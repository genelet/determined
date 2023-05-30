package dethcl

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"reflect"
	"strings"
	"unicode"
)

/*
func Unmarshal(dat []byte, current interface{}, rest ...interface{}) error {
	if rest == nil {
		return unplain(dat, current)
	}

	var spec *Struct
	var ref map[string]interface{}
	var label_values []string
	var ok bool

	switch t := rest[0].(type) {
	case *Struct:
		spec = t
		if len(rest) < 2 {
			return fmt.Errorf("missing object reference")
		}
		ref, ok = rest[1].(map[string]interface{})
		if !ok {
			return fmt.Errorf("wrong object reference")
		}
		if len(rest) > 2 {
			for i := 2; i < len(rest); i++ {
				v, ok := rest[i-2].(string)
				if !ok {
					return fmt.Errorf("label is not string")
				}
				label_values = append(label_values, v)
			}
		}
	case string:
		for _, item := range rest {
			v, ok := item.(string)
			if !ok {
				return fmt.Errorf("label is not string")
			}
			label_values = append(label_values, v)
		}
	default:
		return fmt.Errorf("wrong input data type")
	}
	return unmarshal(dat, current, spec, ref, label_values...)
}
*/

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
	oriValue := reflect.ValueOf(&current).Elem()
	tmp := reflect.New(oriValue.Elem().Type()).Elem()
	tmp.Set(oriValue.Elem())

	n := t.NumField()

	var newFields []reflect.StructField
	tagref := make(map[string]bool)
	for i := 0; i < n; i++ {
		field := t.Field(i)
		name := field.Name
		if unicode.IsUpper([]rune(name)[0]) && field.Tag == "" {
			return fmt.Errorf("missing tag for %s", name)
		}
		if _, ok := objectMap[name]; ok {
			two := tag2(field.Tag)
			tagref[two[0]] = true
		} else {
			newFields = append(newFields, field)
		}
	}
	if newFields != nil && len(newFields) == n {
		return unplain(dat, current, label_values...)
	}

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

	m := 0
	if label_values != nil {
		m = len(label_values)
	}
	k := 0

	j := 0
	for i := 0; i < n; i++ {
		field := t.Field(i)
		fieldType := field.Type
		name := field.Name
		two := tag2(field.Tag)
		f := tmp.Elem().Field(i)
		result, ok := objectMap[name]
		if ok {
			two := tag2(field.Tag)
			blocks := blockref[two[0]]
			if x := result.GetListStruct(); x != nil {
				nextListStructs := x.GetListFields()
				n := len(nextListStructs)
				if n == 0 {
					return fmt.Errorf("missing list struct for %s", name)
				}

				var fSlice, fMap reflect.Value
				if fieldType.Kind() == reflect.Map {
					fMap = reflect.MakeMapWithSize(fieldType, n)
				} else {
					fSlice = reflect.MakeSlice(fieldType, n, n)
				}
				for k := 0; k < n; k++ {
					nextStruct := nextListStructs[k]
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
					if fieldType.Kind() == reflect.Map {
						fMap.SetMapIndex(reflect.ValueOf(labels[0]), reflect.ValueOf(trial))
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
		} else if unicode.IsUpper([]rune(name)[0]) {
			if strings.ToLower(two[1]) == "label" && k < m {
				f.Set(reflect.ValueOf(label_values[k]))
				k++
			} else {
				rawField := rawValue.Field(j)
				j++
				f.Set(rawField)
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
