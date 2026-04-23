# Validation

## Acceptance Criteria

1. `WithProvidedAPIsClusterRoles()` option exists and is exported from the public API
2. Rendering with the option produces 4 additional ClusterRoles per owned CRD (admin/edit/view/crd-view)
3. Generated ClusterRoles have correct names, verbs, API groups, resources, and aggregation labels
4. Rendering without the option produces identical output to before this change
5. `make verify` passes (fmt + vet + lint + test)
6. `make build` still produces a working rv1 binary
7. All existing tests and regression tests still pass

## Test Scenarios

- Render a bundle with 1 owned CRD using `WithProvidedAPIsClusterRoles()` — verify 4 ClusterRoles generated (admin/edit/view/crd-view) with correct naming and verbs
- Render a bundle with 3 owned CRDs — verify 12 ClusterRoles generated
- Render a bundle with no owned CRDs — verify no extra ClusterRoles generated
- Render a bundle with `WithProvidedAPIsClusterRoles()` NOT set — verify no extra ClusterRoles (backward compat)
- Verify aggregation labels: admin role has `aggregate-to-admin: "true"`, edit has `aggregate-to-edit: "true"`, view has `aggregate-to-view: "true"`
- Verify role verbs: admin gets `*`, edit gets `create/update/patch/delete`, view gets `get/list/watch`, crd-view gets `get` on CRD resource
- Run all existing regression tests — verify no output changes

## Quality Gates

- `make verify` (fmt + vet + lint + test)
- `make build`
- All existing unit and regression tests pass unchanged
- New tests cover the generator, the option, and edge cases

## Manual Verification

1. Run `go doc github.com/perdasilva/rv1 WithProvidedAPIsClusterRoles` — verify godoc exists
2. Run `go test -v -run ProvidedAPI ./...` — verify new tests pass
3. Run existing regression tests — verify no golden file changes
