package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/guitcastro/helm-dev-kit/pkg/converter"
	"github.com/guitcastro/helm-dev-kit/pkg/hcl"
	"github.com/guitcastro/helm-dev-kit/pkg/helm"
)

func TestHCLToHelmConversion(t *testing.T) {
	hclContent := `
variable "replicas" {
  type    = "number"
  default = 3
}

variable "namespace" {
  type    = "string"
  default = "default"
}

resource "kubernetes_deployment" "web" {
  replicas = 3
  selector = {
    matchLabels = {
      app = "web"
    }
  }
}

resource "kubernetes_service" "web" {
  type = "ClusterIP"
  selector = {
    app = "web"
  }
}
`

	// Create temporary directory with HCL file
	tempDir, err := os.MkdirTemp("", "integration-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.hcl")
	if err := os.WriteFile(testFile, []byte(hclContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create converter without validator for offline testing
	conv := converter.NewHCLToHelm("test-chart", nil)

	// Convert HCL directory to Helm chart
	chart, err := conv.Convert(tempDir)
	if err != nil {
		t.Fatalf("failed to convert HCL to Helm: %v", err)
	}

	// Verify chart structure
	if chart.Name != "test-chart" {
		t.Errorf("got chart name %s, want test-chart", chart.Name)
	}

	if len(chart.Templates) != 2 {
		t.Errorf("got %d templates, want 2", len(chart.Templates))
	}

	if len(chart.Values) != 2 {
		t.Errorf("got %d values, want 2", len(chart.Values))
	}

	// Verify values
	if chart.Values["replicas"] != float64(3) {
		t.Errorf("got replicas %v, want 3", chart.Values["replicas"])
	}

	if chart.Values["namespace"] != "default" {
		t.Errorf("got namespace %v, want default", chart.Values["namespace"])
	}
}

func TestEndToEndWorkflow(t *testing.T) {
	ctx := context.Background()
	_ = ctx

	// Create temporary directory with HCL file
	tempDir, err := os.MkdirTemp("", "integration-test-e2e")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	hclContent := `
variable "image_tag" {
  type    = "string"
  default = "latest"
}

resource "kubernetes_deployment" "app" {
  replicas = 2
}
`
	testFile := filepath.Join(tempDir, "test.hcl")
	if err := os.WriteFile(testFile, []byte(hclContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Step 1: Parse HCL directory
	parser := hcl.NewParser()
	resources, variables, err := parser.ParseDirectory(tempDir)
	if err != nil {
		t.Fatalf("failed to parse HCL directory: %v", err)
	}

	if len(resources) != 1 {
		t.Errorf("got %d resources, want 1", len(resources))
	}

	if len(variables) != 1 {
		t.Errorf("got %d variables, want 1", len(variables))
	}

	// Step 2: Convert to Helm
	converter := helm.NewConverter("app-chart")
	chart, err := converter.Convert(resources, variables)
	if err != nil {
		t.Fatalf("failed to convert to Helm: %v", err)
	}

	if len(chart.Templates) != 1 {
		t.Errorf("got %d templates, want 1", len(chart.Templates))
	}

	// Step 3: Verify outputs
	valuesYAML, err := converter.GetValuesYAML(chart)
	if err != nil {
		t.Fatalf("failed to get values.yaml: %v", err)
	}

	if valuesYAML == "" {
		t.Error("values.yaml should not be empty")
	}

	chartYAML, err := converter.GetChartYAML(chart)
	if err != nil {
		t.Fatalf("failed to get Chart.yaml: %v", err)
	}

	if chartYAML == "" {
		t.Error("Chart.yaml should not be empty")
	}
}

func TestComplexResource(t *testing.T) {
	hclContent := `
variable "app_name" {
  type    = "string"
  default = "myapp"
}

variable "replicas" {
  type    = "number"
  default = 3
}

resource "kubernetes_deployment" "app" {
  replicas = 3
  selector = {
    matchLabels = {
      app = "myapp"
    }
  }
  template = {
    metadata = {
      labels = {
        app = "myapp"
      }
    }
    spec = {
      containers = [
        {
          name  = "app"
          image = "nginx:latest"
          ports = [
            {
              containerPort = 80
            }
          ]
        }
      ]
    }
  }
}
`

	// Create temporary directory with HCL file
	tempDir, err := os.MkdirTemp("", "integration-test-complex")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.hcl")
	if err := os.WriteFile(testFile, []byte(hclContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	conv := converter.NewHCLToHelm("complex-chart", nil)
	chart, err := conv.Convert(tempDir)
	if err != nil {
		t.Fatalf("failed to convert complex HCL: %v", err)
	}

	if len(chart.Templates) != 1 {
		t.Errorf("got %d templates, want 1", len(chart.Templates))
	}

	if len(chart.Values) != 2 {
		t.Errorf("got %d values, want 2", len(chart.Values))
	}
}
