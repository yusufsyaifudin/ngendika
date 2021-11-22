package msgservice

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/go-resty/resty/v2"
	"github.com/segmentio/encoding/json"
	"github.com/yusufsyaifudin/ngendika/internal/storage/apprepo"
	"github.com/yusufsyaifudin/ngendika/internal/storage/fcmrepo"
	"github.com/yusufsyaifudin/ngendika/pkg/fcm"
)

type ProcessorConfig struct {
	AppRepo apprepo.Repo `validate:"required"`
	FCMRepo fcmrepo.Repo `validate:"required"`

	FCMClient  fcm.Client    `validate:"required"`
	RESTClient *resty.Client `validate:"required"`
}

type Processor struct {
	Config ProcessorConfig
}

var _ Service = (*Processor)(nil)

func NewProcessor(conf ProcessorConfig) (*Processor, error) {
	err := validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	return &Processor{
		Config: conf,
	}, nil
}

func (p *Processor) Process() Process {
	return func(ctx context.Context, task *Task) (out *TaskResult, err error) {
		appRepo := p.Config.AppRepo

		if task.Message == nil {
			task.Message = &Message{}
		}

		err = validator.New().Struct(task)
		if err != nil {
			return
		}

		app, err := appRepo.GetAppByClientID(ctx, task.ClientID)
		if err != nil {
			err = fmt.Errorf("client id %s not found: %w", task.ClientID, err)
			return
		}

		if !app.Enabled {
			err = fmt.Errorf("app %s is diabled", app.ClientID)
			return
		}

		var (
			outFCMMulticast    *FCMMulticastOutput
			outFCMMulticastErr error
			outFCMLegacy       *FCMLegacyOutput
			outFCMLegacyErr    error
			outWebhook         *WebhookOutput
			outWebhookErr      error
		)

		// perform task on each payload
		wg := sync.WaitGroup{}
		wg.Add(3)

		go func() {
			defer wg.Done()

			outFCMMulticast, outFCMMulticastErr = p.FcmMulticast(ctx, &FcmMulticastInput{
				AppID:   app.ID,
				Payload: task.Message.FCMMulticast,
			})
		}()

		go func() {
			defer wg.Done()

			outFCMLegacy, outFCMLegacyErr = p.FcmLegacy(ctx, &FcmLegacyInput{
				AppID:   app.ID,
				Payload: task.Message.FCMLegacy,
			})
		}()

		go func() {
			defer wg.Done()

			outWebhook, outWebhookErr = p.Webhook(ctx, &WebhookInput{
				AppClientID: app.ID,
				Webhook:     task.Message.Webhook,
			})
		}()

		// wait for all go routine to succeed
		wg.Wait()

		out = &TaskResult{
			TaskID:            task.TaskID,
			AppClientID:       task.ClientID,
			FCMMulticast:      outFCMMulticast,
			FCMMulticastError: ErrorString(outFCMMulticastErr),
			FCMLegacy:         outFCMLegacy,
			FCMLegacyError:    ErrorString(outFCMLegacyErr),
			Webhook:           outWebhook,
			WebhookError:      ErrorString(outWebhookErr),
		}

		return
	}
}

func (p *Processor) FcmMulticast(ctx context.Context, input *FcmMulticastInput) (out *FCMMulticastOutput, err error) {
	if input == nil {
		return
	}

	err = validator.New().Struct(input)
	if err != nil {
		return
	}

	fcmSvcAccKeyRepo := p.Config.FCMRepo.FCMServiceAccountKey()

	serviceAccountKeys, err := fcmSvcAccKeyRepo.FetchAll(ctx, input.AppID)
	if err != nil {
		err = fmt.Errorf("error get fcm service account key: %w", err)
		return
	}

	var multicastRes = make([]FCMMulticastResult, 0)

	for _, serviceAccountKey := range serviceAccountKeys {
		key, err := json.Marshal(serviceAccountKey.ServiceAccountKey) // please ensure you pass the right data to marshal here
		if err != nil {
			err = fmt.Errorf("service account json for id %s is error: %w", serviceAccountKey.ID, err)
			multicastRes = append(multicastRes, FCMMulticastResult{
				FCMKeyID: serviceAccountKey.ID,
				Error:    err.Error(),
			})
			continue
		}

		batchRes, err := p.Config.FCMClient.SendMulticast(ctx, key, input.Payload)
		if err != nil {
			multicastRes = append(multicastRes, FCMMulticastResult{
				FCMKeyID: serviceAccountKey.ID,
				Error:    err.Error(),
			})
			continue
		}

		multicastRes = append(multicastRes, FCMMulticastResult{
			FCMKeyID:    serviceAccountKey.ID,
			BatchResult: &batchRes,
		})
	}

	// build final output
	out = &FCMMulticastOutput{
		Result: multicastRes,
	}

	return
}

func (p *Processor) FcmLegacy(ctx context.Context, input *FcmLegacyInput) (out *FCMLegacyOutput, err error) {
	if input == nil {
		return
	}

	err = validator.New().Struct(input)
	if err != nil {
		return
	}

	fcmServerKey := p.Config.FCMRepo.FCMServerKey()

	serverKeys, err := fcmServerKey.FetchAll(ctx, input.AppID)
	if err != nil {
		err = fmt.Errorf("error get fcm service acoount key: %w", err)
		return
	}

	var legacyResp = make([]FCMLegacyResult, 0)

	for _, key := range serverKeys {

		batchRes, err := p.Config.FCMClient.SendLegacy(ctx, key.ServerKey, input.Payload)
		if err != nil {
			legacyResp = append(legacyResp, FCMLegacyResult{
				FCMKeyID: key.ID,
				Error:    err.Error(),
			})
			continue
		}

		legacyResp = append(legacyResp, FCMLegacyResult{
			FCMKeyID:    key.ID,
			BatchResult: &batchRes,
		})
	}

	// build final output
	out = &FCMLegacyOutput{
		Result: legacyResp,
	}

	return
}

func (p *Processor) Webhook(ctx context.Context, input *WebhookInput) (out *WebhookOutput, err error) {
	if input == nil {
		return
	}

	err = validator.New().Struct(input)
	if err != nil {
		return
	}

	var webhookResult = make([]WebhookResult, 0)
	for _, payload := range input.Webhook {

		header := map[string]string{}
		for k, v := range payload.Header {
			header[k] = strings.Join(v, " ")
		}

		req := p.Config.RESTClient.R().
			SetContext(ctx).
			SetHeaders(header).
			SetBody(payload.Body).
			SetQueryParamsFromValues(payload.QueryParam).
			SetFormDataFromValues(payload.FormData)

		resp, errResp := req.Execute(req.Method, payload.URL)

		if errResp != nil {
			webhookResult = append(webhookResult, WebhookResult{
				Error: &WebhookError{
					Code:          ErrWebhookRequest,
					MessageDetail: errResp.Error(),
				},
			})
			continue
		}

		webhookResult = append(webhookResult, WebhookResult{
			Header: resp.Header(),
			Body:   string(resp.Body()),
		})
	}

	// build output
	out = &WebhookOutput{
		Result: webhookResult,
	}
	return
}
