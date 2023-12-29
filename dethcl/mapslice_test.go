package dethcl

import (
	"encoding/json"
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

	if !jsonEqual(d, m) {
		t.Errorf("%#v", d)
		t.Errorf("%#v", m)
		t.Errorf("%s", bs)
	}
}

func jsonEqual(d, m map[string]interface{}) bool {
	xd, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}
	xm, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(xd) == string(xm)
}

func TestDecodeMap(t *testing.T) {
	data := `
io_mode = "async"

service "http" "web_proxy" {
  listen_addr = "127.0.0.1:8080"

  process "main" {
    command = ["/usr/local/bin/awesome-app", "server", "gosh"]
    received = 1
  }

  process "mgmt" {
    command = ["/usr/local/bin/awesome-app", "mgmt"]
  }
}`
	d := map[string]interface{}{}
	err := Unmarshal([]byte(data), &d)
	if err != nil {
		t.Fatal(err)
	}
	//t.Errorf("%#v", d)
	bs, err := Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	//t.Errorf("%s", bs)
	m := make(map[string]interface{})
	err = Unmarshal(bs, &m)
	if err != nil {
		t.Fatal(err)
	}

	if !jsonEqual(d, m) {
		t.Errorf("%#v", d)
		t.Errorf("%#v", m)
		t.Errorf("%s", data)
		t.Errorf("%s", bs)
	}

}
