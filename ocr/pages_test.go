package ocr

import (
	"reflect"
	"testing"
)

func TestParsePageSpec_Valid(t *testing.T) {
	cases := []struct {
		spec string
		want []int
	}{
		{"12-14", []int{12, 13, 14}},
		{"23,34,36", []int{23, 34, 36}},
		{"1-3,5,6-8", []int{1, 2, 3, 5, 6, 7, 8}},
		{"5", []int{5}},
		{"3,1,2", []int{1, 2, 3}},
		{"2-4,2-3", []int{2, 3, 4}},
		{"  1 - 3 ", []int{1, 2, 3}},
	}

	for _, tc := range cases {
		got, err := ParsePageSpec(tc.spec)
		if err != nil {
			t.Errorf("ParsePageSpec(%q) unexpected error: %v", tc.spec, err)
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("ParsePageSpec(%q) = %v, want %v", tc.spec, got, tc.want)
		}
	}
}

func TestParsePageSpec_Range1To999(t *testing.T) {
	got, err := ParsePageSpec("1-999")
	if err != nil {
		t.Fatalf("ParsePageSpec(1-999) unexpected error: %v", err)
	}
	if len(got) != 999 || got[0] != 1 || got[len(got)-1] != 999 {
		t.Fatalf("ParsePageSpec(1-999) length=%d first=%d last=%d", len(got), got[0], got[len(got)-1])
	}
}

func TestParsePageSpec_Errors(t *testing.T) {
	bad := []string{
		"",
		"   ",
		"5-2",
		"abc",
		"1,x,3",
		"5-",
		"-3",
		"0",
		"0-3",
		"1--3",
		"1-2-3",
		"1,,3",
	}
	for _, spec := range bad {
		if _, err := ParsePageSpec(spec); err == nil {
			t.Errorf("ParsePageSpec(%q) expected error, got nil", spec)
		}
	}
}

func TestSelectPages_DropsAboveMax(t *testing.T) {
	got, err := SelectPages("1-3,5,6-8", 6)
	if err != nil {
		t.Fatalf("SelectPages unexpected error: %v", err)
	}
	want := []int{1, 2, 3, 5, 6}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SelectPages = %v, want %v", got, want)
	}
}

func TestSelectPages_Unbounded(t *testing.T) {
	got, err := SelectPages("1-10", 0)
	if err != nil {
		t.Fatalf("SelectPages unexpected error: %v", err)
	}
	want := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SelectPages = %v, want %v", got, want)
	}
}

func TestSelectPages_MalformedPropagates(t *testing.T) {
	if _, err := SelectPages("5-2", 100); err == nil {
		t.Errorf("SelectPages expected error for malformed spec")
	}
}
