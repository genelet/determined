package dethcl

import (
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"reflect"
	"strings"
)

// Marshal marshals object into HCL string
func Marshal(object interface{}) ([]byte, error) {
	return marshal(object, false)
}

func marshal(current interface{}, is ...bool) ([]byte, error) {
	var t reflect.Type
	var oriValue reflect.Value
	switch current.(type) {
	case interface{}:
		t = reflect.TypeOf(current).Elem()
		oriValue = reflect.ValueOf(current).Elem()
	default:
		t = reflect.TypeOf(&current).Elem()
		oriValue = reflect.ValueOf(&current).Elem()
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
	var outliers [][3][]byte
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
		bs = append(bs, (item[0])...)
		bs = append(bs, blank...)
		if item[1] != nil {
			bs = append(bs, (item[1])...)
			bs = append(bs, blank...)
		}
		bs = append(bs, (item[2])...)
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
	return []byte(strings.Join(labels, " ") + " " + str), nil
}
