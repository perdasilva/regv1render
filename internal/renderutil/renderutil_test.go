package renderutil_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/perdasilva/rv1/internal/renderutil"
)

type unmarshalable struct{}

func (u *unmarshalable) MarshalJSON() ([]byte, error) {
	return nil, errors.New("unmarshalable")
}

func TestDeepHashObject(t *testing.T) {
	tests := []struct {
		name         string
		wantPanic    bool
		obj          interface{}
		expectedHash string
	}{
		{
			name: "populated obj with exported fields",
			obj: struct {
				Str string
				Num int
				Obj interface{}
				Arr []int
				B   bool
				N   interface{}
			}{
				Str: "foobar",
				Num: 22,
				Obj: struct{ Foo string }{Foo: "bar"},
				Arr: []int{0, 1},
				B:   true,
				N:   nil,
			},
			expectedHash: "gta1bt5sybll5qjqxdiekmjm7py93glrinmnrfb31fj",
		},
		{
			name: "modified populated obj with exported fields",
			obj: struct {
				Str string
				Num int
				Obj interface{}
				Arr []int
				B   bool
				N   interface{}
			}{
				Str: "foobar",
				Num: 23,
				Obj: struct{ Foo string }{Foo: "bar"},
				Arr: []int{0, 1},
				B:   true,
				N:   nil,
			},
			expectedHash: "1ftn1z2ieih8hsmi2a8c6mkoef6uodrtn4wtt1qapioh",
		},
		{
			name: "populated obj with unexported fields",
			obj: struct {
				str string
				num int
				obj interface{}
				arr []int
				b   bool
				n   interface{}
			}{
				str: "foobar",
				num: 22,
				obj: struct{ foo string }{foo: "bar"},
				arr: []int{0, 1},
				b:   true,
				n:   nil,
			},
			expectedHash: "16jfjhihxbzhfhs1k5mimq740kvioi98pfbea9q6qtf9",
		},
		{
			name:         "empty obj",
			obj:          struct{}{},
			expectedHash: "16jfjhihxbzhfhs1k5mimq740kvioi98pfbea9q6qtf9",
		},
		{
			name:         "string a",
			obj:          "a",
			expectedHash: "1lu1qv1451mq7gv9upu1cx8ffffi07rel5xvbvvc44dh",
		},
		{
			name:         "string b",
			obj:          "b",
			expectedHash: "1ija85ah4gd0beltpfhszipkxfyqqxhp94tf2mjfgq61",
		},
		{
			name:         "nil obj",
			obj:          nil,
			expectedHash: "2im0kl1kwvzn46sr4cdtkvmdzrlurvj51xdzhwdht8l0",
		},
		{
			name:      "unmarshalable obj",
			obj:       &unmarshalable{},
			wantPanic: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			test := func() {
				hash := renderutil.DeepHashObject(tc.obj)
				assert.Equal(t, tc.expectedHash, hash)
			}

			if tc.wantPanic {
				require.Panics(t, test)
			} else {
				require.NotPanics(t, test)
			}
		})
	}
}

func Test_ObjectNameForBaseAndSuffix(t *testing.T) {
	name := renderutil.ObjectNameForBaseAndSuffix("my.object.thing.has.a.really.really.really.really.really.long.name", "suffix")
	require.Len(t, name, 63)
	require.Equal(t, "my.object.thing.has.a.really.really.really.really.really-suffix", name)
}

func Test_ToUnstructured(t *testing.T) {
	for _, tc := range []struct {
		name string
		obj  client.Object
		err  error
	}{
		{
			name: "converts object to unstructured",
			obj: &corev1.Service{
				TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "v1"},
				ObjectMeta: metav1.ObjectMeta{Name: "my-service", Namespace: "my-namespace"},
			},
		}, {
			name: "fails if object doesn't define kind",
			obj: &corev1.Service{
				TypeMeta:   metav1.TypeMeta{Kind: "", APIVersion: "v1"},
				ObjectMeta: metav1.ObjectMeta{Name: "my-service", Namespace: "my-namespace"},
			},
			err: errors.New("object has no kind"),
		}, {
			name: "fails if object doesn't define version",
			obj: &corev1.Service{
				TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: ""},
				ObjectMeta: metav1.ObjectMeta{Name: "my-service", Namespace: "my-namespace"},
			},
			err: errors.New("object has no version"),
		}, {
			name: "fails if object is nil",
			err:  errors.New("object is nil"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := renderutil.ToUnstructured(tc.obj)
			if tc.err != nil {
				require.Error(t, err)
			} else {
				assert.Equal(t, tc.obj.GetObjectKind().GroupVersionKind(), out.GroupVersionKind())
			}
		})
	}
}

func TestMergeMaps(t *testing.T) {
	tests := []struct {
		name      string
		maps      []map[string]string
		expectMap map[string]string
	}{
		{
			name:      "no maps",
			maps:      make([]map[string]string, 0),
			expectMap: map[string]string{},
		},
		{
			name:      "two empty maps",
			maps:      []map[string]string{{}, {}},
			expectMap: map[string]string{},
		},
		{
			name:      "simple add",
			maps:      []map[string]string{{"foo": "bar"}, {"bar": "foo"}},
			expectMap: map[string]string{"foo": "bar", "bar": "foo"},
		},
		{
			name:      "subsequent maps overwrite prior",
			maps:      []map[string]string{{"foo": "bar", "bar": "foo"}, {"foo": "foo"}, {"bar": "bar"}},
			expectMap: map[string]string{"foo": "foo", "bar": "bar"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equalf(t, tc.expectMap, renderutil.MergeMaps(tc.maps...), "maps did not merge as expected")
		})
	}
}
