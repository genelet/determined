package dethcl

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"reflect"
	"strings"
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
	newFields, origTypes, err := loopFields(t, objectMap)
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
	for _, field := range newFields {
		name := field.Name
		two := tag2(field.Tag)
		f := tmp.Elem().FieldByName(name)
		if strings.ToLower(two[1]) == "label" && k < m {
			f.Set(reflect.ValueOf(label_values[k]))
		} else {
			rawField := rawValue.Field(k)
			f.Set(rawField)
			k++
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
			n := len(nextListStructs)
			if n == 0 {
				return fmt.Errorf("missing list struct for %s", name)
			}

			var fSlice, fMap reflect.Value
			if typ.Kind() == reflect.Map {
				fMap = reflect.MakeMapWithSize(typ, n)
			} else {
				fSlice = reflect.MakeSlice(typ, n, n)
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
				if typ.Kind() == reflect.Map {
					fMap.SetMapIndex(reflect.ValueOf(labels[0]), reflect.ValueOf(trial))
				} else {
					fSlice.Index(k).Set(reflect.ValueOf(trial))
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
