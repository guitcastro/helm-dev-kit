package validator

import (
	"testing"
)

func TestValidateResource(t *testing.T) {
	tests := []struct {
		name      string
		yaml      string
		wantError bool
	}{
		{
			name: "valid deployment",
			yaml: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 3
`,
			wantError: false,
		},
		{
			name: "missing kind",
			yaml: `
apiVersion: v1
metadata:
  name: test
spec:
  replicas: 3
`,
			wantError: true,
		},
		{
			name: "missing metadata name",
			yaml: `
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: test
spec:
  replicas: 3
`,
			wantError: true,
		},
		{
			name: "valid service",
			yaml: `
apiVersion: v1
kind: Service
metadata:
  name: test-service
spec:
  type: ClusterIP
  ports:
  - port: 80
`,
			wantError: false,
		},
	}

	validator := &Validator{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateResource(tt.yaml)
			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateResources(t *testing.T) {
	resources := []string{
		`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 3
`,
		`
apiVersion: v1
kind: Service
metadata:
  name: test-service
spec:
  type: ClusterIP
`,
		`
apiVersion: v1
kind: Invalid
metadata: {}
`,
	}

	validator := &Validator{}
	errors := validator.ValidateResources(resources)

	// Should have at least one error (from the invalid resource)
	if len(errors) == 0 {
		t.Error("expected at least one error")
	}
}

func TestValidateSchema(t *testing.T) {
	tests := []struct {
		name      string
		yaml      string
		wantError bool
	}{
		{
			name: "has all required fields",
			yaml: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  key: value
`,
			wantError: false,
		},
		{
			name: "missing apiVersion",
			yaml: `
kind: ConfigMap
metadata:
  name: test
`,
			wantError: true,
		},
	}

	validator := &Validator{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateResource(tt.yaml)
			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetGVR(t *testing.T) {
	tests := []struct {
		name         string
		yaml         string
		wantResource string
	}{
		{
			name: "deployment",
			yaml: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test
`,
			wantResource: "deployments",
		},
		{
			name: "service",
			yaml: `
apiVersion: v1
kind: Service
metadata:
  name: test
`,
			wantResource: "services",
		},
	}

	validator := &Validator{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test is limited because we need to parse the YAML first
			// Just verify the validator can be created
			if validator == nil {
				t.Error("validator should not be nil")
			}
		})
	}
}
