package dethcl

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/genelet/determined/utils"
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
		bs, err := marshalLevel(item, equal, level+1, name)
		if err != nil {
			return err
		}
		var str string
		if equal || depth < 1 {
			str = "="
		}
		// this is the only place where 'equal' matters. if we don't care about it,
		// we can just remove variable 'equal' from functions encoding and loopHash.
		*arr = append(*arr, fmt.Sprintf("%s %s %s", name, str, bs))
	default:
		switch item.(type) {
		case string:
			*arr = append(*arr, fmt.Sprintf("%s = \"%s\"", name, item))
			return nil
		case bool:
			*arr = append(*arr, fmt.Sprintf("%s = %t", name, item))
			return nil
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			*arr = append(*arr, fmt.Sprintf("%s = %d", name, item))
			return nil
		case float32, float64:
			c, err := utils.NativeToCty(item)
			if err != nil {
				return err
			}
			n, err := utils.CtyNumberToNative(c)
			if err != nil {
				return err
			}
			switch n.(type) {
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				*arr = append(*arr, fmt.Sprintf("%s = %d", name, n))
			default:
				*arr = append(*arr, fmt.Sprintf("%s = %f", name, n))
			}
			return nil
		default:
		}
		bs, err := marshalLevel(item, equal, level+1)
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

func encoding(current interface{}, equal bool, level int, keyname ...string) ([]byte, error) {
	var str string
	if current == nil {
		return nil, nil
	}
	leading := strings.Repeat("  ", level+1)
	lessLeading := strings.Repeat("  ", level)

	rv := reflect.ValueOf(current)
	switch rv.Kind() {
	case reflect.Pointer:
		return marshalLevel(rv.Elem().Interface(), equal, level, keyname...)
	case reflect.Map:
		var arr []string
		iter := rv.MapRange()
		for iter.Next() {
			key := iter.Key()
			if key.Kind() != reflect.String {
				return nil, fmt.Errorf("map key must be string, got %v", key.Kind())
			}
			switch iter.Value().Kind() {
			case reflect.Interface, reflect.Ptr, reflect.Map, reflect.Slice, reflect.Func:
				if iter.Value().IsNil() {
					arr = append(arr, fmt.Sprintf("%s = null()", key.String()))
					continue
				}
			default:
			}
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
			bs, err := marshalLevel(rv.Index(i).Interface(), true, level+1)
			if err != nil {
				return nil, err
			}
			item := `[]`
			if bs != nil {
				item = string(bs)
			}
			arr = append(arr, item)
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
