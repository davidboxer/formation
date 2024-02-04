package types

import (
	"context"
	v1 "k8s.io/api/core/v1"
	v11 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resource interface {
	//The type of Resource, Ex configmap
	Type() string
	//The name of the Resource, this name need to be unique per namespace
	Name() string
	//Return an emtpy runtime object of type Resource
	Runtime() client.Object

	//Create the Resource structure
	Create() (client.Object, error)
}

// Builder interfaces

// ToResource is the interface that converts a Builder to a Resource
type ToResource interface {
	ToResource() Resource
}

// AddVolumeToContainer is the interface that adds a Volume to a Container
type AddVolumeToContainer interface {
	AddVolumeToContainer(containerName string, containerVolume v1.VolumeMount, volume v1.VolumeSource)
}

type AddEnvFromSourceToContainer interface {
	AddEnvFromSourceToContainer(containerName string, envFromSource v1.EnvFromSource)
}

// AddTemplateVolumeToContainer is the interface that adds a Volume Template to the builder and which container it belongs to
type AddTemplateVolumeToContainer interface {
	AddVolumeToContainer(containerName string, containerVolume v1.VolumeMount, template v1.PersistentVolumeClaimSpec)
}

// ResourcesName is the interface that returns the name of the resources that the builder creates
// For example, if the resource is a Deployment,
// it will return the with a list of container name with the format <Pod-Name>/<Container-Name>
type ResourcesName interface {
	ResourcesName() []string
}

// Depending on how the controller is implemented, the following interfaces might be useful.
// An Example, if the controller is watching an CR that allow user to customize the resource,
// the following interfaces will be useful for all builder types to implement.

// ConfigurableContainer is the interface that allows the user to customize the resource
type ConfigurableContainer interface {
	// AddResourceRequirements Add resource requirements to the container
	AddResourceRequirements(containerName string, requirements v1.ResourceRequirements)

	//AddEnv Add environment variables to the container
	AddEnv(containerName string, env ...v1.EnvVar)

	SetImage(containerName string, image string)

	SetImagePullPolicy(containerName string, policy v1.PullPolicy)

	// SetStartupProbe Set the startup probe for the container
	SetStartupProbe(containerName string, probe v1.Probe)

	// SetReadinessProbe Set the readiness probe for the container
	SetReadinessProbe(containerName string, probe v1.Probe)

	// SetLivenessProbe Set the liveness probe for the container
	SetLivenessProbe(containerName string, probe v1.Probe)
}

// ConfigurablePod is the interface that allows the user to customize the Pod
type ConfigurablePod interface {
	// AddAffinity Add affinity to the pod
	AddAffinity(affinity v1.Affinity)
	// AddTolerations Add tolerations to the pod
	AddTolerations(toleration ...v1.Toleration)
	// AddTopologySpreadConstraints Add Pod Topology Spread Constraints
	AddTopologySpreadConstraints(topologySpreadConstraint ...v1.TopologySpreadConstraint)

	// AddNodeSelector Add node selector to the pod
	AddNodeSelector(name, value string)

	AddImagePullSecrets(secretNames ...string)

	SetServiceAccountName(name string)
}

// ConfigurableReplicas is the interface that allows the user to customize the number of replicas
type ConfigurableReplicas interface {
	SetReplicas(replica int32)
}

// These interfaces are desgin to give each resource more flexibility to implement their own logic, they are only design to be use with the build in controller

// Converged The default implementation for a list of resources once it is created is to move on to the next one,
// if the controller need to wait for this resource to be ready, it can implement this interface
type Converged interface {
	Converged(ctx context.Context, client client.Client, namespace string) (bool, error)
}

// ConvergedGroup group of resources that need to be converged together but not necessarily in order
type ConvergedGroupInterface interface {
	// SetConvergedGroupID set the group id for the resource. ID must be greater than 0
	// ID of 0 means the resource is not part of any group and must be converged before the next resource will be checked
	SetConvergedGroupID(uid int)
	// GetConvergedGroupID get the group id for the resource
	GetConvergedGroupID() int
}

// Update is the interface that allows each resource to implement their own update logic.
// The default behaviour for the build-in controller is to merge the new into the old.
// To get more control of the resource lifecycle, the controller can implement Reconcile
// Optional
type Update interface {
	Update(ctx context.Context, fromApiServer runtime.Object) error
}

// Reconcile If the resource need to implement their own reconcile logic, they can implement this interface
// Optional
type Reconcile interface {
	Reconcile(ctx context.Context, client client.Client, owner v11.Object) (bool, error)
}
