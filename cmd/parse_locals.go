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

var (
	parseLocalsCache map[string]ParseLocalResult = make(map[string]ParseLocalResult)
)

type ParseLocalResult struct {
	resolvedLocals ResolvedLocals
	err            error
}

// ResolvedLocals are the parsed result of local values this module cares about
type ResolvedLocals struct {
	// The Atlantis workflow to use for some project
	AtlantisWorkflow string

	// Extra dependencies that can be hardcoded in config
	ExtraAtlantisDependencies []string
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

// Parses a given file, returning a map of all it's `local` values
func parseLocals(path string, terragruntOptions *options.TerragruntOptions, includeFromChild *config.IncludeConfig) (ResolvedLocals, error) {
	if cachedResult, ok := parseLocalsCache[path]; ok {
		return cachedResult.resolvedLocals, cachedResult.err
	}

	configString, err := util.ReadFileAsString(path)
	if err != nil {
		parseLocalsCache[path] = ParseLocalResult{err: err}
		return ResolvedLocals{}, err
	}

	// Parse the HCL string into an AST body
	parser := hclparse.NewParser()
	file, err := parseHcl(parser, configString, path)
	if err != nil {
		parseLocalsCache[path] = ParseLocalResult{err: err}
		return ResolvedLocals{}, err
	}

	// Decode just the Base blocks. See the function docs for DecodeBaseBlocks for more info on what base blocks are.
	localsAsCty, _, includeConfig, err := config.DecodeBaseBlocks(terragruntOptions, parser, file, path, includeFromChild)
	if err != nil {
		parseLocalsCache[path] = ParseLocalResult{err: err}
		return ResolvedLocals{}, err
	}

	// Recurse on the parent to merge in the locals from that file
	parentLocals := ResolvedLocals{}
	if includeConfig != nil && includeFromChild == nil {
		// Ignore errors if the parent cannot be parsed. Terragrunt Errors still will be logged
		parentLocals, _ = parseLocals(includeConfig.Path, terragruntOptions, includeConfig)
	}

	childLocals := resolveLocals(*localsAsCty)
	if childLocals.AtlantisWorkflow != "" {
		parentLocals.AtlantisWorkflow = childLocals.AtlantisWorkflow
	}

	for _, dep := range childLocals.ExtraAtlantisDependencies {
		parentLocals.ExtraAtlantisDependencies = append(
			parentLocals.ExtraAtlantisDependencies,
			dep,
		)
	}

	parseLocalsCache[path] = ParseLocalResult{resolvedLocals: parentLocals}
	return parentLocals, nil
}

func resolveLocals(localsAsCty cty.Value) ResolvedLocals {
	resolved := ResolvedLocals{}

	// Return an empty set of locals if no `locals` block was present
	if localsAsCty == cty.NilVal {
		return resolved
	}
	rawLocals := localsAsCty.AsValueMap()

	workflowValue, ok := rawLocals["atlantis_workflow"]
	if ok {
		resolved.AtlantisWorkflow = workflowValue.AsString()
	}

	extraDependenciesAsCty, ok := rawLocals["extra_atlantis_dependencies"]
	if ok {
		it := extraDependenciesAsCty.ElementIterator()
		for it.Next() {
			_, val := it.Element()
			resolved.ExtraAtlantisDependencies = append(
				resolved.ExtraAtlantisDependencies,
				filepath.ToSlash(val.AsString()),
			)
		}
	}

	return resolved
}
