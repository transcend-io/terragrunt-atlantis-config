package cmd

import (
	"regexp"
	"sort"

	log "github.com/sirupsen/logrus"

	"github.com/ghodss/yaml"
	"github.com/gruntwork-io/terragrunt/cli"
	"github.com/gruntwork-io/terragrunt/config"
	"github.com/gruntwork-io/terragrunt/options"
	"github.com/gruntwork-io/terragrunt/util"
	"github.com/spf13/cobra"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"golang.org/x/sync/singleflight"

	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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

// Terragrunt imports can be relative or absolute
// This makes relative paths absolute
func makePathAbsolute(path string, parentPath string) string {
	if strings.HasPrefix(path, filepath.ToSlash(gitRoot)) {
		return path
	}

	parentDir := filepath.Dir(parentPath)
	return filepath.Join(parentDir, path)
}

var requestGroup singleflight.Group

// Set up a cache for the getDependencies function
type getDependenciesOutput struct {
	dependencies []string
	err          error
}

type GetDependenciesCache struct {
	mtx  sync.RWMutex
	data map[string]getDependenciesOutput
}

func newGetDependenciesCache() *GetDependenciesCache {
	return &GetDependenciesCache{data: map[string]getDependenciesOutput{}}
}

func (m *GetDependenciesCache) set(k string, v getDependenciesOutput) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.data[k] = v
}

func (m *GetDependenciesCache) get(k string) (getDependenciesOutput, bool) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	v, ok := m.data[k]
	return v, ok
}

var getDependenciesCache = newGetDependenciesCache()

// Parses the terragrunt config at `path` to find all modules it depends on
func getDependencies(path string, terragruntOptions *options.TerragruntOptions) ([]string, error) {
	res, err, _ := requestGroup.Do(path, func() (interface{}, error) {
		// Check if this path has already been computed
		cachedResult, ok := getDependenciesCache.get(path)
		if ok {
			return cachedResult.dependencies, cachedResult.err
		}
		// if theres no terraform source and we're ignoring parent terragrunt configs
		// return nils to indicate we should skip this project
		isParent, err := isParentModule(path, terragruntOptions)
		if err != nil {
			getDependenciesCache.set(path, getDependenciesOutput{nil, err})
			return nil, err
		}
		if ignoreParentTerragrunt && isParent {
			getDependenciesCache.set(path, getDependenciesOutput{nil, nil})
			return nil, nil
		}

		// Parse the HCL file
		decodeTypes := []config.PartialDecodeSectionType{
			config.DependencyBlock,
			config.DependenciesBlock,
			config.TerraformBlock,
		}
		parsedConfig, err := config.PartialParseConfigFile(path, terragruntOptions, nil, decodeTypes)
		if err != nil {
			getDependenciesCache.set(path, getDependenciesOutput{nil, err})
			return nil, err
		}

		// Parse out locals
		locals, err := parseLocals(path, terragruntOptions, nil)
		if err != nil {
			getDependenciesCache.set(path, getDependenciesOutput{nil, err})
			return nil, err
		}

		// Get deps from locals
		dependencies := []string{}
		if locals.ExtraAtlantisDependencies != nil {
			dependencies = locals.ExtraAtlantisDependencies
		}

		// Get deps from `dependencies` and `dependency` blocks
		if parsedConfig.Dependencies != nil && !ignoreDependencyBlocks {
			for _, parsedPaths := range parsedConfig.Dependencies.Paths {
				dependencies = append(dependencies, filepath.Join(parsedPaths, "terragrunt.hcl"))
			}
		}

		// Get deps from the `Source` field of the `Terraform` block
		if parsedConfig.Terraform != nil && parsedConfig.Terraform.Source != nil {
			source := parsedConfig.Terraform.Source
			// TODO: Make more robust. Check for bitbucket, etc.
			if !strings.Contains(*source, "git::") && !strings.Contains(*source, "github.com") && !strings.Contains(*source, "tfr:///") {
				dependencies = append(dependencies, filepath.Join(*source, "*.tf*"))

				var dir string

				if filepath.IsAbs(*source) {
					dir = *source
				} else {
					dir = util.JoinPath(filepath.Dir(path), *source)
				}
				ls, err := parseTerraformLocalModuleSource(dir)
				if err != nil {
					return nil, err
				}
				sort.Strings(ls)

				dependencies = append(dependencies, ls...)
			}
		}

		// Get deps from `extra_arguments` fields of the `Terraform` block
		if parsedConfig.Terraform != nil && parsedConfig.Terraform.ExtraArgs != nil {
			extraArgs := parsedConfig.Terraform.ExtraArgs
			for _, arg := range extraArgs {
				if arg.RequiredVarFiles != nil {
					dependencies = append(dependencies, *arg.RequiredVarFiles...)
				}
				if arg.OptionalVarFiles != nil {
					dependencies = append(dependencies, *arg.OptionalVarFiles...)
				}
				if arg.Arguments != nil {
					for _, cliFlag := range *arg.Arguments {
						if strings.HasPrefix(cliFlag, "-var-file=") {
							dependencies = append(dependencies, strings.TrimPrefix(cliFlag, "-var-file="))
						}
					}
				}
			}
		}

		// Filter out and dependencies that are the empty string
		nonEmptyDeps := []string{}
		for _, dep := range dependencies {
			if dep != "" {
				childDepAbsPath := dep
				if !filepath.IsAbs(childDepAbsPath) {
					childDepAbsPath = makePathAbsolute(dep, path)
				}
				childDepAbsPath = filepath.ToSlash(childDepAbsPath)
				nonEmptyDeps = append(nonEmptyDeps, childDepAbsPath)
			}
		}

		// Recurse to find dependencies of all dependencies
		cascadedDeps := []string{}
		for _, dep := range nonEmptyDeps {
			cascadedDeps = append(cascadedDeps, dep)

			// The "cascading" feature is protected by a flag
			if !cascadeDependencies {
				continue
			}

			depPath := dep
			terrOpts, _ := options.NewTerragruntOptions(depPath)
			terrOpts.OriginalTerragruntConfigPath = terragruntOptions.OriginalTerragruntConfigPath
			childDeps, err := getDependencies(depPath, terrOpts)
			if err != nil {
				continue
			}

			for _, childDep := range childDeps {
				// If `childDep` is a relative path, it will be relative to `childDep`, as it is from the nested
				// `getDependencies` call on the top level module's dependencies. So here we update any relative
				// path to be from the top level module instead.
				childDepAbsPath := childDep
				if !filepath.IsAbs(childDep) {
					childDepAbsPath, err = filepath.Abs(filepath.Join(depPath, "..", childDep))
					if err != nil {
						getDependenciesCache.set(path, getDependenciesOutput{nil, err})
						return nil, err
					}
				}
				childDepAbsPath = filepath.ToSlash(childDepAbsPath)

				// Ensure we are not adding a duplicate dependency
				alreadyExists := false
				for _, dep := range cascadedDeps {
					if dep == childDepAbsPath {
						alreadyExists = true
						break
					}
				}
				if !alreadyExists {
					cascadedDeps = append(cascadedDeps, childDepAbsPath)
				}
			}
		}

		if filepath.Base(path) == "terragrunt.hcl" {
			dir := filepath.Dir(path)

			ls, err := parseTerraformLocalModuleSource(dir)
			if err != nil {
				return nil, err
			}
			sort.Strings(ls)

			cascadedDeps = append(cascadedDeps, ls...)
		}

		getDependenciesCache.set(path, getDependenciesOutput{cascadedDeps, err})
		return cascadedDeps, nil
	})
	if res != nil {
		return res.([]string), err
	} else {
		return nil, err
	}

}

