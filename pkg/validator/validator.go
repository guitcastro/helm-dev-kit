package validator

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Validator validates Kubernetes resources against basic schemas
type Validator struct {
	// Simple validator without k8s client dependencies
}

// NewValidator creates a new validator instance
func NewValidator() (*Validator, error) {
	return &Validator{}, nil
}

// ValidateResource validates a single Kubernetes resource
func (v *Validator) ValidateResource(resourceYAML string) error {
	var resource map[string]interface{}
	if err := yaml.Unmarshal([]byte(resourceYAML), &resource); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}

	// Basic validation checks
	if err := v.validateSchema(resource); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	return nil
}

// ValidateResources validates multiple Kubernetes resources
func (v *Validator) ValidateResources(resources []string) []error {
	var errors []error
	for i, resource := range resources {
		if err := v.ValidateResource(resource); err != nil {
			errors = append(errors, fmt.Errorf("resource %d: %w", i, err))
		}
	}
	return errors
}

// validateSchema performs basic schema validation
func (v *Validator) validateSchema(resource map[string]interface{}) error {
	// Check for required fields
	if _, ok := resource["apiVersion"]; !ok {
		return fmt.Errorf("missing required field: apiVersion")
	}

	if _, ok := resource["kind"]; !ok {
		return fmt.Errorf("missing required field: kind")
	}

	metadata, ok := resource["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing or invalid metadata field")
	}

	if _, ok := metadata["name"]; !ok {
		return fmt.Errorf("missing required field: metadata.name")
	}

	return nil
}

// GetGVR returns the GroupVersionResource for a given resource
// Simplified version without client-go dependencies
func GetGVR(apiVersion, kind string) (string, string, string, error) {
	// Basic mapping for common resources
	resourceMap := map[string]map[string][]string{
		"apps/v1": {
			"Deployment":  {"apps", "v1", "deployments"},
			"ReplicaSet":  {"apps", "v1", "replicasets"},
			"StatefulSet": {"apps", "v1", "statefulsets"},
		},
		"v1": {
			"Pod":       {"", "v1", "pods"},
			"Service":   {"", "v1", "services"},
			"ConfigMap": {"", "v1", "configmaps"},
			"Secret":    {"", "v1", "secrets"},
			"Namespace": {"", "v1", "namespaces"},
		},
		"extensions/v1beta1": {
			"Ingress": {"extensions", "v1beta1", "ingresses"},
		},
		"networking.k8s.io/v1": {
			"Ingress": {"networking.k8s.io", "v1", "ingresses"},
		},
		"batch/v1": {
			"Job": {"batch", "v1", "jobs"},
		},
		"batch/v1beta1": {
			"CronJob": {"batch", "v1beta1", "cronjobs"},
		},
	}

	if kinds, ok := resourceMap[apiVersion]; ok {
		if gvr, ok := kinds[kind]; ok {
			return gvr[0], gvr[1], gvr[2], nil
		}
	}

	return "", "", "", fmt.Errorf("unknown resource type: %s/%s", apiVersion, kind)
}

// ValidateChart validates all resources in a Helm chart
func (v *Validator) ValidateChart(templates []string) []error {
	return v.ValidateResources(templates)
}
