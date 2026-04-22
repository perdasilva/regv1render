# OLMv0 Rendering Compatibility

Add opt-in support for OLMv0-specific rendering behavior: generating aggregated admin/edit/view ClusterRoles for each owned CRD defined in the CSV. In OLMv0 (operator-lifecycle-manager), these roles are created so that cluster admins can grant fine-grained RBAC access to custom resources via Kubernetes role aggregation. This is exposed as a new `WithProvidedAPIsClusterRoles()` option on the existing renderer.

## Task Group 1: Analyze OLMv0 provided API ClusterRoles (small)

Study how OLMv0 generates the aggregated ClusterRoles and determine the exact output format.

- Read `operator-lifecycle-manager/pkg/controller/operators/olm/operatorgroup.go` — `ensureClusterRolesForCSV` function
- Document the naming convention: `<crd-name>-<version>-<suffix>` where suffix is `admin`, `edit`, or `view`
- Document the aggregation labels: `rbac.authorization.k8s.io/aggregate-to-admin`, `aggregate-to-edit`, `aggregate-to-view`
- Document the RBAC verbs per role: admin gets all verbs, edit gets all except `deletecollection`, view gets `get`/`list`/`watch`
- Document which CRD APIs are included: only owned CRDs from the CSV's `spec.customresourcedefinitions.owned`
- Write down the expected output for a sample CRD to use as a test fixture

## Task Group 2: Implement the generator (medium)

Create a new `ResourceGenerator` that produces admin/edit/view ClusterRoles for each owned CRD.

- Create a new generator function in `internal/render/registryv1/generators/` (or a new file if cleaner)
- For each owned CRD in `rv1.CSV.Spec.CustomResourceDefinitions.Owned`, generate 3 ClusterRoles:
  - `<plural>.<group>-<version>-admin` with verbs `*`
  - `<plural>.<group>-<version>-edit` with verbs `create`, `update`, `patch`, `delete`, `get`, `list`, `watch`
  - `<plural>.<group>-<version>-view` with verbs `get`, `list`, `watch`
- Each ClusterRole should have the appropriate aggregation label
- The generator should only run when the option is enabled (not by default)

## Task Group 3: Add the option and wire it up (small)

Expose the new generator as an opt-in rendering option.

- Add `WithProvidedAPIsClusterRoles()` option to `internal/render/render.go`
- Add a field to `Options` to track whether provided API roles are requested
- Wire the generator into the rendering pipeline — when the option is set, append the generator to the list
- Export the option from the public API at the repo root

## Task Group 4: Add tests (medium)

Write tests covering the new generator and the opt-in behavior.

- Unit tests for the generator: verify correct ClusterRole names, verbs, labels, and aggregation labels for single and multiple owned CRDs
- Test that the generator is NOT invoked by default (DefaultRenderer without the option produces no extra ClusterRoles)
- Test that adding `WithProvidedAPIsClusterRoles()` produces the expected additional ClusterRoles
- Test edge cases: CSV with no owned CRDs, CRDs with multiple versions
- Add regression test case(s) if appropriate

## Task Group 5: Document and clean up (small)

Document the new option and verify everything is consistent.

- Add godoc to the public `WithProvidedAPIsClusterRoles()` option
- Document the OLMv0 compatibility behavior in a brief note (README or godoc)
- Run `make verify` — all tests pass
- Update CLAUDE.md if architecture needs changes
