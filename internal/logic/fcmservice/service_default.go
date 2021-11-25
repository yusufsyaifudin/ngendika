package fcmservice

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/satori/uuid"
	"github.com/segmentio/encoding/json"
	"github.com/yusufsyaifudin/ngendika/internal/logic/appservice"
	"github.com/yusufsyaifudin/ngendika/internal/storage/fcmrepo"
)

type DefaultServiceConfig struct {
	FCMRepo    fcmrepo.Repo       `validate:"required"`
	AppService appservice.Service `validate:"required"` // fcm service required app service
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

func (i *DefaultService) CreateSvcAccKey(ctx context.Context, input CreateSvcAccKeyIn) (out CreateSvcAccKeyOut, err error) {
	err = validator.New().Struct(input)
	if err != nil {
		err = fmt.Errorf("validation error, missing required field: %w", err)
		return
	}

	app, err := i.GetApp(ctx, input.ClientID)
	if err != nil {
		return
	}

	// unmarshal to save in db as string value
	var serviceAccKey fcmrepo.ServiceAccountKey
	err = json.Unmarshal(input.ServiceAccountKey, &serviceAccKey)
	if err != nil {
		err = fmt.Errorf("error build service account object: %w", err)
		return
	}

	fcmKey := fcmrepo.FCMServiceAccountKey{
		ID:                uuid.NewV4().String(),
		AppID:             app.ID,
		ServiceAccountKey: serviceAccKey,
		CreatedAt:         time.Now().UTC(),
	}

	fcmKey, err = i.conf.FCMRepo.FCMServiceAccountKey().Create(ctx, fcmKey)
	if err != nil {
		err = fmt.Errorf("failed create new fcm: %w", err)
		return
	}

	// marshal, so user can know actual string saved in db
	serviceAccKeyBytes, err := json.Marshal(fcmKey.ServiceAccountKey)
	if err != nil {
		err = fmt.Errorf("error converting actual FCM service account key to bytes: %w", err)
		return
	}

	out = CreateSvcAccKeyOut{
		ID:                fcmKey.ID,
		AppID:             fcmKey.AppID,
		ServiceAccountKey: serviceAccKeyBytes,
		CreatedAt:         fcmKey.CreatedAt.UTC(), // always in UTC
	}
	return
}

func (i *DefaultService) GetSvcAccKey(ctx context.Context, input GetSvcAccKeyIn) (out GetSvcAccKeyOut, err error) {
	err = validator.New().Struct(input)
	if err != nil {
		err = fmt.Errorf("validation error, missing required field: %w", err)
		return
	}

	app, err := i.GetApp(ctx, input.ClientID)
	if err != nil {
		return
	}

	keys, err := i.conf.FCMRepo.FCMServiceAccountKey().FetchAll(ctx, app.ID)
	if err != nil {
		err = fmt.Errorf("not found fcm service account for app client id %s: %w", app.ClientID, err)
		return
	}

	// build output
	out.Lists = make([]GetSvcAccKeyOutList, 0)
	for _, key := range keys {
		b, _ := json.Marshal(key.ServiceAccountKey)
		out.Lists = append(out.Lists, GetSvcAccKeyOutList{
			ID:                key.ID,
			AppID:             key.AppID,
			ServiceAccountKey: b,
			CreatedAt:         key.CreatedAt.UTC(), // always in UTC
		})
	}

	return
}

func (i *DefaultService) GetServerKey(ctx context.Context, input GetServerKeyIn) (out GetServerKeyOut, err error) {
	err = validator.New().Struct(input)
	if err != nil {
		err = fmt.Errorf("validation error, missing required field: %w", err)
		return
	}

	app, err := i.GetApp(ctx, input.ClientID)
	if err != nil {
		return
	}

	keys, err := i.conf.FCMRepo.FCMServerKey().FetchAll(ctx, app.ID)
	if err != nil {
		err = fmt.Errorf("not found fcm service account for app client id %s: %w", app.ClientID, err)
		return
	}

	for _, v := range keys {
		out.Lists = append(out.Lists, GetServerKeyOutList{
			ID:        v.ID,
			AppID:     v.AppID,
			ServerKey: v.ServerKey,
			CreatedAt: v.CreatedAt.UTC(),
		})
	}

	return
}

func (i *DefaultService) GetApp(ctx context.Context, clientID string) (app appservice.App, err error) {
	app, err = i.conf.AppService.GetAppByClientID(ctx, clientID)
	if err != nil {
		return
	}

	if !app.Enabled {
		app = appservice.App{} // always use empty value on error
		err = fmt.Errorf("app %s is disabled", clientID)
		return
	}

	return
}
