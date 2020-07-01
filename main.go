package main

import "github.com/transcend-io/terragrunt-atlantis-config/cmd"

var (
	VERSION = "0.4.3"
)

func main() {
	cmd.Execute(VERSION)
}
