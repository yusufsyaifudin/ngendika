package pnpsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/yusufsyaifudin/ngendika/backend"
	"github.com/yusufsyaifudin/ngendika/internal/svc/pnprepo"
	"github.com/yusufsyaifudin/ngendika/pkg/tracer"
	"github.com/yusufsyaifudin/ngendika/pkg/uid"
	"github.com/yusufsyaifudin/ngendika/pkg/validator"
	"go.opentelemetry.io/otel/trace"
	"strings"
	"time"
)

type Config struct {
	UIDGen  uid.UID      `validate:"required"`
	PnpRepo pnprepo.Repo `validate:"required"`
}

type ServiceDefault struct {
	Config Config
}

var _ Service = (*ServiceDefault)(nil)

func New(cfg Config) (svc *ServiceDefault, err error) {
	err = validator.Validate(cfg)
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrValidation, err)
		return
	}

	svc = &ServiceDefault{
		Config: cfg,
	}

	return
}

func (s *ServiceDefault) Create(ctx context.Context, in InCreate) (out OutCreate, err error) {
	// validate the top level field first, i.e: app client id first
	err = validator.Validate(in)
	if err != nil {
		err = fmt.Errorf("%w: error create service provider: %s", ErrValidation, err)
		return
	}

	credentialJson, err := json.Marshal(in.PnProvider.CredentialJSON)

	id, err := s.Config.UIDGen.NextID()
	if err != nil {
		err = fmt.Errorf("cannot generate uid for new record: %w", err)
		return
	}

	now := time.Now().UTC()
	serviceProviderCandidate := pnprepo.PushNotificationProvider{
		ID:             int64(id),
		AppID:          in.AppID,
		Provider:       in.PnProvider.Provider,
		Label:          in.PnProvider.Label,
		CredentialJSON: string(credentialJson),
		CreatedAt:      now.UnixMicro(),
		UpdatedAt:      now.UnixMicro(),
	}

	_, err = backend.MuxBackend().ValidateCredJson(ctx, serviceProviderCandidate.Provider, serviceProviderCandidate.CredentialJSON)
	if err != nil {
		err = fmt.Errorf("cannot validate candidate service provider: %w", err)
		return
	}

	inCreate := pnprepo.InputInsert{
		PnProvider: serviceProviderCandidate,
	}

	outCreate, err := s.Config.PnpRepo.Insert(ctx, inCreate)
	if err != nil {
		err = fmt.Errorf("cannot insert email config record: %w", err)
		return
	}

	out = OutCreate{
		ServiceProvider: FromRepo(outCreate.PnProvider),
	}

	return
}

func (s *ServiceDefault) GetByLabels(ctx context.Context, in InGetByLabels) (out OutGetByLabels, err error) {
	var span trace.Span
	ctx, span = tracer.StartSpan(ctx, "pnpsvc.GetByLabels")
	defer span.End()

	err = validator.Validate(in)
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrValidation, err)
		return
	}

	inGetOneByLabel := pnprepo.InGetByLabels{
		AppID:    in.AppID,
		Provider: in.Provider,
		Labels:   strings.Split(in.Label, ","),
	}
	outGetOneByLabel, err := s.Config.PnpRepo.GetByLabels(ctx, inGetOneByLabel)
	if err != nil {
		err = fmt.Errorf("cannot get by label: %s", err)
		return
	}

	pnProviders := make([]backend.PushNotificationProvider, 0)
	for _, provider := range outGetOneByLabel.PnProvider {
		pnProviders = append(pnProviders, FromRepo(provider))
	}

	out = OutGetByLabels{
		PnProviders: pnProviders,
	}

	return
}

func (s *ServiceDefault) Examples(ctx context.Context) (out OutExamples) {
	out = OutExamples{
		Items: backend.MuxBackend().Examples(ctx),
	}

	return
}
