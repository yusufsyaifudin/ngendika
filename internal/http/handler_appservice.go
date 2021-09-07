package http

import (
	"bytes"
	"encoding/base64"
	"net/http"

	"github.com/segmentio/encoding/json"
	"github.com/yusufsyaifudin/ngendika/internal/logic/appservice"
	"github.com/yusufsyaifudin/ngendika/pkg/logger"
	"github.com/yusufsyaifudin/ngendika/pkg/response"
	"github.com/yusufsyaifudin/ngendika/storage/apprepo"
)

type HandlerAppService struct {
	Logger              logger.Logger                `validate:"required"`
	ResponseConstructor response.HTTPRespConstructor `validate:"required"`
	ResponseWriter      response.Writer              `validate:"required"`
	AppService          appservice.Service           `validate:"required"`
}

// CreateApp .
func (h *HandlerAppService) CreateApp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var reqBody appservice.InputCreateApp
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&reqBody)
	if err != nil {
		resp := h.ResponseConstructor.HTTPError(ctx, response.ErrValidation, err)
		h.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
		return
	}

	out, err := h.AppService.CreateApp(ctx, reqBody)
	if err != nil {
		resp := h.ResponseConstructor.HTTPError(ctx, response.ErrUnhandled, err)
		h.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
		return
	}

	resp := h.ResponseConstructor.HTTPSuccess(ctx, out)
	h.ResponseWriter.JSON(http.StatusOK, w, r, resp)
	return
}

func (h *HandlerAppService) PutFCMServiceAccountKey() func(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		ClientID          string `json:"client_id" validate:"required"`
		ServiceAccountKey string `json:"service_account_key" validate:"required,base64"` // contains service account key JSON in bas64-encoded string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var reqBody Request
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		err := dec.Decode(&reqBody)
		if err != nil {
			resp := h.ResponseConstructor.HTTPError(ctx, response.ErrValidation, err)
			h.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
			return
		}

		svcAccKey, err := base64.URLEncoding.DecodeString(reqBody.ServiceAccountKey)
		if err != nil {
			resp := h.ResponseConstructor.HTTPError(ctx, response.ErrValidation, err)
			h.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
			return
		}

		var fcmServiceAccountKey apprepo.ServiceAccountKey
		dec = json.NewDecoder(bytes.NewReader(svcAccKey))
		dec.DisallowUnknownFields()
		err = dec.Decode(&fcmServiceAccountKey)
		if err != nil {
			resp := h.ResponseConstructor.HTTPError(ctx, response.ErrValidation, err)
			h.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
			return
		}

		out, err := h.AppService.CreateFcmServiceAccountKey(ctx, appservice.InputCreateFcmServiceAccountKey{
			ClientID:             reqBody.ClientID,
			FCMServiceAccountKey: fcmServiceAccountKey,
		})

		if err != nil {
			resp := h.ResponseConstructor.HTTPError(ctx, response.ErrUnhandled, err)
			h.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
			return
		}

		resp := h.ResponseConstructor.HTTPSuccess(ctx, out)
		h.ResponseWriter.JSON(http.StatusOK, w, r, resp)
		return
	}
}
