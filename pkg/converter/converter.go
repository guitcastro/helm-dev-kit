package converter

import (
	"fmt"

	"github.com/guitcastro/helm-dev-kit/pkg/hcl"
	"github.com/guitcastro/helm-dev-kit/pkg/helm"
	"github.com/guitcastro/helm-dev-kit/pkg/validator"
)

// HCLToHelm converts HCL to Helm chart with validation
type HCLToHelm struct {
	parser    *hcl.Parser
	converter *helm.Converter
	validator *validator.Validator
}

// NewHCLToHelm creates a new HCL to Helm converter
func NewHCLToHelm(chartName string, validator *validator.Validator) *HCLToHelm {
	return &HCLToHelm{
		parser:    hcl.NewParser(),
		converter: helm.NewConverter(chartName),
		validator: validator,
	}
}

// ConvertFile converts an HCL file to a validated Helm chart
func (h *HCLToHelm) ConvertFile(filename string) (*helm.Chart, error) {
	// Parse HCL file
	resources, variables, err := h.parser.ParseFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HCL file: %w", err)
	}

	// Convert to Helm chart
	chart, err := h.converter.Convert(resources, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to Helm chart: %w", err)
	}

	// Validate templates if validator is available
	if h.validator != nil {
		if err := h.validateChart(chart); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}
	}

	return chart, nil
}

// ConvertBytes converts HCL content from bytes to a validated Helm chart
func (h *HCLToHelm) ConvertBytes(content []byte, filename string) (*helm.Chart, error) {
	// Parse HCL content
	resources, variables, err := h.parser.ParseBytes(content, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HCL content: %w", err)
	}

	// Convert to Helm chart
	chart, err := h.converter.Convert(resources, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to Helm chart: %w", err)
	}

	// Validate templates if validator is available
	if h.validator != nil {
		if err := h.validateChart(chart); err != nil {
			return nil, fmt.Errorf("validation failed: %w", err)
		}
	}

	return chart, nil
}

// validateChart validates all templates in the chart
func (h *HCLToHelm) validateChart(chart *helm.Chart) error {
	var templates []string
	for _, template := range chart.Templates {
		templates = append(templates, template.Content)
	}

	errors := h.validator.ValidateChart(templates)
	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %v", errors)
	}

	return nil
}