// Creates an AtlantisProject for a directory
func createProject(sourcePath string) (*AtlantisProject, error) {
	options, err := options.NewTerragruntOptions(sourcePath)
	if err != nil {
		return nil, err
	}
	options.OriginalTerragruntConfigPath = sourcePath
	options.RunTerragrunt = cli.RunTerragrunt
	options.Env = getEnvs()

	dependencies, err := getDependencies(sourcePath, options)
	if err != nil {
		return nil, err
	}
	// dependencies being nil is a sign from `getDependencies` that this project should be skipped
	if dependencies == nil {
		return nil, nil
	}

	absoluteSourceDir := filepath.Dir(sourcePath) + string(filepath.Separator)

	locals, err := parseLocals(sourcePath, options, nil)
	if err != nil {
		return nil, err
	}

	// If `atlantis_skip` is true on the module, then do not produce a project for it
	if locals.Skip != nil && *locals.Skip {
		return nil, nil
	}

	// All dependencies depend on their own .hcl file, and any tf files in their directory
	relativeDependencies := []string{
		"*.hcl",
		"*.tf*",
	}

	// Add other dependencies based on their relative paths. We always want to output with Unix path separators
	for _, dependencyPath := range dependencies {
		absolutePath := dependencyPath
		if !filepath.IsAbs(absolutePath) {
			absolutePath = makePathAbsolute(dependencyPath, sourcePath)
		}
		relativePath, err := filepath.Rel(absoluteSourceDir, absolutePath)
		if err != nil {
			return nil, err
		}

		relativeDependencies = append(relativeDependencies, filepath.ToSlash(relativePath))
	}

	// Clean up the relative path to the format Atlantis expects
	relativeSourceDir := strings.TrimPrefix(absoluteSourceDir, gitRoot)
	relativeSourceDir = strings.TrimSuffix(relativeSourceDir, string(filepath.Separator))
	if relativeSourceDir == "" {
		relativeSourceDir = "."
	}

	workflow := defaultWorkflow
	if locals.AtlantisWorkflow != "" {
		workflow = locals.AtlantisWorkflow
	}

	applyRequirements := &defaultApplyRequirements
	if len(defaultApplyRequirements) == 0 {
		applyRequirements = nil
	}
	if locals.ApplyRequirements != nil {
		applyRequirements = &locals.ApplyRequirements
	}

	resolvedAutoPlan := autoPlan
	if locals.AutoPlan != nil {
		resolvedAutoPlan = *locals.AutoPlan
	}

	terraformVersion := defaultTerraformVersion
	if locals.TerraformVersion != "" {
		terraformVersion = locals.TerraformVersion
	}

	project := &AtlantisProject{
		Dir:               filepath.ToSlash(relativeSourceDir),
		Workflow:          workflow,
		TerraformVersion:  terraformVersion,
		ApplyRequirements: applyRequirements,
		Autoplan: AutoplanConfig{
			Enabled:      resolvedAutoPlan,
			WhenModified: relativeDependencies,
		},
	}

	// Terraform Cloud limits the workspace names to be less than 90 characters
	// with letters, numbers, -, and _
	// https://www.terraform.io/docs/cloud/workspaces/naming.html
	// It is not clear from documentation whether the normal workspaces have those limitations
	// However a workspace 97 chars long has been working perfectly.
	// We are going to use the same name for both workspace & project name as it is unique.
	regex := regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	projectName := regex.ReplaceAllString(project.Dir, "_")

	if createProjectName {
		project.Name = projectName
	}

	if createWorkspace {
		project.Workspace = projectName
	}

	return project, nil
}

