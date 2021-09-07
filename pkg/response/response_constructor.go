package response

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type DefaultHTTPRespConstructor struct {
	Debug  bool // set whether return debug message or not
	ErrMap map[interface{}]string
}

func (g *DefaultHTTPRespConstructor) HTTPError(ctx context.Context, reasonKind ErrKind, err error) HTTPError {
	stuff := MustExtract(ctx)

	errMsg := ""
	if err != nil {
		var errJson *json.UnmarshalTypeError
		if errors.As(err, &errJson) {
			// TODO: handle specific type error here
		}

		errMsg = err.Error()
	}

	err = fmt.Errorf("%v %w", err, ErrX)

	reason, ok := ReasonMap[reasonKind]
	if !ok {
		return HTTPError{
			Err: Error{
				Code:    "XX",
				Message: "unknown error kind",
				Debug:   "", // don't show message if unknown type, to prevent security breach
				TraceID: stuff.AppTraceID,
			},
		}
	}

	switch {
	case reasonKind == ErrUnhandled && !g.Debug:
		errMsg = "" // disable message if type ErrUnhandled and debug disabled
	default:
		reason.Message = fmt.Sprintf("%s: %s", reason.Message, errMsg)
	}

	return HTTPError{
		Err: Error{
			Code:    reason.Code,
			Message: reason.Message,
			Debug:   errMsg,
			TraceID: stuff.AppTraceID,
		},
	}
}

func (g *DefaultHTTPRespConstructor) HTTPSuccess(ctx context.Context, data interface{}) HTTPSuccess {
	stuff := MustExtract(ctx)

	return HTTPSuccess{
		TraceID: stuff.AppTraceID,
		Data:    data,
	}
}

var _ HTTPRespConstructor = (*DefaultHTTPRespConstructor)(nil)

func NewResponseConstructor(debug bool) *DefaultHTTPRespConstructor {
	return &DefaultHTTPRespConstructor{
		Debug: debug,
	}
}
