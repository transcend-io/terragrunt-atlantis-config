package cmd

import (
	"strings"

	"github.com/gruntwork-io/terragrunt/util"
	"github.com/hashicorp/terraform/configs"
)

var localModuleSourcePrefixes = []string{
	"./",
	"../",
	".\\",
	"..\\",
}

func parseTerraformLocalModuleSource(path string) ([]string, error) {
	parser := configs.NewParser(nil)
	module, errors := parser.LoadConfigDir(path)
	if len(errors) > 0 {
		return nil, errors[0]
	}

	var sourceMap = map[string]bool{}
	for _, moduleCall := range module.ModuleCalls {
		if isLocalTerraformModuleSource(moduleCall.SourceAddr) {
			modulePath := util.JoinPath(path, moduleCall.SourceAddr)
			modulePathGlob := util.JoinPath(modulePath, "*.tf*")

			if _, exists := sourceMap[modulePathGlob]; exists {
				continue
			}
			sourceMap[modulePathGlob] = true

			// find recursive local module source
			subSources, err := parseTerraformLocalModuleSource(modulePath)
			if err != nil {
				return nil, err
			}

			for _, subSource := range subSources {
				sourceMap[subSource] = true
			}
		}
	}

	var sources = []string{}
	for source := range sourceMap {
		sources = append(sources, source)
	}

	return sources, nil
}

func isLocalTerraformModuleSource(raw string) bool {
	for _, prefix := range localModuleSourcePrefixes {
		if strings.HasPrefix(raw, prefix) {
			return true
		}
	}

	return false
}
