package http

import (
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/segmentio/encoding/json"
	"github.com/yusufsyaifudin/ngendika/internal/logic/appservice"
	"github.com/yusufsyaifudin/ngendika/pkg/response"
)

type HandlerAppServiceConfig struct {
	ResponseConstructor response.HTTPRespConstructor `validate:"required"`
	ResponseWriter      response.Writer              `validate:"required"`
	AppService          appservice.Service           `validate:"required"`
}

type HandlerAppService struct {
	Config HandlerAppServiceConfig
}

func NewHandlerAppService(conf HandlerAppServiceConfig) (*HandlerAppService, error) {
	err := validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	return &HandlerAppService{Config: conf}, nil
}

// CreateApp .
func (h *HandlerAppService) CreateApp() func(http.ResponseWriter, *http.Request) {
	type Request struct {
		ClientID string `json:"client_id" validate:"required"`
		Name     string `json:"name" validate:"required"`
	}

	type Response struct {
		ID        string    `json:"id"`
		ClientID  string    `json:"client_id"`
		Name      string    `json:"name"`
		Enabled   bool      `json:"enabled"`
		CreatedAt time.Time `json:"created_at"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var reqBody Request
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		err := dec.Decode(&reqBody)
		if err != nil {
			resp := h.Config.ResponseConstructor.HTTPError(ctx, response.ErrValidation, err)
			h.Config.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
			return
		}

		out, err := h.Config.AppService.CreateApp(ctx, appservice.CreateAppIn{
			ClientID: reqBody.ClientID,
			Name:     reqBody.Name,
		})
		if err != nil {
			resp := h.Config.ResponseConstructor.HTTPError(ctx, response.ErrUnhandled, err)
			h.Config.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
			return
		}

		respBody := Response{
			ID:        out.App.ID,
			ClientID:  out.App.ClientID,
			Name:      out.App.Name,
			Enabled:   out.App.Enabled,
			CreatedAt: out.App.CreatedAt,
		}

		resp := h.Config.ResponseConstructor.HTTPSuccess(ctx, respBody)
		h.Config.ResponseWriter.JSON(http.StatusOK, w, r, resp)
		return
	}
}
