package bucketlimiter

import (
	"fmt"
	"testing"
	"time"

	"gopkg.in/redis.v4"
)

func Test01(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
	bucket := New(client, "test01")
	for i := 0; i <= 12; i++ {
		time.Sleep(100 * time.Millisecond)
		res, err := bucket.Consume(2)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(res)
	}
	time.Sleep(5 * time.Second)
	res, err := bucket.Consume(2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)

}
