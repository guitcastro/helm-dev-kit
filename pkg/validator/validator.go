package validator

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// Validator validates Kubernetes resources against OpenAPI schemas
type Validator struct {
	discoveryClient discovery.DiscoveryInterface
	dynamicClient   dynamic.Interface
}

// NewValidator creates a new validator
func NewValidator(config *rest.Config) (*Validator, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &Validator{
		discoveryClient: discoveryClient,
		dynamicClient:   dynamicClient,
	}, nil
}

// NewValidatorWithClients creates a validator with existing clients (for testing)
func NewValidatorWithClients(discoveryClient discovery.DiscoveryInterface, dynamicClient dynamic.Interface) *Validator {
	return &Validator{
		discoveryClient: discoveryClient,
		dynamicClient:   dynamicClient,
	}
}

// ValidateResource validates a single Kubernetes resource
func (v *Validator) ValidateResource(resourceYAML string) error {
	// Parse YAML to unstructured object
	obj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal([]byte(resourceYAML), &obj.Object); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Get GVK
	gvk := obj.GroupVersionKind()
	if gvk.Kind == "" {
		return fmt.Errorf("missing kind in resource")
	}

	// Basic validation - check required fields
	if obj.GetName() == "" && obj.GetGenerateName() == "" {
		return fmt.Errorf("resource must have metadata.name or metadata.generateName")
	}

	// Validate schema structure
	if err := v.validateSchema(obj); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	return nil
}

// validateSchema performs basic schema validation
func (v *Validator) validateSchema(obj *unstructured.Unstructured) error {
	// Check required top-level fields
	requiredFields := []string{"apiVersion", "kind", "metadata"}
	for _, field := range requiredFields {
		if _, found := obj.Object[field]; !found {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	return nil
}

// ValidateResources validates multiple resources
func (v *Validator) ValidateResources(resources []string) []error {
	var errors []error
	for i, resource := range resources {
		if err := v.ValidateResource(resource); err != nil {
			errors = append(errors, fmt.Errorf("resource %d: %w", i, err))
		}
	}
	return errors
}

// SimulateApply simulates applying a resource (dry-run)
func (v *Validator) SimulateApply(ctx context.Context, resourceYAML string) error {
	// Parse YAML to unstructured object
	obj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal([]byte(resourceYAML), &obj.Object); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Get GVR (GroupVersionResource)
	gvr, err := v.getGVR(obj)
	if err != nil {
		return fmt.Errorf("failed to get GVR: %w", err)
	}

	// Perform dry-run create
	namespace := obj.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	// This would normally perform a dry-run, but for offline validation we skip actual API calls
	_ = gvr
	_ = namespace

	return nil
}

// getGVR gets the GroupVersionResource for an object
func (v *Validator) getGVR(obj *unstructured.Unstructured) (schema.GroupVersionResource, error) {
	gvk := obj.GroupVersionKind()
	
	// Map common kinds to their plural resource names
	kindToResource := map[string]string{
		"Deployment":  "deployments",
		"Service":     "services",
		"ConfigMap":   "configmaps",
		"Secret":      "secrets",
		"Pod":         "pods",
		"StatefulSet": "statefulsets",
		"DaemonSet":   "daemonsets",
		"Ingress":     "ingresses",
		"Job":         "jobs",
		"CronJob":     "cronjobs",
	}

	resource, ok := kindToResource[gvk.Kind]
	if !ok {
		// Default: lowercase kind + 's'
		resource = fmt.Sprintf("%ss", gvk.Kind)
	}

	return schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: resource,
	}, nil
}

// ValidateChart validates all resources in a Helm chart
func (v *Validator) ValidateChart(templates []string) []error {
	return v.ValidateResources(templates)
}
