package goframework_goredis

import (
	"context"
	"crypto/tls"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
	"net"
	"runtime"
	"strings"
	"time"
)

type RedisConnIns struct {
	name string
	ins  redis.UniversalClient
}

func intValue(cfg map[string]string, key string, defaultValue int) int {
	if v, ok := cfg[key]; ok {
		return cast.ToInt(v)
	} else {
		return defaultValue
	}
}

func durationValue(cfg map[string]string, key string, defaultValue time.Duration) time.Duration {
	if v, ok := cfg[key]; ok {
		return cast.ToDuration(v)
	} else {
		return defaultValue
	}
}

func NewRedisConnIns(name string,
	cfg map[string]string,
	tLSConfig *tls.Config,
	dialer func(ctx context.Context, network, addr string) (net.Conn, error),
	onConnect func(ctx context.Context, cn *redis.Conn) error,
) *RedisConnIns {

	options := redis.UniversalOptions{
		// 单个主机或集群配置
		Addrs: strings.Split(cfg["addr"], ","),
		// ClientName 和 `Options` 相同，会对每个Node节点的每个网络连接配置
		ClientName: cast.ToString(cfg["clientName"]),
		// 设置 DB, 只针对 `Redis Client` 和 `Failover Client`
		DB: cast.ToInt(cfg["db"]),

		Username: cast.ToString(cfg["username"]),
		Password: cast.ToString(cfg["password"]),

		// 命令最大重试次数， 默认为3
		MaxRetries: intValue(cfg, "maxRetries", 3),

		// 每次重试最小间隔时间
		// 默认 8 * time.Millisecond (8毫秒) ，设置-1为禁用
		MinRetryBackoff: durationValue(cfg, "minRetryBackoff", 8*time.Millisecond),

		// 每次重试最大间隔时间
		// 默认 512 * time.Millisecond (512毫秒) ，设置-1为禁用
		MaxRetryBackoff: durationValue(cfg, "maxRetryBackoff", 512*time.Millisecond),

		//// 下面的配置项，和 `Options`、`Sentinel` 基本一致，请参照 `Options` 的说明
		//
		//Dialer    func(ctx context.Context, network, addr string) (net.Conn, error)
		//OnConnect func(ctx context.Context, cn *Conn) error
		//

		// 建立新网络连接时的超时时间
		// 默认5秒
		DialTimeout: durationValue(cfg, "dialTimeout", 5*time.Second),

		// 从网络连接中读取数据超时时间，可能的值：
		//  0 - 默认值，3秒
		// -1 - 无超时，无限期的阻塞
		// -2 - 不进行超时设置，不调用 SetReadDeadline 方法
		ReadTimeout: durationValue(cfg, "readTimeout", 3*time.Second),

		// 把数据写入网络连接的超时时间，可能的值：
		//  0 - 默认值，3秒
		// -1 - 无超时，无限期的阻塞
		// -2 - 不进行超时设置，不调用 SetWriteDeadline 方法
		WriteTimeout: durationValue(cfg, "writeTimeout", 3*time.Second),

		// 是否使用context.Context的上下文截止时间，
		// 有些情况下，context.Context的超时可能带来问题。
		// 默认不使用
		ContextTimeoutEnabled: cast.ToBool(cfg["contextTimeoutEnabled"]),

		// 连接池的类型，有 LIFO 和 FIFO 两种模式，
		// PoolFIFO 为 false 时使用 LIFO 模式，为 true 使用 FIFO 模式。
		// 当一个连接使用完毕时会把连接归还给连接池，连接池会把连接放入队尾，
		// LIFO 模式时，每次取空闲连接会从"队尾"取，就是刚放入队尾的空闲连接，
		// 也就是说 LIFO 每次使用的都是热连接，连接池有机会关闭"队头"的长期空闲连接，
		// 并且从概率上，刚放入的热连接健康状态会更好；
		// 而 FIFO 模式则相反，每次取空闲连接会从"队头"取，相比较于 LIFO 模式，
		// 会使整个连接池的连接使用更加平均，有点类似于负载均衡寻轮模式，会循环的使用
		// 连接池的所有连接，如果你使用 go-redis 当做代理让后端 redis 节点负载更平均的话，
		// FIFO 模式对你很有用。
		// 如果你不确定使用什么模式，请保持默认 PoolFIFO = false
		PoolFIFO: cast.ToBool(cfg["poolFIFO"]),
	}

	// 连接池最大连接数量，注意：这里不包括 pub/sub，pub/sub 将使用独立的网络连接
	// 默认为 10 * runtime.GOMAXPROCS
	if cfg["poolSize"] != "" {
		poolSize := cast.ToInt(cfg["poolSize"])
		options.PoolSize = poolSize * runtime.GOMAXPROCS(-1)
	}

	// PoolTimeout 代表如果连接池所有连接都在使用中，等待获取连接时间，超时将返回错误
	// 默认是 1秒+ReadTimeout
	if cfg["poolTimeout"] != "" {
		options.PoolTimeout = cast.ToDuration(cfg["poolTimeout"])
	}

	// 连接池保持的最小空闲连接数，它受到PoolSize的限制
	// 默认为0，不保持
	if cfg["minIdleConns"] != "" {
		options.MinIdleConns = cast.ToInt(cfg["minIdleConns"])
	}

	// 连接池保持的最大空闲连接数，多余的空闲连接将被关闭
	// 默认为0，不限制
	if cfg["maxIdleConns"] != "" {
		options.MaxIdleConns = cast.ToInt(cfg["maxIdleConns"])
	}

	// ConnMaxIdleTime 是最大空闲时间，超过这个时间将被关闭。
	// 如果 ConnMaxIdleTime <= 0，则连接不会因为空闲而被关闭。
	// 默认值是30分钟，-1禁用
	if cfg["connMaxIdleTime"] != "" {
		options.ConnMaxIdleTime = cast.ToDuration(cfg["connMaxIdleTime"])
	}

	// ConnMaxLifetime 是一个连接的生存时间，
	// 和 ConnMaxIdleTime 不同，ConnMaxLifetime 表示连接最大的存活时间
	// 如果 ConnMaxLifetime <= 0，则连接不会有使用时间限制
	// 默认值为0，代表连接没有时间限制
	if cfg["connMaxLifetime"] != "" {
		options.ConnMaxLifetime = cast.ToDuration(cfg["connMaxLifetime"])
	}

	//
	//TLSConfig *tls.Config
	if tLSConfig != nil {
		options.TLSConfig = tLSConfig
	}

	// 集群配置项
	if cfg["maxRedirects"] != "" {
		options.MaxRedirects = cast.ToInt(cfg["maxRedirects"])
	}

	if cfg["readOnly"] != "" {
		options.ReadOnly = cast.ToBool(cfg["readOnly"])
	}
	if cfg["routeByLatency"] != "" {
		options.RouteByLatency = cast.ToBool(cfg["routeByLatency"])
	}
	if cfg["routeRandomly"] != "" {
		options.RouteRandomly = cast.ToBool(cfg["routeRandomly"])
	}

	// 哨兵 Master Name，仅适用于 `Failover Client`
	if cfg["masterName"] != "" {
		options.MasterName = cast.ToString(cfg["masterName"])
	}
	if cfg["protocol"] != "" {
		options.Protocol = cast.ToInt(cfg["protocol"])
	}
	if cfg["sentinelUsername"] != "" {
		options.SentinelUsername = cast.ToString(cfg["sentinelUsername"])
	}
	if cfg["sentinelPassword"] != "" {
		options.SentinelPassword = cast.ToString(cfg["sentinelPassword"])
	}

	if dialer != nil {
		options.Dialer = dialer
	}

	if onConnect != nil {
		options.OnConnect = onConnect
	}

	return NewRedisConnInsWithRedisOption(name, options)
}

func NewRedisConnInsWithRedisOption(name string, option redis.UniversalOptions) *RedisConnIns {
	client := redis.NewUniversalClient(&option)
	return &RedisConnIns{name: name, ins: client}
}

func (c RedisConnIns) GetName() string {
	return c.name
}

func (c RedisConnIns) GetInstance() interface{} {
	return c.ins
}

func (c RedisConnIns) Close() error {
	return c.ins.Close()
}
