package dethcl

import (
	"fmt"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"math/rand"
	"reflect"
	"strings"
	"unicode"
)

// clone clones a value via pointer
func clone(old interface{}) interface{} {
	obj := reflect.New(reflect.TypeOf(old).Elem())
	oldVal := reflect.ValueOf(old).Elem()
	newVal := obj.Elem()
	for i := 0; i < oldVal.NumField(); i++ {
		newValField := newVal.Field(i)
		if newValField.CanSet() {
			newValField.Set(oldVal.Field(i))
		}
	}

	return obj.Interface()
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
			case reflect.Interface, reflect.Pointer, reflect.Struct:
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
		newFields = append(newFields, &marshalField{field, oriField, pass})
	}
	return newFields, nil
}

func getOutlier(field reflect.StructField, oriField reflect.Value) ([][3][]byte, error) {
	var empty [][3][]byte
	fieldTag := field.Tag

	switch field.Type.Kind() {
	case reflect.Interface, reflect.Pointer:
		newCurrent := oriField.Interface()
		bs, err := marshal(newCurrent, true)
		if err != nil {
			return nil, err
		}
		empty = append(empty, [3][]byte{hcltag(fieldTag), nil, bs})
	case reflect.Struct:
		newCurrent := oriField.Addr().Interface()
		bs, err := marshal(newCurrent, true)
		if err != nil {
			return nil, err
		}
		empty = append(empty, [3][]byte{hcltag(fieldTag), nil, bs})
	case reflect.Slice:
		n := oriField.Len()
		if n < 1 {
			return nil, nil
		}
		switch oriField.Index(0).Kind() {
		case reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.Struct:
			for i := 0; i < n; i++ {
				item := oriField.Index(i)
				bs, err := marshal(item.Interface(), true)
				if err != nil {
					return nil, err
				}
				empty = append(empty, [3][]byte{hcltag(fieldTag), nil, bs})
			}
		default:
		}
	case reflect.Map:
		n := oriField.Len()
		if n < 1 {
			return nil, nil
		}
		switch oriField.MapIndex(oriField.MapKeys()[0]).Kind() {
		case reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.Struct:
			iter := oriField.MapRange()
			for iter.Next() {
				k := iter.Key()
				v := iter.Value()
				bs, err := marshal(v.Interface(), true)
				if err != nil {
					return empty, err
				}
				empty = append(empty, [3][]byte{hcltag(fieldTag), []byte(k.String()), bs})
			}
		default:
		}
	default:
	}
	return empty, nil
}

func getTagref(origTypes []reflect.StructField) map[string]bool {
	tagref := make(map[string]bool)
	for _, field := range origTypes {
		two := tag2(field.Tag)
		tagref[two[0]] = true
	}
	return tagref
}

func loopFields(t reflect.Type, objectMap map[string]*Value) ([]reflect.StructField, []reflect.StructField, error) {
	var newFields []reflect.StructField
	var origTypes []reflect.StructField
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		typ := field.Type
		name := field.Name
		if !unicode.IsUpper([]rune(name)[0]) {
			continue
		}
		two := tag2(field.Tag)
		tcontent := two[0]
		if tcontent == "" {
			switch typ.Kind() {
			case reflect.Interface, reflect.Pointer, reflect.Struct:
				deeps, deepTypes, err := loopFields(field.Type, objectMap)
				if err != nil {
					return nil, nil, err
				}
				for _, v := range deeps {
					newFields = append(newFields, v)
				}
				for _, v := range deepTypes {
					origTypes = append(origTypes, v)
				}
			default:
			}
			continue
		}
		if tcontent == `-` || (len(tcontent) >= 2 && tcontent[len(tcontent)-2:] == `,-`) {
			continue
		}
		if _, ok := objectMap[name]; ok {
			origTypes = append(origTypes, field)
		} else {
			newFields = append(newFields, field)
		}
	}
	return newFields, origTypes, nil
}

func tag2(old reflect.StructTag) [2]string {
	for _, tag := range strings.Fields(string(old)) {
		if len(tag) >= 5 && strings.ToLower(tag[:5]) == "hcl:\"" {
			tag = tag[5 : len(tag)-1]
			two := strings.SplitN(tag, ",", 2)
			if len(two) == 2 {
				return [2]string{two[0], two[1]}
			}
			return [2]string{two[0], ""}
		}
	}
	return [2]string{}
}

func tag2tag(old reflect.StructTag, kind reflect.Kind, ok bool) (reflect.StructTag, [2]string) {
	for _, tag := range strings.Fields(string(old)) {
		if len(tag) >= 5 && strings.ToLower(tag[:5]) == "hcl:\"" {
			tag = tag[5 : len(tag)-1]
			two := strings.SplitN(tag, ",", 2)
			if ok {
				old = reflect.StructTag("hcl:\"" + two[0] + ",remain\"")
			}
			if len(two) == 1 {
				return old, [2]string{two[0], ""}
			}
			if kind == reflect.Map {
				return old, [2]string{two[0], "hash"}
			}
			return old, [2]string{two[0], two[1]}
		}
	}
	return old, [2]string{}
}

func rname() string {
	return fmt.Sprintf("%d.hcl", rand.Int())
}

func hcltag(tag reflect.StructTag) []byte {
	two := strings.SplitN(tag.Get("hcl"), ",", 2)
	return []byte(two[0])
}

func unplain(bs []byte, object interface{}, labels ...string) error {
	/* we may make unplan(bs, object, ref, labels...) working with the hack
	       t := reflect.TypeOf(object).Elem()
	       spec := make(map[string]interface{})

	   	if ref != nil {
	       for i := 0; i < t.NumField(); i++ {
	           field := t.Field(i)
	   		typ := field.Type
	           if typ.Kind() != reflect.Map || typ.Key().Kind() != reflect.String {
	               continue
	           }
	   		// typ.String() == `map[string]*dethcl.circle`
	   		for k, v := range ref {
	   			mapTyp := reflect.MapOf(reflect.TypeOf(string), reflect.TypeOf(v))
	   			if typ.Assignable.(mapTyp) {
	   				objectName = k
	   			}
	   		}
	   		spec[field.Name] = []string{objectName}
	   	}
	   	if len(spec) != 0 {
	       	tr, err := NewStruct(t.Name(), spec)
	   		if err != nil { return nil }
	   		return Unmarshal(bs, object, tr, ref, labels...)
	   	}
	   	}
	*/

	err := hclsimple.Decode(rname(), bs, nil, object)
	if err != nil {
		return err
	}
	addLables(object, labels...)
	return nil
}

func addLables(current interface{}, label_values ...string) {
	if label_values == nil {
		return
	}
	m := len(label_values)
	k := 0

	t := reflect.TypeOf(current).Elem()
	n := t.NumField()

	oriValue := reflect.ValueOf(&current).Elem()
	tmp := reflect.New(oriValue.Elem().Type()).Elem()
	tmp.Set(oriValue.Elem())

	for i := 0; i < n; i++ {
		field := t.Field(i)
		f := tmp.Elem().Field(i)
		_, tag_type := tag2tag(field.Tag, field.Type.Kind(), false)
		if strings.ToLower(tag_type[1]) == "label" && k < m {
			f.Set(reflect.ValueOf(label_values[k]))
			k++
		}
	}
	oriValue.Set(tmp)
}
