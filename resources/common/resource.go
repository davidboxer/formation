package common

import (
	"github.com/Doout/formation/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SimpleResource[T client.Object] struct {
	*types.ConvergedGroup
	onCreate      func(T)
	typeName      string
	DisableUpdate bool
	Obj           T
}

func NewSimpleResourceWithOnCreate[T client.Object](typeName string, obj T, onCreate func(T)) *SimpleResource[T] {
	return &SimpleResource[T]{Obj: obj, onCreate: onCreate, ConvergedGroup: &types.ConvergedGroup{}, typeName: typeName}
}

func NewSimpleResource[T client.Object](typeName string, obj T) *SimpleResource[T] {
	return &SimpleResource[T]{Obj: obj, ConvergedGroup: &types.ConvergedGroup{}, typeName: typeName}
}

func (s *SimpleResource[T]) Type() string { return s.typeName }
func (s *SimpleResource[T]) Name() string { return s.Obj.GetName() }

// Runtime returns a new instance of the object without calling DeepCopy.
// This method uses reflection to create a new instance of the object.
// If the object is a pointer, it will be dereferenced before creating a new instance.
// Returns a pointer to the new instance of the object.
func (s *SimpleResource[T]) Runtime() client.Object {
	t := reflect.TypeOf(s.Obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return reflect.New(t).Elem().Addr().Interface().(client.Object)
}
func (s *SimpleResource[T]) Create() (client.Object, error) {
	if s.DisableUpdate {
		if s.Obj.GetAnnotations() == nil {
			s.Obj.SetAnnotations(make(map[string]string))
		}
		s.Obj.GetAnnotations()[types.UpdateKey] = "disabled"
	}
	if s.onCreate != nil {
		s.onCreate(s.Obj)
	}
	return s.Obj, nil
}
