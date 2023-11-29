package dethcl

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestJsonHcl(t *testing.T) {
	data := `{
"name": "peter", "radius": 1.0, "num": 2, "parties":["one", "two", ["three", "four"], {"five":"51", "six":61}], "roads":{"x":"a","y":"b", "z":{"za":"aa","zb":3.14}, "xy":["ab", true]}
}`
	d := map[string]interface{}{}
	err := json.Unmarshal([]byte(data), &d)
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

	data1, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if string(data1) != `{"io_mode":"async","service":{"http":{"web_proxy":{"listen_addr":"127.0.0.1:8080","process":{"main":{"command":["/usr/local/bin/awesome-app","server","gosh"]},"mgmt":{"command":["/usr/local/bin/awesome-app","mgmt"]}}}}}}` {
		t.Errorf("%s", data)
		t.Errorf("%s", data1)
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(data1, &m)
	if err != nil {
		t.Fatal(err)
	}

	bs, err := Marshal(m)
	if string(bs) != `{
  service http web_proxy {
    listen_addr = "127.0.0.1:8080"
    process main {
	  command = [
	    "/usr/local/bin/awesome-app",
	    "server",
	    "gosh"
	  ],
	  received = 1.000000
    }
    process mgmt {
	  command = [
	    "/usr/local/bin/awesome-app",
	    "mgmt"
	  ]
    }
  },
  io_mode = "async"
}` {
		t.Errorf("%s", data)
		t.Errorf("%s", data1)
		t.Errorf("%s", bs)
	}
}
