package pubsub

import (
	"fmt"
	"log"

	"github.com/go-redis/redis/v7"
)

type RedisClusterClientOpt struct {
	Mode       string   `validate:"required,oneof=single sentinel cluster"`
	Address    []string `validate:"required,unique,dive,required"`
	Username   string   `validate:"-"`
	Password   string   `validate:"-"`
	DB         int      `validate:"-"`
	MasterName string   `validate:"required_if=Mode sentinel"`
}

func (conf RedisClusterClientOpt) MakeRedisClient() interface{} {
	// Define each redis client based on "mode", this way we can explicitly states what infrastructure we use.
	// redis.NewUniversalClient is good, but lacking defining mode explicitly.
	var redisClient redis.UniversalClient
	switch conf.Mode {
	case "single":
		redisClient = redis.NewClient(&redis.Options{
			Addr:     conf.Address[0],
			Username: conf.Username,
			Password: conf.Password,
			DB:       conf.DB,
		})
	case "sentinel":
		redisClient = redis.NewFailoverClient(&redis.FailoverOptions{
			SentinelAddrs: conf.Address,
			Username:      conf.Username,
			Password:      conf.Password,
			DB:            conf.DB,
			MasterName:    conf.MasterName,
		})
	case "cluster":
		// cluster mode is not support DB selection
		redisClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    conf.Address,
			Username: conf.Username,
			Password: conf.Password,
		})
	default:
		err := fmt.Errorf("unknown redis mode: %s", conf.Mode)
		log.Println(err)
		return nil
	}

	err := redisClient.Ping().Err()
	if err != nil {
		panic(err)
	}

	return redisClient
}
