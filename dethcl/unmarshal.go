package dethcl

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	ilang "github.com/genelet/determined/internal/lang"
	"github.com/genelet/determined/utils"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/zclconf/go-cty/cty/function"
)

type Unmarshaler interface {
	UnmarshalHCL([]byte, ...string) error
}

// Unmarshal decodes HCL data
//
//   - dat: Hcl data
//   - current: pointer of struct, []interface{} or map[string]interface{}
//   - optional labels: field values of labels
func Unmarshal(dat []byte, current interface{}, labels ...string) error {
	if current == nil {
		return nil
	}
	rv := reflect.ValueOf(current)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("non-pointer or nil data")
	}
	v, ok := current.(Unmarshaler)
	if ok {
		return v.UnmarshalHCL(dat, labels...)
	}
	return UnmarshalSpec(dat, current, nil, nil, labels...)
}

// UnmarshalSpec decodes HCL struct data with interface specifications.
//
//   - dat: Hcl data
//   - current: object as pointer
//   - spec: Determined for data specs
//   - ref: object map, with key object name and value new object
//   - optional labels: values of labels
func UnmarshalSpec(dat []byte, current interface{}, spec *utils.Struct, ref map[string]interface{}, labels ...string) error {
	if ref == nil {
		ref = make(map[string]interface{})
	}
	top := utils.NewTree(utils.VAR)
	node := top
	ref[utils.ATTRIBUTES] = top
	defaultFuncs := ilang.CoreFunctions(".")
	if ref[utils.FUNCTIONS] == nil {
		ref[utils.FUNCTIONS] = defaultFuncs
	} else {
		for k, v := range defaultFuncs {
			ref[utils.FUNCTIONS].(map[string]function.Function)[k] = v
		}
	}

	return UnmarshalSpecTree(node, dat, current, spec, ref, labels...)
}

