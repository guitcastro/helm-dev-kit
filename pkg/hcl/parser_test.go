package hcl

import (
	"testing"
)

func TestParseBytes(t *testing.T) {
	tests := []struct {
		name          string
		hcl           string
		wantResources int
		wantVariables int
		wantError     bool
	}{
		{
			name: "simple deployment",
			hcl: `
variable "replicas" {
  type    = "number"
  default = 3
}

resource "kubernetes_deployment" "web" {
  replicas = 3
}
`,
			wantResources: 1,
			wantVariables: 1,
			wantError:     false,
		},
		{
			name: "multiple resources",
			hcl: `
resource "kubernetes_deployment" "web" {
  replicas = 3
}

resource "kubernetes_service" "web" {
  type = "ClusterIP"
}
`,
			wantResources: 2,
			wantVariables: 0,
			wantError:     false,
		},
		{
			name: "invalid HCL",
			hcl: `
resource "kubernetes_deployment" {
  replicas = 3
`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			resources, variables, err := parser.ParseBytes([]byte(tt.hcl), "test.hcl")

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(resources) != tt.wantResources {
				t.Errorf("got %d resources, want %d", len(resources), tt.wantResources)
			}

			if len(variables) != tt.wantVariables {
				t.Errorf("got %d variables, want %d", len(variables), tt.wantVariables)
			}
		})
	}
}

func TestParseResource(t *testing.T) {
	hcl := `
resource "kubernetes_deployment" "web" {
  replicas = 3
  name     = "web-deployment"
}
`

	parser := NewParser()
	resources, _, err := parser.ParseBytes([]byte(hcl), "test.hcl")
	if err != nil {
		t.Fatalf("failed to parse HCL: %v", err)
	}

	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}

	resource := resources[0]
	if resource.Type != "kubernetes_deployment" {
		t.Errorf("got type %s, want kubernetes_deployment", resource.Type)
	}

	if resource.Name != "web" {
		t.Errorf("got name %s, want web", resource.Name)
	}

	if resource.Attributes["replicas"] != float64(3) {
		t.Errorf("got replicas %v, want 3", resource.Attributes["replicas"])
	}
}

func TestParseVariable(t *testing.T) {
	hcl := `
variable "replicas" {
  type        = "number"
  default     = 3
  description = "Number of replicas"
}
`

	parser := NewParser()
	_, variables, err := parser.ParseBytes([]byte(hcl), "test.hcl")
	if err != nil {
		t.Fatalf("failed to parse HCL: %v", err)
	}

	if len(variables) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(variables))
	}

	variable := variables[0]
	if variable.Name != "replicas" {
		t.Errorf("got name %s, want replicas", variable.Name)
	}

	if variable.Type != "number" {
		t.Errorf("got type %s, want number", variable.Type)
	}

	if variable.Default != float64(3) {
		t.Errorf("got default %v, want 3", variable.Default)
	}

	if variable.Description != "Number of replicas" {
		t.Errorf("got description %s, want 'Number of replicas'", variable.Description)
	}
}
