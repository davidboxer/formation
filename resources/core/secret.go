package core

import (
	"github.com/Doout/formation/types"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Secret struct {
	secret        *v1.Secret
	DisableUpdate bool
}

func NewSecret(secret *v1.Secret) *Secret {
	return &Secret{secret: secret}
}

func (c *Secret) Type() string           { return "secret" }
func (c *Secret) Name() string           { return c.secret.Name }
func (c *Secret) Runtime() client.Object { return &v1.Secret{} }

func (c *Secret) Create() (client.Object, error) {
	if c.DisableUpdate {
		if c.secret.Annotations == nil {
			c.secret.Annotations = make(map[string]string)
		}
		c.secret.Annotations[types.UpdateKey] = "disabled"
	}
	return c.secret, nil
}
