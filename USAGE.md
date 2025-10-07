# Usage Guide

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/guitcastro/helm-dev-kit.git
cd helm-dev-kit

# Build the CLI
make build

# Or install to GOPATH/bin
make install
```

### Basic Usage

Convert an HCL file to a Helm chart:

```bash
./helm-dev-kit input.hcl output-dir chart-name
```

Example:

```bash
./helm-dev-kit examples/deployment.hcl ./output my-web-app
```

This will create the following structure:

```
output/
└── my-web-app/
    ├── Chart.yaml
    ├── values.yaml
    └── templates/
        ├── web-deployment.yaml
        └── web-service.yaml
```

## HCL Syntax

### Defining Variables

Variables in HCL define configurable values that will be mapped to Helm's `values.yaml`:

```hcl
variable "namespace" {
  type        = "string"
  default     = "default"
  description = "Kubernetes namespace"
}

variable "replicas" {
  type        = "number"
  default     = 3
  description = "Number of pod replicas"
}
```

Supported variable types:
- `string`: Text values
- `number`: Numeric values
- `bool`: Boolean values (true/false)

### Defining Resources

Resources in HCL define Kubernetes objects:

```hcl
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
```

### Supported Resource Types

| HCL Resource Type | Kubernetes Kind | API Version |
|-------------------|-----------------|-------------|
| `kubernetes_deployment` | Deployment | apps/v1 |
| `kubernetes_service` | Service | v1 |
| `kubernetes_config_map` | ConfigMap | v1 |
| `kubernetes_secret` | Secret | v1 |
| `kubernetes_pod` | Pod | v1 |
| `kubernetes_stateful_set` | StatefulSet | apps/v1 |
| `kubernetes_daemon_set` | DaemonSet | apps/v1 |
| `kubernetes_ingress` | Ingress | networking.k8s.io/v1 |
| `kubernetes_job` | Job | batch/v1 |
| `kubernetes_cron_job` | CronJob | batch/v1 |

## Examples

### Example 1: Simple Web Application

**Input (deployment.hcl):**

```hcl
variable "namespace" {
  type    = "string"
  default = "default"
}

variable "replicas" {
  type    = "number"
  default = 3
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

**Command:**

```bash
./helm-dev-kit deployment.hcl ./output web-app
```

**Output Chart Structure:**

```
output/web-app/
├── Chart.yaml
├── values.yaml
└── templates/
    ├── web-deployment.yaml
    └── web-service.yaml
```

**Generated values.yaml:**

```yaml
namespace: default
replicas: 3
```

**Generated web-deployment.yaml:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: '{{ .Values.web.name | default "web" }}'
  namespace: '{{ .Values.namespace | default "default" }}'
spec:
  replicas: 3
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

### Example 2: Application with ConfigMap

**Input (configmap.hcl):**

```hcl
variable "app_name" {
  type    = "string"
  default = "myapp"
}

resource "kubernetes_config_map" "app_config" {
  data = {
    environment = "production"
    log_level   = "info"
  }
}

resource "kubernetes_deployment" "app" {
  replicas = 2
  
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
          image = "myapp:v1.0.0"
          ports = [
            {
              containerPort = 8080
            }
          ]
        }
      ]
    }
  }
}
```

**Command:**

```bash
./helm-dev-kit configmap.hcl ./output myapp
```

## Library Usage

### Basic Conversion

```go
package main

import (
    "fmt"
    "github.com/guitcastro/helm-dev-kit/pkg/converter"
)

func main() {
    // Create converter
    conv := converter.NewHCLToHelm("my-chart", nil)
    
    // Convert HCL file
    chart, err := conv.ConvertFile("deployment.hcl")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Chart: %s\n", chart.Name)
    fmt.Printf("Templates: %d\n", len(chart.Templates))
}
```

### With Validation

```go
package main

import (
    "github.com/guitcastro/helm-dev-kit/pkg/converter"
    "github.com/guitcastro/helm-dev-kit/pkg/validator"
    "k8s.io/client-go/rest"
)

func main() {
    // Create Kubernetes config
    config, err := rest.InClusterConfig()
    if err != nil {
        panic(err)
    }
    
    // Create validator
    val, err := validator.NewValidator(config)
    if err != nil {
        panic(err)
    }
    
    // Create converter with validator
    conv := converter.NewHCLToHelm("my-chart", val)
    
    // Convert and validate
    chart, err := conv.ConvertFile("deployment.hcl")
    if err != nil {
        panic(err)
    }
    
    // Chart is validated
    fmt.Println("Chart validated successfully!")
}
```

### Parsing HCL Directly

```go
package main

import (
    "github.com/guitcastro/helm-dev-kit/pkg/hcl"
)

func main() {
    parser := hcl.NewParser()
    
    resources, variables, err := parser.ParseFile("deployment.hcl")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d resources and %d variables\n", 
        len(resources), len(variables))
}
```

### Converting to Helm

```go
package main

import (
    "github.com/guitcastro/helm-dev-kit/pkg/helm"
    "github.com/guitcastro/helm-dev-kit/pkg/hcl"
)

func main() {
    // Parse HCL
    parser := hcl.NewParser()
    resources, variables, _ := parser.ParseFile("deployment.hcl")
    
    // Convert to Helm
    converter := helm.NewConverter("my-chart")
    chart, err := converter.Convert(resources, variables)
    if err != nil {
        panic(err)
    }
    
    // Get YAML outputs
    chartYAML, _ := converter.GetChartYAML(chart)
    valuesYAML, _ := converter.GetValuesYAML(chart)
    
    fmt.Println(chartYAML)
    fmt.Println(valuesYAML)
}
```

## Testing

### Run All Tests

```bash
make test
```

### Run Specific Package Tests

```bash
go test ./pkg/hcl
go test ./pkg/helm
go test ./pkg/validator
go test ./tests
```

### Run with Coverage

```bash
go test -cover ./...
```

### Run with Race Detection

```bash
make test-race
```

## Makefile Targets

- `make build` - Build the CLI tool
- `make test` - Run all tests
- `make test-race` - Run tests with race detection
- `make clean` - Clean build artifacts
- `make deps` - Install and tidy dependencies
- `make example` - Run example conversion
- `make fmt` - Format code
- `make lint` - Run linter
- `make install` - Install CLI tool to GOPATH/bin
- `make help` - Show help message

## Troubleshooting

### Error: "failed to parse HCL file"

This usually means the HCL syntax is invalid. Check:
- All blocks are properly closed with `}`
- Resource labels are provided (e.g., `resource "type" "name"`)
- Variable blocks have proper structure

### Error: "validation failed"

If validation fails, check:
- Required fields like `apiVersion`, `kind`, and `metadata.name` are present
- Resource types are correctly named
- API versions match the resource kind

### Missing Dependencies

If you get import errors, run:

```bash
make deps
```

Or manually:

```bash
go mod download
go mod tidy
```

## Contributing

See the main [README.md](README.md) for contribution guidelines.