// Finds the absolute paths of all terragrunt.hcl files
func getAllTerragruntFiles() ([]string, error) {
	options, err := options.NewTerragruntOptions(gitRoot)
	if err != nil {
		return nil, err
	}

	// If filterPath is provided, override workingPath instead of gitRoot
	// We do this here because we want to keep the relative path structure of Terragrunt files
	// to root and just ignore the ConfigFiles
	workingPaths := []string{gitRoot}
	if filterPath != "" {
		// get all matching folders
		workingPaths, err = filepath.Glob(filterPath)
		if err != nil {
			return nil, err
		}
	}

	uniqueConfigFilePaths := make(map[string]bool)
	orderedConfigFilePaths := []string{}
	for _, workingPath := range workingPaths {
		paths, err := config.FindConfigFilesInPath(workingPath, options)
		if err != nil {
			return nil, err
		}
		for _, p := range paths {
			// if path not yet seen, insert once
			if !uniqueConfigFilePaths[p] {
				orderedConfigFilePaths = append(orderedConfigFilePaths, p)
				uniqueConfigFilePaths[p] = true
			}
		}
	}

	uniqueConfigFileAbsPaths := []string{}
	for _, uniquePath := range orderedConfigFilePaths {
		uniqueAbsPath, err := filepath.Abs(uniquePath)
		if err != nil {
			return nil, err
		}
		uniqueConfigFileAbsPaths = append(uniqueConfigFileAbsPaths, uniqueAbsPath)
	}

	return uniqueConfigFileAbsPaths, nil
}

func main(cmd *cobra.Command, args []string) error {
	// Ensure the gitRoot has a trailing slash and is an absolute path
	absoluteGitRoot, err := filepath.Abs(gitRoot)
	if err != nil {
		return err
	}
	gitRoot = absoluteGitRoot + string(filepath.Separator)

	// Read in the old config, if it already exists
	oldConfig, err := readOldConfig()
	if err != nil {
		return err
	}

	terragruntFiles, err := getAllTerragruntFiles()
	if err != nil {
		return err
	}

	config := AtlantisConfig{
		Version:       3,
		AutoMerge:     autoMerge,
		ParallelPlan:  parallel,
		ParallelApply: parallel,
	}
	if oldConfig != nil && preserveWorkflows {
		config.Workflows = oldConfig.Workflows
	}

	lock := sync.Mutex{}
	ctx := context.Background()
	errGroup, _ := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(numExecutors)

	// Concurrently looking all dependencies
	for _, terragruntPath := range terragruntFiles {
		terragruntPath := terragruntPath // https://golang.org/doc/faq#closures_and_goroutines

		err := sem.Acquire(ctx, 1)
		if err != nil {
			return err
		}

		errGroup.Go(func() error {
			defer sem.Release(1)
			project, err := createProject(terragruntPath)
			if err != nil {
				return err
			}
			// if project and err are nil then skip this project
			if err == nil && project == nil {
				return nil
			}

			// Lock the list as only one goroutine should be writing to config.Projects at a time
			lock.Lock()
			defer lock.Unlock()

			log.Info("Created project for ", terragruntPath)
			config.Projects = append(config.Projects, *project)

			return nil
		})
	}

	if err := errGroup.Wait(); err != nil {
		return err
	}

	// Sort the projects in config by Dir
	sort.Slice(config.Projects, func(i, j int) bool { return config.Projects[i].Dir < config.Projects[j].Dir })

	// Convert config to YAML string
	yamlBytes, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	// Ensure newline characters are correct on windows machines, as the json encoding function in the stdlib
	// uses "\n" for all newlines regardless of OS: https://github.com/golang/go/blob/master/src/encoding/json/stream.go#L211-L217
	yamlString := string(yamlBytes)
	if strings.Contains(runtime.GOOS, "windows") {
		yamlString = strings.ReplaceAll(yamlString, "\n", "\r\n")
	}

	// Write output
	if len(outputPath) != 0 {
		ioutil.WriteFile(outputPath, []byte(yamlString), 0644)
	} else {
		log.Println(yamlString)
	}

	return nil
}

