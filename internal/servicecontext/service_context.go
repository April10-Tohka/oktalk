package servicecontext

import (
	"oktalk/internal/model"
	"oktalk/internal/pkg/config"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config *config.Config
	DB     *gorm.DB
	Redis  *redis.Client
}

func NewServiceContext(conf *config.Config) *ServiceContext {
	// 1. 初始化 GORM
	db := InitGORM(conf)
	// 2. 执行自动迁移 (建表)
	// 每次启动都会检查表结构，如果模型有变动会自动增加字段
	err := db.AutoMigrate(
		&model.UserLearningRecord{},
		// 以后有新的 Model 往这里加即可
	)
	if err != nil {
		logrus.Errorf("❌ 数据库自动迁移失败: %v", err)
	} else {
		logrus.Info("✅ 数据库模型自动迁移成功")
	}
	// 2. 初始化 Redis
	rdb := InitRedis(conf)

	return &ServiceContext{
		Config: conf,
		DB:     db,
		Redis:  rdb,
	}
}
