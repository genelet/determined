package det

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

// clone creates a new zero-value instance of the same type as old.
// The old parameter must be a pointer to a struct.
// This is used to create fresh instances from type registry templates
// before unmarshaling JSON data into them.
//
// Note: This creates a zero-value instance rather than copying fields.
// This is intentional because:
// 1. Templates in the type registry should be zero-value prototypes
// 2. The instance will be immediately populated by JSON unmarshaling
// 3. This avoids shallow copy issues with pointer fields
func clone(old any) any {
	return reflect.New(reflect.TypeOf(old).Elem()).Interface()
}

func loopFields(t reflect.Type, objectMap map[string]*Value) ([]reflect.StructField, map[string]reflect.StructField, error) {
	var newFields []reflect.StructField
	origTypes := make(map[string]reflect.StructField)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := field.Name
		if !unicode.IsUpper([]rune(name)[0]) {
			continue
		}
		tcontent := field.Tag.Get("json")
		if tcontent == "-" || strings.HasSuffix(tcontent, ",-") {
			continue
		}
		if tcontent == "" && field.Anonymous && field.Type.Kind() == reflect.Struct {
			deeps, deepTypes, err := loopFields(field.Type, objectMap)
			if err != nil {
				return nil, nil, err
			}
			newFields = append(newFields, deeps...)
			for k, v := range deepTypes {
				origTypes[k] = v
			}
			continue
		}
		if result, ok := objectMap[name]; ok {
			newField := reflect.StructField{Name: name, Tag: field.Tag}
			if result.GetMap2Struct() != nil {
				// Nested map for map[[2]string]T: {"k1": {"k2": {...}}}
				newField.Type = reflect.TypeOf(map[string]map[string]json.RawMessage{})
			} else if result.GetMapStruct() != nil {
				newField.Type = reflect.TypeOf(map[string]json.RawMessage{})
			} else if result.GetListStruct() != nil {
				if field.Type.Kind() == reflect.Map {
					newField.Type = reflect.TypeOf(map[string]json.RawMessage{})
				} else {
					newField.Type = reflect.TypeOf([]json.RawMessage{})
				}
			} else {
				newField.Type = reflect.TypeOf(json.RawMessage{})
			}
			newFields = append(newFields, newField)
			origTypes[name] = field
		} else {
			newFields = append(newFields, field)
		}
	}
	return newFields, origTypes, nil
}

// getFirstStructFromMap returns the first struct from a map using deterministic key order.
// This ensures consistent behavior across runs since Go map iteration order is random.
func getFirstStructFromMap(m map[string]*Struct) *Struct {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return m[keys[0]]
}

// getFirstMapStructFromMap returns the first MapStruct from a map using deterministic key order.
func getFirstMapStructFromMap(m map[string]*MapStruct) *MapStruct {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return m[keys[0]]
}
