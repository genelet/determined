package dethcl

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Unmarshal decodes HCL data
//
//   - dat: Hcl data
//   - current: pointer of struct, []interface{} or map[string]interface{}
//   - optional labels: field values of labels
func Unmarshal(dat []byte, current interface{}, labels ...string) error {
	rv := reflect.ValueOf(current)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("non-pointer or nil data")
	}
	rv = rv.Elem()
	switch rv.Kind() {
	case reflect.Struct:
		return UnmarshalSpec(dat, current, nil, nil, labels...)
	case reflect.Map:
		x := current.(*map[string]interface{})
		obj, err := decodeMap(dat)
		if err != nil {
			return err
		}
		for k, v := range obj {
			(*x)[k] = v
		}
	case reflect.Slice:
		x := current.(*[]interface{})
		obj, err := decodeSlice(dat)
		if err != nil {
			return err
		}
		for _, v := range obj {
			*x = append(*x, v)
		}
	default:
		return fmt.Errorf("data type %v not supported", rv.Kind())
	}
	return nil
}

// UnmarshalSpec decodes HCL struct data with interface specifications.
//
//   - dat: Hcl data
//   - current: object as pointer
//   - spec: Determined for data specs
//   - ref: object map, with key object name and value new object
//   - optional labels: values of labels
func UnmarshalSpec(dat []byte, current interface{}, spec *Struct, ref map[string]interface{}, labels ...string) error {
	return unmarshalSpec(nil, dat, current, spec, ref, labels...)
}

