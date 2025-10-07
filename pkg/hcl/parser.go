package hcl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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

// ParseDirectory parses all HCL files in a directory and merges resources and variables
func (p *Parser) ParseDirectory(dirPath string) ([]Resource, []Variable, error) {
	var allResources []Resource
	var allVariables []Variable

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-HCL files
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".hcl") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		resources, variables, err := p.parseBytes(content, path)
		if err != nil {
			return fmt.Errorf("failed to parse file %s: %w", path, err)
		}

		allResources = append(allResources, resources...)
		allVariables = append(allVariables, variables...)

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse directory: %w", err)
	}

	// Merge variables by name (later definitions override earlier ones)
	mergedVariables := p.mergeVariables(allVariables)

	return allResources, mergedVariables, nil
}

// parseBytes parses HCL content from bytes (internal method)
func (p *Parser) parseBytes(content []byte, filename string) ([]Resource, []Variable, error) {
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

// mergeVariables merges variables by name, with later definitions taking precedence
func (p *Parser) mergeVariables(variables []Variable) []Variable {
	variableMap := make(map[string]Variable)

	for _, variable := range variables {
		variableMap[variable.Name] = variable
	}

	var result []Variable
	for _, variable := range variableMap {
		result = append(result, variable)
	}

	return result
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
		// Try to extract variable references from the expression
		if varRef := p.extractVariableReference(attr.Expr); varRef != "" {
			attributes[name] = varRef
		} else {
			val, diags := attr.Expr.Value(nil)
			if diags.HasErrors() {
				// If evaluation fails, try to get the raw string representation
				if rawVal := p.getRawExpressionValue(attr.Expr); rawVal != "" {
					attributes[name] = rawVal
				} else {
					return Resource{}, fmt.Errorf("failed to evaluate attribute %s: %s", name, diags.Error())
				}
			} else {
				attributes[name] = ctyToInterface(val)
			}
		}
	}

	return Resource{
		Type:       block.Labels[0],
		Name:       block.Labels[1],
		Attributes: attributes,
	}, nil
}

// extractVariableReference extracts variable references from HCL expressions
func (p *Parser) extractVariableReference(expr hcl.Expression) string {
	// Convert to syntax to examine the expression
	if syntaxExpr, ok := expr.(*hclsyntax.ScopeTraversalExpr); ok {
		if len(syntaxExpr.Traversal) >= 2 {
			rootName := syntaxExpr.Traversal[0].(hcl.TraverseRoot).Name
			if rootName == "var" {
				varName := syntaxExpr.Traversal[1].(hcl.TraverseAttr).Name
				return "var." + varName
			}
		}
	}
	return ""
}

// getRawExpressionValue attempts to get a string representation of an expression
func (p *Parser) getRawExpressionValue(expr hcl.Expression) string {
	// This is a fallback for when we can't evaluate the expression
	// We'll return a placeholder that the Helm converter can handle
	if syntaxExpr, ok := expr.(*hclsyntax.ScopeTraversalExpr); ok {
		if len(syntaxExpr.Traversal) >= 2 {
			rootName := syntaxExpr.Traversal[0].(hcl.TraverseRoot).Name
			if rootName == "var" {
				varName := syntaxExpr.Traversal[1].(hcl.TraverseAttr).Name
				return "var." + varName
			}
		}
	}
	return ""
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
