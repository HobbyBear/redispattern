package concurrentlimiter

import (
	"fmt"
	"testing"
	"time"
)

func Test01(t *testing.T) {
	fmt.Println(int64(3 * time.Second / time.Millisecond))
	fmt.Println(time.Now().UnixNano() / int64(time.Millisecond))
}
