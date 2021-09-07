package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/segmentio/encoding/json"
	"github.com/yusufsyaifudin/ngendika/internal/logic/msgservice"
	"github.com/yusufsyaifudin/ngendika/pkg/fcm"
	"github.com/yusufsyaifudin/ngendika/pkg/logger"
	"github.com/yusufsyaifudin/ngendika/pkg/response"
	"github.com/yusufsyaifudin/ngendika/pkg/uid"
)

type HandlerMessageService struct {
	Logger               logger.Logger                `validate:"required"`
	UID                  uid.UID                      `validate:"required"`
	ResponseConstructor  response.HTTPRespConstructor `validate:"required"`
	ResponseWriter       response.Writer              `validate:"required"`
	MsgServiceDispatcher msgservice.Service           `validate:"required"`
	MsgServiceProcessor  msgservice.Service           `validate:"required"`
}

func (h *HandlerMessageService) SendMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// swagger:model ReqSendMessage
	type ReqSendMessage struct {
		ClientID string `json:"client_id"`
		Sync     bool   `json:"sync"`

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
		resp := h.ResponseConstructor.HTTPError(ctx, response.ErrValidation, err)
		h.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
		return
	}

	var taskResult *msgservice.TaskResult

	var msgServiceProcess func(ctx context.Context, task *msgservice.Task) (out *msgservice.TaskResult, err error)
	msgServiceProcess = h.MsgServiceDispatcher.Process
	if reqBody.Sync {
		msgServiceProcess = h.MsgServiceProcessor.Process
	}

	fcmMulticast := &msgservice.TaskPayloadFCMMulticast{
		Msg: reqBody.FCMMulticast,
	}

	fcmLegacy := &msgservice.TaskPayloadFCMLegacy{
		Msg: reqBody.FCMLegacy,
	}

	uuid, err := h.UID.NextID()
	if err != nil {
		return
	}

	taskResult, err = msgServiceProcess(ctx, &msgservice.Task{
		TraceInfo:               logger.MustExtract(ctx),
		TaskID:                  fmt.Sprint(uuid),
		AppClientID:             reqBody.ClientID,
		TaskPayloadFCMMulticast: fcmMulticast,
		TaskPayloadFCMLegacy:    fcmLegacy,
		TaskPayloadWebhook:      reqBody.Webhook,
	})

	if err != nil {
		resp := h.ResponseConstructor.HTTPError(ctx, response.ErrUnhandled, err)
		h.ResponseWriter.JSON(http.StatusOK, w, r, resp)
	}

	resp := h.ResponseConstructor.HTTPSuccess(ctx, taskResult)
	h.ResponseWriter.JSON(http.StatusOK, w, r, resp)
	return
}
