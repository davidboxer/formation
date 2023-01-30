package controller

import (
	"github.com/imdario/mergo"
	v1 "k8s.io/api/core/v1"
	"reflect"
)

type Transformers struct {
	transformers []mergo.Transformers
}

func NewTransformersWithDefault() *Transformers {
	return &Transformers{
		transformers: []mergo.Transformers{
			WithOverwriteZeroValue(),
			WithKubernetesVolumeV1OverWrite(),
			WithKubernetesVolumeMountV1OverWrite(),
		},
	}
}

func (t *Transformers) Add(transformers ...mergo.Transformers) {
	t.transformers = append(t.transformers, transformers...)
}

func (t *Transformers) Transformer(reflectType reflect.Type) func(dst, src reflect.Value) error {
	for _, transformer := range t.transformers {
		// Check if transformer return non nil error
		fn := transformer.Transformer(reflectType)
		if fn != nil {
			return fn
		}
	}
	return nil
}

type kubernetesVolumeV1OverWriteTransformer struct{}

func WithKubernetesVolumeV1OverWrite() kubernetesVolumeV1OverWriteTransformer {
	return kubernetesVolumeV1OverWriteTransformer{}
}
func (kubernetesVolumeV1OverWriteTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	// Check if the type is a list of kubernetes volume
	if typ.Kind() == reflect.Slice && typ.Elem().Kind() == reflect.Struct {
		if typ.Elem().Name() == "Volume" && typ.Elem().PkgPath() == "k8s.io/api/core/v1" {
			return func(dst, src reflect.Value) error {
				// cast the src and dst to slice of volume
				srcSlice := src.Interface().([]v1.Volume)
				dstSlice := dst.Interface().([]v1.Volume)
				// create a map of volume name to volume
				srcMap := map[string]v1.Volume{}
				for _, volume := range srcSlice {
					srcMap[volume.Name] = volume
				}
				// Loop over the dst slice and update the volume if exist in the src map
				for i, volume := range dstSlice {
					if srcVolume, ok := srcMap[volume.Name]; ok {
						srcVolume.DeepCopyInto(&dstSlice[i])
					}
				}
				return nil
			}
		}
	}
	return nil
}

type kubernetesVolumeMountV1OverWriteTransformer struct{}

func WithKubernetesVolumeMountV1OverWrite() kubernetesVolumeMountV1OverWriteTransformer {
	return kubernetesVolumeMountV1OverWriteTransformer{}
}
func (kubernetesVolumeMountV1OverWriteTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ.Kind() == reflect.Slice && typ.Elem().Kind() == reflect.Struct {
		if typ.Elem().Name() == "VolumeMount" && typ.Elem().PkgPath() == "k8s.io/api/core/v1" {
			return func(dst, src reflect.Value) error {
				// cast the src and dst to slice of volume
				srcSlice := src.Interface().([]v1.VolumeMount)
				dstSlice := dst.Interface().([]v1.VolumeMount)
				// create a map of volume name to volume
				srcMap := map[string]v1.VolumeMount{}
				for _, volume := range srcSlice {
					srcMap[volume.Name] = volume
				}
				// Loop over the dst slice and update the volume if exist in the src map
				for i, volume := range dstSlice {
					if srcVolume, ok := srcMap[volume.Name]; ok {
						srcVolume.DeepCopyInto(&dstSlice[i])
					}
				}
				return nil
			}
		}
	}
	return nil
}

func WithOverwriteZeroValue() overwriteZeroValueTransformer {
	return overwriteZeroValueTransformer{}
}

// overwriteZeroValueTransformer The package mergo does not overwrite zero value, this transformer will overwrite zero value.
// This is needed to update the resource replicas to 0 if the user need to scale down the resource.
type overwriteZeroValueTransformer struct{}

func (overwriteZeroValueTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if !isNumber(typ) {
		return nil
	}
	return func(dst, src reflect.Value) error {
		if dst.CanSet() {
			dst.Set(src)
			return nil
		}
		return nil
	}
}

func isNumber(v reflect.Type) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}
