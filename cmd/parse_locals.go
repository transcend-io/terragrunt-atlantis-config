package cmd

// Terragrunt doesn't give us an easy way to access all of the Locals from a module
// in an easy to digest way. This file is mostly just follows along how Terragrunt
// parses the `locals` blocks and evaluates their contents.

import (
	"github.com/gruntwork-io/terragrunt/config"
	"github.com/gruntwork-io/terragrunt/options"
	"github.com/gruntwork-io/terragrunt/util"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"

	"path/filepath"
)

// parseHcl uses the HCL2 parser to parse the given string into an HCL file body.
func parseHcl(parser *hclparse.Parser, hcl string, filename string) (file *hcl.File, err error) {
	if filepath.Ext(filename) == ".json" {
		file, parseDiagnostics := parser.ParseJSON([]byte(hcl), filename)
		if parseDiagnostics != nil && parseDiagnostics.HasErrors() {
			return nil, parseDiagnostics
		}

		return file, nil
	}

	file, parseDiagnostics := parser.ParseHCL([]byte(hcl), filename)
	if parseDiagnostics != nil && parseDiagnostics.HasErrors() {
		return nil, parseDiagnostics
	}

	return file, nil
}

// Parses a given file, returning a map of all it's `local` values
func parseLocals(path string) (map[string]cty.Value, error) {
	configString, err := util.ReadFileAsString(path)
	if err != nil {
		return nil, err
	}

	options, err := options.NewTerragruntOptions(path)
	if err != nil {
		return nil, err
	}

	// Parse the HCL string into an AST body
	parser := hclparse.NewParser()
	file, err := parseHcl(parser, configString, path)
	if err != nil {
		return nil, err
	}

	// Decode just the Base blocks. See the function docs for DecodeBaseBlocks for more info on what base blocks are.
	localsAsCty, _, _, err := config.DecodeBaseBlocks(options, parser, file, path, nil)
	if err != nil {
		return nil, err
	}

	// If there are no locals, return early
	if *localsAsCty == cty.NilVal {
		return map[string]cty.Value{}, nil
	}

	return localsAsCty.AsValueMap(), nil
}

// Parses the terragrunt config at <path> to find any `locals` values with
// the name `extra_atlantis_dependencies`
func parseLocalDependencies(path string) ([]string, error) {
	locals, err := parseLocals(path)
	if err != nil {
		return nil, err
	}

	extraDependenciesAsCty, ok := locals["extra_atlantis_dependencies"]
	if !ok {
		return []string{}, nil
	}

	dependencies := []string{}

	it := extraDependenciesAsCty.ElementIterator()
	for it.Next() {
		_, val := it.Element()
		dependencies = append(dependencies, filepath.ToSlash(val.AsString()))
	}

	return dependencies, nil
}
