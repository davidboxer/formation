package core

import (
	"context"

	"github.com/davidboxer/formation/resources/common"
	"github.com/davidboxer/formation/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Secret struct {
	*common.SimpleResource[*v1.Secret]
}

func NewSecret(secret *v1.Secret) *Secret {
	if secret == nil {
		secret = &v1.Secret{}
	}
	return &Secret{
		SimpleResource: common.NewSimpleResource("secret", secret),
	}
}

// NewSecretWithOnCreate is a helper function to create a Secret with a onCreate function
// The onCreate function is called when the Secret is required to be created or updated
// Due to the nature of the Secret, we do not want to generate the secret unless it is required to be created or updated
func NewSecretWithOnCreate(secret *v1.Secret, onCreate func(*v1.Secret)) *Secret {
	if secret == nil {
		secret = &v1.Secret{}
	}
	return &Secret{
		SimpleResource: common.NewSimpleResourceWithOnCreate("secret", secret, onCreate),
	}
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
	// Update all the metadata
	for k, v := range c.Obj.Labels {
		secret.Labels[k] = v
	}
	//Update all annotations are updated
	for k, v := range c.Obj.Annotations {
		secret.Annotations[k] = v
	}

	//Check if the secret is immutable using type
	if secret.Immutable != nil && *secret.Immutable {
		return nil
	}
	secret.Data = c.Obj.Data
	return nil
}
