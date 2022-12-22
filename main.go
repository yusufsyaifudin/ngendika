package main

import (
	_ "github.com/lib/pq"
	"github.com/yusufsyaifudin/ngendika/cmd/gen/genapidoc"
	"log"
	"os"

	"github.com/mitchellh/cli"
	"github.com/yusufsyaifudin/ngendika/cmd/api"
)

func main() {
	const appName, appVersion = "ngendika", "1.0.0"

	apiCmd := api.NewCmd(appName, appVersion)

	c := cli.NewCLI(appName, appVersion)
	c.Args = os.Args[1:]
	c.Autocomplete = true
	c.Commands = map[string]cli.CommandFactory{
		"":    apiCmd, // default command if no subcommand defined
		"api": apiCmd,
		"apidoc": func() (cli.Command, error) {
			return genapidoc.NewApiDocCmd(genapidoc.ApiDocCfg{})
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
