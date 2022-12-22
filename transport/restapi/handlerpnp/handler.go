package handlerpnp

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/schema"
	"github.com/yusufsyaifudin/ngendika/backend"
	"github.com/yusufsyaifudin/ngendika/internal/svc/appsvc"
	"github.com/yusufsyaifudin/ngendika/internal/svc/pnpsvc"
	"github.com/yusufsyaifudin/ngendika/pkg/respbuilder"
	"github.com/yusufsyaifudin/ngendika/pkg/validator"
	"github.com/yusufsyaifudin/ngendika/transport/restapi/httptyped"
	"github.com/yusufsyaifudin/ylog"
	"net/http"
)

type HandlerConfig struct {
	AppService         appsvc.Service `validate:"required"`
	ServiceProviderSvc pnpsvc.Service `validate:"required"`
}

type Handler struct {
	Config HandlerConfig
}

func NewHandler(conf HandlerConfig) (*Handler, error) {
	err := validator.Validate(conf)
	if err != nil {
		return nil, err
	}

	return &Handler{Config: conf}, nil
}

type CreateReqQueryParam struct {
	ClientID string `schema:"client_id"`
}

type CreateReq struct {
	BackendConfig struct {
		Provider       string      `json:"provider" validate:"required"`
		Label          string      `json:"label" validate:"required"`
		CredentialJSON interface{} `json:"credential_json" validate:"required"`
	} `json:"backend_config"`
}

type CreateResp struct {
	App           httptyped.AppEntity              `json:"app"`
	BackendConfig backend.PushNotificationProvider `json:"backend_config"`
}

func (h *Handler) Create() func(http.ResponseWriter, *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			err = fmt.Errorf("failed parse form: %w", err)
			resp := respbuilder.Error(ctx, respbuilder.ErrUnhandled, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		query := CreateReqQueryParam{}
		queryDec := schema.NewDecoder()
		err = queryDec.Decode(&query, r.Form)
		if err != nil {
			err = fmt.Errorf("failed decode query params: %w", err)
			resp := respbuilder.Error(ctx, respbuilder.ErrUnhandled, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		if r.Body == nil {
			err := fmt.Errorf("request body is nil")
			resp := respbuilder.Error(ctx, respbuilder.ErrValidation, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		defer func() {
			if _err := r.Body.Close(); _err != nil {
				ylog.Error(ctx, "cannot close request body", ylog.KV("error", _err))
			}
		}()

		var reqBody CreateReq
		dec := json.NewDecoder(r.Body)
		err = dec.Decode(&reqBody)
		if err != nil {
			resp := respbuilder.Error(ctx, respbuilder.ErrValidation, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		enabled := true
		getAppIn := appsvc.InputGetApp{
			ClientID: query.ClientID,
			Enabled:  &enabled,
		}

		getAppOut, err := h.Config.AppService.GetApp(ctx, getAppIn)
		if err != nil {
			resp := respbuilder.Error(ctx, respbuilder.ErrUnhandled, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		inSvcProvider := pnpsvc.InCreate{
			AppID:      getAppOut.App.ID,
			PnProvider: pnpsvc.InCreatePnProvider(reqBody.BackendConfig),
		}

		outSvcProvider, err := h.Config.ServiceProviderSvc.Create(ctx, inSvcProvider)
		if err != nil {
			resp := respbuilder.Error(ctx, respbuilder.ErrUnhandled, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		respData := CreateResp{
			App:           httptyped.AppEntityFromSvc(getAppOut.App),
			BackendConfig: outSvcProvider.ServiceProvider,
		}

		resp := respbuilder.Success(ctx, respData)
		respbuilder.WriteJSON(http.StatusCreated, w, r, resp)
		return
	}

	return fn
}

type ListByProviderReq struct {
	ClientID string `schema:"client_id"`
	Provider string `schema:"provider"`
	Label    string `schema:"label"`
}

type ListByProviderResp struct {
	Items []backend.PushNotificationProvider `json:"items"`
}

func (h *Handler) ListByProvider() func(http.ResponseWriter, *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			err = fmt.Errorf("failed parse form: %w", err)
			resp := respbuilder.Error(ctx, respbuilder.ErrUnhandled, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		query := ListByProviderReq{}
		queryDec := schema.NewDecoder()
		err = queryDec.Decode(&query, r.Form)
		if err != nil {
			err = fmt.Errorf("failed decode query params: %w", err)
			resp := respbuilder.Error(ctx, respbuilder.ErrUnhandled, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		enabled := true
		getAppIn := appsvc.InputGetApp{
			ClientID: query.ClientID,
			Enabled:  &enabled,
		}

		getAppOut, err := h.Config.AppService.GetApp(ctx, getAppIn)
		if err != nil {
			resp := respbuilder.Error(ctx, respbuilder.ErrUnhandled, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		fetchOut, err := h.Config.ServiceProviderSvc.GetByLabels(ctx, pnpsvc.InGetByLabels{
			AppID:    getAppOut.App.ID,
			Provider: query.Provider,
			Label:    query.Label,
		})
		if err != nil {
			resp := respbuilder.Error(ctx, respbuilder.ErrUnhandled, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		respData := ListByProviderResp{
			Items: fetchOut.PnProviders,
		}

		resp := respbuilder.Success(ctx, respData)
		respbuilder.WriteJSON(http.StatusCreated, w, r, resp)
		return
	}

	return fn
}

type ExamplesResp struct {
	Items []backend.Example `json:"items"`
}

func (h *Handler) Examples() func(http.ResponseWriter, *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		examples := h.Config.ServiceProviderSvc.Examples(ctx)
		respData := ExamplesResp{
			Items: examples.Items,
		}

		resp := respbuilder.Success(ctx, respData)
		respbuilder.WriteJSON(http.StatusCreated, w, r, resp)
		return
	}

	return fn
}
