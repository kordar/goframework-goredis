package goframework_goredis

import (
	"context"
	"crypto/tls"
	"github.com/kordar/godb"
	logger "github.com/kordar/gologger"
	"github.com/redis/go-redis/v9"
	"net"
)

var (
	redispool = godb.NewDbPool()
)

func GetRedisClient(db string) redis.UniversalClient {
	return redispool.Handle(db).(redis.UniversalClient)
}

// AddRedisInstancesArgs 批量添加redis句柄
func AddRedisInstancesArgs(dbs map[string]map[string]string, tls *tls.Config, dialer func(ctx context.Context, network, addr string) (net.Conn, error), onConnect func(ctx context.Context, cn *redis.Conn) error) {
	for db, cfg := range dbs {
		ins := NewRedisConnIns(db, cfg, tls, dialer, onConnect)
		if ins == nil {
			continue
		}
		err := redispool.Add(ins)
		if err != nil {
			logger.Warnf("[godb-redis] 初始化Redis异常，err=%v", err)
		}
	}
}

// AddRedisInstances 批量添加redis句柄
func AddRedisInstances(dbs map[string]map[string]string) {
	AddRedisInstancesArgs(dbs, nil, nil, nil)
}

// AddRedisInstanceArgs 添加redis句柄
func AddRedisInstanceArgs(db string, cfg map[string]string, tls *tls.Config, dialer func(ctx context.Context, network, addr string) (net.Conn, error), onConnect func(ctx context.Context, cn *redis.Conn) error) error {
	ins := NewRedisConnIns(db, cfg, tls, dialer, onConnect)
	return redispool.Add(ins)
}

// AddRedisInstance 添加redis句柄
func AddRedisInstance(db string, cfg map[string]string) error {
	return AddRedisInstanceArgs(db, cfg, nil, nil, nil)
}

func AddRedisInstanceWithRedisOptions(db string, option redis.UniversalOptions) error {
	ins := NewRedisConnInsWithRedisOption(db, option)
	return redispool.Add(ins)
}

// RemoveRedisInstance 移除redis句柄
func RemoveRedisInstance(db string) {
	redispool.Remove(db)
}

// HasRedisInstance redis句柄是否存在
func HasRedisInstance(db string) bool {
	return redispool != nil && redispool.Has(db)
}
