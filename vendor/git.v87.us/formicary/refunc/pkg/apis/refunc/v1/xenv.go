package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// CRD names for Runner templates
const (
	XenvKind       = "Xenv"
	XenvPluralName = "xenvs"
)

var (
	_ runtime.Object            = (*Xenv)(nil)
	_ metav1.ObjectMetaAccessor = (*Xenv)(nil)

	_ runtime.Object          = (*XenvList)(nil)
	_ metav1.ListMetaAccessor = (*XenvList)(nil)
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Xenv is a API object to represent a contaniner based eXecution ENVironment for a function
type Xenv struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec XenvSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// XenvList is a API object to represent a list of Xenv
type XenvList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Xenv `json:"items"`
}

// XenvSpec is the specification to describe a runner
// #
// # ┌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┐
// # | k8s metadata |
// # └╌╌╌╌╌╌╌╌╌╌╌╌╌-┘
// # ┌-╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┐
// # | xenv spec                                                                                            |
// # |======================================================================================================|
// # |                                                                                                      |
// # | // Type of setup, default is "agent"                                                                 |
// # | Type string `json:"type,omitempty"`                                                                  |
// # |                                                                                                      |
// # | // https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.10/#container-v1-core            |
// # | Container        corev1.Container              `json:"container"`                                    |
// # | // https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.10/#volume-v1-core               |
// # | Volumes          []corev1.Volume               `json:"volumes,omitempty"`                            |
// # | // https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.10/#localobjectreference-v1-core |
// # | ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`                   |
// # |                                                                                                      |
// # | // How many pre-initialized from template                                                            |
// # | PoolSize int `json:"poolSize,omitempty"`                                                             |
// # |                                                                                                      |
// # | // ServiceAccount attach to xevn dep                                                                 |
// # | ServiceAccount string `json:"serviceAccount,omitempty"`                                              |
// # |                                                                                                      |
// # | // A key used for runtime builder to access the shell                                                |
// # | SetupKey string `json:"key"`                                                                         |
// # |                                                                                                      |
// # └╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌-┘
// #
// apiVersion: k8s.refunc.io/v1
// kind: Xenv
// metadata:
//   name: mfe
//   namespace: ""
// spec:
//   container:
//     image: registry.v87.us/multifactor/mfe/runner:1f4ea64
//     imagePullPolicy: Always
//     name: executor
//     volumeMounts:
//     - mountPath: /var/run/refunc/cache
//       name: local-cache
//   imagePullSecrets:
//   - name: registry-v87-us
//   key: UiqK9l4w9yrxlSrF5ZuSp
//   poolSize: 6
//   type: agent
//   volumes:
//   - name: local-cache
//     persistentVolumeClaim:
//       claimName: local-cache
type XenvSpec struct {
	// Type of setup, default is agent mode
	Type string `json:"type,omitempty"`

	Container        corev1.Container              `json:"container"`
	Volumes          []corev1.Volume               `json:"volumes,omitempty"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// How many pre-initialized from template
	PoolSize int `json:"poolSize,omitempty"`

	// ServiceAccount attach to xevn dep
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// A key used for runtime builder to access the shell
	SetupKey string `json:"key"`
}

// AsOwner returns *metav1.OwnerReference
func (env *Xenv) AsOwner() *metav1.OwnerReference {
	return &metav1.OwnerReference{
		APIVersion: APIVersion,
		Kind:       XenvKind,
		Name:       env.Name,
		UID:        env.UID,
		Controller: &trueVar,
	}
}

// Ref returns *corev1.ObjectReference
func (env *Xenv) Ref() *corev1.ObjectReference {
	if env == nil {
		return nil
	}
	return &corev1.ObjectReference{
		APIVersion: APIVersion,
		Kind:       XenvKind,
		Namespace:  env.Namespace,
		Name:       env.Name,
		UID:        env.UID,
	}
}
