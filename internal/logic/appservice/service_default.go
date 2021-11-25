package appservice

import (
	"context"
	"fmt"

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
	app := apprepo.NewApp(input.ClientID, input.Name)
	err = validator.New().Struct(app)
	if err != nil {
		err = fmt.Errorf("validation error, missing required field: %w", err)
		return
	}

	app, err = d.conf.AppRepo.CreateApp(ctx, app)
	if err != nil {
		return
	}

	out = CreateAppOut{
		App: App{
			ClientID:  app.ClientID,
			Name:      app.Name,
			Enabled:   app.Enabled,
			CreatedAt: app.CreatedAt.UTC(), // always UTC
		},
	}
	return
}

func (d *DefaultService) GetAppByClientID(ctx context.Context, clientID string) (app App, err error) {
	out, err := d.conf.AppRepo.GetAppByClientID(ctx, clientID)
	if err != nil {
		err = fmt.Errorf("not found app client id '%s': %w", clientID, err)
		return
	}

	app = App{
		ID:        out.ID,
		ClientID:  out.ClientID,
		Name:      out.Name,
		Enabled:   out.Enabled,
		CreatedAt: out.CreatedAt.UTC(),
	}

	return
}
