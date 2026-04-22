# Requirements

## Functional Requirements

- A new `WithProvidedAPIsClusterRoles()` rendering option enables generation of aggregated admin/edit/view ClusterRoles per owned CRD
- For each owned CRD in the CSV's `spec.customresourcedefinitions.owned`, four ClusterRoles are generated:
  - `<name>-<version>-admin` — verbs: `*`, label: `rbac.authorization.k8s.io/aggregate-to-admin: "true"`
  - `<name>-<version>-edit` — verbs: `create`, `update`, `patch`, `delete`, label: `rbac.authorization.k8s.io/aggregate-to-edit: "true"`
  - `<name>-<version>-view` — verbs: `get`, `list`, `watch`, label: `rbac.authorization.k8s.io/aggregate-to-view: "true"`
  - `<name>-<version>-crd-view` — verbs: `get` on `apiextensions.k8s.io/customresourcedefinitions` for the specific CRD, label: `rbac.authorization.k8s.io/aggregate-to-view: "true"`
- The option is opt-in — the `DefaultRenderer` without this option produces no provided API ClusterRoles
- The generated ClusterRoles match OLMv0 behavior from `operator-lifecycle-manager`

## Non-Functional Requirements

- **Backward compatible** — existing rendering behavior is unchanged unless the option is explicitly set
- **Composable** — the option follows the existing functional options pattern (`WithTargetNamespaces`, `WithCertificateProvider`, etc.)
- **Well-tested** — unit tests cover the generator, the option wiring, and edge cases

## Constraints

- Only generate roles for owned CRDs (not required CRDs, not APIServices)
- Do not modify existing generators or validators
- Do not change the default rendering output — this is strictly opt-in
- Follow OLMv0 naming conventions exactly for backward compatibility

## Dependencies

- Existing rendering pipeline from epic #2
- `github.com/operator-framework/api/pkg/operators/v1alpha1` (CSV types, already a dependency)
- `k8s.io/api/rbac/v1` (ClusterRole types, already a dependency)
