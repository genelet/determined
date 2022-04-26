package determined

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// JJUnmarshal unmarshals json data with interfaces determined by JSON-represented DeterminedMap
// dat: the JSON data
// current: pointer of the struct
// objectMap: the DeterminedMap in JSON
// ref: structs involed in the Unmarshaller should be placed here, with key the name and value the new pointer to struct
//
func JJUnmarshal(dat []byte, current interface{}, objectMap []byte, ref map[string]interface{}) error {
	theMap := DeterminedMap{}
	err := json.Unmarshal(objectMap, &theMap)
	if err != nil {
		return err
	}
	return JsonUnmarshal(dat, current, theMap, ref)
}

// JsonUnmarshal unmarshals json data with interfaces determined by DeterminedMap
// dat: the JSON data
// current: pointer of the struct
// objectMap: the DeterminedMap
// ref: structs involed in the Unmarshaller should be placed here, with key the name and value the new pointer to struct
//
func JsonUnmarshal(dat []byte, current interface{}, objectMap DeterminedMap, ref map[string]interface{}) error {
	if objectMap == nil || len(objectMap) == 0 {
		return json.Unmarshal(dat, current)
	}

	t := reflect.TypeOf(current).Elem()
	n := t.NumField()

	var newFields []reflect.StructField
	found := false
	for i:=0; i<n; i++ {
		field := t.Field(i)
		if field.Tag == "" {
			return fmt.Errorf("missing tag for %s", field.Name)
		}
		if result, ok := objectMap[field.Name]; ok {
			newField := reflect.StructField{Name: field.Name, Tag: field.Tag}
			switch result.MetaType {
			case METAMap, METAMapSingle:
				newField.Type = reflect.TypeOf(map[string]json.RawMessage{})
			case METASlice, METASliceSingle:
				newField.Type = reflect.TypeOf([]json.RawMessage{})
			default:
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
			run := func(bs []byte, dex interface{}) (interface{}, error) {
				confName, nextmap, err := result.GetPair(dex)
				if err != nil { return nil, err }
				conf, ok := ref[confName]
				if !ok && nextmap != nil {
					return nil, fmt.Errorf("struct %s not found", confName)
				}
				trial := Clone(conf)
				err = JsonUnmarshal(bs, trial, nextmap, ref)
				return trial, err
			}
			switch result.MetaType {
			case METAMap, METAMapSingle:
				n := rawField.Len()
				fMap := reflect.MakeMap(fieldType)
				keys := rawField.MapKeys()
				for i:=0; i<n; i++ {
					k := keys[i]
					v := rawField.MapIndex(k)
					trial, err := run(v.Bytes(), k.String())
					if err != nil { return err }
					fMap.SetMapIndex(k, reflect.ValueOf(trial))
				}
				f.Set(fMap)
			case METASlice, METASliceSingle:
				n := rawField.Len()
				fSlice := reflect.MakeSlice(fieldType, n, n)
				for k:=0; k<n; k++ {
					v := rawField.Index(k)
					trial, err := run(v.Bytes(), k)
					if err != nil { return err }
					fSlice.Index(k).Set(reflect.ValueOf(trial))
				}
				f.Set(fSlice)
			default:
				trial, err := run(rawField.Bytes(), result.SingleName)
				if err != nil { return err }
				if f.Kind() == reflect.Interface || f.Kind() == reflect.Ptr {
					f.Set(reflect.ValueOf(trial))
				} else {
					f.Set(reflect.ValueOf(trial).Elem())
				}
			}
		} else {
			f.Set(rawField)
		}
	}

	oriValue.Set(tmp)

	return nil
}
