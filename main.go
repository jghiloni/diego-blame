package main

import (
	"os"

	"github.com/cloudfoundry/cli/plugin"

	"github.com/pivotalservices/diego-blame/blameplugin"
)

var (
	Version string = "1.0.0"
)

func main() {
	diegoBlame := &blameplugin.DiegoBlame{
		Version: Version,
		Writer:  os.Stdout,
	}
	plugin.Start(diegoBlame)
}
