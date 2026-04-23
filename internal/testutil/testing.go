package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/perdasilva/rv1/internal/renderutil"
)

func ToUnstructuredT(t *testing.T, obj client.Object) *unstructured.Unstructured {
	u, err := renderutil.ToUnstructured(obj)
	require.NoError(t, err)
	return u
}
