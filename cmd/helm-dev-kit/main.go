package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/guitcastro/helm-dev-kit/pkg/converter"
	"github.com/guitcastro/helm-dev-kit/pkg/helm"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input-directory> <output-dir> [chart-name]\n", os.Args[0])
		os.Exit(1)
	}

	inputDir := os.Args[1]
	outputDir := os.Args[2]
	chartName := "mychart"
	if len(os.Args) >= 4 {
		chartName = os.Args[3]
	}

	// Verify input is a directory
	info, err := os.Stat(inputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error accessing input directory: %v\n", err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: input path '%s' is not a directory\n", inputDir)
		fmt.Fprintf(os.Stderr, "This tool requires a directory containing HCL files as input\n")
		os.Exit(1)
	}

	// Create converter without validator for now (offline mode)
	conv := converter.NewHCLToHelm(chartName, nil)

	// Convert directory
	chart, err := conv.Convert(inputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting HCL directory to Helm: %v\n", err)
		os.Exit(1)
	}

	// Write chart to disk
	if err := writeChart(chart, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing chart: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully processed directory '%s'\n", inputDir)
	fmt.Printf("Successfully generated Helm chart '%s' in %s\n", chartName, outputDir)
}

func writeChart(chart *helm.Chart, outputDir string) error {
	chartDir := filepath.Join(outputDir, chart.Name)
	templatesDir := filepath.Join(chartDir, "templates")

	// Create directories
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Write Chart.yaml
	converter := helm.NewConverter(chart.Name)
	chartYAML, err := converter.GetChartYAML(chart)
	if err != nil {
		return fmt.Errorf("failed to get Chart.yaml: %w", err)
	}
	chartYAMLPath := filepath.Join(chartDir, "Chart.yaml")
	if err := os.WriteFile(chartYAMLPath, []byte(chartYAML), 0644); err != nil {
		return fmt.Errorf("failed to write Chart.yaml: %w", err)
	}

	// Write values.yaml
	valuesYAML, err := converter.GetValuesYAML(chart)
	if err != nil {
		return fmt.Errorf("failed to get values.yaml: %w", err)
	}
	valuesYAMLPath := filepath.Join(chartDir, "values.yaml")
	if err := os.WriteFile(valuesYAMLPath, []byte(valuesYAML), 0644); err != nil {
		return fmt.Errorf("failed to write values.yaml: %w", err)
	}

	// Write templates
	for _, template := range chart.Templates {
		templatePath := filepath.Join(templatesDir, template.Name)
		if err := os.WriteFile(templatePath, []byte(template.Content), 0644); err != nil {
			return fmt.Errorf("failed to write template %s: %w", template.Name, err)
		}
	}

	return nil
}
