package apps

import (
	"github.com/Doout/formation/builder"
	"k8s.io/api/core/v1"
	"reflect"
)

type PodBuilder struct {
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

func (builder *PodBuilder) AddVolumeToContainer(containerName string, containerVolume v1.VolumeMount, volume v1.VolumeSource) *PodBuilder {
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
		container.VolumeMounts = append(container.VolumeMounts, containerVolume)
	}
	return builder
}

func (builder *PodBuilder) GetVolume(name string) *v1.Volume {
	for idx, v := range builder.Spec.Volumes {
		if v.Name == name {
			return &builder.Spec.Volumes[idx]
		}
	}
	return nil
}

func (builder *PodBuilder) GetContainer(name string) *v1.Container {
	return GetContainer(builder.Spec, name)
}

func (builder *PodBuilder) AddEnvToAllContainer(envVar ...v1.EnvVar) *PodBuilder {
	for index := range builder.Spec.Containers {
		ToContainerBuilder(&builder.Spec.Containers[index]).AddEnvironmentVariable(true, envVar...)
	}
	for index := range builder.Spec.InitContainers {
		ToContainerBuilder(&builder.Spec.InitContainers[index]).AddEnvironmentVariable(true, envVar...)
	}
	return builder
}

func (builder *PodBuilder) AddEnvToContainer(containerName string, envVar ...v1.EnvVar) *PodBuilder {
	container := builder.GetContainer(containerName)
	cbuilder := ToContainerBuilder(container)
	if container != nil {
		cbuilder.AddEnvironmentVariable(true, envVar...)
	}
	return builder
}

func (builder *PodBuilder) SetServiceAccount(name string) *PodBuilder {
	builder.Spec.ServiceAccountName = name
	return builder
}

func (builder *PodBuilder) AddContainer(container *v1.Container) *PodBuilder {
	if builder.Spec.Containers == nil {
		builder.Spec.Containers = []v1.Container{*container}
	} else {
		builder.Spec.Containers = append(builder.Spec.Containers, *container)
	}
	return builder
}

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
