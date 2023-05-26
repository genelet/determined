package deth

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
//	"github.com/hashicorp/hcl/v2/hclwrite"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	"reflect"
	"strings"
	"unicode"
)

// HclUnmarshal unmarshals HCL data with interfaces determined by Determined.
//
//   - dat: Hcl data
//   - current: object as interface
//   - endpoint: Determined
//   - ref: struct map, with key being string name and value pointer to struct
//   - optional label_values: fields' values of labels
func HclUnmarshal(dat []byte, current interface{}, endpoint *Struct, ref map[string]interface{}, label_values ...string) error {
	if endpoint == nil {
fmt.Printf("STOP 1\n")
		return unplain(dat, current, label_values...)
	}
	objectMap := endpoint.GetFields()
	if objectMap == nil || len(objectMap) == 0 {
fmt.Printf("STOP 2 endpoint: %s\n data: %s\n current %#v\n", endpoint.String(), dat, current)
		return unplain(dat, current, label_values...)
	}

	t := reflect.TypeOf(current).Elem()
	n := t.NumField()

	tag_types  := make(map[string][2]string) // field name to hcl tag type
	var newFields []reflect.StructField
	found := false
	for i := 0; i < n; i++ {
		field := t.Field(i)
		name := field.Name
		if unicode.IsUpper([]rune(name)[0]) && field.Tag == "" {
			return fmt.Errorf("missing tag for %s", name)
		}
		_, ok := objectMap[name]
		tag, tag_type := tag2tag(field.Tag, field.Type.Kind(), ok)
		if tag_type[1] != "" {
			tag_types[name] = tag_type
		}
		if ok {
			newField := reflect.StructField{Name: name, Tag: tag}
			var victimBody hcl.Body
			newField.Type = reflect.TypeOf(&victimBody).Elem()
			newFields = append(newFields, newField)
			found = true
		} else {
			newFields = append(newFields, field)
		}
	}
	if found == false {
fmt.Printf("STOP 3\n")
		return unplain(dat, current, label_values...)
	}

	newType := reflect.StructOf(newFields)
	raw := reflect.New(newType).Interface()
fmt.Printf("1001: %s\n", dat)
	file, diags := hclsyntax.ParseConfig(dat, rname(), hcl.Pos{Line:1,Column:1})
	if diags.HasErrors() { return diags }
	diags = gohcl.DecodeBody(file.Body, nil, raw)
	if diags.HasErrors() { return diags }
	rawValue := reflect.ValueOf(raw).Elem()
fmt.Printf("33333: %#v\n", rawValue)

	oriValue := reflect.ValueOf(&current).Elem()
	tmp := reflect.New(oriValue.Elem().Type()).Elem()
	tmp.Set(oriValue.Elem())

	m := 0
	if label_values != nil {
		m = len(label_values)
	}
	k := 0

	for i := 0; i < n; i++ {
		field := t.Field(i)
		fieldType := field.Type
		fieldName := field.Name
		f := tmp.Elem().Field(i)
		rawField := rawValue.Field(i)
fmt.Printf("44444444: %T=> %#v\n", rawField, rawField)
		result, ok := objectMap[fieldName]
		tag_type := tag_types[fieldName]
		if ok {
			body := rawField.Interface().(*hclsyntax.Body)
//debugBody(body, file)
			if x := result.GetListStruct(); x != nil {
				nextListStructs := x.GetListFields()
				n := len(nextListStructs)
				if n == 0 {
					return fmt.Errorf("missing list struct for %s", fieldName)
				}

				var fSlice, fMap reflect.Value
				if tag_type[1]=="hash" {
					fMap = reflect.MakeMapWithSize(fieldType, n)
				} else {
					fSlice = reflect.MakeSlice(fieldType, n, n)
				}
				for k := 0; k < n; k++ {
					nextStruct := nextListStructs[k]
					trial := ref[nextStruct.ClassName]
					if trial == nil {
						return protoimpl.X.NewError("ref not found for %s", fieldName)
					}
					trial = clone(trial)
					s, labels, err := getBytes(body.Blocks[k], file)
					if err != nil {
						return err
					}
					err = HclUnmarshal(s, trial, nextStruct, ref, labels...)
					if err != nil {
						return err
					}
					if tag_type[1]=="hash" {
						fMap.SetMapIndex(reflect.ValueOf(labels[0]), reflect.ValueOf(trial))
					} else {
						fSlice.Index(k).Set(reflect.ValueOf(trial))
					}
				}
				if tag_type[1]=="hash" {
					f.Set(fMap)
				} else {
					f.Set(fSlice)
				}
			} else if x := result.GetSingleStruct(); x != nil {
				trial := ref[x.ClassName]
				if trial == nil {
					return protoimpl.X.NewError("class ref not found for %s", x.ClassName)
				}
fmt.Printf("1111 bbbbbb %#v\n", trial)
				trial = clone(trial)
				s, labels, err := getBytes(body.Blocks[0], file)
				if err != nil {
					return err
				}
				err = HclUnmarshal(s, trial, x, ref, labels...)
				if err != nil {
					return err
				}
				if f.Kind() == reflect.Interface || f.Kind() == reflect.Ptr {
					f.Set(reflect.ValueOf(trial))
				} else {
					f.Set(reflect.ValueOf(trial).Elem())
				}
			}
		} else if unicode.IsUpper([]rune(fieldName)[0]) {
			if strings.ToLower(tag_type[1]) == "label" && k<m {
				f.Set(reflect.ValueOf(label_values[k]))
				k++
			} else {
				f.Set(rawField)
			}
		}
	}

	oriValue.Set(tmp)

	return nil
}

func debugBody(x *hclsyntax.Body, file *hcl.File) {
	fmt.Printf("700 Body %v\n", x)
	for k, v := range x.Attributes {
		c, d := v.Expr.Value(nil)
		fmt.Printf("701 Attr %s => %s => %#v\n", k, v.Name, v.Expr)
		fmt.Printf("702 ctyValue %#v => %s => %#v\n", c, c.GoString(), d)
	}
   	for _, block := range x.Blocks {
		fmt.Printf("703 block %#v\n", block)
				rng1 := block.OpenBraceRange
				rng2 := block.CloseBraceRange
				bs := file.Bytes[rng1.End.Byte:rng2.Start.Byte]
fmt.Printf("801 %s\n", bs)
	}
	fmt.Printf("704 range start %#v\n", x.SrcRange.Start)
	fmt.Printf("705 range end %#v\n", x.SrcRange.End)
	fmt.Printf("706 filename %#v\n", x.SrcRange.Filename)
	fmt.Printf("707 range %#v\n", x.SrcRange.String())
}

func getBytes(block *hclsyntax.Block, file *hcl.File) ([]byte, []string, error) {
	if block == nil {
		return nil, nil, protoimpl.X.NewError("block not found")
	}

	rng1 := block.OpenBraceRange
	rng2 := block.CloseBraceRange
	bs := file.Bytes[rng1.End.Byte:rng2.Start.Byte]
fmt.Printf("801 %s\n", bs)
	return bs, block.Labels, nil
}
