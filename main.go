package main

import "github.com/tufitko/terragrunt-atlantis-config/cmd"

// This variable is set at build time using -ldflags parameters.
// But we still set a default here for those using plain `go get` downloads
// For more info, see: http://stackoverflow.com/a/11355611/483528
var VERSION string = "1.20.0"

func main() {
	cmd.Execute(VERSION)
}
