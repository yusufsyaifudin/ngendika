package api

import (
	"context"
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/yusufsyaifudin/ngendika/container"
	"github.com/yusufsyaifudin/ngendika/extd"
	"github.com/yusufsyaifudin/ylog"
	"log"
	"time"
)

const (
	ExitSuccess = 0
	ExitErr     = -1
)

type Cmd struct {
	appName    string
	appVersion string
}

func NewCmd(appName, appVersion string) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		cmd := &Cmd{
			appName:    appName,
			appVersion: appVersion,
		}
		return cmd, nil
	}
}

var _ cli.Command = (*Cmd)(nil)
var _ cli.CommandFactory = NewCmd("", "")

func (c *Cmd) Help() string {
	return `API will start server using HTTP or gRPC`
}

func (c *Cmd) Run(args []string) int {
	// ** define system context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := container.LoadConfig()
	if err != nil {
		err = fmt.Errorf("error load config: %w", err)
		log.Println(err)
		return ExitErr
	}

	// ** register default backends
	err = extd.RegisterDefaultBackends(ctx)
	if err != nil {
		ylog.Error(ctx, "register default backend failed", ylog.KV("error", err))
		return ExitErr
	}

	err = extd.RunServer(ctx, cfg)
	if err != nil {
		ylog.Error(ctx, "cannot start server", ylog.KV("error", err))
		return ExitErr
	}

	return ExitSuccess
}

func (c *Cmd) Synopsis() string {
	return `API will start server using HTTP or gRPC`
}
