# Deployment

1. Run two independent CPA instances: CPA-A and CPA-B.
2. Mount separate credential directories (`auth-a/`, `auth-b/`).
3. Configure Router CPA with the two named OpenAI-compatible upstreams from `examples/router-config.yaml`.
4. Build `sexual-ab-router` for the Router CPA runtime OS/architecture and copy it into the configured plugin directory.
5. Verify the plugin is enabled and both provider keys are available:
   - `openai-compatible-cpa-a`
   - `openai-compatible-cpa-b`
6. Send one known-clean request and verify only CPA-A receives it.
7. Send one known-risk test request and verify only CPA-B receives it.
8. Verify streaming requests still work through Router CPA.
9. Do not configure cross-pool fallback. Router-level `request-retry` should remain `0`; retries belong inside CPA-A or CPA-B.

Unknown or malformed content must fail closed to B.
