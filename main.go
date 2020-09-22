package main

import "github.com/transcend-io/terragrunt-atlantis-config/cmd"

// This variable is set at build time using -ldflags parameters. For more info, see:
// http://stackoverflow.com/a/11355611/483528
var VERSION string

func main() {
	cmd.Execute(VERSION)
}
