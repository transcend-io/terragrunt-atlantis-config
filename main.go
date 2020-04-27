package main

import "github.com/transcend-io/terragrunt-atlantis-config/cmd"

var (
	VERSION = "0.0.5"
)

func main() {
	cmd.Execute(VERSION)
}
