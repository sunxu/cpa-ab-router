.PHONY: test smoke build audit

test:
	go test ./...

smoke:
	./scripts/smoke.sh

build:
	mkdir -p dist
	go build -buildmode=c-shared -o dist/sexual-ab-router.so ./cmd/sexual-ab-router
	rm -f dist/sexual-ab-router.h

audit:
	@test -n "$(LOG_DIR)" || (echo 'usage: make audit LOG_DIR=/path/to/logs' && exit 2)
	go run ./cmd/cpa-ab-audit -o router-audit "$(LOG_DIR)"
