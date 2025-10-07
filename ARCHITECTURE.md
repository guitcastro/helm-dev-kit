# Architecture

## Overview

`helm-dev-kit` is a Go framework that converts Terraform-style HCL configurations into validated Helm charts. The architecture is designed to be modular, testable, and extensible.

## Components

### 1. HCL Parser (`pkg/hcl`)

**Purpose**: Parse Terraform-style HCL files and extract resources and variables.

**Key Types**:
- `Parser`: Main parser struct using `hashicorp/hcl/v2`
- `Resource`: Represents a Kubernetes resource definition
- `Variable`: Represents a variable definition with type, default, and description

**Responsibilities**:
- Parse HCL files and byte streams
- Extract resource blocks with their attributes
- Extract variable blocks with metadata
- Convert HCL types (cty.Value) to Go native types

**Dependencies**:
- `github.com/hashicorp/hcl/v2` - HCL parsing
- `github.com/zclconf/go-cty` - HCL type system

### 2. Helm Converter (`pkg/helm`)

**Purpose**: Convert HCL resources and variables into Helm chart components.

**Key Types**:
- `Converter`: Handles conversion logic
- `Chart`: Represents a complete Helm chart
- `Template`: Represents a single Helm template file
- `Values`: Map of values for values.yaml

**Responsibilities**:
- Map HCL variables to Helm values.yaml
- Convert HCL resources to Kubernetes manifests
- Generate Helm template syntax for dynamic values
- Determine correct Kubernetes API versions for resource types
- Generate Chart.yaml metadata

**Key Methods**:
- `Convert()`: Main conversion entry point
- `resourceToTemplate()`: Converts single resource to template
- `parseKind()`: Extracts Kubernetes kind from HCL resource type
- `getAPIVersion()`: Returns appropriate API version for kind
- `processValue()`: Converts variable references to Helm template syntax

### 3. Kubernetes Validator (`pkg/validator`)

**Purpose**: Validate generated Kubernetes manifests against schemas.

**Key Types**:
- `Validator`: Validation orchestrator

**Responsibilities**:
- Validate Kubernetes resource structure
- Check required fields (apiVersion, kind, metadata)
- Support dry-run validation simulation
- Integrate with Kubernetes discovery and dynamic clients

**Dependencies**:
- `k8s.io/client-go` - Kubernetes client libraries
- `k8s.io/apimachinery` - Kubernetes API machinery
- `gopkg.in/yaml.v3` - YAML processing

### 4. Converter Orchestrator (`pkg/converter`)

**Purpose**: Coordinate the conversion process with validation.

**Key Types**:
- `HCLToHelm`: Main orchestrator

**Responsibilities**:
- Orchestrate parsing, conversion, and validation
- Provide simple API for end-to-end conversion
- Handle error aggregation and reporting

**Workflow**:
1. Parse HCL using `pkg/hcl`
2. Convert to Helm using `pkg/helm`
3. Validate templates using `pkg/validator` (optional)
4. Return validated chart

### 5. CLI Tool (`cmd/helm-dev-kit`)

**Purpose**: Command-line interface for HCL to Helm conversion.

**Responsibilities**:
- Parse command-line arguments
- Invoke converter
- Write chart files to disk
- Provide user feedback

**Usage**:
```bash
helm-dev-kit <input.hcl> <output-dir> <chart-name>
```

## Data Flow

```
┌─────────────┐
│  HCL File   │
└──────┬──────┘
       │
       v
┌─────────────────┐
│  HCL Parser     │
│  (pkg/hcl)      │
└──────┬──────────┘
       │
       v
┌──────────────────┐
│  Resources +     │
│  Variables       │
└──────┬───────────┘
       │
       v
┌─────────────────┐
│  Helm Converter │
│  (pkg/helm)     │
└──────┬──────────┘
       │
       v
┌──────────────────┐
│  Helm Chart      │
│  - Chart.yaml    │
│  - values.yaml   │
│  - templates/    │
└──────┬───────────┘
       │
       v
┌─────────────────┐
│  Validator      │
│  (pkg/validator)│
└──────┬──────────┘
       │
       v
┌──────────────────┐
│  Validated Chart │
└──────────────────┘
```

## Design Decisions

### 1. Modular Package Structure

**Rationale**: Each component has a single responsibility and can be used independently.

**Benefits**:
- Easy to test individual components
- Can use components separately (e.g., just HCL parsing)
- Clear separation of concerns

### 2. Offline-First Validation

**Rationale**: Validator can work without Kubernetes cluster connection.

