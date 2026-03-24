package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Node holds the avro node json object
type Node struct {
	Namespace string           `json:"namespace,omitempty"`
	Name      string           `json:"name"`
	Type      string           `json:"type"`
	Fields    []map[string]any `json:"fields"`
}

// Transform transforms an array of avro json schemas into only one avro json schema.
func Transform(buf []byte) ([]byte, error) {
	// First, check if it's valid JSON
	if !isJSONArray(buf) {
		// If it's not an array, it might be a single object
		// Check if it's valid JSON
		var temp interface{}
		if err := json.Unmarshal(buf, &temp); err != nil {
			return nil, fmt.Errorf("it's not a valid JSON: %w", err)
		}
		return buf, nil
	}

	nodes := make([]Node, 0)
	err := json.Unmarshal(buf, &nodes)
	switch {
	case err != nil:
		return nil, err
	case len(nodes) == 0:
		return buf, nil
	case len(nodes) == 1:
		// If there's only one node, return it as-is
		return json.Marshal(nodes[0])
	}

	schemasMapping, err := generateSchemaMapping(nodes)
	if err != nil {
		return nil, err
	}

	// Update references in all nodes
	for key, node := range schemasMapping {
		updateSchemaReferences(node, schemasMapping)
		schemasMapping[key] = node
	}

	// Find the root node (the one whose name matches the last part of its namespace)
	for _, node := range nodes {
		// If node has no namespace, it can't be the root
		if node.Namespace == "" {
			continue
		}
		namespace := trimSuffix(node.Namespace)
		if namespace == node.Name {
			// Get the updated node from schemasMapping
			refName := strings.Join([]string{node.Namespace, node.Name}, ".")
			if updatedNode, exists := schemasMapping[refName]; exists {
				return json.Marshal(updatedNode)
			}
			return json.Marshal(node)
		}
	}

	// If no root found, return the last node
	return json.Marshal(nodes[len(nodes)-1])
}

// generateSchemaMapping generates an node map by reference.
func generateSchemaMapping(nodes []Node) (map[string]Node, error) {
	schemasMapping := make(map[string]Node, len(nodes))
	for _, node := range nodes {
		if node.Namespace != "" && node.Name != "" {
			refName := strings.Join([]string{node.Namespace, node.Name}, ".")
			schemasMapping[refName] = node
		}
	}
	return schemasMapping, nil
}

// updateSchemaReferences update the node references by nodes.
func updateSchemaReferences(node Node, schemasMapping map[string]Node) {
	for _, element := range node.Fields {
		item := element["type"]
		switch item := item.(type) {
		case []any:
			// Handle union types
			newItems := make([]any, 0, len(item))
			for _, obj := range item {
				v := reflect.ValueOf(obj)
				if node, ok := buildNodeByMapping(v, schemasMapping); ok {
					newItems = append(newItems, Node{
						Name:   node.Name,
						Type:   node.Type,
						Fields: node.Fields,
					})
				} else {
					newItems = append(newItems, obj)
				}
			}
			element["type"] = newItems
		default:
			v := reflect.ValueOf(item)
			if node, ok := buildNodeByMapping(v, schemasMapping); ok {
				updateElement(v, element, node)
			}
		}
	}
}

func updateElement(v reflect.Value, element map[string]any, node Node) {
	switch v.Kind() {
	case reflect.String:
		delete(element, "type")
		element["type"] = Node{
			Name:   node.Name,
			Type:   node.Type,
			Fields: node.Fields,
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			if key.Kind() == reflect.String && key.String() == "items" {
				v.SetMapIndex(key, reflect.ValueOf(node))
			}
		}
	}
}

// buildNodeByMapping attempts to resolve a schema reference by looking up the provided reflect.Value in the schemasMapping.
// It accepts a reflect.Value representing a potential schema reference and a map of schema nodes.
// Returns the resolved Node and a boolean indicating success.
func buildNodeByMapping(v reflect.Value, schemasMapping map[string]Node) (Node, bool) {
	var emptyNode Node

	switch v.Kind() {
	case reflect.String:
		key := v.String()
		if node, ok := schemasMapping[key]; ok {
			return Node{
				Name:   node.Name,
				Type:   node.Type,
				Fields: node.Fields,
			}, true
		}
	case reflect.Pointer, reflect.Interface:
		if !v.IsNil() {
			return buildNodeByMapping(v.Elem(), schemasMapping)
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			if key.Kind() == reflect.String && key.String() == "items" {
				value := v.MapIndex(key)
				return buildNodeByMapping(value, schemasMapping)
			}
		}
	case reflect.Slice:
		for i := range v.Len() {
			elem := v.Index(i)
			if !elem.IsNil() {
				return buildNodeByMapping(elem, schemasMapping)
			}
		}
	}
	// Return false to indicate that no matching node was found in the mapping.
	return emptyNode, false
}

func trimSuffix(namespace string) string {
	n := strings.LastIndex(namespace, ".")
	switch {
	case n <= 0:
		return namespace
	case n == len(namespace)-1:
		return namespace[:n]
	}
	return namespace[n+1:]
}

func isJSONArray(buf []byte) bool {
	for _, b := range buf {
		if b == ' ' || b == '\n' || b == '\t' || b == '\r' {
			continue
		}
		return b == '['
	}
	return false
}
