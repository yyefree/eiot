package dao

import (
	"eiot/internal/model"
	"eiot/pkg/config"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitMySQL 初始化 MySQL 连接
func InitMySQL(cfg config.MySQLConf) error {
	var err error
	// 重试 3 次，等待数据库服务完全就绪
	for i := 0; i < 5; i++ {
		DB, err = gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("[MySQL] connect attempt %d failed: %v, retrying...", i+1, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		return err
	}
	sqlDB, err := DB.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}
	// 自动迁移（飞燕兼容模型
	err = DB.AutoMigrate(
    &model.Project{},
    &model.User{},
    &model.Product{},
    &model.Device{},
    &model.DeviceShare{},
    &model.UserDeviceUI{},
    &model.OperationLog{},
    &model.Dashboard{},
  )
	if err != nil {
		log.Printf("[MySQL] migrate warning: %v", err)
	}
	log.Printf("[MySQL] initialized OK")
	return nil
}
