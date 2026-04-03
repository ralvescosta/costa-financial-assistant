#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INTEGRATION_DIR="${ROOT_DIR}/tests/integration"

required_dirs=(bff bills files identity onboarding payments cross_service helpers)
legacy_pattern='(^|/)us[0-9]+_.*_test\.go$'
filename_pattern='^[a-z0-9]+(_[a-z0-9]+)*_test\.go$'
baseline_file_default="${ROOT_DIR}/scripts/integration_convention_known_failures.txt"

strict_mode="${STRICT_CONVENTION_VALIDATION:-0}"
baseline_file="${INTEGRATION_CONVENTION_BASELINE_FILE:-${baseline_file_default}}"

declare -a baseline_entries=()
declare -A baseline_lookup=()
declare -a unexpected_failures=()
declare -a matched_baseline=()

if [[ "${strict_mode}" != "1" && -f "${baseline_file}" ]]; then
	while IFS= read -r line; do
		if [[ -z "${line}" || "${line}" =~ ^# ]]; then
			continue
		fi
		baseline_entries+=("${line}")
		baseline_lookup["${line}"]=1
	done < "${baseline_file}"
fi

record_failure() {
	local message="$1"
	if [[ "${strict_mode}" != "1" && ${#baseline_lookup[@]} -gt 0 && -n "${baseline_lookup["${message}"]+_}" ]]; then
		matched_baseline+=("${message}")
		return
	fi
	unexpected_failures+=("${message}")
}

for d in "${required_dirs[@]}"; do
	if [[ ! -d "${INTEGRATION_DIR}/${d}" ]]; then
		record_failure "[FAIL] missing required integration directory: ${d}"
	fi
done

if find "${INTEGRATION_DIR}" -maxdepth 1 -type f -name '*_test.go' | grep -q .; then
	record_failure "[FAIL] test files found in root integration directory; tests must live in canonical segment folders"
	while IFS= read -r root_test_file; do
		record_failure "${root_test_file}"
	done < <(find "${INTEGRATION_DIR}" -maxdepth 1 -type f -name '*_test.go' -print | sort)
fi

while IFS= read -r f; do
	base="$(basename "${f}")"
	if [[ ! "${base}" =~ ${filename_pattern} ]]; then
		record_failure "[FAIL] non-canonical test filename: ${f}"
	fi
	if [[ "${f}" =~ ${legacy_pattern} ]]; then
		record_failure "[FAIL] legacy user-story filename detected: ${f}"
	fi

	if ! grep -Eq 'type[[:space:]]+scenario[[:space:]]+struct|scenarios[[:space:]]*:=[[:space:]]*\[\]helpers\.BDDScenario' "${f}"; then
		record_failure "[FAIL] missing table-driven scenario structure: ${f}"
	fi
	if ! grep -Eq 'Given|When|Then' "${f}"; then
		record_failure "[FAIL] missing explicit Given/When/Then structure: ${f}"
	fi
	if ! grep -Eq 'Arrange|Act|Assert' "${f}"; then
		record_failure "[FAIL] missing AAA sections: ${f}"
	fi
done < <(find "${INTEGRATION_DIR}" -mindepth 2 -type f -name '*_test.go' ! -path '*/helpers/*' | sort)

if [[ ${#unexpected_failures[@]} -gt 0 ]]; then
	for failure in "${unexpected_failures[@]}"; do
		echo "${failure}"
	done
	echo "integration test convention validation failed"
	exit 1
fi

if [[ "${strict_mode}" != "1" && ${#baseline_entries[@]} -gt 0 ]]; then
	declare -A matched_lookup=()
	for matched in "${matched_baseline[@]}"; do
		matched_lookup["${matched}"]=1
	done

	declare -a stale_entries=()
	for baseline_entry in "${baseline_entries[@]}"; do
		if [[ -z "${matched_lookup["${baseline_entry}"]+_}" ]]; then
			stale_entries+=("${baseline_entry}")
		fi
	done

	if [[ ${#stale_entries[@]} -gt 0 ]]; then
		echo "integration test convention validation passed, but baseline has stale entries (remove these from ${baseline_file}):"
		for stale in "${stale_entries[@]}"; do
			echo "${stale}"
		done
		exit 0
	fi
fi

echo "integration test convention validation passed"
