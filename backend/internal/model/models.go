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
	ID            uint      `gorm:"primaryKey" json:"id"`
	DeviceName    string    `gorm:"size:128" json:"device_name"`
	DeviceSN      string    `gorm:"size:64;uniqueIndex" json:"device_sn"`
	DeviceSecret  string    `gorm:"size:128" json:"-"`
	ProductSecret string    `gorm:"size:128" json:"-"`
	ProductID     uint      `gorm:"index" json:"product_id"`
	ProductKey    string    `gorm:"size:64" json:"product_key"`
	OwnerID       uint      `gorm:"index" json:"owner_id"`
	Status        int       `gorm:"default:1" json:"status"`
	BindMode      string    `gorm:"size:16;default:device_secret" json:"bind_mode"`
	FirmwareVer   string    `gorm:"size:32" json:"firmware_ver"`
	IPAddress     string    `gorm:"size:45" json:"ip_address"`
	LastOnline    *time.Time `json:"last_online"`
	// 网关/子设备相关
	NodeType      string    `gorm:"size:16;default:device" json:"node_type"`  // gateway / device / sub_device
	GatewayID     uint      `gorm:"index" json:"gateway_id"`                  // 网关设备 ID
	TopoStatus    string    `gorm:"size:16" json:"topo_status"`               // online / offline / adding / removing
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
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

// DeviceEventHistory 设备事件上报历史
type DeviceEventHistory struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DeviceSN   string    `gorm:"size:64;index:idx_ev_sn_ts" json:"device_sn"`
	EventID    string    `gorm:"size:64;index:idx_ev_sn_ts" json:"event_id"`
	EventName  string    `gorm:"size:128" json:"event_name"`
	OutputJSON string    `gorm:"type:text" json:"output_json"`
	CreatedAt  time.Time `gorm:"index:idx_ev_sn_ts" json:"created_at"`
}

// DeviceServiceHistory 设备服务调用历史
type DeviceServiceHistory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	DeviceSN    string    `gorm:"size:64;index:idx_svc_sn_ts" json:"device_sn"`
	ServiceID   string    `gorm:"size:64;index:idx_svc_sn_ts" json:"service_id"`
	ServiceName string    `gorm:"size:128" json:"service_name"`
	InputJSON   string    `gorm:"type:text" json:"input_json"`
	OutputJSON  string    `gorm:"type:text" json:"output_json"`
	Status      string    `gorm:"size:16;default:success" json:"status"`
	CreatedAt   time.Time `gorm:"index:idx_svc_sn_ts" json:"created_at"`
}

// DeviceShadow 设备影子（期望值/上报值）
type DeviceShadow struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	DeviceSN      string    `gorm:"size:64;uniqueIndex" json:"device_sn"`
	DesiredJSON   string    `gorm:"type:text" json:"desired_json"`
	ReportedJSON  string    `gorm:"type:text" json:"reported_json"`
	Version       int       `gorm:"default:0" json:"version"`
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedAt     time.Time `json:"created_at"`
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

// MqttMessage MQTT 报文记录（调试用）
type MqttMessage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DeviceSN  string    `gorm:"size:64;index:idx_mqtt_sn_ts" json:"device_sn"`
	Topic     string    `gorm:"size:256" json:"topic"`
	Direction string    `gorm:"size:8" json:"direction"` // up / down
	Payload   string    `gorm:"type:text" json:"payload"`
	CreatedAt time.Time `gorm:"index:idx_mqtt_sn_ts" json:"created_at"`
}

// ========== 规则引擎/场景联动 ==========

// Rule 规则引擎规则
type Rule struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:128" json:"name"`
	Description string    `gorm:"size:512" json:"description"`
	Type        string    `gorm:"size:32;default:auto" json:"type"` // auto / scene_linkage
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	TriggerJSON string    `gorm:"type:text" json:"trigger_json"` // 触发条件 JSON
	ActionJSON  string    `gorm:"type:text" json:"action_json"`  // 执行动作 JSON
	CreatedBy   uint      `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastRunAt   *time.Time `json:"last_run_at"`
	RunCount    int       `gorm:"default:0" json:"run_count"`
}

// RuleExecution 规则执行记录
type RuleExecution struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RuleID    uint      `gorm:"index" json:"rule_id"`
	TriggeredAt time.Time `json:"triggered_at"`
	Success   bool      `gorm:"default:true" json:"success"`
	ErrorMsg  string    `gorm:"size:512" json:"error_msg"`
	DetailJSON string   `gorm:"type:text" json:"detail_json"`
}

// DeviceGroup 设备分组
type DeviceGroup struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:128" json:"name"`
	Description string    `gorm:"size:512" json:"description"`
	ProductID   uint      `gorm:"index" json:"product_id"` // 0 表示跨产品分组
	OwnerID     uint      `gorm:"index" json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DeviceGroupMember 设备分组成员
type DeviceGroupMember struct {
	GroupID   uint `gorm:"primaryKey" json:"group_id"`
	DeviceID  uint `gorm:"primaryKey" json:"device_id"`
}

// DeviceTag 设备标签
type DeviceTag struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	DeviceID  uint   `gorm:"index:idx_device_tag" json:"device_id"`
	Key       string `gorm:"size:64;index:idx_device_tag" json:"key"`
	Value     string `gorm:"size:255" json:"value"`
}

// DeviceTopology 网关子设备拓扑关系
type DeviceTopology struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	GatewayID  uint      `gorm:"index" json:"gateway_id"`   // 网关设备 ID
	SubDeviceID uint     `gorm:"index" json:"sub_device_id"` // 子设备 ID
	Status     string    `gorm:"size:16;default:online" json:"status"` // online / offline
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// DeviceProvisioning 设备配网记录
type DeviceProvisioning struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ProductID   uint      `gorm:"index" json:"product_id"`
	DeviceSN    string    `gorm:"size:64;index" json:"device_sn"`
	DeviceName  string    `gorm:"size:128" json:"device_name"`
	Method      string    `gorm:"size:32" json:"method"`        // ble / softap / qrcode / zero
	Status      string    `gorm:"size:16;default:pending" json:"status"` // pending / success / failed / timeout
	SSID        string    `gorm:"size:64" json:"ssid"`          // WiFi SSID (SoftAP)
	BSSID       string    `gorm:"size:32" json:"bssid"`         // WiFi BSSID
	PinCode     string    `gorm:"size:16" json:"pin_code"`      // 配网 PIN 码
	QRCode      string    `gorm:"size:512" json:"qr_code"`      // 二维码内容
	ErrorMsg    string    `gorm:"size:512" json:"error_msg"`
	CreatedBy   uint      `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// ProvisioningConfig 配网配置模板
