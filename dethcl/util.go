package dethcl

import (
	"fmt"
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

func hcltag(tag reflect.StructTag) []byte {
	two := tag2(tag)
	return []byte(two[0])
}

func rname() string {
	return fmt.Sprintf("%d.hcl", rand.Int())
}
