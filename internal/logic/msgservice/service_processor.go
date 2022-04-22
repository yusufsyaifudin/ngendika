package msgservice

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/go-resty/resty/v2"
	"github.com/yusufsyaifudin/ngendika/internal/logic/fcmservice"
	"golang.org/x/sync/semaphore"
)

type ProcessorConfig struct {
	FCMService fcmservice.Service `validate:"required"`
	RESTClient *resty.Client      `validate:"required"`
	MaxWorker  int                `validate:"required,min=1"`
}

type Processor struct {
	Config ProcessorConfig
	sem    *semaphore.Weighted
}

var _ Service = (*Processor)(nil)

func NewProcessor(conf ProcessorConfig) (*Processor, error) {
	err := validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	return &Processor{
		Config: conf,
		sem:    semaphore.NewWeighted(int64(conf.MaxWorker)),
	}, nil
}

func (p *Processor) Process() Process {
	return func(ctx context.Context, task *Task) (out *TaskResult, err error) {

		var (
			outFCMMulticast    *fcmservice.FCMMulticastOutput
			outFCMMulticastErr error
			outFCMLegacy       *fcmservice.FCMLegacyOutput
			outFCMLegacyErr    error
			outWebhook         *WebhookOutput
			outWebhookErr      error
		)

		// perform task on each payload
		wg := sync.WaitGroup{}

		if task.Message.FCMMulticast != nil {
			err = p.sem.Acquire(ctx, 1)
			if err != nil {
				err = fmt.Errorf("cannot acquire semaphore: %w", err)
				return
			}

			wg.Add(1)

			go func() {
				defer func() {
					wg.Done()
					p.sem.Release(1)
				}()

				outFCMMulticast, outFCMMulticastErr = p.Config.FCMService.FcmMulticast(ctx, &fcmservice.FcmMulticastInput{
					AppClientID: task.ClientID,
					Payload:     task.Message.FCMMulticast,
				})
			}()
		}

		if task.Message.FCMLegacy != nil {
			err = p.sem.Acquire(ctx, 1)
			if err != nil {
				err = fmt.Errorf("cannot acquire semaphore: %w", err)
				return
			}

			wg.Add(1)

			go func() {
				defer func() {
					wg.Done()
					p.sem.Release(1)
				}()

				outFCMLegacy, outFCMLegacyErr = p.Config.FCMService.FcmLegacy(ctx, &fcmservice.FcmLegacyInput{
					AppClientID: task.ClientID,
					Payload:     task.Message.FCMLegacy,
				})
			}()
		}

		if task.Message.Webhook != nil && len(task.Message.Webhook) > 0 {
			err = p.sem.Acquire(ctx, 1)
			if err != nil {
				err = fmt.Errorf("cannot acquire semaphore: %w", err)
				return
			}

			wg.Add(1)

			go func() {
				defer func() {
					wg.Done()
					p.sem.Release(1)
				}()

				outWebhook, outWebhookErr = p.Webhook(ctx, &WebhookInput{
					AppClientID: task.ClientID,
					Webhook:     task.Message.Webhook,
				})
			}()
		}

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

func (p *Processor) Webhook(ctx context.Context, input *WebhookInput) (out *WebhookOutput, err error) {
	// skip webhook message because it is a slice not struct
	newMsgWebhook := make([]TaskPayloadWebhook, 0)
	for i, webhook := range input.Webhook {
		err = validator.New().Struct(webhook)
		if err != nil {
			err = fmt.Errorf(
				"some webhook payload index %d with reference id '%s' validation error: %w",
				i, webhook.ReferenceID, err,
			)
			return
		}

		newMsgWebhook = append(newMsgWebhook, webhook)
	}

	var webhookResult = make([]WebhookResult, 0)
	for _, payload := range newMsgWebhook {
		header := map[string]string{}
		for k, v := range payload.Header {
			header[k] = strings.Join(v, " ")
		}

		req := p.Config.RESTClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
			R().
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

	if len(webhookResult) <= 0 {
		// return nil if empty webhook results
		out = nil
		err = nil
		return
	}

	// build output
	out = &WebhookOutput{
		Result: webhookResult,
	}
	return
}
