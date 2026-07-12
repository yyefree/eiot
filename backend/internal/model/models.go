package model

import (
	"time"
)

// Project 项目（飞燕标准：项目→产品→设备）
type Project struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:128;uniqueIndex" json:"name"`
	Description string    `gorm:"size:512" json:"description"`
	Industry    string    `gorm:"size:64" json:"industry"`    // 行业分类：智能家居/工业/农业等
	Type        string    `gorm:"size:32;default:consumer" json:"type"` // consumer(消费级)/commercial(商用)
	OwnerID     uint      `json:"owner_id"`                  // 所属管理员
	CreatedBy   uint      `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// User 用户表
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"size:64" json:"username"`
	Password  string    `gorm:"size:255" json:"-"`
	Phone     string    `gorm:"size:32;uniqueIndex" json:"phone"`
	Email     string    `gorm:"size:128" json:"email"`
	Nickname  string    `gorm:"size:64" json:"nickname"`
	Role      string    `gorm:"size:16;default:user" json:"role"` // admin / user
	Avatar    string    `gorm:"size:255" json:"avatar"`
	Status    int       `gorm:"default:1" json:"status"` // 1 正常 0 禁用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Product 产品（物模型二合一（产品内置物模型
type Product struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name           string    `gorm:"size:128;uniqueIndex" json:"name"`
	ProjectID     uint      `gorm:"default:0;index" json:"project_id"` // 所属项目，0表示未分类
	ProductKey     string    `gorm:"size:64;uniqueIndex" json:"product_key"`
	Description    string    `gorm:"size:512" json:"description"`
	Icon           string    `gorm:"size:512" json:"icon"`

	// 物模型（阿里云飞燕兼容 JSON 存储（
	// properties: 属性列表（可读可写
	PropertiesJSON string    `gorm:"type:text" json:"properties_json"` // 属性
	EventsJSON     string    `gorm:"type:text" json:"events_json"`   // 事件列表
	ServicesJSON   string    `gorm:"type:text" json:"services_json"` // 功能（服务）列表

	// 产品级默认移动端 UI 模板（管理员配置
	MobileUIJSON string    `gorm:"type:text" json:"mobile_ui_json"`

	// 联网配置（飞燕标准
	NetworkType string `gorm:"size:32;default:wifi" json:"network_type"` // wifi / ble / cellular
	Category    string `gorm:"size:64" json:"category"`                  // 品类：灯具/开关/传感器等

	CreatedBy     uint      `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Device 设备
type Device struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DeviceName string    `gorm:"size:128" json:"device_name"`
	DeviceSN   string    `gorm:"size:64;uniqueIndex" json:"device_sn"`
	DeviceSecret string  `gorm:"size:128" json:"-"`                  // 一机一密密钥
	ProductSecret string `gorm:"size:128" json:"-"`                  // 一型一密密钥（同产品共用）
	ProductID  uint      `json:"product_id"`
	ProductKey string    `gorm:"size:64" json:"product_key"`
	OwnerID    uint      `json:"owner_id"`
	Status     int       `gorm:"default:1" json:"status"`           // 1 正常 0 禁用
	BindMode   string    `gorm:"size:16;default:device_secret" json:"bind_mode"` // device_secret(一机一密) / product_secret(一型一密)
	FirmwareVer string   `gorm:"size:32" json:"firmware_ver"`        // 固件版本
	IPAddress  string    `gorm:"size:45" json:"ip_address"`          // 最近一次上报IP
	LastOnline *time.Time `json:"last_online"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// DeviceShare 设备共享
type DeviceShare struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DeviceID   uint      `json:"device_id"`
	ShareUserID uint    `json:"share_user_id"`
	Permission string    `gorm:"size:32;default:read" json:"permission"` // read / control
	ExpireAt  *time.Time `json:"expire_at"`
	CreatedBy uint      `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// UserDeviceUI 用户个人设备 UI 配置（用户可基于产品默认模板自定义个人 UI
type UserDeviceUI struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index:idx_user_device" json:"user_id"`
	DeviceID  uint      `gorm:"index:idx_user_device" json:"device_id"`
	LayoutJSON string    `gorm:"type:text" json:"layout_json"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OperationLog 操作日志
type OperationLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"user_id"`
	Action    string    `gorm:"size:255" json:"action"`
	Target    string    `gorm:"size:255" json:"target"`
	Detail    string    `gorm:"size:1024" json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

// Dashboard 用户看板配置（Web 拖拽看板（支持用户个人看板
type Dashboard struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	OwnerID    uint      `json:"owner_id"`
	Type       string    `gorm:"size:32" json:"type"` // web
	Name       string    `gorm:"size:128" json:"name"`
	LayoutJSON string    `gorm:"type:text" json:"layout_json"`
	IsDefault  int       `gorm:"default:0" json:"is_default"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
