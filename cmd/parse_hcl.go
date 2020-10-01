package cmd

import (
	"github.com/gruntwork-io/terragrunt/config"
	"github.com/gruntwork-io/terragrunt/options"
	"github.com/gruntwork-io/terragrunt/util"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"

	log "github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

type parsedHcl struct {
	Terraform                   *terraformConfig           `hcl:"terraform,block"`
	TerraformBinary             *string                    `hcl:"terraform_binary,attr"`
	TerraformVersionConstraint  *string                    `hcl:"terraform_version_constraint,attr"`
	TerragruntVersionConstraint *string                    `hcl:"terragrunt_version_constraint,attr"`
	Inputs                      *cty.Value                 `hcl:"inputs,attr"`
	Include                     *config.IncludeConfig      `hcl:"include,block"`
	RemoteState                 *remoteStateConfigFile     `hcl:"remote_state,block"`
	Dependencies                *config.ModuleDependencies `hcl:"dependencies,block"`
	DownloadDir                 *string                    `hcl:"download_dir,attr"`
	PreventDestroy              *bool                      `hcl:"prevent_destroy,attr"`
	Skip                        *bool                      `hcl:"skip,attr"`
	IamRole                     *string                    `hcl:"iam_role,attr"`
	TerragruntDependencies      []config.Dependency        `hcl:"dependency,block"`
	GenerateBlocks              []terragruntGenerateBlock  `hcl:"generate,block"`

	// This struct is used for validating and parsing the entire terragrunt config. Since locals are evaluated in a
	// completely separate cycle, it should not be evaluated here. Otherwise, we can't support self referencing other
	// elements in the same block.
	Locals *terragruntLocal `hcl:"locals,block"`
}

// Configuration for Terraform remote state as parsed from a terragrunt.hcl config file
type remoteStateConfigFile struct {
	Backend                       string                     `hcl:"backend,attr"`
	DisableInit                   *bool                      `hcl:"disable_init,attr"`
	DisableDependencyOptimization *bool                      `hcl:"disable_dependency_optimization,attr"`
	Generate                      *remoteStateConfigGenerate `hcl:"generate,attr"`
	Config                        cty.Value                  `hcl:"config,attr"`
}

type remoteStateConfigGenerate struct {
	// We use cty instead of hcl, since we are using this type to convert an attr and not a block.
	Path     string `cty:"path"`
	IfExists string `cty:"if_exists"`
}

// We use a struct designed to not parse the block, as locals are parsed and decoded using a special routine that allows
// references to the other locals in the same block.
type terragruntLocal struct {
	Remain hcl.Body `hcl:",remain"`
}

type terraformConfig struct {
	Source *string `hcl:"source,attr"`
}

// Struct used to parse generate blocks. This will later be converted to GenerateConfig structs so that we can go
// through the codegen routine.
type terragruntGenerateBlock struct {
	Name             string  `hcl:",label"`
	Path             string  `hcl:"path,attr"`
	IfExists         string  `hcl:"if_exists,attr"`
	CommentPrefix    *string `hcl:"comment_prefix,attr"`
	Contents         string  `hcl:"contents,attr"`
	DisableSignature *bool   `hcl:"disable_signature,attr"`
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

// Create a cty Function that can be used to for calling read_terragrunt_config.
func readTerragruntConfig(dependencies *[]string) function.Function {
	log.Info("IN WRAPPER")

	return function.New(&function.Spec{
		// Takes one required string param
		Params: []function.Parameter{function.Parameter{Type: cty.String}},
		// And optional param that takes anything
		VarParam: &function.Parameter{Type: cty.DynamicPseudoType},
		// We don't know the return type until we parse the terragrunt config, so we use a dynamic type
		Type: function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			log.Info("IN IMPL")
			configPath := args[0].AsString()
			*dependencies = append(*dependencies, configPath)
			return cty.NilVal, nil
		},
	})
}

// func createEvalContext(
// 	filename string,
// 	terragruntOptions *options.TerragruntOptions,
// 	dependencies *[]string,
// ) *hcl.EvalContext {
// 	// DO NOT SUBMIT
// 	log.Info("IN CREATE TERRAGRUNT EVAL CONTEXT WITH PATH: " + filename)

// 	tfscope := tflang.Scope{
// 		BaseDir: filepath.Dir(filename),
// 	}

// 	terragruntFunctions := map[string]function.Function{
// 		// "find_in_parent_folders":                       wrapStringSliceToStringAsFuncImpl(findInParentFolders, extensions.Include, terragruntOptions),
// 		// "path_relative_to_include":                     wrapVoidToStringAsFuncImpl(pathRelativeToInclude, extensions.Include, terragruntOptions),
// 		// "path_relative_from_include":                   wrapVoidToStringAsFuncImpl(pathRelativeFromInclude, extensions.Include, terragruntOptions),
// 		// "get_env":                                      wrapStringSliceToStringAsFuncImpl(getEnvironmentVariable, extensions.Include, terragruntOptions),
// 		// "run_cmd":                                      wrapStringSliceToStringAsFuncImpl(runCommand, extensions.Include, terragruntOptions),
// 		// "read_terragrunt_config":                       readTerragruntConfigAsFuncImpl(terragruntOptions),
// 		// "get_platform":                                 wrapVoidToStringAsFuncImpl(getPlatform, extensions.Include, terragruntOptions),
// 		// "get_terragrunt_dir":                           wrapVoidToStringAsFuncImpl(getTerragruntDir, extensions.Include, terragruntOptions),
// 		// "get_terraform_command":                        wrapVoidToStringAsFuncImpl(getTerraformCommand, extensions.Include, terragruntOptions),
// 		// "get_terraform_cli_args":                       wrapVoidToStringSliceAsFuncImpl(getTerraformCliArgs, extensions.Include, terragruntOptions),
// 		// "get_parent_terragrunt_dir":                    wrapVoidToStringAsFuncImpl(getParentTerragruntDir, extensions.Include, terragruntOptions),
// 		// "get_aws_account_id":                           wrapVoidToStringAsFuncImpl(getAWSAccountID, extensions.Include, terragruntOptions),
// 		// "get_aws_caller_identity_arn":                  wrapVoidToStringAsFuncImpl(getAWSCallerIdentityARN, extensions.Include, terragruntOptions),
// 		// "get_aws_caller_identity_user_id":              wrapVoidToStringAsFuncImpl(getAWSCallerIdentityUserID, extensions.Include, terragruntOptions),
// 		// "get_terraform_commands_that_need_vars":        wrapStaticValueToStringSliceAsFuncImpl(TERRAFORM_COMMANDS_NEED_VARS),
// 		// "get_terraform_commands_that_need_locking":     wrapStaticValueToStringSliceAsFuncImpl(TERRAFORM_COMMANDS_NEED_LOCKING),
// 		// "get_terraform_commands_that_need_input":       wrapStaticValueToStringSliceAsFuncImpl(TERRAFORM_COMMANDS_NEED_INPUT),
// 		// "get_terraform_commands_that_need_parallelism": wrapStaticValueToStringSliceAsFuncImpl(TERRAFORM_COMMANDS_NEED_PARALLELISM),
// 		// "sops_decrypt_file":                            wrapStringSliceToStringAsFuncImpl(sopsDecryptFile, extensions.Include, terragruntOptions),

// 		"read_terragrunt_config": readTerragruntConfig(dependencies),
// 	}

// 	functions := map[string]function.Function{}
// 	for k, v := range tfscope.Functions() {
// 		functions[k] = v
// 	}
// 	for k, v := range terragruntFunctions {
// 		functions[k] = v
// 	}

// 	ctx := &hcl.EvalContext{
// 		Functions: functions,
// 	}
// 	ctx.Variables = map[string]cty.Value{}

// 	// DO NOT SUBMIT
// 	log.Info("RETURNING CTX")
// 	log.Info(ctx.Functions["read_terragrunt_config"])

// 	return ctx
// }

func extractDependencies(path string, terragruntOptions *options.TerragruntOptions) ([]string, error) {
	// DO NOT SUBMIT
	log.Info("IN EXTRACT DEPS WITH PATH: " + path)

	configString, err := util.ReadFileAsString(path)
	if err != nil {
		return nil, err
	}

	parser := hclparse.NewParser()
	file, err := parseHcl(parser, configString, path)
	if err != nil {
		return nil, err
	}

	// Have functions that read in files update a dependency list
	dependencies := []string{}
	extensions := config.EvalContextExtensions{}
	evalContext := config.CreateTerragruntEvalContext(path, terragruntOptions, extensions)
	delete(evalContext.Functions, "read_terragrunt_config")
	// evalContext.Functions["read_terragrunt_config"] = readTerragruntConfig(&dependencies)

	// ctx := createEvalContext(path, terragruntOptions, &dependencies)

	// The HCL2 parser and especially cty conversions will panic in many types of errors, so we have to recover from
	// those panics here and convert them to normal errors
	// defer func() {
	// 	if recovered := recover(); recovered != nil {
	// 		// err = errors.WithStackTrace(PanicWhileParsingConfig{RecoveredValue: recovered, ConfigFile: filename})
	// 		log.Error(recovered)
	// 	}
	// }()

	var parsed parsedHcl
	diags := gohcl.DecodeBody(file.Body, evalContext, &parsed)

	// DO NOT SUBMIT
	// DO NOT SUBMIT
	log.Info("#################")
	log.Info(diags)
	log.Info("#################")
	// log.Info(dependencies)
	// log.Info(*parsed.Terraform.Source)

	return dependencies, nil
}