type ProvisioningConfig struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ProductID     uint      `gorm:"index" json:"product_id"`
	Method        string    `gorm:"size:32" json:"method"`        // ble / softap / qrcode
	BLEConfig     string    `gorm:"type:text" json:"ble_config"`  // 蓝牙配网参数 JSON
	SoftAPConfig  string    `gorm:"type:text" json:"softap_config"` // SoftAP 配网参数 JSON
	QRCodeConfig  string    `gorm:"type:text" json:"qrcode_config"` // 扫码配网参数 JSON
	CreatedBy     uint      `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// DataFlow 数据流转规则
type DataFlow struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:128" json:"name"`
	Description string    `gorm:"size:512" json:"description"`
	Type        string    `gorm:"size:32" json:"type"`        // republish / forward / sql
	SourceType  string    `gorm:"size:32" json:"source_type"` // topic / sql
	SourceTopic string    `gorm:"size:256" json:"source_topic"`
	SQL         string    `gorm:"type:text" json:"sql"`       // SQL 语句
	TargetType  string    `gorm:"size:32" json:"target_type"` // republish / mysql / timeseries / http
	TargetTopic string    `gorm:"size:256" json:"target_topic"`
	TargetConfig string   `gorm:"type:text" json:"target_config"` // 目标配置 JSON
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	CreatedBy   uint      `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DataFlowExecution 数据流转执行记录
type DataFlowExecution struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FlowID    uint      `gorm:"index" json:"flow_id"`
	Status    string    `gorm:"size:16" json:"status"` // success / failed
	ErrorMsg  string    `gorm:"size:512" json:"error_msg"`
	InputJSON string    `gorm:"type:text" json:"input_json"`
	OutputJSON string   `gorm:"type:text" json:"output_json"`
	CreatedAt time.Time `json:"created_at"`
}

// ========== 多协议网关 ==========

// ProtocolGateway 协议网关配置
type ProtocolGateway struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"size:128" json:"name"`
	Type       string    `gorm:"size:32" json:"type"`        // coap / http / mqtt / modbus
	Host       string    `gorm:"size:128" json:"host"`       // 监听地址
	Port       int       `json:"port"`                       // 监听端口
	Config     string    `gorm:"type:text" json:"config"`    // 协议特定配置 JSON
	Enabled    bool      `gorm:"default:true" json:"enabled"`
	CreatedBy  uint      `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ProtocolMapping 协议映射（将协议数据映射为物模型）
type ProtocolMapping struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	GatewayID   uint      `gorm:"index" json:"gateway_id"`
	ProductID   uint      `gorm:"index" json:"product_id"`
	ProtocolKey string    `gorm:"size:128" json:"protocol_key"` // 协议侧标识（如 Modbus 寄存器地址）
	PropertyID  string    `gorm:"size:64" json:"property_id"`   // 物模型属性标识符
	DataType    string    `gorm:"size:32" json:"data_type"`     // 数据类型转换
	Scale       float64   `json:"scale"`                        // 缩放因子
	Offset      float64   `json:"offset"`                       // 偏移量
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ========== 设备诊断/监控 ==========

// DeviceDiagnostic 设备诊断记录
type DeviceDiagnostic struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DeviceSN   string    `gorm:"size:64;index:idx_diag_sn_ts" json:"device_sn"`
	Type       string    `gorm:"size:32" json:"type"`        // connectivity / performance / firmware / security
	Level      string    `gorm:"size:16" json:"level"`       // info / warning / critical
	Title      string    `gorm:"size:128" json:"title"`
	Detail     string    `gorm:"type:text" json:"detail"`
	Status     string    `gorm:"size:16;default:open" json:"status"` // open / resolved / ignored
	CreatedAt  time.Time `gorm:"index:idx_diag_sn_ts" json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at"`
}

// DeviceMetrics 设备指标统计
type DeviceMetrics struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DeviceSN     string    `gorm:"size:64;index:idx_metrics_sn_ts" json:"device_sn"`
	OnlineTime   int64     `json:"online_time"`    // 在线时长(秒)
	OfflineTime  int64     `json:"offline_time"`   // 离线时长(秒)
	MessageCount int64     `json:"message_count"`  // 消息数
	ErrorCount   int64     `json:"error_count"`    // 错误数
	AvgLatency   float64   `json:"avg_latency"`    // 平均延迟(ms)
	Date         time.Time `gorm:"index:idx_metrics_sn_ts" json:"date"` // 统计日期
}
