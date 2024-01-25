package det

import (
	"encoding/json"
	"fmt"
	"reflect"

	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
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
	}

	oriValue.Set(tmp)

	return nil
}
