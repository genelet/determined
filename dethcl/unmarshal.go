package dethcl

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"reflect"
	"strings"
	"unicode"
)

// Unmarshal decodes HCL data with interfaces determined by Determined.
//
//   - dat: Hcl data
//   - current: object as interface
//   - spec: Determined
//   - ref: map, with key being string name, and value referenced value
//   - optional label_values: fields' values of labels
func Unmarshal(dat []byte, current interface{}, spec *Struct, ref map[string]interface{}, label_values ...string) error {
	if spec == nil {
		return unplain(dat, current, label_values...)
	}
	objectMap := spec.GetFields()
	if objectMap == nil || len(objectMap) == 0 {
		return unplain(dat, current, label_values...)
	}

	file, diags := hclsyntax.ParseConfig(dat, rname(), hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return diags
	}

	t := reflect.TypeOf(current).Elem()
	newFields, origTypes, err := loopFields(t, objectMap, ref)
	if err != nil {
		return err
	}
	if len(origTypes) == 0 {
		return unplain(dat, current, label_values...)
	}

	tagref := getTagref(origTypes)
	body := &hclsyntax.Body{
		Attributes: file.Body.(*hclsyntax.Body).Attributes,
		SrcRange:   file.Body.(*hclsyntax.Body).SrcRange,
		EndRange:   file.Body.(*hclsyntax.Body).EndRange}
	blockref := make(map[string][]*hclsyntax.Block)
	for _, block := range file.Body.(*hclsyntax.Body).Blocks {
		if tagref[block.Type] {
			blockref[block.Type] = append(blockref[block.Type], block)
		} else {
			body.Blocks = append(body.Blocks, block)
		}
	}

	newType := reflect.StructOf(newFields)
	//raw := reflect.New(newType).Elem().Addr().Interface()
	raw := reflect.New(newType).Interface()
	diags = gohcl.DecodeBody(body, nil, raw)
	if diags.HasErrors() {
		return diags
	}
	rawValue := reflect.ValueOf(raw).Elem()

	oriValue := reflect.ValueOf(&current).Elem()
	tmp := reflect.New(oriValue.Elem().Type()).Elem()
	tmp.Set(oriValue.Elem())

	m := 0
	if label_values != nil {
		m = len(label_values)
	}
	k := 0
	for i, field := range newFields {
		name := field.Name
		two := tag2(field.Tag)
		f := tmp.Elem().FieldByName(name)
		if strings.ToLower(two[1]) == "label" && k < m {
			f.Set(reflect.ValueOf(label_values[k]))
			k++
		} else {
			rawField := rawValue.Field(i)
			f.Set(rawField)
		}
	}

	for _, field := range origTypes {
		name := field.Name
		typ := field.Type
		f := tmp.Elem().FieldByName(name)
		two := tag2(field.Tag)
		blocks := blockref[two[0]]
		result := objectMap[name]
		if x := result.GetListStruct(); x != nil {
			nextListStructs := x.GetListFields()
			nSmaller := len(nextListStructs)
			if nSmaller == 0 {
				continue
			}
			first := nextListStructs[0]

			n := len(blocks)

			var fSlice, fMap reflect.Value
			if typ.Kind() == reflect.Map {
				fMap = reflect.MakeMapWithSize(typ, n)
			} else {
				fSlice = reflect.MakeSlice(typ, n, n)
			}
			for k := 0; k < n; k++ {
				nextStruct := first
				if k < nSmaller {
					nextStruct = nextListStructs[k]
				}
				trial := ref[nextStruct.ClassName]
				if trial == nil {
					return fmt.Errorf("ref not found for %s", name)
				}
				trial = clone(trial)
				s, labels, err := getBlockBytes(blocks[k], file)
				if err != nil {
					return err
				}
				err = Unmarshal(s, trial, nextStruct, ref, labels...)
				if err != nil {
					return err
				}
				knd := typ.Elem().Kind() // units' kind in hash or array
				if typ.Kind() == reflect.Map {
					strKey := reflect.ValueOf(labels[0])
					if knd == reflect.Interface || knd == reflect.Ptr {
						fMap.SetMapIndex(strKey, reflect.ValueOf(trial))
					} else {
						fMap.SetMapIndex(strKey, reflect.ValueOf(trial).Elem())
					}
				} else {
					if knd == reflect.Interface || knd == reflect.Ptr {
						fSlice.Index(k).Set(reflect.ValueOf(trial))
					} else {
						fSlice.Index(k).Set(reflect.ValueOf(trial).Elem())
					}
				}
			}
			if typ.Kind() == reflect.Map {
				f.Set(fMap)
			} else {
				f.Set(fSlice)
			}
		} else if x := result.GetSingleStruct(); x != nil {
			trial := ref[x.ClassName]
			if trial == nil {
				return fmt.Errorf("class ref not found for %s", x.ClassName)
			}
			trial = clone(trial)
			s, labels, err := getBlockBytes(blocks[0], file)
			if err != nil {
				return err
			}
			err = Unmarshal(s, trial, x, ref, labels...)
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

func getBlockBytes(block *hclsyntax.Block, file *hcl.File) ([]byte, []string, error) {
	if block == nil {
		return nil, nil, fmt.Errorf("block not found")
	}
	rng1 := block.OpenBraceRange
	rng2 := block.CloseBraceRange
	bs := file.Bytes[rng1.End.Byte:rng2.Start.Byte]
	return bs, block.Labels, nil
}

func getTagref(origTypes []reflect.StructField) map[string]bool {
	tagref := make(map[string]bool)
	for _, field := range origTypes {
		two := tag2(field.Tag)
		tagref[two[0]] = true
	}
	return tagref
}

// newFields for normal fields which can be decoded withe gohcl
// origTypes for interface which needs decoded individually as body
func loopFields(t reflect.Type, objectMap map[string]*Value, ref map[string]interface{}) ([]reflect.StructField, []reflect.StructField, error) {
	var newFields []reflect.StructField
	var origTypes []reflect.StructField
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		typ := field.Type
		name := field.Name
		if !unicode.IsUpper([]rune(name)[0]) {
			continue
		}
		tcontent := (tag2(field.Tag))[0]
		if tcontent == `-` || (len(tcontent) >= 2 && tcontent[len(tcontent)-2:] == `,-`) {
			continue
		}
		if _, ok := objectMap[name]; ok {
			two := tag2(field.Tag)
			field.Tag = reflect.StructTag("hcl:\"" + two[0] + ",block\"")
			origTypes = append(origTypes, field)
			continue
		}
		if tcontent == "" {
			switch typ.Kind() {
			case reflect.Interface:
				continue
			case reflect.Pointer, reflect.Struct:
				var deeps, deepTypes []reflect.StructField
				var err error
				if typ.Kind() == reflect.Pointer {
					deeps, deepTypes, err = loopFields(field.Type.Elem(), objectMap, ref)
				} else {
					deeps, deepTypes, err = loopFields(field.Type, objectMap, ref)
				}
				if err != nil {
					return nil, nil, err
				}
				for _, v := range deeps {
					newFields = append(newFields, v)
				}
				for _, v := range deepTypes {
					origTypes = append(origTypes, v)
				}
			default:
			}
			continue
		} else if typ.Kind() == reflect.Map { // for non-interface map
			eType := typ.Elem()
			s := eType.String()
			if eType.Kind() == reflect.Pointer || eType.Kind() == reflect.Interface {
				ref[s] = reflect.New(eType.Elem()).Interface()
			} else if eType.Kind() == reflect.Struct {
				ref[s] = reflect.New(eType).Interface()
			} else {
				field.Tag = addOptional(field.Tag)
				newFields = append(newFields, field)
				continue
			}
			v, err := NewValue([]string{s})
			if err != nil {
				return nil, nil, err
			}
			objectMap[field.Name] = v
			two := tag2(field.Tag)
			field.Tag = reflect.StructTag("hcl:\"" + two[0] + ",block\"")
			origTypes = append(origTypes, field)
			continue
		} else {
			field.Tag = addOptional(field.Tag)
			newFields = append(newFields, field)
		}
	}
	return newFields, origTypes, nil
}

func unplain(bs []byte, object interface{}, labels ...string) error {
	// start to investigate map of non-interface
	// see the same code in loopFields
	t := reflect.TypeOf(object).Elem()
	spec := make(map[string]interface{})
	ref := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		typ := field.Type
		if typ.Kind() != reflect.Map {
			continue
		}
		eType := typ.Elem()
		s := eType.String()
		if eType.Kind() == reflect.Pointer || eType.Kind() == reflect.Interface {
			ref[s] = reflect.New(eType.Elem()).Interface()
		} else if eType.Kind() == reflect.Struct {
			ref[s] = reflect.New(eType).Interface()
		} else {
			continue
		}
		spec[field.Name] = []string{s}
	}
	if len(spec) != 0 {
		tr, err := NewStruct(t.Name(), spec)
		if err != nil {
			return nil
		}
		return Unmarshal(bs, object, tr, ref, labels...)
	}
	// end map of non-interface

	err := hclsimple.Decode(rname(), bs, nil, object)
	if err != nil {
		return err
	}
	addLables(object, labels...)
	return nil
}

func addLables(current interface{}, label_values ...string) {
	if label_values == nil {
		return
	}
	m := len(label_values)
	k := 0

	t := reflect.TypeOf(current).Elem()
	n := t.NumField()

	oriValue := reflect.ValueOf(&current).Elem()
	tmp := reflect.New(oriValue.Elem().Type()).Elem()
	tmp.Set(oriValue.Elem())

	for i := 0; i < n; i++ {
		field := t.Field(i)
		f := tmp.Elem().Field(i)
		two := tag2(field.Tag)
		if strings.ToLower(two[1]) == "label" && k < m {
			f.Set(reflect.ValueOf(label_values[k]))
			k++
		}
	}
	oriValue.Set(tmp)
}

func addOptional(old reflect.StructTag) reflect.StructTag {
	two := tag2(old)
	if two[0] == "" {
		return reflect.StructTag(two[0])
	} else if two[1] == "" {
		two[1] = "optional"
	}
	return reflect.StructTag("hcl:\"" + two[0] + "," + two[1] + "\"")
}
