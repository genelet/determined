package det

import (
"github.com/zclconf/go-cty/cty"

	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcl/v2/hclsyntax"
//	"github.com/hashicorp/hcl/v2/hclwrite"
	"math/rand"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	"reflect"
	"strings"
	"unicode"
)

type Hash struct {
	//A map[string]hcl.Block `cty:",remain"`
	A map[string]cty.Value `hcl:",remain"`
}
type List struct {
	//A []map[string]hcl.Body `hcl:",remain"`
	A []cty.Value `hcl:",remain"`
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

	tagref := make(map[string]string)
	var newFields []reflect.StructField
	found := false
	for i := 0; i < n; i++ {
		field := t.Field(i)
		name := field.Name
		if unicode.IsUpper([]rune(name)[0]) && field.Tag == "" {
			return fmt.Errorf("missing tag for %s", name)
		}
		if result, ok := objectMap[name]; ok {
			tag, hcltag := tag2tag(field.Tag)
			tagref[name] = hcltag
fmt.Printf("ddddd  %v\n", tagref)
			newField := reflect.StructField{Name: name, Tag: tag}
			if result.GetMapStruct() != nil {
					//var victimBody map[string]hcl.Body
					var victimBody *Hash
					newField.Type = reflect.TypeOf(&victimBody).Elem()
			} else if result.GetListStruct() != nil {
					var victimBody *List
					newField.Type = reflect.TypeOf(&victimBody).Elem()
			} else {
					var victimBody hcl.Body
					newField.Type = reflect.TypeOf(&victimBody).Elem()
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
fmt.Printf("1000: %#v\n", raw)
fmt.Printf("1001: %s\n", dat)
		file, diags := hclsyntax.ParseConfig(dat, fmt.Sprintf("%d.hcl", rand.Int()), hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() { return diags }
fmt.Printf("1002\n")
		diags = gohcl.DecodeBody(file.Body, nil, raw)
fmt.Printf("1003: %v\n", diags)
		if diags.HasErrors() { return diags }
fmt.Printf("2222222222: %#v\n", raw)
	rawValue := reflect.ValueOf(raw).Elem()
fmt.Printf("33333: %#v\n", rawValue)

	oriValue := reflect.ValueOf(&current).Elem()
	tmp := reflect.New(oriValue.Elem().Type()).Elem()
	tmp.Set(oriValue.Elem())

	for i := 0; i < n; i++ {
		field := t.Field(i)
		fieldType := field.Type
		fieldName := field.Name
		f := tmp.Elem().Field(i)
		rawField := rawValue.Field(i)
fmt.Printf("44444444: %T=> %#v\n", rawField, rawField)
		result, ok := objectMap[fieldName]
		if ok {
			if x := result.GetMapStruct(); x != nil {
tmp := rawField.Interface()
tmp1 := tmp.(*Hash).A
fmt.Printf("55555 %#v\n", tmp1)
//debugBody(tmp.(*Hash).A.(*hclsyntax.Body))
fmt.Printf("55555\n")
				nextMapStructs := x.GetMapFields()
				nSmaller := len(nextMapStructs)
				if nSmaller == 0 {
					return fmt.Errorf("missing map struct for %s", fieldName)
				}
				var first *Struct
				for _, first = range nextMapStructs {
					break
				}

for k, v := range tmp1 {
fmt.Printf("666666 %s => %#v\n", k, v)
fmt.Printf("666666 %s\n", v.GoString())
//debugBody(tmp.(*Hash).A.(*hclsyntax.Body))
fmt.Printf("666666\n")
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
					s := getBytes(v, tagref[fieldName], file)
					err := HclUnmarshal(s, trial, nextStruct, ref)
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
					return fmt.Errorf("missing list struct for %s", fieldName)
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
					s := getBytes(v, tagref[fieldName], file)
					err := HclUnmarshal(s, trial, nextStruct, ref)
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
				s := getBytes(rawField, tagref[fieldName], file)
fmt.Printf("bbbbbb %s\n", s)
fmt.Printf("bbbbbb %#v\n", trial)
fmt.Printf("bbbbbb %#v\n", x)
				err := HclUnmarshal(s, trial, x, ref)
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
		} else if unicode.IsUpper([]rune(fieldName)[0]) {
			f.Set(rawField)
		}
	}

	oriValue.Set(tmp)

	return nil
}

func tag2tag(old reflect.StructTag) (reflect.StructTag, string) {
		for _, tag := range strings.Fields(string(old)) {
			if len(tag) >= 5 && strings.ToLower(tag[:5]) == "hcl:\"" {
				two := strings.SplitN(tag, ",", 2)
				//return reflect.StructTag(two[0]+",remain\""), two[0][5:]
				return old, two[0][5:]
			} 
		}
	return old, ""
}

func debugBody(x *hclsyntax.Body) {
	fmt.Printf("700 Body %v\n", x)
	for k, v := range x.Attributes {
		c, d := v.Expr.Value(nil)
		fmt.Printf("701 Attr %s => %s => %#v\n", k, v.Name, v.Expr)
		fmt.Printf("702 ctyValue %#v => %s => %#v\n", c, c.GoString(), d)
	}
   	for _, block := range x.Blocks {
		fmt.Printf("703 block %#v\n", block)
	}
	fmt.Printf("704 range start %#v\n", x.SrcRange.Start)
	fmt.Printf("705 range end %#v\n", x.SrcRange.End)
	fmt.Printf("706 filename %#v\n", x.SrcRange.Filename)
	fmt.Printf("707 range %#v\n", x.SrcRange.String())
}

func getBytes(rawfield reflect.Value, tagname string, file *hcl.File) []byte {
		x := rawfield.Interface().(*hclsyntax.Body)
		debugBody(x)
		for _, block := range x.Blocks {
			if block.Type == tagname {
				rng1 := block.OpenBraceRange
				rng2 := block.CloseBraceRange
				bs := file.Bytes[rng1.End.Byte:rng2.Start.Byte]
fmt.Printf("801 %s\n", bs)
		return bs
			}
		}

	return rawfield.Bytes()
}
