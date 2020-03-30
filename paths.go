package main

import (
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

func main() {
	source := "infra/vault/env-dev/settings/roles/atlantis-role/terragrunt.hcl"
	dest := "infra/atlantis/instance/terragrunt.hcl"

	rel, err := filepath.Rel(source, dest)
	if err != nil {
		log.Fatal("Couldn't find relative path")
	}
	log.Info(rel)
}
