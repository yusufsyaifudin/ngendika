package appservice

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/yusufsyaifudin/ngendika/pkg/cache"
	"github.com/yusufsyaifudin/ylog"
	"time"
)

type CachedServiceConfig struct {
	ServicePersistent Service       `validate:"required"`
	CacheExpiry       time.Duration `validate:"required"`
	CachePrefixKey    string        `validate:"required,alphanumeric"`
	Cache             cache.Cache   `validate:"required"`
}

type CachedService struct {
	cfg CachedServiceConfig
}

var _ Service = (*CachedService)(nil)

func NewCached(cfg CachedServiceConfig) (Service, error) {
	if err := validator.New().Struct(cfg); err != nil {
		return nil, err
	}

	return &CachedService{
		cfg: cfg,
	}, nil
}

func (c *CachedService) CreateApp(ctx context.Context, input InputCreateApp) (out OutCreateApp, err error) {
	err = validator.New().Struct(input)
	if err != nil {
		err = fmt.Errorf("validation error, missing required field: %w", err)
		return
	}

	// Get cached data, if exists then return error
	existingApp, err := c.getByClientID(ctx, input.ClientID)
	if err != nil {
		// log and then discard error
		ylog.Error(ctx, "get app by id from cache error, continuing to try to insert", ylog.KV("error", err))
		err = nil
	}

	if existingApp.ID != "" {
		err = fmt.Errorf("app with client id '%s' already exist", existingApp.ClientID)
		return
	}

	// persist to db
	out, err = c.cfg.ServicePersistent.CreateApp(ctx, input)
	if err != nil {
		err = fmt.Errorf("persist app to db error: %w", err)
		return OutCreateApp{}, err
	}

	// if ok, save to cache
	c.setByClientID(ctx, out.App)
	return
}

func (c *CachedService) GetApp(ctx context.Context, input InputGetApp) (out OutGetApp, err error) {
	// Get from cache first
	app, err := c.getByClientID(ctx, input.ClientID)
	if err == nil && app.ID != "" {
		return
	}

	// If error occurred, then try get from persistent storage
	if err != nil {
		ylog.Info(ctx, fmt.Sprintf("app client id %s error get from cache", input.ClientID), ylog.KV("error", err))
		err = nil
	}

	out, err = c.cfg.ServicePersistent.GetApp(ctx, input)
	if err != nil {
		err = fmt.Errorf("persistence storage fetch error: %w", err)
		return
	}

	// Try cache, only log when error
	c.setByClientID(ctx, app)
	return
}

// -- cache

func (c *CachedService) cacheKeyByClientID(clientID string) string {
	return fmt.Sprintf("%s:%s", c.cfg.CachePrefixKey, clientID)
}

func (c *CachedService) getByClientID(ctx context.Context, clientID string) (App, error) {
	var app App
	err := c.cfg.Cache.GetAs(ctx, c.cacheKeyByClientID(clientID), &app)
	if err != nil {
		return App{}, err
	}

	ylog.Debug(ctx, fmt.Sprintf("get app id %s from cache", clientID))
	return app, nil
}

func (c *CachedService) setByClientID(ctx context.Context, app App) {
	err := c.cfg.Cache.SetExp(ctx, c.cacheKeyByClientID(app.ClientID), app, c.cfg.CacheExpiry)
	if err != nil {
		ylog.Error(ctx, fmt.Sprintf("cannot save cache app id %s", app.ClientID), ylog.KV("error", err))
		return
	}

	ylog.Debug(ctx, fmt.Sprintf("caching app id %s", app.ClientID))
	return
}
