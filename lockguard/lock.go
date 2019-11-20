package lockguard

import (
	"time"

	"gopkg.in/redis.v4"
)

type lock struct {
	redis      *redis.Client
	Key        string
	Value      string
	Expiration time.Duration
}



