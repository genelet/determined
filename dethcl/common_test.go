package dethcl

import (
	"reflect"
	"strings"
	"testing"
)

func TestCommonString(t *testing.T) {
	spec, err := NewStruct(
		"Geo", map[string]interface{}{
			"TheString0": "Circle",
			"TheList0":   []string{"CircleClass1", "CircleClass2"},
		},
	)
	if err != nil {
		panic(err)
	}
	fields := spec.GetFields()
	if strings.ReplaceAll(fields["TheString0"].String()," ", "") != "single_struct:{className:\"Circle\"}" {
		t.Errorf("%#v", fields["TheString0"].String())
	}
	if strings.ReplaceAll(fields["TheList0"].String()," ", "") != "list_struct:{list_fields:{className:\"CircleClass1\"}list_fields:{className:\"CircleClass2\"}}" {
		t.Errorf("%#v", fields["TheList0"].String())
	}
}

func TestCommonStruct(t *testing.T) {
	spec, err := NewStruct(
		"Geo", map[string]interface{}{
			"Shape1": [2]interface{}{
				"Class1", map[string]interface{}{"Field1": "Circle1"}},
			"Shape2": [2]interface{}{
				"Class2", map[string]interface{}{"Field2": []string{"Circle2","Circle3"}}},
		},
	)
	if err != nil {
		panic(err)
	}
	shapeFields := spec.GetFields()

	shapeEndpoint := shapeFields["Shape1"].GetSingleStruct()
	field1Fields := shapeEndpoint.GetFields()
	field1Endpoint := field1Fields["Field1"].GetSingleStruct()
	if spec.ClassName != "Geo" ||
		shapeEndpoint.ClassName != "Class1" ||
		field1Endpoint.ClassName != "Circle1" {
		t.Errorf("shape spec: %s", shapeEndpoint.String())
		t.Errorf("field 1 spec: %s", field1Endpoint.String())
	}

	shape2Endpoint := shapeFields["Shape2"].GetSingleStruct()
	field2Fields := shape2Endpoint.GetFields()
	field2Endpoint := field2Fields["Field2"].GetListStruct()
	if spec.ClassName != "Geo" ||
		shape2Endpoint.ClassName != "Class2" ||
		field2Endpoint.ListFields[0].ClassName != "Circle2" ||
		field2Endpoint.ListFields[1].ClassName != "Circle3" {
		t.Errorf("shape spec: %s", shape2Endpoint.String())
		t.Errorf("field 2 spec: %s", field2Endpoint.String())
	}
}

func TestCommonList(t *testing.T) {
	spec, err := NewStruct(
		"Geo", map[string]interface{}{
			"ListShapes": [][2]interface{}{
				{"Class2", map[string]interface{}{"Field3": "Circle"}},
				{"Class3", map[string]interface{}{"Field5": "Circle"}}},
		},
	)
	if err != nil {
		panic(err)
	}
	shapeFields := spec.GetFields()
	shapeEndpoint := shapeFields["ListShapes"].GetListStruct().GetListFields()[1]
	field1Fields := shapeEndpoint.GetFields()
	field1Endpoint := field1Fields["Field5"].GetSingleStruct()
	if spec.ClassName != "Geo" ||
		shapeEndpoint.ClassName != "Class3" ||
		field1Endpoint.ClassName != "Circle" {
		t.Errorf("shape spec: %s", shapeEndpoint.String())
		t.Errorf("field 1 spec: %s", field1Endpoint.String())
	}
}

type xclass struct {
    Name   string              `json:"name" hcl:"name"`
    Squares map[string]*square `json:"squares" hcl:"squares"`
    Circles map[string]*circle `json:"circles" hcl:"circles"`
}

func TestMapList(t *testing.T) {
	x := &xclass{Name: "xclass name",
			Squares: map[string]*square{
				"k1": &square{SX: 1, SY: 2}, "k2": &square{SX: 3, SY: 4}},
			Circles: map[string]*circle{
				"k5": &circle{5.6}, "k6": &circle{6.7}}}
	bs, err := Marshal(x)
	if err != nil { t.Fatal(err) }

	typ := reflect.TypeOf(x).Elem()
    n := typ.NumField()
	oriValue := reflect.ValueOf(x).Elem()

	ref  := make(map[string]interface{})
	spec := make(map[string]interface{})

	for i := 0; i < n; i++ {
		field := typ.Field(i)
		rawField := oriValue.Field(i)	
		if field.Type.Kind() != reflect.Map {
			continue
		}
		var arr []string
		iter := rawField.MapRange()
    	for iter.Next() {
        	//k := iter.Key()
        	v := iter.Value()
			s := v.Type().String()
			arr = append(arr, s)
			ref[s] = reflect.New(v.Type().Elem()).Elem().Addr().Interface()
		}
		spec[field.Name] = arr
	}

	tr, err := NewStruct(typ.Name(), spec)
	xc := &xclass{}
    err = Unmarshal(bs, xc, tr, ref)
    if err != nil { t.Fatal(err) }

	if x.Squares["k1"].SX != xc.Squares["k1"].SX ||
		x.Squares["k2"].SX != xc.Squares["k2"].SX ||
		x.Circles["k5"].Radius != xc.Circles["k5"].Radius ||
		x.Circles["k6"].Radius != xc.Circles["k6"].Radius {
		t.Errorf("%#v", x.Squares["k1"])
		t.Errorf("%#v", x.Squares["k2"])
		t.Errorf("%#v", x.Circles["k5"])
		t.Errorf("%#v", x.Circles["k6"])
		t.Errorf("%#v", xc.Squares["k1"])
		t.Errorf("%#v", xc.Squares["k2"])
		t.Errorf("%#v", xc.Circles["k5"])
		t.Errorf("%#v", xc.Circles["k6"])
	}
}
