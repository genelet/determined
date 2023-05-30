package det

import (
	"encoding/json"
	"fmt"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	"reflect"
	"unicode"
)

// JsonUnmarshal unmarshals JSON data with interfaces determined by Determined.
//
//   - dat: JSON data
//   - current: object as interface
//   - spec: Determined
//   - ref: struct map, with key being string name and value pointer to struct
func JsonUnmarshal(dat []byte, current interface{}, spec *Struct, ref map[string]interface{}) error {
	if spec == nil {
		return json.Unmarshal(dat, current)
	}
	objectMap := spec.GetFields()
	if objectMap == nil || len(objectMap) == 0 {
		return json.Unmarshal(dat, current)
	}

	t := reflect.TypeOf(current).Elem()
	oriValue := reflect.ValueOf(&current).Elem()
	tmp := reflect.New(oriValue.Elem().Type()).Elem()
	tmp.Set(oriValue.Elem())

	n := t.NumField()

	var newFields []reflect.StructField
	found := false
	for i := 0; i < n; i++ {
		field := t.Field(i)
		name := field.Name
		if unicode.IsUpper([]rune(name)[0]) && field.Tag == "" {
			return fmt.Errorf("missing tag for %s", name)
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
			found = true
		} else {
			newFields = append(newFields, field)
		}
	}
	if found == false {
		return json.Unmarshal(dat, current)
	}

	newType := reflect.StructOf(newFields)
	raw := reflect.New(newType).Interface()
	err := json.Unmarshal(dat, raw)
	if err != nil {
		return err
	}
	rawValue := reflect.ValueOf(raw).Elem()

	for i := 0; i < n; i++ {
		field := t.Field(i)
		fieldType := field.Type
		name := field.Name
		f := tmp.Elem().Field(i)
		rawField := rawValue.Field(i)
		result, ok := objectMap[name]
		if ok {
			if x := result.GetMapStruct(); x != nil {
				nextMapStructs := x.GetMapFields()
				nSmaller := len(nextMapStructs)
				if nSmaller == 0 {
					return fmt.Errorf("missing map struct for %s", name)
				}
				var first *Struct
				for _, first = range nextMapStructs {
					break
				}

				n := rawField.Len()
				keys := rawField.MapKeys()
				fMap := reflect.MakeMap(fieldType)
				for i := 0; i < n; i++ {
					key := keys[i]
					v := rawField.MapIndex(key)
					nextStruct := first
					if i < nSmaller {
						if tmp, ok := nextMapStructs[key.String()]; ok {
							nextStruct = tmp
						}
					}
					trial := clone(ref[nextStruct.ClassName])
					err := JsonUnmarshal(v.Bytes(), trial, nextStruct, ref)
					if err != nil {
						return err
					}
					fMap.SetMapIndex(key, reflect.ValueOf(trial))
				}
				f.Set(fMap)
			} else if x := result.GetListStruct(); x != nil {
				n := rawField.Len()

				nextListStructs := x.GetListFields()
				nSmaller := len(nextListStructs)
				if nSmaller == 0 {
					return fmt.Errorf("missing list struct for %s", field.Name)
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

					trial := clone(ref[nextStruct.ClassName])
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
			} else if x := result.GetSingleStruct(); x != nil {
				trial := ref[x.ClassName]
				if trial == nil {
					return protoimpl.X.NewError("class ref not found for %s", x.ClassName)
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
			}
		} else if unicode.IsUpper([]rune(field.Name)[0]) {
			f.Set(rawField)
		}
	}

	oriValue.Set(tmp)

	return nil
}
