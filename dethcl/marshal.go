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

	n := t.NumField()
	var newFields []reflect.StructField

	for i := 0; i < n; i++ {
		field := t.Field(i)
		oriField := oriValue.Field(i)
		pass := false
		switch field.Type.Kind() {
		case reflect.Interface, reflect.Pointer, reflect.Struct:
				pass = true
		case reflect.Slice:
			switch oriField.Index(0).Kind() {
			case reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.Struct:
				pass = true
			default:
			}
		case reflect.Map:
			switch oriField.MapIndex(oriField.MapKeys()[0]).Kind() {
			case reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.Struct:
				pass = true
			default:
			}
		default:
		}
		if !pass {
			if _, ok := field.Tag.Lookup("hcl"); ok {
				newFields = append(newFields, field)
			}
		}
	}

	newType := reflect.StructOf(newFields)
	tmp := reflect.New(newType).Elem()
	var outliers [][3][]byte
	var labels []string

	k := 0
	for i := 0; i < n; i++ {
		field := t.Field(i)
		fieldTag := field.Tag
		oriField := oriValue.Field(i)
		pass := false
		switch field.Type.Kind() {
		case reflect.Interface, reflect.Pointer:
			newCurrent := oriField.Interface()
			bs, err := marshal(newCurrent, true)
			if err != nil { return nil, err }
			outliers = append(outliers, [3][]byte{hcltag(fieldTag), nil, bs})	
			pass = true	
		case reflect.Struct:
			newCurrent := oriField.Addr().Interface()
			bs, err := marshal(newCurrent, true)
			if err != nil { return nil, err }
			outliers = append(outliers, [3][]byte{hcltag(fieldTag), nil, bs})	
			pass = true	
		case reflect.Slice:
			switch oriField.Index(0).Kind() {
			case reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.Struct:
				pass = true	
				n := oriField.Len()
				for i:=0; i<n; i++ {
					item := oriField.Index(i)
					bs, err := marshal(item.Interface(), true)
					if err != nil { return nil, err }
					outliers = append(outliers, [3][]byte{hcltag(fieldTag), nil, bs})	
				}
			default:
			}
		case reflect.Map:
			switch oriField.MapIndex(oriField.MapKeys()[0]).Kind() {
			case reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.Struct:
				pass = true
				iter := oriField.MapRange()
				for iter.Next() {
					k := iter.Key()
					v := iter.Value()
					bs, err := marshal(v.Interface(), true)
					if err != nil { return nil, err }
					outliers = append(outliers, [3][]byte{hcltag(fieldTag), []byte(k.String()), bs})
				}
			default:
			}
		default:
		}
		if !pass {
			hcl := tag2(fieldTag)
			if hcl[1] == "label" {
				labels = append(labels, oriField.Interface().(string))
				continue
			}
			tmp.Field(k).Set(oriValue.Field(i))
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

	if is == nil || is[0] == false {
        return bs, nil
    }

	str := strings.ReplaceAll(string(bs), "\n", "\n\t")
    str = "{\n\t" + str[0:len(str)-1] + "}\n"
    if labels == nil {
        return []byte(str), nil
    }
    return []byte(strings.Join(labels, " ") + " " + str), nil
}
