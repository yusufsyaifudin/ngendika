package handlermsg

import (
	"github.com/segmentio/encoding/json"
	"github.com/yusufsyaifudin/ngendika/internal/svc/msgsvc"
	"github.com/yusufsyaifudin/ngendika/pkg/respbuilder"
	"github.com/yusufsyaifudin/ngendika/pkg/tracer"
	"github.com/yusufsyaifudin/ngendika/pkg/validator"
	"github.com/yusufsyaifudin/ngendika/transport/restapi/httptyped"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

type HandlerConfig struct {
	MsgServiceProcessor msgsvc.Service `validate:"required"`
}

type Handler struct {
	Config HandlerConfig
}

func NewHandler(cfg HandlerConfig) (*Handler, error) {
	err := validator.Validate(cfg)
	if err != nil {
		return nil, err
	}

	return &Handler{Config: cfg}, nil
}

type SendMessageReq struct {
	TaskID   string                   `json:"task_id"`
	ClientID string                   `json:"client_id"`
	Label    string                   `json:"label"`
	Payloads map[string][]interface{} `json:"payloads"` // fcm:[{}, {}]
}

type SendMessageResp struct {
	TaskID  string              `json:"task_id"`
	App     httptyped.AppEntity `json:"app"`
	Errors  []string            `json:"errors,omitempty"`
	Reports any                 `json:"reports,omitempty"`
}

func (h *Handler) SendMessage() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var span trace.Span
		ctx, span = tracer.StartSpan(ctx, "handlermsg.SendMessage")
		defer span.End()

		var reqBody SendMessageReq
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		err := dec.Decode(&reqBody)
		if err != nil {
			resp := respbuilder.Error(ctx, respbuilder.ErrValidation, err)
			respbuilder.WriteJSON(http.StatusBadRequest, w, r, resp)
			return
		}

		processMsgIn := &msgsvc.InputProcess{
			TaskID:   reqBody.TaskID,
			ClientID: reqBody.ClientID,
			Label:    reqBody.Label,
			Payloads: reqBody.Payloads,
		}

		processMsgOut, processMsgErr := h.Config.MsgServiceProcessor.Process(ctx, processMsgIn)
		if processMsgErr != nil {
			resp := respbuilder.Error(ctx, respbuilder.ErrUnhandled, processMsgErr)
			respbuilder.WriteJSON(http.StatusOK, w, r, resp)
			return
		}

		respBody := SendMessageResp{
			TaskID:  processMsgOut.TaskID,
			App:     httptyped.AppEntityFromSvc(processMsgOut.App),
			Errors:  processMsgOut.Errors,
			Reports: processMsgOut.ReportGroup,
		}

		resp := respbuilder.Success(ctx, respBody)
		respbuilder.WriteJSON(http.StatusOK, w, r, resp)
		return
	}
}
