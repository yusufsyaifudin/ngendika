package http

import (
	"fmt"
	"net/http"

	"github.com/segmentio/encoding/json"
	"github.com/yusufsyaifudin/ngendika/pkg/fcm"
	"github.com/yusufsyaifudin/ylog"

	"github.com/go-playground/validator/v10"
	"github.com/yusufsyaifudin/ngendika/internal/logic/msgservice"
	"github.com/yusufsyaifudin/ngendika/pkg/response"
	"github.com/yusufsyaifudin/ngendika/pkg/uid"
)

type HandlerMessageServiceConfig struct {
	UID                 uid.UID                      `validate:"required"`
	ResponseConstructor response.HTTPRespConstructor `validate:"required"`
	ResponseWriter      response.Writer              `validate:"required"`
	MsgServiceProcessor msgservice.Service           `validate:"required"`
}

type HandlerMessageService struct {
	Config HandlerMessageServiceConfig
}

func NewHandlerMessageService(config HandlerMessageServiceConfig) (*HandlerMessageService, error) {
	err := validator.New().Struct(config)
	if err != nil {
		return nil, err
	}

	return &HandlerMessageService{Config: config}, nil
}

func (h *HandlerMessageService) SendMessage() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// swagger:model ReqSendMessage
		type ReqSendMessage struct {
			ClientID string `json:"client_id"`

			// this payload must copies from msgservice.Task
			FCMMulticast *fcm.MulticastMessage           `json:"fcm_multicast,omitempty"`
			FCMLegacy    *fcm.LegacyMessage              `json:"fcm_legacy,omitempty"`
			Webhook      []msgservice.TaskPayloadWebhook `json:"webhook,omitempty"`
		}

		var reqBody ReqSendMessage
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		err := dec.Decode(&reqBody)
		if err != nil {
			resp := h.Config.ResponseConstructor.HTTPError(ctx, response.ErrValidation, err)
			h.Config.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
			return
		}

		uuid, err := h.Config.UID.NextID()
		if err != nil {
			return
		}

		var taskResult *msgservice.TaskResult
		taskResult, err = h.Config.MsgServiceProcessor.Process(ctx, &msgservice.Task{
			TraceInfo: ylog.MustExtract(ctx),
			TaskID:    fmt.Sprint(uuid),
			ClientID:  reqBody.ClientID,
			Message: &msgservice.Message{
				FCMMulticast: reqBody.FCMMulticast,
				FCMLegacy:    reqBody.FCMLegacy,
				Webhook:      reqBody.Webhook,
			},
		})

		if err != nil {
			resp := h.Config.ResponseConstructor.HTTPError(ctx, response.ErrUnhandled, err)
			h.Config.ResponseWriter.JSON(http.StatusOK, w, r, resp)
		}

		resp := h.Config.ResponseConstructor.HTTPSuccess(ctx, taskResult)
		h.Config.ResponseWriter.JSON(http.StatusOK, w, r, resp)
		return
	}
}
