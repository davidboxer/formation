package types

import (
	"errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ErrWrongResourceType = errors.New("wrong resource type")
)

const (
	//HashKey conation the hash of the resource in the annotations
	HashKey = "formation/hash"

	//UpdateKey if is set to "disabled", the resource will not be updated.
	//Once this is set on a resource, it will not be updated unless the annotation is removed.
	UpdateKey = "formation/update"
)

type ResourceState string

const (
	Creating ResourceState = "Creating"
	Ready    ResourceState = "Ready"
	Waiting  ResourceState = "Waiting"
)

// StorageConfigType manages the formation's interpretation of a StorageConfig; each
// of the types has near-identical specifications but slightly different
// behavior.
// +kubebuilder:validation:Enum=template;existing;create;""
type StorageConfigType string

const (
	// StorageConfigTypeTemplate will typically be templated in to any
	// StatefulSet managed by the ResourceConfig parent of this StorageConfig.
	// They will be ignored for any workload object (e.g., Deployment) that
	// doesn't accept VolumeClaimTemplates.
	StorageConfigTypeTemplate StorageConfigType = "template"
	// StorageConfigTypeExisting will typically be inserted in to the Volumes
	// field of the PodSpec for all workload objects (Deployment, StatefulSet)
	// managed by this ResourceConfig.
	StorageConfigTypeExisting StorageConfigType = "existing"
	// StorageConfigTypeCreate are expected to be created by the controllers and then
	// inserted in to the Volumes field of workload objects, just as with
	// Existing.
	StorageConfigTypeCreate StorageConfigType = "create"
)

type LinkVolumeData struct {
	//The name of the container that this volume is mounted to
	// Format is <PodName>.<ContainerName>
	// Wildcard is supported, e.g. <PodName>.* will mount to all containers in the pod
	// Regex is not supported
	Visibility []string `json:"visibility,omitempty" yaml:"visibility"`

	VolumeMount v1.VolumeMount `json:"volumeMount,omitempty" yaml:"volumeMount"`

	//If the Volume is pre-created, or was created by the Formation, what is the reference to it
	VolumeSource *v1.VolumeSource `json:"volumeSource,omitempty" yaml:"volumeSource"`

	// If Template is set, the volume will be created by the template
	Template *v1.PersistentVolumeClaimSpec `json:"template,omitempty" yaml:"template"`

	// The sources to populate environment variables in the container.
	// The keys defined within a source must be a C_IDENTIFIER. All invalid keys
	// will be reported as an event when the container is starting. When a key exists in multiple
	// sources, the value associated with the last source will take precedence.
	// Values defined by an Env with a duplicate key will take precedence.
	EnvFromSource *v1.EnvFromSource `json:"envFromSource,omitempty" yaml:"envFromSource"`
}

type FormationStatusInterface interface {
	GetStatus() *FormationStatus
}
type ResourceStatus struct {
	Name  string        `json:"name,omitempty"`
	Type  string        `json:"type,omitempty"`
	Group string        `json:"group,omitempty"`
	State ResourceState `json:"state,omitempty"`
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format="date-time"
	LastUpdate metav1.Time `json:"lastUpdate,omitempty"`
}

type FormationStatus struct {
	Resources []*ResourceStatus `json:"resources,omitempty" yaml:"resources"`
}

func (in *FormationStatus) DeepCopyInto(t *FormationStatus) {
	t.Resources = make([]*ResourceStatus, 0, len(in.Resources))
	for _, res := range in.Resources {
		if res == nil {
			continue
		}
		t.Resources = append(t.Resources,
			&ResourceStatus{Name: res.Name, Group: res.Group, Type: res.Type, State: res.State})
	}
}

func (in *FormationStatus) DeepCopy() *FormationStatus {
	if in == nil {
		return nil
	}
	out := new(FormationStatus)
	in.DeepCopyInto(out)
	return out
}

type ConvergedGroup struct {
	id int
}

func (c *ConvergedGroup) SetConvergedGroupID(id int) {
	c.id = id
}

func (c *ConvergedGroup) GetConvergedGroupID() int {
	if c == nil {
		return 0
	}
	if c.id <= 0 {
		return 0
	}
	return c.id
}
