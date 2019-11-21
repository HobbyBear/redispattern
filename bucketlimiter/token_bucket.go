package bucketlimiter

import (
	"time"

	"gopkg.in/redis.v4"
)

const script = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local expiration = tonumber(ARGV[2])
local rate_limit_info = redis.call("HMGET",key,'last_time','exit_token_num')
local last_time = tonumber(rate_limit_info[1])
local exit_token_num = tonumber(rate_limit_info[2])

redis.call('set','test',last_time)

return 1


`

type TokenBucket struct {
	redis      *redis.Client
	TokenNum   int64 // total token num (bucket size)
	Key        string
	Expiration time.Duration
	Rate       time.Duration // put one token every rate
}

type setter func(*TokenBucket)

func WithTokenNum(num int64) setter {
	return func(bucket *TokenBucket) {
		bucket.TokenNum = num
	}
}

func WithExpiration(expiration time.Duration) setter {
	return func(bucket *TokenBucket) {
		bucket.Expiration = expiration
	}
}

func WithRate(rate time.Duration) setter {
	return func(bucket *TokenBucket) {
		bucket.Rate = rate
	}
}

func New(redis *redis.Client, key string, setters ...setter) *TokenBucket {
	tokenBucket := TokenBucket{
		redis:      redis,
		TokenNum:   20,
		Key:        key,
		Expiration: 3 * time.Second,
		Rate:       time.Second,
	}

	for _, setter := range setters {
		setter(&tokenBucket)
	}

	return &tokenBucket
}

func Consume(num int64) bool {
	return false
}
