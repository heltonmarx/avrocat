package main

import "testing"

func TestAvroTransformer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Transform([]byte(tt.input))
			if err != nil {
				t.Fatalf("Transform() error = %v", err)
			}
			if string(result) != tt.expected {
				t.Errorf("Transform() = %s, want %s", result, tt.expected)
			}
		})
	}
}
