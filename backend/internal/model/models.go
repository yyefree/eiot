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
	DeviceSecret string  `gorm:"size:128" json:"-"`
	ProductSecret string `gorm:"size:128" json:"-"`
	ProductID  uint      `gorm:"index" json:"product_id"`
	ProductKey string    `gorm:"size:64" json:"product_key"`
	OwnerID    uint      `gorm:"index" json:"owner_id"`
	Status     int       `gorm:"default:1" json:"status"`
	BindMode   string    `gorm:"size:16;default:device_secret" json:"bind_mode"`
	FirmwareVer string   `gorm:"size:32" json:"firmware_ver"`
	IPAddress  string    `gorm:"size:45" json:"ip_address"`
	LastOnline *time.Time `json:"last_online"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// DeviceShare 设备共享
type DeviceShare struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DeviceID   uint      `gorm:"index" json:"device_id"`
	ShareUserID uint    `gorm:"index" json:"share_user_id"`
	Permission string    `gorm:"size:32;default:read" json:"permission"`
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
	UserID    uint      `gorm:"index" json:"user_id"`
	Action    string    `gorm:"size:255" json:"action"`
	Target    string    `gorm:"size:255" json:"target"`
	Detail    string    `gorm:"size:1024" json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

// Dashboard 用户看板配置
type Dashboard struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	OwnerID    uint      `gorm:"index" json:"owner_id"`
	Type       string    `gorm:"size:32" json:"type"`
	Name       string    `gorm:"size:128" json:"name"`
	LayoutJSON string    `gorm:"type:text" json:"layout_json"`
	IsDefault  int       `gorm:"default:0" json:"is_default"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ========== 家庭/房间/场景/消息/OTA（云智能App核心模型）==========

// Home 家庭
type Home struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:64" json:"name"`
	Address   string    `gorm:"size:255" json:"address"`
	Icon      string    `gorm:"size:255" json:"icon"`
	OwnerID   uint      `gorm:"index" json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// HomeMember 家庭成员
type HomeMember struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	HomeID   uint   `gorm:"index" json:"home_id"`
	UserID   uint   `gorm:"index" json:"user_id"`
	Role     string `gorm:"size:16;default:member" json:"role"` // owner / admin / member
	Nickname string `gorm:"size:64" json:"nickname"`
}

// Room 房间
type Room struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	HomeID    uint      `gorm:"index" json:"home_id"`
	Name      string    `gorm:"size:64" json:"name"`
	Icon      string    `gorm:"size:255" json:"icon"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RoomDevice 房间-设备关联
type RoomDevice struct {
	RoomID   uint `gorm:"primaryKey" json:"room_id"`
	DeviceID uint `gorm:"primaryKey" json:"device_id"`
}

// Scene 场景
type Scene struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	HomeID    uint      `gorm:"index" json:"home_id"`
	Name      string    `gorm:"size:64" json:"name"`
	Icon      string    `gorm:"size:255" json:"icon"`
	Type      string    `gorm:"size:16;default:manual" json:"type"` // manual / auto
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SceneCondition 场景触发条件
type SceneCondition struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	SceneID    uint   `gorm:"index" json:"scene_id"`
	Type       string `gorm:"size:32" json:"type"` // time / device / location
	ConfigJSON string `gorm:"type:text" json:"config_json"`
}

// SceneAction 场景执行动作
type SceneAction struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	SceneID    uint   `gorm:"index" json:"scene_id"`
	DeviceID   uint   `json:"device_id"`
	ActionJSON string `gorm:"type:text" json:"action_json"`
	SortOrder  int    `gorm:"default:0" json:"sort_order"`
}

// Message 消息通知
type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Type      string    `gorm:"size:32" json:"type"` // system / device / scene
	Title     string    `gorm:"size:128" json:"title"`
	Content   string    `gorm:"type:text" json:"content"`
	Read      bool      `gorm:"default:false" json:"read"`
	ExtraJSON string    `gorm:"type:text" json:"extra_json"`
	CreatedAt time.Time `json:"created_at"`
}

// OTAFirmware OTA固件包
type OTAFirmware struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ProductID  uint      `gorm:"index" json:"product_id"`
	Version    string    `gorm:"size:32" json:"version"`
	Changelog  string    `gorm:"type:text" json:"changelog"`
	FileURL    string    `gorm:"size:512" json:"file_url"`
	Size       int64     `json:"size"`
	Status     string    `gorm:"size:16;default:pending" json:"status"` // pending / pushing / done
	CreatedBy  uint      `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// DeviceDataHistory 设备属性历史数据
type DeviceDataHistory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DeviceSN  string    `gorm:"size:64;index:idx_sn_ts" json:"device_sn"`
	Property  string    `gorm:"size:64;index:idx_sn_prop_ts" json:"property"`
	Value     string    `gorm:"size:255" json:"value"`
	CreatedAt time.Time `gorm:"index:idx_sn_ts" json:"created_at"`
}

// AlertRule 告警规则
type AlertRule struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ProductID   uint      `gorm:"index" json:"product_id"`
	DeviceSN    string    `gorm:"size:64;index" json:"device_sn"` // 为空表示产品级规则
	Property    string    `gorm:"size:64" json:"property"`
	Operator    string    `gorm:"size:8" json:"operator"` // >, <, ==, >=, <=, !=
	Threshold   string    `gorm:"size:64" json:"threshold"`
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	NotifyType  string    `gorm:"size:16;default:message" json:"notify_type"` // message / webhook
	NotifyURL   string    `gorm:"size:512" json:"notify_url"`
	CreatedBy   uint      `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
