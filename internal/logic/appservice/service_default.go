package appservice

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yusufsyaifudin/ylog"

	"github.com/go-playground/validator/v10"
	"github.com/yusufsyaifudin/ngendika/internal/storage/apprepo"
)

type DefaultServiceConfig struct {
	AppRepo apprepo.Repo `validate:"required"`
}

type DefaultService struct {
	conf DefaultServiceConfig
}

var _ Service = (*DefaultService)(nil)

func New(dep DefaultServiceConfig) (*DefaultService, error) {
	if err := validator.New().Struct(dep); err != nil {
		return nil, err
	}

	return &DefaultService{
		conf: dep,
	}, nil
}

// CreateApp is a function that knows business logic.
// It doesn't know whether the input come from HTTP or GRPC or any input.
func (d *DefaultService) CreateApp(ctx context.Context, input CreateAppIn) (out CreateAppOut, err error) {
	err = validator.New().Struct(input)
	if err != nil {
		err = fmt.Errorf("validation error, missing required field: %w", err)
		return
	}

	existingApp, err := d.GetAppByClientID(ctx, input.ClientID)
	if err != nil {
		// log and then discard error
		ylog.Error(ctx, "get app by id error, continuing to try to insert", ylog.KV("error", err))
		err = nil
	}

	if existingApp != nil {
		err = fmt.Errorf("app with client id '%s' already exist", existingApp.ClientID)
		return
	}

	app := &apprepo.App{
		ClientID:  strings.ToLower(input.ClientID),
		Name:      input.Name,
		Enabled:   true,
		CreatedAt: time.Now().UTC(),
	}

	app, err = d.conf.AppRepo.CreateApp(ctx, app)
	if err != nil {
		return
	}

	out = CreateAppOut{
		App: AppFromRepo(app),
	}
	return
}

func (d *DefaultService) GetAppByClientID(ctx context.Context, clientID string) (app *App, err error) {
	out, err := d.conf.AppRepo.GetAppByClientID(ctx, clientID)
	if err != nil {
		err = fmt.Errorf("not found app client id '%s': %w", clientID, err)
		return
	}

	app = AppFromRepo(out)
	return
}
