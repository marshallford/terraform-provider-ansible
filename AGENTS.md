# Agent Instructions

## Commands

- `make lint/golangci` -- lint go code with golangci-lint
- `make lint/pkg` -- unit tests (only tests `pkg/` packages, does not test any `internal/` packages)
- `make test/acc` -- acceptance tests (only tests `internal/provider` package, does not test any `pkg/` packages)
- `make test/acc TESTARGS="--run TestName"` -- run individual acceptance tests
- `make build` -- quick compile check

## Package Design

- `pkg/ansible` -- generic ansible types, interfaces, constants. Nothing terraform-specific.
- `pkg/ansible/navigator` -- ansible-navigator orchestration. Depends on `pkg/ansible`. Nothing terraform-specific.
- `internal/provider` -- terraform provider. Depends on `pkg/`

## Conventions

- Compile-time interface checks: `var _ Interface = (*impl)(nil)` in the same file as the implementation.
- Do not edit `docs/` directly. They are generated from `examples/` with `make docs`.
