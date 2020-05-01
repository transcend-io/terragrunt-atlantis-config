package main

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"os"

	"github.com/transcend-io/terragrunt-atlantis-config/cmd"
)

func main() {
	os.Mkdir("test_artifacts", os.ModePerm)

	content, err := cmd.RunWithFlags([]string{})
	if err != nil {
		log.Fatal(err)
	}

	goldenContents, err := ioutil.ReadFile("atlantis.yaml")
	if err != nil {
		log.Fatal(err)
	}

	if string(content) != string(goldenContents) {
		log.Fatalf("Expected content did not match golden file. Expected content: %s", string(content))
	}

	log.Info("atlantis config matches :)")
}
