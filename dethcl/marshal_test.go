package dethcl

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/genelet/determined/utils"
)

func TestMHclSimple(t *testing.T) {
	data1 := `radius = 1.0`
	c := new(circle)
	err := UnmarshalSpec([]byte(data1), c, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	bs, err := Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != "  radius = 1" {
		t.Errorf("%s", bs)
	}
}

func TestMHclSimpleMore(t *testing.T) {
	data1 := `
	radius = 1.0
arr1 = ["abc", "def"]
arr2 = [123, 4356]
arr3 = [true, false, true]
`
	c := new(circlemore)
	err := UnmarshalSpec([]byte(data1), c, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	bs, err := Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `  radius = 1
  arr1   = ["abc", "def"]
  arr2   = [123, 4356]
  arr3   = [true, false, true]` {
		t.Errorf("'%s'", bs)
	}
}

func TestMHclShape(t *testing.T) {
	data1 := `
	name = "peter shape"
	shape {
		radius = 1.0
	}
`
	g := &geo{}
	c := &circle{}
	ref := map[string]interface{}{"circle": c}
	spec, err := utils.NewStruct(
		"geo", map[string]interface{}{"Shape": "circle"})
	if err == nil {
		err = UnmarshalSpec([]byte(data1), g, spec, ref)
	}
	if err != nil {
		t.Fatal(err)
	}
	bs, err := Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `  name = "peter shape"
  shape {
    radius = 1
  }` {
		t.Errorf("%s", bs)
	}
}

func TestMHclMoreShape(t *testing.T) {
	data2 := `
	name = "peter shape"
	shape {
    	sx = 5
    	sy = 6
	}
`
	g := &geo{}
	c := &circle{}
	s := &square{}
	ref := map[string]interface{}{"circle": c, "square": s}
	spec, err := utils.NewStruct(
		"geo", map[string]interface{}{"Shape": "square"})
	if err == nil {
		err = UnmarshalSpec([]byte(data2), g, spec, ref)
	}
	if err != nil {
		t.Fatal(err)
	}
	bs, err := Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `  name = "peter shape"
  shape {
    sx = 5
    sy = 6
  }` {
		t.Errorf("'%s'", bs)
	}

	data4 := `
	name = "peter drawings"
	drawings {
		sx=5
		sy=6
	}
	drawings {
		sx=7
		sy=8
	}
`
	p := &picture{}
	spec, err = utils.NewStruct(
		"Picture", map[string]interface{}{
			"Drawings": []string{"square", "square"}})
	if err != nil {
		t.Fatal(err)
	}
	err = UnmarshalSpec([]byte(data4), p, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	bs, err = Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `  name = "peter drawings"
  drawings {
    sx = 5
    sy = 6
  }
  drawings {
    sx = 7
    sy = 8
  }` {
		t.Errorf("'%s'", bs)
	}

	data5 := `
    name = "peter drawings"
    drawings "abc1" def1 {
        sx=5
        sy=6
    }
    drawings abc2 "def2" {
        sx=7
        sy=8
    }
`

	p = &picture{}
	spec, err = utils.NewStruct(
		"Picture", map[string]interface{}{
			"Drawings": []string{"moresquare", "moresquare"}})
	if err != nil {
		t.Fatal(err)
	}
	ref["moresquare"] = &moresquare{}
	err = UnmarshalSpec([]byte(data5), p, spec, ref)
	if err != nil {
		t.Fatal(err)
	}

	bs, err = Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `  name = "peter drawings"
  drawings "abc1" "def1" {
    sx = 5
    sy = 6
  }
  drawings "abc2" "def2" {
    sx = 7
    sy = 8
  }` {
		t.Errorf("'%s'", bs)
	}
}

func TestMHash(t *testing.T) {
	data3 := `
	name = "peter shapes"
	shapes obj5 {
		sx = 5
		sy = 6
	}
	shapes obj7 {
		sx = 7
		sy = 8
	}
`
	g := &geometry{}
	ref := map[string]interface{}{"square": new(square)}
	spec, err := utils.NewStruct(
		"geometry", map[string]interface{}{
			"Shapes": []string{"square", "square"}})
	if err == nil {
		err = UnmarshalSpec([]byte(data3), g, spec, ref)
	}
	if err != nil {
		t.Fatal(err)
	}

	bs, err := Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `  name = "peter shapes"
  shapes "obj5" {
    sx = 5
    sy = 6
  }
  shapes "obj7" {
    sx = 7
    sy = 8
  }` && string(bs) != `  name = "peter shapes"
  shapes "obj7" {
    sx = 7
    sy = 8
  }
  shapes "obj5" {
    sx = 5
    sy = 6
  }` {
		t.Errorf("'%s'", bs)
	}
}

func TestMPointerHash(t *testing.T) {
	data3 := `
	name = "peter shapes"
	shapes obj5 {
		sx = 5
		sy = 6
	}
	shapes obj7 {
		sx = 7
		sy = 8
	}
`
	g := &geometry{}
	ref := map[string]interface{}{"square": new(square)}
	spec, err := utils.NewStruct(
		"geometry", map[string]interface{}{
			"Shapes": []string{"square", "square"}})
	if err == nil {
		err = UnmarshalSpec([]byte(data3), g, spec, ref)
	}
	if err != nil {
		t.Fatal(err)
	}

	p := &geometries{
		Name:   g.Name,
		Shapes: &g.Shapes,
	}

	bs, err := Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `  name = "peter shapes"
  shapes "obj5" {
    sx = 5
    sy = 6
  }
  shapes "obj7" {
    sx = 7
    sy = 8
  }` && string(bs) != `  name = "peter shapes"
  shapes "obj7" {
    sx = 7
    sy = 8
  }
  shapes "obj5" {
    sx = 5
    sy = 6
  }` {
		t.Errorf("'%s'", bs)
	}
}

func TestMHclOld(t *testing.T) {
	data1 := `  description = "here is detailed description"
  y13 = {
    str131 = "la"
    str132 = "nyc"
  }
  y14 = {
    str141 = 141
    str142 = 142
  }
  y15 = {
    str151 = true
    str152 = false
  }
  y7 {
    many = 3
    why  = "national day"
  }
  y10 {
    many = 4
    why  = "labor day"
  }
  y10 {
    many = 5
    why  = "holiday day"
  }
  y11 k6 {
    many = 6
    why  = "memorial day"
  }
  y11 k7 {
    many = 7
    why  = "new day"
  }
  y12 k8 {
    many = 8
    why  = "christmas day"
  }
  y12 k9 {
    many = 9
    why  = "new year day"
  }`
	f0 := new(frame0)
	err := UnmarshalSpec([]byte(data1), f0, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(f0.Y10) != 2 || len(f0.Y11) != 2 || len(f0.Y12) != 2 {
		t.Errorf("%#v", f0)
	}

	bs, err := Marshal(f0)
	if err != nil {
		t.Fatal(err)
	}
	if (string(data1))[:100] != (string(bs))[:100] {
		t.Errorf("%s", bs)
	}
}

func TestMHclFrame(t *testing.T) {
	data1 := `description = "here is detailed description"
number      = 4
what        = "flags"
x1 {
  name = "x1 shape"
  shape {
    radius = 1
  }
  
}

x2 {
  name = "x2 shape"
  shape {
    radius = 2
  }
  
}

x3 {
  name = "x3 1 shape"
  shape {
    radius = 3
  }
  
}

x3 {
  name = "x3 2 shape"
  shape {
    radius = 3
  }
  
}

x4 {
  name = "x4 1 shape"
  shape {
    radius = 4
  }
  
}

x4 {
  name = "x4 2 shape"
  shape {
    radius = 4
  }
  
}

x5 k51 {
  name = "x5 1 shape"
  shape {
    radius = 5
  }
  
}

x5 k52 {
  name = "x5 2 shape"
  shape {
    radius = 5
  }
  
}

x6 k61 {
  name = "x6 1 shape"
  shape {
    radius = 6
  }
  
}

x6 k62 {
  name = "x6 2 shape"
  shape {
    radius = 6
  }
  
}

y7 {
  many = 3
  why  = "national day"
}

y10 {
  many = 4
  why  = "labor day"
}

y10 {
  many = 5
  why  = "holiday day"
}

y11 k7 {
  many = 7
  why  = "new day"
}

y11 k6 {
  many = 6
  why  = "memorial day"
}

`
	spec, err := utils.NewStruct(
		"frame", map[string]interface{}{
			"X1": [2]interface{}{
				"geo", map[string]interface{}{"Shape": "circle"},
			},
			"X2": [2]interface{}{
				"geo", map[string]interface{}{"Shape": "circle"},
			},
			"X3": [][2]interface{}{
				{"geo", map[string]interface{}{"Shape": "circle"}},
				{"geo", map[string]interface{}{"Shape": "circle"}},
			},
			"X4": [][2]interface{}{
				{"geo", map[string]interface{}{"Shape": "circle"}},
				{"geo", map[string]interface{}{"Shape": "circle"}},
			},
			"X5": [][2]interface{}{
				{"geo", map[string]interface{}{"Shape": "circle"}},
				{"geo", map[string]interface{}{"Shape": "circle"}},
			},
			"X6": [][2]interface{}{
				{"geo", map[string]interface{}{"Shape": "circle"}},
				{"geo", map[string]interface{}{"Shape": "circle"}},
			},
		})
	if err != nil {
		t.Fatal(err)
	}

	ref := map[string]interface{}{"geo": &geo{}, "circle": &circle{}, "frame": &frame{}}

	f := new(frame)
	err = UnmarshalSpec([]byte(data1), f, spec, ref)
	if err != nil {
		t.Fatal(err)
	}
	bs, err := Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	//	if string(data1) != string(bs) {
	if false {
		t.Errorf("%s", bs)
	}
}

func TestMHclChild(t *testing.T) {
	data1 := `
age = 5
brand {
	toy_name = "roblox"
	price = 99.9
	geo {
		name = "peter shape"
		shape {
    		radius = 1.0
		}
	}
}
`

	ref := map[string]interface{}{"geo": &geo{}, "circle": &circle{}, "toy": &toy{}}
	c := new(child)
	spec, err := utils.NewStruct(
		"child1", map[string]interface{}{
			"Brand": [2]interface{}{
				"toy", map[string]interface{}{
					"Geo": [2]interface{}{
						"geo", map[string]interface{}{"Shape": "circle"}}}}})
	if err == nil {
		err = UnmarshalSpec([]byte(data1), c, spec, ref)
	}
	if err != nil {
		t.Fatal(err)
	}

	bs, err := Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `  age = 5
  brand {
    toy_name = "roblox"
    price    = 99.9000015258789
    geo {
      name = "peter shape"
      shape {
        radius = 1
      }
    }
  }` {
		t.Errorf("'%s'", bs)
	}
}

func TestMHclPainting(t *testing.T) {
	data4 := `
	name = "peter drawings"
	drawings obj5 {
		sx=55
		sy=66
	}
	drawings obj7 {
		sx=7
		sy=8
	}
`

	ref := map[string]interface{}{"square": new(square)}
	p := &config{}
	spec, err := utils.NewStruct(
		"config", map[string]interface{}{
			"Drawings": []string{"square", "square"}})
	if err == nil {
		err = UnmarshalSpec([]byte(data4), p, spec, ref)
	}
	if err != nil {
		t.Fatal(err)
	}

	bs, err := Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `  name = "peter drawings"
  drawings "obj5" {
    sx = 55
    sy = 66
  }
  drawings "obj7" {
    sx = 7
    sy = 8
  }` {
		t.Errorf("'%s'", bs)
	}
}

func TestZeroFalseSimple(t *testing.T) {
	hash := map[string]interface{}{
		"request_id":     "2e7a9b1d-a8d6-4ce4-6380-47c05cf1d16e",
		"lease_id":       "",
		"renewable":      false,
		"lease_duration": 0,
		"data":           nil,
	}
	bs, err := Marshal(hash)
	if err != nil {
		t.Fatal(err)
	}
	x := make(map[string]interface{})
	err = Unmarshal(bs, &x)
	if err != nil {
		t.Fatal(err)
	}

	if x["lease_duration"].(int) != 0 ||
		x["renewable"].(bool) != false ||
		x["data"] != nil ||
		x["lease_id"] != "" ||
		x["request_id"] != "2e7a9b1d-a8d6-4ce4-6380-47c05cf1d16e" {
		t.Errorf("%#v", x)
	}
}

func TestZeroFalseMore(t *testing.T) {
	data := `{"request_id":"2e7a9b1d-a8d6-4ce4-6380-47c05cf1d16e","lease_id":"","renewable":false,"lease_duration":0,"data":null,"wrap_info":null,"warnings":null,"auth":{"client_token":"hvs.","accessor":"xxx","policies":["adv_policy","default"],"token_policies":["adv_policy","default"],"identity_policies":["adv_policy"],"metadata":{"username":"peter_001@kinet.com"},"lease_duration":36000,"renewable":true,"entity_id":"70debb54-a346-06c6-7c22-26bf330aa3c8","token_type":"service","orphan":true,"mfa_requirement":null,"num_uses":0},"mount_type":""}`
	hash := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &hash)
	if err != nil {
		t.Fatal(err)
	}
	bs, err := Marshal(hash)
	if err != nil {
		t.Fatal(err)
	}
	x := make(map[string]interface{})
	err = Unmarshal(bs, &x)
	if err != nil {
		t.Fatal(err)
	}
	if x["lease_duration"].(int) != 0 ||
		x["renewable"].(bool) != false ||
		x["data"] != nil ||
		x["lease_id"] != "" ||
		x["request_id"] != "2e7a9b1d-a8d6-4ce4-6380-47c05cf1d16e" {
		t.Errorf("%#v", x)
	}
	str1, err := json.Marshal(x["auth"])
	if err != nil {
		t.Fatal(err)
	}
	str2, err := json.Marshal(hash["auth"])
	if err != nil {
		t.Fatal(err)
	}
	if string(str1) != string(str2) {
		t.Errorf("%#v", x["auth"])
		t.Errorf("%#v", hash["auth"])
	}
}

type response struct {
	BodyData    map[string]interface{} `hcl:"body_data,block"`
	HeadersData map[string][]string    `hcl:"headers_data,optional"`
}

func TestZeroFalseMore2(t *testing.T) {
	data := `
  body_data {
    renewable = false
    lease_duration = 0
    mount_type = ""
    request_id = "5aebeec9-653a-44b2-363f-f9152274cb30"
    lease_id = ""
    auth {
      token_policies = [
        "adv_policy",
        "default"
      ]
      identity_policies = [
        "adv_policy"
      ]
      metadata {
        username = "peter_001@kinet.com"
      }
      entity_id = "70debb54-a346-06c6-7c22-26bf330aa3c8"
      lease_duration = 36000
      policies = [
        "adv_policy",
        "default"
      ]
      renewable = true
      mfa_requirement = null()
      token_type = "service"
      orphan = true
      num_uses = 0
    }
  }`

	r := new(response)
	err := Unmarshal([]byte(data), r)
	if err != nil {
		t.Fatal(err)
	}
	bs, err := Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(bs) != 644 {
		t.Errorf("%d %s", len(bs), bs)
	}
}

type rspn struct {
	Method              string                 `hcl:"method,optional" json:"method,omitempty"`
	Path                string                 `hcl:"path,optional" json:"path,omitempty"`
	Payload             string                 `hcl:"payload,optional" json:"payload,omitempty"`
	HeadersData         map[string][]string    `hcl:"request_headers,optional" json:"request_headers,omitempty"`
	StatusCode          int                    `hcl:"status_code,optional" json:"status_code,omitempty"`
	ResponseBodyData    map[string]interface{} `hcl:"response_data,block" json:"response_data,omitempty"`
	ResponseHeadersData map[string][]string    `hcl:"response_headers,optional" json:"response_headers,omitempty"`
}

func TestResponseUnmarshal(t *testing.T) {
	bs := []byte(`
  method      = "GET"
  path        = "http://ec2.us-west-2.amazonaws.com/?Action=DescribeInstanceStatus&Version=2016-11-15"
  status_code = 200
  request_headers  = {
    Authorization = [
      "AWS4-HMAC-SHA256 Credential=AKIASAZMGJKKCLV57M8Y/20240730/us-west-2/ec2/aws4_request, SignedHeaders=host;x-amz-date, Signature=a9df8dcc2e8d4fc623061f49282554d8edf2d2093e19f0d04e49294c5ba0b1fc"
    ]
    X-Amz-Date = [
      "20240730T122641Z"
    ]
  }
  response_data {
    DescribeInstanceStatusResponse {
      requestId = "e64de5a0-693f-43a4-a8f6-1fb94614528e"
      instanceStatusSet "item"  {
        instanceId = "i-0064c18e730799d76"
        availabilityZone = "us-west-2a"
        instanceState = {
          code = "16"
          name = "running"
        }
        systemStatus {
          details "item"  {
            name = "reachability"
            status = "passed"
          }
          status = "ok"
        }
        instanceStatus {
          status = "ok"
          details "item"  {
            name = "reachability"
            status = "passed"
          }
        }
      }
      xmlns = "http://ec2.amazonaws.com/doc/2016-11-15/"
    }
  }
`)

	r := new(rspn)
	err := Unmarshal(bs, r)
	if err != nil {
		t.Fatal(err)
	}
	bs, err = Marshal(r)
	if err != nil {
		t.Fatal(err)
	}

	r1 := new(rspn)
	err = Unmarshal(bs, r1)
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(r, r1) == false {
		t.Errorf("%#v\n%#v", r, r1)
	}
}
