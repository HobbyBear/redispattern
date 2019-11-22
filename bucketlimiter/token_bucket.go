package bucketlimiter

import (
	"errors"
	"time"

	"gopkg.in/redis.v4"
)

const script = `
local key = KEYS[1];
local now = tonumber(ARGV[1])
local expiration = tonumber(ARGV[2])
local tokenNum = tonumber(ARGV[3])
local rate = tonumber(ARGV[4])
local consumeTokenNum = tonumber(ARGV[5])
local rate_limit_info = redis.call('HMGET',key,'last_time','exit_token_num');
local last_time = now
local exit_token_num  = tokenNum
if (rate_limit_info[1]) then
	last_time = tonumber(rate_limit_info[1])
	exit_token_num = tonumber(rate_limit_info[2])
end
local expectedTokenNum = math.floor(exit_token_num + (now - last_time) / rate)
expectedTokenNum = math.min(expectedTokenNum,tokenNum)
if expectedTokenNum < consumeTokenNum then 
	return 0
end
	exit_token_num = expectedTokenNum - consumeTokenNum
	redis.call("HMSET",key,"last_time",now,"exit_token_num",exit_token_num)
	redis.call("Expire",key,expiration)

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
		Expiration: 100 * time.Second,
		Rate:       1 * time.Second,
	}

	for _, setter := range setters {
		setter(&tokenBucket)
	}

	return &tokenBucket
}

func (tokenBucket *TokenBucket) Consume(num int64) (bool, error) {
	ret, err := tokenBucket.redis.Eval(script,
		[]string{tokenBucket.Key},
		time.Now().UnixNano()/int64(time.Millisecond),
		tokenBucket.Expiration.Seconds(),
		tokenBucket.TokenNum,
		int64(tokenBucket.Rate/time.Millisecond),
		num,
	).Result()
	if err != nil {
		return false, err
	}
	res, ok := ret.(int64)
	if !ok {
		return false, errors.New("unkonwn error")
	}
	return res == 1, nil
}
