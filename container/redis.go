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
	ctx          context.Context
	conf         config.Redis
	redisConnMap map[string]redis.UniversalClient
	closer       []Closer
}

func NewRedisConnMaker(ctx context.Context, conf config.Redis) (*RedisConnMaker, error) {
	err := validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	instance := &RedisConnMaker{
		ctx:          ctx,
		redisConnMap: make(map[string]redis.UniversalClient),
		closer:       make([]Closer, 0),
	}

	err = instance.connect()
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
		if err := validator.New().Var(key, "required,alphanumeric"); err != nil {
			err = fmt.Errorf("error connecting to database key '%s': %w", key, err)
			return err
		}

		var redisClient redis.UniversalClient
		switch connInfo.Mode {
		case "single":
			redisClient = redis.NewClient(&redis.Options{
				Addr:     connInfo.Address[0],
				Username: connInfo.Username,
				Password: connInfo.Password,
				DB:       connInfo.DB,
			})
		case "sentinel":
			redisClient = redis.NewFailoverClient(&redis.FailoverOptions{
				SentinelAddrs: connInfo.Address,
				Username:      connInfo.Username,
				Password:      connInfo.Password,
				DB:            connInfo.DB,
				MasterName:    connInfo.MasterName,
			})
		case "cluster":
			// cluster mode is not support DB selection
			redisClient = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:    connInfo.Address,
				Username: connInfo.Username,
				Password: connInfo.Password,
			})
		default:
			err := fmt.Errorf("unknown redis mode: %s", connInfo.Mode)
			return err
		}

		err := redisClient.Ping(ctx).Err()
		if err != nil {
			return err
		}

		i.redisConnMap[key] = redisClient
		i.closer = append(i.closer, NewNamedCloser(key, redisClient))
	}

	return nil
}

func (i *RedisConnMaker) Get(key string) (redis.UniversalClient, error) {
	v, ok := i.redisConnMap[key]
	if !ok {
		return nil, fmt.Errorf("key %s is not exist on redis list", key)
	}

	return v, nil
}

func (i *RedisConnMaker) CloseAll() error {
	ctx := i.ctx

	logger.Debug(ctx, "db sql: trying to close")

	var err error
	for _, closer := range i.closer {
		if closer == nil {
			continue
		}

		if e := closer.Close(); e != nil {
			err = fmt.Errorf("%v: %w", err, e)
		} else {
			logger.Debug(ctx, fmt.Sprintf("db sql: %s success to close", closer.Name()))
		}
	}

	if err != nil {
		logger.Error(ctx, "db sql: some error occurred when closing dep", logger.KV("error", err))
	}

	return err
}
