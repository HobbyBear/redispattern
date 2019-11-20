package redislimiter

import (
	"time"
	"github.com/go-redis/redis"
)

type Limiter struct {
	redis    *redis.Client
	key      string
	limit    int64
	duration time.Duration
}

type option struct {
	limit    int64
	duration time.Duration
}

type Setter func(*option)

func WithLimit(limit int64) Setter {
	return func(o *option) {
		o.limit = limit
	}
}

func WithDuration(duration time.Duration) Setter {
	return func(o *option) {
		o.duration = duration
	}
}

func New(client *redis.Client, key string, setters ...Setter) *Limiter {
	op := new(option)
	for _, setter := range setters {
		setter(op)
	}
	return &Limiter{
		redis:    client,
		key:      key,
		limit:    op.limit,
		duration: op.duration,
	}
}

func (l1 *Limiter) Limit() bool {
	current, _ := l1.redis.LLen(l1.key).Result()
	value := "0"
	if current > l1.limit {
		return true
	} else {
		exit, err := l1.redis.Exists(l1.key).Result()
		if err != nil {
			return true
		}
		if exit == 0 {
			pipe := l1.redis.TxPipeline()
			pipe.RPush(l1.key, value)
			pipe.Expire(l1.key, l1.duration)
			_, err := pipe.Exec()
			if err != nil {
				return true
			}
		} else {
			l1.redis.RPush(value)
		}
	}
	return false
}
