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

// ResolvedLocals are the parsed result of local values this module cares about
type ResolvedLocals struct {
	// The Atlantis workflow to use for some project
	AtlantisWorkflow string

	// Apply requirements to override the global `--apply-requirements` flag
	ApplyRequirements []string

	// Extra dependencies that can be hardcoded in config
	ExtraAtlantisDependencies []string

	// If set, a single module will have autoplan turned to this setting
	AutoPlan *bool

	// If set to true, the module will not be included in the output
	Skip *bool

	// Terraform version to use just for this project
	TerraformVersion string
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
	configString, err := util.ReadFileAsString(path)
	if err != nil {
		return ResolvedLocals{}, err
	}

	// Parse the HCL string into an AST body
	parser := hclparse.NewParser()
	file, err := parseHcl(parser, configString, path)
	if err != nil {
		return ResolvedLocals{}, err
	}

	// Decode just the Base blocks. See the function docs for DecodeBaseBlocks for more info on what base blocks are.
	localsAsCty, _, includeConfig, err := config.DecodeBaseBlocks(terragruntOptions, parser, file, path, includeFromChild)
	if err != nil {
		return ResolvedLocals{}, err
	}

	// Recurse on the parent to merge in the locals from that file
	parentLocals := ResolvedLocals{}
	if includeConfig != nil && includeFromChild == nil {
		// Ignore errors if the parent cannot be parsed. Terragrunt Errors still will be logged
		parentLocals, _ = parseLocals(includeConfig.Path, terragruntOptions, includeConfig)
	}
	childLocals := resolveLocals(*localsAsCty)

	// Merge in values from child => parent local values
	if childLocals.AtlantisWorkflow != "" {
		parentLocals.AtlantisWorkflow = childLocals.AtlantisWorkflow
	}

	if childLocals.TerraformVersion != "" {
		parentLocals.TerraformVersion = childLocals.TerraformVersion
	}

	if childLocals.AutoPlan != nil {
		parentLocals.AutoPlan = childLocals.AutoPlan
	}

	if childLocals.Skip != nil {
		parentLocals.Skip = childLocals.Skip
	}

	if childLocals.ApplyRequirements != nil || len(childLocals.ApplyRequirements) > 0 {
		parentLocals.ApplyRequirements = childLocals.ApplyRequirements
	}

	for _, dep := range childLocals.ExtraAtlantisDependencies {
		parentLocals.ExtraAtlantisDependencies = append(
			parentLocals.ExtraAtlantisDependencies,
			dep,
		)
	}

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

	versionValue, ok := rawLocals["atlantis_terraform_version"]
	if ok {
		resolved.TerraformVersion = versionValue.AsString()
	}

	autoPlanValue, ok := rawLocals["atlantis_autoplan"]
	if ok {
		hasValue := autoPlanValue.True()
		resolved.AutoPlan = &hasValue
	}

	skipValue, ok := rawLocals["atlantis_skip"]
	if ok {
		hasValue := skipValue.True()
		resolved.Skip = &hasValue
	}

	applyReqs, ok := rawLocals["atlantis_apply_requirements"]
	if ok {
		it := applyReqs.ElementIterator()
		for it.Next() {
			_, val := it.Element()
			resolved.ApplyRequirements = append(resolved.ApplyRequirements, val.AsString())
		}
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
