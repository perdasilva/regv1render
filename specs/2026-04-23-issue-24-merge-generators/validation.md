# Validation

## Acceptance Criteria

1. `internal/render/registryv1/` package no longer exists (including generators and validator subdirs)
2. `internal/render/resourceutil/` package exists with builder helpers
3. `internal/render/validator/` package exists with bundle validator and ValidationError
4. All generators are unexported functions in `internal/render/generators.go`
5. `ResourceGenerator`, `ResourceGenerators`, `BundleValidator`, and `Options` types are not exported
6. `NewRendererBuilder()` takes no arguments
7. `make verify` passes
8. All regression tests pass without golden file changes
9. Test coverage is maintained (same scenarios and assertions)

## Test Scenarios

- `Render()` on a bundle with deployments — verify Deployment objects in output with correct annotations, volumes
- `Render()` on a bundle with permissions — verify Role/RoleBinding in output
- `Render()` on a bundle with cluster permissions — verify ClusterRole/ClusterRoleBinding
- `Render()` on a bundle with CRDs — verify CRD objects
- `Render()` on a bundle with webhooks — verify webhook configs, services, cert resources
- `Render()` on a bundle with provided APIs flag — verify ClusterRoles
- `Render()` on a bundle with unsupported resource — verify error
- All existing regression test cases pass

## Quality Gates

- `make verify` (fmt + vet + lint + test)
- `make build`
- All regression tests pass unchanged
- `ls internal/render/registryv1/` should fail (directory deleted)
- `ls internal/render/validator/` should succeed

## Manual Verification

1. `go doc github.com/perdasilva/rv1` — no `ResourceGenerator` types
2. Run `make build && bin/rv1 --version` — CLI works
3. Run regression tests — golden files unchanged
