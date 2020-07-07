package main

import "github.com/transcend-io/terragrunt-atlantis-config/cmd"

var (
	VERSION = "0.4.4"
)

func main() {
	cmd.Execute(VERSION)
}
