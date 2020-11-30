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
func parseLocals(path string, terragruntOptions *options.TerragruntOptions, includeFromChild *config.IncludeConfig) error {
	configString, err := util.ReadFileAsString(path)
	if err != nil {
		return err
	}

	// Parse the HCL string into an AST body
	parser := hclparse.NewParser()
	file, err := parseHcl(parser, configString, path)
	if err != nil {
		return err
	}

	// Decode just the Base blocks. See the function docs for DecodeBaseBlocks for more info on what base blocks are.
	localsAsCty, _, includeConfig, err := config.DecodeBaseBlocks(terragruntOptions, parser, file, path, includeFromChild)
	if err != nil {
		return err
	}

	// Recurse on the parent to merge in the locals from that file
	if includeConfig != nil && includeFromChild == nil {
		// Ignore errors if the parent cannot be parsed. Terragrunt Errors still will be logged
		parseLocals(includeConfig.Path, terragruntOptions, includeConfig)
	}

	// If no `locals` block was found, just exit cleanly
	if *localsAsCty == cty.NilVal {
		return nil
	}

	// Store all `locals` values onto the param
	rawLocals := localsAsCty.AsValueMap()
	for _, param := range []*ParameterValue{&workflowParameter, &extraDependenciesParameter} {
		if param.LocalsName != "" {
			val, ok := rawLocals[(*param).LocalsName]
			if ok {
				if includeConfig == nil {
					(*param).LocalValue = &val
				} else {
					(*param).ParentLocalValue = &val
				}
			} else {
				if includeConfig == nil {
					(*param).LocalValue = nil
				} else {
					(*param).ParentLocalValue = nil
				}
			}
		}
	}

	return nil
}
