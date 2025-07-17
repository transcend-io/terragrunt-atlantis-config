# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

terragrunt-atlantis-config is a Go CLI tool that generates Atlantis configuration files (`atlantis.yaml`) for Terragrunt projects. It analyzes Terragrunt dependencies and creates proper Atlantis workflows that respect the dependency graph.

## Common Development Commands

### Building
- `make build` - Build binary for current OS/arch
- `make build-all` - Cross-compile for Linux/Darwin/Windows (amd64/arm64)
- `make install` - Install to ~/.local/bin/

### Testing
- `make gotestsum` - Run tests with gotestsum (preferred)
- `make test` - Run standard Go tests
- To run a single test: `go test -run TestName ./cmd`

### Development
- `make clean` - Clean build artifacts
- `make sign` - Generate SHA256 checksums for releases

## Code Architecture

### Command Structure
The codebase follows a command pattern using Cobra:
- `main.go` - Entry point, sets version
- `cmd/root.go` - Base command setup
- `cmd/generate.go` - Core logic for scanning and generating config
- `cmd/parse_*.go` - Parsing utilities for HCL, locals, and Terraform configs
- `cmd/config.go` - Configuration management

### Key Architectural Patterns
1. **Dependency Graph Building**: The tool builds a DAG of Terragrunt modules by parsing `terragrunt.hcl` files and evaluating dependencies
2. **Golden File Testing**: Tests use golden files in `cmd/golden/` to verify expected YAML outputs
3. **Recursive Directory Walking**: Scans for all `terragrunt.hcl` files in the repository
4. **HCL Parsing**: Uses HashiCorp's HCL v2 library for parsing Terragrunt configurations

### Important Implementation Details
- The tool evaluates Terragrunt locals and terraform blocks to understand module relationships
- Supports both local and remote module sources
- Can handle complex dependency chains and multiple includes
- Respects Terragrunt-specific locals like `atlantis_skip`, `atlantis_workflow`, etc.

## Testing Approach

Tests are located in `cmd/generate_test.go` with test fixtures in `test_examples/`. When adding new features:
1. Add test cases to `test_examples/` directory
2. Generate expected output in `cmd/golden/` directory
3. Use `gotestsum` for better test output visibility

## Key Dependencies

- Terragrunt library (v0.83.2) - Core dependency evaluation
- Cobra - CLI framework
- HCL v2 - Configuration parsing
- Various AWS/Azure/GCP SDKs - Cloud provider support