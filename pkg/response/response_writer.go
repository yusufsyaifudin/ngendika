package response

import (
	"net/http"

	"github.com/segmentio/encoding/json"
)

type Writer interface {
	JSON(httpStatus int, rw http.ResponseWriter, r *http.Request, data interface{})
}

type writerImpl struct{}

func (w *writerImpl) JSON(httpStatus int, rw http.ResponseWriter, r *http.Request, data interface{}) {
	tracer := MustExtract(r.Context())

	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Tracer-ID", tracer.AppTraceID)
	rw.WriteHeader(httpStatus)

	enc := json.NewEncoder(rw)
	err := enc.Encode(data)
	if err != nil {
		reason := ReasonMap[ErrValidation]
		errPayload, _ := json.Marshal(HTTPError{
			Err: Error{
				Code:    reason.Code,
				Message: reason.Message,
				Debug:   err.Error(),
				TraceID: tracer.AppTraceID,
			},
		})

		_, _ = rw.Write(errPayload)
		return
	}

	return
}

var _ Writer = (*writerImpl)(nil)

func New() *writerImpl {
	return &writerImpl{}
}
