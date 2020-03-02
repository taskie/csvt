package csvt

import (
	"fmt"
	"testing"
)

var records01 = [][]string{
	{"A", "B", "C"},
	{"1", "2", "3"},
	{"4", "5", "6"},
	{"7", "8", "9"},
}

var records01t = [][]string{
	{"A", "1", "4", "7"},
	{"B", "2", "5", "8"},
	{"C", "3", "6", "9"},
}

var items01 = []map[string]string{
	{"A": "1", "B": "2", "C": "3"},
	{"A": "4", "B": "5", "C": "6"},
	{"A": "7", "B": "8", "C": "9"},
}

var slice01 = [][]string{
	{"5", "6"},
	{"8", "9"},
}

var records02 = [][]string{
	{"a"},
	{"b", "c", "d"},
	{"e", "f"},
}

var records02t = [][]string{
	{"a", "b", "e"},
	{"", "c", "f"},
	{"", "d", ""},
}

var items02 = []map[string]string{
	{"a": "b"},
	{"a": "e"},
}

var records03 = [][]string{
	{"a", "c"},
	{"b", ""},
	{"", "d"},
}

var items03 = []map[string]string{
	{"a": "b"},
	{"c": "d"},
}

func assertEqual(expected interface{}, actual interface{}) error {
	if expected != actual {
		return fmt.Errorf("expected: %v, actual: %v", expected, actual)
	}
	return nil
}

func assertEqualStrings(expected []string, actual []string) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("invalid length: expected: %d, actual: %d", len(expected), len(actual))
	}
	for i, e := range expected {
		err := assertEqual(e, actual[i])
		if err != nil {
			return fmt.Errorf("[%d]: %s", i, err.Error())
		}
	}
	return nil
}

func assertEqualRecords(expected [][]string, actual [][]string) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("invalid length: expected: %d, actual: %d", len(expected), len(actual))
	}
	for i, e := range expected {
		err := assertEqualStrings(e, actual[i])
		if err != nil {
			return fmt.Errorf("[%d]: %s", i, err.Error())
		}
	}
	return nil
}

func assertEqualStringStringMap(expected map[string]string, actual map[string]string) error {
	for k, e := range expected {
		if a, ok := actual[k]; ok {
			err := assertEqual(e, a)
			if err != nil {
				return fmt.Errorf("[\"%s\"]: %s", k, err.Error())
			}
		} else {
			return fmt.Errorf("[\"%s\"]: no entry", k)
		}
	}
	return nil
}

func assertEqualItems(expected []map[string]string, actual []map[string]string) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("invalid length: expected: %d, actual: %d", len(expected), len(actual))
	}
	for i, e := range expected {
		err := assertEqualStringStringMap(e, actual[i])
		if err != nil {
			return fmt.Errorf("[%d]: %s", i, err.Error())
		}
	}
	return nil
}

func TestTransposer01(t *testing.T) {
	transposer := Transposer{}
	actualRecords01t, err := transposer.Transpose(records01)
	if err != nil {
		t.Fatal(err)
	}
	err = assertEqualRecords(records01t, actualRecords01t)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTransposer01LengthChecked(t *testing.T) {
	transposer := Transposer{LengthChecked: true}
	actualRecords01t, err := transposer.Transpose(records01)
	if err != nil {
		t.Fatal(err)
	}
	err = assertEqualRecords(records01t, actualRecords01t)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTransposer02(t *testing.T) {
	transposer := Transposer{}
	actualRecords01t, err := transposer.Transpose(records02)
	if err != nil {
		t.Fatal(err)
	}
	err = assertEqualRecords(records02t, actualRecords01t)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTransposerLengthChecked02(t *testing.T) {
	transposer := Transposer{LengthChecked: true}
	_, err := transposer.Transpose(records02)
	if err == nil {
		t.Fatal(fmt.Errorf("must fail in strict mode"))
	}
}

func TestMapper01(t *testing.T) {
	mapper := Mapper{}
	actualItems01, err := mapper.MapAll(records01)
	if err != nil {
		t.Fatal(err)
	}
	err = assertEqualItems(items01, actualItems01)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMapper02(t *testing.T) {
	mapper := Mapper{}
	actualItems02, err := mapper.MapAll(records02)
	if err != nil {
		t.Fatal(err)
	}
	err = assertEqualItems(items02, actualItems02)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMapperLengthChecked02(t *testing.T) {
	mapper := Mapper{LengthChecked: true}
	_, err := mapper.MapAll(records02)
	if err == nil {
		t.Fatal(fmt.Errorf("must fail in strict mode"))
	}
}

func TestUnmapper01(t *testing.T) {
	unmapper := Unmapper{}
	actualRecords01, err := unmapper.UnmapAll(items01)
	if err != nil {
		t.Fatal(err)
	}
	err = assertEqualRecords(records01, actualRecords01)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmapper03(t *testing.T) {
	unmapper := Unmapper{}
	actualRecords03, err := unmapper.UnmapAll(items03)
	if err != nil {
		t.Fatal(err)
	}
	err = assertEqualRecords(records03, actualRecords03)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmapperKeyChecked03(t *testing.T) {
	unmapper := Unmapper{KeyChecked: true}
	_, err := unmapper.UnmapAll(items03)
	if err == nil {
		t.Fatal(fmt.Errorf("must fail in strict mode"))
	}
}

func TestUnmapperKeyCheckedEach03(t *testing.T) {
	unmapper := Unmapper{KeyChecked: true}
	for i, item := range items03 {
		_, err := unmapper.Unmap(item)
		if i == 1 {
			if err == nil {
				t.Fatal(fmt.Errorf("must fail in strict mode"))
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestSlicer01(t *testing.T) {
	slicer := Slicer{
		RowRanges: Ranges{NewRange(2, 4)},
		ColRanges: Ranges{NewRange(1, 3)},
	}
	actualSlice01, err := slicer.Slice(records01)
	if err != nil {
		t.Fatal(err)
	}
	err = assertEqualRecords(slice01, actualSlice01)
	if err != nil {
		t.Fatal(err)
	}
}
