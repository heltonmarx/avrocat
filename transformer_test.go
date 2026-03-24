package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestAvroTransformer(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:     "Single schema",
			input:    `{"namespace": "com.example", "name": "user", "type": "record", "fields": [{"name": "id", "type": "int"}]}`,
			expected: `{"namespace": "com.example", "name": "user", "type": "record", "fields": [{"name": "id", "type": "int"}]}`,
		},
		{
			name: "Multiple schemas",
			input: `[
				{"namespace": "com.example.fulfilled", "name": "user", "type": "record", "fields": [{"name": "id", "type": "int"}]},
				{"namespace": "com.example.fulfilled", "name": "order", "type": "record", "fields": [{"name": "orderId", "type": "string", "name": "user", "type": "com.example.fulfilled.user"}]},
				{"namespace": "com.example.fulfilled", "name": "fulfilled", "type": "record", "fields": [{"name": "order", "type": "com.example.fulfilled.order"}]}
			]`,
			expected: `{"namespace":"com.example.fulfilled","name":"fulfilled","type":"record","fields":[{"name":"order","type":{"name":"order","type":"record","fields":[{"name":"user","type":{"name":"user","type":"record","fields":[{"name":"id","type":"int"}]}}]}}]}`,
		},
		{
			name:     "Empty array",
			input:    `[]`,
			expected: `[]`,
		},
		{
			name:     "Array with single element",
			input:    `[{"namespace": "test", "name": "single", "type": "record", "fields": [{"name": "value", "type": "string"}]}]`,
			expected: `{"namespace":"test","name":"single","type":"record","fields":[{"name":"value","type":"string"}]}`,
		},
		{
			name:        "Invalid JSON",
			input:       `{invalid json}`,
			expectError: true,
		},
		{
			name: "Schemas with array types",
			input: `[
				{"namespace": "com.example", "name": "item", "type": "record", "fields": [{"name": "id", "type": "int"}]},
				{"namespace": "com.example", "name": "container", "type": "record", "fields": [{"name": "items", "type": {"type": "array", "items": "com.example.item"}}]}
			]`,
			expected: `{"namespace":"com.example","name":"container","type":"record","fields":[{"name":"items","type":{"type":"array","items":{"name":"item","type":"record","fields":[{"name":"id","type":"int"}]}}}]}`,
		},
		{
			name: "Schemas with union types",
			input: `[
				{"namespace": "com.example", "name": "A", "type": "record", "fields": [{"name": "field", "type": "string"}]},
				{"namespace": "com.example", "name": "B", "type": "record", "fields": [{"name": "field", "type": "int"}]},
				{"namespace": "com.example", "name": "C", "type": "record", "fields": [{"name": "value", "type": ["null", "com.example.A", "com.example.B"]}]}
			]`,
			expected: `{"namespace":"com.example","name":"C","type":"record","fields":[{"name":"value","type":["null",{"name":"A","type":"record","fields":[{"name":"field","type":"string"}]},{"name":"B","type":"record","fields":[{"name":"field","type":"int"}]}]}]}`,
		},
		{
			name: "Schemas without namespace",
			input: `[
				{"name": "Simple", "type": "record", "fields": [{"name": "id", "type": "int"}]}
			]`,
			expected: `{"name":"Simple","type":"record","fields":[{"name":"id","type":"int"}]}`,
		},
		{
			name: "Nested references",
			input: `[
				{"namespace": "test", "name": "base", "type": "record", "fields": [{"name": "id", "type": "int"}]},
				{"namespace": "test", "name": "middle", "type": "record", "fields": [{"name": "base", "type": "test.base"}]},
				{"namespace": "test", "name": "top", "type": "record", "fields": [{"name": "middle", "type": "test.middle"}]}
			]`,
			expected: `{"namespace":"test","name":"top","type":"record","fields":[{"name":"middle","type":{"name":"middle","type":"record","fields":[{"name":"base","type":{"name":"base","type":"record","fields":[{"name":"id","type":"int"}]}}]}}]}`,
		},
		{
			name: "Empty fields array",
			input: `{"namespace": "test", "name": "Empty", "type": "record", "fields": []}`,
			expected: `{"namespace": "test", "name": "Empty", "type": "record", "fields": []}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Transform([]byte(tt.input))
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Transform() expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Transform() error = %v", err)
			}
			
			// Parse both JSON to compare structures
			var gotObj, wantObj interface{}
			if err := json.Unmarshal(result, &gotObj); err != nil {
				t.Fatalf("Failed to parse result JSON: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.expected), &wantObj); err != nil {
				t.Fatalf("Failed to parse expected JSON: %v", err)
			}
			
			// Compare using DeepEqual
			if !reflect.DeepEqual(gotObj, wantObj) {
				// Marshal both for error message
				gotBytes, _ := json.MarshalIndent(gotObj, "", "  ")
				wantBytes, _ := json.MarshalIndent(wantObj, "", "  ")
				t.Errorf("Transform() mismatch\nGot:\n%s\nWant:\n%s", gotBytes, wantBytes)
			}
		})
	}
}

func TestIsJSONArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Empty string", "", false},
		{"Whitespace only", "   \n\t\r", false},
		{"Starts with [", "[", true},
		{"Starts with [ after whitespace", "  [", true},
		{"Starts with {", "{", false},
		{"Valid array", `[1,2,3]`, true},
		{"Valid object", `{"a":1}`, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isJSONArray([]byte(tt.input))
			if result != tt.expected {
				t.Errorf("isJSONArray(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTrimSuffix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", ""},
		{"No dot", "namespace", "namespace"},
		{"Single dot at start", ".suffix", ".suffix"},
		{"Multiple dots", "com.example.test", "test"},
		{"Dot at end", "test.", "test"},
		{"Multiple consecutive dots", "com..test", "test"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimSuffix(tt.input)
			if result != tt.expected {
				t.Errorf("trimSuffix(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
