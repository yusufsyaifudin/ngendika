package msgservice

import (
	"context"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/go-resty/resty/v2"
	"github.com/yusufsyaifudin/ngendika/internal/logic/fcmservice"
	"github.com/yusufsyaifudin/ngendika/pkg/fcm"
)

type ProcessorConfig struct {
	FCMService fcmservice.Service `validate:"required"`

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
		err = validator.New().Struct(task)
		if err != nil {
			return nil, err
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
				AppClientID: task.ClientID,
				Payload:     task.Message.FCMMulticast,
			})
		}()

		go func() {
			defer wg.Done()

			outFCMLegacy, outFCMLegacyErr = p.FcmLegacy(ctx, &FcmLegacyInput{
				AppClientID: task.ClientID,
				Payload:     task.Message.FCMLegacy,
			})
		}()

		go func() {
			defer wg.Done()

			outWebhook, outWebhookErr = p.Webhook(ctx, &WebhookInput{
				AppClientID: task.ClientID,
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

	fcmService, err := p.Config.FCMService.GetSvcAccKey(ctx, fcmservice.GetSvcAccKeyIn{
		ClientID: input.AppClientID,
	})
	if err != nil {
		return nil, err
	}

	var multicastRes = make([]FCMMulticastResult, 0)
	for _, serviceAccountKey := range fcmService.Lists {
		batchRes, err := p.Config.FCMClient.SendMulticast(ctx, serviceAccountKey.ServiceAccountKey, input.Payload)
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

	fcmServerKeys, err := p.Config.FCMService.GetServerKey(ctx, fcmservice.GetServerKeyIn{
		ClientID: input.AppClientID,
	})
	if err != nil {
		return nil, err
	}

	var legacyResp = make([]FCMLegacyResult, 0)

	for _, key := range fcmServerKeys.Lists {
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
				ReferenceID: payload.ReferenceID,
				Error: &WebhookError{
					Code:          ErrWebhookRequest,
					MessageDetail: errResp.Error(),
				},
			})
			continue
		}

		webhookResult = append(webhookResult, WebhookResult{
			ReferenceID: payload.ReferenceID,
			Header:      resp.Header(),
			Body:        string(resp.Body()),
		})
	}

	// build output
	out = &WebhookOutput{
		Result: webhookResult,
	}
	return
}
