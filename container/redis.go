package container

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/yusufsyaifudin/ngendika/config"
	"github.com/yusufsyaifudin/ngendika/pkg/logger"
)

type RedisConnMaker struct {
	ctx           context.Context
	conf          config.Redis
	redisSingle   map[string]*redis.Client
	redisSentinel map[string]*redis.Client
	redisCluster  map[string]*redis.ClusterClient
	closer        []Closer
}

func NewRedisConnMaker(ctx context.Context, conf config.Redis) (*RedisConnMaker, error) {
	instance := &RedisConnMaker{
		ctx:           ctx,
		conf:          conf,
		redisSingle:   map[string]*redis.Client{},
		redisSentinel: map[string]*redis.Client{},
		redisCluster:  map[string]*redis.ClusterClient{},
		closer:        make([]Closer, 0),
	}

	err := instance.connect()
	if err != nil {
		// close previous opened connection if error happen
		if _err := instance.CloseAll(); _err != nil {
			err = fmt.Errorf("close redis error: %w: %s", err, _err)
		}

		return nil, err
	}

	return instance, nil
}

func (i *RedisConnMaker) connect() error {

	ctx := i.ctx

	for key, connInfo := range i.conf {
		key = strings.TrimSpace(strings.ToLower(key))
		if err := validator.New().Var(key, "required,alphanum"); err != nil {
			err = fmt.Errorf("error connecting to redis key '%s': %w", key, err)
			return err
		}

		var redisClient redis.UniversalClient
		switch connInfo.Mode {
		case "single":
			single := redis.NewClient(&redis.Options{
				Addr:     connInfo.Address[0],
				Username: connInfo.Username,
				Password: connInfo.Password,
				DB:       connInfo.DB,
			})

			i.redisSingle[key] = single
			redisClient = single

		case "sentinel":
			sentinel := redis.NewFailoverClient(&redis.FailoverOptions{
				SentinelAddrs: connInfo.Address,
				Username:      connInfo.Username,
				Password:      connInfo.Password,
				DB:            connInfo.DB,
				MasterName:    connInfo.MasterName,
			})

			i.redisSentinel[key] = sentinel
			redisClient = sentinel

		case "cluster":
			// cluster mode is not support DB selection
			cluster := redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:    connInfo.Address,
				Username: connInfo.Username,
				Password: connInfo.Password,
			})

			i.redisCluster[key] = cluster
			redisClient = cluster

		default:
			err := fmt.Errorf("unknown redis mode: %s", connInfo.Mode)
			return err
		}

		if redisClient == nil {
			return fmt.Errorf("redis client %s is nil", key)
		}

		err := redisClient.Ping(ctx).Err()
		if err != nil {
			err = fmt.Errorf("error ping redis %s: %w", key, err)
			return err
		}

		i.closer = append(i.closer, NewNamedCloser(key, redisClient)) // register the closer
	}

	return nil
}

func (i *RedisConnMaker) GetSingle(key string) (*redis.Client, error) {
	key = strings.TrimSpace(strings.ToLower(key))
	v, ok := i.redisSingle[key]
	if !ok {
		return nil, fmt.Errorf("key %s is not found on any redis with single architecture", key)
	}

	return v, nil
}

func (i *RedisConnMaker) GetSentinel(key string) (*redis.Client, error) {
	key = strings.TrimSpace(strings.ToLower(key))
	v, ok := i.redisSentinel[key]
	if !ok {
		return nil, fmt.Errorf("key %s is not found on any redis with sentinel architecture", key)
	}

	return v, nil
}

func (i *RedisConnMaker) GetCluster(key string) (*redis.ClusterClient, error) {
	key = strings.TrimSpace(strings.ToLower(key))
	v, ok := i.redisCluster[key]
	if !ok {
		return nil, fmt.Errorf("key %s is not found on any redis with cluster architecture", key)
	}

	return v, nil
}

func (i *RedisConnMaker) Get(key string) (v redis.UniversalClient, err error) {
	v, err = i.GetSingle(key)
	if err == nil {
		return v, nil
	}

	v, err = i.GetSentinel(key)
	if err == nil {
		return v, nil
	}

	v, err = i.GetCluster(key)
	if err == nil {
		return v, nil
	}

	return nil, fmt.Errorf("key %s is not found in any redis topology", key)
}

func (i *RedisConnMaker) CloseAll() error {
	ctx := i.ctx

	logger.Debug(ctx, "redis: trying to close")

	var err error
	for _, closer := range i.closer {
		if closer == nil {
			continue
		}

		if e := closer.Close(); e != nil {
			err = fmt.Errorf("%v: %w", err, e)
		} else {
			logger.Debug(ctx, fmt.Sprintf("redis: %s success to close", closer.Name()))
		}
	}

	if err != nil {
		logger.Error(ctx, "redis: some error occurred when closing dep", logger.KV("error", err))
	}

	return err
}
