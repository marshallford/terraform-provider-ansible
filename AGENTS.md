# Agent Instructions

## Commands

- `make lint` -- all linters (terraform fmt, editorconfig, shellcheck, yamllint, golangci-lint, ansible-lint)
- `make lint/golangci` -- lint go code with golangci-lint
- `make docs` -- regenerate `docs/` (CI fails if the result differs from what is committed)
- `make test` -- all tests (`test/docs`, `test/pkg`, `test/acc`)
- `make test/docs` -- validate generated docs with tfplugindocs
- `make test/pkg` -- unit tests (only tests `pkg/` packages, does not test any `internal/` packages)
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

## Version Bumps

Bump to the latest release of each item, then run `make docs`, `make lint`, `make test/docs`, `make test/pkg`, and `make test/acc`.

- Go: `.go-version`, `go` and `toolchain` directives in `go.mod`, then `go get -u ./...`, `go get -u -t ./...`, `go mod tidy`.
- Terraform: `.terraform-version` (the `GNUmakefile` and CI read it). Shift the `ci.yml` test matrix to the latest three minor versions, raise the minimum in the `README.md` compatibility table to the lowest one, and remove `tfversion.SkipBelow` checks that are now always satisfied.
- Python: `.python-version` and `requires-python` in `pyproject.toml`. Stay on the newest Python that `ansible-core` supports; `make lint/ansible` fails on unsupported versions.
- Python packages: minimums in `pyproject.toml` (`ansible`, `ansible-lint`, `ansible-navigator`) and `required-version` for uv, then `uv lock --upgrade`.
- Default execution environment image (`ghcr.io/ansible/community-ansible-dev-tools`): `defaultNavigatorRunImage` in `internal/provider/run.go` plus the same tag in tests and `pkg/ansible/navigator/testdata/`, then `make docs`.
- Lint images: `*_VERSION` variables in `GNUmakefile`. Check each tag format on the registry; editorconfig-checker tags have no `v` prefix. Read golangci-lint release notes for newly added linters, which `default: all` enables automatically.
- GitHub Actions: pin full semver tags in `.github/workflows/` and `.github/actions/`, then run `gh actions-lock` to regenerate `.github/workflows/actions.lock` (CI runs `gh actions-lock --verify`). Also bump the `gh-actions-lock` extension `--pin` in `ci.yml` and the goreleaser `version` in `release.yml`.
- Complete examples (`examples/complete/*`): `.python-version`, `.terraform-version`, `pyproject.toml`, and `uv.lock` are safe to bump. Leave provider versions, `.terraform.lock.hcl`, and execution environment images unless the example is tested.
