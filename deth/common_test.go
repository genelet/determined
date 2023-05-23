package deth

import (
	"testing"
	"strings"
)

func TestCommonString(t *testing.T) {
	endpoint, err := NewStruct(
		"Geo", map[string]interface{}{
			"TheString0": "Circle",
			"TheString1": [1]string{"Circle"},
			"TheString2": [2]string{"Circle"},

			"TheList0":   []string{"CircleClass1", "CircleClass2"},
			"TheList1":   [][1]string{{"CircleClass1"}, {"CircleClass2"}},
			"TheList2":   [][2]string{{"CircleClass1"}, {"CircleClass2"}},
		},
	)
	if err != nil {
		panic(err)
	}
	fields := endpoint.GetFields()
	if strings.ReplaceAll(fields["TheString0"].String()," ", "") != "single_struct:{className:\"Circle\"}" {
		t.Errorf("%#v", fields["TheString0"].String())
	}
	if strings.ReplaceAll(fields["TheString1"].String()," ", "") != "single_struct:{className:\"Circle\"nLabels:1}" {
		t.Errorf("%#v", fields["TheString1"].String())
	}
	if strings.ReplaceAll(fields["TheString2"].String()," ", "") != "single_struct:{className:\"Circle\"nLabels:2}" {
		t.Errorf("%#v", fields["TheString2"].String())
	}
	if strings.ReplaceAll(fields["TheList0"].String()," ", "") != "list_struct:{list_fields:{className:\"CircleClass1\"}list_fields:{className:\"CircleClass2\"}}" {
		t.Errorf("%#v", fields["TheList0"].String())
	}
	if strings.ReplaceAll(fields["TheList1"].String()," ", "") != "list_struct:{list_fields:{className:\"CircleClass1\"nLabels:1}list_fields:{className:\"CircleClass2\"nLabels:1}}" {
		t.Errorf("%#v", fields["TheList1"].String())
	}
	if strings.ReplaceAll(fields["TheList2"].String()," ", "") != "list_struct:{list_fields:{className:\"CircleClass1\"nLabels:2}list_fields:{className:\"CircleClass2\"nLabels:2}}" {
		t.Errorf("%#v", fields["TheList2"].String())
	}
}

func TestCommonStruct(t *testing.T) {
	endpoint, err := NewStruct(
		"Geo", map[string]interface{}{
			"Shape1": [2]interface{}{
				"Class1", map[string]interface{}{"Field1": "Circle1"}},
			"Shape2": [3]interface{}{
				"Class2", map[string]interface{}{"Field2": [2]string{"Circle2"}}},
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
	field2Endpoint := field2Fields["Field2"].GetSingleStruct()
	if endpoint.ClassName != "Geo" || endpoint.NLabels != 0 ||
		shape2Endpoint.ClassName != "Class2" || shape2Endpoint.NLabels != 1 ||
		field2Endpoint.ClassName != "Circle2" || field2Endpoint.NLabels != 2 {
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
