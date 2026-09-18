# A/B experiment

## Primary endpoint

A credential is counted as `AUTH_DEAD` only when both are observed:

1. upstream request: `401 UNAUTHENTICATED`
2. token refresh: `invalid_grant`

429, 5xx, TLS errors and timeouts are not account deaths.

## Pools

- **A** receives only payloads classified high-confidence ordinary.
- **B** receives sexual, uncertain and fail-closed traffic.

Historical upstream audit was approximately A 9% / B 91%. Treat this only as an initial sizing estimate: run `cpa-ab-audit` against the same raw request corpus with the production classifier before assigning pool sizes.

Keep average upstream attempts per alive credential reasonably comparable between A and B; otherwise load becomes a confounder.

## Metrics

At minimum record per credential:

- pool (A/B)
- first_used_at / last_success_at
- logical_requests / upstream_attempts
- first_401_at
- first_invalid_grant_at
- auth_dead

Compare:

```text
A death rate = AUTH_DEAD_A / A credentials
B death rate = AUTH_DEAD_B / B credentials
```

Also audit **A contamination**. Any post-hoc confirmed sexual request routed to A is a critical experimental error.
