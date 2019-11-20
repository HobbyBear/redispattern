package lockguard

import (
	"crypto/rand"
	"crypto/rc4"
	"time"

	"gopkg.in/redis.v4"
)

const (
	redisLockKey = "HHsYC5oVzLjFuWE4KMz923QT"
	delScript    = `
	if redis.call('get','KEYS[1]') == ARGV[1]
	then 
	return redis.call('del',KEYS[1])
	else
	return 0
end
`
)

type LockGuard struct {
	lock lock
}

type Setter func(*lock)

func WithExpiration(expiration time.Duration) Setter {
	return func(lo *lock) {
		lo.Expiration = expiration
	}
}

func New(redis *redis.Client, key string, setters ...Setter) *LockGuard {
	l := lock{
		redis:      redis,
		Key:        key,
		Value:      "",
		Expiration: 30 * time.Second,
	}
	for _, setter := range setters {
		setter(&l)
	}
	return &LockGuard{lock: l}
}

func (guard *LockGuard) Lock() bool {
	src := make([]byte, len(redisLockKey))
	_, err := rand.Read(src)
	if err != nil {
		return false
	}
	cipher, err := rc4.NewCipher([]byte(redisLockKey))
	if err != nil {
		return false
	}
	cipher.XORKeyStream(src, src)
	guard.lock.Value = string(src)
	flag, err := guard.lock.redis.SetNX(guard.lock.Key, guard.lock.Value, guard.lock.Expiration).Result()
	if err != nil {
		return false
	}
	return flag
}

func (guard *LockGuard) Unlock() {
	if len(guard.lock.Key) > 0 && len(guard.lock.Value) > 0 {
		guard.lock.redis.Eval(delScript, []string{guard.lock.Key}, guard.lock.Value)
	}
}
