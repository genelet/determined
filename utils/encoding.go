package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func isHashAll(item interface{}) (int, map[string]interface{}) {
	var found int
	next, ok := item.(map[string]interface{})
	if ok {
		found = 1
		for _, v := range next {
			_, ok1 := v.(map[string]interface{})
			if ok1 {
				found = 2
			} else {
				found = 1
				break
			}
		}
	}
	return found, next
}

func loopHash(arr *[]string, name string, item interface{}, level int) error {
	found, next := isHashAll(item)
	switch found {
	case 2:
		for key, value := range next {
			name2 := name + ` "` + key + `"`
			err := loopHash(arr, name2, value, level)
			if err != nil {
				return err
			}
		}
	case 1:
		bs, err := Encoding(item, level+1)
		if err != nil {
			return err
		}
		*arr = append(*arr, fmt.Sprintf("%s %s", name, bs))
	default:
		bs, err := Encoding(item, level+1)
		if err != nil {
			return err
		}
		*arr = append(*arr, fmt.Sprintf("%s = %s", name, bs))
	}
	return nil
}

func Encoding(current interface{}, level int) ([]byte, error) {
	var str string

	leading := strings.Repeat("  ", level+1)
	lessLeading := strings.Repeat("  ", level)

	rv := reflect.ValueOf(current)
	switch rv.Kind() {
	case reflect.Map:
		var arr []string
		for name, item := range current.(map[string]interface{}) {
			/* // This is the old code for the this "for" loop
			bs, err := Encoding(item, level+1)
			if err != nil {
				return nil, err
			}
			if string(bs)[0] == '{' {
				arr = append(arr, fmt.Sprintf("%s %s", name, bs))
			} else {
				arr = append(arr, fmt.Sprintf("%s = %s", name, bs))
			}
			*/
			err := loopHash(&arr, name, item, level)
			if err != nil {
				return nil, err
			}
		}
		if level == 0 {
			str = fmt.Sprintf("\n%s\n%s", leading+strings.Join(arr, "\n"+leading), lessLeading)
		} else {
			str = fmt.Sprintf("{\n%s\n%s}", leading+strings.Join(arr, "\n"+leading), lessLeading)
		}
	case reflect.Slice, reflect.Array:
		var arr []string
		for _, item := range current.([]interface{}) {
			bs, err := Encoding(item, level+1)
			if err != nil {
				return nil, err
			}
			arr = append(arr, string(bs))
		}
		str = fmt.Sprintf("[\n%s\n%s]", leading+strings.Join(arr, ",\n"+leading), lessLeading)
		// both the above and the following code are correct
		// str = fmt.Sprintf("[%s]", strings.Join(arr, ", "))
	case reflect.String:
		str = "\"" + rv.String() + "\""
	case reflect.Bool:
		str = fmt.Sprintf("%t", rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		str = fmt.Sprintf("%d", rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		str = fmt.Sprintf("%d", rv.Uint())
	case reflect.Float32:
		str = strconv.FormatFloat(rv.Float(), 'f', -1, 32)
	case reflect.Float64:
		str = strconv.FormatFloat(rv.Float(), 'f', -1, 64)
	default:
		return nil, fmt.Errorf("data type %v not supported", rv.Kind())
	}

	return []byte(str), nil
}
