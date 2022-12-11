package core

import (
	"context"
	"github.com/Doout/formation/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

func (c *Secret) Update(ctx context.Context, fromApiServer runtime.Object) error {
	// Check if fromApiServer is type Secret
	secret, ok := fromApiServer.(*v1.Secret)
	if !ok {
		return types.ErrWrongResourceType
	}

	// Check if the secret is immutable using annotations
	if c.secret.Annotations[types.UpdateKey] == "disabled" || secret.Annotations[types.UpdateKey] == "disabled" {
		return nil
	}
	//Check if the secret is immutable using type
	if secret.Immutable != nil && *secret.Immutable {
		return nil
	}
	secret.Data = c.secret.Data
	return nil
}
