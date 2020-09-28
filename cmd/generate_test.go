package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// Resets all flag values to their defaults in between tests
func resetDefaultFlags() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	gitRoot = pwd
	autoPlan = false
	ignoreParentTerragrunt = false
	parallel = true
	createWorkspace = false
	createProjectName = false
	defaultWorkflow = ""
	outputPath = ""

	return nil
}

// Runs a test, asserting the output produced matches a golden file
func runTest(t *testing.T, goldenFile string, args []string) {
	err := resetDefaultFlags()
	if err != nil {
		t.Error("Failed to reset default flags")
		return
	}

	content, err := RunWithFlags(args)
	if err != nil {
		t.Error("Failed to read file")
		return
	}

	goldenContents, err := ioutil.ReadFile(goldenFile)
	if err != nil {
		t.Error("Failed to read golden file")
		return
	}

	if string(content) != string(goldenContents) {
		t.Errorf("Content did not match golden file.\n\nExpected Content: %s\n\nContent: %s", string(goldenContents), string(content))
	}
}

func TestSettingRoot(t *testing.T) {
	runTest(t, filepath.Join("golden", "basic.yaml"), []string{
		"--root",
		filepath.Join("..", "test_examples", "basic_module"),
	})
}

func TestRootPathBeingAbsolute(t *testing.T) {
	parent, err := filepath.Abs(filepath.Join("..", "test_examples", "basic_module"))
	if err != nil {
		t.Error("Failed to find parent directory")
	}

	runTest(t, filepath.Join("golden", "basic.yaml"), []string{
		"--root",
		parent,
	})
}

func TestRootPathHavingTrailingSlash(t *testing.T) {
	runTest(t, filepath.Join("golden", "basic.yaml"), []string{
		"--root",
		filepath.Join("..", "test_examples", "basic_module") + string(filepath.Separator),
	})
}

func TestWithNoTerragruntFiles(t *testing.T) {
	runTest(t, filepath.Join("golden", "empty.yaml"), []string{
		"--root",
		".", // There are no terragrunt files in this directory
		filepath.Join("..", "test_examples", "no_modules"),
	})
}

func TestWithParallelizationDisabled(t *testing.T) {
	runTest(t, filepath.Join("golden", "noParallel.yaml"), []string{
		"--root",
		filepath.Join("..", "test_examples", "basic_module"),
		"--parallel=false",
	})
}

func TestIgnoringParentTerragrunt(t *testing.T) {
	runTest(t, filepath.Join("golden", "withoutParent.yaml"), []string{
		"--root",
		filepath.Join("..", "test_examples", "with_parent"),
		"--ignore-parent-terragrunt",
	})
}

func TestNotIgnoringParentTerragrunt(t *testing.T) {
	runTest(t, filepath.Join("golden", "withParent.yaml"), []string{
		"--root",
		filepath.Join("..", "test_examples", "with_parent"),
	})
}

func TestEnablingAutoplan(t *testing.T) {
	runTest(t, filepath.Join("golden", "withAutoplan.yaml"), []string{
		"--root",
		filepath.Join("..", "test_examples", "basic_module"),
		"--autoplan",
	})
}

func TestSettingWorkflowName(t *testing.T) {
	runTest(t, filepath.Join("golden", "namedWorkflow.yaml"), []string{
		"--root",
		filepath.Join("..", "test_examples", "basic_module"),
		"--workflow",
		"someWorkflow",
	})
}

func TestExtraDeclaredDependencies(t *testing.T) {
	runTest(t, filepath.Join("golden", "extra_dependencies.yaml"), []string{
		"--root",
		filepath.Join("..", "test_examples", "extra_dependency"),
	})
}

func TestLocalTerraformModuleSource(t *testing.T) {
	runTest(t, filepath.Join("golden", "local_terraform_module.yaml"), []string{
		"--root",
		filepath.Join("..", "test_examples", "local_terraform_module_source"),
	})
}

func TestTerragruntDependencies(t *testing.T) {
	runTest(t, filepath.Join("golden", "terragrunt_dependency.yaml"), []string{
		"--root",
		filepath.Join("..", "test_examples", "terragrunt_dependency"),
	})
}

func TestCustomWorkflowName(t *testing.T) {
	runTest(t, filepath.Join("golden", "different_workflow_names.yaml"), []string{
		"--root",
		filepath.Join("..", "test_examples", "different_workflow_names"),
	})
}

// This test covers parent Terragrunt files that are not runnable as modules themselves.
// Sometimes it is possible to have parent files that only are runnable when included
// into child modules.
func TestUnparseableParent(t *testing.T) {
	runTest(t, filepath.Join("golden", "invalid_parent_module.yaml"), []string{
		"--root",
		filepath.Join("..", "test_examples", "invalid_parent_module"),
		"--ignore-parent-terragrunt",
	})
}

func TestWithWorkspaces(t *testing.T) {
	runTest(t, filepath.Join("golden", "withWorkspace.yaml"), []string{
		"--root",
		filepath.Join("..", "test_examples", "basic_module"),
		"--create-workspace",
	})
}

func TestWithProjectNames(t *testing.T) {
	runTest(t, filepath.Join("golden", "withProjectName.yaml"), []string{
		"--root",
		filepath.Join("..", "test_examples", "invalid_parent_module"),
		"--ignore-parent-terragrunt", "--create-project-name",
	})
}

func TestMergingLocalDependenciesFromParent(t *testing.T) {
	runTest(t, filepath.Join("golden", "mergeParentDependencies.yaml"), []string{
		"--root",
		filepath.Join("..", "test_examples", "parent_with_extra_deps"),
		"--ignore-parent-terragrunt",
	})
}

func TestWorkflowFromParentInLocals(t *testing.T) {
	runTest(t, filepath.Join("golden", "parentDefinedWorkflow.yaml"), []string{
		"--root",
		filepath.Join("..", "test_examples", "parent_with_workflow_local"),
		"--ignore-parent-terragrunt",
	})
}

func TestChildWorkflowOverridesParentWorkflow(t *testing.T) {
	runTest(t, filepath.Join("golden", "parentAndChildDefinedWorkflow.yaml"), []string{
		"--root",
		filepath.Join("..", "test_examples", "child_and_parent_specify_workflow"),
		"--ignore-parent-terragrunt",
	})
}
