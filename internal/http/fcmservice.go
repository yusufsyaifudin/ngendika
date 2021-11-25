package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/segmentio/encoding/json"
	"github.com/yusufsyaifudin/ngendika/internal/logic/fcmservice"
	"github.com/yusufsyaifudin/ngendika/pkg/logger"
	"github.com/yusufsyaifudin/ngendika/pkg/response"
)

type HandlerFCMServiceConfig struct {
	ResponseConstructor response.HTTPRespConstructor `validate:"required"`
	ResponseWriter      response.Writer              `validate:"required"`
	FCMService          fcmservice.Service           `validate:"required"`
}

type HandlerFCMService struct {
	Config HandlerFCMServiceConfig
}

func NewHandlerFCMService(conf HandlerFCMServiceConfig) (*HandlerFCMService, error) {
	err := validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	return &HandlerFCMService{Config: conf}, nil
}

func (h *HandlerFCMService) List() func(http.ResponseWriter, *http.Request) {
	type Request struct {
		ClientID string `validate:"required"`
	}

	type FCMServiceAccount struct {
		ID                string      `json:"id"`
		CreatedAt         time.Time   `json:"created_at"`
		ServiceAccountKey interface{} `json:"service_account_key"` // in JSON format
	}

	type Response struct {
		FCMServiceAccounts []FCMServiceAccount `json:"fcm_service_accounts"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		clientID := r.Header.Get(ClientIDHeaderKey)
		out, err := h.Config.FCMService.GetSvcAccKey(ctx, fcmservice.GetSvcAccKeyIn{
			ClientID: clientID,
		})
		if err != nil {
			resp := h.Config.ResponseConstructor.HTTPError(ctx, response.ErrUnhandled, err)
			h.Config.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
			return
		}

		respBody := Response{
			FCMServiceAccounts: []FCMServiceAccount{},
		}

		for _, v := range out.Lists {
			var svcAccKey interface{}
			if _err := json.Unmarshal(v.ServiceAccountKey, &svcAccKey); _err != nil {
				logger.Error(ctx, fmt.Sprintf("error unmarshal fcm service account %s", v.ID),
					logger.KV("error", _err),
				)
				continue
			}

			respBody.FCMServiceAccounts = append(respBody.FCMServiceAccounts, FCMServiceAccount{
				ID:                v.ID,
				CreatedAt:         v.CreatedAt,
				ServiceAccountKey: svcAccKey,
			})
		}

		resp := h.Config.ResponseConstructor.HTTPSuccess(ctx, respBody)
		h.Config.ResponseWriter.JSON(http.StatusOK, w, r, resp)
		return
	}
}

func (h *HandlerFCMService) Upload() func(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		ID                string      `json:"id"`
		ServiceAccountKey interface{} `json:"service_account_key"`
		CreatedAt         time.Time   `json:"created_at"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		clientID := r.PostFormValue("client_id")
		file, fileHeader, err := r.FormFile("file") // default Go 32MB
		if err != nil {
			resp := h.Config.ResponseConstructor.HTTPError(ctx, response.ErrValidation, err)
			h.Config.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
			return
		}

		defer func() {
			if file == nil {
				return
			}

			if _err := file.Close(); _err != nil {
				logger.Error(ctx, "failed close uploaded file", logger.KV("error", _err))
			}
		}()

		fileBuffer := bytes.Buffer{}
		_, err = io.Copy(&fileBuffer, file)
		if err != nil {
			err = fmt.Errorf("error read file: %w", err)
			resp := h.Config.ResponseConstructor.HTTPError(ctx, response.ErrValidation, err)
			h.Config.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
			return
		}

		out, err := h.Config.FCMService.CreateSvcAccKey(ctx, fcmservice.CreateSvcAccKeyIn{
			ClientID:          clientID,
			ServiceAccountKey: fileBuffer.Bytes(),
			Metadata: fcmservice.SvcAccKeyMetadata{
				Filename: fileHeader.Filename,
			},
		})

		if err != nil {
			resp := h.Config.ResponseConstructor.HTTPError(ctx, response.ErrUnhandled, err)
			h.Config.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
			return
		}

		var fcmSvcAccKey interface{} = map[string]interface{}{} // default to empty object
		err = json.Unmarshal(out.ServiceAccountKey, &fcmSvcAccKey)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("error unmarshal fcm service account %s", out.ID),
				logger.KV("error", err),
			)

			err = nil // discard error and continue to response
		}

		respData := Response{
			ID:                out.ID,
			ServiceAccountKey: fcmSvcAccKey,
			CreatedAt:         out.CreatedAt,
		}

		resp := h.Config.ResponseConstructor.HTTPSuccess(ctx, respData)
		h.Config.ResponseWriter.JSON(http.StatusOK, w, r, resp)
		return
	}
}
