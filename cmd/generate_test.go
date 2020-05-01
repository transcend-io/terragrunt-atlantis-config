package cmd

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
)

// Runs a set of arguments, returning the output
func RunWithFlags(args []string) ([]byte, error) {
	randomInt := rand.Int()
	filename := fmt.Sprintf("test_artifacts/%d.yaml", randomInt)

	defer os.Remove(filename)

	allArgs := append([]string{
		"generate",
		"--output",
		filename,
	}, args...)
	rootCmd.SetArgs(allArgs)
	rootCmd.Execute()

	return ioutil.ReadFile(filename)
}

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
	runTest(t, "golden/settingRoot.yaml", []string{
		"--root",
		"..",
	})
}

func TestWithoutTrailingSlash(t *testing.T) {
	runTest(t, "golden/settingRoot.yaml", []string{
		"--root",
		"../../terragrunt-atlantis-config",
	})
}

func TestWithTrailingSlash(t *testing.T) {
	runTest(t, "golden/settingRoot.yaml", []string{
		"--root",
		"../../terragrunt-atlantis-config/",
	})
}

func TestWithNoTerragruntFiles(t *testing.T) {
	runTest(t, "golden/empty.yaml", []string{
		"--root",
		".", // There are no terragrunt files in this directory
	})
}

func TestIgnoringParentTerragrunt(t *testing.T) {
	runTest(t, "golden/withoutParent.yaml", []string{
		"--root",
		"..",
		"--ignore-parent-terragrunt",
	})
}

func TestEnablingAutoplan(t *testing.T) {
	runTest(t, "golden/withAutoplan.yaml", []string{
		"--root",
		"..",
		"--autoplan",
	})
}

func TestSettingWorkflowName(t *testing.T) {
	runTest(t, "golden/namedWorkflow.yaml", []string{
		"--root",
		"..",
		"--workflow",
		"someWorkflow",
	})
}
