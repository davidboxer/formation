package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// GetValueWithDefaultValue Try to find the value given some key, if the key does not exist return a default value
func GetValueWithDefaultValue(obj interface{}, path string, defaultValue interface{}) interface{} {
	if obj == nil {
		return defaultValue
	}
	value, err := GetValue2(obj, path)
	if err != nil {
		return defaultValue
	}
	return value
}

func GetValue2(obj interface{}, path string) (reflect.Value, error) {
	return GetValue(obj, strings.Split(path, "."))
}
func GetValue(obj interface{}, keys []string) (reflect.Value, error) {
	if obj == nil {
		return reflect.Value{}, errors.New("object can not be nil")
	}
	objType := reflect.ValueOf(obj)
	var err error
	for _, key := range keys {
		if objType, err = get(objType, key); err != nil {
			return reflect.Value{}, err
		}
	}
	return objType, nil
}

func get(value reflect.Value, key string) (reflect.Value, error) {
	var err error
	//reflect and cast is about the same speed outside of pure loops
	// Since Array and Slice are not the same thing []interface{} will not work for both
	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		if value, err = getIndexFromList(value, key); err != nil {
			return reflect.Value{}, err
		}
	case reflect.Map:
		value = value.MapIndex(reflect.ValueOf(key))
		if !value.IsValid() {
			return reflect.Value{}, fmt.Errorf("key: %s does not exist in map", key)
		}
	case reflect.Struct:
		value = value.FieldByName(key)
		if !value.IsValid() {
			return reflect.Value{}, fmt.Errorf("key: %s does not exist in struct", key)
		}
		return value, nil
	case reflect.Ptr:
		return get(reflect.Indirect(value), key)
	}
	if value.Kind() == reflect.String {
		return value, nil
	}

	return value.Elem(), nil
}

/*helper func for getting an index of a list
 */
func getIndexFromList(value reflect.Value, key string) (reflect.Value, error) {
	index, err := strconv.Atoi(key)
	if err != nil {
		return value, err
	}
	if index > value.Len() {
		return value, fmt.Errorf("Index: %d out of bounds  ", index)
	}
	return value.Index(index), nil
}
