package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httptracer"
	"github.com/go-playground/validator/v10"
	"github.com/opentracing/opentracing-go"
	"github.com/satori/uuid"
	"github.com/yusufsyaifudin/ngendika/internal/logic/appservice"
	"github.com/yusufsyaifudin/ngendika/internal/logic/msgservice"
	"github.com/yusufsyaifudin/ngendika/pkg/logger"
	"github.com/yusufsyaifudin/ngendika/pkg/response"
	"github.com/yusufsyaifudin/ngendika/pkg/uid"
)

type Config struct {
	DebugError        bool
	UID               uid.UID            `validate:"required"`
	Log               logger.Logger      `validate:"required"`
	AppService        appservice.Service `validate:"required"`
	MessageProcessor  msgservice.Service `validate:"required"`
	MessageDispatcher msgservice.Service `validate:"required"`
}

type defaultHTTP struct {
	router *chi.Mux
}

func NewHTTPTransport(config Config) (*defaultHTTP, error) {
	if err := validator.New().Struct(config); err != nil {
		return nil, fmt.Errorf("http transport config error: %w", err)
	}

	handlerApp := &HandlerAppService{
		Logger:              config.Log,
		ResponseConstructor: response.NewResponseConstructor(config.DebugError),
		ResponseWriter:      response.New(),
		AppService:          config.AppService,
	}

	handlerPN := &HandlerMessageService{
		Logger:               config.Log,
		UID:                  config.UID,
		ResponseConstructor:  response.NewResponseConstructor(config.DebugError),
		ResponseWriter:       response.New(),
		MsgServiceDispatcher: config.MessageDispatcher,
		MsgServiceProcessor:  config.MessageProcessor,
	}

	if err := validator.New().Struct(handlerApp); err != nil {
		return nil, fmt.Errorf("http transport HandlerAppService error: %w", err)
	}

	router := chi.NewRouter()

	// TODO: add open telemetry here
	router.Use(httptracer.Tracer(opentracing.NoopTracer{}, httptracer.Config{
		ServiceName:    "ngendika",
		ServiceVersion: "v0.1.0",
		SampleRate:     1,
		SkipFunc: func(r *http.Request) bool {
			return r.URL.Path == "/health"
		},
		Tags: map[string]interface{}{},
	}))

	// add trace FCMKeyID and also log request response
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t1 := time.Now().UTC()
			ctx := r.Context()

			traceID := uuid.NewV4().String()
			loggerTracer := logger.Tracer{
				RemoteAddr: r.RemoteAddr,
				AppTraceID: traceID,
			}

			responseTracer := response.Tracer{
				RemoteAddr: r.RemoteAddr,
				AppTraceID: traceID,
			}

			ctx = Inject(ctx, loggerTracer, responseTracer)
			r = r.WithContext(ctx)

			reqBody := make([]byte, 0)
			var reqBodyData interface{}
			if r.Body != nil {
				reqBody, _ = ioutil.ReadAll(r.Body)
				r.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))
			}

			if _err := json.Unmarshal(reqBody, &reqBodyData); _err != nil {
				reqBodyData = string(reqBody)
			}

			// continue serve, and record the response
			rec := httptest.NewRecorder()
			next.ServeHTTP(rec, r)

			// read, copy, restore
			respBody := make([]byte, 0)
			if rec.Result().Body != nil {
				respBody, _ = ioutil.ReadAll(rec.Result().Body)
				rec.Result().Body = ioutil.NopCloser(bytes.NewBuffer(respBody))
			}

			var respBodyData interface{}
			if _err := json.Unmarshal(respBody, &respBodyData); _err != nil {
				respBodyData = string(respBody)
			}

			for k, v := range rec.Result().Header {
				w.Header()[k] = v
			}

			w.WriteHeader(rec.Code)
			_, err := bytes.NewReader(respBody).WriteTo(w)
			if err != nil {
				err = fmt.Errorf("write response body error: %w", err)

				logger.Access(ctx, logger.AccessLogData{
					Path:        r.URL.Path,
					ReqBody:     reqBodyData,
					RespBody:    respBodyData,
					Error:       err.Error(),
					ElapsedTime: time.Since(t1).Milliseconds(),
				})

				return
			}

			// log request
			logger.Access(ctx, logger.AccessLogData{
				Path:        r.RequestURI,
				ReqBody:     reqBodyData,
				RespBody:    respBodyData,
				ElapsedTime: time.Since(t1).Milliseconds(),
			})
		})
	})

	router.Route("/apps", func(r chi.Router) {
		r.Post("/", handlerApp.CreateApp)                    // create apps
		r.Get("/", nil)                                      // list of apps
		r.Delete("/", nil)                                   // delete apps
		r.Post("/fcm", handlerApp.PutFCMServiceAccountKey()) // add fcm cert
		r.Delete("/fcm/{fcm_id}", nil)                       // delete fcm cert
	})

	router.Route("/messages", func(r chi.Router) {
		r.Post("/", handlerPN.SendMessage) // send message
	})

	return &defaultHTTP{router: router}, nil
}

// Server .
func (a *defaultHTTP) Server() http.Handler {
	return a.router
}
