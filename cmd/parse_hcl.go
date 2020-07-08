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
}

type terraformConfig struct {
	Source *string `hcl:"source,attr"`
}

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
	evalContext := config.CreateTerragruntEvalContext(path, terragruntOptions, extensions)

	// Mock all the functions out so they don't do anything. Otherwise they may throw errors that we don't care about
	evalContext.Functions = map[string]function.Function{}

	// We don't need to check the errors/diagnostics coming from `DecodeBody`, as when errors come up,
	// it will leave the partially parsed result in the output object.
	var parsed parsedHcl
	gohcl.DecodeBody(file.Body, evalContext, &parsed)

	if parsed.Terraform == nil || parsed.Terraform.Source == nil {
		return true, nil
	}

	return false, nil
}
