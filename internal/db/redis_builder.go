package db

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"messaging-server/internal/logger"
	"messaging-server/internal/model"
	"os"
	"strconv"
	"time"
)

type redisTemplate struct {
	client *redis.Client
	ctx    context.Context
	ttl    time.Duration
}

var RedisConnection *redisTemplate

// InitRedis sets up Redis struct and opens the initial connection.
func InitRedis() error {
	RedisConnection = &redisTemplate{}
	return RedisConnection.connect()
}

// connect (re)connect Redis in initial and when connection lost
func (r *redisTemplate) connect() error {
	ctx := context.Background()
	// control if redis conn already exist and alive
	if r.client != nil {
		_, err := r.client.Ping(ctx).Result()
		if err == nil {
			return nil
		} else {
			_ = r.client.Close()
		}
	}

	// create ttl for redis keys
	var redisTtl time.Duration

	if sec, err := strconv.Atoi(os.Getenv("REDIS_TTL")); err == nil {
		redisTtl = time.Duration(sec) * time.Second
	} else {
		// default 1 hour
		redisTtl = time.Duration(3600) * time.Second
	}

	// create connection if connection not exist or lost
	client := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST"),
	})

	RedisConnection = &redisTemplate{client: client, ctx: ctx, ttl: redisTtl}

	logger.Sugar.Info("redis connection established successfully")
	return nil
}

// InsertMessage set new key for given message
func (r *redisTemplate) InsertMessage(msg model.RedisMessage) error {
	m, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// create key with id
	key := "messages:" + msg.MessageID

	// set data as value
	if err = r.client.Set(r.ctx, key, m, r.ttl).Err(); err != nil {
		return err
	}

	logger.Sugar.Infof("message %s inserted to redis", msg.MessageID)
	return nil
}
