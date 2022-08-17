package builder

import (
	"github.com/Doout/formation/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//Builder is a useful wrapper around core resources in Kubernetes.
//It provides a simple way to add values
type Builder struct {
	Object client.Object
}

type BuilderInterface interface {
	CreateResource() types.Resource
}

func NewBuilder(object client.Object) *Builder {
	return &Builder{Object: object}
}

func (b *Builder) Labels() MapBuilder {
	if b.Object.GetLabels() == nil {
		b.Object.SetLabels(make(map[string]string))
	}
	return b.Object.GetLabels()
}

func (b *Builder) Annotations() MapBuilder {
	if b.Object.GetAnnotations() == nil {
		b.Object.SetAnnotations(make(map[string]string))
	}
	return b.Object.GetAnnotations()
}

type MapBuilder map[string]string

func (b MapBuilder) Add(key, value string) MapBuilder {
	b[key] = value
	return b
}
func (b MapBuilder) AddMany(more map[string]string) MapBuilder {
	for k, v := range more {
		b[k] = v
	}
	return b
}

func (b MapBuilder) Remove(key string) MapBuilder {
	delete(b, key)
	return b
}
