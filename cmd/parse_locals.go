package cmd

// Terragrunt doesn't give us an easy way to access all of the Locals from a module
// in an easy to digest way. This file is mostly just follows along how Terragrunt
// parses the `locals` blocks and evaluates their contents.

import (
	"encoding/json"

	"github.com/gruntwork-io/terragrunt/config"
	terragruntErrors "github.com/gruntwork-io/terragrunt/errors"
	"github.com/gruntwork-io/terragrunt/options"
	"github.com/gruntwork-io/terragrunt/util"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"

	"path/filepath"
)

// This is a hacky workaround to convert a cty Value to a Go map[string]interface{}. cty does not support this directly
// (https://github.com/hashicorp/hcl2/issues/108) and doing it with gocty.FromCtyValue is nearly impossible, as cty
// requires you to specify all the output types and will error out when it hits interface{}. So, as an ugly workaround,
// we convert the given value to JSON using cty's JSON library and then convert the JSON back to a
// map[string]interface{} using the Go json library.
func parseCtyValueToMap(value cty.Value) (map[string]interface{}, error) {
	jsonBytes, err := ctyjson.Marshal(value, cty.DynamicPseudoType)
	if err != nil {
		return nil, terragruntErrors.WithStackTrace(err)
	}

	var ctyJsonOutput config.CtyJsonOutput
	if err := json.Unmarshal(jsonBytes, &ctyJsonOutput); err != nil {
		return nil, terragruntErrors.WithStackTrace(err)
	}

	return ctyJsonOutput.Value, nil
}

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

// Parses the terragrunt config at <path> to find any `locals` values with
// the name `extra_atlantis_dependencies`
func parseLocalDependencies(path string) ([]string, error) {
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
		return []string{}, nil
	}

	extraDependenciesAsCty, ok := localsAsCty.AsValueMap()["extra_atlantis_dependencies"]
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
