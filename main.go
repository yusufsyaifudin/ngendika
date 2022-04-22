package main

import (
	"log"
	"os"

	_ "github.com/lib/pq"

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
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
