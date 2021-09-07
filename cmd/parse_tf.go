package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/gruntwork-io/terragrunt/util"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

var localModuleSourcePrefixes = []string{
	"./",
	"../",
	".\\",
	"..\\",
}

type ModuleCall struct {
	Name   string `hcl:"name,label"`
	Source string `hcl:"source,attr"`
}

type File struct {
	ModuleCalls []ModuleCall `hcl:"module,block"`
}

type parser struct {
	p *hclparse.Parser
}

func parseTerraformLocalModuleSource(path string) ([]string, error) {
	parser := &parser{
		p: hclparse.NewParser(),
	}

	modules, diags := parser.loadConfigDir(path)
	if diags.HasErrors() {
		return nil, errors.New(diags.Error())
	}

	var sourceMap = map[string]bool{}
	for _, source := range modules {
		if isLocalTerraformModuleSource(source) {
			modulePath := util.JoinPath(path, source)
			modulePathGlob := util.JoinPath(modulePath, "*.tf*")

			if _, exists := sourceMap[modulePathGlob]; exists {
				continue
			}
			sourceMap[modulePathGlob] = true

			// find local module source recursively
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

func (p *parser) loadConfigDir(path string) (map[string]string, hcl.Diagnostics) {
	primaryPaths, overridePaths, diags := p.dirFiles(path)
	if diags.HasErrors() {
		return nil, diags
	}

	primary, fDiags := p.loadFiles(primaryPaths, false)
	diags = append(diags, fDiags...)
	override, fDiags := p.loadFiles(overridePaths, true)
	diags = append(diags, fDiags...)

	module := p.mergeModules(primary, override)

	return module, diags
}

func (p *parser) mergeModules(primaryFiles, overrideFiles []*File) map[string]string {
	modules := map[string]string{}

	for _, f := range primaryFiles {
		for _, m := range f.ModuleCalls {
			modules[m.Name] = m.Source
		}
	}

	for _, f := range overrideFiles {
		for _, m := range f.ModuleCalls {
			modules[m.Name] = m.Source
		}
	}

	return modules
}

func (p *parser) dirFiles(dir string) (primary, override []string, diags hcl.Diagnostics) {
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed to read module directory",
			Detail:   fmt.Sprintf("Module directory %s does not exist or cannot be read.", dir),
		})
		return
	}

	for _, info := range infos {
		if info.IsDir() {
			// We only care about files
			continue
		}

		name := info.Name()
		ext := fileExt(name)
		if ext == "" || isIgnoredFile(name) {
			continue
		}

		baseName := name[:len(name)-len(ext)] // strip extension
		isOverride := baseName == "override" || strings.HasSuffix(baseName, "_override")

		fullPath := filepath.Join(dir, name)
		if isOverride {
			override = append(override, fullPath)
		} else {
			primary = append(primary, fullPath)
		}
	}

	return
}

func (p *parser) loadFiles(paths []string, override bool) ([]*File, hcl.Diagnostics) {
	var files []*File
	var diags hcl.Diagnostics

	for _, path := range paths {
		var f *File
		var fDiags hcl.Diagnostics
		if override {
			f, fDiags = p.loadConfigFile(path, true)
		} else {
			f, fDiags = p.loadConfigFile(path, false)
		}

		diags = append(diags, fDiags...)
		if f != nil {
			files = append(files, f)
		}
	}

	return files, diags
}

func (p *parser) loadConfigFile(path string, override bool) (*File, hcl.Diagnostics) {

	body, diags := p.loadHCLFile(path)
	if body == nil {
		return nil, diags
	}

	file := &File{}
	gohcl.DecodeBody(body, nil, file)

	return file, diags
}

func (p *parser) loadHCLFile(path string) (hcl.Body, hcl.Diagnostics) {
	src, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to read file",
				Detail:   fmt.Sprintf("The file %q could not be read.", path),
			},
		}
	}

	var file *hcl.File
	var diags hcl.Diagnostics
	switch {
	case strings.HasSuffix(path, ".json"):
		file, diags = p.p.ParseJSON(src, path)
	default:
		file, diags = p.p.ParseHCL(src, path)
	}

	// If the returned file or body is nil, then we'll return a non-nil empty
	// body so we'll meet our contract that nil means an error reading the file.
	if file == nil || file.Body == nil {
		return hcl.EmptyBody(), diags
	}

	return file.Body, diags
}

// fileExt returns the Terraform configuration extension of the given
// path, or a blank string if it is not a recognized extension.
func fileExt(path string) string {
	if strings.HasSuffix(path, ".tf") {
		return ".tf"
	} else if strings.HasSuffix(path, ".tf.json") {
		return ".tf.json"
	} else {
		return ""
	}
}

// IsIgnoredFile returns true if the given filename (which must not have a
// directory path ahead of it) should be ignored as e.g. an editor swap file.
func isIgnoredFile(name string) bool {
	return strings.HasPrefix(name, ".") || // Unix-like hidden files
		strings.HasSuffix(name, "~") || // vim
		strings.HasPrefix(name, "#") && strings.HasSuffix(name, "#") // emacs
}
