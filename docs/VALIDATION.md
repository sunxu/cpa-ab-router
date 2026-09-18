# Validation

The repository intentionally contains no real request logs or credentials.

Local validation performed while preparing the initial implementation:

- `go test ./...`: PASS
- `scripts/smoke.sh`: PASS
- `go build -buildmode=c-shared`: PASS on Linux/amd64
- historical CPA log replay:
  - 113 logs: A 12 / B 101 = 10.62% / 89.38%
  - 514 logs: A 47 / B 467 = 9.14% / 90.86%
- on the 514-log validation set, no request classified `A` by the new Router gate corresponded to an old upstream attempt classified as confirmed `sexual` by the previous v3 audit.

The old v3 upstream audit and the Router classifier are not identical mechanisms. Before production rollout, replay the full historical corpus with the exact production build and manually review a sample of A results.

Recommended gate before rollout:

1. `go test ./...`
2. `make audit LOG_DIR=/path/to/historical/logs`
3. review `router-audit/requests.csv` for every A row or a statistically meaningful sample
4. deploy with a small credential cohort first
