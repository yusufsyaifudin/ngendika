package handlerapp

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
	"github.com/segmentio/encoding/json"
	"github.com/yusufsyaifudin/ngendika/internal/svc/appsvc"
	"github.com/yusufsyaifudin/ngendika/pkg/respbuilder"
	"github.com/yusufsyaifudin/ngendika/pkg/validator"
	"github.com/yusufsyaifudin/ngendika/transport/restapi/httptyped"
	"github.com/yusufsyaifudin/ylog"
	"net/http"
	"strings"
	"unicode/utf8"
)

type HandlerConfig struct {
	AppService appsvc.Service `validate:"required"`
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

type CreateAppReq struct {
	ClientID string `json:"client_id"`
	Name     string `json:"name"`
}

type CreateAppResp struct {
	App httptyped.AppEntity `json:"app"`
}

// CreateApp CreateSvcAccKey new application using unique Client FCMServiceAccountKey.
// Path         : POST /api/v1/apps
// Request Body : CreateAppReq
// Response     : CreateAppResp
func (h *Handler) CreateApp() func(http.ResponseWriter, *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

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

		var reqBody CreateAppReq
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&reqBody)
		if err != nil {
			resp := respbuilder.Error(ctx, respbuilder.ErrValidation, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		createAppIn := appsvc.InputCreateApp{
			ClientID: reqBody.ClientID,
			Name:     reqBody.Name,
		}

		createAppOut, err := h.Config.AppService.CreateApp(ctx, createAppIn)
		if err != nil {
			resp := respbuilder.Error(ctx, respbuilder.ErrUnhandled, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		respBody := CreateAppResp{
			App: httptyped.AppEntityFromSvc(createAppOut.App),
		}

		resp := respbuilder.Success(ctx, respBody)
		respbuilder.WriteJSON(http.StatusCreated, w, r, resp)
	}

	return handler
}

type PutAppRequest struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type PutAppResp struct {
	App httptyped.AppEntity `json:"app"`
}

// PutApp Create or Replace application using unique Client FCMServiceAccountKey.
// Path         : PUT /api/v1/apps/{client_id}
// Request Body : PutAppRequest
// Response     : PutAppResp
func (h *Handler) PutApp() func(http.ResponseWriter, *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		clientID := strings.TrimSpace(chi.URLParam(r, "client_id"))
		if !utf8.ValidString(clientID) {
			err := fmt.Errorf("client id '%s' is not valid utf8", clientID)
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

		var reqBody PutAppRequest
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&reqBody)
		if err != nil {
			resp := respbuilder.Error(ctx, respbuilder.ErrValidation, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		putAppIn := appsvc.InputPutApp{
			ClientID: clientID,
			Name:     reqBody.Name,
		}

		putAppOut, err := h.Config.AppService.PutApp(ctx, putAppIn)
		if err != nil {
			resp := respbuilder.Error(ctx, respbuilder.ErrUnhandled, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		respBody := PutAppResp{
			App: httptyped.AppEntityFromSvc(putAppOut.App),
		}

		resp := respbuilder.Success(ctx, respBody)
		respbuilder.WriteJSON(http.StatusOK, w, r, resp)
	}

	return handler
}

type ListAppsReq struct {
	Limit int64 `schema:"limit"`
	MaxID int64 `schema:"max_id"`
	MinID int64 `schema:"min_id"`
}

type ListAppsResp struct {
	Total int64                 `json:"total"`
	Limit int64                 `json:"limit"`
	Items []httptyped.AppEntity `json:"items"`
}

// ListApps ListApp apps
// Path          : GET /api/v1/apps
// Request Query : ListAppsReq
// Response      : ListAppsResp
func (h *Handler) ListApps() func(http.ResponseWriter, *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := r.ParseForm()
		if err != nil {
			err = fmt.Errorf("failed parse form: %w", err)
			resp := respbuilder.Error(ctx, respbuilder.ErrUnhandled, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		query := ListAppsReq{}
		queryDec := schema.NewDecoder()
		err = queryDec.Decode(&query, r.Form)
		if err != nil {
			err = fmt.Errorf("failed decode query params: %w", err)
			resp := respbuilder.Error(ctx, respbuilder.ErrUnhandled, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		appListIn := appsvc.InputListApp{
			Limit:    query.Limit,
			BeforeID: query.MaxID,
			AfterID:  query.MinID,
		}

		listOut, err := h.Config.AppService.ListApp(ctx, appListIn)
		if err != nil {
			resp := respbuilder.Error(ctx, respbuilder.ErrUnhandled, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		apps := make([]httptyped.AppEntity, 0)
		for _, app := range listOut.Apps {
			apps = append(apps, httptyped.AppEntityFromSvc(app))
		}

		respBody := ListAppsResp{
			Total: listOut.Total,
			Limit: listOut.Limit,
			Items: apps,
		}

		resp := respbuilder.Success(ctx, respBody)
		respbuilder.WriteJSON(http.StatusOK, w, r, resp)
	}

	return handler
}

type GetAppByClientIDResp struct {
	App httptyped.AppEntity `json:"app"`
}

// GetByClientID Get one by client id
// Path          : GET /api/v1/apps/{client_id}
// Response      : GetAppByClientIDResp
func (h *Handler) GetByClientID() func(http.ResponseWriter, *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		clientID := strings.TrimSpace(chi.URLParam(r, "client_id"))
		if !utf8.ValidString(clientID) {
			err := fmt.Errorf("client id '%s' is not valid utf8", clientID)
			resp := respbuilder.Error(ctx, respbuilder.ErrUnhandled, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		enabled := true
		getAppIn := appsvc.InputGetApp{
			ClientID: clientID,
			Enabled:  &enabled,
		}

		getAppOut, err := h.Config.AppService.GetApp(ctx, getAppIn)
		if err != nil {
			resp := respbuilder.Error(ctx, respbuilder.ErrUnhandled, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		respBody := GetAppByClientIDResp{
			App: httptyped.AppEntityFromSvc(getAppOut.App),
		}

		resp := respbuilder.Success(ctx, respBody)
		respbuilder.WriteJSON(http.StatusOK, w, r, resp)
	}

	return handler
}

type DelAppByClientIDResp struct {
	Success bool `json:"success"`
}

// DelAppByClientID Delete one by client id
// Path          : DEL /api/v1/apps/{client_id}
// Response      : httptyped.DelAppByClientIDResp
func (h *Handler) DelAppByClientID() func(http.ResponseWriter, *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		clientID := strings.TrimSpace(chi.URLParam(r, "client_id"))
		if !utf8.ValidString(clientID) {
			err := fmt.Errorf("client id '%s' is not valid utf8", clientID)
			resp := respbuilder.Error(ctx, respbuilder.ErrUnhandled, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		delAppIn := appsvc.InputDelApp{
			ClientID: clientID,
		}

		delAppOut, err := h.Config.AppService.DelApp(ctx, delAppIn)
		if err != nil {
			resp := respbuilder.Error(ctx, respbuilder.ErrUnhandled, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		respBody := DelAppByClientIDResp{
			Success: delAppOut.Success,
		}

		resp := respbuilder.Success(ctx, respBody)
		respbuilder.WriteJSON(http.StatusOK, w, r, resp)
	}

	return handler
}
