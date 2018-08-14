package v1

import (
	"fmt"
	"path/filepath"
	reflect "reflect"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// CRD names for Funcdef
const (
	FuncinstKind       = "Funcinst"
	FuncinstPluralName = "funcinsts"
)

// static asserts
var (
	_ runtime.Object            = (*Funcinst)(nil)
	_ metav1.ObjectMetaAccessor = (*Funcinst)(nil)

	_ runtime.Object          = (*FuncinstList)(nil)
	_ metav1.ListMetaAccessor = (*FuncinstList)(nil)
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Funcinst is a API object to represent a FUNCtion DEClaration
type Funcinst struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   FuncinstSpec   `json:"spec"`
	Status FuncinstStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FuncinstList is a API object to represent a list of Refuncs
type FuncinstList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Funcinst `json:"items"`
}

// FuncinstSpec is the specification that describes a funcinst for refunc
type FuncinstSpec struct {
	FuncdefRef *corev1.ObjectReference `json:"funcdefRef,omitempty"`
	TriggerRef *corev1.ObjectReference `json:"triggerRef,omitempty"`

	Runtime RuntimeConfig

	// LastActivity is the last valid active for current group of func insts
	// The operator will refresh this field periodically to keep func alive
	LastActivity metav1.Time `json:"lastActivity"`
}

// FuncinstPhase is label to indicates current state for a func
type FuncinstPhase string

// Different phases during life time of a funcinst
//
//	Funcinst's states transition
//
//	         funcdef found                 xenv is ok (should be really fast)
//	( o ) ╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌> │ pending │ ╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌> │ ready │ ╌╌╌╌┐
//	                              ┆  ^                               ┆         ┆
//	    ( x )             funcdef ┆  ┆        xenv removed           v         ┆
//	      ^               removed v  ┣╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌<╌╌╌╌╌╌╌╌┐  ┆         ┆
//	      ┆                       ┆  ┆                            ┆╌╌┫      rs ┆
//	 │ inactive │ <╌╌╌╌╌╌┳╌╌╌<╌╌╌╌┻╌╌(╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌<╌╌╌╌╌╌╌╌┘  ┆ created ┆
//	                     ┆           ┆     funcdef removed           ┆         ┆
//	     hash changed or ^      xenv ┆                               ^         ┆
//	     funcdef removed ┆   removed ^                               ┆         ┆
//	                ┌╌<╌╌┻╌╌<╌┳╌╌>╌╌╌┘                         ┌╌╌╌╌╌┘         ┆
//	        tapping ┆         ┆       active pods > 0    ┌╌╌╌╌╌v╌╌╌╌╌╌┐        ┆
//	                └╌╌> │ active │ <╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌ ┆ activating ┆ <╌╌╌╌╌╌┘
//                            ┆  active pods drop to 0   └╌╌╌╌╌^╌╌╌╌╌╌┘
//                            └╌╌╌╌╌>╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┘
const (
	Inactive FuncinstPhase = "Inactive" // funcinst cannot accept new events
	Pending  FuncinstPhase = "Pending"  // waiting for a valid xenv is ready
	Ready    FuncinstPhase = "Ready"    // runner is ready, refunc is idle that can be activated
	Active   FuncinstPhase = "Active"   // can be invoked
)

// BackendRef contains a valid pod's IP and its ref
type BackendRef struct {
	IP     string                  `json:"string"`
	PodRef *corev1.ObjectReference `json:"podRef,omitempty"`
}

// FuncinstStatus is the running status for a refunc
type FuncinstStatus struct {
	Phase FuncinstPhase `json:"phase"` // current refunc state

	Backends []BackendRef `json:"backends,omitempty"`
}

// RuntimeConfig is runtime configuration for funcinst
type RuntimeConfig struct {
	Credentials Credentials `json:"credentials"`
	Permissions Permissions `json:"permissions"`
}

// Credentials provides runtime credentials for funcinst
type Credentials struct {
	AccessKey string `json:"accessKey,omitempty"`
	SecretKey string `json:"secretKey,omitempty"`
	Token     string `json:"token,omitempty"`
}

// Permissions provides runtime permissions for funcinst
type Permissions struct {
	Scope     string   `json:"scope,omitempty"`
	Publish   []string `json:"publish,omitempty"`
	Subscribe []string `json:"subscribe,omitempty"`
}

// AsOwner returns *metav1.OwnerReference
func (t *Funcinst) AsOwner() *metav1.OwnerReference {
	return &metav1.OwnerReference{
		APIVersion: APIVersion,
		Kind:       FuncinstKind,
		Name:       t.Name,
		UID:        t.UID,
		Controller: &trueVar,
	}
}

// Ref returns *corev1.ObjectReference
func (t *Funcinst) Ref() *corev1.ObjectReference {
	if t == nil {
		return nil
	}
	return &corev1.ObjectReference{
		APIVersion: APIVersion,
		Kind:       FuncinstKind,
		Namespace:  t.Namespace,
		Name:       t.Name,
		UID:        t.UID,
	}
}

// SortBackends sorts backends by IP
func SortBackends(backends []BackendRef) {
	sort.Slice(backends, func(i int, j int) bool {
		bi, bj := backends[i], backends[j]
		if bi.PodRef != nil && bj.PodRef != nil {
			return bi.PodRef.Name < bj.PodRef.Name
		}
		return bi.IP < bj.IP
	})
}

// AppendBackend appends or updates backends
func AppendBackend(backends []BackendRef, pod *corev1.Pod) []BackendRef {
	ref := &corev1.ObjectReference{
		Kind:       pod.Kind,
		Namespace:  pod.Namespace,
		Name:       pod.Name,
		UID:        pod.UID,
		APIVersion: pod.APIVersion,
	}

	// duplicates checking
	for i, b := range backends {
		if reflect.DeepEqual(b.PodRef, ref) {
			return backends
		}
		// replace old
		if b.IP == pod.Status.PodIP {
			backends[i].PodRef = ref
		}
	}

	return append(backends, BackendRef{
		IP:     pod.Status.PodIP,
		PodRef: ref,
	})
}

// OnlyLastActivityChanged checks two funcinst returns true if only LastActivity is changed
func OnlyLastActivityChanged(left, right *Funcinst) bool {
	l, r := left.DeepCopy(), right.DeepCopy()
	l.TypeMeta = r.TypeMeta
	l.ObjectMeta = r.ObjectMeta
	l.Spec.LastActivity = r.Spec.LastActivity
	return reflect.DeepEqual(l, r)
}

// NewDefaultPermissions returns default permissions for given inst
func NewDefaultPermissions(fni *Funcinst, scopeRoot string, isSys bool) Permissions {
	if isSys {
		return Permissions{
			Scope: scopeRoot,
			Publish: []string{
				"refunc.>",
				"_refunc.>",
				"_INBOX.*",
				"_INBOX.*.*",
			},
			Subscribe: []string{
				"refunc.>",
				"_refunc.>",
				"_INBOX.*",
				"_INBOX.*.*",
			},
		}
	}
	return Permissions{
		Scope: filepath.Join(scopeRoot, fni.Spec.FuncdefRef.Namespace, fni.Spec.FuncdefRef.Name, "data") + "/",
		Publish: []string{
			// request endpoint
			"refunc.*.*",
			"refunc.*.*._meta",
			// reply
			"_INBOX.*",   // old style
			"_INBOX.*.*", // new style
			// logs forwarding endpoint to client
			"_refunc.forwardlogs.*",
			fni.EventsPubEndpoint(),
			fni.LoggingEndpoint(),
			fni.CryingEndpoint(),
			fni.TappingEndpoint(),
		},
		Subscribe: []string{
			// public
			"_INBOX.*",   // old style
			"_INBOX.*.*", // new style
			// internal
			fni.EventsSubEndpoint(),
			fni.ServiceEndpoint(),
			fni.CryServiceEndpoint(),
		},
	}
}

// EventsPubEndpoint is endpoint for publishing events
func (t *Funcinst) EventsPubEndpoint() string {
	return fmt.Sprintf("refunc.%s.%s.events.>", t.Spec.FuncdefRef.Namespace, t.Spec.FuncdefRef.Name)
}

// LoggingEndpoint is endpoint for logging
func (t *Funcinst) LoggingEndpoint() string {
	return fmt.Sprintf("refunc.%s.%s.logs.%s", t.Spec.FuncdefRef.Namespace, t.Spec.FuncdefRef.Name, t.Name)
}

// CryingEndpoint is endpoint to signal birth of a inst
func (t *Funcinst) CryingEndpoint() string {
	return fmt.Sprintf("_refunc._cry_.%s/%s", t.Namespace, t.Name)
}

// TappingEndpoint is endpoint for tapping
func (t *Funcinst) TappingEndpoint() string {
	return fmt.Sprintf("_refunc._tap_.%s/%s", t.Namespace, t.Name)
}

// EventsSubEndpoint is endpoint for subscribing events within same ns
func (t *Funcinst) EventsSubEndpoint() string {
	return fmt.Sprintf("refunc.%s.*.events.>", t.Spec.FuncdefRef.Namespace)
}

// ServiceEndpoint is endpoint for inst to listen at in order to proivde services
func (t Funcinst) ServiceEndpoint() string {
	return fmt.Sprintf("_refunc._insts_.%s.%s", t.Spec.FuncdefRef.Namespace, t.Name)
}

// CryServiceEndpoint is endpoint to poke a inst to cry
func (t *Funcinst) CryServiceEndpoint() string {
	return t.ServiceEndpoint() + "._cry_"
}
