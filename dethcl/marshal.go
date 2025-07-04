// Package dethcl implements marshal and unmarshal between HCL string and go struct.
package dethcl

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type Marshaler interface {
	MarshalHCL() ([]byte, error)
}

// Marshal marshals object into HCL string
func Marshal(current interface{}) ([]byte, error) {
	if current == nil {
		return nil, nil
	}
	return MarshalLevel(current, 0)
}

func MarshalLevel(current interface{}, level int) ([]byte, error) {
	return marshalLevel(current, false, level)
}

func marshalLevel(current interface{}, equal bool, level int, keyname ...string) ([]byte, error) {
	rv := reflect.ValueOf(current)
	if rv.IsValid() && rv.IsZero() {
		return nil, nil
	}

	switch rv.Kind() {
	case reflect.Pointer, reflect.Struct:
		return marshal(current, level, keyname...)
	default:
	}

	return encoding(current, equal, level, keyname...)
}

func marshal(current interface{}, level int, keyname ...string) ([]byte, error) {
	if current == nil {
		return nil, nil
	}
	leading := strings.Repeat("  ", level+1)
	lessLeading := strings.Repeat("  ", level)

	if v, ok := current.(Marshaler); ok {
		bs, err := v.MarshalHCL()
		if err != nil {
			return nil, err
		}

		str := string(bs)
		if level > 0 {
			if isBlanck(bs) {
				str = "\n"
			} else {
				str = "\n" + str + "\n"
			}
		}
		str = strings.ReplaceAll(str, "\n", "\n"+lessLeading)
		if level > 0 {
			str = fmt.Sprintf("{%s%s}", leading, str)
		}
		return []byte(str), nil
	}

	t := reflect.TypeOf(current)
	oriValue := reflect.ValueOf(current)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
		oriValue = oriValue.Elem()
	}

	switch t.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		if oriValue.IsValid() {
			return []byte(fmt.Sprintf("= %v", oriValue.Interface())), nil
		}
		return nil, nil
	case reflect.String:
		if oriValue.IsValid() {
			return []byte(" = " + oriValue.String()), nil
		}
		return nil, nil
	case reflect.Pointer:
		return marshal(oriValue.Elem().Interface(), level, keyname...)
	default:
	}

	newFields, err := getFields(t, oriValue)
	if err != nil {
		return nil, err
	}

	var plains []reflect.StructField
	for _, mField := range newFields {
		if !mField.out {
			plains = append(plains, mField.field)
		}
	}
	newType := reflect.StructOf(plains)
	tmp := reflect.New(newType).Elem()
	var outliers []*marshalOut
	var labels []string

	k := 0
	for _, mField := range newFields {
		field := mField.field
		oriField := mField.value
		if mField.out {
			outlier, err := getOutlier(field, oriField, level)
			if err != nil {
				return nil, err
			}
			outliers = append(outliers, outlier...)
		} else {
			fieldTag := field.Tag
			hcl := tag2(fieldTag)
			if hcl[1] == "label" {
				label := oriField.Interface().(string)
				if keyname == nil || keyname[0] != label {
					labels = append(labels, label)
				}
				k++
				continue
			}
			tmp.Field(k).Set(oriField)
			k++
		}
	}

	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(tmp.Addr().Interface(), f.Body())
	bs := f.Bytes()

	str := string(bs)
	str = leading + strings.ReplaceAll(str, "\n", "\n"+leading)

	var lines []string
	for _, item := range outliers {
		line := string(item.b0) + " "
		if item.encode {
			line += "= "
		}
		if len(item.b1) > 0 {
			line += `"` + strings.Join(item.b1, `" "`) + `" `
		}
		line += string(item.b2)
		lines = append(lines, line)
	}
	if len(lines) > 0 {
		str += strings.Join(lines, "\n"+leading)
	}

	str = strings.TrimRight(str, " \t\n\r")
	if level > 0 { // not root
		str = fmt.Sprintf("{\n%s\n%s}", str, lessLeading)
		if labels != nil {
			str = "\"" + strings.Join(labels, "\" \"") + "\" " + str
		}
	}

	return []byte(str), nil
}

type marshalField struct {
	field reflect.StructField
	value reflect.Value
	out   bool
}

func getFields(t reflect.Type, oriValue reflect.Value) ([]*marshalField, error) {
	var newFields []*marshalField
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		typ := field.Type
		if !unicode.IsUpper([]rune(field.Name)[0]) {
			continue
		}
		oriField := oriValue.Field(i)
		two := tag2(field.Tag)
		tcontent := two[0]
		if tcontent == `-` || (len(tcontent) >= 2 && tcontent[len(tcontent)-2:] == `,-`) {
			continue
		}

		if field.Anonymous && tcontent == "" {
			switch typ.Kind() {
			case reflect.Ptr:
				mfs, err := getFields(typ.Elem(), oriField.Elem())
				if err != nil {
					return nil, err
				}
				newFields = append(newFields, mfs...)
			case reflect.Struct:
				mfs, err := getFields(typ, oriField)
				if err != nil {
					return nil, err
				}
				newFields = append(newFields, mfs...)
			default:
			}
			continue
		}

		// treat field of type pointer e.g. *map[string]*Example, the same as map[string]*Example
		if typ.Kind() == reflect.Pointer && (typ.Elem().Kind() == reflect.Slice || typ.Elem().Kind() == reflect.Map) {
			typ = typ.Elem()
			oriField = oriField.Elem()
			if !oriField.IsValid() {
				continue
			}
		}
		pass := false
		switch typ.Kind() {
		case reflect.Interface, reflect.Pointer, reflect.Struct:
			pass = true
		case reflect.Slice:
			if oriField.Len() == 0 {
				pass = true
				break
			}
			switch oriField.Index(0).Kind() {
			case reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.Struct:
				pass = true
			default:
			}
		case reflect.Map:
			if oriField.Len() == 0 {
				pass = true
				break
			}
			switch oriField.MapIndex(oriField.MapKeys()[0]).Kind() {
			case reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.Struct:
				pass = true
			default:
			}
		default:
			if oriField.IsValid() && oriField.IsZero() {
				continue
			}
		}
		if tcontent == "" {
			if pass {
				field.Tag = reflect.StructTag(fmt.Sprintf(`hcl:"%s,block"`, strings.ToLower(field.Name)))
			} else {
				field.Tag = reflect.StructTag(fmt.Sprintf(`hcl:"%s,optional"`, strings.ToLower(field.Name)))
			}
		}
		newFields = append(newFields, &marshalField{field, oriField, pass})
	}
	return newFields, nil
}

