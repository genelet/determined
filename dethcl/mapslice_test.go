package dethcl

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestJsonHcl(t *testing.T) {
	data1 := `{
"name": "peter", "radius": 1.0, "num": 2, "parties":["one", "two", ["three", "four"], {"five":"51", "six":61}], "roads":{"x":"a","y":"b", "z":{"za":"aa","zb":3.14}, "xy":["ab", true]}
}`
	d := map[string]interface{}{}
	err := json.Unmarshal([]byte(data1), &d)
	if err != nil {
		t.Fatal(err)
	}

	bs, err := Marshal(d)
	if err != nil {
		t.Fatal(err)
	}

	m := make(map[string]interface{})
	err = Unmarshal(bs, &m)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(d, m) {
		t.Errorf("%#v", d)
		t.Errorf("%s", bs)
		t.Errorf("%#v", m)
	}
}
