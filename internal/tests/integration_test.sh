#!/usr/bin/env bash

set -eu -o pipefail

declare root_dir
root_dir=$(git rev-parse --show-toplevel)
readonly root_dir

declare -r work_dir=/tmp/nodeinfo_client_test
declare -r main_bin="${work_dir}/main"
declare -r popular_host='mastodon.social'

export GOCOVERDIR="${work_dir}/covdata"

function main () {
	umask 077

	rm -rf "${work_dir}"
	mkdir -pv "${work_dir}" "${GOCOVERDIR}"

	go build -o "${main_bin}" -cover "${root_dir}/internal/tests"

	if ! "${main_bin}" -H "${popular_host}" "discover_one" | tee "${work_dir}/discover_one.json" ; then
		>&2 echo "discover_one failed"
		return 1
	fi

	local href
	href=$(jq -r '.[0].href' < "${work_dir}/discover_one.json")
	readonly href

	if ! "${main_bin}" -U "${href}" "get_one" | tee "${work_dir}/get_one.json" ; then
		>&2 echo "get_one failed"
		return 1
	fi

	if ! "${main_bin}" -client-timeout 10s "batch_discovery" < "${root_dir}/internal/tests/popular_hosts.txt" | tee "${work_dir}/batch_discover.json" ; then
		>&2 echo "batch_discovery failed"
		return 1
	fi

	# collect href values where the declared version is either 2.0 or 2.1
	jq -rc 'select(.err == null).links[]? | select(.rel | (endswith("2.0") or endswith("2.1"))).href' < "${work_dir}/batch_discover.json" > "${work_dir}/batch_discover_href_v2"
	if ! "${main_bin}" "batch_nodeinfo" < "${work_dir}/batch_discover_href_v2" | tee "${work_dir}/batch_nodeinfo_v2.json" ; then
		>&2 echo "batch_nodeinfo failed"
		return 1
	fi

	>&2 echo "OK, see data in ${work_dir}"

	go tool covdata merge -i "${GOCOVERDIR}" -o "${work_dir}"
	go tool covdata percent -i "${work_dir}" -pkg github.com/rafaelespinoza/nodeinfo/nodeinfo
}

main "${@}"
