package goframework_goredis_test

import (
	"context"
	goframeworkgoredis "github.com/kordar/goframework-goredis"
	"testing"
	"time"
)

func TestRedis(t *testing.T) {
	_ = goframeworkgoredis.AddRedisInstanceArgs("aa", map[string]string{
		"addr":     "xxxxxx",
		"db":       "2",
		"password": "xxxxx",
	}, nil, nil, nil)
	client := goframeworkgoredis.GetRedisClient("aa")
	ctx := context.Background()

	client.Set(ctx, "bbb", "ccc", 10*time.Second)
}
