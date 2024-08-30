# goframework-goredis

包装`redis`对象，实现[`godb`](https://github.com/kordar/godb)接口。

## 安装
```go
go get github.com/kordar/goframework-goredis v0.0.1
```

## 使用

- 添加实例

```go
_ = goframeworkgoredis.AddRedisInstanceArgs("aa", map[string]string{
"addr":     "127.0.0.1:1234",
"db":       "2",
"password": "xxxxxx",
}, nil, nil, nil)
client := goframeworkgoredis.GetRedisClient("aa")
ctx := context.Background()

client.Set(ctx, "bbb", "ccc", 10*time.Second)
```





