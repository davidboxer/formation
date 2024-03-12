package apps

import (
	"strings"

	"github.com/davidboxer/formation/types"
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
						// Check if the Volume is an EnvFromSource
						if volume.EnvFromSource != nil {
							// check if object have type.AddEnvFromSource
							if addEnvFromSource, ok := obj.(types.AddEnvFromSourceToContainer); ok {
								// Add the envFromSource
								addEnvFromSource.AddEnvFromSourceToContainer(resContainerName, *volume.EnvFromSource)
							}

						} else if volume.Template != nil { // Check if the Volume is an Template
							// check if object have type.AddTemplateVolume
							if addTemplateVolume, ok := obj.(types.AddTemplateVolumeToContainer); ok {
								// Add the template volume
								addTemplateVolume.AddVolumeToContainer(resContainerName, volume.VolumeMount, *volume.Template)
							}
						} else if volume.VolumeSource != nil { // Check if the Volume is an VolumeSource
							// check if object have type.AddVolume
							if addVolume, ok := obj.(types.AddVolumeToContainer); ok {
								// Add the volume
								addVolume.AddVolumeToContainer(resContainerName, volume.VolumeMount, *volume.VolumeSource)
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
		// If the character before the suffix is a dash, then it is a match.
		// If the podName == visibility, then it is a match as well.
		if len(podName) == len(visibility) || podName[len(podName)-len(visibility)-1] == '-' {
			return true
		}
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

func FindAllPodBuilderWithContainerName(builders []*PodBuilder, containerName string) []*PodBuilder {
	retList := []*PodBuilder{}
	for index, builder := range builders {
		if container := builder.GetContainer(containerName); container != nil {
			retList = append(retList, builders[index])
		}
	}
	return retList
}
