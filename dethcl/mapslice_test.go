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

/*
func TestDecodeMap(t *testing.T) {
	data1 := `
io_mode = "async"

service "http" "web_proxy" {
  listen_addr = "127.0.0.1:8080"

  process "main" {
    command = ["/usr/local/bin/awesome-app", "server", "gosh"]
  }

  process "mgmt" {
    command = ["/usr/local/bin/awesome-app", "mgmt"]
  }
}`
	d := map[string]interface{}{}
	err := Unmarshal([]byte(data1), &d)
	if err != nil {
		t.Fatal(err)
	}

	bs, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		t.Fatal(err)
	}

	bs, err = Marshal(m)

	if !reflect.DeepEqual(d, m) {
		t.Errorf("%#v", d)
		t.Errorf("%s", bs)
		t.Errorf("%#v", m)
	}
}

io_mode = "async"

service "http" "web_proxy" {
  listen_addr = "127.0.0.1:8080"

  process "main" {
	command = ["/usr/local/bin/awesome-app", "server", "gosh"]
  }

  process "mgmt" {
	command = ["/usr/local/bin/awesome-app", "mgmt"]
  }
}

io_mode = "async",
service = [
  {
	process = [
	  {
		command = [
		  "/usr/local/bin/awesome-app",
		  "server",
		  "gosh"
		],
		process_label_0 = "main"
	  },
	  {
		command = [
		  "/usr/local/bin/awesome-app",
		  "mgmt"
		],
		process_label_0 = "mgmt"
	  }
	],
	service_label_0 = "http",
	service_label_1 = "web_proxy",
	listen_addr = "127.0.0.1:8080"
  }
]
*/
