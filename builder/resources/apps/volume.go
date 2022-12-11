package apps

import (
	"github.com/Doout/formation/types"
	"strings"
)

func LinkVolumes(objects []any, volumes []types.LinkVolumeData) {
	// Reduce the list of object to only types.ResourcesName
	var objs []any
	for _, obj := range objects {
		_, ok := obj.(types.ResourcesName)
		if ok {
			objs = append(objs, obj)
		}
	}

	//Loop over all the volumes
	for _, volume := range volumes {
		for _, obj := range objs {
			// get the list of resourcesName
			resourcesName := obj.(types.ResourcesName)
			// Check if any of the resourcesName match the visibility
			for _, visibility := range volume.Visibility {
				// split the visibility into podName and containerName
				podName, containerName := splitVisibility(visibility)
				// Look over resourcesName to see if any of them match the visibility
				for _, name := range resourcesName.ResourcesName() {
					// split the name into podName and containerName
					resPodName, resContainerName := splitVisibility(name)
					// Check if the podName and containerName match the visibility
					if podNameMatchVisibility(resPodName, podName) && containerNameMatchVisibility(resContainerName, containerName) {
						// Check if the Volume is an Template
						if volume.Template != nil {
							// check if object have type.AddTemplateVolume
							if addTemplateVolume, ok := obj.(types.AddTemplateVolumeToContainer); ok {
								// Add the template volume
								addTemplateVolume.AddVolumeToContainer(resContainerName, volume.VolumeMount, *volume.Template)
							}
						} else {
							// check if object have type.AddVolume
							if addVolume, ok := obj.(types.AddVolumeToContainer); ok {
								// Add the volume
								addVolume.AddVolumeToContainer(resContainerName, volume.VolumeMount, volume.VolumeSource)
							}
						}
					}
				}
			}
		}
	}
}

func splitVisibility(visibility string) (string, string) {
	podName := visibility
	containerName := ""
	if idx := strings.Index(visibility, "/"); idx != -1 {
		podName = visibility[:idx]
		containerName = visibility[idx+1:]
	}
	return podName, containerName
}

func podNameMatchVisibility(podName string, visibility string) bool {
	if visibility == "*" {
		return true
	}
	// Since the podName can also include the instance prefix. Check if the visibility match the podName end
	if strings.HasSuffix(podName, visibility) {
		return true
	}
	return false
}
func containerNameMatchVisibility(containerName string, visibility string) bool {
	if visibility == "*" {
		return true
	}
	if visibility == containerName {
		return true
	}
	return false
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
