package main

import "github.com/olivermking/spin-aks-plugin/cmd"

var version string // set through ldflags at build time like "go run -ldflags "-X main.version=0.0.1" main.go"

func main() {
	cmd.Execute(cmd.Config{
		Version: version,
	})
}
