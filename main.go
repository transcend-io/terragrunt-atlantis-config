package main

import "github.com/transcend-io/terragrunt-atlantis-config/cmd"

var (
	VERSION = "0.7.0"
)

func main() {
	cmd.Execute(VERSION)
}
