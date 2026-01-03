---
description: "GitHub Actions CI rules for Tunez"
applyTo: ".github/workflows/**/*.yml"
---

## Goals
- Keep CI fast, deterministic, and secure.
- Prefer official actions and pinned versions where possible.

## Required checks
- `go test ./...`
- `gofmt` check (or `go fmt` + git diff clean)
- Optional: `golangci-lint` if configured

## Security
- Never print secrets.
- Use least-privilege permissions.
- Cache Go build/test artifacts safely (keyed by go.sum).
