package struct_merge

import (
	"fmt"
	"reflect"
)

// mergeStructs merges non-zero fields from src into dst.
// Both dst and src must be pointers to structs.
func Merge[T any](dst, src *T) error {
	dstVal := reflect.ValueOf(dst).Elem()
	srcVal := reflect.ValueOf(src).Elem()

	// Ensure both values are structs.
	if dstVal.Kind() != reflect.Struct || srcVal.Kind() != reflect.Struct {
		return fmt.Errorf("mergeStructs only supports struct types")
	}

	typ := dstVal.Type()
	for i := 0; i < typ.NumField(); i++ {
		sField := srcVal.Field(i)
		dField := dstVal.Field(i)

		// Skip if destination field cannot be set.
		if !dField.CanSet() {
			continue
		}

		// If the source field is not the zero value, copy it over.
		if !isZero(sField) {
			dField.Set(sField)
		}
	}
	return nil
}

// isZero checks whether a reflect.Value is the zero value for its type.
func isZero(v reflect.Value) bool {
	zero := reflect.Zero(v.Type())
	return reflect.DeepEqual(v.Interface(), zero.Interface())
}