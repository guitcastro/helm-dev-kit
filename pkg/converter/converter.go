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

// Convert converts all HCL files in a directory to a validated Helm chart
func (h *HCLToHelm) Convert(dirPath string) (*helm.Chart, error) {
	// Parse all HCL files in directory
	resources, variables, err := h.parser.ParseDirectory(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HCL directory: %w", err)
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
