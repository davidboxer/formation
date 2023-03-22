package core

import (
	"github.com/Doout/formation/resources/common"
	"k8s.io/api/core/v1"
)

type ConfigMap struct {
	*common.SimpleResource[*v1.ConfigMap]
}

func NewConfigMap(configMap *v1.ConfigMap) *ConfigMap {
	return &ConfigMap{
		SimpleResource: common.NewSimpleResource("configmap", configMap),
	}
}

func NewConfigMapWithOnCreate(configMap *v1.ConfigMap, onCreate func(*v1.ConfigMap)) *ConfigMap {
	return &ConfigMap{
		SimpleResource: common.NewSimpleResourceWithOnCreate("configmap", configMap, onCreate),
	}
}
