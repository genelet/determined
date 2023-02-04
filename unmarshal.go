package determined

import (
	"encoding/json"
	"fmt"
	"reflect"
	"unicode"

	det "github.com/genelet/determined/xdetermined"
)

// JJUnmarshal unmarshals JSON data with interfaces determined by JSON-represented Determined.
//
//  - dat: JSON data
//  - current: pointer of the struct
//  - endpoint: Determined expressed in JSON
//  - ref: struct map, in which the key is its string name and value the pointer to struct
//
func JJUnmarshal(dat []byte, current interface{}, endpoint []byte, ref map[string]interface{}) error {
	theEndpoint := Determined{}
	err := json.Unmarshal(endpoint, &theEndpoint)
	if err != nil {
		return err
	}
	return JsonUnmarshal(dat, current, &theEndpoint, ref)
}

// JsonUnmarshal unmarshals JSON data with interfaces determined by Determined.
//
//  - dat: JSON data
//  - current: object as interface
//  - endpoint: Determined
//  - ref: struct map, in which the key is its string name and value the pointer to struct
//
func JsonUnmarshal(dat []byte, current interface{}, endpoint *structpb.Struct, ref map[string]interface{}) error {
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
			switch result.Kind {
			case *det.Value_MapEnd, *x.Value_MapStruct:
				newField.Type = reflect.TypeOf(map[string]json.RawMessage{})
			case *det.Value_ListEnd, *x.Value_ListStruct:
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
			run := func(bs []byte, dex ...interface{}) (interface{}, error) {
				switch 
				confName, nextmap, err := result.getPair(dex...)
				if err != nil { return nil, err }
				conf, ok := ref[confName]
				if !ok && nextmap != nil {
					return nil, fmt.Errorf("struct %s not found", confName)
				}
				trial := Clone(conf)
				err = JsonUnmarshal(bs, trial, &Determined{MetaType:METASingle, SingleField:nextmap}, ref)
				return trial, err
			}

			if x := result.GetSingleEnd(); x != nil {
				trial := Clone(ref[x])
				err := json.Unmarshal(rawField.Bytes(), trial)
				if err != nil { return err }
				if f.Kind() == reflect.Interface || f.Kind() == reflect.Ptr {
					f.Set(reflect.ValueOf(trial))
				} else {
					f.Set(reflect.ValueOf(trial).Elem())
				}
			} else if x := result.GetMapEnd(); x != nil {
				n := rawField.Len()
				fMap := reflect.MakeMap(fieldType)
				keys := rawField.MapKeys()
				nextValues := x.GetMapEndFields()
				for i:=0; i<n; i++ {
					k := keys[i]
					v := rawField.MapIndex(k)
					key := k.String()
					nextValue := nextValues[key]
					trial := Clone(ref[nextValue])
					err := json.Unmarshal(v.Bytes(), trial)
					if err != nil { return err }
					fMap.SetMapIndex(k, reflect.ValueOf(trial))
				}
				f.Set(fMap)
			} else if x := result.GetListEnd(); x != nil {
				n := rawField.Len()
				fSlice := reflect.MakeSlice(fieldType, n, n)
				nextValues := x.GetListEndFields()
				for k:=0; k<n; k++ {
					v := rawField.Index(k)
					nextValue := nextValues[k]
					trial := Clone(ref[nextValue])
					err := json.Unmarshal(v.Bytes(), trial)
					if err != nil { return err }
					fSlice.Index(k).Set(reflect.ValueOf(trial))
				}
				f.Set(fSlice)
			} else if x := result.GetSingleStruct(); x != nil {
				err = JsonUnmarshal(bs, trial, &Determined{MetaType:METASingle, SingleField:nextmap}, ref)
				trial := Clone(ref[x])
				err := json.Unmarshal(rawField.Bytes(), trial)
				if err != nil { return err }
				if f.Kind() == reflect.Interface || f.Kind() == reflect.Ptr {
					f.Set(reflect.ValueOf(trial))
				} else {
					f.Set(reflect.ValueOf(trial).Elem())
				}
			case *structpb.Value_StringValue:
				:wq
			default:
				
				trial, err := run(rawField.Bytes(), result.AsIntgerface())
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
