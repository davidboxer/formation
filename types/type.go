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
	Creating ResourceState = "creating"
	Ready    ResourceState = "ready"
	Waiting  ResourceState = "waiting"
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
	Visibility []string

	VolumeMount v1.VolumeMount

	//If the Volume is pre-created, or was created by the Formation, what is the reference to it
	VolumeSource v1.VolumeSource

	// If Template is set, the volume will be created by the template
	Template *v1.PersistentVolumeClaimSpec
}

type AddVolumeToContainer interface {
	AddVolumeToContainer(containerName string, containerVolume v1.VolumeMount, volume v1.VolumeSource)
}

type AddTemplateVolumeToContainer interface {
	AddVolumeToContainer(containerName string, containerVolume v1.VolumeMount, template v1.PersistentVolumeClaimSpec)
}

type ResourcesName interface {
	// ResourcesName returns a list of names of the resources.
	// For example, if the resource is a Deployment,
	//it will return the with a list of container name with the format <Pod-Name>/<Container-Name>
	ResourcesName() []string
}

type FormationStatusInterface interface {
	GetStatus() *FormationStatus
}
type ResourceStatus struct {
	Name  string        `json:"name,omitempty"`
	Type  string        `json:"type,omitempty"`
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
		t.Resources = append(t.Resources, &ResourceStatus{Name: res.Name, Type: res.Type, State: res.State})
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
