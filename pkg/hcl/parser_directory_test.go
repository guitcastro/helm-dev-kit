package hcl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDirectory(t *testing.T) {
	// Create a temporary directory with test HCL files
	tempDir, err := os.MkdirTemp("", "hcl-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	file1Content := `
variable "namespace" {
  type        = "string"
  default     = "default"
  description = "Kubernetes namespace"
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
        }
      ]
    }
  }
}
`

	file2Content := `
variable "replicas" {
  type    = "number"
  default = 2
  description = "Number of replicas"
}

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

	// Write test files
	file1Path := filepath.Join(tempDir, "deployment.hcl")
	if err := os.WriteFile(file1Path, []byte(file1Content), 0644); err != nil {
		t.Fatalf("Failed to write test file 1: %v", err)
	}

	file2Path := filepath.Join(tempDir, "service.hcl")
	if err := os.WriteFile(file2Path, []byte(file2Content), 0644); err != nil {
		t.Fatalf("Failed to write test file 2: %v", err)
	}

	// Create a non-HCL file to test filtering
	nonHCLPath := filepath.Join(tempDir, "readme.txt")
	if err := os.WriteFile(nonHCLPath, []byte("This is not an HCL file"), 0644); err != nil {
		t.Fatalf("Failed to write non-HCL file: %v", err)
	}

	// Parse directory
	parser := NewParser()
	resources, variables, err := parser.ParseDirectory(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse directory: %v", err)
	}

	// Verify resources
	if len(resources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(resources))
	}

	resourceTypes := make(map[string]bool)
	for _, resource := range resources {
		resourceTypes[resource.Type] = true
	}

	if !resourceTypes["kubernetes_deployment"] {
		t.Error("Expected kubernetes_deployment resource")
	}
	if !resourceTypes["kubernetes_service"] {
		t.Error("Expected kubernetes_service resource")
	}

	// Verify variables (should be merged)
	if len(variables) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(variables))
	}

	variableNames := make(map[string]bool)
	for _, variable := range variables {
		variableNames[variable.Name] = true
	}

	if !variableNames["namespace"] {
		t.Error("Expected namespace variable")
	}
	if !variableNames["replicas"] {
		t.Error("Expected replicas variable")
	}
}

func TestParseDirectoryWithDuplicateVariables(t *testing.T) {
	// Create a temporary directory with test HCL files
	tempDir, err := os.MkdirTemp("", "hcl-test-duplicate")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files with duplicate variable names
	file1Content := `
variable "replicas" {
  type        = "number"
  default     = 3
  description = "First definition"
}
`

	file2Content := `
variable "replicas" {
  type        = "number"
  default     = 5
  description = "Second definition"
}
`

	// Write test files
	file1Path := filepath.Join(tempDir, "file1.hcl")
	if err := os.WriteFile(file1Path, []byte(file1Content), 0644); err != nil {
		t.Fatalf("Failed to write test file 1: %v", err)
	}

	file2Path := filepath.Join(tempDir, "file2.hcl")
	if err := os.WriteFile(file2Path, []byte(file2Content), 0644); err != nil {
		t.Fatalf("Failed to write test file 2: %v", err)
	}

	// Parse directory
	parser := NewParser()
	_, variables, err := parser.ParseDirectory(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse directory: %v", err)
	}

	// Should have only one variable (merged)
	if len(variables) != 1 {
		t.Errorf("Expected 1 variable after merge, got %d", len(variables))
	}

	if variables[0].Name != "replicas" {
		t.Errorf("Expected variable name 'replicas', got '%s'", variables[0].Name)
	}
}

func TestParseEmptyDirectory(t *testing.T) {
	// Create an empty temporary directory
	tempDir, err := os.MkdirTemp("", "hcl-test-empty")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Parse empty directory
	parser := NewParser()
	resources, variables, err := parser.ParseDirectory(tempDir)
	if err != nil {
		t.Fatalf("Failed to parse empty directory: %v", err)
	}

	// Should have no resources or variables
	if len(resources) != 0 {
		t.Errorf("Expected 0 resources in empty directory, got %d", len(resources))
	}

	if len(variables) != 0 {
		t.Errorf("Expected 0 variables in empty directory, got %d", len(variables))
	}
}
