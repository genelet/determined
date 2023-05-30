package dethcl

import (
	"fmt"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"math/rand"
	"reflect"
	"strings"
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
	   		return unmarshal(bs, object, tr, ref, labels...)
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
