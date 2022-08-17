package core

import (
	"github.com/Doout/formation/types"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Configmap struct {
	cm            *v1.ConfigMap
	DisableUpdate bool
}

func NewConfigmap(cm *v1.ConfigMap) *Configmap {
	return &Configmap{cm: cm}
}

func (c *Configmap) Type() string { return "configmap" }

func (c *Configmap) Name() string { return c.cm.Name }

func (c *Configmap) Runtime() client.Object { return &v1.ConfigMap{} }

func (c *Configmap) Create() (client.Object, error) {
	if c.DisableUpdate {
		if c.cm.Annotations == nil {
			c.cm.Annotations = make(map[string]string)
		}
		c.cm.Annotations[types.UpdateKey] = "disabled"
	}
	return c.cm, nil

}
