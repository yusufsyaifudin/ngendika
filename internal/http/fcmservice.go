package http

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/segmentio/encoding/json"
	"github.com/yusufsyaifudin/ngendika/internal/logic/appservice"
	"github.com/yusufsyaifudin/ngendika/pkg/logger"
	"github.com/yusufsyaifudin/ngendika/pkg/response"
	"github.com/yusufsyaifudin/ngendika/storage/fcmsvcacckeyrepo"
)

type ConfigFCMService struct {
	Logger              logger.Logger                `validate:"required"`
	ResponseConstructor response.HTTPRespConstructor `validate:"required"`
	ResponseWriter      response.Writer              `validate:"required"`
	AppService          appservice.Service           `validate:"required"`
}

type HandlerFCMService struct {
	Config ConfigFCMService
}

func NewHandlerFCMService(conf ConfigFCMService) (*HandlerFCMService, error) {
	err := validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	return &HandlerFCMService{Config: conf}, nil
}

func (h *HandlerFCMService) List() func(http.ResponseWriter, *http.Request) {
	type Request struct {
		Body appservice.GetFcmSvcAccKeyIn
	}

	type Response struct {
		appservice.GetFcmSvcAccKeyOut
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		clientID := r.Header.Get(ClientIDHeaderKey)
		out, err := h.Config.AppService.GetFcmSvcAccKey(ctx, appservice.GetFcmSvcAccKeyIn{
			ClientID: clientID,
		})
		if err != nil {
			resp := h.Config.ResponseConstructor.HTTPError(ctx, response.ErrUnhandled, err)
			h.Config.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
			return
		}

		respBody := Response{
			GetFcmSvcAccKeyOut: out,
		}

		resp := h.Config.ResponseConstructor.HTTPSuccess(ctx, respBody)
		h.Config.ResponseWriter.JSON(http.StatusOK, w, r, resp)
		return
	}
}

func (h *HandlerFCMService) Upload() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		clientID := r.PostFormValue("client_id")
		file, _, err := r.FormFile("file")
		if err != nil {
			resp := h.Config.ResponseConstructor.HTTPError(ctx, response.ErrValidation, err)
			h.Config.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
			return
		}

		var fcmServiceAccountKey fcmsvcacckeyrepo.ServiceAccountKey
		dec := json.NewDecoder(file)
		dec.DisallowUnknownFields()
		err = dec.Decode(&fcmServiceAccountKey)
		if err != nil {
			resp := h.Config.ResponseConstructor.HTTPError(ctx, response.ErrValidation, err)
			h.Config.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
			return
		}

		out, err := h.Config.AppService.CreateFcmSvcAccKey(ctx, appservice.CreateFcmSvcAccKeyIn{
			ClientID:             clientID,
			FCMServiceAccountKey: fcmServiceAccountKey,
		})

		if err != nil {
			resp := h.Config.ResponseConstructor.HTTPError(ctx, response.ErrUnhandled, err)
			h.Config.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
			return
		}

		resp := h.Config.ResponseConstructor.HTTPSuccess(ctx, out)
		h.Config.ResponseWriter.JSON(http.StatusOK, w, r, resp)
		return
	}
}
