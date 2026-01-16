package servicecontext

import (
	"context"
	"fmt"
	"oktalk/internal/pkg/config"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func InitRedis(conf *config.Config) *redis.Client {

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", conf.Redis.Host, conf.Redis.Port), // 例如 "127.0.0.1:6379"
		Password: conf.Redis.Password,                                    // 密码
		DB:       conf.Redis.DB,                                          // 默认数据库
		PoolSize: 10,                                                     // 连接池大小
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		logrus.Fatalf("❌ Redis 连接失败: %v", err)
	}

	logrus.Info("✅ Redis 初始化成功")
	return rdb
}