// UnmarshalSpecTree decodes HCL struct data with interface specifications, at specifc tree node
//
//   - node: tree node
//   - dat: Hcl data
//   - current: object as pointer
//   - spec: Determined for data specs
//   - ref: object map, with key object name and value new object
//   - optional labels: values of labels
func UnmarshalSpecTree(node *utils.Tree, dat []byte, current interface{}, spec *utils.Struct, ref map[string]interface{}, labels ...string) error {
	rv := reflect.ValueOf(current)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("non-pointer or nil data")
	}
	rv = rv.Elem()

	switch rv.Kind() {
	case reflect.Map:
		obj, err := decodeMap(ref, node, dat)
		if err != nil {
			return err
		}
		x := current.(*map[string]interface{})
		for k, v := range obj {
			(*x)[k] = v
		}
		return nil
	case reflect.Slice:
		obj, err := decodeSlice(ref, node, dat)
		if err != nil {
			return err
		}
		x := current.(*[]interface{})
		for _, v := range obj {
			*x = append(*x, v)
		}
		return nil
	default:
	}

	t := reflect.TypeOf(current)
	if t.Kind() != reflect.Pointer {
		return fmt.Errorf("non-pointer or nil data")
	}
	t = t.Elem()

	var objectMap map[string]*utils.Value
	if spec != nil {
		objectMap = spec.GetFields()
	}
	if objectMap == nil {
		objectMap = make(map[string]*utils.Value)
	}

	file, diags := hclsyntax.ParseConfig(dat, rname(), hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return diags
	}
	bd := file.Body.(*hclsyntax.Body)

	for k, v := range bd.Attributes {
		cv, err := utils.ExpressionToCty(ref, node, v.Expr)
		if err != nil {
			return err
		}
		v.Expr = utils.CtyToExpression(cv, v.Range())
		node.AddItem(k, cv)
	}

	for _, block := range bd.Blocks {
		node.AddNodes(block.Type, block.Labels...)
	}

	newLabels, newFields, oriFields, decFields, err := loopFields(t, objectMap, ref)
	if err != nil {
		return err
	}

	labelExprs, rawValue, decattrs, decblock, oriblock, diags := refreshBody(bd, oriFields, decFields, newFields, newLabels)
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
	for _, field := range newLabels {
		name := field.Name
		f := oriTobe.Elem().FieldByName(name)
		var label string
		if labelExprs != nil && len(labelExprs) > k {
			cv, diags := labelExprs[k].Value(nil)
			if diags.HasErrors() {
				return diags
			}
			label = cv.AsString()
			if k < m {
				labels[k] = label // useless but if pass ...*string as labels
			}
		} else if k < m {
			label = labels[k]
		}
		f.Set(reflect.ValueOf(label))
		k++
	}

	for i, field := range newFields {
		name := field.Name
		f := oriTobe.Elem().FieldByName(name)
		rawField := rawValue.Field(i)
		f.Set(rawField)
	}

	for _, field := range decFields {
		var bs []byte
		var err error
		tag := (tag2(field.Tag))[0]
		if attr, ok := decattrs[tag]; ok {
			bs = file.Bytes[attr.EqualsRange.End.Byte:attr.SrcRange.End.Byte]
		} else if blkd, ok := decblock[tag]; ok { // not supposed to happen
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
			obj, err := decodeSlice(ref, node, bs)
			if err != nil {
				return err
			}
			f.Set(reflect.ValueOf(obj))
		} else {
			obj, err := decodeMap(ref, node, bs)
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
		result := objectMap[name]

		if x := result.GetMap2Struct(); x != nil {
			nextMap2Structs := x.GetMap2Fields()
			if typ.Kind() != reflect.Map {
				return fmt.Errorf("type mismatch for %s", name)
			}
			var first *utils.MapStruct
			var firstFirst *utils.Struct
			for _, first = range nextMap2Structs {
				for _, firstFirst = range first.GetMapFields() {
					break
				}
				break
			}
			n := len(blocks)
			fMap := reflect.MakeMapWithSize(typ, n)

			for k := 0; k < n; k++ {
				block := blocks[k]
				subnode := node.GetNode(tag, block.Labels...)
				keystring0 := block.Labels[0]
				var keystring1 string
				if len(block.Labels) > 1 {
					keystring1 = block.Labels[1]
				}
				nextMapStruct, ok := nextMap2Structs[keystring0]
				if !ok {
					nextMapStruct = first
				}
				nextStruct, ok := nextMapStruct.GetMapFields()[keystring1]
				if !ok {
					nextStruct = firstFirst
				}
				trial := ref[nextStruct.ClassName]
				if trial == nil {
					return fmt.Errorf("ref not found for %s", nextStruct.ClassName)
				}
				trial = clone(trial)
				s, lbls, err := getBlockBytes(block, file)
				if err != nil {
					return err
				}
				if len(lbls) > 2 {
					return fmt.Errorf("only two labels are allowed for map2 struct %s", name)
				}
				err = plusUnmarshalSpecTree(subnode, s, trial, nextStruct, ref, lbls...)
				if err != nil {
					return err
				}
				knd := typ.Elem().Kind() // units' kind in hash or array
				strKey := reflect.ValueOf([2]string{keystring0, keystring1})

				if knd == reflect.Interface || knd == reflect.Ptr {
					fMap.SetMapIndex(strKey, reflect.ValueOf(trial))
				} else {
					fMap.SetMapIndex(strKey, reflect.ValueOf(trial).Elem())
				}
			}
			f.Set(fMap)
		} else if x := result.GetMapStruct(); x != nil {
			nextMapStructs := x.GetMapFields()
			if typ.Kind() != reflect.Map {
				return fmt.Errorf("type mismatch for %s", name)
			}
			var first *utils.Struct
			for _, first = range nextMapStructs {
				break
			}
			n := len(blocks)
			fMap := reflect.MakeMapWithSize(typ, n)

			for k := 0; k < n; k++ {
				block := blocks[k]
				subnode := node.GetNode(tag, block.Labels...)
				keystring := block.Labels[0]
				nextStruct, ok := nextMapStructs[keystring]
				if !ok {
					nextStruct = first
				}

				trial := ref[nextStruct.ClassName]
				if trial == nil {
					return fmt.Errorf("ref not found for %s", nextStruct.ClassName)
				}
				trial = clone(trial)
				s, lbls, err := getBlockBytes(block, file)
				if err != nil {
					return err
				}
				if len(lbls) > 1 {
					return fmt.Errorf("only one label is allowed for map struct %s", name)
				}
				err = plusUnmarshalSpecTree(subnode, s, trial, nextStruct, ref, lbls...)
				if err != nil {
					return err
				}
				knd := typ.Elem().Kind() // units' kind in hash or array
				strKey := reflect.ValueOf(keystring)

				if knd == reflect.Interface || knd == reflect.Ptr {
					fMap.SetMapIndex(strKey, reflect.ValueOf(trial))
				} else {
					fMap.SetMapIndex(strKey, reflect.ValueOf(trial).Elem())
				}
			}
			f.Set(fMap)
		} else if x := result.GetListStruct(); x != nil {
			nextListStructs := x.GetListFields()
			nSmaller := len(nextListStructs)
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
				if k < nSmaller && (typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array) {
					nextStruct = nextListStructs[k]
					// map is only using the first struct
				}
				block := blocks[k]
				subnode := node.GetNode(tag, block.Labels...)
				trial := ref[nextStruct.ClassName]
				if trial == nil {
					return fmt.Errorf("class ref not found for %s", nextStruct.ClassName)
				}
				trial = clone(trial)
				s, lbls, err := getBlockBytes(block, file)
				if err != nil {
					return err
				}
				err = plusUnmarshalSpecTree(subnode, s, trial, nextStruct, ref, lbls...)
				if err != nil {
					return err
				}
				knd := typ.Elem().Kind() // units' kind in hash or array
				if typ.Kind() == reflect.Map {
					strKey := reflect.ValueOf(lbls[0])
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
			subnode := node.GetNode(tag, blocks[0].Labels...)
			trial := ref[x.ClassName]
			if trial == nil {
				return fmt.Errorf("class ref not found for %s", x.ClassName)
			}
			trial = clone(trial)
			s, lbls, err := getBlockBytes(blocks[0], file)
			if err != nil {
				return err
			}
			err = plusUnmarshalSpecTree(subnode, s, trial, x, ref, lbls...)
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

func plusUnmarshalSpecTree(subnode *utils.Tree, s []byte, trial interface{}, nextStruct *utils.Struct, ref map[string]interface{}, labels ...string) error {
	v, ok := trial.(Unmarshaler)
	if ok {
		return v.UnmarshalHCL(s, labels...)
	}
	return UnmarshalSpecTree(subnode, s, trial, nextStruct, ref, labels...)
}

func refreshBody(bd *hclsyntax.Body, oriFields, decFields, newFields, newLabels []reflect.StructField) ([]hclsyntax.Expression, reflect.Value, map[string]*hclsyntax.Attribute, map[string][]*hclsyntax.Block, map[string][]*hclsyntax.Block, hcl.Diagnostics) {
	body := &hclsyntax.Body{SrcRange: bd.SrcRange, EndRange: bd.EndRange}

	oriref := getTagref(oriFields)
	decref := getTagref(decFields)

	var labelExprs []hclsyntax.Expression
	var decattrs map[string]*hclsyntax.Attribute
	for k, v := range bd.Attributes {
		if decref[k] {
			if decattrs == nil {
				decattrs = make(map[string]*hclsyntax.Attribute)
			}
			decattrs[k] = v
		} else {
			var found bool
			for _, s := range newLabels {
				tags := tag2(s.Tag)
				if len(tags) == 2 && tags[0] == k && strings.ToLower(tags[1]) == "label" {
					found = true
					break
				}
			}
			if found {
				labelExprs = append(labelExprs, v.Expr)
			} else {
				if body.Attributes == nil {
					body.Attributes = make(map[string]*hclsyntax.Attribute)
				}
				body.Attributes[k] = v
			}
		}
	}

	oriblock := make(map[string][]*hclsyntax.Block)
	decblock := make(map[string][]*hclsyntax.Block)
	for _, block := range bd.Blocks {
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
		return nil, reflect.Zero(newType), nil, nil, nil, diags
	}

	return labelExprs, reflect.ValueOf(raw).Elem(), decattrs, decblock, oriblock, nil
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
func loopFields(t reflect.Type, objectMap map[string]*utils.Value, ref map[string]interface{}) ([]reflect.StructField, []reflect.StructField, []reflect.StructField, []reflect.StructField, error) {
	var newLabels []reflect.StructField
	var newFields []reflect.StructField
	var oriFields []reflect.StructField
	var decFields []reflect.StructField
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		typ := field.Type
		if typ.Kind() == reflect.Pointer {
			typ = typ.Elem()
		}
		name := field.Name
		if !unicode.IsUpper([]rune(name)[0]) {
			continue
		}
		tag := (tag2(field.Tag))[0]
		what := (tag2(field.Tag))[1]
		if strings.ToLower(what) == "label" {
			newLabels = append(newLabels, field)
			continue
		}
		if tag == `-` || (len(tag) >= 2 && tag[len(tag)-2:] == `,-`) {
			continue
		}
		if _, ok := objectMap[name]; ok {
			oriFields = append(oriFields, field)
			continue
		}

		if tag == "" {
			switch typ.Kind() {
			case reflect.Struct:
				_, deeps, deepTypes, deepDecs, err := loopFields(field.Type, objectMap, ref)
				if err != nil {
					return nil, nil, nil, nil, err
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
			continue
		}

		if typ.Kind() == reflect.Struct {
			s := typ.String()
			ref[s] = reflect.New(typ).Interface()
			v, err := utils.NewValue(s)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			objectMap[field.Name] = v
			oriFields = append(oriFields, field)
		} else if typ.Kind() == reflect.Map && typ.Key().Kind() == reflect.Array && typ.Key().Len() == 2 {
			// this is map[[2]string]string
			eType := typ.Elem()
			s := eType.String()

			switch eType.Kind() {
			case reflect.Struct:
				ref[s] = reflect.New(eType).Interface()
			case reflect.Pointer:
				ref[s] = reflect.New(eType.Elem()).Interface()
			case reflect.Interface:
				decFields = append(decFields, field)
				continue
			default:
				newFields = append(newFields, field)
				continue
			}
			// use 2 empty strings here as key, then firstFirst in unmarshaling as default
			v, err := utils.NewValue(map[[2]string]string{{"", ""}: s})
			if err != nil {
				return nil, nil, nil, nil, err
			}
			objectMap[field.Name] = v
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
				decFields = append(decFields, field)
				continue
			default:
				newFields = append(newFields, field)
				continue
			}

			v, err := utils.NewValue([]string{s})
			if err != nil {
				return nil, nil, nil, nil, err
			}
			objectMap[field.Name] = v
			oriFields = append(oriFields, field)
		} else {
			newFields = append(newFields, field)
		}
	}
	return newLabels, newFields, oriFields, decFields, nil
}

func addLabels(current interface{}, labels ...string) {
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
