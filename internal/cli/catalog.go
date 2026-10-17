package cli

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
)

// Parameter describes a named path identifier, query option or JSON body field.
type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// Operation is one named application command backed by an existing handler.
// A static query in Path selects a specific representation of a shared endpoint.
type Operation struct {
	Name        string      `json:"name"`
	Method      string      `json:"method"`
	Path        string      `json:"path"`
	Description string      `json:"description"`
	Response    string      `json:"response"`
	Effect      string      `json:"effect"`
	Parameters  []Parameter `json:"parameters"`
	Body        bool        `json:"body"`
	Multipart   string      `json:"multipart,omitempty"`
	Source      string      `json:"source"`
}

//go:embed operations.json
var operationsJSON []byte

var operationCatalog = loadOperations()

func loadOperations() []Operation {
	var operations []Operation
	decoder := json.NewDecoder(bytes.NewReader(operationsJSON))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&operations); err != nil {
		panic(fmt.Sprintf("invalid embedded command catalog: %v", err))
	}
	names := make(map[string]bool, len(operations))
	for _, operation := range operations {
		if operation.Name == "" || names[operation.Name] {
			panic(fmt.Sprintf("invalid or duplicate catalog command: %q", operation.Name))
		}
		names[operation.Name] = true
	}
	return operations
}

// Operations returns the complete command catalog. Callers may sort or annotate
// their copy without modifying the embedded definitions used by other callers.
func Operations() []Operation {
	operations := append([]Operation(nil), operationCatalog...)
	for i := range operations {
		operations[i].Parameters = append([]Parameter(nil), operations[i].Parameters...)
	}
	return operations
}
