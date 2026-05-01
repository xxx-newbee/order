package svc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xxx-newbee/order/internal/config"
	"github.com/xxx-newbee/order/internal/model"
	"github.com/xxx-newbee/storage"
	"github.com/xxx-newbee/storage/cache"
	"github.com/xxx-newbee/storage/locker"
	"github.com/xxx-newbee/storage/queue"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config            config.Config
	Cache             storage.AdapterCache
	Locker            storage.AdapterLocker
	MemoryQueue       storage.AdapterQueue
	RedisQueue        storage.AdapterQueue
	OrderMainModel    model.OrderMainModel
	OrderItemModel    model.OrderItemModel
	SeckillStockModel model.SeckillStockModel
	SeckillActivity   model.SeckillActivityModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := InitDB(c)
	redis := InitRedis(c)
	redisQueue := InitRedisQueue(c)

	svcCtx := &ServiceContext{
		Config:            c,
		Cache:             cache.NewRedis(redis, nil),
		Locker:            locker.NewRedisLocker(redis),
		MemoryQueue:       queue.NewMemoryQueue(c.Queue.Memory.PoolSize),
		RedisQueue:        redisQueue,
		OrderMainModel:    model.NewOrderMainModel(db),
		OrderItemModel:    model.NewOrderItemModel(db),
		SeckillStockModel: model.NewSeckillStockModel(db),
		SeckillActivity:   model.NewSeckillActivityModel(db),
	}

	// 注册消息队列消费者（全局仅一次）
	redisQueue.Register(model.SeckillStock{}.TableName(), svcCtx.SeckillStockConsumer)
	redisQueue.Run()

	return svcCtx
}

func InitDB(c config.Config) *gorm.DB {
	dsn := c.Database.DataSource
	var err error
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	sqlDb, err := db.DB()
	if err != nil {
		panic("failed to get database: " + err.Error())
	}
	sqlDb.SetMaxOpenConns(c.Database.MaxOpenConns)
	sqlDb.SetMaxIdleConns(c.Database.MaxIdleConns)
	sqlDb.SetConnMaxLifetime(time.Duration(c.Database.ConnMaxLifetime) * time.Second)

	if sqlDb.PingContext(context.Background()) != nil {
		panic("failed to ping database: " + err.Error())
	}
	println("MySQL connected ✅")
	return db
}

func InitRedis(c config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     c.Cache.Redis.Addr,
		Password: c.Cache.Redis.Password,
		DB:       c.Cache.Redis.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	return client
}

func (s *ServiceContext) SeckillStockConsumer(msg storage.Messager) error {
	order := struct {
		UserId     int64  `json:"user_id"`
		OrderNo    string `json:"order_no"`
		ActivityId int64  `json:"activity_id"`
		ProductId  int64  `json:"product_id"`
	}{}
	rb, err := json.Marshal(msg.GetValues())
	if err != nil {
		return err
	}

	if err = json.Unmarshal(rb, &order); err != nil {
		return err
	}

	// 乐观锁扣减库存
	affected, err := s.SeckillStockModel.DecreaseStock(order.ActivityId)
	if err != nil || affected == 0 {
		// 扣减失败，更新订单为取消状态
		logx.Errorf("数据库库存扣减失败，error: %v", err)
		_ = s.OrderMainModel.UpdateStatus(order.OrderNo, 4)
		if err != nil {
			return err
		}
		return nil
	}
	// 扣减成功，更新订单为待付款状态
	logx.Infof("秒杀库存扣减成功，orderNo: %s", order.OrderNo)
	return s.OrderMainModel.UpdateStatus(order.OrderNo, 0)
}

func InitRedisQueue(c config.Config) storage.AdapterQueue {
	client := redis.NewClient(&redis.Options{
		Addr:     c.Queue.Redis.Addr,
		Password: c.Queue.Redis.Password,
		DB:       c.Queue.Redis.DB,
	})
	return queue.NewRedisQueue(client, c.Queue.Redis.Prefix, c.Queue.Redis.MaxRetry)
}
