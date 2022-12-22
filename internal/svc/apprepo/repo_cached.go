package apprepo

import (
	"context"
	"fmt"
	"github.com/yusufsyaifudin/ngendika/pkg/cache"
	"github.com/yusufsyaifudin/ngendika/pkg/validator"
	"github.com/yusufsyaifudin/ylog"
	"time"
)

type CachedConfig struct {
	Persistent     Repo          `validate:"required"`
	CacheExpiry    time.Duration `validate:"required"`
	CachePrefixKey string        `validate:"required,alphanumeric"`
	Cache          cache.Cache   `validate:"required"`
}

type CachedRepo struct {
	Config CachedConfig
}

var _ Repo = (*CachedRepo)(nil)

func NewCached(cfg CachedConfig) (*CachedRepo, error) {
	if err := validator.Validate(cfg); err != nil {
		return nil, err
	}

	return &CachedRepo{
		Config: cfg,
	}, nil
}

func (c *CachedRepo) Create(ctx context.Context, in InputCreate) (out OutCreate, err error) {
	// Get cached data, if exists then return in
	existingApp, err := c.getByClientID(ctx, in.App.ClientID)
	if err != nil {
		// log and then discard error
		ylog.Error(ctx, "get app by id from cache error, continuing to try to insert", ylog.KV("error", err))
		err = nil
	}

	if existingApp.ClientID == in.App.ClientID {
		err = fmt.Errorf("app with client id '%s' already exist", existingApp.ClientID)
		return
	}

	// persist to db
	out, err = c.Config.Persistent.Create(ctx, in)
	if err != nil {
		err = fmt.Errorf("persist app to db error: %w", err)
		return
	}

	// if ok, save to cache
	c.setByClientID(ctx, out.App)
	return
}

func (c *CachedRepo) Upsert(ctx context.Context, in InputUpsert) (out OutUpsert, err error) {
	err = validator.Validate(in)
	if err != nil {
		err = fmt.Errorf("validation error, missing required field: %w", err)
		return
	}

	out, err = c.Config.Persistent.Upsert(ctx, in)
	if err != nil {
		err = fmt.Errorf("upsert via cache error on persistent store: %w", err)
		return
	}

	// if ok, save to cache
	c.setByClientID(ctx, out.App)
	return
}

func (c *CachedRepo) GetByClientID(ctx context.Context, in InputGetByClientID) (out OutGetByClientID, err error) {
	// Get from cache first
	app, err := c.getByClientID(ctx, in.ClientID)
	if err == nil && app.ClientID == in.ClientID {
		return
	}

	// If error occurred, then try get from persistent storage
	if err != nil {
		ylog.Error(ctx, fmt.Sprintf("app client id %s error get from cache", in.ClientID), ylog.KV("error", err))
		err = nil
	}

	out, err = c.Config.Persistent.GetByClientID(ctx, in)
	if err != nil {
		err = fmt.Errorf("persistence storage fetch error: %w", err)
		return
	}

	// Try cache, only log when error
	c.setByClientID(ctx, app)
	return
}

// List of cached apps now will not use cache. It hard to maintain list in cache.
func (c *CachedRepo) List(ctx context.Context, in InputList) (out OutList, err error) {
	return c.Config.Persistent.List(ctx, in)
}

func (c *CachedRepo) DelByClientID(ctx context.Context, in InputDelByClientID) (out OutDelByClientID, err error) {
	out, err = c.Config.Persistent.DelByClientID(ctx, in)
	if err != nil {
		return
	}

	err = c.delByClientID(ctx, in.ClientID)
	return
}

// -- cache

func (c *CachedRepo) genCacheKeyByClientID(clientID string) string {
	return fmt.Sprintf("%s:%s", c.Config.CachePrefixKey, clientID)
}

func (c *CachedRepo) getByClientID(ctx context.Context, clientID string) (App, error) {
	var app App
	err := c.Config.Cache.GetAs(ctx, c.genCacheKeyByClientID(clientID), &app)
	if err != nil {
		return App{}, err
	}

	ylog.Debug(ctx, fmt.Sprintf("get app id %s from cache", clientID))
	return app, nil
}

func (c *CachedRepo) setByClientID(ctx context.Context, app App) {
	err := c.Config.Cache.SetExp(ctx, c.genCacheKeyByClientID(app.ClientID), app, c.Config.CacheExpiry)
	if err != nil {
		ylog.Error(ctx, fmt.Sprintf("cannot save cache app id %s", app.ClientID), ylog.KV("error", err))
		return
	}

	ylog.Debug(ctx, fmt.Sprintf("caching app id %s", app.ClientID))
	return
}

func (c *CachedRepo) delByClientID(ctx context.Context, clientID string) error {
	return c.Config.Cache.Delete(ctx, c.genCacheKeyByClientID(clientID))
}