**Benefits**:
- CI/CD friendly
- No cluster required for basic validation
- Fast feedback during development

**Trade-offs**:
- Cannot validate against actual cluster state
- Limited schema validation without API server

### 3. Template Naming Convention

**Rationale**: Template files named as `{resource-name}-{kind}.yaml`

**Benefits**:
- Unique names prevent collisions
- Clear identification of resource type
- Follows Helm best practices

**Example**: `web-deployment.yaml`, `web-service.yaml`

### 4. Variable Mapping Strategy

**Rationale**: HCL variables map directly to Helm values with same names.

**Benefits**:
- Predictable and intuitive
- Simple to understand and debug
- No complex transformations

**Example**:
```hcl
variable "replicas" {
  default = 3
}
```
↓
```yaml
replicas: 3
```

### 5. Value Processing

**Rationale**: Variable references (`var.name`) converted to Helm syntax (`{{ .Values.name }}`).

**Benefits**:
- Preserves dynamic nature of variables
- Allows runtime value customization
- Standard Helm idiom

## Testing Strategy

### Unit Tests

- Each package has comprehensive unit tests
- Test individual functions and methods
- Mock external dependencies
- Coverage: 40-70% per package

**Locations**:
- `pkg/hcl/parser_test.go`
- `pkg/helm/converter_test.go`
- `pkg/validator/validator_test.go`

### Integration Tests

- End-to-end workflow tests
- Test component interactions
- Verify complete conversion pipeline

**Location**:
- `tests/integration_test.go`

### Manual Testing

- CLI tool testing with real examples
- Visual inspection of generated charts
- Makefile `example` target for quick testing

## Extension Points

### Adding New Resource Types

1. Add kind mapping in `pkg/helm/converter.go`:
```go
func (c *Converter) getAPIVersion(kind string) string {
    apiVersions := map[string]string{
        "NewKind": "group/version",
        // ...
    }
}
```

### Custom Validation Rules

1. Extend `pkg/validator/validator.go`:
```go
func (v *Validator) ValidateCustomRule(resource string) error {
    // Custom validation logic
}
```

### Additional Output Formats

1. Create new converter in `pkg/`:
```go
package kustomize

type Converter struct {
    // Convert to Kustomize format
}
```

## Dependencies

### Core Dependencies

- `github.com/hashicorp/hcl/v2` - HCL parsing
- `github.com/zclconf/go-cty` - HCL type system
- `helm.sh/helm/v3` - Helm chart structures
- `k8s.io/client-go` - Kubernetes client
- `k8s.io/apimachinery` - Kubernetes API types
- `sigs.k8s.io/e2e-framework` - Testing framework
- `gopkg.in/yaml.v3` - YAML marshaling

### Development Dependencies

- Go 1.21+ (uses standard library features)
- Make (for build automation)

## Performance Considerations

### Memory

- HCL files parsed in memory (suitable for typical configuration sizes)
- Large files (>100MB) may require streaming approach
- Template generation creates strings in memory

### Speed

- Parsing: Fast (sub-second for typical files)
- Conversion: O(n) where n = number of resources
- Validation: Depends on Kubernetes client availability

### Optimization Opportunities

1. **Parallel Processing**: Convert multiple resources concurrently
2. **Caching**: Cache parsed HCL or validation results
3. **Streaming**: Stream large files instead of loading entirely

## Security Considerations

1. **Input Validation**: HCL parser validates syntax
2. **Path Traversal**: File operations use absolute paths
3. **Resource Limits**: No explicit limits on resource count (TODO)
4. **Secrets**: No special handling for sensitive data (use Kubernetes secrets)

## Future Enhancements

1. **Variable References**: Support for `var.name` in HCL (currently must use literals)
2. **Advanced Validation**: OpenAPI schema validation with Kubernetes API
3. **Multi-Chart Support**: Generate multiple charts from single HCL file
4. **Helm Hooks**: Support for Helm lifecycle hooks
5. **Chart Dependencies**: Generate Chart.yaml dependencies section
6. **Interactive Mode**: CLI prompts for missing values
7. **Watch Mode**: Auto-regenerate on file changes
8. **Helm Packaging**: Directly package charts into .tgz files

## Contributing

See [README.md](README.md) for contribution guidelines.

### Code Style

- Follow Go conventions and idioms
- Use `gofmt` for formatting
- Add tests for new features
- Document public APIs
- Keep functions focused and small

### Pull Request Process

1. Write tests for new functionality
2. Ensure all tests pass (`make test`)
3. Update documentation
4. Run `make fmt` and `make lint`
5. Submit PR with clear description
