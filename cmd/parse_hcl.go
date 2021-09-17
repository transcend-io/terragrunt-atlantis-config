package cmd

import (
	"github.com/gruntwork-io/terragrunt/config"
	"github.com/gruntwork-io/terragrunt/options"
	"github.com/gruntwork-io/terragrunt/util"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty/function"
)

type parsedHcl struct {
	Terraform *terraformConfig `hcl:"terraform,block"`
	Includes  *includeConfig   `hcl:"include,block"`
}

type terraformConfig struct {
	Source *string `hcl:"source,attr"`
}

type includeConfig struct {
	Path *string `hcl:"path,attr"`
}

// Not all modules need an include statement, as they could define everything in one file without a parent
// The key signifiers of a parent are:
//   - no include statement
//   - no terraform source defined
// If both of those are true, it is likely a parent module
func isParentModule(path string, terragruntOptions *options.TerragruntOptions) (bool, error) {
	configString, err := util.ReadFileAsString(path)
	if err != nil {
		return false, err
	}

	parser := hclparse.NewParser()
	file, err := parseHcl(parser, configString, path)
	if err != nil {
		return false, err
	}

	extensions := config.EvalContextExtensions{}
	evalContext, err := config.CreateTerragruntEvalContext(path, terragruntOptions, extensions)
	if err != nil {
		return false, err
	}

	// Mock all the functions out so they don't do anything. Otherwise they may throw errors that we don't care about
	evalContext.Functions = map[string]function.Function{}

	// We don't need to check the errors/diagnostics coming from `DecodeBody`, as when errors come up,
	// it will leave the partially parsed result in the output object.
	var parsed parsedHcl
	gohcl.DecodeBody(file.Body, evalContext, &parsed)

	// If the file has an `includes` block, it cannot be a parent as terragrunt only allows one level of inheritance
	if parsed.Includes != nil && parsed.Includes.Path != nil {
		return false, nil
	}

	// If the file does not define a terraform source block, it is likely a parent (though not guaranteed)
	if parsed.Terraform == nil || parsed.Terraform.Source == nil {
		return true, nil
	}

	return false, nil
}
