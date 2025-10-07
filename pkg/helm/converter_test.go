package helm

import (
	"strings"
	"testing"

	"github.com/guitcastro/helm-dev-kit/pkg/hcl"
)

func TestConvert(t *testing.T) {
	variables := []hcl.Variable{
		{
			Name:    "replicas",
			Type:    "number",
			Default: float64(3),
		},
		{
			Name:    "namespace",
			Type:    "string",
			Default: "default",
		},
	}

	resources := []hcl.Resource{
		{
			Type: "kubernetes_deployment",
			Name: "web",
			Attributes: map[string]interface{}{
				"replicas": float64(3),
			},
		},
	}

	converter := NewConverter("test-chart")
	chart, err := converter.Convert(resources, variables)
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	if chart.Name != "test-chart" {
		t.Errorf("got chart name %s, want test-chart", chart.Name)
	}

	if len(chart.Templates) != 1 {
		t.Errorf("got %d templates, want 1", len(chart.Templates))
	}

	if len(chart.Values) != 2 {
		t.Errorf("got %d values, want 2", len(chart.Values))
	}

	if chart.Values["replicas"] != float64(3) {
		t.Errorf("got replicas value %v, want 3", chart.Values["replicas"])
	}
}

func TestResourceToTemplate(t *testing.T) {
	resource := hcl.Resource{
		Type: "kubernetes_deployment",
		Name: "web",
		Attributes: map[string]interface{}{
			"replicas": float64(3),
		},
	}

	converter := NewConverter("test-chart")
	template, err := converter.resourceToTemplate(resource)
	if err != nil {
		t.Fatalf("failed to convert resource: %v", err)
	}

	if template.Name != "web-deployment.yaml" {
		t.Errorf("got template name %s, want web-deployment.yaml", template.Name)
	}

	if !strings.Contains(template.Content, "kind: Deployment") {
		t.Error("template should contain 'kind: Deployment'")
	}

	if !strings.Contains(template.Content, "apiVersion: apps/v1") {
		t.Error("template should contain 'apiVersion: apps/v1'")
	}
}

func TestParseKind(t *testing.T) {
	tests := []struct {
		resourceType string
		wantKind     string
	}{
		{"kubernetes_deployment", "Deployment"},
		{"kubernetes_service", "Service"},
		{"kubernetes_config_map", "ConfigMap"},
		{"stateful_set", "StatefulSet"},
	}

	converter := NewConverter("test")
	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			got := converter.parseKind(tt.resourceType)
			if got != tt.wantKind {
				t.Errorf("parseKind(%s) = %s, want %s", tt.resourceType, got, tt.wantKind)
			}
		})
	}
}

func TestGetAPIVersion(t *testing.T) {
	tests := []struct {
		kind       string
		wantAPIVer string
	}{
		{"Deployment", "apps/v1"},
		{"Service", "v1"},
		{"ConfigMap", "v1"},
		{"Ingress", "networking.k8s.io/v1"},
		{"Job", "batch/v1"},
	}

	converter := NewConverter("test")
	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			got := converter.getAPIVersion(tt.kind)
			if got != tt.wantAPIVer {
				t.Errorf("getAPIVersion(%s) = %s, want %s", tt.kind, got, tt.wantAPIVer)
			}
		})
	}
}

func TestProcessValue(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{
			name:  "simple string",
			input: "test",
			want:  "test",
		},
		{
			name:  "variable reference",
			input: "var.replicas",
			want:  "{{ .Values.replicas }}",
		},
		{
			name: "nested map",
			input: map[string]interface{}{
				"key": "var.value",
			},
			want: map[string]interface{}{
				"key": "{{ .Values.value }}",
			},
		},
	}

	converter := NewConverter("test")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := converter.processValue(tt.input, "test")
			if !compareValues(got, tt.want) {
				t.Errorf("processValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func compareValues(a, b interface{}) bool {
	switch va := a.(type) {
	case string:
		vb, ok := b.(string)
		return ok && va == vb
	case map[string]interface{}:
		vb, ok := b.(map[string]interface{})
		if !ok || len(va) != len(vb) {
			return false
		}
		for k, v := range va {
			if !compareValues(v, vb[k]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}

func TestGetValuesYAML(t *testing.T) {
	chart := &Chart{
		Name: "test-chart",
		Values: Values{
			"replicas":  float64(3),
			"namespace": "default",
		},
	}

	converter := NewConverter("test-chart")
	yaml, err := converter.GetValuesYAML(chart)
	if err != nil {
		t.Fatalf("failed to get values YAML: %v", err)
	}

	if !strings.Contains(yaml, "replicas:") {
		t.Error("YAML should contain 'replicas:'")
	}

	if !strings.Contains(yaml, "namespace:") {
		t.Error("YAML should contain 'namespace:'")
	}
}

func TestGetChartYAML(t *testing.T) {
	chart := &Chart{
		Name: "test-chart",
	}

	converter := NewConverter("test-chart")
	yaml, err := converter.GetChartYAML(chart)
	if err != nil {
		t.Fatalf("failed to get chart YAML: %v", err)
	}

	if !strings.Contains(yaml, "name: test-chart") {
		t.Error("YAML should contain 'name: test-chart'")
	}

	if !strings.Contains(yaml, "apiVersion: v2") {
		t.Error("YAML should contain 'apiVersion: v2'")
	}
}