var gitRoot string
var autoPlan bool
var autoMerge bool
var ignoreParentTerragrunt bool
var ignoreDependencyBlocks bool
var parallel bool
var createWorkspace bool
var createProjectName bool
var defaultTerraformVersion string
var defaultWorkflow string
var filterPath string
var outputPath string
var preserveWorkflows bool
var cascadeDependencies bool
var defaultApplyRequirements []string
var numExecutors int64

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Makes atlantis config",
	Long:  `Logs Yaml representing Atlantis config to stderr`,
	RunE:  main,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	generateCmd.PersistentFlags().BoolVar(&autoPlan, "autoplan", false, "Enable auto plan. Default is disabled")
	generateCmd.PersistentFlags().BoolVar(&autoMerge, "automerge", false, "Enable auto merge. Default is disabled")
	generateCmd.PersistentFlags().BoolVar(&ignoreParentTerragrunt, "ignore-parent-terragrunt", true, "Ignore parent terragrunt configs (those which don't reference a terraform module). Default is enabled")
	generateCmd.PersistentFlags().BoolVar(&ignoreDependencyBlocks, "ignore-dependency-blocks", false, "When true, dependencies found in `dependency` blocks will be ignored")
	generateCmd.PersistentFlags().BoolVar(&parallel, "parallel", true, "Enables plans and applys to happen in parallel. Default is enabled")
	generateCmd.PersistentFlags().BoolVar(&createWorkspace, "create-workspace", false, "Use different workspace for each project. Default is use default workspace")
	generateCmd.PersistentFlags().BoolVar(&createProjectName, "create-project-name", false, "Add different name for each project. Default is false")
	generateCmd.PersistentFlags().BoolVar(&preserveWorkflows, "preserve-workflows", true, "Preserves workflows from old output files. Default is true")
	generateCmd.PersistentFlags().BoolVar(&cascadeDependencies, "cascade-dependencies", true, "When true, dependencies will cascade, meaning that a module will be declared to depend not only on its dependencies, but all dependencies of its dependencies all the way down. Default is true")
	generateCmd.PersistentFlags().StringVar(&defaultWorkflow, "workflow", "", "Name of the workflow to be customized in the atlantis server. Default is to not set")
	generateCmd.PersistentFlags().StringSliceVar(&defaultApplyRequirements, "apply-requirements", []string{}, "Requirements that must be satisfied before `atlantis apply` can be run. Currently the only supported requirements are `approved` and `mergeable`. Can be overridden by locals")
	generateCmd.PersistentFlags().StringVar(&outputPath, "output", "", "Path of the file where configuration will be generated. Default is not to write to file")
	generateCmd.PersistentFlags().StringVar(&filterPath, "filter", "", "Path or glob expression to the directory you want scope down the config for. Default is all files in root")
	generateCmd.PersistentFlags().StringVar(&gitRoot, "root", pwd, "Path to the root directory of the git repo you want to build config for. Default is current dir")
	generateCmd.PersistentFlags().StringVar(&defaultTerraformVersion, "terraform-version", "", "Default terraform version to specify for all modules. Can be overriden by locals")
	generateCmd.PersistentFlags().Int64Var(&numExecutors, "num-executors", 15, "Number of executors used for parallel generation of projects. Default is 15")
}

// Runs a set of arguments, returning the output
func RunWithFlags(filename string, args []string) ([]byte, error) {
	rootCmd.SetArgs(args)
	rootCmd.Execute()

	return ioutil.ReadFile(filename)
}
