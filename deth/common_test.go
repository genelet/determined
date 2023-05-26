package deth

import (
	"reflect"
	"strings"
	"testing"
)

func TestCommonString(t *testing.T) {
	endpoint, err := NewStruct(
		"Geo", map[string]interface{}{
			"TheString0": "Circle",
			"TheList0":   []string{"CircleClass1", "CircleClass2"},
		},
	)
	if err != nil {
		panic(err)
	}
	fields := endpoint.GetFields()
	if strings.ReplaceAll(fields["TheString0"].String()," ", "") != "single_struct:{className:\"Circle\"}" {
		t.Errorf("%#v", fields["TheString0"].String())
	}
	if strings.ReplaceAll(fields["TheList0"].String()," ", "") != "list_struct:{list_fields:{className:\"CircleClass1\"}list_fields:{className:\"CircleClass2\"}}" {
		t.Errorf("%#v", fields["TheList0"].String())
	}
}

func TestCommonStruct(t *testing.T) {
	endpoint, err := NewStruct(
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
	shapeFields := endpoint.GetFields()

	shapeEndpoint := shapeFields["Shape1"].GetSingleStruct()
	field1Fields := shapeEndpoint.GetFields()
	field1Endpoint := field1Fields["Field1"].GetSingleStruct()
	if endpoint.ClassName != "Geo" ||
		shapeEndpoint.ClassName != "Class1" ||
		field1Endpoint.ClassName != "Circle1" {
		t.Errorf("shape endpoint: %s", shapeEndpoint.String())
		t.Errorf("field 1 endpoint: %s", field1Endpoint.String())
	}

	shape2Endpoint := shapeFields["Shape2"].GetSingleStruct()
	field2Fields := shape2Endpoint.GetFields()
	field2Endpoint := field2Fields["Field2"].GetListStruct()
	if endpoint.ClassName != "Geo" ||
		shape2Endpoint.ClassName != "Class2" ||
		field2Endpoint.ListFields[0].ClassName != "Circle2" ||
		field2Endpoint.ListFields[1].ClassName != "Circle3" {
		t.Errorf("shape endpoint: %s", shape2Endpoint.String())
		t.Errorf("field 2 endpoint: %s", field2Endpoint.String())
	}
}

func TestCommonList(t *testing.T) {
	endpoint, err := NewStruct(
		"Geo", map[string]interface{}{
			"ListShapes": [][2]interface{}{
				{"Class2", map[string]interface{}{"Field3": "Circle"}},
				{"Class3", map[string]interface{}{"Field5": "Circle"}}},
		},
	)
	if err != nil {
		panic(err)
	}
	shapeFields := endpoint.GetFields()
	shapeEndpoint := shapeFields["ListShapes"].GetListStruct().GetListFields()[1]
	field1Fields := shapeEndpoint.GetFields()
	field1Endpoint := field1Fields["Field5"].GetSingleStruct()
	if endpoint.ClassName != "Geo" ||
		shapeEndpoint.ClassName != "Class3" ||
		field1Endpoint.ClassName != "Circle" {
		t.Errorf("shape endpoint: %s", shapeEndpoint.String())
		t.Errorf("field 1 endpoint: %s", field1Endpoint.String())
	}
}

type xclass struct {
    Name   string              `json:"name" hcl:"name"`
    Squares map[string]*square `json:"squares" hcl:"squares,block"`
    Circles map[string]*circle `json:"circles" hcl:"circles,block"`
}

func TestMapList(t *testing.T) {
	x := &xclass{Name: "xclass name",
			Squares: map[string]*square{
				"k1": &square{SX: 1, SY: 2}, "k2": &square{SX: 3, SY: 4}},
			Circles: map[string]*circle{
				"k5": &circle{5.6}, "k6": &circle{6.7}}}
	typ := reflect.TypeOf(x).Elem()
    n := typ.NumField()
	oriValue := reflect.ValueOf(x).Elem()

	fields := make(map[string]*Value)
	ref := make(map[string]interface{})

	for i := 0; i < n; i++ {
		field := typ.Field(i)
		rawField := oriValue.Field(i)	
		if field.Type.Kind() != reflect.Map {
			continue
		}
		var arr []string
		iter := rawField.MapRange()
    	for iter.Next() {
        	k := iter.Key()
        	v := iter.Value()
			t.Errorf("%#v", k.Interface().(string))
			s := v.Type().String()
			arr = append(arr, s)
			ref[s] = reflect.New(v.Type()).Elem().Interface()
		}
		val, err := NewValue(arr)
		if err != nil {
			t.Fatal(err)
		}
		fields[field.Name] = val
	}
	tr := &Struct{Fields:fields}
	t.Errorf("%#v", tr.String())
	t.Errorf("%#v", ref)

	data6 := `
    name = "peter drawings"
    squares "abc1" def1 {
        sx=5
        sy=6
    }
    squares abc2 "def2" {
        sx=7
        sy=8
    }
    circles "xyz4" def1 {
        radius=5.6
    }
    circles xyz5 "def2" {
        radius=6.7
    }
`
	xc := &xclass{}
    err := HclUnmarshal([]byte(data6), xc, tr, ref)
    if err != nil {
        t.Fatal(err)
    }
	t.Errorf("%#v", xc)
}
