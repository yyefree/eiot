package cache

import (
	"context"
	"eiot/pkg/config"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis(cfg config.RedisConf) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := RDB.Ping(ctx).Result()
	return err
}

// SetOnline 设置设备在线状态
func SetOnline(deviceSN string, online bool, expireSec int) {
	if RDB == nil || deviceSN == "" {
		return
	}
	status := "offline"
	if online {
		status = "online"
	}
	RDB.Set(context.Background(), "device:status:"+deviceSN, status, time.Duration(expireSec)*time.Second)
}

// GetOnline 获取设备在线状态
func GetOnline(deviceSN string) string {
	if RDB == nil {
		return "unknown"
	}
	v, _ := RDB.Get(context.Background(), "device:status:"+deviceSN).Result()
	if v == "" {
		return "offline"
	}
	return v
}

// SetLatest 设置设备最新值
func SetLatest(deviceSN, key string, value interface{}) {
	if RDB == nil || deviceSN == "" {
		return
	}
	RDB.HSet(context.Background(), "device:latest:"+deviceSN, key, value)
}

// GetLatest 获取设备最新值
func GetLatest(deviceSN string) map[string]string {
	if RDB == nil {
		return map[string]string{}
	}
	m, err := RDB.HGetAll(context.Background(), "device:latest:"+deviceSN).Result()
	if err != nil {
		return map[string]string{}
	}
	return m
}
