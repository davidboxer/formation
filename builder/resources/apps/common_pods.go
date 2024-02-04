package apps

import (
	"github.com/Doout/formation/builder"
	"github.com/Doout/formation/types"
	"github.com/Doout/formation/utils"
	"k8s.io/api/core/v1"
	"reflect"
)

type PodBuilder struct {
	*types.ConvergedGroup
	builder.Builder
	Spec *v1.PodSpec
}
type VolumeTemplateBuilder interface {
	HandleTemplate(containerName string, containerVolume v1.VolumeMount, pvc v1.PersistentVolumeClaim)
}

var (
	ImagePullPolicy = v1.PullIfNotPresent
)

func (builder *PodBuilder) SetRestartPolicy(restart v1.RestartPolicy) *PodBuilder {
	builder.Spec.RestartPolicy = restart
	return builder
}

// AddEnvFromSourceToContainer Add environment variables from a source to the container
func (builder *PodBuilder) AddEnvFromSourceToContainer(containerName string, envFromSource v1.EnvFromSource) {
	container := builder.GetContainer(containerName)
	if container != nil {
		if container.EnvFrom == nil {
			container.EnvFrom = []v1.EnvFromSource{}
		}
		container.EnvFrom = append(container.EnvFrom, envFromSource)
	}
}

// AddVolumeToContainer Add a volume to the container
func (builder *PodBuilder) AddVolumeToContainer(containerName string, containerVolume v1.VolumeMount, volume v1.VolumeSource) {
	container := builder.GetContainer(containerName)
	if container != nil {
		if container.VolumeMounts == nil {
			container.VolumeMounts = []v1.VolumeMount{}
		}
		//Check if we already add VolumeSource into spec.Volumes. We are only allow to add it once
		skipPodVolume := false
		for _, item := range builder.Spec.Volumes {
			if reflect.DeepEqual(item.VolumeSource, volume) {
				containerVolume.Name = item.Name
				skipPodVolume = true
				break
			}
		}

		newVolume := v1.Volume{Name: containerVolume.Name, VolumeSource: volume}
		if builder.Spec.Volumes == nil {
			builder.Spec.Volumes = []v1.Volume{}
		}
		if !skipPodVolume {
			builder.Spec.Volumes = append(builder.Spec.Volumes, newVolume)
		}
		//Check if this volume already exist in container.VolumeMounts
		for _, item := range container.VolumeMounts {
			if item.Name == containerVolume.Name {
				return
			}
		}
		container.VolumeMounts = append(container.VolumeMounts, containerVolume)
	}
}

// GetVolume Get a volume from the pod that match the name
func (builder *PodBuilder) GetVolume(name string) *v1.Volume {
	for idx, v := range builder.Spec.Volumes {
		if v.Name == name {
			return &builder.Spec.Volumes[idx]
		}
	}
	return nil
}

// GetContainer Get a container from the pod that match the name
func (builder *PodBuilder) GetContainer(name string) *v1.Container {
	return GetContainer(builder.Spec, name)
}

// AddEnvToAllContainer Add environment variables to all containers
func (builder *PodBuilder) AddEnvToAllContainer(envVar ...v1.EnvVar) *PodBuilder {
	for index := range builder.Spec.Containers {
		ToContainerBuilder(&builder.Spec.Containers[index]).AddEnvironmentVariable(true, envVar...)
	}
	for index := range builder.Spec.InitContainers {
		ToContainerBuilder(&builder.Spec.InitContainers[index]).AddEnvironmentVariable(true, envVar...)
	}
	return builder
}

// AddEnvToContainer Add environment variables to the container
func (builder *PodBuilder) AddEnvToContainer(containerName string, envVar ...v1.EnvVar) *PodBuilder {
	container := builder.GetContainer(containerName)
	cbuilder := ToContainerBuilder(container)
	if container != nil {
		cbuilder.AddEnvironmentVariable(true, envVar...)
	}
	return builder
}

// SetServiceAccount Set the service account for the pod
func (builder *PodBuilder) SetServiceAccount(name string) *PodBuilder {
	builder.Spec.ServiceAccountName = name
	return builder
}

// AddContainer Add a container to the pod
func (builder *PodBuilder) AddContainer(container *v1.Container) *PodBuilder {
	if builder.Spec.Containers == nil {
		builder.Spec.Containers = []v1.Container{*container}
	} else {
		builder.Spec.Containers = append(builder.Spec.Containers, *container)
	}
	return builder
}

