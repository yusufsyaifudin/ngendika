package appservice

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/satori/uuid"
	"github.com/yusufsyaifudin/ngendika/storage/apprepo"
)

type Config struct {
	AppRepo apprepo.Repo `validate:"required"`
}

type DefaultService struct {
	conf Config
}

// CreateApp is a function that knows business logic.
// It doesn't know whether the input come from HTTP or GRPC or any input.
func (d *DefaultService) CreateApp(ctx context.Context, input InputCreateApp) (out OutputCreateApp, err error) {

	appRepo := d.conf.AppRepo

	app := apprepo.NewApp(input.ClientID, input.Name)
	err = validator.New().Struct(app)
	if err != nil {
		err = fmt.Errorf("validation error, missing required field: %w", err)
		return
	}

	app, err = appRepo.CreateApp(ctx, app)
	if err != nil {
		return
	}

	out = OutputCreateApp{
		App: App{
			ClientID:  app.ClientID,
			Name:      app.Name,
			Enabled:   app.Enabled,
			CreatedAt: app.CreatedAt,
		},
	}
	return
}

func (d *DefaultService) CreateFcmServiceAccountKey(ctx context.Context, input InputCreateFcmServiceAccountKey) (out OutputCreateFcmServiceAccountKey, err error) {
	appRepo := d.conf.AppRepo

	app, err := appRepo.GetAppByClientID(ctx, input.ClientID)
	if err != nil {
		err = fmt.Errorf("client id %s not found: %w", input.ClientID, err)
		return
	}

	if !app.Enabled {
		err = fmt.Errorf("client_id %s is disabled", input.ClientID)
		return
	}

	fcmKey := apprepo.FCMServiceAccountKey{
		ID:                uuid.NewV4().String(),
		AppClientID:       app.ClientID,
		ServiceAccountKey: input.FCMServiceAccountKey,
		CreatedAt:         time.Now().UTC(),
	}
	err = validator.New().Struct(fcmKey)
	if err != nil {
		err = fmt.Errorf("validation error, missing required field: %w", err)
		return
	}

	fcmKey, err = appRepo.CreateFCMServiceAccountKey(ctx, fcmKey)
	if err != nil {
		err = fmt.Errorf("failed create new fcm: %w", err)
		return
	}

	out = OutputCreateFcmServiceAccountKey{
		ServiceAccountKey: fcmKey.ServiceAccountKey,
		CreatedAt:         fcmKey.CreatedAt,
	}
	return
}

var _ Service = (*DefaultService)(nil)

func New(dep Config) (*DefaultService, error) {
	if err := validator.New().Struct(dep); err != nil {
		return nil, err
	}

	return &DefaultService{
		conf: dep,
	}, nil
}
