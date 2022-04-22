package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/go-playground/validator/v10"
	"github.com/segmentio/encoding/json"
	"github.com/yusufsyaifudin/ngendika/internal/logic/fcmservice"
	"github.com/yusufsyaifudin/ngendika/pkg/response"
	"github.com/yusufsyaifudin/ylog"
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
				ylog.Error(ctx, fmt.Sprintf("error unmarshal fcm service account %s", v.ID),
					ylog.KV("error", _err),
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

		clientID := chi.RouteContext(ctx).URLParam("client_id")

		// default Go 32MB https://cs.opensource.google/go/go/+/refs/tags/go1.18:src/net/http/request.go;l=1376
		file, fileHeader, err := r.FormFile("file")
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
				ylog.Error(ctx, "failed close uploaded file", ylog.KV("error", _err))
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
			ylog.Error(ctx, fmt.Sprintf("error unmarshal fcm service account %s", out.ID),
				ylog.KV("error", err),
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
