package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
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
	if ok := isArray(buf); !ok {
		return buf, nil
	}
	nodes := make([]Node, 0)
	err := json.Unmarshal(buf, &nodes)
	switch {
	case err != nil:
		return nil, err
	case len(nodes) == 0:
		return buf, nil
	}

	schemasMapping, err := generateSchemaMapping(nodes)
	if err != nil {
		return nil, err
	}

	for _, node := range schemasMapping {
		updateSchemaReferences(node, schemasMapping)
	}

	for _, node := range nodes {
		namespace := trimSuffix(node.Namespace)
		if namespace == node.Name {
			return json.Marshal(node)
		}
	}
	return nil, errors.New("could not transform schema")
}

// generateSchemaMapping generates an node map by reference.
func generateSchemaMapping(nodes []Node) (map[string]Node, error) {
	schemasMapping := make(map[string]Node)
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
			m := item
			for i, obj := range m {
				v := reflect.ValueOf(obj)
				if node, ok := buildNodeByMapping(v, schemasMapping); ok {
					// remove object from node
					m = slices.Delete(m, i, i+1)
					// append the new node
					m = append(m, Node{
						Name:   node.Name,
						Type:   node.Type,
						Fields: node.Fields,
					})
				}
			}
		default:
			v := reflect.ValueOf(item)
			if node, ok := buildNodeByMapping(v, schemasMapping); ok {
				updateElement(v, element, node)
			}
		}
	}
}

func updateElement(v reflect.Value, element map[string]interface{}, node Node) {
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

// buildNodeByMapping substitute a reference by a node.
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
	case reflect.Ptr, reflect.Interface:
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
	return emptyNode, false
}

func trimSuffix(namespace string) string {
	n := strings.LastIndex(namespace, ".")
	if n == 0 || n == len(namespace) {
		return namespace
	}
	return namespace[n+1:]
}

func isArray(buf []byte) bool {
	v := bytes.TrimLeft(buf, " \t\r\n")
	return len(v) > 0 && v[0] == '['
}
