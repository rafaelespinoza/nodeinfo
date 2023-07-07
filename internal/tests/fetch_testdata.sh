#!/usr/bin/env bash

set -eu -o pipefail

declare CURR_SCRIPT
CURR_SCRIPT="$(basename "$0")"
readonly CURR_SCRIPT

declare -r BASE_URI='https://api.fedidb.org'
declare -r BASE_WORK_DIR=/tmp/fedidb
declare -r MAX_NUM_FETCHES=200
declare -r DOMAINS_FILE='domains.txt'

function _get_servers () {
	local -r cursor="${1:-''}"

	local req_path="v1/servers?limit=40"
	if [[ -n "${cursor}" ]]; then
		req_path+="&cursor=${cursor}"
	fi

	local -r req_uri="${BASE_URI}/${req_path}"
	>&2 echo "# getting servers, req_uri=${req_uri}"
	curl -H 'Accept:application/json' "${req_uri}"
}

function usage () {
	echo "Usage:
	${CURR_SCRIPT} [-c cursor] [-d data_dir] [-h]

Description:
	This script could be used for creating testdata for integration tests. It
	fetches paginated data as JSON about servers in the Fediverse from
	api.fedidb.org. API documentation: https://fedidb.org/docs/api/v1.

	Data is written to a temporary directory, created upon each script
	invocation. By default it attempts to fetch as much data as is available,
	but it can be configured to write to a pick up where it left off on a
	previous invocation.

	The data directory can be overridden with flag -d.
	The pagination cursor can be specified with flag -c.

	The script will write some variables, including the pagination cursor, to
	STDERR for easier introspection The pagination cursor may also be found in
	the response body data at key path: '.meta.next_cursor'

	At the end, results are sorted and written to a file: ${DOMAINS_FILE}.
	Whether or not to copy the results into version control is left up to the
	caller's discretion.

Examples:
	# fetch everything
	${CURR_SCRIPT}

	# did it stop partway through? pick it up again by specifying the data
	# directory from last time and the next pagination cursor.
	${CURR_SCRIPT} -d /tmp/fedidb.EXAMPLE -c EXAMPLE_CURSOR_eyJ1c2VyX2N
"
}

function main () {
	umask 077

	local data_dir=''
	local cursor=''

	while getopts ":c:d:h" opt; do
		case "${opt}" in
			c )
				cursor="${OPTARG}"
				;;
			d )
				data_dir="${OPTARG}"
				;;
			h )
				usage
				return 0
				;;
			* )
				;;
		esac
	done

	if [[ -z "${data_dir}" ]]; then
		data_dir=$(mktemp -d "${BASE_WORK_DIR}.XXXXX")
		>&2 echo "data_dir=${data_dir}"
	fi
	readonly data_dir

	local result_file=''
	local num_fetches=0
	local num_results_last_fetch=0

	while : ; do
		if [[ "$num_fetches" -ge "${MAX_NUM_FETCHES}" ]] ; then
			>&2 echo "# we've reached the max number of fetches, goodbye"
			break
		fi

		result_file="${data_dir}/list_servers.$(date --utc +%s.%N).json" # sequence filenames
		_get_servers "${cursor}" | jq -c > "${result_file}"
		num_results_last_fetch=$(jq -rc '.data | length' < "${result_file}")
		cursor=$(jq -rc '.meta.next_cursor' < "${result_file}")
		num_fetches=$((num_fetches + 1))

		# capture application testdata.
		jq -cr '.data[].domain' < "${result_file}" >> "${data_dir}/.unsorted_${DOMAINS_FILE}"

		>&2 echo "
# num_fetches=${num_fetches}
# max_num_fetches=${MAX_NUM_FETCHES}
# num_results_last_fetch=${num_results_last_fetch}
# cursor=${cursor}"

		if [[ "${num_results_last_fetch}" -lt 1 ]]; then
			>&2 echo "# apparently there are no more results, goodbye"
			break
		fi

		>&2 echo "# sleeping"
		sleep 1 # be nice to the API
	done

	sort < "${data_dir}/.unsorted_${DOMAINS_FILE}" > "${data_dir}/${DOMAINS_FILE}"

	>&2 echo "# see collected domain names at ${data_dir}/${DOMAINS_FILE}"
}

main "$@"
