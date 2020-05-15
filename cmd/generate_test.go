package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// Resets all flag values to their defaults
func resetDefaultFlags() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	gitRoot = pwd
	autoPlan = false
	ignoreParentTerragrunt = false
	workflow = ""
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
		t.Errorf("Expected content did not match golden file. Expected content: %s", string(content))
	}
}

func TestSettingRoot(t *testing.T) {
	runTest(t, filepath.Join("golden", "settingRoot.yaml"), []string{
		"--root",
		"..",
	})
}

func TestWithoutTrailingSlash(t *testing.T) {
	parent, err := filepath.Abs("..")
	if err != nil {
		t.Error("Failed to find parent directory")
	}

	runTest(t, filepath.Join("golden", "settingRoot.yaml"), []string{
		"--root",
		parent,
	})
}

func TestWithTrailingSlash(t *testing.T) {
	parent, err := filepath.Abs("..")
	if err != nil {
		t.Error("Failed to find parent directory")
	}

	runTest(t, filepath.Join("golden", "settingRoot.yaml"), []string{
		"--root",
		parent + string(filepath.Separator),
	})
}

func TestWithNoTerragruntFiles(t *testing.T) {
	runTest(t, filepath.Join("golden", "empty.yaml"), []string{
		"--root",
		".", // There are no terragrunt files in this directory
	})
}

func TestIgnoringParentTerragrunt(t *testing.T) {
	runTest(t, filepath.Join("golden", "withoutParent.yaml"), []string{
		"--root",
		"..",
		"--ignore-parent-terragrunt",
	})
}

func TestEnablingAutoplan(t *testing.T) {
	runTest(t, filepath.Join("golden", "withAutoplan.yaml"), []string{
		"--root",
		"..",
		"--autoplan",
	})
}

func TestSettingWorkflowName(t *testing.T) {
	runTest(t, filepath.Join("golden", "namedWorkflow.yaml"), []string{
		"--root",
		"..",
		"--workflow",
		"someWorkflow",
	})
}
