package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

var C Config

type Config struct {
	zrpc.RpcServerConf
	Database struct {
		DataSource      string
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime int
	}
	Cache struct {
		Redis  RedisConf
		Memory string
	}
	Queue struct {
		Redis  RedisQueueConf
		Memory MemoryQueueConf
	}
}

type RedisConf struct {
	Addr     string
	Password string
	DB       int
}

type RedisQueueConf struct {
	Addr     string
	Password string
	DB       int
	Prefix   string
	MaxRetry int
}

type MemoryQueueConf struct {
	PoolSize int
}