type marshalOut struct {
	b0     []byte
	b1     []string
	b2     []byte
	encode bool
}

func getOutlier(field reflect.StructField, oriField reflect.Value, level int) ([]*marshalOut, error) {
	var empty []*marshalOut
	fieldTag := field.Tag
	typ := field.Type
	newlevel := level + 1

	// treat ptr the same as the underlying type e.g. *Example, Example
	if typ.Kind() == reflect.Ptr && (typ.Elem().Kind() == reflect.Map || typ.Elem().Kind() == reflect.Slice) {
		typ = typ.Elem()
	}

	switch typ.Kind() {
	case reflect.Interface, reflect.Pointer:
		newCurrent := oriField.Interface()
		bs, err := MarshalLevel(newCurrent, newlevel)
		if err != nil {
			return nil, err
		}
		if isBlanck(bs) {
			return nil, nil
		}
		empty = append(empty, &marshalOut{hcltag(fieldTag), nil, bs, false})
	case reflect.Struct:
		var newCurrent interface{}
		if oriField.CanAddr() {
			newCurrent = oriField.Addr().Interface()
		} else {
			newCurrent = oriField.Interface()
		}
		bs, err := MarshalLevel(newCurrent, newlevel)
		if err != nil {
			return nil, err
		}
		if isBlanck(bs) {
			return nil, nil
		}
		empty = append(empty, &marshalOut{hcltag(fieldTag), nil, bs, false})
	case reflect.Slice:
		if oriField.IsNil() {
			return nil, nil
		}

		n := oriField.Len()
		if n < 1 {
			return []*marshalOut{{hcltag(fieldTag), nil, []byte(`[]`), true}}, nil
		}

		first := oriField.Index(0)
		var isLoop bool
		switch first.Kind() {
		case reflect.Pointer, reflect.Struct:
			isLoop = true
		case reflect.Interface:
			if first.Elem().Kind() == reflect.Pointer || first.Elem().Kind() == reflect.Struct {
				isLoop = true
			}
		default:
		}

		if isLoop {
			for i := 0; i < n; i++ {
				item := oriField.Index(i)
				bs, err := MarshalLevel(item.Interface(), newlevel)
				if err != nil {
					return nil, err
				}
				if isBlanck(bs) {
					continue
				}
				empty = append(empty, &marshalOut{hcltag(fieldTag), nil, bs, false})
			}
		} else {
			bs, err := MarshalLevel(oriField.Interface(), newlevel)
			if err != nil {
				return nil, err
			}
			if isBlanck(bs) {
				return nil, nil
			}
			empty = append(empty, &marshalOut{hcltag(fieldTag), nil, bs, true})
		}
	case reflect.Map:
		if oriField.IsNil() {
			return nil, nil
		}

		n := oriField.Len()
		if n < 1 {
			leading := strings.Repeat("  ", level+1)
			return []*marshalOut{{hcltag(fieldTag), nil, []byte("{\n" + leading + "}"), false}}, nil
		}

		first := oriField.MapIndex(oriField.MapKeys()[0])
		var isLoop bool
		switch first.Kind() {
		case reflect.Pointer, reflect.Struct:
			isLoop = true
		case reflect.Interface:
			if first.Elem().Kind() == reflect.Pointer || first.Elem().Kind() == reflect.Struct {
				isLoop = true
			}
		default:
		}

		if isLoop {
			iter := oriField.MapRange()
			for iter.Next() {
				k := iter.Key()
				var arr []string
				switch k.Kind() {
				case reflect.Array, reflect.Slice:
					for i := 0; i < k.Len(); i++ {
						item := k.Index(i)
						if !item.IsZero() {
							arr = append(arr, item.String())
						}
					}
				default:
					arr = append(arr, k.String())
				}

				v := iter.Value()
				var bs []byte
				var err error
				bs, err = marshal(v.Interface(), newlevel, arr...)
				if err != nil {
					return empty, err
				}
				if isBlanck(bs) {
					continue
				}
				empty = append(empty, &marshalOut{hcltag(fieldTag), arr, bs, false})
			}
		} else {
			bs, err := MarshalLevel(oriField.Interface(), newlevel)
			if err != nil {
				return nil, err
			}
			if isBlanck(bs) {
				return nil, nil
			}
			equal := true
			if typ.Elem().Kind() == reflect.Interface {
				equal = false
			}
			empty = append(empty, &marshalOut{hcltag(fieldTag), nil, bs, equal})
		}
	default:
	}
	return empty, nil
}

func isBlanck(bs []byte) bool {
	for _, b := range bs {
		if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
			return false
		}
	}
	return true
}
