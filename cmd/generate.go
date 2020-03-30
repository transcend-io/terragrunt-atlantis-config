/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"

	"github.com/gruntwork-io/terragrunt/cli"
	"github.com/gruntwork-io/terragrunt/config"
	"github.com/gruntwork-io/terragrunt/options"

	"github.com/ghodss/yaml"

	"os"
	"path/filepath"
	"strings"
)

// Parse env vars into a map
func getEnvs() map[string]string {
	envs := os.Environ()
	m := make(map[string]string)

	for _, env := range envs {
		results := strings.Split(env, "=")
		m[results[0]] = results[1]
	}

	return m
}

// Represents an entire config file
type AtlantisConfig struct {
	// Version of the config syntax
	Version int `json:"version"`

	// If Atlantis should merge after finishing `atlantis apply`
	AutoMerge bool `json:"automerge"`

	// The project settings
	Projects []AtlantisProject `json:"projects"`
}

// Represents an Atlantis Project directory
type AtlantisProject struct {
	// The directory with the terragrunt.hcl file
	Dir string `json:"dir"`

	// Autoplan settings for which plans affect other plans
	Autoplan AutoplanConfig `json:"autoplan"`
}

// Autoplan settings for which plans affect other plans
type AutoplanConfig struct {
	// Relative paths from this modules directory to modules it depends on
	WhenModified []string `json:"when_modified"`

	// If autoplan should be enabled for this dir
	Enabled bool `json:"enabled"`
}

// Terragrunt imports can be relative or absolute
// This makes relative paths absolute
func makePathAbsolute(path string, parentPath string) string {
	if strings.HasPrefix(path, gitRoot) {
		return path
	}

	parentDir := filepath.Dir(parentPath)
	return filepath.Join(parentDir, path)
}

// Parses the terragrunt config at <path> to find all modules it depends on
func getDependencies(path string) ([]string, error) {
	decodeTypes := []config.PartialDecodeSectionType{
		config.DependencyBlock,
		config.DependenciesBlock,
		config.TerraformBlock,
	}

	options, err := options.NewTerragruntOptions(path)
	if err != nil {
		return nil, err
	}
	options.RunTerragrunt = cli.RunTerragrunt
	options.Env = getEnvs()

	parsedConfig, err := config.PartialParseConfigFile(path, options, nil, decodeTypes)
	if err != nil {
		return nil, err
	}

	dependencies := []string{}

	if parsedConfig.Dependencies != nil {
		dependencies = parsedConfig.Dependencies.Paths
	}

	if parsedConfig.Terraform != nil && parsedConfig.Terraform.Source != nil {
		source := parsedConfig.Terraform.Source
		// TODO: Make more robust. Check for bitbucket, etc.
		if !strings.Contains(*source, "git") {
			dependencies = append(dependencies, *source)
		}
	}

	return dependencies, nil
}

// Creates an AtlantisProject for a directory
func createProject(sourcePath string) (*AtlantisProject, error) {
	dependencies, err := getDependencies(sourcePath)
	if err != nil {
		return nil, err
	}

	absoluteSourceDir := filepath.Dir(sourcePath)

	// All dependencies depend on their own .hcl file, and any tf files in their directory
	relativeDependencies := []string{
		"terragrunt.hcl",
		"*.tf*",
	}

	// Add other dependencies based on their relative paths
	for _, dependencyPath := range dependencies {
		absolutePath := makePathAbsolute(dependencyPath, sourcePath)
		relativePath, err := filepath.Rel(absoluteSourceDir, absolutePath)
		if err != nil {
			return nil, err
		}
		relativeDependencies = append(relativeDependencies, relativePath)
	}

	relativeSourceDir := strings.TrimPrefix(absoluteSourceDir, gitRoot)

	project := &AtlantisProject{
		Dir: relativeSourceDir,

		Autoplan: AutoplanConfig{
			Enabled:      false,
			WhenModified: relativeDependencies,
		},
	}
	return project, nil
}

// Finds the absolute paths of all terragrunt.hcl files
func getAllTerragruntFiles() ([]string, error) {
	options, err := options.NewTerragruntOptions(gitRoot)
	if err != nil {
		return nil, err
	}

	paths, err := config.FindConfigFilesInPath(gitRoot, options)
	if err != nil {
		return nil, err
	}

	return paths, nil
}

// Logs out an Atlantis repo yaml config file contents.
// Limitations:
//   - Only goes one level deep
//   - Does not work with `read_terragrunt_config` dependencies
//   - Some atlantis env vars are not respected (would need to use their CLI context)
//   - Poorly handles paths
func main(cmd *cobra.Command, args []string) {
	terragruntFiles, err := getAllTerragruntFiles()
	if err != nil {
		log.Fatal("Could not list all terragrunt files: ", err)
	}
	log.Info("AHHHHHHH")
	log.Info(terragruntFiles)

	config := AtlantisConfig{
		Version:   3,
		AutoMerge: false,
	}

	for _, terragruntPath := range terragruntFiles {
		project, err := createProject(terragruntPath)
		if err != nil {
			log.Fatal("Could not create project for ", terragruntPath, " with err: ", err)
		}
		log.Info("Created project for ", terragruntPath)
		config.Projects = append(config.Projects, *project)
	}

	yamlString, err := yaml.Marshal(&config)
	if err != nil {
		log.Fatal("Could not serialize Config into yaml")
	}
	log.Println(string(yamlString))
}

var gitRoot string

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Makes atlantis config",
	Long:  `Logs Yaml representing Atlantis config to stderr`,
	Run:   main,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.PersistentFlags().StringVar(&gitRoot, "root", ".", "Path to the root directory of the github repo you want to build config for. Default is current dir")
}
