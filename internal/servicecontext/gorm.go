package servicecontext

import (
	"context"
	"fmt"
	"time"

	"oktalk/internal/pkg/config"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormLogger 自定义 GORM 日志实现，对接我们的 logrus
type GormLogger struct{}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface { return l }
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	logrus.WithContext(ctx).Infof(msg, data...)
}
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	logrus.WithContext(ctx).Warnf(msg, data...)
}
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	logrus.WithContext(ctx).Errorf(msg, data...)
}
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	fields := logrus.Fields{
		"elapsed": elapsed,
		"rows":    rows,
	}
	if err != nil {
		fields["error"] = err
		logrus.WithContext(ctx).WithFields(fields).Errorf("SQL ERROR: %s", sql)
	} else {
		logrus.WithContext(ctx).WithFields(fields).Infof("SQL: %s", sql)
	}
}

// InitGORM 初始化数据库连接
func InitGORM(conf *config.Config) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.Database.User,
		conf.Database.Password,
		conf.Database.Host,
		conf.Database.Port,
		conf.Database.Database)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		// 使用我们自定义的 Logger，这样 SQL 就会带上 TraceID
		Logger: &GormLogger{},
	})

	if err != nil {
		logrus.Fatalf("❌ 数据库连接失败: %v", err)
	}

	// 配置连接池
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(conf.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(conf.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(conf.Database.ConnMaxLifetime) * time.Second)

	logrus.Info("✅ 数据库(GORM)初始化成功")
	return db
}
