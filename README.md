# Terragrunt Atlantis Config

Generates Atlantis Config for Terragrunt projects"

## What is this?

Atlantis (runatlantis.io) is an awesome tool for automating terraform pull request automation. Each repo can define a yaml config file that specifies which terraform modules depend on other modules so that PRs that affect dependent modules will also make `terraform plan`s appear for the depending modules.

Bring in terragrunt (https://terragrunt.gruntwork.io/), a thin terraform wrapper that has a concept of dependencies built into its configuration.

This tool creates yaml for terragrunt projects by:

- Finding all `terragrunt.hcl` in a repo
- Evaluating their dependency and terraform source blocks to find their dependencies
- Creating a Directed Acyclic Graph of all dependencies
- Constructing and logging yaml in Atlantis' config spec that reflects the graph

This is especially useful for companies that use monorepos for their terragrunt config (like we do at Transcend-io), and have thousands of lines of config.

## Installation and Usage

Install via brew:

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

Check the log output for the yaml