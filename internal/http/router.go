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
	"github.com/yusufsyaifudin/ngendika/internal/logic/fcmservice"
	"github.com/yusufsyaifudin/ngendika/internal/logic/msgservice"
	"github.com/yusufsyaifudin/ngendika/pkg/logger"
	"github.com/yusufsyaifudin/ngendika/pkg/response"
	"github.com/yusufsyaifudin/ngendika/pkg/uid"
)

const (
	ClientIDHeaderKey = "Client-ID"
)

type Config struct {
	AppServiceName string `validate:"required"`
	AppVersion     string `validate:"required"`

	DebugError        bool
	UID               uid.UID            `validate:"required"`
	AppService        appservice.Service `validate:"required"`
	FCMService        fcmservice.Service `validate:"required"`
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

	respConstructor := response.NewResponseConstructor(config.DebugError)
	respWriter := response.New()

	handlerApp, err := NewHandlerAppService(HandlerAppServiceConfig{
		ResponseConstructor: respConstructor,
		ResponseWriter:      respWriter,
		AppService:          config.AppService,
	})
	if err != nil {
		return nil, err
	}

	handlerFCM, err := NewHandlerFCMService(HandlerFCMServiceConfig{
		ResponseConstructor: respConstructor,
		ResponseWriter:      respWriter,
		FCMService:          config.FCMService,
	})

	if err != nil {
		return nil, err
	}

	handlerMessage, err := NewHandlerMessageService(HandlerMessageServiceConfig{
		UID:                  config.UID,
		ResponseConstructor:  respConstructor,
		ResponseWriter:       respWriter,
		MsgServiceDispatcher: config.MessageDispatcher,
		MsgServiceProcessor:  config.MessageProcessor,
	})

	if err != nil {
		return nil, err
	}

	router := chi.NewRouter()

	// TODO: add open telemetry here
	router.Use(httptracer.Tracer(opentracing.NoopTracer{}, httptracer.Config{
		ServiceName:    config.AppServiceName,
		ServiceVersion: config.AppVersion,
		SampleRate:     1,
		SkipFunc: func(r *http.Request) bool {
			return r.URL.Path == "/health"
		},
		Tags: map[string]interface{}{},
	}))

	// add trace FCMKeyID and also log request response
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var err error
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
				reqBody, err = ioutil.ReadAll(r.Body)
				if err != nil {
					logger.Error(ctx, "error read request body", logger.KV("error", err))
					reqBody = []byte(``)
				}

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
				respBody, err = ioutil.ReadAll(rec.Result().Body)
				if err != nil {
					logger.Error(ctx, "error read response body", logger.KV("error", err))
					respBody = []byte(``)
				}

				rec.Result().Body = ioutil.NopCloser(bytes.NewBuffer(respBody))
			}

			var respBodyData interface{} = map[string]interface{}{}
			if _err := json.Unmarshal(respBody, &respBodyData); _err != nil {
				respBodyData = map[string]interface{}{
					"raw_response_body": respBodyData,
					"error":             _err,
				}
			}

			for k, v := range rec.Result().Header {
				w.Header()[k] = v
			}

			w.WriteHeader(rec.Code)
			_, err = bytes.NewReader(respBody).WriteTo(w)
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

	todoHandler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"todo": true}`))
	}

	router.Route("/apps", func(r chi.Router) {
		r.Post("/", handlerApp.CreateApp())   // create apps
		r.Get("/", todoHandler)               // list of apps
		r.Put("/{client_id}", todoHandler)    // modify some field in apps (support patching)
		r.Delete("/{client_id}", todoHandler) // delete apps

		r.Get("/fcm", handlerFCM.List())          // get all fcm cert
		r.Get("/fcm/{fcm_id}", handlerFCM.List()) // get one fcm cert
		r.Post("/fcm", handlerFCM.Upload())       // add fcm cert
		r.Delete("/fcm", nil)                     // delete fcm cert, array of id

		r.Get("/apns", todoHandler)    // get apns cert
		r.Post("/apns", todoHandler)   // add apns cert
		r.Delete("/apns", todoHandler) // delete apns cert
	})

	router.Route("/messages", func(r chi.Router) {
		r.Post("/", handlerMessage.SendMessage) // send message
	})

	instance := &defaultHTTP{
		router: router,
	}

	return instance, nil
}

// Server .
func (a *defaultHTTP) Server() http.Handler {
	return a.router
}
