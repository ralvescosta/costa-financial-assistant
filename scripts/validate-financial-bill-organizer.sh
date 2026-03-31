#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

printf "\n[1/5] Validating docker compose...\n"
cd "$ROOT_DIR"
docker compose config >/dev/null

printf "\n[2/5] Building backend...\n"
cd "$ROOT_DIR/backend"
go build ./...

if [[ "${SKIP_INTEGRATION_TESTS:-0}" != "1" ]]; then
  printf "\n[3/5] Running backend integration checks (history + contracts + metrics)...\n"
  go test -tags integration ./tests/integration/ -run 'History|Timeline|Compliance|OpenAPI|Metrics'
else
  printf "\n[3/5] Skipping backend integration checks (SKIP_INTEGRATION_TESTS=1).\n"
fi

printf "\n[4/5] Building frontend...\n"
cd "$ROOT_DIR/frontend"
npm run build

printf "\n[5/5] Running focused frontend hook tests...\n"
npm test -- --run useHistoryDashboard
npm test -- --run usePaymentDashboard
npm test -- --run useReconciliation

printf "\nValidation completed successfully.\n"
