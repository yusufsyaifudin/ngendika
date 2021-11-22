package main

import (
	"log"
	"os"

	_ "github.com/lib/pq"

	"github.com/mitchellh/cli"
	"github.com/yusufsyaifudin/ngendika/cmd/api"
	"github.com/yusufsyaifudin/ngendika/cmd/worker"
)

func main() {
	apiCmd := api.NewCmd()

	c := cli.NewCLI("ngendika", "1.0.0")
	c.Args = os.Args[1:]
	c.Autocomplete = true
	c.Commands = map[string]cli.CommandFactory{
		"":       apiCmd, // default command if no subcommand defined
		"api":    apiCmd,
		"worker": worker.NewCmd(),
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