func unmarshalSpec(node *Tree, dat []byte, current interface{}, spec *Struct, ref map[string]interface{}, labels ...string) error {
	t := reflect.TypeOf(current)
	if t.Kind() != reflect.Pointer {
		return fmt.Errorf("non-pointer or nil data")
	}
	t = t.Elem()

	var objectMap map[string]*Value
	if spec != nil {
		objectMap = spec.GetFields()
	}
	if objectMap == nil {
		objectMap = make(map[string]*Value)
	}
	if ref == nil {
		ref = make(map[string]interface{})
	}

	file, diags := hclsyntax.ParseConfig(dat, rname(), hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return diags
	}
	bd := file.Body.(*hclsyntax.Body)

	if ref[ATTRIBUTES] == nil || node == nil {
		top := NewTree(VAR)
		node = top
		ref[ATTRIBUTES] = top
	}

	var found bool
	for k, v := range bd.Attributes {
		cv, err := expressionToCty(ref, node, v.Expr)
		if err != nil {
			return err
		}
		found = true
		v.Expr = ctyToExpression(cv, v.Range())
		node.AddItem(k, cv)
	}

	newFields, oriFields, decFields, err := loopFields(t, objectMap, ref)
	if err != nil {
		return err
	}
	if oriFields == nil && decFields == nil && !found {
		return unplain(dat, current, labels...)
	}

	rawValue, decattrs, decblock, oriblock, diags := refreshBody(bd, oriFields, decFields, newFields)
	if diags.HasErrors() {
		return diags
	}

	oriValue := reflect.ValueOf(&current).Elem()
	oriTobe := reflect.New(oriValue.Elem().Type()).Elem()
	oriTobe.Set(oriValue.Elem())

	m := 0
	if labels != nil {
		m = len(labels)
	}
	k := 0
	for i, field := range newFields {
		two := tag2(field.Tag)
		name := field.Name
		f := oriTobe.Elem().FieldByName(name)
		if strings.ToLower(two[1]) == "label" && k < m {
			f.Set(reflect.ValueOf(labels[k]))
			k++
		} else {
			rawField := rawValue.Field(i)
			f.Set(rawField)
		}
	}

	for _, field := range decFields {
		var bs []byte
		var err error
		tag := (tag2(field.Tag))[0]
		if attr, ok := decattrs[tag]; ok {
			bs = file.Bytes[attr.EqualsRange.End.Byte:attr.SrcRange.End.Byte]
		} else if blkd, ok := decblock[tag]; ok {
			// this is not supposed to happen
			bs, _, err = getBlockBytes(blkd[0], file)
			if err != nil {
				return err
			}
		} else {
			continue
		}

		name := field.Name
		typ := field.Type
		f := oriTobe.Elem().FieldByName(name)
		if typ.Kind() == reflect.Slice {
			obj, err := decodeSlice(bs)
			if err != nil {
				return err
			}
			f.Set(reflect.ValueOf(obj))
		} else {
			obj, err := decodeMap(bs)
			if err != nil {
				return err
			}
			f.Set(reflect.ValueOf(obj))
		}
	}

	for _, field := range oriFields {
		tag := (tag2(field.Tag))[0]
		blocks := oriblock[tag]
		if blocks == nil || len(blocks) == 0 {
			continue
		}

		name := field.Name
		typ := field.Type
		f := oriTobe.Elem().FieldByName(name)
		subNode := node.AddNode(name)
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
				err = unmarshalSpec(subNode, s, trial, nextStruct, ref, labels...)
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
			err = unmarshalSpec(subNode, s, trial, x, ref, labels...)
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

	oriValue.Set(oriTobe)

	return nil
}

func refreshBody(bd *hclsyntax.Body, oriFields []reflect.StructField, decFields []reflect.StructField, newFields []reflect.StructField) (reflect.Value, map[string]*hclsyntax.Attribute, map[string][]*hclsyntax.Block, map[string][]*hclsyntax.Block, hcl.Diagnostics) {
	body := &hclsyntax.Body{SrcRange: bd.SrcRange, EndRange: bd.EndRange}

	oriref := getTagref(oriFields)
	decref := getTagref(decFields)

	var decattrs map[string]*hclsyntax.Attribute
	for k, v := range bd.Attributes {
		if decref[k] {
			if decattrs == nil {
				decattrs = make(map[string]*hclsyntax.Attribute)
			}
			decattrs[k] = v
		} else {
			if body.Attributes == nil {
				body.Attributes = make(map[string]*hclsyntax.Attribute)
			}
			body.Attributes[k] = v
		}
	}

	oriblock := make(map[string][]*hclsyntax.Block)
	decblock := make(map[string][]*hclsyntax.Block)
	for _, block := range bd.Blocks {
		// block.Type is the tag string
		tag := block.Type
		if oriref[tag] {
			oriblock[tag] = append(oriblock[tag], block)
		} else if decref[tag] {
			decblock[tag] = append(decblock[tag], block)
		} else {
			body.Blocks = append(body.Blocks, block)
		}
	}

	newType := reflect.StructOf(newFields)
	raw := reflect.New(newType).Interface()
	diags := gohcl.DecodeBody(body, nil, raw)
	if diags.HasErrors() {
		return reflect.Zero(newType), nil, nil, nil, diags
	}

	return reflect.ValueOf(raw).Elem(), decattrs, decblock, oriblock, nil
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

func getTagref(oriFields []reflect.StructField) map[string]bool {
	tagref := make(map[string]bool)
	for _, field := range oriFields {
		tag := (tag2(field.Tag))[0]
		tagref[tag] = true
	}
	return tagref
}

// newFields for normal fields, can be decoded withe gohcl
// oriFields for blocks, decoded individually as body
// decFields for map[string]interface{} or []interface{}
func loopFields(t reflect.Type, objectMap map[string]*Value, ref map[string]interface{}) ([]reflect.StructField, []reflect.StructField, []reflect.StructField, error) {
	var newFields []reflect.StructField
	var oriFields []reflect.StructField
	var decFields []reflect.StructField
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		typ := field.Type
		name := field.Name
		if !unicode.IsUpper([]rune(name)[0]) {
			continue
		}
		tag := (tag2(field.Tag))[0]
		if tag == `-` || (len(tag) >= 2 && tag[len(tag)-2:] == `,-`) {
			continue
		}
		if _, ok := objectMap[name]; ok {
			field.Tag = reflect.StructTag("hcl:\"" + tag + ",block\"")
			oriFields = append(oriFields, field)
			continue
		}
		if tag == "" {
			switch typ.Kind() {
			case reflect.Interface:
				continue
			case reflect.Pointer, reflect.Struct:
				var deeps, deepTypes, deepDecs []reflect.StructField
				var err error
				if typ.Kind() == reflect.Pointer {
					deeps, deepTypes, deepDecs, err = loopFields(field.Type.Elem(), objectMap, ref)
				} else {
					deeps, deepTypes, deepDecs, err = loopFields(field.Type, objectMap, ref)
				}
				if err != nil {
					return nil, nil, nil, err
				}
				for _, v := range deeps {
					newFields = append(newFields, v)
				}
				for _, v := range deepTypes {
					oriFields = append(oriFields, v)
				}
				for _, v := range deepDecs {
					decFields = append(decFields, v)
				}
			default:
			}
		} else if typ.Kind() == reflect.Struct || typ.Kind() == reflect.Pointer {
			eType := typ
			if typ.Kind() == reflect.Pointer {
				eType = typ.Elem()
			}
			s := eType.String()
			ref[s] = reflect.New(eType).Interface()
			v, err := NewValue(s)
			if err != nil {
				return nil, nil, nil, err
			}
			objectMap[field.Name] = v
			field.Tag = reflect.StructTag("hcl:\"" + tag + ",block\"")
			oriFields = append(oriFields, field)
		} else if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Map {
			eType := typ.Elem()
			s := eType.String()
			switch eType.Kind() {
			case reflect.Struct:
				ref[s] = reflect.New(eType).Interface()
			case reflect.Pointer:
				ref[s] = reflect.New(eType.Elem()).Interface()
			case reflect.Interface:
				field.Tag = addOptional(field.Tag)
				decFields = append(decFields, field)
				continue
			default:
				field.Tag = addOptional(field.Tag)
				newFields = append(newFields, field)
				continue
			}
			v, err := NewValue([]string{s})
			if err != nil {
				return nil, nil, nil, err
			}
			objectMap[field.Name] = v
			field.Tag = reflect.StructTag("hcl:\"" + tag + ",block\"")
			oriFields = append(oriFields, field)
			//		} else if typ.Kind()==reflect.Interface {
			//			field.Tag = addOptional(field.Tag)
			//			decFields = append(decFields, field)
		} else {
			field.Tag = addOptional(field.Tag)
			newFields = append(newFields, field)
		}
	}
	return newFields, oriFields, decFields, nil
}

func unplain(bs []byte, object interface{}, labels ...string) error {
	// start map of non-interface
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
		switch eType.Kind() {
		case reflect.Interface:
			return Unmarshal(bs, object, labels...)
		case reflect.Pointer:
			ref[s] = reflect.New(eType.Elem()).Interface()
		case reflect.Struct:
			ref[s] = reflect.New(eType).Interface()
		default:
			continue
		}
		spec[field.Name] = []string{s}
	}
	if len(spec) != 0 {
		tr, err := NewStruct(t.Name(), spec)
		if err != nil {
			return nil
		}
		return UnmarshalSpec(bs, object, tr, ref, labels...)
	}
	// end map of non-interface

	err := hclsimple.Decode(rname(), bs, nil, object)
	if err != nil {
		return err
	}
	addLables(object, labels...)
	return nil
}

func addLables(current interface{}, labels ...string) {
	if labels == nil {
		return
	}
	m := len(labels)
	k := 0

	t := reflect.TypeOf(current).Elem()
	n := t.NumField()

	oriValue := reflect.ValueOf(&current).Elem()
	oriTobe := reflect.New(oriValue.Elem().Type()).Elem()
	oriTobe.Set(oriValue.Elem())

	for i := 0; i < n; i++ {
		field := t.Field(i)
		f := oriTobe.Elem().Field(i)
		two := tag2(field.Tag)
		if strings.ToLower(two[1]) == "label" && k < m {
			f.Set(reflect.ValueOf(labels[k]))
			k++
		}
	}
	oriValue.Set(oriTobe)
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
