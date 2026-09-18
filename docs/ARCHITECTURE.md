# Architecture

## Goal

Test whether credentials exposed to sexual/uncertain request content have a higher `AUTH_DEAD` rate than credentials receiving only high-confidence ordinary content.

## Data path

```text
Client
  -> Router CPA
      -> sexual-ab-router ModelRouter plugin
          -> A: openai-compatible-cpa-a -> CPA-A -> auth-a/
          -> B: openai-compatible-cpa-b -> CPA-B -> auth-b/
```

## Frozen responsibilities

- **Plugin**: inspect the current complete request body and choose A or B.
- **Router CPA**: protocol handling, OpenAI-compatible upstream execution, streaming/SSE, response forwarding.
- **CPA-A / CPA-B**: Antigravity credentials, retries, cooldown, OAuth and Google upstream execution.

The plugin is not an executor and does not issue HTTP requests.

## Routing rule

- A: high-confidence ordinary only.
- B: confirmed sexual, ambiguous sexual, unknown/unparseable, or otherwise not safe enough for A.
- USER_CHAT and INTERNAL are treated identically.
- No session-sticky state. Carried history is classified because it is part of the current request body.
- Fail closed: any classifier or request parsing uncertainty routes to B.

## Isolation

CPA-A and CPA-B must use separate credential directories. Retry/fallback may happen inside a pool, never across pools. Router CPA contains no Antigravity credentials.
