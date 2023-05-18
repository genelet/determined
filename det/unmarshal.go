package det

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"math/rand"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	"reflect"
	"strings"
	"unicode"
)

// JsonUnmarshal unmarshals JSON data with interfaces determined by Determined.
//
//   - dat: JSON data
//   - current: object as interface
//   - endpoint: Determined
//   - ref: struct map, with key being string name and value pointer to struct
func JsonUnmarshal(dat []byte, current interface{}, endpoint *Struct, ref map[string]interface{}) error {
	return unmarshal(dat, current, endpoint, ref, json.Unmarshal, "json")
}

// HclUnmarshal unmarshals HCL data with interfaces determined by Determined.
//
//   - dat: Hcl data
//   - current: object as interface
//   - endpoint: Determined
//   - ref: struct map, with key being string name and value pointer to struct
func HclUnmarshal(dat []byte, current interface{}, endpoint *Struct, ref map[string]interface{}) error {
	general := func(bs []byte, object interface{}) error {
fmt.Printf("STOP 4 data: %s\n object %#v\n", bs, object)
		return hclsimple.Decode(fmt.Sprintf("%d.hcl", rand.Int()), bs, nil, object)
	}
	return unmarshal(dat, current, endpoint, ref, general, "hcl")
}

func unmarshal(dat []byte, current interface{}, endpoint *Struct, ref map[string]interface{}, general func([]byte, any) error, postfix string) error {
	if endpoint == nil {
fmt.Printf("STOP 1\n")
		return general(dat, current)
	}
	objectMap := endpoint.GetFields()
	if objectMap == nil || len(objectMap) == 0 {
fmt.Printf("STOP 2 endpoint: %s\n data: %s\n current %#v\n", endpoint.String(), dat, current)
		return general(dat, current)
	}

	t := reflect.TypeOf(current).Elem()
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
			newField := reflect.StructField{Name: name, Tag: tag2tag(field.Tag, postfix)}
			if result.GetMapStruct() != nil {
				if postfix == "hcl" {
					newField.Type = reflect.TypeOf(map[string]map[string]string{})
				} else {
					newField.Type = reflect.TypeOf(map[string]json.RawMessage{})
				}
			} else if result.GetListStruct() != nil {
				if postfix == "hcl" {
					newField.Type = reflect.TypeOf([]map[string]string{})
				} else {
					newField.Type = reflect.TypeOf([]json.RawMessage{})
				}
			} else {
				if postfix == "hcl" {
					newField.Type = reflect.TypeOf(map[string][]byte{})
				} else {
					newField.Type = reflect.TypeOf(json.RawMessage{})
				}
			}
			newFields = append(newFields, newField)
			found = true
		} else {
			newFields = append(newFields, field)
		}
	}
	if found == false {
fmt.Printf("STOP 3\n")
		return general(dat, current)
	}

	newType := reflect.StructOf(newFields)
	raw := reflect.New(newType).Interface()
fmt.Printf("1111111111: %#v\n", raw)
fmt.Printf("1111111111: %s\n", dat)
	err := general(dat, raw)
	if err != nil {
		return err
	}
fmt.Printf("2222222222: %#v\n", raw)
	rawValue := reflect.ValueOf(raw).Elem()
fmt.Printf("33333: %#v\n", rawValue)

	oriValue := reflect.ValueOf(&current).Elem()
	tmp := reflect.New(oriValue.Elem().Type()).Elem()
	tmp.Set(oriValue.Elem())

	for i := 0; i < n; i++ {
		field := t.Field(i)
		fieldType := field.Type
		f := tmp.Elem().Field(i)
		rawField := rawValue.Field(i)
fmt.Printf("44444444: %T=> %#v\n", rawField, rawField)
		result, ok := objectMap[field.Name]
		if ok {
			if x := result.GetMapStruct(); x != nil {
				nextMapStructs := x.GetMapFields()
				nSmaller := len(nextMapStructs)
				if nSmaller == 0 {
					return fmt.Errorf("missing map struct for %s", field.Name)
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
					s := getBytes(v, postfix)
					err := unmarshal(s, trial, nextStruct, ref, general, postfix)
					if err != nil {
						return err
					}
					fMap.SetMapIndex(key, reflect.ValueOf(trial))
				}
				f.Set(fMap)
			} else if x := result.GetListStruct(); x != nil {
				nextListStructs := x.GetListFields()
				nSmaller := len(nextListStructs)
				if nSmaller == 0 {
					return fmt.Errorf("missing list struct for %s", field.Name)
				}

				n := rawField.Len()
				fSlice := reflect.MakeSlice(fieldType, n, n)
				first := nextListStructs[0]
				for k := 0; k < n; k++ {
					v := rawField.Index(k)
					nextStruct := first
					if k < nSmaller {
						nextStruct = nextListStructs[k]
					}
					trial := clone(ref[nextStruct.ClassName])
					s := getBytes(v, postfix)
					err := unmarshal(s, trial, nextStruct, ref, general, postfix)
					if err != nil {
						return err
					}
					fSlice.Index(k).Set(reflect.ValueOf(trial))
				}
				f.Set(fSlice)
			} else if x := result.GetSingleStruct(); x != nil {
				trial := ref[x.ClassName]
				if trial == nil {
					return protoimpl.X.NewError("class ref not found for %s", x.ClassName)
				}
fmt.Printf("1111 bbbbbb %#v\n", trial)
				trial = clone(trial)
				s := getBytes(rawField, postfix)
fmt.Printf("bbbbbb %s\n", s)
fmt.Printf("bbbbbb %#v\n", trial)
fmt.Printf("bbbbbb %#v\n", x)
				err := unmarshal(s, trial, x, ref, general, postfix)
fmt.Printf("cccccc %#v\n", err)
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

func tag2tag(old reflect.StructTag, postfix string) reflect.StructTag {
	if postfix == "hcl" {
		for _, tag := range strings.Fields(string(old)) {
			if len(tag) >= 5 && strings.ToLower(tag[:5]) == "hcl:\"" {
				two := strings.SplitN(tag, ",", 2)
				return reflect.StructTag(two[0]+"\"")
			} 
		}
	}
	return old
}

func getBytes(rawfield reflect.Value, postfix string) []byte {
	type Hash struct {
		Z map[string]string `hcl:"z"`
	}
	if postfix == "hcl" {
		x := rawfield.Interface().(map[string]string)
		y := Hash{x}
		f := hclwrite.NewEmptyFile()
		gohcl.EncodeIntoBody(&y, f.Body())
		z := string(f.Bytes())
		n := len(z)
		a := z[5:n-2]
fmt.Printf("88888 %s\n", a)
fmt.Printf("99999 %#v\n", strings.Split(a, ""))
		return []byte(a)
	}
	return rawfield.Bytes()
}
