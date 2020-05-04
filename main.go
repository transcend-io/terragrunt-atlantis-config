package main

import "github.com/transcend-io/terragrunt-atlantis-config/cmd"

var (
	VERSION = "0.1.2"
)

func main() {
	cmd.Execute(VERSION)
}
