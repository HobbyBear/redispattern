package bucketlimiter

import (
	"testing"
	"time"

	"gopkg.in/redis.v4"
)

func Test01(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
	_, err := client.Eval(script, []string{"testset"}, time.Now().Second(), time.Millisecond.Seconds()).Result()
	if err != nil {
		t.Fatal(err)
	}
}
