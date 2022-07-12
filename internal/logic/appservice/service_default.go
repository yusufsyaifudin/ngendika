package appservice

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/yusufsyaifudin/ngendika/internal/storage/apprepo"
	"github.com/yusufsyaifudin/ylog"
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
func (d *DefaultService) CreateApp(ctx context.Context, input InputCreateApp) (out OutCreateApp, err error) {
	err = validator.New().Struct(input)
	if err != nil {
		err = fmt.Errorf("validation error, missing required field: %w", err)
		return
	}

	existingApp, err := d.GetApp(ctx, InputGetApp{
		ClientID: input.ClientID,
	})
	if err != nil {
		// log and then discard error
		ylog.Error(ctx, "get app by id error, continuing to try to insert", ylog.KV("error", err))
		err = nil
	}

	if existingApp.App.ID != "" {
		err = fmt.Errorf("app with client id '%s' already exist", existingApp.App.ClientID)
		return
	}

	appInput := apprepo.App{
		ClientID:  strings.ToLower(input.ClientID),
		Name:      input.Name,
		Enabled:   true,
		CreatedAt: time.Now().UTC(),
	}

	appOut, err := d.conf.AppRepo.CreateApp(ctx, apprepo.InputCreateApp{
		App: appInput,
	})
	if err != nil {
		return
	}

	out = OutCreateApp{
		App: AppFromRepo(appOut.App),
	}
	return
}

func (d *DefaultService) GetApp(ctx context.Context, input InputGetApp) (out OutGetApp, err error) {
	outGetApp, err := d.conf.AppRepo.GetApp(ctx, apprepo.InputGetApp(input))
	if err != nil {
		err = fmt.Errorf("not found app client id '%s': %w", input.ClientID, err)
		return
	}

	out = OutGetApp{
		App: AppFromRepo(outGetApp.App),
	}
	return
}
