package dethcl

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"github.com/hashicorp/hcl/v2/hclsimple"
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
			tag = tag[5:len(tag)-1]
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
			tag = tag[5:len(tag)-1]
            two := strings.SplitN(tag, ",", 2)
			if ok {
				old = reflect.StructTag("hcl:\""+two[0]+",remain\"")
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
/* arr is empty, so the following comment code  does not work
    t := reflect.TypeOf(object).Elem()
    n := t.NumField()
    oriValue := reflect.ValueOf(object).Elem()

    ref  := make(map[string]interface{})
    spec := make(map[string]interface{})

    for i := 0; i < n; i++ {
        field := t.Field(i)
        rawField := oriValue.Field(i)   
        if field.Type.Kind() != reflect.Map {
            continue
        }
        var arr []string
        iter := rawField.MapRange()
        for iter.Next() {
            //k := iter.Key()
            v := iter.Value()
            s := v.Type().String()
            arr = append(arr, s)
            ref[s] = reflect.New(v.Type().Elem()).Elem().Addr().Interface()
        }
        spec[field.Name] = arr
	}
	if len(spec) != 0 {
    	tr, err := NewStruct(t.Name(), spec)
		if err != nil { return nil }
		return Unmarshal(bs, object, tr, ref, labels...)
	}
*/

	err := hclsimple.Decode(rname(), bs, nil, object)
	if err != nil { return err }
	addLables(object, labels...)
	return nil
}

func addLables(current interface{}, label_values ...string) {
	if label_values == nil { return }
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
