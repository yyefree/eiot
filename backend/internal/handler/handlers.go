package handler

import (
	"eiot/internal/dao"
	"eiot/internal/logic"
	"eiot/internal/model"
	"eiot/internal/svc"
	"eiot/pkg/middleware"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ========== 辅助：统一响应

// deviceItem 前端展示
type deviceItem struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	DeviceSN    string `json:"device_sn"`
	DeviceKey   string `json:"deviceKey"`
	ProductID   uint   `json:"product_id"`
	ProductKey  string `json:"product_key"`
	ProductName string `json:"productName"`
	BindMode    string `json:"bind_mode"`
	FirmwareVer string `json:"firmware_ver"`
	IPAddress   string `json:"ip_address"`
	OwnerID     uint   `json:"owner_id"`
	OwnerName   string `json:"ownerName"`
	Status      int    `json:"status"`
	Online      bool   `json:"online"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func toDeviceItem(d *model.Device, productName, ownerName string) deviceItem {
	online := false
	if d.LastOnline != nil {
		online = time.Since(*d.LastOnline) < 5*time.Minute
	}
	return deviceItem{
		ID: d.ID, Name: d.DeviceName, DeviceSN: d.DeviceSN, DeviceKey: d.DeviceSN,
		ProductID: d.ProductID, ProductKey: d.ProductKey, ProductName: productName,
		BindMode: d.BindMode, FirmwareVer: d.FirmwareVer, IPAddress: d.IPAddress,
		OwnerID: d.OwnerID, OwnerName: ownerName, Status: d.Status, Online: online,
		CreatedAt: d.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: d.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func adaptDeviceList(list []model.Device) []deviceItem {
	if len(list) == 0 {
		return []deviceItem{}
	}
	pIDs := []uint{}
	uIDs := []uint{}
	for _, d := range list {
		if d.ProductID > 0 {
			pIDs = append(pIDs, d.ProductID)
		}
		if d.OwnerID > 0 {
			uIDs = append(uIDs, d.OwnerID)
		}
	}
	pIDs = uniqueUint(pIDs)
	uIDs = uniqueUint(uIDs)
	pMap := map[uint]string{}
	if len(pIDs) > 0 {
		products := []model.Product{}
		dao.DB.Where("id IN ?", pIDs).Find(&products)
		for _, p := range products {
			pMap[p.ID] = p.Name
		}
	}
	uMap := map[uint]string{}
	if len(uIDs) > 0 {
		users := []model.User{}
		dao.DB.Where("id IN ?", uIDs).Find(&users)
		for _, u := range users {
			uMap[u.ID] = u.Nickname
		}
	}
	out := make([]deviceItem, len(list))
	for i, d := range list {
		out[i] = toDeviceItem(&d, pMap[d.ProductID], uMap[d.OwnerID])
	}
	return out
}

func uniqueUint(in []uint) []uint {
	seen := map[uint]bool{}
	out := []uint{}
	for _, v := range in {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

// deviceDetailResp 设备详情
type deviceDetailResp struct {
	Device     deviceItem             `json:"device"`
	Latest     map[string]string      `json:"latest"`
	ThingModel map[string]interface{} `json:"thingModel"`
	UserUI     string                 `json:"userUI"`
}

// userItem 用户信息
type userItem struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Phone     string `json:"phone"`
	Nickname  string `json:"nickname"`
	Role      string `json:"role"`
	Status    int    `json:"status"`
	CreatedAt string `json:"createdAt"`
}

func toUserItem(u *model.User) userItem {
	return userItem{
		ID: u.ID, Username: u.Username, Phone: u.Phone,
		Nickname: u.Nickname, Role: u.Role, Status: u.Status,
		CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ========== 路由注册

func RegisterHandlers(r *gin.Engine, sc *svc.ServiceContext) {
	// 认证
	api := r.Group("/api")
	{
		api.POST("/auth/login", wrap(sc, handleLogin))
		api.POST("/auth/send-code", wrap(sc, handleSendCode))
		api.POST("/auth/login-code", wrap(sc, handleLoginCode))
		api.GET("/user/info", middleware.AuthMiddleware(sc.Config), wrap(sc, handleUserInfo))
		api.PUT("/user/info", middleware.AuthMiddleware(sc.Config), wrap(sc, handleUpdateUserInfo))

	// 健康检查
	api.GET("/health", wrap(sc, handleHealthCheck))

	// WebSocket 实时推送
	r.GET("/ws", HandleWebSocket)

		// 设备历史数据（用户）
		api.GET("/device/data/:sn", middleware.AuthMiddleware(sc.Config), wrap(sc, handleDeviceDataHistory))

		// 管理员模块
		admin := api.Group("/admin", middleware.AuthMiddleware(sc.Config), middleware.AdminMiddleware())
		{
			// 项目（飞燕标准层级
			admin.POST("/project", wrap(sc, handleCreateProject))
			admin.GET("/project", wrap(sc, handleListProject))
			admin.GET("/project/:id", wrap(sc, handleGetProject))
			admin.PUT("/project/:id", wrap(sc, handleUpdateProject))
			admin.DELETE("/project/:id", wrap(sc, handleDeleteProject))

			// 产品（物模型合一
			admin.POST("/product", wrap(sc, handleCreateProduct))
			admin.GET("/product", wrap(sc, handleListProduct))
			admin.GET("/product/:id", wrap(sc, handleGetProduct))
			admin.PUT("/product/:id", wrap(sc, handleUpdateProduct))
			admin.DELETE("/product/:id", wrap(sc, handleDeleteProduct))
			// 产品物模型
			admin.PUT("/product/:id/thing-model", wrap(sc, handleUpdateProductThingModel))
			// 产品移动端 UI
			admin.GET("/product/:id/mobile-ui", wrap(sc, handleGetProductMobileUI))
			admin.PUT("/product/:id/mobile-ui", wrap(sc, handleSaveProductMobileUI))
			// 设备二维码绑定信息
			admin.GET("/device/:id/bind-qr", wrap(sc, handleDeviceBindQR))

			// 设备
			admin.POST("/device/batch", wrap(sc, handleBatchDevice))
			admin.GET("/device/export/:product_id", wrap(sc, handleExportDevice))
          admin.GET("/device", wrap(sc, handleListDevice))
          admin.PUT("/device/:id", wrap(sc, handleUpdateAdminDevice))
          admin.DELETE("/device/:id", wrap(sc, handleDeleteAdminDevice))

			// 告警规则
			admin.POST("/alert-rule", wrap(sc, handleCreateAlertRule))
			admin.GET("/alert-rule", wrap(sc, handleListAlertRules))
			admin.DELETE("/alert-rule/:id", wrap(sc, handleDeleteAlertRule))
			admin.PUT("/alert-rule/:id/toggle", wrap(sc, handleToggleAlertRule))

			// 用户
			admin.GET("/user", wrap(sc, handleListUser))

			// 看板统计
			admin.GET("/stats", wrap(sc, handleAdminStats))
		}

		// 设备绑定（用户
		api.POST("/device/bind", middleware.AuthMiddleware(sc.Config), wrap(sc, handleBindDevice))
		api.GET("/device", middleware.AuthMiddleware(sc.Config), wrap(sc, handleListDeviceUser))
		api.GET("/device/:id", middleware.AuthMiddleware(sc.Config), wrap(sc, handleDeviceDetail))
    api.POST("/device/:id/control", middleware.AuthMiddleware(sc.Config), wrap(sc, handleDeviceControl))
    // 用户个人 UI
    api.GET("/device/:id/ui", middleware.AuthMiddleware(sc.Config), wrap(sc, handleGetUserDeviceUI))
    api.PUT("/device/:id/ui", middleware.AuthMiddleware(sc.Config), wrap(sc, handleSaveUserDeviceUI))

		// 设备共享
		share := api.Group("/device/share", middleware.AuthMiddleware(sc.Config))
		{
			share.POST("", wrap(sc, handleShareDevice))
			share.DELETE("/:id", wrap(sc, handleRevokeShare))
			share.GET("", wrap(sc, handleListShare))
		}

		// Web 看板
		dash := api.Group("/dashboard", middleware.AuthMiddleware(sc.Config))
		{
			dash.POST("", wrap(sc, handleSaveDashboard))
			dash.GET("", wrap(sc, handleListDashboard))
		}

		// ========== 家庭/房间/场景/消息（云智能App）==========
		auth := middleware.AuthMiddleware(sc.Config)

		// 家庭
		home := api.Group("/home", auth)
		{
			home.POST("", wrap(sc, handleCreateHome))
			home.GET("", wrap(sc, handleListHomes))
			home.GET("/:id", wrap(sc, handleGetHomeDetail))
			home.PUT("/:id", wrap(sc, handleUpdateHome))
			home.DELETE("/:id", wrap(sc, handleDeleteHome))
			home.POST("/:id/member", wrap(sc, handleAddHomeMember))
			home.DELETE("/:id/member/:uid", wrap(sc, handleRemoveHomeMember))
			home.GET("/:id/member", wrap(sc, handleListHomeMembers))
		}

		// 房间
		room := api.Group("/room", auth)
		{
			room.POST("", wrap(sc, handleCreateRoom))
			room.GET("", wrap(sc, handleListRooms))
			room.PUT("/:id", wrap(sc, handleUpdateRoom))
			room.DELETE("/:id", wrap(sc, handleDeleteRoom))
			room.PUT("/reorder", wrap(sc, handleReorderRooms))
			room.POST("/:id/device", wrap(sc, handleAddDeviceToRoom))
			room.DELETE("/:id/device/:did", wrap(sc, handleRemoveDeviceFromRoom))
			room.GET("/:id/device", wrap(sc, handleListRoomDevices))
		}

		// 场景
		scene := api.Group("/scene", auth)
		{
			scene.POST("", wrap(sc, handleCreateScene))
			scene.GET("", wrap(sc, handleListScenes))
			scene.PUT("/:id", wrap(sc, handleUpdateScene))
			scene.DELETE("/:id", wrap(sc, handleDeleteScene))
			scene.PUT("/:id/toggle", wrap(sc, handleToggleScene))
			scene.POST("/:id/run", wrap(sc, handleRunScene))
			scene.POST("/:id/condition", wrap(sc, handleAddSceneCondition))
			scene.DELETE("/:id/condition/:cid", wrap(sc, handleRemoveSceneCondition))
			scene.POST("/:id/action", wrap(sc, handleAddSceneAction))
			scene.DELETE("/:id/action/:aid", wrap(sc, handleRemoveSceneAction))
			scene.PUT("/:id/action/reorder", wrap(sc, handleReorderSceneActions))
		}

		// 消息
		msg := api.Group("/message", auth)
		{
			msg.GET("", wrap(sc, handleListMessages))
			msg.GET("/unread", wrap(sc, handleUnreadCount))
			msg.PUT("/:id/read", wrap(sc, handleMarkMessageRead))
			msg.PUT("/read-all", wrap(sc, handleMarkAllRead))
			msg.DELETE("/:id", wrap(sc, handleDeleteMessage))
		}

		// OTA 管理（Admin）
		admin.POST("/ota", wrap(sc, handleCreateFirmware))
		admin.GET("/ota", wrap(sc, handleListFirmwares))
		admin.POST("/ota/:id/push", wrap(sc, handlePushOTA))
	}
}

// ========== 统一响应包装
// BizError 携带业务 HTTP 状态码的错误类型
type BizError struct {
	HTTPCode int    // Gin 的 HTTP 状态码
	Code     int    // 业务 code（透传给前端）
	Message  string // 错误信息
}

func (e *BizError) Error() string { return e.Message }

func NewBizError(httpCode, bizCode int, msg string) *BizError {
	return &BizError{HTTPCode: httpCode, Code: bizCode, Message: msg}
}

func wrap(sc *svc.ServiceContext, h func(c *gin.Context, sc *svc.ServiceContext) (interface{}, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := h(c, sc)
		// 如果 handler 已经直接写入响应（如 CSV 导出），不再追加 JSON
		if c.Writer.Written() {
			return
		}
		if err != nil {
			if be, ok := err.(*BizError); ok {
				c.JSON(be.HTTPCode, gin.H{"code": be.Code, "msg": be.Message, "data": nil})
				return
			}
			// 默认 500 Internal Server Error
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "服务器内部错误", "data": nil})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok", "data": data})
	}
}

// 辅助
func bodyMap(c *gin.Context) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	if err := c.ShouldBindBodyWithJSON(&m); err != nil {
		return nil, NewBizError(http.StatusBadRequest, 400, "请求 body 解析失败: "+err.Error())
	}
	return m, nil
}

func sVal(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func pageSize(c *gin.Context) (int, int) {
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	sz, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if p <= 0 {
		p = 1
	}
	if sz <= 0 {
		sz = 20
	}
	if sz > 200 {
		sz = 200
	}
	return p, sz
}

// ========== 认证 ==========

func handleLogin(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	token, u, err := logic.LoginByPassword(sVal(b["phone"]), sVal(b["password"]), sc.Config.JWTSecret)
	if err != nil {
		return nil, err
	}
	return gin.H{"token": token, "user": u}, nil
}

func handleSendCode(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	code, err := logic.SendCode(sVal(b["phone"]))
	if err != nil {
		return nil, err
	}
	return gin.H{"expire_sec": 600, "code": code}, nil
}

func handleLoginCode(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	token, u, err := logic.LoginOrRegisterByCode(sVal(b["phone"]), sVal(b["code"]), sc.Config.JWTSecret)
	if err != nil {
		return nil, err
	}
	return gin.H{"token": token, "user": u}, nil
}

func handleUserInfo(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	u, err := logic.GetUserInfo(middleware.UID(c))
	if err != nil {
		return nil, err
	}
	return gin.H{"id": u.ID, "username": u.Username, "phone": u.Phone, "nickname": u.Nickname, "role": u.Role, "avatar": u.Avatar, "status": u.Status, "createdAt": u.CreatedAt.Format("2006-01-02 15:04:05")}, nil
}

// ========== 产品（物模型 ==========

func handleCreateProduct(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	name := sVal(b["name"])
	description := sVal(b["description"])
	icon := sVal(b["icon"])
	productKey := sVal(b["product_key"])
	propertiesJSON := sVal(b["properties_json"])
	eventsJSON := sVal(b["events_json"])
	servicesJSON := sVal(b["services_json"])
	projectID, _ := b["project_id"].(float64)
	networkType := sVal(b["network_type"])
	if propertiesJSON == "" {
		propertiesJSON = "[]"
	}
	if eventsJSON == "" {
		eventsJSON = "[]"
	}
	if servicesJSON == "" {
		servicesJSON = "[]"
	}
	if networkType == "" {
		networkType = "wifi"
	}
	return logic.CreateProduct(name, description, icon, productKey, propertiesJSON, eventsJSON, servicesJSON, uint(projectID), networkType, middleware.UID(c))
}

func handleListProduct(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	page, size := pageSize(c)
	projectID, _ := strconv.Atoi(c.Query("project_id"))
	total, list, err := logic.ListProducts(page, size, uint(projectID))
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(list))
	for _, p := range list {
		var devCount int64
		dao.DB.Model(&model.Device{}).Where("product_id = ?", p.ID).Count(&devCount)
		var props []map[string]interface{}
		json.Unmarshal([]byte(p.PropertiesJSON), &props)
		out = append(out, map[string]interface{}{
			"id":           p.ID,
			"name":         p.Name,
			"description":  p.Description,
			"product_key":  p.ProductKey,
			"project_id":   p.ProjectID,
			"network_type": p.NetworkType,
			"properties":   props,
			"has_devices":  devCount > 0,
			"dev_count":    devCount,
			"created_at":   p.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return gin.H{"total": total, "list": out, "page": page, "size": size}, nil
}

func handleGetProduct(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	p, err := logic.GetProduct(uint(id))
	if err != nil {
		return nil, err
	}
	// 统计设备数
	var devCount int64
	dao.DB.Model(&model.Device{}).Where("product_id = ?", p.ID).Count(&devCount)
	// 解析 properties 方便前端
	var properties, events, services []interface{}
	_ = json.Unmarshal([]byte(p.PropertiesJSON), &properties)
	_ = json.Unmarshal([]byte(p.EventsJSON), &events)
	_ = json.Unmarshal([]byte(p.ServicesJSON), &services)
	return gin.H{
		"id":             p.ID,
		"name":            p.Name,
		"description":     p.Description,
		"product_key":     p.ProductKey,
		"properties":      properties,
		"events":          events,
		"services":        services,
		"properties_json":    p.PropertiesJSON,
		"events_json":     p.EventsJSON,
		"services_json":   p.ServicesJSON,
		"mobile_ui_json": p.MobileUIJSON,
		"has_devices":    devCount > 0,
		"created_at":    p.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func handleUpdateProduct(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	projectID, _ := b["project_id"].(float64)
	networkType := sVal(b["network_type"])
	err = logic.UpdateProduct(uint(id), sVal(b["name"]), sVal(b["description"]), sVal(b["icon"]), uint(projectID), networkType)
	if err != nil {
		return nil, err
	}
	return "ok", nil
}

func handleDeleteProduct(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := logic.DeleteProduct(uint(id))
	if err != nil {
		return nil, err
	}
	return "ok", nil
}

func handleUpdateProductThingModel(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	propertiesJSON := sVal(b["properties_json"])
	eventsJSON := sVal(b["events_json"])
	servicesJSON := sVal(b["services_json"])
	err = logic.UpdateProductThingModel(uint(id), propertiesJSON, eventsJSON, servicesJSON)
	if err != nil {
		return nil, err
	}
	return "ok", nil
}

// ========== 产品移动端 UI

func handleGetProductMobileUI(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	p, err := logic.GetProduct(uint(id))
	if err != nil {
		return nil, err
	}
	// 解析 properties
	var properties []interface{}
	_ = json.Unmarshal([]byte(p.PropertiesJSON), &properties)
	return gin.H{
		"product_id":    p.ID,
		"product_name":  p.Name,
		"properties":   properties,
		"mobile_ui_json": p.MobileUIJSON,
	}, nil
}

func handleSaveProductMobileUI(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	err = logic.SaveProductMobileUI(uint(id), sVal(b["mobile_ui_json"]))
	if err != nil {
		return nil, err
	}
	return gin.H{"ok": true}, nil
}

// ========== 设备（管理员

func handleBatchDevice(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	pid, _ := b["product_id"].(float64)
	prefix := sVal(b["prefix"])
	if prefix == "" {
		prefix = "DEV"
	}
	cnt, _ := b["count"].(float64)
	if cnt == 0 {
		cnt = 10
	}
	bindMode := sVal(b["bind_mode"]) // device_secret(一机一密) / product_secret(一型一密)
	var devs []model.Device
	if bindMode == "product_secret" {
		devs, err = logic.BatchGenerateDevicesWithMode(uint(pid), prefix, int(cnt), "product_secret")
	} else {
		devs, err = logic.BatchGenerateDevices(uint(pid), prefix, int(cnt))
	}
	if err != nil {
		return nil, err
	}
	return gin.H{"count": len(devs), "bind_mode": bindMode, "devices": devs}, nil
}

func handleExportDevice(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("product_id"))
	var devices []model.Device
	if err := dao.DB.Where("product_id = ?", id).Find(&devices).Error; err != nil {
		return nil, err
	}
	// 飞燕标准三元组 CSV 导出（不包含密钥）
	out := "DeviceName,DeviceSN,ProductKey,BindMode\n"
	for _, d := range devices {
		out += fmt.Sprintf("%q,%q,%q,%q\n", d.DeviceName, d.DeviceSN, d.ProductKey, d.BindMode)
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=devices_"+strconv.Itoa(int(id))+".csv")
	c.String(http.StatusOK, "\xEF\xBB\xBF"+out)
	return nil, nil
}

func handleListDevice(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	page, size := pageSize(c)
	var pid uint
	if v := c.Query("product_id"); v != "" {
		fmt.Sscanf(v, "%d", &pid)
	}
	total, list, err := logic.ListDevices(0, "admin", page, size, pid)
	if err != nil {
		return nil, err
	}
	items := adaptDeviceList(list)
	return gin.H{"total": total, "list": items, "page": page, "size": size}, nil
}

// ========== 用户列表

func handleListUser(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	page, size := pageSize(c)
	total, list, err := logic.ListUsers(page, size)
	if err != nil {
		return nil, err
	}
	users := make([]userItem, len(list))
	for i, u := range list {
		users[i] = toUserItem(&u)
	}
	return gin.H{"total": total, "list": users, "page": page, "size": size}, nil
}

// ========== 用户视角设备

func handleBindDevice(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	sn := sVal(b["sn"])
	d, err := logic.BindDeviceBySN(middleware.UID(c), sn)
	if err != nil {
		return nil, err
	}
	return gin.H{"id": d.ID, "name": d.DeviceName, "device_sn": d.DeviceSN}, nil
}

func handleListDeviceUser(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	page, size := pageSize(c)
	total, list, err := logic.ListDevices(middleware.UID(c), middleware.Role(c), page, size, 0)
	if err != nil {
		return nil, err
	}
	items := adaptDeviceList(list)
	return gin.H{"total": total, "list": items, "page": page, "size": size}, nil
}

func handleDeviceDetail(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	d, prod, latest, err := logic.GetDeviceDetail(middleware.UID(c), middleware.Role(c), uint(id))
	if err != nil {
		return nil, err
	}
	ownerName := ""
	if d.OwnerID > 0 {
		var owner model.User
		dao.DB.First(&owner, d.OwnerID)
		ownerName = owner.Nickname
	}
	// 物模型
	tm := map[string]interface{}{
		"product_id": prod.ID, "product_name": prod.Name, "product_key": prod.ProductKey,
	}
	var properties []interface{}
	var events []interface{}
	var services []interface{}
	_ = json.Unmarshal([]byte(prod.PropertiesJSON), &properties)
	_ = json.Unmarshal([]byte(prod.EventsJSON), &events)
	_ = json.Unmarshal([]byte(prod.ServicesJSON), &services)
	tm["properties"] = properties
	tm["events"] = events
	tm["services"] = services

	// 用户个人 UI
	userUI := ""
	if middleware.UID(c) > 0 && d.OwnerID == middleware.UID(c) {
		userUI, _ = logic.GetUserDeviceUI(middleware.UID(c), d.ID)
	}

	return deviceDetailResp{
		Device:     toDeviceItem(d, prod.Name, ownerName),
		Latest:    latest,
		ThingModel: tm,
		UserUI:    userUI,
	}, nil
}

func handleDeviceControl(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	params, _ := b["params"].(map[string]interface{})
	if params == nil {
		params = b
	}
	// 过滤掉非物模型中只读字段（如 token 等内部字段
	if err := logic.ControlDevice(middleware.UID(c), middleware.Role(c), uint(id), params, sc.EMQX); err != nil {
		return nil, err
	}
	return "ok", nil
}

// ========== 用户个人设备 UI

func handleGetUserDeviceUI(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	// 验证权限：只有设备所有者
	var d model.Device
	if err := dao.DB.First(&d, id).Error; err != nil {
		return nil, errors.New("设备不存在")
	}
	if d.OwnerID != middleware.UID(c) {
		return nil, errors.New("无权限")
	}
	ui, _ := logic.GetUserDeviceUI(middleware.UID(c), d.ID)
	return gin.H{"device_id": d.ID, "layout_json": ui}, nil
}

func handleSaveUserDeviceUI(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	// 验证权限
	var d model.Device
	if err := dao.DB.First(&d, id).Error; err != nil {
		return nil, errors.New("设备不存在")
	}
	if d.OwnerID != middleware.UID(c) {
		return nil, errors.New("无权限")
	}
	err = logic.SaveUserDeviceUI(middleware.UID(c), d.ID, sVal(b["layout_json"]))
	if err != nil {
		return nil, err
	}
	return gin.H{"ok": true}, nil
}

// ========== 共享

func handleShareDevice(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	devID, _ := b["device_id"].(float64)
	shareUID, _ := b["share_user_id"].(float64)
	perm := sVal(b["permission"])
	if perm == "" {
		perm = "read"
	}
	hours, _ := b["hours"].(float64)
	var exp time.Time
	if hours > 0 {
		exp = time.Now().Add(time.Duration(hours) * time.Hour)
	}
	err = logic.ShareDevice(middleware.UID(c), uint(devID), uint(shareUID), perm, exp)
	if err != nil {
		return nil, err
	}
	return "ok", nil
}

func handleRevokeShare(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := logic.RevokeShare(middleware.UID(c), uint(id))
	if err != nil {
		return nil, err
	}
	return "ok", nil
}

func handleListShare(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	page, size := pageSize(c)
	total, list, err := logic.ListShares(middleware.UID(c), page, size)
	if err != nil {
		return nil, err
	}
	return gin.H{"total": total, "list": list, "page": page, "size": size}, nil
}

// ========== 看板

func handleSaveDashboard(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	typ := sVal(b["type"])
	name := sVal(b["name"])
	if name == "" {
		name = "default"
	}
	layout := sVal(b["layout_json"])
	err = logic.SaveDashboard(middleware.UID(c), typ, name, layout)
	if err != nil {
		return nil, err
	}
	return "ok", nil
}

func handleListDashboard(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	typ := strings.TrimSpace(c.Query("type"))
	list, err := logic.ListDashboards(middleware.UID(c), typ)
	if err != nil {
		return nil, err
	}
	return gin.H{"list": list}, nil
}

// ========== 管理员统计

func handleAdminStats(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	var deviceTotal, productTotal, userTotal int64
	dao.DB.Model(&model.Device{}).Count(&deviceTotal)
	dao.DB.Model(&model.Product{}).Count(&productTotal)
	dao.DB.Model(&model.User{}).Count(&userTotal)
	var online int64
	dao.DB.Model(&model.Device{}).Where("updated_at > ?", time.Now().Add(-5*time.Minute)).Count(&online)
	offline := deviceTotal - online

	// 最近 20 台设备
	var recentDevices []model.Device
	dao.DB.Order("id DESC").Limit(20).Find(&recentDevices)
	deviceItems := adaptDeviceList(recentDevices)

	return gin.H{
		"devices":       deviceTotal,
		"products":      productTotal,
		"users":         userTotal,
		"online":        online,
		"offline":       offline,
		"thingModels":   productTotal,
		"devices_list":  deviceItems,
	}, nil
}

// ========== 项目（飞燕标准：项目→产品→设备 ==========

func handleCreateProject(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	name := sVal(b["name"])
	description := sVal(b["description"])
	industry := sVal(b["industry"])
	projType := sVal(b["type"])
	if projType == "" {
		projType = "consumer"
	}
	return logic.CreateProject(name, description, industry, projType, middleware.UID(c))
}

func handleListProject(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	page, size := pageSize(c)
	total, list, err := logic.ListProjects(page, size)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(list))
	for _, p := range list {
		var prodCount int64
		dao.DB.Model(&model.Product{}).Where("project_id = ?", p.ID).Count(&prodCount)
		out = append(out, map[string]interface{}{
			"id":          p.ID,
			"name":        p.Name,
			"description": p.Description,
			"industry":    p.Industry,
			"type":        p.Type,
			"prod_count":  prodCount,
			"created_at":  p.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return gin.H{"total": total, "list": out, "page": page, "size": size}, nil
}

func handleGetProject(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	var p model.Project
	if err := dao.DB.First(&p, id).Error; err != nil {
		return nil, errors.New("项目不存在")
	}
	// 统计产品和设备
	var prodCount, devCount int64
	dao.DB.Model(&model.Product{}).Where("project_id = ?", p.ID).Count(&prodCount)
	if prodCount > 0 {
		var pids []uint
		dao.DB.Model(&model.Product{}).Where("project_id = ?", p.ID).Pluck("id", &pids)
		dao.DB.Model(&model.Device{}).Where("product_id IN ?", pids).Count(&devCount)
	}
	return gin.H{
		"id": p.ID, "name": p.Name, "description": p.Description,
		"industry": p.Industry, "type": p.Type,
		"prod_count": prodCount, "dev_count": devCount,
		"created_at": p.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func handleUpdateProject(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	err = logic.UpdateProject(uint(id), sVal(b["name"]), sVal(b["description"]), sVal(b["industry"]), sVal(b["type"]))
	if err != nil {
		return nil, err
	}
	return "ok", nil
}

func handleDeleteProject(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := logic.DeleteProject(uint(id))
	if err != nil {
		return nil, err
	}
	return "ok", nil
}

// ========== 设备二维码绑定 ==========

func handleDeviceBindQR(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	// 生成包含三元组的绑定 URL
	var d model.Device
	if err := dao.DB.First(&d, id).Error; err != nil {
		return nil, errors.New("设备不存在")
	}
	// 返回三元组供客户端生成二维码（不含 DeviceSecret）
	return gin.H{
		"product_key": d.ProductKey,
		"device_name": d.DeviceName,
		"device_sn":   d.DeviceSN,
		"bind_url": fmt.Sprintf("eiot://bind?pk=%s&dn=%s&sn=%s", d.ProductKey, d.DeviceName, d.DeviceSN),
	}, nil
}

// ========== 家庭管理 ==========

func handleCreateHome(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	uid := middleware.UID(c)
	return logic.CreateHome(sVal(b["name"]), sVal(b["address"]), sVal(b["icon"]), uid)
}

func handleListHomes(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	return logic.ListHomes(middleware.UID(c))
}

func handleGetHomeDetail(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	home, members, roomCount, err := logic.GetHomeDetail(uint(id))
	if err != nil {
		return nil, err
	}
	return gin.H{"home": home, "members": members, "room_count": roomCount}, nil
}

func handleUpdateHome(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	return nil, logic.UpdateHome(uint(id), sVal(b["name"]), sVal(b["address"]), sVal(b["icon"]))
}

func handleDeleteHome(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	return nil, logic.DeleteHome(uint(id))
}

func handleAddHomeMember(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	uid := uint(b["user_id"].(float64))
	m, e := logic.AddHomeMember(uint(id), uid, sVal(b["role"]), sVal(b["nickname"]))
	return m, e
}

func handleRemoveHomeMember(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	uid, _ := strconv.Atoi(c.Param("uid"))
	return nil, logic.RemoveHomeMember(uint(id), uint(uid))
}

func handleListHomeMembers(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	return logic.ListHomeMembers(uint(id))
}

// ========== 房间管理 ==========

func handleCreateRoom(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	homeID := uint(b["home_id"].(float64))
	sortOrder := 0
	if v, ok := b["sort_order"].(float64); ok {
		sortOrder = int(v)
	}
	return logic.CreateRoom(homeID, sVal(b["name"]), sVal(b["icon"]), sortOrder)
}

func handleListRooms(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	homeID, _ := strconv.Atoi(c.Query("home_id"))
	return logic.ListRooms(uint(homeID))
}

func handleUpdateRoom(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	return nil, logic.UpdateRoom(uint(id), sVal(b["name"]), sVal(b["icon"]))
}

func handleDeleteRoom(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	return nil, logic.DeleteRoom(uint(id))
}

func handleReorderRooms(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	homeID := uint(b["home_id"].(float64))
	items, _ := b["items"].([]interface{})
	var ids []uint
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			ids = append(ids, uint(m["id"].(float64)))
		}
	}
	return nil, logic.ReorderRooms(homeID, ids)
}

func handleAddDeviceToRoom(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	deviceID := uint(b["device_id"].(float64))
	return nil, logic.AddDeviceToRoom(uint(id), deviceID)
}

func handleRemoveDeviceFromRoom(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	did, _ := strconv.Atoi(c.Param("did"))
	return nil, logic.RemoveDeviceFromRoom(uint(id), uint(did))
}

func handleListRoomDevices(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	return logic.ListRoomDevices(uint(id))
}

// ========== 场景管理 ==========

func handleCreateScene(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	homeID := uint(b["home_id"].(float64))
	sortOrder := 0
	if v, ok := b["sort_order"].(float64); ok {
		sortOrder = int(v)
	}
	return logic.CreateScene(homeID, sVal(b["name"]), sVal(b["icon"]), sVal(b["type"]), sortOrder)
}

func handleListScenes(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	homeID, _ := strconv.Atoi(c.Query("home_id"))
	return logic.ListScenes(uint(homeID))
}

func handleUpdateScene(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	return nil, logic.UpdateScene(uint(id), sVal(b["name"]), sVal(b["icon"]), sVal(b["type"]))
}

func handleDeleteScene(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	return nil, logic.DeleteScene(uint(id))
}

func handleToggleScene(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, _ := bodyMap(c)
	id, _ := strconv.Atoi(c.Param("id"))
	enabled := true
	if v, ok := b["enabled"].(bool); ok {
		enabled = v
	}
	return nil, logic.ToggleScene(uint(id), enabled)
}

func handleRunScene(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	return nil, logic.RunScene(uint(id), sc.EMQX)
}

func handleAddSceneCondition(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	m, e := logic.AddSceneCondition(uint(id), sVal(b["type"]), sVal(b["config_json"]))
	return m, e
}

func handleRemoveSceneCondition(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	cid, _ := strconv.Atoi(c.Param("cid"))
	return nil, logic.RemoveSceneCondition(uint(cid))
}

func handleAddSceneAction(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	deviceID := uint(b["device_id"].(float64))
	sortOrder := 0
	if v, ok := b["sort_order"].(float64); ok {
		sortOrder = int(v)
	}
	m, e := logic.AddSceneAction(uint(id), deviceID, sVal(b["action_json"]), sortOrder)
	return m, e
}

func handleRemoveSceneAction(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	aid, _ := strconv.Atoi(c.Param("aid"))
	return nil, logic.RemoveSceneAction(uint(aid))
}

func handleReorderSceneActions(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	sceneID := uint(b["scene_id"].(float64))
	items, _ := b["items"].([]interface{})
	var ids []uint
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			ids = append(ids, uint(m["id"].(float64)))
		}
	}
	return nil, logic.ReorderSceneActions(sceneID, ids)
}

// ========== 消息管理 ==========

func handleListMessages(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	uid := middleware.UID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	total, list, err := logic.ListMessages(uid, page, size)
	if err != nil {
		return nil, err
	}
	return gin.H{"total": total, "list": list, "page": page, "size": size}, nil
}

func handleUnreadCount(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	return logic.UnreadCount(middleware.UID(c))
}

func handleMarkMessageRead(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	return nil, logic.MarkMessageRead(uint(id), middleware.UID(c))
}

func handleMarkAllRead(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	return nil, logic.MarkAllRead(middleware.UID(c))
}

func handleDeleteMessage(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	return nil, logic.DeleteMessage(uint(id), middleware.UID(c))
}

// ========== OTA 管理 ==========

func handleCreateFirmware(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	productID := uint(b["product_id"].(float64))
	return logic.CreateFirmware(productID, sVal(b["version"]), sVal(b["changelog"]), sVal(b["file_url"]), int64(b["size"].(float64)), middleware.UID(c))
}

func handleListFirmwares(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	productID, _ := strconv.Atoi(c.Query("product_id"))
	total, list, err := logic.ListFirmwares(uint(productID), page, size)
	if err != nil {
		return nil, err
	}
	return gin.H{"total": total, "list": list, "page": page, "size": size}, nil
}

func handlePushOTA(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	return nil, logic.PushOTA(uint(id), sc.EMQX)
}

// ========== 管理员设备管理 ==========

func handleUpdateAdminDevice(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	name := sVal(b["name"])
	status := -1
	if v, ok := b["status"].(float64); ok {
		status = int(v)
	}
	return nil, logic.UpdateAdminDevice(uint(id), name, status)
}

func handleDeleteAdminDevice(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	return nil, logic.DeleteAdminDevice(uint(id))
}

// ========== 设备历史数据 ==========

func handleDeviceDataHistory(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	sn := c.Param("sn")
	property := c.Query("property")
	startStr := c.Query("start_time")
	endStr := c.Query("end_time")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "500"))

	var startTime, endTime time.Time
	if startStr != "" {
		startTime, _ = time.Parse("2006-01-02T15:04:05", startStr)
	}
	if endStr != "" {
		endTime, _ = time.Parse("2006-01-02T15:04:05", endStr)
	}

	list, err := logic.GetDeviceDataHistory(sn, property, startTime, endTime, limit)
	if err != nil {
		return nil, err
	}
	return gin.H{"list": list, "total": len(list)}, nil
}

// ========== 用户资料编辑 ==========

func handleUpdateUserInfo(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	return nil, logic.UpdateUserInfo(middleware.UID(c), sVal(b["nickname"]), sVal(b["avatar"]), sVal(b["email"]))
}

// ========== 健康检查 ==========

func handleHealthCheck(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	return gin.H{"status": "ok", "time": time.Now().Format("2006-01-02 15:04:05")}, nil
}

// ========== 告警规则 ==========

func handleCreateAlertRule(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	b, err := bodyMap(c)
	if err != nil {
		return nil, err
	}
	productID := uint(b["product_id"].(float64))
	return logic.CreateAlertRule(productID, sVal(b["device_sn"]), sVal(b["property"]),
		sVal(b["operator"]), sVal(b["threshold"]), sVal(b["notify_type"]),
		sVal(b["notify_url"]), middleware.UID(c))
}

func handleListAlertRules(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	page, size := pageSize(c)
	productID, _ := strconv.Atoi(c.Query("product_id"))
	total, list, err := logic.ListAlertRules(uint(productID), page, size)
	if err != nil {
		return nil, err
	}
	return gin.H{"total": total, "list": list, "page": page, "size": size}, nil
}

func handleDeleteAlertRule(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	return nil, logic.DeleteAlertRule(uint(id))
}

func handleToggleAlertRule(c *gin.Context, sc *svc.ServiceContext) (interface{}, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	b, _ := bodyMap(c)
	enabled := true
	if v, ok := b["enabled"].(bool); ok {
		enabled = v
	}
	return nil, logic.ToggleAlertRule(uint(id), enabled)
}
