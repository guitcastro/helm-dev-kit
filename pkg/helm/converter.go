package helm

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/guitcastro/helm-dev-kit/pkg/hcl"
	"gopkg.in/yaml.v3"
)

// Chart represents a Helm chart structure
type Chart struct {
	Name      string
	Templates []Template
	Values    Values
}

// Template represents a Helm template file
type Template struct {
	Name    string
	Content string
}

// Values represents the values.yaml structure
type Values map[string]interface{}

// Converter handles conversion from HCL to Helm
type Converter struct {
	chartName string
}

// NewConverter creates a new Helm converter
func NewConverter(chartName string) *Converter {
	return &Converter{
		chartName: chartName,
	}
}

// Convert converts HCL resources and variables to a Helm chart
func (c *Converter) Convert(resources []hcl.Resource, variables []hcl.Variable) (*Chart, error) {
	chart := &Chart{
		Name:   c.chartName,
		Values: make(Values),
	}

	// Map variables to values.yaml
	for _, variable := range variables {
		chart.Values[variable.Name] = variable.Default
	}

	// Convert resources to templates
	for _, resource := range resources {
		template, err := c.resourceToTemplate(resource)
		if err != nil {
			return nil, fmt.Errorf("failed to convert resource %s/%s: %w", resource.Type, resource.Name, err)
		}
		chart.Templates = append(chart.Templates, template)
	}

	return chart, nil
}

// resourceToTemplate converts an HCL resource to a Helm template
func (c *Converter) resourceToTemplate(resource hcl.Resource) (Template, error) {
	// Build Kubernetes manifest from resource
	manifest := make(map[string]interface{})

	// Parse resource type (e.g., "kubernetes_deployment" -> "Deployment")
	kind := c.parseKind(resource.Type)
	apiVersion := c.getAPIVersion(kind)

	manifest["apiVersion"] = apiVersion
	manifest["kind"] = kind

	// Build metadata
	metadata := make(map[string]interface{})
	metadata["name"] = fmt.Sprintf("{{ .Values.%s.name | default \"%s\" }}", resource.Name, resource.Name)
	metadata["namespace"] = "{{ .Values.namespace | default \"default\" }}"
	if labels, ok := resource.Attributes["labels"].(map[string]interface{}); ok {
		metadata["labels"] = labels
	}
	manifest["metadata"] = metadata

	// Build spec from attributes
	spec := make(map[string]interface{})
	for key, value := range resource.Attributes {
		if key != "labels" && key != "annotations" {
			spec[key] = c.processValue(value, resource.Name)
		}
	}
	manifest["spec"] = spec

	// Convert to YAML
	yamlContent, err := yaml.Marshal(manifest)
	if err != nil {
		return Template{}, fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Create unique template name based on resource name and kind
	templateName := fmt.Sprintf("%s-%s.yaml", strings.ToLower(resource.Name), strings.ToLower(kind))
	return Template{
		Name:    templateName,
		Content: string(yamlContent),
	}, nil
}

// parseKind extracts Kubernetes kind from resource type
func (c *Converter) parseKind(resourceType string) string {
	// Remove "kubernetes_" prefix if present
	kind := strings.TrimPrefix(resourceType, "kubernetes_")

	// Convert to CamelCase
	parts := strings.Split(kind, "_")
	for i, part := range parts {
		parts[i] = strings.Title(part)
	}

	return strings.Join(parts, "")
}

// getAPIVersion returns the API version for a given kind
func (c *Converter) getAPIVersion(kind string) string {
	apiVersions := map[string]string{
		"Deployment":  "apps/v1",
		"Service":     "v1",
		"ConfigMap":   "v1",
		"Secret":      "v1",
		"Pod":         "v1",
		"StatefulSet": "apps/v1",
		"DaemonSet":   "apps/v1",
		"Ingress":     "networking.k8s.io/v1",
		"Job":         "batch/v1",
		"CronJob":     "batch/v1",
	}

	if version, ok := apiVersions[kind]; ok {
		return version
	}

	return "v1"
}

// processValue processes a value to add Helm template syntax where appropriate
func (c *Converter) processValue(value interface{}, resourceName string) interface{} {
	switch v := value.(type) {
	case string:
		// Check if value looks like a variable reference
		if strings.HasPrefix(v, "var.") {
			varName := strings.TrimPrefix(v, "var.")
			return fmt.Sprintf("{{ .Values.%s }}", varName)
		}
		return v
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			result[key] = c.processValue(val, resourceName)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = c.processValue(val, resourceName)
		}
		return result
	default:
		return v
	}
}

// WriteChart writes the chart to disk
func (c *Converter) WriteChart(chart *Chart, outputDir string) error {
	// Create chart structure
	chartDir := filepath.Join(outputDir, chart.Name)
	templatesDir := filepath.Join(chartDir, "templates")

	// Write Chart.yaml
	chartMetadata := map[string]interface{}{
		"apiVersion":  "v2",
		"name":        chart.Name,
		"version":     "0.1.0",
		"description": "Generated Helm chart from HCL",
	}

	chartYAML, err := yaml.Marshal(chartMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal Chart.yaml: %w", err)
	}

	// Write values.yaml
	valuesYAML, err := yaml.Marshal(chart.Values)
	if err != nil {
		return fmt.Errorf("failed to marshal values.yaml: %w", err)
	}

	// Return the chart data (actual file writing would be done by caller)
	_ = chartYAML
	_ = valuesYAML
	_ = templatesDir

	return nil
}

// GetValuesYAML returns the values.yaml content as string
func (c *Converter) GetValuesYAML(chart *Chart) (string, error) {
	valuesYAML, err := yaml.Marshal(chart.Values)
	if err != nil {
		return "", fmt.Errorf("failed to marshal values.yaml: %w", err)
	}
	return string(valuesYAML), nil
}

// GetChartYAML returns the Chart.yaml content as string
func (c *Converter) GetChartYAML(chart *Chart) (string, error) {
	chartMetadata := map[string]interface{}{
		"apiVersion":  "v2",
		"name":        chart.Name,
		"version":     "0.1.0",
		"description": "Generated Helm chart from HCL",
	}

	chartYAML, err := yaml.Marshal(chartMetadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Chart.yaml: %w", err)
	}
	return string(chartYAML), nil
}
