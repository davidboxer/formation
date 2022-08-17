package types

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Converged interface {
	Converged(ctx context.Context, client client.Client, namespace string) (bool, error)
}

type Update interface {
	Update(ctx context.Context, fromApiServer runtime.Object) error
}

type Resource interface {
	//The type of Resource, Ex configmap
	Type() string
	//The name of the Resource, this name need to be unique per namespace
	Name() string
	//Return an emtpy runtime object of type Resource
	Runtime() client.Object

	//Create the Resource structure
	Create() (client.Object, error)
}
