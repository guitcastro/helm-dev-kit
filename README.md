# helm-dev-kit

A Go framework that converts Terraform-style HCL configurations into validated Helm charts. It parses HCL resources and variables, maps variables to `values.yaml`, validates resources with Kubernetes OpenAPI schemas, and uses `sigs.k8s.io/e2e-framework` for testing.

## Features

- **HCL Parser**: Parse Terraform-style HCL resources and variables
- **Helm Converter**: Convert HCL resources to Helm templates
- **Variable Mapping**: Automatically map HCL variables to Helm `values.yaml`
- **Kubernetes Validation**: Validate resources against Kubernetes OpenAPI schemas
- **E2E Testing**: Integration with `sigs.k8s.io/e2e-framework` for testing
- **CLI Tool**: Command-line interface for converting HCL to Helm charts

## Installation

```bash
go get github.com/guitcastro/helm-dev-kit
```

## Usage

### As a Library

```go
package main

import (
    "fmt"
    "github.com/guitcastro/helm-dev-kit/pkg/converter"
)

func main() {
    // Create converter (without validator for offline mode)
    conv := converter.NewHCLToHelm("my-chart", nil)
    
    // Convert HCL file to Helm chart
    chart, err := conv.ConvertFile("deployment.hcl")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Generated chart: %s\n", chart.Name)
    fmt.Printf("Templates: %d\n", len(chart.Templates))
    fmt.Printf("Values: %d\n", len(chart.Values))
}
```

### As a CLI Tool

```bash
# Build the CLI
go build -o helm-dev-kit ./cmd/helm-dev-kit

# Convert HCL to Helm chart
./helm-dev-kit input.hcl output-dir my-chart
```

## HCL Input Format

Define your Kubernetes resources in HCL format:

```hcl
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
  replicas = var.replicas
  
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
```

## Output

The framework generates a complete Helm chart structure:

```
my-chart/
├── Chart.yaml          # Chart metadata
├── values.yaml         # Values from HCL variables
└── templates/
    ├── web.yaml        # Deployment template
    └── web-service.yaml # Service template
```

### Example Chart.yaml

```yaml
apiVersion: v2
name: my-chart
version: 0.1.0
description: Generated Helm chart from HCL
```

### Example values.yaml

```yaml
namespace: default
replicas: 3
```

### Example Template

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Values.web.name | default "web" }}
  namespace: {{ .Values.namespace | default "default" }}
spec:
  replicas: {{ .Values.replicas }}
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: web
        image: nginx:latest
        ports:
        - containerPort: 80
```

## Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/hcl
go test ./pkg/helm
go test ./pkg/validator
go test ./tests
```

## Architecture

### Package Structure

```
helm-dev-kit/
├── cmd/
│   └── helm-dev-kit/      # CLI application
├── pkg/
│   ├── hcl/               # HCL parser
│   ├── helm/              # Helm converter
│   ├── validator/         # Kubernetes validator
│   └── converter/         # Main converter logic
├── examples/              # Example HCL files
└── tests/                 # Integration tests
```

### Components

1. **HCL Parser** (`pkg/hcl`): Parses Terraform-style HCL files and extracts resources and variables
2. **Helm Converter** (`pkg/helm`): Converts HCL resources to Helm templates and maps variables to values
3. **Validator** (`pkg/validator`): Validates Kubernetes resources against OpenAPI schemas
4. **Converter** (`pkg/converter`): Orchestrates the conversion process with validation

## Supported Resource Types

- `kubernetes_deployment` → Deployment (apps/v1)
- `kubernetes_service` → Service (v1)
- `kubernetes_config_map` → ConfigMap (v1)
- `kubernetes_secret` → Secret (v1)
- `kubernetes_pod` → Pod (v1)
- `kubernetes_stateful_set` → StatefulSet (apps/v1)
- `kubernetes_daemon_set` → DaemonSet (apps/v1)
- `kubernetes_ingress` → Ingress (networking.k8s.io/v1)
- `kubernetes_job` → Job (batch/v1)
- `kubernetes_cron_job` → CronJob (batch/v1)

## Variable Mapping

HCL variables are automatically mapped to Helm values:

| HCL Variable | Helm Value Path |
|--------------|-----------------|
| `variable "replicas"` | `.Values.replicas` |
| `variable "namespace"` | `.Values.namespace` |
| `variable "image_tag"` | `.Values.image_tag` |

Variable references in HCL (`var.replicas`) are converted to Helm template syntax (`{{ .Values.replicas }}`).

## Validation

The framework validates:

- HCL syntax and structure
- Kubernetes resource schema (required fields)
- Resource references and dependencies
- API versions and kinds

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details.

## Dependencies

- `github.com/hashicorp/hcl/v2` - HCL parsing
- `helm.sh/helm/v3` - Helm chart handling
- `k8s.io/client-go` - Kubernetes client
- `k8s.io/apimachinery` - Kubernetes API machinery
- `sigs.k8s.io/e2e-framework` - End-to-end testing framework
- `gopkg.in/yaml.v3` - YAML processing