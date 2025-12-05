package det

import (
	"reflect"
	"strings"
)

// collectStructTypesFromObject returns a type registry for unmarshaling any struct.
// It discovers all struct types by traversing the object's fields recursively.
// The implementations map provides concrete types for interface fields, mapping
// interface names to slices of concrete instances that implement them.
//
// The returned map contains both short names ("Config") and package-qualified
// names ("cell.Config") for each discovered struct type.
func collectStructTypesFromObject(obj any, implementations map[string][]any) map[string]any {
	ref := make(map[string]any)
	collectStructTypes(reflect.TypeOf(obj), ref, implementations)
	return ref
}

// collectStructTypes recursively collects all struct types from a type.
// It adds the type itself and all nested struct types to the ref map
// with both short and package-qualified names.
func collectStructTypes(t reflect.Type, ref map[string]any, implementations map[string][]any) {
	if t == nil {
		return
	}
	// Dereference pointer
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Handle interface types - register known implementations
	if t.Kind() == reflect.Interface {
		if impls, ok := implementations[t.Name()]; ok {
			for _, impl := range impls {
				collectStructTypes(reflect.TypeOf(impl), ref, implementations)
			}
		}
		return
	}

	// Handle map types - process value type
	if t.Kind() == reflect.Map {
		collectStructTypes(t.Elem(), ref, implementations)
		return
	}

	// Handle slice types - process element type
	if t.Kind() == reflect.Slice {
		collectStructTypes(t.Elem(), ref, implementations)
		return
	}

	if t.Kind() != reflect.Struct {
		return
	}

	// Add this struct type to ref (skip if already added or unnamed)
	name := t.Name()
	if name == "" || ref[name] != nil {
		return
	}
	obj := reflect.New(t).Interface()
	ref[name] = obj
	// Add package-qualified name (e.g., "cell.Config")
	if pkgPath := t.PkgPath(); pkgPath != "" {
		parts := strings.Split(pkgPath, "/")
		ref[parts[len(parts)-1]+"."+name] = obj
	}

	// Recurse into fields
	for i := 0; i < t.NumField(); i++ {
		collectStructTypes(t.Field(i).Type, ref, implementations)
	}
}
