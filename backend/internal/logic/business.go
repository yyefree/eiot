package logic

import (
	"bytes"
	"eiot/internal/dao"
	"eiot/internal/model"
	"eiot/pkg/cache"
	"eiot/pkg/mqtt"
	"eiot/pkg/util"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// BizError 业务错误
type BizError struct {
	HTTPCode int
	Code     int
	Message  string
}

func (e *BizError) Error() string { return e.Message }

func NewBizError(httpCode, code int, msg string) *BizError {
	return &BizError{HTTPCode: httpCode, Code: code, Message: msg}
}

// codeStore 保存验证码（内存存储，简化解法）
type codeItem struct {
	Code  string
	Until time.Time
}

// OnDeviceData 设备数据上报回调（由 handler 包设置，避免循环依赖）
var OnDeviceData func(deviceSN string, data map[string]interface{})

var (
	codeMap = map[string]*codeItem{}
	codeMu  sync.RWMutex
)

func init() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			codeMu.Lock()
			now := time.Now()
			for k, v := range codeMap {
				if now.After(v.Until) {
					delete(codeMap, k)
				}
			}
			codeMu.Unlock()
		}
	}()
}

// ========== 认证 ==========

// LoginByPassword 账号密码登录
func LoginByPassword(phone, password, secret string) (string, *model.User, error) {
	if phone == "" || password == "" {
		return "", nil, NewBizError(http.StatusBadRequest, 400, "账号或密码不能为空")
	}
	var u model.User
	if err := dao.DB.Where("phone = ? OR username = ?", phone, phone).First(&u).Error; err != nil {
		return "", nil, NewBizError(http.StatusUnauthorized, 401, "账号或密码错误")
	}
	if u.Status != 1 {
		return "", nil, NewBizError(http.StatusForbidden, 403, "账号已禁用")
	}
	// 仅使用 bcrypt 校验，禁止明文密码 fallback
	if !util.CheckPassword(password, u.Password) {
		return "", nil, NewBizError(http.StatusUnauthorized, 401, "账号或密码错误")
	}
	token, err := util.GenerateJWT(u.ID, u.Role, secret)
	return token, &u, err
}


// SendCode 发送验证码
func SendCode(phoneOrEmail string) (string, error) {
	if phoneOrEmail == "" {
		return "", errors.New("请输入手机号或邮箱")
	}
	code := util.GenerateCode()
	codeMu.Lock()
	codeMap[phoneOrEmail] = &codeItem{Code: code, Until: time.Now().Add(10 * time.Minute)}
	codeMu.Unlock()
	return code, nil
}

// GetLastCode 获取最近生成的验证码（仅开发环境使用）
func GetLastCode(phoneOrEmail string) string {
	codeMu.RLock()
	defer codeMu.RUnlock()
	if item, ok := codeMap[phoneOrEmail]; ok {
		return item.Code
	}
	return ""
}

// LoginOrRegisterByCode 使用验证码登录或注册
func LoginOrRegisterByCode(phone, code, secret string) (string, *model.User, error) {
	codeMu.Lock()
	item, ok := codeMap[phone]
	if !ok || item == nil || time.Now().After(item.Until) {
		codeMu.Unlock()
		return "", nil, errors.New("验证码无效或已过期")
	}
	if subtle.ConstantTimeCompare([]byte(item.Code), []byte(code)) != 1 {
		codeMu.Unlock()
		return "", nil, errors.New("验证码错误")
	}
	delete(codeMap, phone)
	codeMu.Unlock()

	nickname := "用户"
	if len(phone) >= 4 {
		nickname = "用户" + phone[len(phone)-4:]
	}

	var u model.User
	err := dao.DB.Where("phone = ?", phone).First(&u).Error
	if err != nil {
		hashed, _ := util.HashPassword("123456")
		u = model.User{
			Username: "u_" + phone, Phone: phone, Password: hashed,
			Nickname: nickname,
			Role: "user", Status: 1,
		}
		if err := dao.DB.Create(&u).Error; err != nil {
			return "", nil, err
		}
	}
	if u.Status != 1 {
		return "", nil, errors.New("账号已禁用")
	}
	token, _ := util.GenerateJWT(u.ID, u.Role, secret)
	return token, &u, nil
}

// RegisterByPassword 密码注册（开发模式，无需验证码）
func RegisterByPassword(phone, password, nickname, secret string) (string, error) {
	if phone == "" || password == "" {
		return "", errors.New("手机号和密码不能为空")
	}
	if len(password) < 6 {
		return "", errors.New("密码长度不能少于6位")
	}
	// 检查手机号是否已注册
	var count int64
	dao.DB.Model(&model.User{}).Where("phone = ?", phone).Count(&count)
	if count > 0 {
		return "", errors.New("该手机号已注册")
	}
	if nickname == "" {
		nickname = "用户"
		if len(phone) >= 4 {
			nickname = "用户" + phone[len(phone)-4:]
		}
	}
	hashed, err := util.HashPassword(password)
	if err != nil {
		return "", err
	}
	u := model.User{
		Username: "u_" + phone, Phone: phone, Password: hashed,
		Nickname: nickname,
		Role: "user", Status: 1,
	}
	if err := dao.DB.Create(&u).Error; err != nil {
		return "", err
	}
	token, err := util.GenerateJWT(u.ID, u.Role, secret)
	return token, err
}

// RefreshToken 刷新 Token（简单实现：从旧 token 解析 uid 重新签发）
func RefreshToken(refreshToken, secret string) (string, error) {
	claims, err := util.ParseJWT(refreshToken, secret)
	if err != nil {
		return "", errors.New("refresh_token 无效或已过期")
	}
	uidFloat, ok := claims["uid"].(float64)
	if !ok {
		return "", errors.New("无效的 token payload")
	}
	var u model.User
	if err := dao.DB.First(&u, uint(uidFloat)).Error; err != nil {
		return "", errors.New("用户不存在")
	}
	if u.Status != 1 {
		return "", errors.New("账号已禁用")
	}
	return util.GenerateJWT(u.ID, u.Role, secret)
}

// GetUserInfo 获取用户信息
func GetUserInfo(uid uint) (*model.User, error) {
	var u model.User
	if err := dao.DB.First(&u, uid).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	// 脱敏密码等
	u.Password = ""
	return &u, nil
}

// ========== 初始化 ==========

// InitAdmin 确保至少存在一个 admin
func InitAdmin(phone, password string) {
	if phone == "" {
		phone = "13800000000"
	}
	if password == "" {
		password = "admin123"
	}
	var count int64
	dao.DB.Model(&model.User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		return
	}
	hashed, _ := util.HashPassword(password)
	dao.DB.Create(&model.User{
		Username: "admin", Phone: phone, Password: hashed,
		Nickname: "超级管理员", Role: "admin", Status: 1,
	})
}

// SeedDemoData 生成演示数据
func SeedDemoData(adminPhone, adminPassword string) {
	InitAdmin(adminPhone, adminPassword)
	// 取管理员
	var admin model.User
	dao.DB.Where("role = ?", "admin").First(&admin)
	// 普通用户
	userCount := int64(0)
	dao.DB.Model(&model.User{}).Where("role = ?", "user").Count(&userCount)
	if userCount == 0 {
		for i := 1; i <= 5; i++ {
			phone := fmt.Sprintf("139%08d", i)
			hashed, _ := util.HashPassword("123456")
			dao.DB.Create(&model.User{
				Username: fmt.Sprintf("user_%02d", i), Phone: phone, Password: hashed,
				Nickname: fmt.Sprintf("用户%02d", i), Role: "user", Status: 1,
			})
		}
	}
	// 物模型合一（产品
	var prodCount int64
	dao.DB.Model(&model.Product{}).Count(&prodCount)
	if prodCount == 0 {

	// 预定义物模型属性 JSON
	propsTemp := `[{"identifier":"temp_01","name":"温度","accessMode":"r","dataType":{"type":"float","specs":{"min":-40,"max":85,"step":0.1,"unit":"℃"}}},{"identifier":"hum_01","name":"湿度","accessMode":"r","dataType":{"type":"float","specs":{"min":0,"max":100,"step":1,"unit":"%"}}}]`
	propsSwitch := `[{"identifier":"switch_01","name":"开关","accessMode":"rw","dataType":{"type":"bool","specs":{}}},{"identifier":"power","name":"功率","accessMode":"r","dataType":{"type":"float","specs":{"min":0,"max":3000,"step":1,"unit":"W"}}}]`
	propsEnergy := `[{"identifier":"voltage","name":"电压","accessMode":"r","dataType":{"type":"float","specs":{"min":0,"max":260,"step":1,"unit":"V"}}},{"identifier":"current","name":"电流","accessMode":"r","dataType":{"type":"float","specs":{"min":0,"max":100,"step":0.1,"unit":"A"}}},{"identifier":"energy","name":"能耗","accessMode":"r","dataType":{"type":"float","specs":{"min":0,"max":100000,"step":1,"unit":"kWh"}}}]`
	propsCurtain := `[{"identifier":"open_pct","name":"开合度","accessMode":"rw","dataType":{"type":"int","specs":{"min":0,"max":100,"step":1,"unit":"%"}}},{"identifier":"motor","name":"电机状态","accessMode":"rw","dataType":{"type":"bool","specs":{}}}]`
	propsAir := `[{"identifier":"pm25","name":"PM2.5","accessMode":"r","dataType":{"type":"float","specs":{"min":0,"max":1000,"step":1,"unit":"μg/m³"}}},{"identifier":"co2","name":"CO₂","accessMode":"r","dataType":{"type":"int","specs":{"min":0,"max":5000,"step":1,"unit":"ppm"}}},{"identifier":"tvoc","name":"TVOC","accessMode":"r","dataType":{"type":"float","specs":{"min":0,"max":10,"step":0.1,"unit":"mg/m³"}}}]`

	products := []struct {
		Name        string
		ProductKey  string
		Description string
		PropsJSON   string
		DeviceCount int
	}{
		{"温湿度传感器", "PK_TEMP001", "采集环境温度、湿度", propsTemp, 10},
		{"智能开关", "PK_SWITCH001", "远程控制通断", propsSwitch, 10},
		{"能耗监测仪", "PK_ENERGY001", "监测电压、电流、能耗", propsEnergy, 10},
		{"智能窗帘", "PK_CURTAIN001", "远程控制窗帘开合", propsCurtain, 10},
		{"空气质量监测", "PK_AIR001", "PM2.5/CO2/TVOC", propsAir, 10},
	}

	for _, s := range products {
		prod := model.Product{
			Name: s.Name, ProductKey: s.ProductKey, Description: s.Description,
			PropertiesJSON: s.PropsJSON, EventsJSON: `[]`, ServicesJSON: `[]`,
			CreatedBy: admin.ID,
		}
		dao.DB.Create(&prod)
		for i := 1; i <= s.DeviceCount; i++ {
			sn := fmt.Sprintf("%sSN%05d", s.ProductKey, i)
			dao.DB.Create(&model.Device{
				DeviceName: fmt.Sprintf("%s-%03d", s.Name, i),
				DeviceSN:   sn, DeviceSecret: "secret_" + sn,
				ProductID: prod.ID, ProductKey: s.ProductKey, OwnerID: admin.ID, Status: 1,
			})
		}
	}

	// 共享：给用户1绑定一台
	var users []model.User
	dao.DB.Where("role = ?", "user").Order("id asc").Find(&users)
	var devices []model.Device
	dao.DB.Limit(3).Find(&devices)
	if len(users) >= 1 && len(devices) >= 1 {
		dao.DB.Model(&devices[0]).Update("owner_id", users[0].ID)
		if len(users) >= 2 {
			dao.DB.Create(&model.DeviceShare{DeviceID: devices[1].ID, ShareUserID: users[1].ID, Permission: "read", CreatedBy: users[0].ID})
		}
	}

	// 填充最新模拟数据
	var allDevices []model.Device
	dao.DB.Find(&allDevices)
	for _, d := range allDevices {
		cache.SetOnline(d.DeviceSN, true, 300)
		cache.SetLatest(d.DeviceSN, "temperature", fmt.Sprintf("%.1f", 20+rand.Float64()*10))
		cache.SetLatest(d.DeviceSN, "humidity", fmt.Sprintf("%.1f", 40+rand.Float64()*30))
		cache.SetLatest(d.DeviceSN, "voltage", fmt.Sprintf("%.1f", 220+rand.Float64()*10))
		cache.SetLatest(d.DeviceSN, "current", fmt.Sprintf("%.2f", rand.Float64()*5))
		now := time.Now()
		dao.DB.Model(&d).Updates(map[string]interface{}{"last_online": &now})
	}

	// 默认看板
	var dc int64
	dao.DB.Model(&model.Dashboard{}).Where("is_default = ?", 1).Count(&dc)
	if dc == 0 {
		dao.DB.Create(&model.Dashboard{
			OwnerID: 0, Type: "web", Name: "默认看板",
			LayoutJSON: `{"widgets":[{"type":"stat","title":"设备总数"},{"type":"chart","title":"近24小时能耗"}]}`,
			IsDefault:  1,
		})
	}
	} // end if prodCount == 0

	// ========== 默认家庭/房间/场景（云智能App）==========
	// （无论产品是否已存在，都检查家庭数据）
	var homeCount int64
	dao.DB.Model(&model.Home{}).Count(&homeCount)
	if homeCount == 0 {
		home := model.Home{Name: "我的家", Address: "示例地址", Icon: "home", OwnerID: admin.ID}
		dao.DB.Create(&home)
		dao.DB.Create(&model.HomeMember{HomeID: home.ID, UserID: admin.ID, Role: "owner", Nickname: admin.Nickname})

		// 房间
		rooms := []model.Room{
			{HomeID: home.ID, Name: "客厅", Icon: "living_room", SortOrder: 0},
			{HomeID: home.ID, Name: "卧室", Icon: "bedroom_parent", SortOrder: 1},
			{HomeID: home.ID, Name: "厨房", Icon: "kitchen", SortOrder: 2},
			{HomeID: home.ID, Name: "书房", Icon: "desk", SortOrder: 3},
		}
		for _, r := range rooms {
			dao.DB.Create(&r)
		}

		// 把前几个设备分配到房间
		var devices []model.Device
		dao.DB.Limit(8).Find(&devices)
		var roomList []model.Room
		dao.DB.Where("home_id = ?", home.ID).Order("sort_order asc").Find(&roomList)
		if len(roomList) > 0 && len(devices) > 0 {
			for i, d := range devices {
				rid := roomList[i%len(roomList)].ID
				dao.DB.Create(&model.RoomDevice{RoomID: rid, DeviceID: d.ID})
			}
		}

		// 场景
		scene1 := model.Scene{HomeID: home.ID, Name: "回家模式", Icon: "home", Type: "manual", Enabled: true, SortOrder: 0}
		dao.DB.Create(&scene1)
		scene2 := model.Scene{HomeID: home.ID, Name: "离家模式", Icon: "logout", Type: "manual", Enabled: true, SortOrder: 1}
		dao.DB.Create(&scene2)
		scene3 := model.Scene{HomeID: home.ID, Name: "睡眠模式", Icon: "bedtime", Type: "manual", Enabled: true, SortOrder: 2}
		dao.DB.Create(&scene3)

		// 给场景添加示例动作（如果有设备的话）
		if len(devices) > 0 {
			dao.DB.Create(&model.SceneAction{SceneID: scene1.ID, DeviceID: devices[0].ID, ActionJSON: `{"switch_01":true}`, SortOrder: 0})
			dao.DB.Create(&model.SceneAction{SceneID: scene2.ID, DeviceID: devices[0].ID, ActionJSON: `{"switch_01":false}`, SortOrder: 0})
			if len(devices) > 1 {
				dao.DB.Create(&model.SceneAction{SceneID: scene3.ID, DeviceID: devices[1].ID, ActionJSON: `{"switch_01":false}`, SortOrder: 0})
			}
		}

		// 示例消息
		dao.DB.Create(&model.Message{UserID: admin.ID, Type: "system", Title: "欢迎使用飞燕IoT平台", Content: "您的账户已成功创建，开始探索智能物联世界吧！"})
		dao.DB.Create(&model.Message{UserID: admin.ID, Type: "device", Title: "设备上线通知", Content: "温湿度传感器-001 已上线"})
	}
}

