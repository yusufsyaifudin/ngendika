package appsvc

import (
	"context"
	"fmt"
	"github.com/yusufsyaifudin/ngendika/pkg/tracer"
	"github.com/yusufsyaifudin/ngendika/pkg/uid"
	"go.opentelemetry.io/otel/trace"
	"strings"
	"time"

	"github.com/yusufsyaifudin/ngendika/internal/svc/apprepo"
	"github.com/yusufsyaifudin/ngendika/pkg/validator"
	"github.com/yusufsyaifudin/ylog"
)

type DefaultServiceConfig struct {
	UIDGen  uid.UID      `validate:"required"`
	AppRepo apprepo.Repo `validate:"required"`
}

type DefaultService struct {
	Config DefaultServiceConfig
}

var _ Service = (*DefaultService)(nil)

func New(dep DefaultServiceConfig) (*DefaultService, error) {
	if err := validator.Validate(dep); err != nil {
		return nil, err
	}

	return &DefaultService{
		Config: dep,
	}, nil
}

// CreateApp is a function that knows business logic.
// It doesn't know whether the input come from HTTP or GRPC or any input.
func (d *DefaultService) CreateApp(ctx context.Context, input InputCreateApp) (out OutCreateApp, err error) {
	err = validator.Validate(input)
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

	if existingApp.App.ClientID != "" {
		err = fmt.Errorf("app with client id '%s' already exist", existingApp.App.ClientID)
		return
	}

	nextID, err := d.Config.UIDGen.NextID()
	if err != nil {
		err = fmt.Errorf("cannot get next id: %w", err)
		return
	}

	now := time.Now().UTC()
	appInput := apprepo.App{
		ID:        int64(nextID),
		ClientID:  strings.ToLower(input.ClientID),
		Name:      input.Name,
		CreatedAt: now.UnixMicro(),
		UpdatedAt: now.UnixMicro(),
	}

	appOut, err := d.Config.AppRepo.Create(ctx, apprepo.InputCreate{
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

// PutApp .
func (d *DefaultService) PutApp(ctx context.Context, input InputPutApp) (out OutPutApp, err error) {
	err = validator.Validate(input)
	if err != nil {
		err = fmt.Errorf("validation error, missing required field: %w", err)
		return
	}

	nextID, err := d.Config.UIDGen.NextID()
	if err != nil {
		err = fmt.Errorf("cannot get next id: %w", err)
		return
	}

	now := time.Now().UTC()
	appCandidate := apprepo.App{
		ID:        int64(nextID),
		ClientID:  strings.ToLower(input.ClientID),
		Name:      input.Name,
		CreatedAt: now.UnixMicro(),
		UpdatedAt: now.UnixMicro(),
	}

	appOut, err := d.Config.AppRepo.Upsert(ctx, apprepo.InputUpsert{
		App: appCandidate,
	})
	if err != nil {
		return
	}

	out = OutPutApp{
		App: AppFromRepo(appOut.App),
	}
	return
}

func (d *DefaultService) GetApp(ctx context.Context, input InputGetApp) (out OutGetApp, err error) {
	var span trace.Span
	ctx, span = tracer.StartSpan(ctx, "appsvc.GetApp")
	defer span.End()

	err = validator.Validate(input)
	if err != nil {
		err = fmt.Errorf("validation error, missing required field: %w", err)
		return
	}

	outGetApp, err := d.Config.AppRepo.GetByClientID(ctx, apprepo.InputGetByClientID(input))
	if err != nil {
		err = fmt.Errorf("not found app client id '%s': %w", input.ClientID, err)
		return
	}

	appData := outGetApp.App
	out = OutGetApp{
		App: AppFromRepo(appData),
	}
	return
}

func (d *DefaultService) ListApp(ctx context.Context, in InputListApp) (out OutListApp, err error) {
	err = validator.Validate(in)
	if err != nil {
		err = fmt.Errorf("validation error, missing required field: %w", err)
		return
	}

	// set to the default value
	if in.Limit <= 0 || in.Limit > 100 {
		in.Limit = 100
	}

	outList, err := d.Config.AppRepo.List(ctx, apprepo.InputList{
		Limit:    in.Limit,
		BeforeID: in.BeforeID,
		AfterID:  in.AfterID,
	})
	if err != nil {
		err = fmt.Errorf("list apps error: %w", err)
		return
	}

	apps := make([]App, 0)
	for _, app := range outList.Apps {
		apps = append(apps, AppFromRepo(app))
	}

	out = OutListApp{
		Total: outList.Total,
		Limit: in.Limit,
		Apps:  apps,
	}

	return
}

func (d *DefaultService) DelApp(ctx context.Context, input InputDelApp) (out OutDelApp, err error) {
	err = validator.Validate(input)
	if err != nil {
		err = fmt.Errorf("validation error, missing required field: %w", err)
		return
	}

	inDelApp := apprepo.InputDelByClientID{
		ClientID:  input.ClientID,
		DeletedAt: time.Now().UTC().UnixMicro(),
	}

	outDelApp, err := d.Config.AppRepo.DelByClientID(ctx, inDelApp)
	if err != nil {
		err = fmt.Errorf("db delete error '%s': %w", input.ClientID, err)
		return
	}

	out = OutDelApp{
		Success: outDelApp.Success,
	}
	return
}
