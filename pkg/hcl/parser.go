package hcl

import (
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

// Resource represents a Kubernetes resource defined in HCL
type Resource struct {
	Type       string
	Name       string
	Attributes map[string]interface{}
}

// Variable represents a variable definition in HCL
type Variable struct {
	Name        string
	Type        string
	Default     interface{}
	Description string
}

// Parser handles HCL parsing
type Parser struct {
	parser *hclparse.Parser
}

// NewParser creates a new HCL parser
func NewParser() *Parser {
	return &Parser{
		parser: hclparse.NewParser(),
	}
}

// ParseFile parses an HCL file and extracts resources and variables
func (p *Parser) ParseFile(filename string) ([]Resource, []Variable, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	return p.ParseBytes(content, filename)
}

// ParseBytes parses HCL content from bytes
func (p *Parser) ParseBytes(content []byte, filename string) ([]Resource, []Variable, error) {
	file, diags := p.parser.ParseHCL(content, filename)
	if diags.HasErrors() {
		return nil, nil, fmt.Errorf("failed to parse HCL: %s", diags.Error())
	}

	resources, variables, err := p.extractContent(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract content: %w", err)
	}

	return resources, variables, nil
}

// extractContent extracts resources and variables from HCL file
func (p *Parser) extractContent(file *hcl.File) ([]Resource, []Variable, error) {
	content, diags := file.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "resource", LabelNames: []string{"type", "name"}},
			{Type: "variable", LabelNames: []string{"name"}},
		},
	})
	if diags.HasErrors() {
		return nil, nil, fmt.Errorf("failed to get content: %s", diags.Error())
	}

	var resources []Resource
	var variables []Variable

	// Parse resources
	for _, block := range content.Blocks {
		switch block.Type {
		case "resource":
			resource, err := p.parseResource(block)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse resource: %w", err)
			}
			resources = append(resources, resource)
		case "variable":
			variable, err := p.parseVariable(block)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse variable: %w", err)
			}
			variables = append(variables, variable)
		}
	}

	return resources, variables, nil
}

// parseResource parses a resource block
func (p *Parser) parseResource(block *hcl.Block) (Resource, error) {
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return Resource{}, fmt.Errorf("failed to get attributes: %s", diags.Error())
	}

	attributes := make(map[string]interface{})
	for name, attr := range attrs {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return Resource{}, fmt.Errorf("failed to evaluate attribute %s: %s", name, diags.Error())
		}
		attributes[name] = ctyToInterface(val)
	}

	return Resource{
		Type:       block.Labels[0],
		Name:       block.Labels[1],
		Attributes: attributes,
	}, nil
}

// parseVariable parses a variable block
func (p *Parser) parseVariable(block *hcl.Block) (Variable, error) {
	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		return Variable{}, fmt.Errorf("failed to get attributes: %s", diags.Error())
	}

	variable := Variable{
		Name: block.Labels[0],
	}

	for name, attr := range attrs {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			continue // Some attributes might reference other values
		}

		switch name {
		case "type":
			if val.Type() == cty.String {
				variable.Type = val.AsString()
			}
		case "default":
			variable.Default = ctyToInterface(val)
		case "description":
			if val.Type() == cty.String {
				variable.Description = val.AsString()
			}
		}
	}

	return variable, nil
}

// ctyToInterface converts cty.Value to Go interface{}
func ctyToInterface(val cty.Value) interface{} {
	if val.IsNull() {
		return nil
	}

	switch val.Type() {
	case cty.String:
		return val.AsString()
	case cty.Number:
		bf := val.AsBigFloat()
		if f, accuracy := bf.Float64(); accuracy == 0 {
			return f
		}
		return bf.String()
	case cty.Bool:
		return val.True()
	}

	if val.Type().IsListType() || val.Type().IsTupleType() {
		var result []interface{}
		for it := val.ElementIterator(); it.Next(); {
			_, v := it.Element()
			result = append(result, ctyToInterface(v))
		}
		return result
	}

	if val.Type().IsMapType() || val.Type().IsObjectType() {
		result := make(map[string]interface{})
		for it := val.ElementIterator(); it.Next(); {
			k, v := it.Element()
			result[k.AsString()] = ctyToInterface(v)
		}
		return result
	}

	return val.AsString()
}
