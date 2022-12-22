package restapi

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/yusufsyaifudin/ngendika/assets"
	"github.com/yusufsyaifudin/ngendika/internal/svc/appsvc"
	"github.com/yusufsyaifudin/ngendika/internal/svc/msgsvc"
	"github.com/yusufsyaifudin/ngendika/internal/svc/pnpsvc"
	"github.com/yusufsyaifudin/ngendika/pkg/tracer"
	"github.com/yusufsyaifudin/ngendika/pkg/validator"
	"github.com/yusufsyaifudin/ngendika/transport/restapi/handlerapp"
	"github.com/yusufsyaifudin/ngendika/transport/restapi/handlermsg"
	"github.com/yusufsyaifudin/ngendika/transport/restapi/handlerpnp"
	"go.opentelemetry.io/otel"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

type Config struct {
	AppServiceName string         `validate:"required"`
	AppVersion     string         `validate:"required"`
	AppService     appsvc.Service `validate:"required"`
	PNPService     pnpsvc.Service `validate:"required"`
	MsgService     msgsvc.Service `validate:"required"`
}

type DefaultHTTP struct {
	router *chi.Mux
}

func NewHTTPTransport(cfg Config) (*DefaultHTTP, error) {
	if err := validator.Validate(cfg); err != nil {
		return nil, fmt.Errorf("http transport cfg error: %w", err)
	}

	// ** Application handler
	handlerAppCfg := handlerapp.HandlerConfig{
		AppService: cfg.AppService,
	}

	handlerApp, err := handlerapp.NewHandler(handlerAppCfg)
	if err != nil {
		return nil, err
	}

	// ** Service provider handler
	handlSvcProviderCfg := handlerpnp.HandlerConfig{
		AppService:         cfg.AppService,
		ServiceProviderSvc: cfg.PNPService,
	}
	handlSvcProvider, err := handlerpnp.NewHandler(handlSvcProviderCfg)
	if err != nil {
		return nil, err
	}

	// ** Messaging service handler
	handlerMsgCfg := handlermsg.HandlerConfig{
		MsgServiceProcessor: cfg.MsgService,
	}
	handlerMessage, err := handlermsg.NewHandler(handlerMsgCfg)

	if err != nil {
		return nil, err
	}

	router := chi.NewRouter()

	skip := func(r *http.Request) bool {
		switch strings.TrimSpace(path.Clean(r.URL.Path)) {
		case "/swaggerui",
			"/health",
			"/ping":
			return true
		}

		return false
	}

	router.Use(middleware.StripSlashes)

	router.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	router.Use(func(next http.Handler) http.Handler {
		return tracer.Middleware(tracer.MiddlewareConfig{
			TracerName:     "github.com/yusufsyaifudin/ngendika",
			ServiceName:    assets.ServiceName,
			SkipFunc:       skip,
			TracerProvider: otel.GetTracerProvider(),    // global tracer provider
			TextPropagator: otel.GetTextMapPropagator(), // use global text map propagator
		}, next)
	})

	// add trace id and also log request response
	router.Use(func(next http.Handler) http.Handler {
		return requestLogger(skip, next)
	})

	todoHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"todo": true}`))
	}

	swaggerDir, _ := fs.Sub(assets.SwaggerUI, ".")
	router.Mount("/", http.FileServer(http.FS(swaggerDir)))

	// Resource: apps
	router.Route("/api/v1/apps", func(r chi.Router) {
		r.Post("/", handlerApp.CreateApp())                     // create apps
		r.Get("/", handlerApp.ListApps())                       // list of apps
		r.Get("/{client_id}", handlerApp.GetByClientID())       // list of apps
		r.Put("/{client_id}", handlerApp.PutApp())              // replace all existing field in apps (does not support patching)
		r.Delete("/{client_id}", handlerApp.DelAppByClientID()) // delete apps
	})

	// Resource: service providers
	router.Route("/api/v1/pnp", func(r chi.Router) {
		r.Post("/", handlSvcProvider.Create())                        // create new
		r.Put("/{label}", todoHandler)                                // create or replace entirely
		r.Get("/examples", handlSvcProvider.Examples())               // create or replace entirely
		r.Get("/list/by-provider", handlSvcProvider.ListByProvider()) // get list under this client_id
		r.Get("/{label}", todoHandler)                                // get one
		r.Delete("/{label}", todoHandler)                             // delete one
	})

	// Resource: messages
	router.Route("/api/v1/messages", func(r chi.Router) {
		r.Post("/", handlerMessage.SendMessage()) // send message
	})

	instance := &DefaultHTTP{
		router: router,
	}

	return instance, nil
}

// Server .
func (a *DefaultHTTP) Server() http.Handler {
	return a.router
}
