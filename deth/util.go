package deth

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

/*
func marshal(object interface{}, is_label ...bool) ([]byte, error) {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(object, f.Body())
	if is_label == nil || is_label[0] == false {
		return f.Bytes(), nil
	}

	str := strings.ReplaceAll(string(f.Bytes()), "\n", "\n\t")
	str = "{\n\t" + str[0:len(str)-1] + "}\n"
	arr := getLabels(object)
	if arr == nil {
		return []byte(str), nil
	}
	return []byte(strings.Join(arr, " ") + " " + str), nil
}

func getLabels(current interface{}) []string {
	t := reflect.TypeOf(current).Elem()
	n := t.NumField()

	oriValue := reflect.ValueOf(current).Elem()

	var labels []string
	for i := 0; i < n; i++ {
		field := t.Field(i)
		_, tag_type := tag2tag(field.Tag, field.Type.Kind(), false)
		if strings.ToLower(tag_type[1]) == "label" {
			oriField := oriValue.Field(i)
			labels = append(labels, oriField.Interface().(string))
		}
	}

	return labels
}
*/

func hcltag(tag reflect.StructTag) []byte {
	two := strings.SplitN(tag.Get("hcl"), ",", 2)
	return []byte(two[0])
}

func unmarshal(bs []byte, object interface{}, labels ...string) error {
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
