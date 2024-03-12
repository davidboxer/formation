package core

import (
	"github.com/davidboxer/formation/resources/common"
	v1 "k8s.io/api/core/v1"
)

type PersistentVolumeClaim struct {
	*common.SimpleResource[*v1.PersistentVolumeClaim]
}

func NewPersistentVolumeClaim(pvc *v1.PersistentVolumeClaim) *PersistentVolumeClaim {
	return &PersistentVolumeClaim{
		SimpleResource: common.NewSimpleResource("persistentvolumeclaim", pvc),
	}
}
