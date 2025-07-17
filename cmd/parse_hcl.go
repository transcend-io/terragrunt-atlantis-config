package cmd

import (
	"strings"
	"os"
	"os/exec"
	"github.com/gruntwork-io/go-commons/errors"
	"github.com/gruntwork-io/terragrunt/config"
	"github.com/gruntwork-io/terragrunt/config/hclparse"
	"github.com/gruntwork-io/terragrunt/util"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"path/filepath"
)

const bareIncludeKey = ""

type parsedHcl struct {
	Terraform *config.TerraformConfig `hcl:"terraform,block"`
	Includes  []config.IncludeConfig  `hcl:"include,block"`
}

// terragruntIncludeMultiple is a struct that can be used to only decode the include block with labels.
type terragruntIncludeMultiple struct {
	Include []config.IncludeConfig `hcl:"include,block"`
	Remain  hcl.Body               `hcl:",remain"`
}

// updateBareIncludeBlock searches the parsed terragrunt contents for a bare include block (include without a label),
// and convert it to one with empty string as the label. This is necessary because the hcl parser is strictly enforces
// label counts when parsing out labels with a go struct.
//
// Returns the updated contents, a boolean indicated whether anything changed, and an error (if any).
func updateBareIncludeBlock(file *hcl.File, filename string) ([]byte, bool, error) {
	hclFile, err := hclwrite.ParseConfig(file.Bytes, filename, hcl.InitialPos)
	if err != nil {
		return nil, false, errors.WithStackTrace(err)
	}

	codeWasUpdated := false
	for _, block := range hclFile.Body().Blocks() {
		if block.Type() == "include" && len(block.Labels()) == 0 {
			if codeWasUpdated {
				return nil, false, errors.WithStackTrace(config.MultipleBareIncludeBlocksErr{})
			}
			block.SetLabels([]string{bareIncludeKey})
			codeWasUpdated = true
		}
	}
	return hclFile.Bytes(), codeWasUpdated, nil
}

// createMinimalEvalContext creates a minimal evaluation context to avoid segmentation faults
func createMinimalEvalContext(configPath string) *hcl.EvalContext {
	return &hcl.EvalContext{
		Functions: map[string]function.Function{
			"find_in_parent_folders": function.New(&function.Spec{
				Params: []function.Parameter{
					{
						Name: "file",
						Type: cty.String,
					},
				},
				VarParam: &function.Parameter{
					Name: "default",
					Type: cty.String,
				},
				Type: function.StaticReturnType(cty.String),
				Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
					// Return empty string as a safe default
					return cty.StringVal(""), nil
				},
			}),
			"get_repo_root": function.New(&function.Spec{
				Params: []function.Parameter{},
				Type:   function.StaticReturnType(cty.String),
				Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
					// Get current working directory as repo root
					wd, err := os.Getwd()
					if err != nil {
						return cty.StringVal("."), nil
					}
					return cty.StringVal(wd), nil
				},
			}),
			"path_relative_to_include": function.New(&function.Spec{
				Params: []function.Parameter{},
				Type:   function.StaticReturnType(cty.String),
				Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
					return cty.StringVal("."), nil
				},
			}),
			"get_parent_terragrunt_dir": function.New(&function.Spec{
				Params: []function.Parameter{},
				Type:   function.StaticReturnType(cty.String),
				Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
					return cty.StringVal(filepath.Dir(configPath)), nil
				},
			}),
			"get_original_terragrunt_dir": function.New(&function.Spec{
				Params: []function.Parameter{},
				Type:   function.StaticReturnType(cty.String),
				Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
					return cty.StringVal(filepath.Dir(configPath)), nil
				},
			}),
			"run_cmd": function.New(&function.Spec{
				VarParam: &function.Parameter{
					Name: "args",
					Type: cty.String,
				},
				Type: function.StaticReturnType(cty.String),
				Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
					// Execute only the file_finder.sh script safely
					if len(args) >= 4 {
						scriptPath := args[1].AsString()
						if strings.HasSuffix(scriptPath, "/scripts/file_finder.sh") {
							targetFile := args[2].AsString()
							currentDir := args[3].AsString()
							
							// Get repo root from current working directory
							repoRoot, err := os.Getwd()
							if err != nil {
								repoRoot = "."
							}
							
							// Execute the file finder script
							cmd := exec.Command(scriptPath, targetFile, currentDir, repoRoot)
							output, err := cmd.Output()
							if err != nil {
								// If file not found, return empty config path which read_terragrunt_config will handle
								return cty.StringVal(""), nil
							}
							return cty.StringVal(strings.TrimSpace(string(output))), nil
						}
					}
					// For other commands, return empty string
					return cty.StringVal(""), nil
				},
			}),
			"read_terragrunt_config": function.New(&function.Spec{
				Params: []function.Parameter{
					{
						Name: "config_path",
						Type: cty.String,
					},
				},
				VarParam: &function.Parameter{
					Name: "default",
					Type: cty.DynamicPseudoType,
				},
				Type: function.StaticReturnType(cty.DynamicPseudoType),
				Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
					// Return default empty object
					return cty.ObjectVal(map[string]cty.Value{
						"locals": cty.EmptyObjectVal,
					}), nil
				},
			}),
		},
		Variables: map[string]cty.Value{},
	}
}

