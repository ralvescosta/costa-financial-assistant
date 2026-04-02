#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INTEGRATION_DIR="${ROOT_DIR}/tests/integration"

required_dirs=(bff bills files identity onboarding payments cross_service helpers)
legacy_pattern='(^|/)us[0-9]+_.*_test\.go$'
filename_pattern='^[a-z0-9]+(_[a-z0-9]+)*_test\.go$'

fail=0

for d in "${required_dirs[@]}"; do
	if [[ ! -d "${INTEGRATION_DIR}/${d}" ]]; then
		echo "[FAIL] missing required integration directory: ${d}"
		fail=1
	fi
done

if find "${INTEGRATION_DIR}" -maxdepth 1 -type f -name '*_test.go' | grep -q .; then
	echo "[FAIL] test files found in root integration directory; tests must live in canonical segment folders"
	find "${INTEGRATION_DIR}" -maxdepth 1 -type f -name '*_test.go' -print
	fail=1
fi

while IFS= read -r f; do
	base="$(basename "${f}")"
	if [[ ! "${base}" =~ ${filename_pattern} ]]; then
		echo "[FAIL] non-canonical test filename: ${f}"
		fail=1
	fi
	if [[ "${f}" =~ ${legacy_pattern} ]]; then
		echo "[FAIL] legacy user-story filename detected: ${f}"
		fail=1
	fi

	if ! grep -Eq 'type[[:space:]]+scenario[[:space:]]+struct|scenarios[[:space:]]*:=[[:space:]]*\[\]helpers\.BDDScenario' "${f}"; then
		echo "[FAIL] missing table-driven scenario structure: ${f}"
		fail=1
	fi
	if ! grep -Eq 'Given|When|Then' "${f}"; then
		echo "[FAIL] missing explicit Given/When/Then structure: ${f}"
		fail=1
	fi
	if ! grep -Eq 'Arrange|Act|Assert' "${f}"; then
		echo "[FAIL] missing AAA sections: ${f}"
		fail=1
	fi
done < <(find "${INTEGRATION_DIR}" -mindepth 2 -type f -name '*_test.go' ! -path '*/helpers/*' | sort)

if [[ ${fail} -ne 0 ]]; then
	echo "integration test convention validation failed"
	exit 1
fi

echo "integration test convention validation passed"
