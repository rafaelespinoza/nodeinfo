#!/usr/bin/env bash

set -eu -o pipefail

declare GO="${GO:-$(command -v go)}"

declare root_dir
root_dir=$(git rev-parse --show-toplevel)
readonly root_dir

declare -r work_dir=/tmp/nodeinfo_client_test
declare -r main_bin="${work_dir}/main"

export GOCOVERDIR="${work_dir}/covdata"

function main () {
	local -r num_hosts="${1:-1024}"

	umask 077

	rm -rf "${work_dir}"
	mkdir -pv "${work_dir}" "${GOCOVERDIR}"

	"${GO}" build -o "${main_bin}" -cover "${root_dir}/internal/tests"

	# There may be several thousand hostnames to choose from. Do we need to check all of them?
	# Probably not. But let's keep it interesting and pick some random ones.
	local hostnames
	hostnames=$(shuf -n "${num_hosts}" "${root_dir}/internal/tests/testdata/popular_hosts.txt")
	readonly hostnames

	local -r batch_discover_file="${work_dir}/batch_discover.json"
	if ! "${main_bin}" "batch_discovery" < <(echo "${hostnames}") | tee "${batch_discover_file}" ; then
		>&2 echo "batch_discovery failed"
		return 1
	fi

	# collect href values where the declared version is either 2.0 or 2.1
	local -r hrefs_file="${work_dir}/batch_discover_href_v2"
	jq -rc 'select(.err == null).links[]? | select(.rel | (endswith("2.0") or endswith("2.1"))).href' < "${batch_discover_file}" > "${hrefs_file}"
	if ! "${main_bin}" "batch_nodeinfo" < "${hrefs_file}" | tee "${work_dir}/batch_nodeinfo_v2.json" ; then
		>&2 echo "batch_nodeinfo failed"
		return 1
	fi

	>&2 echo "OK, see data in ${work_dir}"

	"${GO}" tool covdata merge -i "${GOCOVERDIR}" -o "${work_dir}"
	"${GO}" tool covdata percent -i "${work_dir}" -pkg github.com/rafaelespinoza/nodeinfo/nodeinfo
}

main "${@}"