// AddInitContainer Add a init container to the pod
func (builder *PodBuilder) AddInitContainer(container *v1.Container) *PodBuilder {
	syncContainer(container)
	if builder.Spec.InitContainers == nil {
		builder.Spec.InitContainers = []v1.Container{*container}
	} else {
		builder.Spec.InitContainers = append(builder.Spec.InitContainers, *container)
	}
	return builder
}

func syncContainer(container *v1.Container) {
	if container.ImagePullPolicy == "" {
		container.ImagePullPolicy = ImagePullPolicy
	}
}

// GetContainer Get a container from the pod that match the name
func GetContainer(spec *v1.PodSpec, name string) *v1.Container {
	list := spec.Containers
	for index, item := range list {
		if item.Name == name {
			return &spec.Containers[index]
		}
	}
	list = spec.InitContainers
	for index, item := range list {
		if item.Name == name {
			return &spec.InitContainers[index]
		}
	}
	return nil
}

// ResourcesName returns a list of name of the resources
// The format is <podName>/<containerName>
func (builder *PodBuilder) ResourcesName() []string {
	var names []string
	for _, container := range builder.Spec.Containers {
		names = append(names, builder.Builder.Name+"/"+container.Name)
	}
	return names
}

// AddResourceRequirements add resource requirements to the container
func (builder *PodBuilder) AddResourceRequirements(containerName string, resourceRequirements v1.ResourceRequirements) {
	container := builder.GetContainer(containerName)
	if container != nil {
		container.Resources = resourceRequirements
	}
}

// AddEnv add environment variables to the container
func (builder *PodBuilder) AddEnv(containerName string, envs ...v1.EnvVar) {
	container := builder.GetContainer(containerName)
	if container != nil {
		ToContainerBuilder(container).AddEnvironmentVariable(true, envs...)
	}
}

// AddAffinity add affinity to the pod
func (builder *PodBuilder) AddAffinity(affinity v1.Affinity) {
	builder.Spec.Affinity = utils.MergeAffinity(*builder.Spec.Affinity, affinity)

}

// AddTolerations add tolerations to the pod
func (builder *PodBuilder) AddTolerations(tolerations ...v1.Toleration) {
	builder.Spec.Tolerations = utils.MergeTolerations(builder.Spec.Tolerations, tolerations)
}

// AddTopologySpreadConstraints add topology spread constraints to the pod
func (builder *PodBuilder) AddTopologySpreadConstraints(constraints ...v1.TopologySpreadConstraint) {
	builder.Spec.TopologySpreadConstraints = utils.MergeTopologySpreadConstraints(builder.Spec.TopologySpreadConstraints, constraints)
}

// SetImage set the image of the container
func (builder *PodBuilder) SetImage(containerName string, image string) {
	container := builder.GetContainer(containerName)
	if container != nil {
		container.Image = image
	}
}

// SetImagePullPolicy set the image pull policy of the container
func (builder *PodBuilder) SetImagePullPolicy(containerName string, policy v1.PullPolicy) {
	container := builder.GetContainer(containerName)
	if container != nil {
		container.ImagePullPolicy = policy
	}
}

// ImagePullSecrets add image pull secrets to the pod
func (builder *PodBuilder) AddImagePullSecrets(secretNames ...string) {
	secretReference := []v1.LocalObjectReference{}
	for _, secretName := range secretNames {
		secretReference = append(secretReference, v1.LocalObjectReference{Name: secretName})
	}
	builder.Spec.ImagePullSecrets = utils.MergeLocalObjectReference(builder.Spec.ImagePullSecrets, secretReference)
}

// SetStartupProbe set the startup probe of the container
func (builder *PodBuilder) SetStartupProbe(containerName string, probe v1.Probe) {
	container := builder.GetContainer(containerName)
	if container != nil {
		container.StartupProbe = &probe
	}
}

// SetLivenessProbe set the liveness probe of the container
func (builder *PodBuilder) SetLivenessProbe(containerName string, probe v1.Probe) {
	container := builder.GetContainer(containerName)
	if container != nil {
		container.LivenessProbe = &probe
	}
}

// SetReadinessProbe set the readiness probe of the container
func (builder *PodBuilder) SetReadinessProbe(containerName string, probe v1.Probe) {
	container := builder.GetContainer(containerName)
	if container != nil {
		container.ReadinessProbe = &probe
	}
}

// SetServiceAccountName set the service account of the pod
func (builder *PodBuilder) SetServiceAccountName(serviceAccountName string) {
	builder.Spec.ServiceAccountName = serviceAccountName
}

// AddNodeSelector add node selector to the pod
func (builder *PodBuilder) AddNodeSelector(name string, value string) {
	if builder.Spec.NodeSelector == nil {
		builder.Spec.NodeSelector = map[string]string{}
	}
	builder.Spec.NodeSelector[name] = value
}
