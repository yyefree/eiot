package tsdb

import (
	"log"
	"time"
)

// InitTDengine 初始化 TDengine 客户端；当前版本先跳过，预留入口
// 真实环境下会建库建超表并写入时序数据
func InitTDengine(dsn string) error {
	if dsn == "" {
		return nil
	}
	log.Printf("[TDengine] placeholder init: %s (当前版本未启用，数据仅写入 Redis latest)", dsn)
	return nil
}

// InsertProperty 插入一条属性时序数据（当前实现无操作，返回 nil）
func InsertProperty(deviceSN, identifier string, value interface{}, ts time.Time) error {
	return nil
}
