package restapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/satori/uuid"
	"github.com/yusufsyaifudin/ngendika/pkg/tracer"
	"go.uber.org/multierr"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/yusufsyaifudin/ngendika/pkg/respbuilder"
	"github.com/yusufsyaifudin/ylog"
)

func toSimpleMap(h http.Header) map[string]string {
	out := map[string]string{}
	for k, v := range h {
		out[k] = strings.Join(v, " ")
	}

	return out
}

func requestLogger(skipFunc func(r *http.Request) bool, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if skipFunc(r) {
			next.ServeHTTP(w, r)
			return
		}

		var globalErr error
		t1 := time.Now().UTC()
		ctx := r.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		traceID := uuid.NewV4().String()

		propagateData := tracer.LogData{
			RemoteAddr: r.RemoteAddr,
			TraceID:    traceID,
		}

		var logTraceData *ylog.Tracer
		logTraceData, err := ylog.NewTracer(propagateData, ylog.WithTag("tracer"))
		if err != nil {
			// this should never happen, but once it happens, we need to log in the response
			globalErr = multierr.Append(globalErr, fmt.Errorf("error prepare log tracer data: %w", err))
		}

		responseTracer := respbuilder.Tracer{
			RemoteAddr: r.RemoteAddr,
			AppTraceID: traceID,
		}

		// Inject logger and response tracer at same time
		ctx = ylog.Inject(ctx, logTraceData)
		ctx = respbuilder.Inject(ctx, responseTracer)
		r = r.WithContext(ctx)

		reqBody := make([]byte, 0)
		if r.Body != nil {
			defer func() {
				if _err := r.Body.Close(); _err != nil {
					_err = fmt.Errorf("cannot close request body: %w", _err)
					globalErr = multierr.Append(globalErr, _err)
				}
			}()

			reqBody, err = io.ReadAll(r.Body)
			if err != nil {
				globalErr = multierr.Append(globalErr, fmt.Errorf("error read request body: %w", err))
				reqBody = []byte(``)
			}

			r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		var reqBodyStr = string(reqBody)
		var reqBodyObj interface{} = map[string]interface{}{}
		if _err := json.Unmarshal(reqBody, &reqBodyObj); _err != nil {
			globalErr = multierr.Append(globalErr, fmt.Errorf("error marshal request body: %w", _err))
		} else {
			reqBodyStr = "" // set to empty string if valid json payload
		}

		// continue serve, and record the response
		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r)

		// read, copy, restore
		respBody := make([]byte, 0)
		if rec.Result().Body != nil {
			respBody, err = io.ReadAll(rec.Result().Body)
			if err != nil {
				globalErr = multierr.Append(globalErr, fmt.Errorf("error read response body: %w", err))
				respBody = []byte(``)
			}

			rec.Result().Body = io.NopCloser(bytes.NewBuffer(respBody))
		}

		var respBodyStr = string(respBody)
		var respBodyData interface{}
		if _err := json.Unmarshal(respBody, &respBodyData); _err != nil {
			globalErr = multierr.Append(globalErr, fmt.Errorf("error marshal response body: %w", _err))
		} else {
			respBodyStr = "" // set to empty string if success as json object
		}

		for k, v := range rec.Header() {
			w.Header()[k] = v
		}

		w.WriteHeader(rec.Code)
		_, err = bytes.NewReader(respBody).WriteTo(w)
		if err != nil {
			globalErr = multierr.Append(globalErr, fmt.Errorf("error write response body: %w", err))
		}

		errStr := ""
		if globalErr != nil {
			errStr = globalErr.Error()
		}

		// log request
		ylog.Access(ctx, ylog.AccessLogData{
			Path: r.RequestURI,
			Request: ylog.HTTPData{
				Header:     toSimpleMap(r.Header),
				DataObject: reqBodyObj,
				DataString: reqBodyStr,
			},
			Response: ylog.HTTPData{
				Header:     toSimpleMap(rec.Header()),
				DataObject: respBodyData,
				DataString: respBodyStr,
			},
			Error:       errStr,
			ElapsedTime: time.Since(t1).Milliseconds(),
		})
	}
}
