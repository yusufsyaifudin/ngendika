package http

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/segmentio/encoding/json"
	"github.com/yusufsyaifudin/ngendika/internal/logic/appservice"
	"github.com/yusufsyaifudin/ngendika/pkg/response"
)

type ConfigAppService struct {
	ResponseConstructor response.HTTPRespConstructor `validate:"required"`
	ResponseWriter      response.Writer              `validate:"required"`
	AppService          appservice.Service           `validate:"required"`
}

type HandlerAppService struct {
	Config ConfigAppService
}

func NewHandlerAppService(conf ConfigAppService) (*HandlerAppService, error) {
	err := validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	return &HandlerAppService{Config: conf}, nil
}

// CreateApp .
func (h *HandlerAppService) CreateApp() func(http.ResponseWriter, *http.Request) {
	type Request struct {
		Body appservice.CreateAppIn
	}

	type Response struct {
		appservice.CreateAppOut
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var reqBody Request
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		err := dec.Decode(&reqBody.Body)
		if err != nil {
			resp := h.Config.ResponseConstructor.HTTPError(ctx, response.ErrValidation, err)
			h.Config.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
			return
		}

		out, err := h.Config.AppService.CreateApp(ctx, reqBody.Body)
		if err != nil {
			resp := h.Config.ResponseConstructor.HTTPError(ctx, response.ErrUnhandled, err)
			h.Config.ResponseWriter.JSON(http.StatusBadRequest, w, r, resp)
			return
		}

		respBody := Response{
			CreateAppOut: out,
		}

		resp := h.Config.ResponseConstructor.HTTPSuccess(ctx, respBody)
		h.Config.ResponseWriter.JSON(http.StatusOK, w, r, resp)
		return
	}
}
