package main

import (
	"github.com/cloudfoundry/cli/plugin"

	"github.com/pivotalservices/diego-blame/blameplugin"
)

var (
	Version string = "1.0.0"
)

func main() {
	diegoBlame := &blameplugin.DiegoBlame{Version: Version}
	plugin.Start(diegoBlame)
}
