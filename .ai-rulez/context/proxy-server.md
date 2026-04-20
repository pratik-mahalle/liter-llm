---
priority: high
---

# Proxy Server (`crates/liter-llm-proxy`)

OpenAI-compatible HTTP proxy built on axum. Served via `liter-llm api --config <path>`.

## Virtual API Keys

- Configured in `liter-llm-proxy.toml` under `[[keys]]`
- Master key (unrestricted) and virtual keys (per-key model allowlist + rate/budget limits)
- `VirtualKeyConfig`: `key`, `description`, `models` (allowlist), `rpm`, `tpm`, `budget_limit`
- Config supports `${ENV_VAR}` interpolation for secrets

## Model Routing

- Each model declared as `[[models]]` with `name` (public alias) and `provider_model`
- `ServicePool` maps model name → Tower `BoxCloneService` stack
- `AliasEntry` (`[[aliases]]`) for pattern-based routing with optional credential overrides

## Tower Middleware Stack (innermost → outermost)

1. `CacheLayer` — semantic response cache
2. `HealthCheckLayer` — periodic probe
3. `CooldownLayer` — circuit-breaker after consecutive errors
4. `ModelRateLimitLayer` — RPM/TPM enforcement
5. `CostTrackingLayer` — token cost accounting (pricing from `schemas/pricing.json`)
6. `BudgetLayer` — hard/soft budget enforcement
7. `TracingLayer` — tracing spans

## API Surface (22 endpoints under `/v1/`)

Chat completions, embeddings, models, images, audio, moderations, files, batches, responses, rerank, search, OCR. Health and OpenAPI endpoints unauthenticated.

## Contract Tests

- `task proxy:schemathesis` — runs schemathesis against `/openapi.json`
- Integration tests in `crates/liter-llm-proxy/tests/`
