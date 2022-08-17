package utils

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type PointerType interface {
	int | int8 | int16 | int32 | int64 | float32 | float64 | string | bool
}

func Pointer[T PointerType](b T) *T {
	return &b
}

// CreateResourceList the first element of the slice is cpu, second is memory and third is ephemeral storage
func CreateResourceList(resourceLimits ...string) v1.ResourceList {
	res := v1.ResourceList{}
	if len(resourceLimits) > 0 {
		res["cpu"] = resource.MustParse(resourceLimits[0])
	}
	if len(resourceLimits) > 1 {
		res["memory"] = resource.MustParse(resourceLimits[1])
	}
	if len(resourceLimits) > 2 {
		res["ephemeral-storage"] = resource.MustParse(resourceLimits[2])
	}
	return res
}

func MergeResourceRequirements(dst, src *v1.ResourceRequirements) (*v1.ResourceRequirements, bool) {
	if dst == nil {
		dst = &v1.ResourceRequirements{}
	}
	if src == nil {
		src = &v1.ResourceRequirements{}
	}
	copy := dst.DeepCopy()
	updateValue := false
	//Apply all the value in src onto dst
	for key, value := range src.Limits {
		v2, ok := copy.Limits[key]
		if !ok || v2.Cmp(value) != 0 {
			updateValue = true
			copy.Limits[key] = value
			continue
		}
	}

	for key, value := range src.Requests {
		v2, ok := copy.Requests[key]
		if !ok || v2.Cmp(value) != 0 {
			updateValue = true
			copy.Requests[key] = value
			continue
		}
	}

	return copy, updateValue
}
