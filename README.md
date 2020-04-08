# Terragrunt Atlantis Config

Generate Atlantis Config for Terragrunt projects.

## What is this

[Atlantis](runatlantis.io) is an awesome tool for Terraform pull request automation. Each repo can have a YAML config file that defines Terraform module dependendcies, so that PRs that affect dependent modules will automatically generate `terraform plan`s for those modules.

[Terragrunt](https://terragrunt.gruntwork.io) is a Terraform wrapper, which has the concept of dependencies built in to its configuration.

This tool creates YAML configurations for Terragrunt projects by:

- Finding all `terragrunt.hcl` in a repo
- Evaluating their dependency and Terraform source blocks to find their dependencies
- Creating a Directed Acyclic Graph of all dependencies
- Constructing and logging YAML in Atlantis' config spec that reflects the graph

This is especially useful for organizations that use monorepos for their Terragrunt config (as we do at Transcend), and have thousands of lines of config.

## Installation and Usage

Install via Homebrew:

```bash
brew install transcend-io/tap/terragrunt-atlantis-config
```

Usage:

```bash
# From the root of your repo
terragrunt-atlantis-config generate

# or from anywhere
terragrunt-atlantis-config generate --root /some/path/to/your/repo/root
```

Finally, check the log output for the YAML.
