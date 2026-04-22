package render

import (
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
)

// DeploymentConfig is a type alias for v1alpha1.SubscriptionConfig
// to maintain clear naming in the OLMv1 context while reusing the v0 type.
type DeploymentConfig = v1alpha1.SubscriptionConfig
