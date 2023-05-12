package det

import (
	"encoding/json"
	"fmt"
	"reflect"
	"unicode"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

// JsonUnmarshal unmarshals JSON data with interfaces determined by Determined.
//
//  - dat: JSON data
//  - current: object as interface
//  - endpoint: Determined
//  - ref: struct map, with key being string name and value pointer to struct
//
func JsonUnmarshal(dat []byte, current interface{}, endpoint *Struct, ref map[string]interface{}) error {
	if endpoint == nil {
		return json.Unmarshal(dat, current)
	}
	objectMap := endpoint.GetFields()
	if objectMap == nil || len(objectMap) == 0 {
		return json.Unmarshal(dat, current)
	}

	t := reflect.TypeOf(current).Elem()
	n := t.NumField()

	var newFields []reflect.StructField
	found := false
	for i:=0; i<n; i++ {
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
				newField.Type = reflect.TypeOf([]json.RawMessage{})
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
	if err != nil { return err }
	rawValue := reflect.ValueOf(raw).Elem()

	oriValue := reflect.ValueOf(&current).Elem()
	tmp := reflect.New(oriValue.Elem().Type()).Elem()
	tmp.Set(oriValue.Elem())

	for i:=0; i<n; i++ {
		field := t.Field(i)
		fieldType := field.Type
		f := tmp.Elem().Field(i)
		rawField := rawValue.Field(i)
		result, ok := objectMap[field.Name]
		if ok {
			if x := result.GetMapStruct(); x != nil {
				nextMapStructs := x.GetMapFields()
				nSmaller := len(nextMapStructs)
				var first *Struct
				for _, first = range nextMapStructs { break }

				n := rawField.Len()
				keys := rawField.MapKeys()
				fMap := reflect.MakeMap(fieldType)
				for i:=0; i<n; i++ {
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
					if err != nil { return err }
					fMap.SetMapIndex(key, reflect.ValueOf(trial))
				}
				f.Set(fMap)
			} else if x := result.GetListStruct(); x != nil {
				nextListStructs := x.GetListFields()
				nSmaller := len(nextListStructs)

				n := rawField.Len()
				fSlice := reflect.MakeSlice(fieldType, n, n)
				first := nextListStructs[0]
				for k:=0; k<n; k++ {
					v := rawField.Index(k)
					nextStruct := first
					if k < nSmaller {
						nextStruct = nextListStructs[k]
					}
					trial := clone(ref[nextStruct.ClassName])
					err := JsonUnmarshal(v.Bytes(), trial, nextStruct, ref)
					if err != nil { return err }
					fSlice.Index(k).Set(reflect.ValueOf(trial))
				}
				f.Set(fSlice)
			} else if x := result.GetSingleStruct(); x != nil {
				trial := ref[x.ClassName]
				if trial == nil {
					return protoimpl.X.NewError("class ref not found for %s", x.ClassName)
				}
				trial = clone(trial)
				err := JsonUnmarshal(rawField.Bytes(), trial, x, ref)
				if err != nil { return err }
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
