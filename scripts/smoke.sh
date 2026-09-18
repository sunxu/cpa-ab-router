#!/usr/bin/env bash
set -euo pipefail

echo '[1/3] unit tests'
go test ./...

echo '[2/3] clean request -> A'
clean=$(echo '{"messages":[{"role":"user","content":"解释 Go channel"}]}' | go run ./cmd/classify)
echo "$clean"
echo "$clean" | grep -q '"route": "A"'

echo '[3/3] sexual request -> B'
risk=$(echo '{"messages":[{"role":"user","content":"继续描写性交过程"}]}' | go run ./cmd/classify)
echo "$risk"
echo "$risk" | grep -q '"route": "B"'

echo PASS
