package converter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConvert(t *testing.T) {
	// Create a temporary directory with test HCL files
	tempDir, err := os.MkdirTemp("", "converter-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	deploymentContent := `
variable "namespace" {
  type        = "string"
  default     = "default"
  description = "Kubernetes namespace"
}

variable "replicas" {
  type    = "number"
  default = 3
  description = "Number of replicas"
}

resource "kubernetes_deployment" "web" {
  replicas = 3
  
  selector = {
    matchLabels = {
      app = "web"
    }
  }
  
  template = {
    metadata = {
      labels = {
        app = "web"
      }
    }
    spec = {
      containers = [
        {
          name  = "web"
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

	serviceContent := `
resource "kubernetes_service" "web" {
  type = "ClusterIP"
  
  selector = {
    app = "web"
  }
  
  ports = [
    {
      port       = 80
      targetPort = 80
      protocol   = "TCP"
    }
  ]
}
`

	configMapContent := `
variable "app_name" {
  type        = "string"
  default     = "myapp"
  description = "Application name"
}

resource "kubernetes_config_map" "app_config" {
  data = {
    environment = "production"
    log_level   = "info"
  }
}
`

	// Write test files
	deploymentPath := filepath.Join(tempDir, "deployment.hcl")
	if err := os.WriteFile(deploymentPath, []byte(deploymentContent), 0644); err != nil {
		t.Fatalf("Failed to write deployment file: %v", err)
	}

	servicePath := filepath.Join(tempDir, "service.hcl")
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		t.Fatalf("Failed to write service file: %v", err)
	}

	configMapPath := filepath.Join(tempDir, "configmap.hcl")
	if err := os.WriteFile(configMapPath, []byte(configMapContent), 0644); err != nil {
		t.Fatalf("Failed to write configmap file: %v", err)
	}

	// Convert directory
	converter := NewHCLToHelm("test-chart", nil)
	chart, err := converter.Convert(tempDir)
	if err != nil {
		t.Fatalf("Failed to convert directory: %v", err)
	}

	// Verify chart
	if chart.Name != "test-chart" {
		t.Errorf("Expected chart name 'test-chart', got '%s'", chart.Name)
	}

	// Should have 3 templates (deployment, service, configmap)
	if len(chart.Templates) != 3 {
		t.Errorf("Expected 3 templates, got %d", len(chart.Templates))
	}

	// Should have 3 unique variables
	if len(chart.Values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(chart.Values))
	}

	// Check that expected variables are present
	expectedVars := map[string]bool{
		"namespace": false,
		"replicas":  false,
		"app_name":  false,
	}

	for varName := range chart.Values {
		if _, exists := expectedVars[varName]; exists {
			expectedVars[varName] = true
		}
	}

	for varName, found := range expectedVars {
		if !found {
			t.Errorf("Expected variable '%s' not found in chart values", varName)
		}
	}

	// Verify template names
	templateNames := make(map[string]bool)
	for _, template := range chart.Templates {
		templateNames[template.Name] = true
	}

	expectedTemplates := []string{"web-deployment.yaml", "web-service.yaml", "app_config-configmap.yaml"}
	for _, expectedName := range expectedTemplates {
		if !templateNames[expectedName] {
			t.Errorf("Expected template '%s' not found", expectedName)
		}
	}
}

func TestConvertEmpty(t *testing.T) {
	// Create an empty temporary directory
	tempDir, err := os.MkdirTemp("", "converter-test-empty")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Convert empty directory
	converter := NewHCLToHelm("empty-chart", nil)
	chart, err := converter.Convert(tempDir)
	if err != nil {
		t.Fatalf("Failed to convert empty directory: %v", err)
	}

	// Should have empty chart
	if chart.Name != "empty-chart" {
		t.Errorf("Expected chart name 'empty-chart', got '%s'", chart.Name)
	}

	if len(chart.Templates) != 0 {
		t.Errorf("Expected 0 templates in empty chart, got %d", len(chart.Templates))
	}

	if len(chart.Values) != 0 {
		t.Errorf("Expected 0 values in empty chart, got %d", len(chart.Values))
	}
}

func TestConvertWithSubdirectories(t *testing.T) {
	// Create a temporary directory with subdirectories
	tempDir, err := os.MkdirTemp("", "converter-test-subdir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create HCL files in root and subdirectory
	rootFileContent := `
resource "kubernetes_deployment" "root_app" {
  replicas = 1
  
  selector = {
    matchLabels = {
      app = "root-app"
    }
  }
  
  template = {
    metadata = {
      labels = {
        app = "root-app"
      }
    }
    spec = {
      containers = [
        {
          name  = "root-app"
          image = "nginx:latest"
        }
      ]
    }
  }
}
`

	subFileContent := `
resource "kubernetes_service" "sub_service" {
  type = "ClusterIP"
  
  selector = {
    app = "sub-app"
  }
  
  ports = [
    {
      port       = 8080
      targetPort = 8080
    }
  ]
}
`

	// Write files
	rootFilePath := filepath.Join(tempDir, "root.hcl")
	if err := os.WriteFile(rootFilePath, []byte(rootFileContent), 0644); err != nil {
		t.Fatalf("Failed to write root file: %v", err)
	}

	subFilePath := filepath.Join(subDir, "sub.hcl")
	if err := os.WriteFile(subFilePath, []byte(subFileContent), 0644); err != nil {
		t.Fatalf("Failed to write sub file: %v", err)
	}

	// Convert directory (should include subdirectories)
	converter := NewHCLToHelm("multi-dir-chart", nil)
	chart, err := converter.Convert(tempDir)
	if err != nil {
		t.Fatalf("Failed to convert directory with subdirectories: %v", err)
	}

	// Should have 2 templates (from root and subdirectory)
	if len(chart.Templates) != 2 {
		t.Errorf("Expected 2 templates (including subdirectory), got %d", len(chart.Templates))
	}
}
