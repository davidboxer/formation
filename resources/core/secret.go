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
	onCreate      func(*v1.Secret)
	DisableUpdate bool
}

func NewSecret(secret *v1.Secret) *Secret {
	if secret == nil {
		secret = &v1.Secret{}
	}
	return &Secret{secret: secret}
}

// NewSecretWithOnCreate is a helper function to create a Secret with a onCreate function
// The onCreate function is called when the Secret is required to be created or updated
// Due to the nature of the Secret, we do not want to generate the secret unless it is required to be created or updated
func NewSecretWithOnCreate(secret *v1.Secret, onCreate func(*v1.Secret)) *Secret {
	if secret == nil {
		secret = &v1.Secret{}
	}
	return &Secret{secret: secret, onCreate: onCreate}
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
	if c.onCreate != nil {
		c.onCreate(c.secret)
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
	if secret.Annotations[types.UpdateKey] == "disabled" {
		return nil
	}
	// Update metadata
	//update all labels are updated
	for k, v := range c.secret.Labels {
		secret.Labels[k] = v
	}
	//Update all annotations are updated
	for k, v := range c.secret.Annotations {
		secret.Annotations[k] = v
	}

	//Check if the secret is immutable using type
	if secret.Immutable != nil && *secret.Immutable {
		return nil
	}
	secret.Data = c.secret.Data
	return nil
}
