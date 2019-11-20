package concurrentlimiter

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"time"

	"gopkg.in/redis.v4"
)

const (
	enterScript = `
	local key = KEYS[1]
	local limit = tonumber(ARGV[1])
	local ttl = tonumber(ARGV[2])
	local now = tonumber(ARGV[3])
	local random = ARGV[4]
		redis.call('zremrangebyscore',key,'-inf', now - ttl )
	local count = redis.call('zcard',key)
	if count < limit then
		redis.call('zadd',key,now,random)
		return 1	
	end
	return 0
`

	leaveScript = `
		local key = KEYS[1]
		local random = ARGV[1]
		local ret = redis.call("zrem", key, random)
		return ret
`
)

type Limiter struct {
	redis  *redis.Client
	key    string
	option Option
}

type Option struct {
	ttl   int64 // time to live
	limit int64 // maximum running limit
}

type Setter func(*Option)

func WithTTL(d time.Duration) Setter {
	return func(option *Option) {
		option.ttl = int64(d / time.Millisecond)
	}
}

func WithLimit(limit int64) Setter {
	return func(option *Option) {
		option.limit = limit
	}
}

func New(redis *redis.Client, key string, setters ...Setter) *Limiter {
	op := Option{
		ttl:   int64(3 * time.Second / time.Millisecond),
		limit: 10,
	}
	for _, setter := range setters {
		setter(&op)
	}
	l1 := Limiter{
		redis:  redis,
		key:    key,
		option: op,
	}
	return &l1
}

func (l1 *Limiter) Enter(random string) (bool, error) {
	digest := scriptDigest(enterScript)
	exists, err := l1.redis.ScriptExists(enterScript).Result()
	if err != nil {
		return false, err
	}
	if !exists[0] {
		_, err = l1.redis.ScriptLoad(enterScript).Result()
		if err != nil {
			return false, err
		}
	}
	ret, err := l1.redis.EvalSha(digest,
		[]string{l1.key},
		l1.option.limit,
		l1.option.ttl,
		time.Now().UnixNano()/int64(time.Millisecond),
		random,
	).Result()
	if err != nil {
		return false, err
	}
	r, ok := ret.(int64)
	if !ok {
		return false, errors.New(" unknown error")
	}
	return r == 1, nil

}

func scriptDigest(script string) string {
	sha1 := sha1.New()
	sha1.Write([]byte(enterScript))
	return hex.EncodeToString(sha1.Sum(nil))
}

func (l1 *Limiter) Leave(random string) error {
	d := scriptDigest(leaveScript)
	exist, err := l1.redis.ScriptExists(d).Result()
	if err != nil {
		return err
	}
	if !exist[0] {
		_, err := l1.redis.ScriptLoad(leaveScript).Result()
		if err != nil {
			return err
		}
	}
	_, err = l1.redis.EvalSha(d, []string{l1.key}, random).Result()
	if err != nil {
		return err
	}
	return nil
}
