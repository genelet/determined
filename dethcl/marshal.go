package dethcl

import (
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"reflect"
	"strings"
	"unicode"
)

// Marshal marshals object into HCL string
func Marshal(current interface{}) ([]byte, error) {
	rv := reflect.ValueOf(current)
	if rv.IsValid() && rv.IsZero() {
		return nil, nil
	}

	switch rv.Kind() {
	case reflect.Pointer, reflect.Struct:
		return marshal(current, false)
	default:
	}
	return encode(current)
}

func marshal(current interface{}, is ...bool) ([]byte, error) {
	t := reflect.TypeOf(current)
	oriValue := reflect.ValueOf(current)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
		oriValue = oriValue.Elem()
	}
	if t.Kind() != reflect.Struct {
		return Marshal(current)
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
			outlier, err := getOutlier(field, oriField)
			if err != nil {
				return nil, err
			}
			outliers = append(outliers, outlier...)
		} else {
			fieldTag := field.Tag
			hcl := tag2(fieldTag)
			if hcl[1] == "label" {
				labels = append(labels, oriField.Interface().(string))
				k++
				continue
			}
			tmp.Field(k).Set(oriField)
			k++
		}
	}

	f := hclwrite.NewEmptyFile()
	// use tmp.Addr().Interface() as the constructed object
	gohcl.EncodeIntoBody(tmp.Addr().Interface(), f.Body())
	bs := f.Bytes()

	blank := []byte(" ")
	nl := []byte("\n")
	for _, item := range outliers {
		bs = append(bs, (item.b0)...)
		bs = append(bs, blank...)
		if item.encode {
			bs = append(bs, ([]byte("="))...)
			bs = append(bs, blank...)
		}
		if item.b1 != nil {
			bs = append(bs, (item.b1)...)
			bs = append(bs, blank...)
		}
		bs = append(bs, (item.b2)...)
		bs = append(bs, nl...)
	}

	if is == nil || is[0] == false || bs == nil {
		return bs, nil
	}

	str := strings.ReplaceAll(string(bs), "\n", "\n  ")
	str = "{\n  " + str[0:len(str)-2] + "}\n"
	if labels == nil {
		return []byte(str), nil
	}
	return []byte("\"" + strings.Join(labels, "\" \"") + "\" " + str), nil
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
		if tcontent == "" {
			switch typ.Kind() {
			case reflect.Ptr:
				mfs, err := getFields(typ.Elem(), oriField.Elem())
				if err != nil {
					return nil, err
				}
				for _, v := range mfs {
					newFields = append(newFields, v)
				}
			case reflect.Struct:
				mfs, err := getFields(typ, oriField)
				if err != nil {
					return nil, err
				}
				for _, v := range mfs {
					newFields = append(newFields, v)
				}
			default:
			}
			continue
		}
		if tcontent == `-` || (len(tcontent) >= 2 && tcontent[len(tcontent)-2:] == `,-`) {
			continue
		}
		pass := false
		switch typ.Kind() {
		case reflect.Interface, reflect.Pointer, reflect.Struct:
			pass = true
		case reflect.Slice:
			if oriField.Len() == 0 {
				continue
			}
			switch oriField.Index(0).Kind() {
			case reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.Struct:
				pass = true
			default:
			}
		case reflect.Map:
			if oriField.Len() == 0 {
				continue
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
		newFields = append(newFields, &marshalField{field, oriField, pass})
	}
	return newFields, nil
}

type marshalOut struct {
	b0     []byte
	b1     []byte
	b2     []byte
	encode bool
}

func getOutlier(field reflect.StructField, oriField reflect.Value) ([]*marshalOut, error) {
	var empty []*marshalOut
	fieldTag := field.Tag
	typ := field.Type

	switch typ.Kind() {
	case reflect.Interface, reflect.Pointer:
		newCurrent := oriField.Interface()
		bs, err := marshal(newCurrent, true)
		if err != nil {
			return nil, err
		}
		if bs == nil {
			return nil, nil
		}
		empty = append(empty, &marshalOut{hcltag(fieldTag), nil, bs, false})
	case reflect.Struct:
		newCurrent := oriField.Addr().Interface()
		bs, err := marshal(newCurrent, true)
		if err != nil {
			return nil, err
		}
		if bs == nil {
			return nil, nil
		}
		empty = append(empty, &marshalOut{hcltag(fieldTag), nil, bs, false})
	case reflect.Slice:
		n := oriField.Len()
		if n < 1 {
			return nil, nil
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
				bs, err := marshal(item.Interface(), true)
				if err != nil {
					return nil, err
				}
				if bs == nil {
					continue
				}
				empty = append(empty, &marshalOut{hcltag(fieldTag), nil, bs, false})
			}
		} else {
			bs, err := encode(oriField.Interface())
			if err != nil {
				return nil, err
			}
			empty = append(empty, &marshalOut{hcltag(fieldTag), nil, bs, true})
		}
	case reflect.Map:
		n := oriField.Len()
		if n < 1 {
			return nil, nil
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
				v := iter.Value()
				var bs []byte
				var err error
				bs, err = marshal(v.Interface(), true)
				if err != nil {
					return empty, err
				}
				if bs == nil {
					continue
				}
				empty = append(empty, &marshalOut{hcltag(fieldTag), []byte(k.String()), bs, false})
			}
		} else {
			bs, err := encode(oriField.Interface())
			if err != nil {
				return nil, err
			}
			empty = append(empty, &marshalOut{hcltag(fieldTag), nil, bs, true})
		}
	default:
	}
	return empty, nil
}
