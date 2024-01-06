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

func loopHash(arr *[]string, name string, item interface{}, equal bool, depth, level int, keyname ...string) error {
	found, next := isHashAll(item)
	if depth >= 2 && found == 2 {
		found = 1
	}
	switch found {
	case 2:
		for key, value := range next {
			name2 := name + ` "` + key + `"`
			err := loopHash(arr, name2, value, false, depth+1, level)
			if err != nil {
				return err
			}
		}
	case 1:
		// pass 'name' as the keyname to the next 'default' below
		bs, err := encoding(item, equal, level+1, name)
		if err != nil {
			return err
		}
		var str string
		if equal {
			str = "="
		}
		// this is the only place where 'equal' matters. if we don't care about it,
		// we can just remove variable 'equal' from functions encoding and loopHash.
		*arr = append(*arr, fmt.Sprintf("%s %s %s", name, str, bs))
	default:
		bs, err := encoding(item, equal, level+1)
		if err != nil {
			return err
		}
		if keyname != nil && matchlast(keyname[0], string(bs)) {
			return nil
		}
		*arr = append(*arr, fmt.Sprintf("%s = %s", name, bs))
	}
	return nil
}

func matchlast(keyname string, name string) bool {
	names := strings.Split(keyname, " ")
	keyname = names[len(names)-1]
	if keyname == name {
		return true
	}
	return false
}

// Encoding encode the data to HCL format
func Encoding(current interface{}, level int) ([]byte, error) {
	return encoding(current, false, level)
}

func encoding(current interface{}, equal bool, level int, keyname ...string) ([]byte, error) {
	var str string

	leading := strings.Repeat("  ", level+1)
	lessLeading := strings.Repeat("  ", level)

	rv := reflect.ValueOf(current)
	switch rv.Kind() {
	case reflect.Pointer:
		return encoding(rv.Elem().Interface(), equal, level, keyname...)
	case reflect.Map:
		var arr []string
		iter := rv.MapRange()
		for iter.Next() {
			key := iter.Key()
			err := loopHash(&arr, key.String(), iter.Value().Interface(), equal, 0, level, keyname...)
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
		for i := 0; i < rv.Len(); i++ {
			bs, err := encoding(rv.Index(i).Interface(), true, level+1)
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
