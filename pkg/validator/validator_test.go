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
`,
			wantError: true,
		},
		{
			name: "missing metadata name",
			yaml: `
apiVersion: v1
kind: Pod
metadata: {}
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

	validator, err := NewValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

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
	validator, err := NewValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	resources := []string{
		`
apiVersion: v1
kind: Service
metadata:
  name: test-service
`,
		`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
`,
		`
apiVersion: v1
kind: Pod
metadata: {}
`, // Invalid - missing name
	}

	errors := validator.ValidateResources(resources)
	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}
}

func TestValidateSchema(t *testing.T) {
	validator, err := NewValidator()
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	tests := []struct {
		name     string
		resource map[string]interface{}
		wantErr  bool
	}{
		{
			name: "has all required fields",
			resource: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"name": "test-pod",
				},
			},
			wantErr: false,
		},
		{
			name: "missing apiVersion",
			resource: map[string]interface{}{
				"kind": "Pod",
				"metadata": map[string]interface{}{
					"name": "test-pod",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateSchema(tt.resource)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetGVR(t *testing.T) {
	tests := []struct {
		name             string
		apiVersion       string
		kind             string
		expectedGroup    string
		expectedVersion  string
		expectedResource string
		wantErr          bool
	}{
		{
			name:             "deployment",
			apiVersion:       "apps/v1",
			kind:             "Deployment",
			expectedGroup:    "apps",
			expectedVersion:  "v1",
			expectedResource: "deployments",
			wantErr:          false,
		},
		{
			name:             "service",
			apiVersion:       "v1",
			kind:             "Service",
			expectedGroup:    "",
			expectedVersion:  "v1",
			expectedResource: "services",
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, version, resource, err := GetGVR(tt.apiVersion, tt.kind)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if group != tt.expectedGroup || version != tt.expectedVersion || resource != tt.expectedResource {
				t.Errorf("expected (%s, %s, %s), got (%s, %s, %s)",
					tt.expectedGroup, tt.expectedVersion, tt.expectedResource,
					group, version, resource)
			}
		})
	}
}
