package blocker

import (
	"reflect"
	"testing"
)

func TestRemoveComment(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{name: "Full line comment", input: []byte("# This is a full line comment"), expected: []byte("This is a full line comment")},
		{name: "Inline comment", input: []byte("code here # inline comment"), expected: []byte("code here # inline comment")},
		{name: "Whitespace only", input: []byte("   "), expected: []byte("   ")},
		{name: "No comment", input: []byte("no comment on this line"), expected: []byte("no comment on this line")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := stripComment(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected: %v, got: %v", tc.expected, result)
			}
		})
	}
}

func TestPrependComment(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{name: "No comment", input: []byte("no comment on this line"), expected: []byte("# no comment on this line")},
		{name: "Inline comment", input: []byte("code here # inline comment"), expected: []byte("# code here # inline comment")},
		{name: "Whitespace only", input: []byte("   "), expected: []byte("#    ")},
		{name: "Full line comment", input: []byte("# This is a full line comment"), expected: []byte("# This is a full line comment")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := addComment(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected: %v, got: %v", tc.expected, result)
			}
		})
	}
}
