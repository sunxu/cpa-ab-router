# cpa-ab-router

Minimal CLIProxyAPI ModelRouter plugin for a sexual-content A/B experiment.

```text
Client -> Router CPA -> plugin -> CPA-A / CPA-B -> Antigravity
```

Routing:

- high-confidence ordinary -> **A**
- sexual / ambiguous / parse failure / unknown -> **B**

The plugin only chooses a provider. Router CPA keeps responsibility for HTTP, streaming/SSE and OpenAI-compatible execution.

## Build

Linux host/container:

```bash
mkdir -p plugins/linux/$(go env GOARCH)
go build -buildmode=c-shared \
  -o plugins/linux/$(go env GOARCH)/sexual-ab-router.so \
  ./cmd/sexual-ab-router
rm -f plugins/linux/$(go env GOARCH)/sexual-ab-router.h
```

macOS:

```bash
mkdir -p plugins/darwin/$(go env GOARCH)
go build -buildmode=c-shared \
  -o plugins/darwin/$(go env GOARCH)/sexual-ab-router.dylib \
  ./cmd/sexual-ab-router
rm -f plugins/darwin/$(go env GOARCH)/sexual-ab-router.h
```

The plugin must be built for the **Router CPA runtime GOOS/GOARCH**, not necessarily the developer Mac.

## Test

```bash
go test ./...
```

Manual classifier:

```bash
echo '{"messages":[{"role":"user","content":"解释 Go channel"}]}' | go run ./cmd/classify
```

Historical log audit:

```bash
go run ./cmd/cpa-ab-audit ~/cpa/logs -o router-audit
```

Outputs:

```text
router-audit/summary.json
router-audit/requests.csv
```

## Configure Router CPA

Start from [`examples/router-config.yaml`](examples/router-config.yaml). The two named OpenAI-compatible providers must remain separate:

```text
cpa-a -> openai-compatible-cpa-a
cpa-b -> openai-compatible-cpa-b
```

See [Architecture](docs/ARCHITECTURE.md), [Experiment](docs/EXPERIMENT.md), [Deployment](docs/DEPLOYMENT.md), and [Validation](docs/VALIDATION.md).
