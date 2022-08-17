package apps

import (
	"github.com/Doout/formation/types"
	v1 "k8s.io/api/core/v1"
)

type Volume struct {
	Type types.StorageConfigType
	//The name of the container that this volume is mounted to
	Visibility   []string
	VolumeMount  v1.VolumeMount
	VolumeSource v1.VolumeSource
}

//Type of link volume
// 1) Insert PVC into Volume
// 2) Insert Configmap/Secret as Volume

func LinkVolumes(builder []any, volumes []Volume) {
	podBuilders := FindAllPodBuilder(builder)
	for _, vol := range volumes {
		if vol.Type == types.StorageConfigTypeTemplate {
			//TODO add support for template
			continue
		}
		for _, containerName := range vol.Visibility {
			for _, builder := range FindAllPodBuilderWithContainerName(podBuilders, containerName) {
				builder.AddVolumeToContainer(containerName, vol.VolumeMount, vol.VolumeSource)
			}
		}
	}
}

func FindAllPodBuilder(builder []any) []*PodBuilder {
	retList := []*PodBuilder{}
	for _, v := range builder {
		switch v.(type) {
		case *DeploymentBuilder:
			retList = append(retList, v.(*DeploymentBuilder).PodBuilder)
		}
	}
	return retList
}

func FindAllPodBuilderWithContainerName(builders []*PodBuilder, containerName string) []*PodBuilder {
	retList := []*PodBuilder{}
	for index, builder := range builders {
		if container := builder.GetContainer(containerName); container != nil {
			retList = append(retList, builders[index])
		}
	}
	return retList
}