// ========== 项目（飞燕标准层级 ==========

// CreateProject 创建项目
func CreateProject(name, description, industry, projType string, createdBy uint) (*model.Project, error) {
	if name == "" {
		return nil, errors.New("项目名称不可为空")
	}
	if industry == "" {
		industry = "智能家居"
	}
	p := &model.Project{
		Name: name, Description: description, Industry: industry, Type: projType,
		OwnerID: createdBy, CreatedBy: createdBy,
	}
	if err := dao.DB.Create(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

// ListProjects 项目列表
func ListProjects(page, size int) (int64, []model.Project, error) {
	var total int64
	var list []model.Project
	dao.DB.Model(&model.Project{}).Count(&total)
	err := dao.DB.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return total, list, err
}

// UpdateProject 更新项目
func UpdateProject(id uint, name, description, industry, projType string) error {
	updates := map[string]interface{}{"updated_at": time.Now()}
	if name != "" {
		updates["name"] = name
	}
	if description != "" {
		updates["description"] = description
	}
	if industry != "" {
		updates["industry"] = industry
	}
	if projType != "" {
		updates["type"] = projType
	}
	return dao.DB.Model(&model.Project{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteProject 删除项目（仅当项目下无产品时允许）
func DeleteProject(id uint) error {
	var prodCount int64
	dao.DB.Model(&model.Product{}).Where("project_id = ?", id).Count(&prodCount)
	if prodCount > 0 {
		return errors.New("项目下存在产品，请先删除产品后再删除项目")
	}
	return dao.DB.Delete(&model.Project{}, id).Error
}

// ========== 产品（物模型 ==========

// CreateProduct 创建产品（物模型
func CreateProduct(name, description, icon, productKey, propertiesJSON, eventsJSON, servicesJSON string, projectID uint, networkType string, createdBy uint) (*model.Product, error) {
	if name == "" {
		return nil, errors.New("产品名称不可为空")
	}
	if productKey == "" {
		productKey = "PK_" + strings.ToUpper(strconv.FormatInt(time.Now().Unix(), 36))
	}
	if networkType == "" {
		networkType = "wifi"
	}
	p := &model.Product{
		Name: name, Description: description, Icon: icon, ProductKey: productKey,
		ProjectID: projectID, NetworkType: networkType,
		PropertiesJSON: propertiesJSON, EventsJSON: eventsJSON, ServicesJSON: servicesJSON,
		CreatedBy: createdBy,
	}
	if err := dao.DB.Create(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

// ListProducts 产品列表
func ListProducts(page, size int, projectID uint) (int64, []model.Product, error) {
	var total int64
	var list []model.Product
	q := dao.DB.Model(&model.Product{})
	if projectID > 0 {
		q = q.Where("project_id = ?", projectID)
	}
	q.Count(&total)
	err := q.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return total, list, err
}

// GetProduct 获取产品详情
func GetProduct(id uint) (*model.Product, error) {
	var p model.Product
	if err := dao.DB.First(&p, id).Error; err != nil {
		return nil, errors.New("产品不存在")
	}
	return &p, nil
}

// UpdateProduct 更新产品基础信息
func UpdateProduct(id uint, name, description, icon string, projectID uint, networkType string) error {
	if name == "" {
		return errors.New("名称不可为空")
	}
	updates := map[string]interface{}{"updated_at": time.Now()}
	if name != "" {
		updates["name"] = name
	}
	if description != "" {
		updates["description"] = description
	}
	if icon != "" {
		updates["icon"] = icon
	}
	updates["project_id"] = projectID
	if networkType != "" {
		updates["network_type"] = networkType
	}
	return dao.DB.Model(&model.Product{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateProductThingModel 更新产品物模型
// 检查产品下有设备则不允许修改物模型结构
func UpdateProductThingModel(id uint, propertiesJSON, eventsJSON, servicesJSON string) error {
	// 检查是否已有设备
	var devCount int64
	dao.DB.Model(&model.Device{}).Where("product_id = ?", id).Count(&devCount)
	if devCount > 0 {
		return errors.New("产品下已有设备，物模型结构锁定，不可修改")
	}
	return dao.DB.Model(&model.Product{}).Where("id = ?", id).Updates(
		map[string]interface{}{
			"properties_json": propertiesJSON,
			"events_json":     eventsJSON,
			"services_json":   servicesJSON,
			"updated_at":      time.Now(),
		},
	).Error
}

// DeleteProduct 删除产品
func DeleteProduct(id uint) error {
	// 检查是否有设备
	var devCount int64
	dao.DB.Model(&model.Device{}).Where("product_id = ?", id).Count(&devCount)
	if devCount > 0 {
		return errors.New("产品下仍有设备，不可删除")
	}
	return dao.DB.Delete(&model.Product{}, id).Error
}

// GetProductMobileUI 获取产品移动端 UI 配置
func GetProductMobileUI(id uint) (string, error) {
	var p model.Product
	if err := dao.DB.First(&p, id).Error; err != nil {
		return "", errors.New("产品不存在")
	}
	return p.MobileUIJSON, nil
}

// SaveProductMobileUI 保存产品移动端 UI 配置
func SaveProductMobileUI(id uint, uiJSON string) error {
	var p model.Product
	if err := dao.DB.First(&p, id).Error; err != nil {
		return errors.New("产品不存在")
	}
	return dao.DB.Model(&p).Updates(map[string]interface{}{"mobile_ui_json": uiJSON, "updated_at": time.Now()}).Error
}

// ========== 设备 ==========

// BatchGenerateDevices 批量生成设备
func BatchGenerateDevices(productID uint, prefix string, count int) ([]model.Device, error) {
	return BatchGenerateDevicesWithMode(productID, prefix, count, "device_secret")
}

func BatchGenerateDevicesWithMode(productID uint, prefix string, count int, bindMode string) ([]model.Device, error) {
	if productID == 0 {
		return nil, errors.New("产品ID不可为空")
	}
	if count <= 0 || count > 5000 {
		count = 10
	}
	if prefix == "" {
		prefix = "DEV"
	}
	if bindMode == "" {
		bindMode = "device_secret"
	}
	var prod model.Product
	if err := dao.DB.First(&prod, productID).Error; err != nil {
		return nil, errors.New("产品不存在")
	}
	devs := make([]model.Device, 0, count)
	ts := time.Now().UnixNano()
	// 一型一密：同产品共用 ProductSecret
	productSecret := ""
	if bindMode == "product_secret" {
		productSecret = fmt.Sprintf("ps_%s_%d", prod.ProductKey, ts/1000000)
	}
	for i := 1; i <= count; i++ {
		sn := fmt.Sprintf("%s%08d%04d", prod.ProductKey, ts%100000000, i)
		d := model.Device{
			DeviceName: fmt.Sprintf("%s-%03d", prefix, i),
			DeviceSN:   sn,
			ProductID:  prod.ID,
			ProductKey: prod.ProductKey,
			BindMode:   bindMode,
			Status:     1,
		}
		if bindMode == "device_secret" {
			d.DeviceSecret = fmt.Sprintf("ds_%s_%d", sn, ts%100000)
		} else {
			d.ProductSecret = productSecret
		}
		devs = append(devs, d)
	}
	if err := dao.DB.Create(&devs).Error; err != nil {
		return nil, err
	}
	return devs, nil
}

// ListDevices 设备列表（admin 全量/用户视角
func ListDevices(uid uint, role string, page, size int, productID uint) (int64, []model.Device, error) {
	var total int64
	var list []model.Device
	query := dao.DB.Model(&model.Device{})
	if productID > 0 {
		query = query.Where("product_id = ?", productID)
	}
	if role != "admin" {
		shared := []uint{}
		dao.DB.Table("device_shares").Where("share_user_id = ?", uid).Pluck("device_id", &shared)
		inIDs := append([]uint{}, shared...)
		if uid > 0 {
			query = query.Where("owner_id = ? OR id IN ?", uid, inIDs)
		}
	}
	query.Count(&total)
	err := query.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return total, list, err
}

// GetDeviceDetail 获取设备详情
func GetDeviceDetail(uid uint, role string, id uint) (*model.Device, *model.Product, map[string]string, error) {
	var d model.Device
	if err := dao.DB.First(&d, id).Error; err != nil {
		return nil, nil, nil, errors.New("设备不存在")
	}
	// 权限检查
	if role != "admin" {
		if d.OwnerID != uid {
			var c int64
			dao.DB.Model(&model.DeviceShare{}).Where("device_id = ? AND share_user_id = ?", id, uid).Count(&c)
			if c == 0 {
				return nil, nil, nil, errors.New("无权限查看该设备")
			}
		}
	}
	// 产品信息
	var prod model.Product
	dao.DB.First(&prod, d.ProductID)
	// 最新数据
	latest := cache.GetLatest(d.DeviceSN)
	return &d, &prod, latest, nil
}

// ControlDevice 下发设备控制指令
func ControlDevice(uid uint, role string, id uint, params map[string]interface{}, emqx interface{}) error {
	var d model.Device
	if err := dao.DB.First(&d, id).Error; err != nil {
		return errors.New("设备不存在")
	}
	// 权限检查：管理员 / 设备所有者 / 拥有可写共享权限
	if role != "admin" && d.OwnerID != uid {
		var share model.DeviceShare
		err := dao.DB.Where("device_id = ? AND share_user_id = ?", id, uid).First(&share).Error
		if err != nil {
			return errors.New("无权限控制该设备")
		}
		// 共享权限必须包含写权限（兼容 control / readwrite / rw
		if share.Permission != "control" && share.Permission != "readwrite" && share.Permission != "rw" {
			return errors.New("当前共享权限仅支持查看，不可控制设备")
		}
	}

	// 通过 EMQX 下发
	if emqx != nil {
		if cli, ok := emqx.(interface{ Publish(string, interface{}) error }); ok {
			topic := fmt.Sprintf("/sys/cmd/%s", d.DeviceSN)
			payload := map[string]interface{}{
				"product_key": d.ProductKey,
				"params":      params,
				"ts":          time.Now().Unix(),
			}
			_ = cli.Publish(topic, payload)
			data, _ := json.Marshal(payload)
			LogMqttMessage(d.DeviceSN, topic, "down", string(data))
		}
	}
	// 记录日志
	buf, _ := json.Marshal(params)
	dao.DB.Create(&model.OperationLog{
		UserID: uid, Action: "control", Target: fmt.Sprintf("device:%d", id),
		Detail: fmt.Sprintf("下发控制指令: %s", string(buf)), CreatedAt: time.Now(),
	})
	return nil
}

// HandleDeviceReport 处理设备上报（通过 SN）
func HandleDeviceReport(sn string, params map[string]interface{}) {
	HandleDeviceReportByPKAndName("", sn, params)
}

// HandleDeviceReportByPKAndName 通过 ProductKey + DeviceName 处理设备上报（飞燕标准
func HandleDeviceReportByPKAndName(productKey, deviceName string, params map[string]interface{}) {
	if params == nil {
		return
	}
	var d model.Device
	var err error
	if productKey != "" && deviceName != "" {
		// 优先用 PK+Name 精确匹配（飞燕标准）
		err = dao.DB.Where("product_key = ? AND device_name = ?", productKey, deviceName).First(&d).Error
	} else {
		err = dao.DB.Where("device_sn = ?", deviceName).First(&d).Error // 兼容旧方式（直接按 SN 匹配）
	}
	if err != nil {
		return
	}
	sn := d.DeviceSN
	cache.SetOnline(sn, true, 300)
	wsData := map[string]interface{}{}
	for k, v := range params {
		cache.SetLatest(sn, k, fmt.Sprintf("%v", v))
		SaveDeviceData(sn, k, fmt.Sprintf("%v", v))
		wsData[k] = v
	}
	// 广播到 WebSocket 客户端（通过回调避免循环依赖）
	if OnDeviceData != nil {
		OnDeviceData(sn, wsData)
	}
	now := time.Now()
	updates := map[string]interface{}{"last_online": &now, "updated_at": now}
	// 记录上报IP（如果有）
	if ip, ok := params["ip"].(string); ok && ip != "" {
		updates["ip_address"] = ip
	}
	dao.DB.Model(&d).Updates(updates)
}

// HandleDeviceFirmwareReport 处理固件版本上报（飞燕标准 /ota/device/inform/
func HandleDeviceFirmwareReport(productKey, deviceName, version string) {
	var d model.Device
	dao.DB.Where("product_key = ? AND device_name = ?", productKey, deviceName).First(&d)
	if d.ID == 0 {
		return
	}
	dao.DB.Model(&d).Update("firmware_ver", version)
}

// BindDeviceBySN 绑定设备（用户扫码绑定
func BindDeviceBySN(uid uint, sn string) (*model.Device, error) {
	if sn == "" {
		return nil, errors.New("设备SN不可为空")
	}
	var d model.Device
	if err := dao.DB.Where("device_sn = ?", sn).First(&d).Error; err != nil {
		return nil, errors.New("设备不存在")
	}
	if d.OwnerID != 0 {
		return nil, errors.New("该设备已被其他用户绑定")
	}
	dao.DB.Model(&d).Update("owner_id", uid)
	return &d, nil
}

// ListUsers 用户列表
func ListUsers(page, size int) (int64, []model.User, error) {
	var total int64
	var list []model.User
	dao.DB.Model(&model.User{}).Count(&total)
	err := dao.DB.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return total, list, err
}

// ========== 共享 ==========

// ShareDevice 共享设备
func ShareDevice(uid uint, deviceID uint, shareUID uint, permission string, exp time.Time) error {
	if deviceID == 0 || shareUID == 0 {
		return errors.New("参数错误")
	}
	var d model.Device
	dao.DB.First(&d, deviceID)
	if d.ID == 0 {
		return errors.New("设备不存在")
	}
	var u model.User
	dao.DB.First(&u, uid)
	if d.OwnerID != uid && u.Role != "admin" {
		return errors.New("仅设备所有者可发起共享")
	}
	share := model.DeviceShare{
		DeviceID: deviceID, ShareUserID: shareUID, Permission: permission, CreatedBy: uid,
	}
	if !exp.IsZero() {
		share.ExpireAt = &exp
	}
	return dao.DB.Create(&share).Error
}

// RevokeShare 取消共享
func RevokeShare(uid uint, id uint) error {
	var s model.DeviceShare
	if err := dao.DB.First(&s, id).Error; err != nil {
		return errors.New("共享记录不存在")
	}
	var d model.Device
	dao.DB.First(&d, s.DeviceID)
	if d.OwnerID != uid {
		return errors.New("无权限取消他人共享")
	}
	return dao.DB.Delete(&s).Error
}

// ListShares 共享列表（当前用户发起的共享
func ListShares(uid uint, page, size int) (int64, []map[string]interface{}, error) {
	var total int64
	raw := []map[string]interface{}{}
	dao.DB.Table("device_shares").Where("created_by = ?", uid).Count(&total).
		Order("id desc").Offset((page-1)*size).Limit(size).Find(&raw)
	return total, raw, nil
}

// ========== 用户个人设备 UI 配置 ==========

// GetUserDeviceUI 获取用户对某设备的个人 UI 配置
func GetUserDeviceUI(uid uint, deviceID uint) (string, error) {
	var ui model.UserDeviceUI
	err := dao.DB.Where("user_id = ? AND device_id = ?", uid, deviceID).First(&ui).Error
	if err != nil {
		return "", nil
	}
	return ui.LayoutJSON, nil
}

// SaveUserDeviceUI 保存用户个人 UI 配置
func SaveUserDeviceUI(uid uint, deviceID uint, layoutJSON string) error {
	var existing model.UserDeviceUI
	err := dao.DB.Where("user_id = ? AND device_id = ?", uid, deviceID).First(&existing).Error
	if err != nil {
		return dao.DB.Create(&model.UserDeviceUI{UserID: uid, DeviceID: deviceID, LayoutJSON: layoutJSON}).Error
	}
	return dao.DB.Model(&existing).Updates(map[string]interface{}{"layout_json": layoutJSON, "updated_at": time.Now()}).Error
}

// ========== 看板 ==========

func SaveDashboard(uid uint, typ string, name, layout string) error {
	var existing model.Dashboard
	err := dao.DB.Where("owner_id = ? AND type = ?", uid, typ).First(&existing).Error
	if err != nil {
		return dao.DB.Create(&model.Dashboard{OwnerID: uid, Type: typ, Name: name, LayoutJSON: layout}).Error
	}
	return dao.DB.Model(&existing).Updates(map[string]interface{}{"name": name, "layout_json": layout, "updated_at": time.Now()}).Error
}

func ListDashboards(uid uint, typ string) ([]model.Dashboard, error) {
	var list []model.Dashboard
	q := dao.DB.Where("owner_id = ? OR is_default = 1", uid)
	if typ != "" {
		q = q.Where("type = ?", typ)
	}
	err := q.Order("id desc").Find(&list).Error
	return list, err
}

// ========== 家庭 ==========

// CreateHome 创建家庭
func CreateHome(name, address, icon string, ownerID uint) (*model.Home, error) {
	if name == "" {
		return nil, errors.New("家庭名称不可为空")
	}
	h := &model.Home{
		Name: name, Address: address, Icon: icon, OwnerID: ownerID,
	}
	if err := dao.DB.Create(h).Error; err != nil {
		return nil, err
	}
	// 自动把创建者添加为家庭 owner
	dao.DB.Create(&model.HomeMember{HomeID: h.ID, UserID: ownerID, Role: "owner", Nickname: ""})
	return h, nil
}

// ListHomes 家庭列表（包含用户作为成员的家庭）
func ListHomes(uid uint) ([]model.Home, error) {
	var list []model.Home
	// 查找用户作为成员的家庭ID
	var memberHomeIDs []uint
	dao.DB.Model(&model.HomeMember{}).Where("user_id = ?", uid).Pluck("home_id", &memberHomeIDs)
	// 合并用户拥有的和作为成员的
	query := dao.DB.Where("owner_id = ?", uid)
	if len(memberHomeIDs) > 0 {
		query = query.Or("id IN ?", memberHomeIDs)
	}
	err := query.Order("id desc").Find(&list).Error
	return list, err
}

// UpdateHome 更新家庭
func UpdateHome(id uint, name, address, icon string) error {
	if name == "" {
		return errors.New("家庭名称不可为空")
	}
	updates := map[string]interface{}{"updated_at": time.Now()}
	updates["name"] = name
	if address != "" {
		updates["address"] = address
	}
	if icon != "" {
		updates["icon"] = icon
	}
	return dao.DB.Model(&model.Home{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteHome 删除家庭
func DeleteHome(id uint) error {
	// 删除关联房间设备
	var rooms []model.Room
	dao.DB.Where("home_id = ?", id).Find(&rooms)
	roomIDs := make([]uint, 0)
	for _, r := range rooms {
		roomIDs = append(roomIDs, r.ID)
	}
	if len(roomIDs) > 0 {
		dao.DB.Where("room_id IN ?", roomIDs).Delete(&model.RoomDevice{})
	}
	dao.DB.Where("home_id = ?", id).Delete(&model.Room{})
	dao.DB.Where("home_id = ?", id).Delete(&model.Scene{})
	dao.DB.Where("home_id = ?", id).Delete(&model.HomeMember{})
	return dao.DB.Delete(&model.Home{}, id).Error
}

// GetHomeDetail 获取家庭详情（含成员和房间数）
func GetHomeDetail(id uint) (*model.Home, []model.HomeMember, int64, error) {
	var h model.Home
	if err := dao.DB.First(&h, id).Error; err != nil {
		return nil, nil, 0, errors.New("家庭不存在")
	}
	var members []model.HomeMember
	dao.DB.Where("home_id = ?", id).Find(&members)
	var roomCount int64
	dao.DB.Model(&model.Room{}).Where("home_id = ?", id).Count(&roomCount)
	return &h, members, roomCount, nil
}

// ========== 家庭成员 ==========

// AddHomeMember 添加家庭成员
func AddHomeMember(homeID, userID uint, role, nickname string) (*model.HomeMember, error) {
	// 检查家庭是否存在
	var h model.Home
	if err := dao.DB.First(&h, homeID).Error; err != nil {
		return nil, errors.New("家庭不存在")
	}
	// 检查是否已是成员
	var count int64
	dao.DB.Model(&model.HomeMember{}).Where("home_id = ? AND user_id = ?", homeID, userID).Count(&count)
	if count > 0 {
		return nil, errors.New("用户已是该家庭成员")
	}
	if role == "" {
		role = "member"
	}
	m := &model.HomeMember{HomeID: homeID, UserID: userID, Role: role, Nickname: nickname}
	if err := dao.DB.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// RemoveHomeMember 移除家庭成员
func RemoveHomeMember(homeID, userID uint) error {
	result := dao.DB.Where("home_id = ? AND user_id = ?", homeID, userID).Delete(&model.HomeMember{})
	if result.RowsAffected == 0 {
		return errors.New("成员不存在")
	}
	return result.Error
}

// ListHomeMembers 家庭成员列表
func ListHomeMembers(homeID uint) ([]model.HomeMember, error) {
	var list []model.HomeMember
	err := dao.DB.Where("home_id = ?", homeID).Find(&list).Error
	return list, err
}

// ========== 房间 ==========

// CreateRoom 创建房间
func CreateRoom(homeID uint, name, icon string, sortOrder int) (*model.Room, error) {
	if name == "" {
		return nil, errors.New("房间名称不可为空")
	}
	// 检查家庭是否存在
	var h model.Home
	if err := dao.DB.First(&h, homeID).Error; err != nil {
		return nil, errors.New("家庭不存在")
	}
	r := &model.Room{
		HomeID: homeID, Name: name, Icon: icon, SortOrder: sortOrder,
	}
	if err := dao.DB.Create(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

// ListRooms 房间列表
func ListRooms(homeID uint) ([]model.Room, error) {
	var list []model.Room
	err := dao.DB.Where("home_id = ?", homeID).Order("sort_order asc, id asc").Find(&list).Error
	return list, err
}

// UpdateRoom 更新房间
func UpdateRoom(id uint, name, icon string) error {
	if name == "" {
		return errors.New("房间名称不可为空")
	}
	updates := map[string]interface{}{"updated_at": time.Now(), "name": name}
	if icon != "" {
		updates["icon"] = icon
	}
	return dao.DB.Model(&model.Room{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteRoom 删除房间
func DeleteRoom(id uint) error {
	dao.DB.Where("room_id = ?", id).Delete(&model.RoomDevice{})
	return dao.DB.Delete(&model.Room{}, id).Error
}

// ReorderRooms 房间排序
func ReorderRooms(homeID uint, roomIDs []uint) error {
	for i, rid := range roomIDs {
		dao.DB.Model(&model.Room{}).Where("id = ? AND home_id = ?", rid, homeID).Update("sort_order", i)
	}
	return nil
}

// ========== 房间设备 ==========

// AddDeviceToRoom 添加设备到房间
func AddDeviceToRoom(roomID, deviceID uint) error {
	// 检查房间是否存在
	var r model.Room
	if err := dao.DB.First(&r, roomID).Error; err != nil {
		return errors.New("房间不存在")
	}
	// 检查设备是否存在
	var d model.Device
	if err := dao.DB.First(&d, deviceID).Error; err != nil {
		return errors.New("设备不存在")
	}
	// 检查是否已存在
	var count int64
	dao.DB.Model(&model.RoomDevice{}).Where("room_id = ? AND device_id = ?", roomID, deviceID).Count(&count)
	if count > 0 {
		return errors.New("设备已在该房间中")
	}
	return dao.DB.Create(&model.RoomDevice{RoomID: roomID, DeviceID: deviceID}).Error
}

// RemoveDeviceFromRoom 从房间移除设备
func RemoveDeviceFromRoom(roomID, deviceID uint) error {
	result := dao.DB.Where("room_id = ? AND device_id = ?", roomID, deviceID).Delete(&model.RoomDevice{})
	if result.RowsAffected == 0 {
		return errors.New("设备不在该房间中")
	}
	return result.Error
}

// ListRoomDevices 房间设备列表
func ListRoomDevices(roomID uint) ([]model.Device, error) {
	var deviceIDs []uint
	dao.DB.Model(&model.RoomDevice{}).Where("room_id = ?", roomID).Pluck("device_id", &deviceIDs)
	var list []model.Device
	if len(deviceIDs) == 0 {
		return list, nil
	}
	err := dao.DB.Where("id IN ?", deviceIDs).Find(&list).Error
	return list, err
}

// MoveDeviceToRoom 移动设备到另一个房间
func MoveDeviceToRoom(fromRoomID, toRoomID, deviceID uint) error {
	// 先从原房间移除
	dao.DB.Where("room_id = ? AND device_id = ?", fromRoomID, deviceID).Delete(&model.RoomDevice{})
	// 添加到新房间
	return AddDeviceToRoom(toRoomID, deviceID)
}

// ========== 场景 ==========

// CreateScene 创建场景
func CreateScene(homeID uint, name, icon, sceneType string, sortOrder int) (*model.Scene, error) {
	if name == "" {
		return nil, errors.New("场景名称不可为空")
	}
	if sceneType == "" {
		sceneType = "manual"
	}
	s := &model.Scene{
		HomeID: homeID, Name: name, Icon: icon, Type: sceneType,
		Enabled: true, SortOrder: sortOrder,
	}
	if err := dao.DB.Create(s).Error; err != nil {
		return nil, err
	}
	return s, nil
}

// ListScenes 场景列表
func ListScenes(homeID uint) ([]model.Scene, error) {
	var list []model.Scene
	err := dao.DB.Where("home_id = ?", homeID).Order("sort_order asc, id asc").Find(&list).Error
	return list, err
}

// UpdateScene 更新场景
func UpdateScene(id uint, name, icon, sceneType string) error {
	if name == "" {
		return errors.New("场景名称不可为空")
	}
	updates := map[string]interface{}{"updated_at": time.Now(), "name": name}
	if icon != "" {
		updates["icon"] = icon
	}
	if sceneType != "" {
		updates["type"] = sceneType
	}
	return dao.DB.Model(&model.Scene{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteScene 删除场景
func DeleteScene(id uint) error {
	dao.DB.Where("scene_id = ?", id).Delete(&model.SceneCondition{})
	dao.DB.Where("scene_id = ?", id).Delete(&model.SceneAction{})
	return dao.DB.Delete(&model.Scene{}, id).Error
}

// ToggleScene 启用/禁用场景
func ToggleScene(id uint, enabled bool) error {
	return dao.DB.Model(&model.Scene{}).Where("id = ?", id).Updates(map[string]interface{}{
		"enabled":    enabled,
		"updated_at": time.Now(),
	}).Error
}

// ========== 场景条件/动作 ==========

// AddSceneCondition 添加场景条件
func AddSceneCondition(sceneID uint, condType, configJSON string) (*model.SceneCondition, error) {
	var s model.Scene
	if err := dao.DB.First(&s, sceneID).Error; err != nil {
		return nil, errors.New("场景不存在")
	}
	c := &model.SceneCondition{
		SceneID: sceneID, Type: condType, ConfigJSON: configJSON,
	}
	if err := dao.DB.Create(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

// RemoveSceneCondition 移除场景条件
func RemoveSceneCondition(id uint) error {
	result := dao.DB.Delete(&model.SceneCondition{}, id)
	if result.RowsAffected == 0 {
		return errors.New("条件不存在")
	}
	return result.Error
}

// AddSceneAction 添加场景动作
func AddSceneAction(sceneID, deviceID uint, actionJSON string, sortOrder int) (*model.SceneAction, error) {
	var s model.Scene
	if err := dao.DB.First(&s, sceneID).Error; err != nil {
		return nil, errors.New("场景不存在")
	}
	a := &model.SceneAction{
		SceneID: sceneID, DeviceID: deviceID, ActionJSON: actionJSON, SortOrder: sortOrder,
	}
	if err := dao.DB.Create(a).Error; err != nil {
		return nil, err
	}
	return a, nil
}

// RemoveSceneAction 移除场景动作
func RemoveSceneAction(id uint) error {
	result := dao.DB.Delete(&model.SceneAction{}, id)
	if result.RowsAffected == 0 {
		return errors.New("动作不存在")
	}
	return result.Error
}

// ReorderSceneActions 场景动作排序
func ReorderSceneActions(sceneID uint, actionIDs []uint) error {
	for i, aid := range actionIDs {
		dao.DB.Model(&model.SceneAction{}).Where("id = ? AND scene_id = ?", aid, sceneID).Update("sort_order", i)
	}
	return nil
}

// ========== 执行场景 ==========

// RunScene 手动执行场景 —— 遍历动作并通过 MQTT 下发设备控制指令
func RunScene(sceneID uint, emqx *mqtt.EMQXClient) error {
	var s model.Scene
	if err := dao.DB.First(&s, sceneID).Error; err != nil {
		return errors.New("场景不存在")
	}
	if !s.Enabled {
		return errors.New("场景已禁用")
	}
	var actions []model.SceneAction
	if err := dao.DB.Where("scene_id = ?", sceneID).Order("sort_order asc").Find(&actions).Error; err != nil {
		return err
	}
	if len(actions) == 0 {
		return errors.New("场景无动作")
	}
	for _, a := range actions {
		var d model.Device
		if err := dao.DB.First(&d, a.DeviceID).Error; err != nil {
			continue // 设备不存在则跳过
		}
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(a.ActionJSON), &params); err != nil {
			continue // JSON 解析失败则跳过
		}
		if emqx != nil {
			_ = emqx.Publish(fmt.Sprintf("/sys/cmd/%s", d.DeviceSN), map[string]interface{}{
				"product_key": d.ProductKey,
				"params":      params,
				"ts":          time.Now().Unix(),
			})
		}
	}
	// 记录操作日志
	dao.DB.Create(&model.OperationLog{
		Action: "run_scene", Target: fmt.Sprintf("scene:%d", sceneID),
		Detail: fmt.Sprintf("执行场景 %s，共 %d 个动作", s.Name, len(actions)), CreatedAt: time.Now(),
	})
	return nil
}

// ========== 消息 ==========

// CreateMessage 创建消息
func CreateMessage(userID uint, msgType, title, content, extraJSON string) (*model.Message, error) {
	m := &model.Message{
		UserID: userID, Type: msgType, Title: title, Content: content, ExtraJSON: extraJSON,
	}
	if err := dao.DB.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// ListMessages 消息列表
func ListMessages(userID uint, page, size int) (int64, []model.Message, error) {
	var total int64
	var list []model.Message
	dao.DB.Model(&model.Message{}).Where("user_id = ?", userID).Count(&total)
	err := dao.DB.Where("user_id = ?", userID).Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return total, list, err
}

// MarkMessageRead 标记单条消息已读
func MarkMessageRead(id, userID uint) error {
	return dao.DB.Model(&model.Message{}).Where("id = ? AND user_id = ?", id, userID).Update("read", true).Error
}

// MarkAllRead 标记全部消息已读
func MarkAllRead(userID uint) error {
	return dao.DB.Model(&model.Message{}).Where("user_id = ? AND `read` = ?", userID, false).Update("read", true).Error
}

// UnreadCount 未读消息数
func UnreadCount(userID uint) (int64, error) {
	var count int64
	err := dao.DB.Model(&model.Message{}).Where("user_id = ? AND `read` = ?", userID, false).Count(&count).Error
	return count, err
}

// DeleteMessage 删除消息
func DeleteMessage(id, userID uint) error {
	result := dao.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Message{})
	if result.RowsAffected == 0 {
		return errors.New("消息不存在")
	}
	return result.Error
}

// ========== OTA ==========

// CreateFirmware 创建固件包
func CreateFirmware(productID uint, version, changelog, fileURL string, size int64, createdBy uint) (*model.OTAFirmware, error) {
	if version == "" {
		return nil, errors.New("固件版本不可为空")
	}
	// 检查版本是否已存在
	var count int64
	dao.DB.Model(&model.OTAFirmware{}).Where("product_id = ? AND version = ?", productID, version).Count(&count)
	if count > 0 {
		return nil, errors.New("该版本固件已存在")
	}
	f := &model.OTAFirmware{
		ProductID: productID, Version: version, Changelog: changelog,
		FileURL: fileURL, Size: size, Status: "pending", CreatedBy: createdBy,
	}
	if err := dao.DB.Create(f).Error; err != nil {
		return nil, err
	}
	return f, nil
}

// ListFirmwares 固件列表
func ListFirmwares(productID uint, page, size int) (int64, []model.OTAFirmware, error) {
	var total int64
	var list []model.OTAFirmware
	q := dao.DB.Model(&model.OTAFirmware{})
	if productID > 0 {
		q = q.Where("product_id = ?", productID)
	}
	q.Count(&total)
	err := q.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return total, list, err
}

// PushOTA 推送OTA升级（更新设备固件版本并下发升级指令）
func PushOTA(firmwareID uint, emqx *mqtt.EMQXClient) error {
	var f model.OTAFirmware
	if err := dao.DB.First(&f, firmwareID).Error; err != nil {
		return errors.New("固件包不存在")
	}
	if f.Status == "done" {
		return errors.New("该固件已推送完成")
	}
	// 更新状态为推送中
	dao.DB.Model(&f).Update("status", "pushing")

	// 查找该产品下所有设备
	var devices []model.Device
	dao.DB.Where("product_id = ?", f.ProductID).Find(&devices)

	if len(devices) == 0 {
		dao.DB.Model(&f).Update("status", "done")
		return errors.New("该产品下无设备")
	}

	// 通过 MQTT 下发升级指令
	for _, d := range devices {
		if emqx != nil {
			_ = emqx.Publish(fmt.Sprintf("/ota/upgrade/%s", d.DeviceSN), map[string]interface{}{
				"product_key": d.ProductKey,
				"device_sn":   d.DeviceSN,
				"version":     f.Version,
				"file_url":    f.FileURL,
				"ts":          time.Now().Unix(),
			})
		}
	}

	// 记录日志
	dao.DB.Create(&model.OperationLog{
		Action: "push_ota", Target: fmt.Sprintf("firmware:%d", firmwareID),
		Detail: fmt.Sprintf("推送OTA固件 %s，共 %d 台设备", f.Version, len(devices)), CreatedAt: time.Now(),
	})

	dao.DB.Model(&f).Update("status", "done")
	return nil
}

// Ensure 利用 gorm.ErrRecordNotFound 的辅助判断（如需更精确的 not-found 语义）
func ensureRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// ========== 操作日志 ==========

// ListOperationLogs 操作日志列表
func ListOperationLogs(page, size int) (int64, []model.OperationLog, error) {
	var total int64
	var list []model.OperationLog
	dao.DB.Model(&model.OperationLog{}).Count(&total)
	err := dao.DB.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return total, list, err
}

// ========== 固件删除 ==========

// DeleteFirmware 删除固件包
func DeleteFirmware(id uint) error {
	return dao.DB.Delete(&model.OTAFirmware{}, id).Error
}

// ========== 设备历史数据 ==========

// SaveDeviceData 保存设备上报数据到历史记录
func SaveDeviceData(deviceSN, property, value string) {
	h := model.DeviceDataHistory{
		DeviceSN:  deviceSN,
		Property:  property,
		Value:     value,
		CreatedAt: time.Now(),
	}
	dao.DB.Create(&h)
}

// GetDeviceDataHistory 查询设备属性历史数据
func GetDeviceDataHistory(deviceSN, property string, startTime, endTime time.Time, limit int) ([]model.DeviceDataHistory, error) {
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	var list []model.DeviceDataHistory
	q := dao.DB.Where("device_sn = ?", deviceSN)
	if property != "" {
		q = q.Where("property = ?", property)
	}
	if !startTime.IsZero() {
		q = q.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		q = q.Where("created_at <= ?", endTime)
	}
	err := q.Order("created_at desc").Limit(limit).Find(&list).Error
	return list, err
}

// ========== 事件 ==========

// ReportDeviceEvent 设备上报事件
func ReportDeviceEvent(deviceSN, eventID, eventName, outputJSON string) error {
	history := model.DeviceEventHistory{
		DeviceSN:   deviceSN,
		EventID:    eventID,
		EventName:  eventName,
		OutputJSON: outputJSON,
		CreatedAt:  time.Now(),
	}
	if err := dao.DB.Create(&history).Error; err != nil {
		return err
	}
	dao.DB.Create(&model.Message{
		UserID:    0,
		Type:      "device",
		Title:     fmt.Sprintf("设备事件: %s", eventName),
		Content:   fmt.Sprintf("设备 %s 上报事件 %s", deviceSN, eventName),
		ExtraJSON: fmt.Sprintf(`{"device_sn":"%s","event_id":"%s"}`, deviceSN, eventID),
		CreatedAt: time.Now(),
	})
	return nil
}

// ListDeviceEvents 设备事件历史
func ListDeviceEvents(deviceSN, eventID string, page, size int) (int64, []model.DeviceEventHistory, error) {
	var total int64
	var list []model.DeviceEventHistory
	q := dao.DB.Where("device_sn = ?", deviceSN)
	if eventID != "" {
		q = q.Where("event_id = ?", eventID)
	}
	q.Model(&model.DeviceEventHistory{}).Count(&total)
	err := q.Order("id desc").Offset((page-1)*size).Limit(size).Find(&list).Error
	return total, list, err
}

// ========== 服务 ==========

// InvokeDeviceService 云端调用设备服务
func InvokeDeviceService(deviceSN, serviceID, serviceName, inputJSON string) (*model.DeviceServiceHistory, error) {
	history := model.DeviceServiceHistory{
		DeviceSN:    deviceSN,
		ServiceID:   serviceID,
		ServiceName: serviceName,
		InputJSON:   inputJSON,
		Status:      "success",
		CreatedAt:   time.Now(),
	}
	if err := dao.DB.Create(&history).Error; err != nil {
		return nil, err
	}
	return &history, nil
}

// ListDeviceServiceHistory 设备服务调用历史
func ListDeviceServiceHistory(deviceSN, serviceID string, page, size int) (int64, []model.DeviceServiceHistory, error) {
	var total int64
	var list []model.DeviceServiceHistory
	q := dao.DB.Where("device_sn = ?", deviceSN)
	if serviceID != "" {
		q = q.Where("service_id = ?", serviceID)
	}
	q.Model(&model.DeviceServiceHistory{}).Count(&total)
	err := q.Order("id desc").Offset((page-1)*size).Limit(size).Find(&list).Error
	return total, list, err
}

// ========== 设备影子 ==========

// GetDeviceShadow 获取设备影子
func GetDeviceShadow(deviceSN string) (*model.DeviceShadow, error) {
	var shadow model.DeviceShadow
	err := dao.DB.Where("device_sn = ?", deviceSN).First(&shadow).Error
	if err == gorm.ErrRecordNotFound {
		shadow = model.DeviceShadow{
			DeviceSN:     deviceSN,
			DesiredJSON:  "{}",
			ReportedJSON: "{}",
		}
		return &shadow, nil
	}
	return &shadow, err
}

// UpdateDeviceShadowDesired 更新设备影子期望值
func UpdateDeviceShadowDesired(deviceSN string, desired map[string]interface{}) error {
	desiredJSON, _ := json.Marshal(desired)
	var shadow model.DeviceShadow
	err := dao.DB.Where("device_sn = ?", deviceSN).First(&shadow).Error
	if err == gorm.ErrRecordNotFound {
		shadow = model.DeviceShadow{
			DeviceSN:     deviceSN,
			DesiredJSON:  string(desiredJSON),
			ReportedJSON: "{}",
			Version:      1,
		}
		return dao.DB.Create(&shadow).Error
	}
	shadow.DesiredJSON = string(desiredJSON)
	shadow.Version++
	return dao.DB.Save(&shadow).Error
}

// SyncDeviceShadowReported 同步设备上报值到影子
func SyncDeviceShadowReported(deviceSN string, reported map[string]interface{}) {
	reportedJSON, _ := json.Marshal(reported)
	var shadow model.DeviceShadow
	if err := dao.DB.Where("device_sn = ?", deviceSN).First(&shadow).Error; err != nil {
		shadow = model.DeviceShadow{
			DeviceSN:     deviceSN,
			DesiredJSON:  "{}",
			ReportedJSON: string(reportedJSON),
		}
		dao.DB.Create(&shadow)
		return
	}
	shadow.ReportedJSON = string(reportedJSON)
	shadow.Version++
	dao.DB.Save(&shadow)
}

// ========== 物模型导入导出 ==========

// ExportThingModel 导出飞燕标准物模型 JSON
func ExportThingModel(productID uint) (map[string]interface{}, error) {
	var p model.Product
	if err := dao.DB.First(&p, productID).Error; err != nil {
		return nil, errors.New("产品不存在")
	}
	var properties, events, services interface{}
	json.Unmarshal([]byte(p.PropertiesJSON), &properties)
	json.Unmarshal([]byte(p.EventsJSON), &events)
	json.Unmarshal([]byte(p.ServicesJSON), &services)

	return map[string]interface{}{
		"product_key": p.ProductKey,
		"name":        p.Name,
		"description": p.Description,
		"category":    p.Category,
		"properties":  properties,
		"events":      events,
		"services":    services,
	}, nil
}

// ImportThingModel 导入飞燕标准物模型 JSON
func ImportThingModel(productID uint, data map[string]interface{}) error {
	var p model.Product
	if err := dao.DB.First(&p, productID).Error; err != nil {
		return errors.New("产品不存在")
	}
	var devCount int64
	dao.DB.Model(&model.Device{}).Where("product_id = ?", productID).Count(&devCount)
	if devCount > 0 {
		return errors.New("产品下已有设备，物模型结构锁定，不可修改")
	}

	if props, ok := data["properties"]; ok {
		b, _ := json.Marshal(props)
		p.PropertiesJSON = string(b)
	}
	if events, ok := data["events"]; ok {
		b, _ := json.Marshal(events)
		p.EventsJSON = string(b)
	}
	if services, ok := data["services"]; ok {
		b, _ := json.Marshal(services)
		p.ServicesJSON = string(b)
	}
	if name, ok := data["name"].(string); ok && name != "" {
		p.Name = name
	}
	if desc, ok := data["description"].(string); ok {
		p.Description = desc
	}
	if cat, ok := data["category"].(string); ok {
		p.Category = cat
	}

	return dao.DB.Save(&p).Error
}

// ValidatePropertyReport 校验设备上报属性是否符合物模型
func ValidatePropertyReport(productID uint, params map[string]interface{}) []string {
	var p model.Product
	if err := dao.DB.First(&p, productID).Error; err != nil {
		return nil
	}
	var specs []map[string]interface{}
	if err := json.Unmarshal([]byte(p.PropertiesJSON), &specs); err != nil {
		return nil
	}
	var warnings []string
	for _, spec := range specs {
		identifier, _ := spec["identifier"].(string)
		if _, exists := params[identifier]; !exists {
			continue
		}
		dt, _ := spec["dataType"].(map[string]interface{})
		dataType, _ := dt["type"].(string)
		specsMap, _ := dt["specs"].(map[string]interface{})

		val := params[identifier]
		switch dataType {
		case "int", "float":
			numVal, ok := val.(float64)
			if !ok {
				warnings = append(warnings, fmt.Sprintf("属性 %s 期望数值类型", identifier))
				continue
			}
			if min, ok := specsMap["min"].(float64); ok && numVal < min {
				warnings = append(warnings, fmt.Sprintf("属性 %s 值 %v 小于最小值 %v", identifier, numVal, min))
			}
			if max, ok := specsMap["max"].(float64); ok && numVal > max {
				warnings = append(warnings, fmt.Sprintf("属性 %s 值 %v 大于最大值 %v", identifier, numVal, max))
			}
		case "bool":
			if _, ok := val.(bool); !ok {
				warnings = append(warnings, fmt.Sprintf("属性 %s 期望布尔类型", identifier))
			}
		case "enum":
			enumVals, _ := specsMap["values"].([]interface{})
			found := false
			for _, ev := range enumVals {
				if fmt.Sprintf("%v", ev) == fmt.Sprintf("%v", val) {
					found = true
					break
				}
			}
			if !found {
				warnings = append(warnings, fmt.Sprintf("属性 %s 值 %v 不在枚举范围内", identifier, val))
			}
		}
	}
	return warnings
}

// ========== 管理员设备管理 ==========

// UpdateAdminDevice 管理员更新设备
func UpdateAdminDevice(id uint, name string, status int) error {
	updates := map[string]interface{}{"updated_at": time.Now()}
	if name != "" {
		updates["device_name"] = name
	}
	if status >= 0 {
		updates["status"] = status
	}
	return dao.DB.Model(&model.Device{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteAdminDevice 管理员删除设备
func DeleteAdminDevice(id uint) error {
	// 先删关联
	dao.DB.Where("device_id = ?", id).Delete(&model.RoomDevice{})
	dao.DB.Where("device_id = ?", id).Delete(&model.DeviceShare{})
	dao.DB.Where("device_id = ?", id).Delete(&model.UserDeviceUI{})
	return dao.DB.Delete(&model.Device{}, id).Error
}

// ========== 用户资料编辑 ==========

// UpdateUserInfo 更新用户资料
func UpdateUserInfo(uid uint, nickname, avatar, email string) error {
	updates := map[string]interface{}{"updated_at": time.Now()}
	if nickname != "" {
		updates["nickname"] = nickname
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}
	if email != "" {
		updates["email"] = email
	}
	return dao.DB.Model(&model.User{}).Where("id = ?", uid).Updates(updates).Error
}

// ChangePassword 修改密码
func ChangePassword(uid uint, oldPwd, newPwd string) error {
	var u model.User
	if err := dao.DB.First(&u, uid).Error; err != nil {
		return errors.New("用户不存在")
	}
	if !util.CheckPassword(oldPwd, u.Password) {
		return errors.New("旧密码错误")
	}
	hashed, err := util.HashPassword(newPwd)
	if err != nil {
		return err
	}
	return dao.DB.Model(&u).Update("password", hashed).Error
}

// AdminToggleUserStatus 管理员启用/禁用用户
func AdminToggleUserStatus(uid uint, status int) error {
	return dao.DB.Model(&model.User{}).Where("id = ?", uid).Update("status", status).Error
}

// AdminDeleteUser 管理员删除用户
func AdminDeleteUser(uid uint) error {
	var u model.User
	if err := dao.DB.First(&u, uid).Error; err != nil {
		return errors.New("用户不存在")
	}
	if u.Role == "admin" {
		return errors.New("不可删除管理员账户")
	}
	return dao.DB.Delete(&u).Error
}

// AdminResetPassword 管理员重置用户密码
func AdminResetPassword(uid uint, newPwd string) error {
	hashed, err := util.HashPassword(newPwd)
	if err != nil {
		return err
	}
	return dao.DB.Model(&model.User{}).Where("id = ?", uid).Update("password", hashed).Error
}

// ========== 告警规则 ==========

// CreateAlertRule 创建告警规则
func CreateAlertRule(productID uint, deviceSN, property, operator, threshold, notifyType, notifyURL string, createdBy uint) (*model.AlertRule, error) {
	if property == "" || operator == "" {
		return nil, errors.New("属性和操作符不可为空")
	}
	r := &model.AlertRule{
		ProductID:  productID,
		DeviceSN:   deviceSN,
		Property:   property,
		Operator:   operator,
		Threshold:  threshold,
		Enabled:    true,
		NotifyType: notifyType,
		NotifyURL:  notifyURL,
		CreatedBy:  createdBy,
	}
	if err := dao.DB.Create(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

// ListAlertRules 告警规则列表
func ListAlertRules(productID uint, page, size int) (int64, []model.AlertRule, error) {
	var total int64
	var list []model.AlertRule
	q := dao.DB.Model(&model.AlertRule{})
	if productID > 0 {
		q = q.Where("product_id = ?", productID)
	}
	q.Count(&total)
	err := q.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return total, list, err
}

// DeleteAlertRule 删除告警规则
func DeleteAlertRule(id uint) error {
	return dao.DB.Delete(&model.AlertRule{}, id).Error
}

// ToggleAlertRule 启用/禁用告警规则
func ToggleAlertRule(id uint, enabled bool) error {
	return dao.DB.Model(&model.AlertRule{}).Where("id = ?", id).Update("enabled", enabled).Error
}

// ========== 规则引擎/场景联动 ==========

// RuleTrigger 规则触发条件
type RuleTrigger struct {
	Type       string `json:"type"`       // device_property / device_event / timer / manual
	DeviceSN   string `json:"device_sn"`  // 设备序列号（设备级触发
	ProductID  uint   `json:"product_id"` // 产品ID（产品级触发
	Property   string `json:"property"`   // 属性标识符
	Operator   string `json:"operator"`   // >, <, ==, >=, <=, !=, changed
	Threshold  string `json:"threshold"`  // 阈值
	EventID    string `json:"event_id"`   // 事件标识符
	CronExpr   string `json:"cron_expr"`  // 定时表达式
	Expression string `json:"expression"` // 复杂表达式
}

// RuleAction 规则执行动作
type RuleAction struct {
	Type        string                 `json:"type"`         // device_control / scene_run / notify / webhook / delay
	DeviceSN    string                 `json:"device_sn"`    // 目标设备
	Property    string                 `json:"property"`     // 设置的属性
	Value       interface{}            `json:"value"`        // 设置的值
	ServiceID   string                 `json:"service_id"`   // 调用的服务
	ServiceName string                 `json:"service_name"` // 服务名称
	Input       map[string]interface{} `json:"input"`        // 服务输入参数
	SceneID     uint                   `json:"scene_id"`     // 执行的场景
	Title       string                 `json:"title"`        // 通知标题
	Content     string                 `json:"content"`      // 通知内容
	NotifyType  string                 `json:"notify_type"`  // message / push / sms
	WebhookURL  string                 `json:"webhook_url"`  // webhook地址
	DelaySec    int                    `json:"delay_sec"`    // 延迟秒数
}

// CreateRule 创建规则
func CreateRule(name, description, ruleType string, trigger RuleTrigger, actions []RuleAction, createdBy uint) (*model.Rule, error) {
	triggerJSON, _ := json.Marshal(trigger)
	actionJSON, _ := json.Marshal(actions)
	r := &model.Rule{
		Name:        name,
		Description: description,
		Type:        ruleType,
		Enabled:     true,
		TriggerJSON: string(triggerJSON),
		ActionJSON:  string(actionJSON),
		CreatedBy:   createdBy,
	}
	if err := dao.DB.Create(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

// ListRules 规则列表
func ListRules(page, size int) (int64, []model.Rule, error) {
	var total int64
	var list []model.Rule
	q := dao.DB.Model(&model.Rule{})
	q.Count(&total)
	err := q.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return total, list, err
}

// GetRule 获取规则详情
func GetRule(id uint) (*model.Rule, error) {
	var r model.Rule
	if err := dao.DB.First(&r, id).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

// UpdateRule 更新规则
func UpdateRule(id uint, name, description string, trigger RuleTrigger, actions []RuleAction, enabled bool) error {
	triggerJSON, _ := json.Marshal(trigger)
	actionJSON, _ := json.Marshal(actions)
	return dao.DB.Model(&model.Rule{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":         name,
		"description":  description,
		"trigger_json": string(triggerJSON),
		"action_json":  string(actionJSON),
		"enabled":      enabled,
	}).Error
}

// DeleteRule 删除规则
func DeleteRule(id uint) error {
	return dao.DB.Delete(&model.Rule{}, id).Error
}

// ToggleRule 启用/禁用规则
func ToggleRule(id uint, enabled bool) error {
	return dao.DB.Model(&model.Rule{}).Where("id = ?", id).Update("enabled", enabled).Error
}

// EvaluateRule 评估并执行规则
func EvaluateRule(triggerData map[string]interface{}) {
	var rules []model.Rule
	dao.DB.Where("enabled = ?", true).Find(&rules)

	for _, rule := range rules {
		if matchTrigger(rule, triggerData) {
			executeRule(rule, triggerData)
		}
	}
}

func matchTrigger(rule model.Rule, data map[string]interface{}) bool {
	var trigger RuleTrigger
	json.Unmarshal([]byte(rule.TriggerJSON), &trigger)

	switch trigger.Type {
	case "device_property":
		// 属性触发：检查设备上报的属性值
		if data["type"] != "property_report" {
			return false
		}
		if trigger.DeviceSN != "" && data["device_sn"] != trigger.DeviceSN {
			return false
		}
		if trigger.ProductID > 0 && data["product_id"] != trigger.ProductID {
			return false
		}
		if trigger.Property != "" {
			if props, ok := data["properties"].(map[string]interface{}); ok {
				val := props[trigger.Property]
				return compareValue(val, trigger.Operator, trigger.Threshold)
			}
		}
		return true

	case "device_event":
		// 事件触发
		if data["type"] != "event_report" {
			return false
		}
		if trigger.DeviceSN != "" && data["device_sn"] != trigger.DeviceSN {
			return false
		}
		if trigger.EventID != "" && data["event_id"] != trigger.EventID {
			return false
		}
		return true

	case "timer":
		// 定时触发由外部调度器处理
		return false

	case "manual":
		// 手动触发
		return data["manual_trigger"] == true && data["rule_id"] == rule.ID

	default:
		return false
	}
}

func compareValue(val interface{}, operator, threshold string) bool {
	if val == nil {
		return false
	}
	// 尝试转为数值比较
	var fVal float64
	switch v := val.(type) {
	case float64:
		fVal = v
	case int:
		fVal = float64(v)
	case string:
		fVal, _ = strconv.ParseFloat(v, 64)
	default:
		return false
	}
	fThresh, _ := strconv.ParseFloat(threshold, 64)

	switch operator {
	case ">":
		return fVal > fThresh
	case "<":
		return fVal < fThresh
	case ">=":
		return fVal >= fThresh
	case "<=":
		return fVal <= fThresh
	case "==":
		return fVal == fThresh
	case "!=":
		return fVal != fThresh
	case "changed":
		return true // 只要上报就触发
	default:
		return false
	}
}

func executeRule(rule model.Rule, triggerData map[string]interface{}) {
	var actions []RuleAction
	json.Unmarshal([]byte(rule.ActionJSON), &actions)

	execution := &model.RuleExecution{
		RuleID:      rule.ID,
		TriggeredAt: time.Now(),
		Success:     true,
		DetailJSON:  "",
	}
	detail := map[string]interface{}{"actions": []map[string]interface{}{}}

	for _, action := range actions {
		actionDetail := map[string]interface{}{"type": action.Type, "success": true}
		var err error

		switch action.Type {
		case "device_control":
			// 设备属性控制
			if action.DeviceSN != "" && action.Property != "" {
				// 通过 SN 查找设备
				var d model.Device
				if err = dao.DB.Where("device_sn = ?", action.DeviceSN).First(&d).Error; err == nil {
					params := map[string]interface{}{action.Property: action.Value}
					err = ControlDevice(0, "admin", d.ID, params, nil)
				} else {
					err = errors.New("设备不存在: " + action.DeviceSN)
				}
			}

		case "device_service":
			// 调用设备服务
			if action.DeviceSN != "" && action.ServiceID != "" {
				inputJSON := "{}"
				if action.Input != nil {
					ib, _ := json.Marshal(action.Input)
					inputJSON = string(ib)
				}
				_, err = InvokeDeviceService(action.DeviceSN, action.ServiceID, action.ServiceName, inputJSON)
			}

		case "scene_run":
			// 执行场景
			if action.SceneID > 0 {
				err = RunScene(action.SceneID, nil)
			}

		case "notify":
			// 发送站内消息
			err = sendRuleNotification(action.Title, action.Content, action.NotifyType)

		case "webhook":
			// 调用 Webhook
			err = callWebhook(action.WebhookURL, triggerData, rule)

		case "delay":
			// 延迟（仅记录，实际延迟由前端或调度器处理）
			actionDetail["delay_sec"] = action.DelaySec
		}

		if err != nil {
			execution.Success = false
			execution.ErrorMsg = err.Error()
			actionDetail["success"] = false
			actionDetail["error"] = err.Error()
		}

		detail["actions"] = append(detail["actions"].([]map[string]interface{}), actionDetail)
	}

	detailJSON, _ := json.Marshal(detail)
	execution.DetailJSON = string(detailJSON)

	// 更新规则统计
	dao.DB.Model(&rule).Updates(map[string]interface{}{
		"last_run_at": time.Now(),
		"run_count":   gorm.Expr("run_count + 1"),
	})

	dao.DB.Create(execution)
}

// 辅助函数
func sendRuleNotification(title, content, notifyType string) error {
	// 这里可以集成推送服务，暂存为系统消息给管理员
	var admin model.User
	dao.DB.Where("role = ?", "admin").First(&admin)
	if admin.ID > 0 {
		msg := &model.Message{
			UserID: admin.ID,
			Type:   "rule",
			Title:  title,
			Content: content,
		}
		dao.DB.Create(msg)
	}
	return nil
}

func callWebhook(url string, triggerData map[string]interface{}, rule model.Rule) error {
	payload := map[string]interface{}{
		"rule_id":    rule.ID,
		"rule_name":  rule.Name,
		"triggered_at": time.Now().Format(time.RFC3339),
		"trigger_data": triggerData,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return errors.New("webhook returned status: " + strconv.Itoa(resp.StatusCode))
	}
	return nil
}

// 规则执行记录
func ListRuleExecutions(ruleID uint, page, size int) (int64, []model.RuleExecution, error) {
	var total int64
	var list []model.RuleExecution
	q := dao.DB.Model(&model.RuleExecution{})
	if ruleID > 0 {
		q = q.Where("rule_id = ?", ruleID)
	}
	q.Count(&total)
	err := q.Order("triggered_at desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return total, list, err
}

// ========== 设备分组 ==========

// CreateDeviceGroup 创建设备分组
func CreateDeviceGroup(name, description string, productID, ownerID uint) (*model.DeviceGroup, error) {
	g := &model.DeviceGroup{
		Name:        name,
		Description: description,
		ProductID:   productID,
		OwnerID:     ownerID,
	}
	if err := dao.DB.Create(g).Error; err != nil {
		return nil, err
	}
	return g, nil
}

// ListDeviceGroups 分组列表
func ListDeviceGroups(productID, ownerID uint, page, size int) (int64, []model.DeviceGroup, error) {
	var total int64
	var list []model.DeviceGroup
	q := dao.DB.Model(&model.DeviceGroup{})
	if productID > 0 {
		q = q.Where("product_id = ?", productID)
	}
	if ownerID > 0 {
		q = q.Where("owner_id = ?", ownerID)
	}
	q.Count(&total)
	err := q.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return total, list, err
}

// AddDeviceToGroup 添加设备到分组
func AddDeviceToGroup(groupID, deviceID uint) error {
	return dao.DB.Create(&model.DeviceGroupMember{GroupID: groupID, DeviceID: deviceID}).Error
}

// RemoveDeviceFromGroup 从分组移除设备
func RemoveDeviceFromGroup(groupID, deviceID uint) error {
	return dao.DB.Delete(&model.DeviceGroupMember{}, "group_id = ? AND device_id = ?", groupID, deviceID).Error
}

// ListGroupDevices 获取分组内设备
func ListGroupDevices(groupID uint) ([]model.Device, error) {
	var devices []model.Device
	err := dao.DB.Table("devices").
		Joins("JOIN device_group_members ON devices.id = device_group_members.device_id").
		Where("device_group_members.group_id = ?", groupID).
		Find(&devices).Error
	return devices, err
}

// ========== 设备标签 ==========

// SetDeviceTag 设置设备标签
func SetDeviceTag(deviceID uint, key, value string) error {
	tag := model.DeviceTag{DeviceID: deviceID, Key: key, Value: value}
	return dao.DB.Where("device_id = ? AND key = ?", deviceID, key).
		Assign(tag).FirstOrCreate(&tag).Error
}

// GetDeviceTags 获取设备标签
func GetDeviceTags(deviceID uint) ([]model.DeviceTag, error) {
	var tags []model.DeviceTag
	err := dao.DB.Where("device_id = ?", deviceID).Find(&tags).Error
	return tags, err
}

// DeleteDeviceTag 删除设备标签
func DeleteDeviceTag(deviceID uint, key string) error {
	return dao.DB.Delete(&model.DeviceTag{}, "device_id = ? AND key = ?", deviceID, key).Error
}

// SearchDevicesByTag 标签搜索设备
func SearchDevicesByTag(key, value string) ([]model.Device, error) {
	var devices []model.Device
	err := dao.DB.Table("devices").
		Joins("JOIN device_tags ON devices.id = device_tags.device_id").
		Where("device_tags.key = ? AND device_tags.value = ?", key, value).
		Find(&devices).Error
	return devices, err
}

// LogMqttMessage 记录 MQTT 报文
func LogMqttMessage(deviceSN, topic, direction, payload string) {
	msg := &model.MqttMessage{
		DeviceSN:  deviceSN,
		Topic:     topic,
		Direction: direction,
		Payload:   payload,
	}
	dao.DB.Create(msg)
}

// ========== 网关子设备管理 ==========

// AddSubDevice 添加子设备到网关
func AddSubDevice(gatewayID, subDeviceID uint) error {
	var gateway model.Device
	if err := dao.DB.First(&gateway, gatewayID).Error; err != nil {
		return errors.New("网关设备不存在")
	}
	if gateway.NodeType != "gateway" {
		return errors.New("该设备不是网关类型")
	}

	var subDevice model.Device
	if err := dao.DB.First(&subDevice, subDeviceID).Error; err != nil {
		return errors.New("子设备不存在")
	}

	// 更新子设备网关关联
	subDevice.GatewayID = gatewayID
	subDevice.NodeType = "sub_device"
	if err := dao.DB.Save(&subDevice).Error; err != nil {
		return err
	}

	// 创建拓扑关系
	topo := &model.DeviceTopology{
		GatewayID:  gatewayID,
		SubDeviceID: subDeviceID,
		Status:     "online",
	}
	return dao.DB.Create(topo).Error
}

// RemoveSubDevice 从网关移除子设备
func RemoveSubDevice(gatewayID, subDeviceID uint) error {
	// 删除拓扑关系
	if err := dao.DB.Where("gateway_id = ? AND sub_device_id = ?", gatewayID, subDeviceID).Delete(&model.DeviceTopology{}).Error; err != nil {
		return err
	}

	// 更新子设备
	var subDevice model.Device
	if err := dao.DB.First(&subDevice, subDeviceID).Error; err != nil {
		return errors.New("子设备不存在")
	}
	subDevice.GatewayID = 0
	subDevice.NodeType = "device"
	return dao.DB.Save(&subDevice).Error
}

// ListGatewaySubDevices 获取网关下的子设备列表
func ListGatewaySubDevices(gatewayID uint) ([]model.Device, error) {
	var devices []model.Device
	err := dao.DB.Table("devices").
		Joins("JOIN device_topologies ON devices.id = device_topologies.sub_device_id").
		Where("device_topologies.gateway_id = ?", gatewayID).
		Find(&devices).Error
	return devices, err
}

// GetDeviceTopology 获取设备拓扑信息
func GetDeviceTopology(deviceID uint) (*model.DeviceTopology, error) {
	var topo model.DeviceTopology
	err := dao.DB.Where("gateway_id = ? OR sub_device_id = ?", deviceID, deviceID).First(&topo).Error
	if err != nil {
		return nil, err
	}
	return &topo, nil
}

// UpdateSubDeviceStatus 更新子设备在线状态
func UpdateSubDeviceStatus(subDeviceID uint, status string) error {
	return dao.DB.Model(&model.DeviceTopology{}).
		Where("sub_device_id = ?", subDeviceID).
		Update("status", status).Error
}

// BatchAddSubDevices 批量添加子设备
func BatchAddSubDevices(gatewayID uint, subDeviceIDs []uint) error {
	return dao.DB.Transaction(func(tx *gorm.DB) error {
		var gateway model.Device
		if err := tx.First(&gateway, gatewayID).Error; err != nil {
			return errors.New("网关设备不存在")
		}
		if gateway.NodeType != "gateway" {
			return errors.New("该设备不是网关类型")
		}

		for _, subID := range subDeviceIDs {
			var subDevice model.Device
			if err := tx.First(&subDevice, subID).Error; err != nil {
				return errors.New("子设备不存在: " + strconv.Itoa(int(subID)))
			}
			subDevice.GatewayID = gatewayID
			subDevice.NodeType = "sub_device"
			if err := tx.Save(&subDevice).Error; err != nil {
				return err
			}

			topo := &model.DeviceTopology{
				GatewayID:   gatewayID,
				SubDeviceID: subID,
				Status:      "online",
			}
			if err := tx.Create(topo).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ========== 设备配网 ==========

// CreateProvisioningConfig 创建配网配置
func CreateProvisioningConfig(productID uint, method, bleConfig, softapConfig, qrcodeConfig string, createdBy uint) (*model.ProvisioningConfig, error) {
	cfg := &model.ProvisioningConfig{
		ProductID:    productID,
		Method:       method,
		BLEConfig:    bleConfig,
		SoftAPConfig: softapConfig,
		QRCodeConfig: qrcodeConfig,
		CreatedBy:    createdBy,
	}
	if err := dao.DB.Create(cfg).Error; err != nil {
		return nil, err
	}
	return cfg, nil
}

// GetProvisioningConfig 获取产品的配网配置
func GetProvisioningConfig(productID uint) (*model.ProvisioningConfig, error) {
	var cfg model.ProvisioningConfig
	err := dao.DB.Where("product_id = ?", productID).First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// UpdateProvisioningConfig 更新配网配置
func UpdateProvisioningConfig(productID uint, method, bleConfig, softapConfig, qrcodeConfig string) error {
	return dao.DB.Model(&model.ProvisioningConfig{}).
		Where("product_id = ?", productID).
		Updates(map[string]interface{}{
			"method":        method,
			"ble_config":    bleConfig,
			"softap_config": softapConfig,
			"qrcode_config": qrcodeConfig,
		}).Error
}

// StartProvisioning 开始配网
func StartProvisioning(productID uint, deviceSN, deviceName, method, ssid, bssid, pinCode string, createdBy uint) (*model.DeviceProvisioning, error) {
	prov := &model.DeviceProvisioning{
		ProductID:  productID,
		DeviceSN:   deviceSN,
		DeviceName: deviceName,
		Method:     method,
		Status:     "pending",
		SSID:       ssid,
		BSSID:      bssid,
		PinCode:    pinCode,
		CreatedBy:  createdBy,
	}

	// 根据方式生成配网凭证
	switch method {
	case "ble":
		// 蓝牙配网：生成随机 PIN 码
		if prov.PinCode == "" {
			prov.PinCode = fmt.Sprintf("%06d", rand.Intn(1000000))
		}
	case "softap":
		// SoftAP 配网：需要 SSID
		if ssid == "" {
			return nil, errors.New("SoftAP 配网需要提供 SSID")
		}
	case "qrcode":
		// 扫码配网：生成二维码内容
		prov.QRCode = fmt.Sprintf("https://iot.aliyun.com/provision?sn=%s&pk=%s", deviceSN, "")
	case "zero":
		// 零配：无需额外参数
	}

	if err := dao.DB.Create(prov).Error; err != nil {
		return nil, err
	}
	return prov, nil
}

// GetProvisioningStatus 获取配网状态
func GetProvisioningStatus(deviceSN string) (*model.DeviceProvisioning, error) {
	var prov model.DeviceProvisioning
	err := dao.DB.Where("device_sn = ?", deviceSN).Order("id desc").First(&prov).Error
	if err != nil {
		return nil, err
	}
	return &prov, nil
}

// CompleteProvisioning 完成配网（设备上线回调）
func CompleteProvisioning(deviceSN string, success bool, errorMsg string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       "success",
		"completed_at": &now,
	}
	if !success {
		updates["status"] = "failed"
		updates["error_msg"] = errorMsg
	}
	return dao.DB.Model(&model.DeviceProvisioning{}).
		Where("device_sn = ? AND status = 'pending'", deviceSN).
		Updates(updates).Error
}

// ListProvisioningRecords 配网记录列表
func ListProvisioningRecords(productID uint, page, size int) (int64, []model.DeviceProvisioning, error) {
	var total int64
	var list []model.DeviceProvisioning
	q := dao.DB.Model(&model.DeviceProvisioning{})
	if productID > 0 {
		q = q.Where("product_id = ?", productID)
	}
	q.Count(&total)
	err := q.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return total, list, err
}

// GenerateDeviceQRCode 生成设备配网二维码
func GenerateDeviceQRCode(deviceSN, productKey string) (string, error) {
	// 生成配网二维码链接（飞燕标准）
	qrContent := fmt.Sprintf("aliyun_iot://provision?sn=%s&pk=%s", deviceSN, productKey)
	return qrContent, nil
}

// ========== 数据流转/SQL分析 ==========

// CreateDataFlow 创建数据流转规则
func CreateDataFlow(name, description, flowType, sourceType, sourceTopic, sql, targetType, targetTopic, targetConfig string, createdBy uint) (*model.DataFlow, error) {
	df := &model.DataFlow{
		Name:         name,
		Description:  description,
		Type:         flowType,
		SourceType:   sourceType,
		SourceTopic:  sourceTopic,
		SQL:          sql,
		TargetType:   targetType,
		TargetTopic:  targetTopic,
		TargetConfig: targetConfig,
		Enabled:      true,
		CreatedBy:    createdBy,
	}
	if err := dao.DB.Create(df).Error; err != nil {
		return nil, err
	}
	return df, nil
}

// ListDataFlows 数据流转规则列表
func ListDataFlows(page, size int) (int64, []model.DataFlow, error) {
	var total int64
	var list []model.DataFlow
	q := dao.DB.Model(&model.DataFlow{})
	q.Count(&total)
	err := q.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return total, list, err
}

// GetDataFlow 获取数据流转规则详情
func GetDataFlow(id uint) (*model.DataFlow, error) {
	var df model.DataFlow
	if err := dao.DB.First(&df, id).Error; err != nil {
		return nil, err
	}
	return &df, nil
}

// UpdateDataFlow 更新数据流转规则
func UpdateDataFlow(id uint, name, description, flowType, sourceType, sourceTopic, sql, targetType, targetTopic, targetConfig string, enabled bool) error {
	return dao.DB.Model(&model.DataFlow{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":          name,
		"description":   description,
		"type":          flowType,
		"source_type":   sourceType,
		"source_topic":  sourceTopic,
		"sql":           sql,
		"target_type":   targetType,
		"target_topic":  targetTopic,
		"target_config": targetConfig,
		"enabled":       enabled,
	}).Error
}

// DeleteDataFlow 删除数据流转规则
func DeleteDataFlow(id uint) error {
	return dao.DB.Delete(&model.DataFlow{}, id).Error
}

// ToggleDataFlow 启用/禁用数据流转规则
func ToggleDataFlow(id uint, enabled bool) error {
	return dao.DB.Model(&model.DataFlow{}).Where("id = ?", id).Update("enabled", enabled).Error
}

// ExecuteDataFlow 执行数据流转（简化版，实际应由规则引擎调用）
func ExecuteDataFlow(flowID uint, input map[string]interface{}) error {
	df, err := GetDataFlow(flowID)
	if err != nil {
		return err
	}
	if !df.Enabled {
		return nil
	}

	execution := &model.DataFlowExecution{
		FlowID:    flowID,
		Status:    "success",
		InputJSON: "",
	}
	inputJSON, _ := json.Marshal(input)
	execution.InputJSON = string(inputJSON)

	var output interface{}
	var execErr error

	// 简化的 SQL 执行（实际应使用更完整的 SQL 引擎）
	switch df.Type {
	case "sql":
		// 这里只是记录，实际 SQL 执行需要专门引擎
		output = map[string]interface{}{"sql": df.SQL, "input": input}
	case "republish":
		// 重新发布到 MQTT
		// 实际实现需要 EMQX 客户端
		output = map[string]interface{}{"target_topic": df.TargetTopic, "data": input}
	case "forward":
		// 转发到 HTTP
		output = map[string]interface{}{"forwarded": true, "data": input}
	}

	outputJSON, _ := json.Marshal(output)
	execution.OutputJSON = string(outputJSON)

	if execErr != nil {
		execution.Status = "failed"
		execution.ErrorMsg = execErr.Error()
	}

	dao.DB.Create(execution)
	return execErr
}

// ListDataFlowExecutions 数据流转执行记录
func ListDataFlowExecutions(flowID uint, page, size int) (int64, []model.DataFlowExecution, error) {
	var total int64
	var list []model.DataFlowExecution
	q := dao.DB.Model(&model.DataFlowExecution{})
	if flowID > 0 {
		q = q.Where("flow_id = ?", flowID)
	}
	q.Count(&total)
	err := q.Order("created_at desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return total, list, err
}

// ========== 多协议网关 ==========

// CreateProtocolGateway 创建协议网关
func CreateProtocolGateway(name, gatewayType, host string, port int, config string, createdBy uint) (*model.ProtocolGateway, error) {
	g := &model.ProtocolGateway{
		Name:      name,
		Type:      gatewayType,
		Host:      host,
		Port:      port,
		Config:    config,
		Enabled:   true,
		CreatedBy: createdBy,
	}
	if err := dao.DB.Create(g).Error; err != nil {
		return nil, err
	}
	return g, nil
}

// ListProtocolGateways 协议网关列表
func ListProtocolGateways(page, size int) (int64, []model.ProtocolGateway, error) {
	var total int64
	var list []model.ProtocolGateway
	q := dao.DB.Model(&model.ProtocolGateway{})
	q.Count(&total)
	err := q.Order("id desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return total, list, err
}

// CreateProtocolMapping 创建协议映射
func CreateProtocolMapping(gatewayID, productID uint, protocolKey, propertyID, dataType string, scale, offset float64) (*model.ProtocolMapping, error) {
	m := &model.ProtocolMapping{
		GatewayID:   gatewayID,
		ProductID:   productID,
		ProtocolKey: protocolKey,
		PropertyID:  propertyID,
		DataType:    dataType,
		Scale:       scale,
		Offset:      offset,
	}
	if err := dao.DB.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// ListProtocolMappings 协议映射列表
func ListProtocolMappings(gatewayID, productID uint) ([]model.ProtocolMapping, error) {
	var list []model.ProtocolMapping
	q := dao.DB.Model(&model.ProtocolMapping{})
	if gatewayID > 0 {
		q = q.Where("gateway_id = ?", gatewayID)
	}
	if productID > 0 {
		q = q.Where("product_id = ?", productID)
	}
	err := q.Find(&list).Error
	return list, err
}

// ========== 设备诊断/监控 ==========

// CreateDeviceDiagnostic 创建设备诊断记录
func CreateDeviceDiagnostic(deviceSN, diagType, level, title, detail string) (*model.DeviceDiagnostic, error) {
	d := &model.DeviceDiagnostic{
		DeviceSN: deviceSN,
		Type:     diagType,
		Level:    level,
		Title:    title,
		Detail:   detail,
		Status:   "open",
	}
	if err := dao.DB.Create(d).Error; err != nil {
		return nil, err
	}
	return d, nil
}

// ListDeviceDiagnostics 设备诊断列表
func ListDeviceDiagnostics(deviceSN string, page, size int) (int64, []model.DeviceDiagnostic, error) {
	var total int64
	var list []model.DeviceDiagnostic
	q := dao.DB.Model(&model.DeviceDiagnostic{})
	if deviceSN != "" {
		q = q.Where("device_sn = ?", deviceSN)
	}
	q.Count(&total)
	err := q.Order("created_at desc").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return total, list, err
}

// ResolveDeviceDiagnostic 解决诊断问题
func ResolveDeviceDiagnostic(id uint) error {
	now := time.Now()
	return dao.DB.Model(&model.DeviceDiagnostic{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      "resolved",
		"resolved_at": &now,
	}).Error
}

// GetDeviceMetrics 获取设备指标
func GetDeviceMetrics(deviceSN string, days int) ([]model.DeviceMetrics, error) {
	var list []model.DeviceMetrics
	startDate := time.Now().AddDate(0, 0, -days)
	err := dao.DB.Where("device_sn = ? AND date >= ?", deviceSN, startDate).Order("date desc").Find(&list).Error
	return list, err
}

// RecordDeviceMetrics 记录设备指标（定时任务调用）
func RecordDeviceMetrics(deviceSN string) error {
	// 这里简化实现，实际应从 MQTT/Redis 统计
	var metrics model.DeviceMetrics
	today := time.Now().Truncate(24 * time.Hour)
	err := dao.DB.Where("device_sn = ? AND date = ?", deviceSN, today).First(&metrics).Error
	if err != nil {
		// 创建新记录
		metrics = model.DeviceMetrics{
			DeviceSN: deviceSN,
			Date:     today,
		}
	}
	// 更新统计
	metrics.MessageCount++
	metrics.OnlineTime += 60 // 假设每分钟记录一次
	if err := dao.DB.Save(&metrics).Error; err != nil {
		return err
	}
	return nil
}
