package det

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/genelet/schema"
)

// JsonUnmarshal unmarshals JSON data with interfaces determined by Determined.
//
//   - dat: JSON data
//   - current: object as interface
//   - spec: Determined
//   - ref: struct map, with key being string name and value pointer to struct.
//     The ref map can also contain []any values to specify interface implementations,
//     which will be used for auto-discovery of struct types.
//
// Example with implementations:
//
//	ref := map[string]any{
//	    "Shape": []any{new(Circle), new(Square)},  // interface implementations
//	}
//	err := JsonUnmarshal(data, &config, spec, ref)
func JsonUnmarshal(dat []byte, current any, spec *schema.Struct, ref map[string]any) error {
	if spec == nil {
		return json.Unmarshal(dat, current)
	}
	objectMap := spec.GetFields()
	if len(objectMap) == 0 {
		return json.Unmarshal(dat, current)
	}

	// Extract implementations from ref (values that are []any)
	implementations := make(map[string][]any)
	for k, v := range ref {
		if impls, ok := v.([]any); ok {
			implementations[k] = impls
		}
	}

	// Auto-collect struct types from the target object
	autoRef := collectStructTypesFromObject(current, implementations)

	// Merge passed ref into autoRef (passed ref takes precedence)
	for k, v := range ref {
		if _, isSlice := v.([]any); !isSlice {
			autoRef[k] = v
		}
	}
	ref = autoRef

	t := reflect.TypeOf(current).Elem()
	newFields, origTypes, err := loopFields(t, objectMap)
	if err != nil {
		return err
	}

	newType := reflect.StructOf(newFields)
	raw := reflect.New(newType).Interface()
	err = json.Unmarshal(dat, raw)
	if err != nil {
		return err
	}
	rawValue := reflect.ValueOf(raw).Elem()

	oriValue := reflect.ValueOf(&current).Elem()
	tmp := reflect.New(oriValue.Elem().Type()).Elem()
	tmp.Set(oriValue.Elem())

	for i, field2 := range newFields {
		rawField := rawValue.Field(i)
		name := field2.Name
		field, ok := origTypes[name]
		if !ok {
			field = field2
		}
		fieldType := field.Type
		f := tmp.Elem().FieldByName(name)
		result, ok := objectMap[name]
		if !ok {
			f.Set(rawField)
			continue
		}
		if x := result.GetMap2Struct(); x != nil {
			if err := handleMap2Struct(f, rawField, fieldType, x, ref, name); err != nil {
				return err
			}
		} else if x := result.GetMapStruct(); x != nil {
			if err := handleMapStruct(f, rawField, fieldType, x, ref, name); err != nil {
				return err
			}
		} else if x := result.GetListStruct(); x != nil {
			if err := handleListOrMapStruct(f, rawField, fieldType, x, ref, field.Name); err != nil {
				return err
			}
		} else if x := result.GetSingleStruct(); x != nil {
			if err := handleSingleStruct(f, rawField, x, ref); err != nil {
				return err
			}
		}
	}

	oriValue.Set(tmp)

	return nil
}

func handleMapStruct(f, rawField reflect.Value, fieldType reflect.Type, x *schema.MapStruct, ref map[string]any, name string) error {
	nextMapStructs := x.GetMapFields()
	if len(nextMapStructs) == 0 {
		return fmt.Errorf("missing map struct for %s", name)
	}

	// Get default struct using deterministic key order
	first := getFirstStructFromMap(nextMapStructs)

	n := rawField.Len()
	keys := rawField.MapKeys()
	fMap := reflect.MakeMap(fieldType)
	for i := 0; i < n; i++ {
		key := keys[i]
		v := rawField.MapIndex(key)
		nextStruct := first
		if tmp, ok := nextMapStructs[key.String()]; ok {
			nextStruct = tmp
		}
		trial := ref[nextStruct.ClassName]
		if trial == nil {
			return fmt.Errorf("class ref not found for %s", nextStruct.ClassName)
		}
		trial = clone(trial)
		err := JsonUnmarshal(v.Bytes(), trial, nextStruct, ref)
		if err != nil {
			return err
		}
		fMap.SetMapIndex(key, reflect.ValueOf(trial))
	}
	f.Set(fMap)
	return nil
}

