package utils

type PointerType interface {
	int | int8 | int16 | int32 | int64 | float32 | float64 | string | bool
}

func Pointer[T PointerType](b T) *T {
	return &b
}
