package goseq

import (
	"reflect"
	"sort"
	"testing"
)

var (
	err_type_coercion  string = "Value returned is not value input (unexpected type coercion)!"
	err_val_corruption        = "Value set is not value returned (value corruption)!"
)

func testFilterSuite(keys map[string]interface{}) (Filter, error) {
	fil := NewFilter()
	for key, val := range keys {
		err := fil.Set(key, val)
		if err != nil {
			return fil, err
		}
	}
	return fil, nil
}

// some other, non-supported type
type _unknown struct {
	x string
	y int32
	z bool
}

func TestFilter_Set_string(t *testing.T) {
	_, err := testFilterSuite(map[string]interface{}{"somekey": "someval"})
	if err != nil {
		t.Log("Unexpected error on string value:")
		t.Log(err)
		t.FailNow()
	}
}

func TestFilter_Set_int(t *testing.T) {
	_, err := testFilterSuite(map[string]interface{}{"somekey52": 43836})
	if err != nil {
		t.Log("Unexpected error on int value:")
		t.Log(err)
		t.FailNow()
	}
}

func TestFilter_Set_bool(t *testing.T) {
	_, err := testFilterSuite(map[string]interface{}{"somekey3": true})
	if err != nil {
		t.Log("Unexpected error on bool value:")
		t.Log(err)
		t.FailNow()
	}
}

func TestFilter_Set_unsupported(t *testing.T) {
	_, err := testFilterSuite(map[string]interface{}{"letestkey": _unknown{}})
	if err == nil {
		t.Log("Expected error for unknown value type!")
		t.FailNow()
	}
}

func TestFilter_Get_string(t *testing.T) {
	key := "somekey"
	value := "somevalue"

	fil, _ := testFilterSuite(map[string]interface{}{key: value})
	interfaceback, err := fil.Get(key)
	if err != nil {
		t.Log("Unexpected error:")
		t.Log(err)
		t.FailNow()
	}

	valback, ok := interfaceback.(string)

	if !ok {
		t.Log(err_type_coercion)
	}

	if valback != value {
		t.Log(err_val_corruption)
	}
}

func TestFilter_Get_int(t *testing.T) {
	key := "as83hasj_sa"
	value := 38263

	fil, _ := testFilterSuite(map[string]interface{}{key: value})
	interfaceback, err := fil.Get(key)
	if err != nil {
		t.Log("Unexpected error:")
		t.Log(err)
		t.FailNow()
	}

	valback, ok := interfaceback.(int)

	if !ok {
		t.Log(err_type_coercion)
		t.FailNow()
	}

	if valback != value {
		t.Log(err_val_corruption)
		t.FailNow()
	}
}

func TestFilter_Get_bool(t *testing.T) {
	key := "uhsdoygfas"
	value := false

	fil, _ := testFilterSuite(map[string]interface{}{key: value})
	interfaceback, err := fil.Get(key)
	if err != nil {
		t.Log("Unexpected error:")
		t.Log(err)
		t.FailNow()
	}

	valback, ok := interfaceback.(bool)

	if !ok {
		t.Log(err_type_coercion)
		t.FailNow()
	}

	if valback != value {
		t.Log(err_val_corruption)
		t.FailNow()
	}
}

func TestFilter_Has(t *testing.T) {
	key := "48ahusn4sa"
	fil, _ := testFilterSuite(map[string]interface{}{key: 84724})
	if !fil.Has(key) {
		t.Log("Has did not register our key!")
		t.FailNow()
	}
}

func TestFilter_Keys(t *testing.T) {
	valmap := map[string]interface{}{
		"string": "what what",
		"int":    3242,
		"bool":   true,
	}

	actualkeys := make([]string, len(valmap))

	for key, _ := range valmap {
		actualkeys = append(actualkeys, key)
	}

	fil, err := testFilterSuite(valmap)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	sortedActual := sort.StringSlice(actualkeys)
	sortedActual.Sort()
	sortedReported := sort.StringSlice(fil.Keys())
	sortedReported.Sort()

	if !reflect.DeepEqual(sortedActual, sortedReported) {
		t.Log("Array keys are not the same!")
	}
}
