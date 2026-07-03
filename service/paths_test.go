package service

import (
	"testing"
)

func TestExtractNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"1_file.md", 1},
		{"10_doc.md", 10},
		{"file.md", 0},
		{"02_test.md", 2},
	}

	for _, tt := range tests {
		result := ExtractNumber(tt.input)
		if result != tt.expected {
			t.Errorf("ExtractNumber(%s) = %d; want %d", tt.input, result, tt.expected)
		}
	}
}

func TestSortByNumber(t *testing.T) {
	input := []string{"3_b.md", "1_a.md", "2_c.md"}
	expected := []string{"1_a.md", "2_c.md", "3_b.md"}
	
	result := SortByNumber(input)
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("SortByNumber(%v) = %v; want %v", input, result, expected)
		}
	}
}

func TestEncodeUrl(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"  Spaces  ", "spaces"},
		{"UPPERCASE", "uppercase"},
	}

	for _, tt := range tests {
		result := EncodeUrl(tt.input)
		if result != tt.expected {
			t.Errorf("EncodeUrl(%s) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}