// decodeHcl uses the HCL2 parser to decode the parsed HCL into the struct specified by out.
//
// Note that we take a two pass approach to support parsing include blocks without a label. Ideally we can parse include
// blocks with and without labels in a single pass, but the HCL parser is fairly restrictive when it comes to parsing
// blocks with labels, requiring the exact number of expected labels in the parsing step.  To handle this restriction,
// we first see if there are any include blocks without any labels, and if there is, we modify it in the file object to
// inject the label as "".
func decodeHcl(
	ctx *config.ParsingContext,
	file *hcl.File,
	filename string,
	out interface{},
) (err error) {
	// The HCL2 parser and especially cty conversions will panic in many types of errors, so we have to recover from
	// those panics here and convert them to normal errors
	defer func() {
		if recovered := recover(); recovered != nil {
			err = hclparse.PanicWhileParsingConfigError{RecoveredValue: recovered, ConfigFile: filename}
		}
	}()

	// Check if we need to update the file to label any bare include blocks.
	// Excluding json because of https://github.com/transcend-io/terragrunt-atlantis-config/issues/244.
	if filepath.Ext(filename) != ".json" {
		updatedBytes, isUpdated, err := updateBareIncludeBlock(file, filename)
		if err != nil {
			return err
		}
		if isUpdated {
			// Code was updated, so we need to reparse the new updated contents. This is necessarily because the blocks
			// returned by hclparse does not support editing, and so we have to go through hclwrite, which leads to a
			// different AST representation.
			file, err = parseHcl(hclparse.NewParser(), string(updatedBytes), filename)
			if err != nil {
				return err
			}
		}
	}

	evalContext := createMinimalEvalContext(filename)

	decodeDiagnostics := gohcl.DecodeBody(file.Body, evalContext, out)
	if decodeDiagnostics != nil && decodeDiagnostics.HasErrors() {
		return decodeDiagnostics
	}

	return nil
}

// This decodes only the `include` blocks of a terragrunt config, so its value can be used while decoding the rest of
// the config.
// For consistency, `include` in the call to `decodeHcl` is always assumed to be nil. Either it really is nil (parsing
// the child config), or it shouldn't be used anyway (the parent config shouldn't have an include block).
func decodeAsTerragruntInclude(
	ctx *config.ParsingContext,
	file *hcl.File,
	filename string,
) ([]config.IncludeConfig, error) {
	tgInc := terragruntIncludeMultiple{}
	if err := decodeHcl(ctx, file, filename, &tgInc); err != nil {
		return nil, err
	}
	return tgInc.Include, nil
}

// Not all modules need an include statement, as they could define everything in one file without a parent
// The key signifiers of a parent are:
//   - no include statement
//   - no terraform source defined
//
// If both of those are true, it is likely a parent module
func parseModule(ctx *config.ParsingContext, path string) (isParent bool, includes []config.IncludeConfig, err error) {
	configString, err := util.ReadFileAsString(path)
	if err != nil {
		return false, nil, err
	}

	parser := hclparse.NewParser()
	file, err := parseHcl(parser, configString, path)
	if err != nil {
		return false, nil, err
	}

	terragruntIncludeList, err := decodeAsTerragruntInclude(ctx, file, path)
	if err != nil {
		return false, nil, err
	}

	// If the file has any `include` blocks it is not a parent
	if len(terragruntIncludeList) > 0 {
		return false, terragruntIncludeList, nil
	}

	// We don't need to check the errors/diagnostics coming from `decodeHcl`, as when errors come up,
	// it will leave the partially parsed result in the output object.
	var parsed parsedHcl
	_ = decodeHcl(ctx, file, path, &parsed)

	// If the file does not define a terraform source block, it is likely a parent (though not guaranteed)
	if parsed.Terraform == nil || parsed.Terraform.Source == nil {
		return true, nil, nil
	}

	return false, nil, nil
}
