# Merge Generators into Renderer and Remove Generators Package

Move all 11 generator functions from the `generators` package into the `Renderer` as unexported methods, extract resource builder helpers into a `resourceutil` package, delete the `generators` package, and consolidate the `registryv1` package into the renderer.

## Task Group 1: Extract resource builders into resourceutil (medium)

Move the resource builder helpers out of generators so they can be used by the renderer.

- Create `internal/render/resourceutil/` package
- Move `CreateDeploymentResource`, `CreateRoleResource`, `CreateRoleBindingResource`, `CreateClusterRoleResource`, `CreateClusterRoleBindingResource`, `CreateServiceAccountResource`, `CreateServiceResource` and their option types from `generators/resources.go`
- Move `resources_test.go` tests
- Update imports in generators to use the new package
- Verify `make verify` passes

## Task Group 2: Move generators into renderer (large)

Move all 11 generator functions from exported package-level functions to unexported functions in `internal/render/`.

- Move each generator function from `generators/generators.go` to `internal/render/generators.go` (as unexported functions)
- Wire generators via `defaultGenerators()` called from `NewRendererBuilder`
- Remove `ResourceGenerator` function type and `ResourceGenerators` slice type (now unexported `resourceGenerator`)
- Delete `internal/render/registryv1/generators/` package
- Merge `render.go` into `renderer.go`
- Verify `make verify` passes

## Task Group 3: Refactor tests (large)

Move from testing individual generators to integration-testing `Renderer.Render()` output.

- For each generator test, create an equivalent test that calls `Render()` and checks that expected resources are present in the output (by kind, name, namespace, key fields)
- Preserve all error message assertions and edge case coverage
- Keep regression tests unchanged (golden files verify exact output)
- Remove individual generator test functions
- Merge `render_test.go` into `renderer_test.go`
- Verify test counts and coverage are maintained

## Task Group 4: Remove registryv1 package (medium)

Remove the `registryv1` package by absorbing its responsibilities into the renderer.

- Change `NewRendererBuilder()` to take no args — always uses standard validator
- Move validator package from `registryv1/validator/` to `internal/render/validator/`
- Move `ValidationError` from `render` to `validator` package (breaks import cycle)
- Update public API `render.go` to import validator directly
- Delete `internal/render/registryv1/` entirely
- Move `registryv1_test.go` tests into `renderer_test.go`
- Update all test bundles with `PackageName` and valid fields for the always-present validator

## Task Group 5: Unexport internal types (small)

- Unexport `BundleValidator` → `bundleValidator` (only one implementation, always wired internally)
- Unexport `Options` → `options` (only used by internal generators)
- Move `certprovider_test.go` to `package render` (internal test) to access unexported types
- Update CLAUDE.md architecture section
- Run final `make verify`
