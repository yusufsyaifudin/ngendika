package appservice

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/satori/uuid"
	"github.com/yusufsyaifudin/ngendika/internal/storage/apprepo"
	"github.com/yusufsyaifudin/ngendika/internal/storage/fcmrepo"
)

type Config struct {
	AppRepo apprepo.Repo `validate:"required"`
	FCMRepo fcmrepo.Repo `validate:"required"`
}

type DefaultService struct {
	conf Config
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

// CreateApp is a function that knows business logic.
// It doesn't know whether the input come from HTTP or GRPC or any input.
func (d *DefaultService) CreateApp(ctx context.Context, input CreateAppIn) (out CreateAppOut, err error) {

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

	out = CreateAppOut{
		App: App{
			ClientID:  app.ClientID,
			Name:      app.Name,
			Enabled:   app.Enabled,
			CreatedAt: app.CreatedAt,
		},
	}
	return
}

func (d *DefaultService) CreateFcmSvcAccKey(ctx context.Context, input CreateFcmSvcAccKeyIn) (out CreateFcmSvcAccKeyOut, err error) {
	appRepo := d.conf.AppRepo
	fcmSvcAccKeyRepo := d.conf.FCMRepo.FCMServiceAccountKey()

	err = validator.New().Struct(input)
	if err != nil {
		err = fmt.Errorf("validation error, missing required field: %w", err)
		return
	}

	app, err := appRepo.GetAppByClientID(ctx, input.ClientID)
	if err != nil {
		err = fmt.Errorf("client id %s not found: %w", input.ClientID, err)
		return
	}

	if !app.Enabled {
		err = fmt.Errorf("client_id %s is disabled", input.ClientID)
		return
	}

	fcmKey := fcmrepo.FCMServiceAccountKey{
		ID:                uuid.NewV4().String(),
		AppID:             app.ID,
		ServiceAccountKey: input.FCMServiceAccountKey,
		CreatedAt:         time.Now().UTC(),
	}

	fcmKey, err = fcmSvcAccKeyRepo.Create(ctx, fcmKey)
	if err != nil {
		err = fmt.Errorf("failed create new fcm: %w", err)
		return
	}

	out = CreateFcmSvcAccKeyOut{
		ServiceAccountKey: fcmKey.ServiceAccountKey,
		CreatedAt:         fcmKey.CreatedAt,
	}
	return
}

func (d *DefaultService) GetFcmSvcAccKey(ctx context.Context, input GetFcmSvcAccKeyIn) (out GetFcmSvcAccKeyOut, err error) {
	err = validator.New().Struct(input)
	if err != nil {
		err = fmt.Errorf("validation error, missing required field: %w", err)
		return
	}

	appRepo := d.conf.AppRepo
	fcmSvcAccKeyRepo := d.conf.FCMRepo.FCMServiceAccountKey()

	app, err := appRepo.GetAppByClientID(ctx, input.ClientID)
	if err != nil {
		err = fmt.Errorf("client id %s not found: %w", input.ClientID, err)
		return
	}

	if !app.Enabled {
		err = fmt.Errorf("client_id %s is disabled", input.ClientID)
		return
	}

	keys, err := fcmSvcAccKeyRepo.FetchAll(ctx, app.ID)
	if err != nil {
		return GetFcmSvcAccKeyOut{}, err
	}

	out.Lists = make([]GetFcmSvcAccKeyOutList, 0)
	for _, key := range keys {
		out.Lists = append(out.Lists, GetFcmSvcAccKeyOutList{
			ID:                key.ID,
			ServiceAccountKey: key.ServiceAccountKey,
		})
	}

	return
}
