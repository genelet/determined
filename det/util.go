package det

import (
	"encoding/json"
	"reflect"
	"strings"
	"unicode"
)

// clone clones a value via pointer
func clone(old any) any {
	obj := reflect.New(reflect.TypeOf(old).Elem())
	oldVal := reflect.ValueOf(old).Elem()
	newVal := obj.Elem()
	for i := 0; i < oldVal.NumField(); i++ {
		newValField := newVal.Field(i)
		if newValField.CanSet() {
			newValField.Set(oldVal.Field(i))
		}
	}

	return obj.Interface()
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
			if result.GetMapStruct() != nil {
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