// handleListOrMapStruct handles both slice fields ([]T) and map fields (map[string]T)
// where the spec provides a list of struct types. For slices, the index determines
// which struct spec to use. For maps, all values use the first struct spec as the type.
func handleListOrMapStruct(f, rawField reflect.Value, fieldType reflect.Type, x *schema.ListStruct, ref map[string]any, fieldName string) error {
	n := rawField.Len()

	nextListStructs := x.GetListFields()
	nSmaller := len(nextListStructs)
	if nSmaller == 0 {
		return fmt.Errorf("missing list struct for %s", fieldName)
	}
	first := nextListStructs[0]

	var fSlice, fMap reflect.Value
	var keys []reflect.Value
	if fieldType.Kind() == reflect.Map {
		keys = rawField.MapKeys()
		fMap = reflect.MakeMapWithSize(fieldType, n)
	} else {
		fSlice = reflect.MakeSlice(fieldType, n, n)
	}

	for k := 0; k < n; k++ {
		var key, v reflect.Value
		if fieldType.Kind() == reflect.Map {
			key = keys[k]
			v = rawField.MapIndex(key)
		} else {
			v = rawField.Index(k)
		}

		nextStruct := first
		if k < nSmaller {
			nextStruct = nextListStructs[k]
		}

		trial := ref[nextStruct.ClassName]
		if trial == nil {
			return fmt.Errorf("class ref not found for %s", nextStruct.ClassName)
		}
		trial = clone(trial)
		err := JsonUnmarshal(v.Bytes(), trial, nextStruct, ref)
		if err != nil {
			return err
		}
		if fieldType.Kind() == reflect.Map {
			fMap.SetMapIndex(key, reflect.ValueOf(trial))
		} else {
			fSlice.Index(k).Set(reflect.ValueOf(trial))
		}
	}
	if fieldType.Kind() == reflect.Map {
		f.Set(fMap)
	} else {
		f.Set(fSlice)
	}
	return nil
}

func handleSingleStruct(f, rawField reflect.Value, x *schema.Struct, ref map[string]any) error {
	trial := ref[x.ClassName]
	if trial == nil {
		return fmt.Errorf("class ref not found for %s", x.ClassName)
	}
	trial = clone(trial)
	err := JsonUnmarshal(rawField.Bytes(), trial, x, ref)
	if err != nil {
		return err
	}
	if f.Kind() == reflect.Interface || f.Kind() == reflect.Ptr {
		f.Set(reflect.ValueOf(trial))
	} else {
		f.Set(reflect.ValueOf(trial).Elem())
	}
	return nil
}

// handleMap2Struct handles map[[2]string]T fields where JSON is represented as nested objects.
// JSON format: {"key1": {"key2": {...}}}
// Go type: map[[2]string]*SomeStruct
func handleMap2Struct(f, rawField reflect.Value, fieldType reflect.Type, x *schema.Map2Struct, ref map[string]any, name string) error {
	nextMap2Structs := x.GetMap2Fields()
	if len(nextMap2Structs) == 0 {
		return fmt.Errorf("missing map2 struct for %s", name)
	}

	// Get a default struct for unmarshaling using deterministic key order
	firstMapStruct := getFirstMapStructFromMap(nextMap2Structs)
	if firstMapStruct == nil {
		return fmt.Errorf("missing inner map struct for map2 %s", name)
	}
	firstStruct := getFirstStructFromMap(firstMapStruct.GetMapFields())
	if firstStruct == nil {
		return fmt.Errorf("missing inner struct for map2 %s", name)
	}

	// Create the result map with [2]string keys
	fMap := reflect.MakeMap(fieldType)

	// Iterate over outer map (first level keys)
	outerKeys := rawField.MapKeys()
	for _, key0 := range outerKeys {
		key0Str := key0.String()
		innerRaw := rawField.MapIndex(key0)

		// Get the MapStruct for this key0, or use first as default
		mapStruct := firstMapStruct
		if ms, ok := nextMap2Structs[key0Str]; ok {
			mapStruct = ms
		}
		innerStructs := mapStruct.GetMapFields()

		// innerRaw should be a map[string]json.RawMessage
		if innerRaw.Kind() != reflect.Map {
			return fmt.Errorf("expected nested map for key %s in map2 %s", key0Str, name)
		}

		// Iterate over inner map (second level keys)
		innerKeys := innerRaw.MapKeys()
		for _, key1 := range innerKeys {
			key1Str := key1.String()
			v := innerRaw.MapIndex(key1)

			// Get the Struct for this key1, or use first as default
			nextStruct := firstStruct
			if s, ok := innerStructs[key1Str]; ok {
				nextStruct = s
			}

			trial := ref[nextStruct.ClassName]
			if trial == nil {
				return fmt.Errorf("class ref not found for %s", nextStruct.ClassName)
			}
			trial = clone(trial)

			err := JsonUnmarshal(v.Bytes(), trial, nextStruct, ref)
			if err != nil {
				return err
			}

			// Create [2]string key
			keyArr := [2]string{key0Str, key1Str}
			fMap.SetMapIndex(reflect.ValueOf(keyArr), reflect.ValueOf(trial))
		}
	}

	f.Set(fMap)
	return nil
}
