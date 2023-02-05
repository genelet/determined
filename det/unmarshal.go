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
//  - ref: struct map, in which the key is its string name and value the pointer to struct
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
		if result, ok := objectMap[field.Name]; ok {
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
				n := rawField.Len()
				fMap := reflect.MakeMap(fieldType)
				keys := rawField.MapKeys()
				nextMapStructs := x.GetMapFields()
				var last, trial interface{}
				for i:=0; i<n; i++ {
					k := keys[i]
					v := rawField.MapIndex(k)
					key := k.String()
					nextStruct := nextMapStructs[key]
					if nextStruct != nil { last = nextStruct }
					trial = ref[nextStruct.ClassName]
					if trial == nil { trial = last }
					trial = clone(trial)
					err := JsonUnmarshal(v.Bytes(), trial, nextStruct, ref)
					if err != nil { return err }
					fMap.SetMapIndex(k, reflect.ValueOf(trial))
				}
				f.Set(fMap)
			} else if x := result.GetListStruct(); x != nil {
				n := rawField.Len()
				fSlice := reflect.MakeSlice(fieldType, n, n)
				nextListStructs := x.GetListFields()
				var last, trial interface{}
				for k:=0; k<n; k++ {
					v := rawField.Index(k)
					nextStruct := nextListStructs[k]
					if nextStruct != nil { last = nextStruct }
					trial = ref[nextStruct.ClassName]
					if trial == nil { trial = last }
					trial = clone(trial)
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
