# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive GitHub Actions CI/CD pipeline
- Multi-platform builds (Linux, macOS, Windows)
- Docker containerization with GitHub Container Registry
- Code quality checks with golangci-lint and staticcheck
- Security scanning with gosec and nancy
- Test coverage reporting with codecov
- Automated releases on version tags
- Contributing guidelines (CONTRIBUTING.md)
- golangci-lint configuration file

### Changed
- **BREAKING**: Removed support for single file processing
- **BREAKING**: Simplified API to directory-only processing
- Renamed `ConvertDirectory()` to `Convert()` in converter package
- Modernized from deprecated `io/ioutil` to `os` package functions
- Updated all tests to use modern Go stdlib functions
- Improved error handling and validation

### Removed
- **BREAKING**: `ConvertFile()` method from converter package
- **BREAKING**: `ConvertBytes()` method from converter package
- **BREAKING**: `ParseFile()` method from HCL parser package
- **BREAKING**: `ParseBytes()` method from HCL parser package
- Usage of deprecated `io/ioutil` package

### Technical
- Upgraded to Go 1.20+ requirements
- Enhanced test coverage across all packages
- Improved CLI input validation
- Better error messages and user experience

## [Previous Versions]

### [0.1.0] - Initial Release
- Basic HCL to Helm chart conversion
- Support for both file and directory processing
- Core converter and parser functionality
- Basic CLI interface
- Initial test suite

---

## Migration Guide

### From 0.x to 1.0

**Breaking Changes:**

1. **API Simplification**: The converter now only accepts directories as input.

   **Before:**
   ```go
   // These methods no longer exist
   converter.ConvertFile("config.hcl", "output/")
   converter.ConvertBytes(data, "output/")
   converter.ConvertDirectory("input/", "output/")
   ```

   **After:**
   ```go
   // Only directory-to-directory conversion
   converter.Convert("input/", "output/")
   ```

2. **CLI Usage**: Single file processing is no longer supported.

   **Before:**
   ```bash
   helm-dev-kit config.hcl output/
   ```

   **After:**
   ```bash
   # Only directory processing
   helm-dev-kit input-dir/ output-dir/
   ```

3. **Import Changes**: If you were using internal parser methods, they are no longer public.

   **Before:**
   ```go
   parser.ParseFile("config.hcl")
   parser.ParseBytes(data)
   ```

   **After:**
   ```go
   // Only directory parsing is public
   parser.ParseDirectory("input-dir/")
   ```

**Migration Steps:**

1. Update your code to use directory-based processing
2. Organize single files into directories if needed
3. Update import statements if using internal APIs
4. Test with the new API to ensure compatibility

**Benefits of Migration:**

- Simpler, more focused API
- Better support for complex Helm charts
- Improved error handling
- Modern Go practices
- Enhanced CI/CD pipeline
- Better security and code quality