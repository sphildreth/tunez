---
description: "Go development rules for Tunez"
applyTo: "**/*.go"
---

## Go standards
- Target Go >= 1.22 (use idiomatic, standard library-first patterns).
- Run `gofmt` on all changes.
- Prefer small packages and explicit dependencies.
- Avoid global mutable state.

## Errors & logging
- Wrap errors with `%w` and add context.
- Return errors rather than printing.
- No panics in normal control flow.
- Never log secrets / tokens / full filesystem paths unless explicitly required.

## Concurrency
- Use `context.Context` for cancellation.
- Bound concurrency (worker pools, semaphores) when scanning or fetching.
- Avoid goroutine leaks: every goroutine must exit on context cancellation.

## Testing
- Use table-driven tests where appropriate.
- Prefer deterministic tests (no sleeps; use fakes/time control when needed).
- Keep tests fast and runnable with `go test ./...`.

## Bubble Tea constraints
- Never do I/O directly inside `Update`.
- Use `tea.Cmd` to start background work and return a message.
