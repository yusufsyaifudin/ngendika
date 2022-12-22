package msgsvc

import (
	"context"
	"fmt"
	"github.com/yusufsyaifudin/ngendika/backend"
	"github.com/yusufsyaifudin/ngendika/internal/svc/appsvc"
	"github.com/yusufsyaifudin/ngendika/internal/svc/pnpsvc"
	"github.com/yusufsyaifudin/ngendika/pkg/tracer"
	"github.com/yusufsyaifudin/ngendika/pkg/validator"
	"go.opentelemetry.io/otel/trace"
	"sync"
	"time"
)

type SvcSyncConfig struct {
	AppSvc        appsvc.Service    `validate:"required"`
	PNProviderSvc pnpsvc.Service    `validate:"required"`
	PNSender      backend.SenderMux `validate:"required"`
	MaxBuffer     int               `validate:"required,min=1"`
	MaxWorker     int               `validate:"required,min=1"` // MaxWorker number of maximum go routine for all backend type
}

type SvcSync struct {
	Config       SvcSyncConfig
	MessageQueue chan senderWorkerJob
}

var _ Service = (*SvcSync)(nil)

func New(cfg SvcSyncConfig) (*SvcSync, error) {
	err := validator.Validate(cfg)
	if err != nil {
		return nil, err
	}

	svc := &SvcSync{
		Config:       cfg,
		MessageQueue: make(chan senderWorkerJob, cfg.MaxBuffer),
	}

	for i := 1; i <= cfg.MaxWorker; i++ {
		go senderWorker(i, cfg.PNSender, svc.MessageQueue)
	}

	return svc, nil
}

func (p *SvcSync) Process(ctx context.Context, input *InputProcess) (out *OutProcess, err error) {
	var span trace.Span
	ctx, span = tracer.StartSpan(ctx, "msgsvc.Process")
	defer span.End()

	err = validator.Validate(input)
	if err != nil {
		err = fmt.Errorf("validation error: %w", err)
		return
	}

	getAppIn := appsvc.InputGetApp{ClientID: input.ClientID}
	getAppOut, err := p.Config.AppSvc.GetApp(ctx, getAppIn)
	if err != nil {
		return
	}

	app := getAppOut.App

	lock := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	errs := make([]string, 0)
	wgReport := &senderWorkerJobReport{
		BackendReports: map[int64][]backendReport{},
	}

	allPnpMapByID := make(map[int64]backend.PushNotificationProvider)

	for provider, payloads := range input.Payloads {

		// get push notification provider only one per provider
		outGetServiceProvider, _err := p.Config.PNProviderSvc.GetByLabels(ctx, pnpsvc.InGetByLabels{
			AppID:    app.ID,
			Provider: provider,
			Label:    input.Label,
		})
		if _err != nil {
			_err = fmt.Errorf("failed get service provider '%s': %w", provider, _err)
			errs = append(errs, _err.Error())
			continue
		}

		// we may get push notification config more than one, because we use label:* or label1,label2.
		// so, we need to iterate every push notification configuration.
		// And since one push notification can have array of payloads, we need to iterate that.
		for _, pnProvider := range outGetServiceProvider.PnProviders {
			allPnpMapByID[pnProvider.ID] = pnProvider

			for _, payload := range payloads {
				msg := &backend.Message{
					ReferenceID: input.TaskID,
					RawPayload:  payload,
				}

				wg.Add(1)
				p.MessageQueue <- senderWorkerJob{
					Ctx:             ctx,
					Lock:            lock,
					Wg:              wg,
					SubmitTime:      time.Now(),
					ServiceProvider: pnProvider,
					Message:         msg,
					Report:          wgReport,
				}
			}

		}
	}

	wg.Wait()

	// Group report with the push notification provider
	reportGroup := make([]ReportGroup, 0)
	for pnpID, reports := range wgReport.BackendReports {
		pnp := allPnpMapByID[pnpID]

		collectiveErr := make([]string, 0)
		collectiveReport := make([]*backend.Report, 0)

		for _, report := range reports {
			if report.BackendError != "" {
				collectiveErr = append(collectiveErr, report.BackendError)
			}

			if report.BackendReport != nil {
				collectiveReport = append(collectiveReport, report.BackendReport)
			}
		}

		reportGroup = append(reportGroup, ReportGroup{
			PNP:            pnp,
			BackendErrors:  collectiveErr,
			BackendReports: collectiveReport,
		})
	}

	out = &OutProcess{
		TaskID:      input.TaskID,
		App:         app,
		Errors:      errs,
		ReportGroup: reportGroup,
	}

	return
}
