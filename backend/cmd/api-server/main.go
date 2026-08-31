package main

import (
	"eiot/internal/dao"
	"eiot/internal/handler"
	"eiot/internal/logic"
	"eiot/internal/model"
	"eiot/internal/svc"
	"eiot/pkg/cache"
	"eiot/pkg/config"
	"eiot/pkg/middleware"
	"eiot/pkg/mqtt"
	"eiot/pkg/tsdb"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// 默认配置
func defaultConfig() config.Config {
	return config.Config{
		Name:          "eiot-api",
		Host:          "0.0.0.0",
		Port:          8080,
		JWTSecret:     "eiot-super-secret-key-change-me",
		AdminPhone:    "13800000000",
		AdminPassword: "admin123",
		MySQL:         config.MySQLConf{DSN: "root:eiot@tcp(mysql:3306)/eiot?charset=utf8mb4&parseTime=True&loc=Local"},
		Redis:         config.RedisConf{Addr: "redis:6379"},
		TDengine:      config.TDengineConf{DSN: ""},
		EMQX:          config.EMQXConf{Broker: "tcp://emqx:1883", Username: "eiot", Password: "eiot2026"},
	}
}

func applyEnv(c *config.Config) {
	if v := os.Getenv("EIOT_JWT_SECRET"); v != "" {
		c.JWTSecret = v
	}
	if v := os.Getenv("EIOT_ADMIN_PHONE"); v != "" {
		c.AdminPhone = v
	}
	if v := os.Getenv("EIOT_ADMIN_PASSWORD"); v != "" {
		c.AdminPassword = v
	}
	if v := os.Getenv("EIOT_DB"); v != "" {
		c.MySQL.DSN = v
	}
	if v := os.Getenv("EIOT_REDIS"); v != "" {
		c.Redis.Addr = v
	}
	if v := os.Getenv("EIOT_EMQX"); v != "" {
		c.EMQX.Broker = v
	}
	if v := os.Getenv("EIOT_EMQX_USER"); v != "" {
		c.EMQX.Username = v
	}
	if v := os.Getenv("EIOT_EMQX_PASS"); v != "" {
		c.EMQX.Password = v
	}
	if v := os.Getenv("EIOT_PORT"); v != "" {
		var p int
		if _, err := fmt.Sscanf(v, "%d", &p); err == nil && p > 0 {
			c.Port = p
		}
	}
}

func main() {
	c := defaultConfig()
	if data, err := os.ReadFile("api/config.yaml"); err == nil {
		_ = yaml.Unmarshal(data, &c)
	}
	applyEnv(&c)

	log.Printf("[EIOT] starting server %s:%d", c.Host, c.Port)

	if err := dao.InitMySQL(c.MySQL); err != nil {
		log.Fatalf("[FATAL] MySQL init failed: %v", err)
	}

	logic.InitAdmin(c.AdminPhone, c.AdminPassword)
	if err := cache.InitRedis(c.Redis); err != nil {
		log.Printf("[WARN] Redis init failed: %v", err)
	}
	if err := tsdb.InitTDengine(c.TDengine.DSN); err != nil {
		log.Printf("[WARN] TDengine init skipped: %v", err)
	}

	logic.SeedDemoData(c.AdminPhone, c.AdminPassword)

	var emqx *mqtt.EMQXClient
	if cli, err := mqtt.NewEMQXClient(c.EMQX.Broker, c.EMQX.Username, c.EMQX.Password, c.EMQX.ClientID); err != nil {
		log.Printf("[WARN] EMQX connect failed: %v", err)
	} else {
		emqx = cli
		// 飞燕标准 Topic：设备属性上报 /sys/{ProductKey}/{DeviceName}/prop/post
		// 兼容 Alink JSON 格式：{"id":"xxx","version":"1.0","params":{"temp":25},"method":"thing.event.property.post"}
		_ = emqx.Subscribe("/sys/+/+/prop/post", func(topic string, payload []byte) {
			segs := strings.Split(strings.Trim(topic, "/"), "/")
			if len(segs) < 4 {
				return
			}
			productKey := segs[1]
			deviceName := segs[2]
			var body map[string]interface{}
			if err := json.Unmarshal(payload, &body); err != nil {
				return
			}
			// 兼容飞燕格式：params 内为属性值，或直接是顶层
			params, ok := body["params"].(map[string]interface{})
			if !ok {
				params = body
			}
			// 查找设备 SN 并记录 MQTT 报文（异步）
			var d model.Device
			if err := dao.DB.Where("product_key = ? AND device_name = ?", productKey, deviceName).First(&d).Error; err == nil {
				logic.LogMqttMessageAsync(d.DeviceSN, topic, "up", string(payload))
				logic.HandleDeviceReportByPKAndName(productKey, deviceName, params)
			}
		})

		// 飞燕标准 Topic：设备批量属性上报 /sys/{ProductKey}/{DeviceName}/thing/event/property/batch_post
		_ = emqx.Subscribe("/sys/+/+/thing/event/property/batch_post", func(topic string, payload []byte) {
			segs := strings.Split(strings.Trim(topic, "/"), "/")
			if len(segs) < 4 {
				return
			}
			productKey := segs[1]
			deviceName := segs[2]
			var body map[string]interface{}
			if err := json.Unmarshal(payload, &body); err != nil {
				return
			}
			// 批量上报格式: {"id":"xxx","version":"1.0","method":"thing.event.property.batch.post","params":[{"time":1234567890,"value":{"temp":25}},{"time":1234567891,"value":{"temp":26}}]}
			paramsList, ok := body["params"].([]interface{})
			if !ok {
				return
			}
			var d model.Device
			if err := dao.DB.Where("product_key = ? AND device_name = ?", productKey, deviceName).First(&d).Error; err != nil {
				return
			}
			logic.LogMqttMessageAsync(d.DeviceSN, topic, "up", string(payload))
			for _, item := range paramsList {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if props, ok := itemMap["value"].(map[string]interface{}); ok {
						logic.HandleDeviceReportByPKAndName(productKey, deviceName, props)
					}
				}
			}
		})

		// 飞燕标准 Topic：设备历史属性上报 /sys/{ProductKey}/{DeviceName}/thing/event/property/history/post
		_ = emqx.Subscribe("/sys/+/+/thing/event/property/history/post", func(topic string, payload []byte) {
			segs := strings.Split(strings.Trim(topic, "/"), "/")
			if len(segs) < 4 {
				return
			}
			productKey := segs[1]
			deviceName := segs[2]
			var body map[string]interface{}
			if err := json.Unmarshal(payload, &body); err != nil {
				return
			}
			// 历史数据上报格式: {"id":"xxx","version":"1.0","method":"thing.event.property.history.post","params":[{"time":1234567890,"value":{"temp":25}},{"time":1234567891,"value":{"temp":26}}]}
			paramsList, ok := body["params"].([]interface{})
			if !ok {
				return
			}
			var d model.Device
			if err := dao.DB.Where("product_key = ? AND device_name = ?", productKey, deviceName).First(&d).Error; err != nil {
				return
			}
			logic.LogMqttMessageAsync(d.DeviceSN, topic, "up", string(payload))
			for _, item := range paramsList {
				if itemMap, ok := item.(map[string]interface{}); ok {
					timeVal, _ := itemMap["time"].(float64)
					if props, ok := itemMap["value"].(map[string]interface{}); ok {
						// 保存历史数据，包含时间戳
						for k, v := range props {
							logic.SaveDeviceDataWithTime(d.DeviceSN, k, fmt.Sprintf("%v", v), time.Unix(int64(timeVal), 0))
						}
					}
				}
			}
		})

		// 飞燕标准 Topic：设备固件版本上报 /ota/device/inform/{productKey}/{deviceName}
		_ = emqx.Subscribe("/ota/device/inform/+/+", func(topic string, payload []byte) {
			segs := strings.Split(strings.Trim(topic, "/"), "/")
			if len(segs) < 4 {
				return
			}
			productKey := segs[2]
			deviceName := segs[3]
			var body map[string]interface{}
			if err := json.Unmarshal(payload, &body); err != nil {
				return
			}
			if ver, ok := body["version"].(string); ok {
				logic.HandleDeviceFirmwareReport(productKey, deviceName, ver)
			}
		})

		// 飞燕标准 Topic：设备事件上报 /sys/{ProductKey}/{DeviceName}/event/post
		_ = emqx.Subscribe("/sys/+/+/event/post", func(topic string, payload []byte) {
			segs := strings.Split(strings.Trim(topic, "/"), "/")
			if len(segs) < 4 {
				return
			}
			productKey := segs[1]
			deviceName := segs[2]
			var body map[string]interface{}
			if err := json.Unmarshal(payload, &body); err != nil {
				return
			}
			eventId, _ := body["event_id"].(string)
			eventName, _ := body["event_name"].(string)
			output, _ := json.Marshal(body["output"])

			var d model.Device
			if err := dao.DB.Where("product_key = ? AND device_name = ?", productKey, deviceName).First(&d).Error; err != nil {
				return
			}
			logic.LogMqttMessageAsync(d.DeviceSN, topic, "up", string(payload))
			logic.ReportDeviceEvent(d.DeviceSN, eventId, eventName, string(output))
		})

		// 飞燕标准 Topic：设备服务响应 /sys/{ProductKey}/{DeviceName}/service/{ServiceId}/reply
		_ = emqx.Subscribe("/sys/+/+/service/+/reply", func(topic string, payload []byte) {
			segs := strings.Split(strings.Trim(topic, "/"), "/")
			if len(segs) < 6 {
				return
			}
			serviceId := segs[4]
			var body map[string]interface{}
			if err := json.Unmarshal(payload, &body); err != nil {
				return
			}
			productKey := segs[1]
			deviceName := segs[2]
			var d model.Device
			if err := dao.DB.Where("product_key = ? AND device_name = ?", productKey, deviceName).First(&d).Error; err != nil {
				return
			}
			logic.LogMqttMessageAsync(d.DeviceSN, topic, "up", string(payload))
			outputJSON, _ := json.Marshal(body["output"])
			dao.DB.Model(&model.DeviceServiceHistory{}).
				Where("device_sn = ? AND service_id = ? AND status = 'success'", d.DeviceSN, serviceId).
				Update("output_json", string(outputJSON))
		})

		// 飞燕标准 Topic：RRPC 同步服务调用请求 /sys/{ProductKey}/{DeviceName}/rrpc/request/${messageId}
		_ = emqx.Subscribe("/sys/+/+/rrpc/request/+", func(topic string, payload []byte) {
			segs := strings.Split(strings.Trim(topic, "/"), "/")
			if len(segs) < 6 {
				return
			}
			productKey := segs[1]
			deviceName := segs[2]
			messageId := segs[5]
			var body map[string]interface{}
			if err := json.Unmarshal(payload, &body); err != nil {
				return
			}
			var d model.Device
			if err := dao.DB.Where("product_key = ? AND device_name = ?", productKey, deviceName).First(&d).Error; err != nil {
				return
			}
			logic.LogMqttMessageAsync(d.DeviceSN, topic, "up", string(payload))
			
			// 提取服务调用参数
			method, _ := body["method"].(string)
			params, _ := body["params"].(map[string]interface{})
			
			// 记录服务调用历史（等待设备响应）
			history := &model.DeviceServiceHistory{
				DeviceSN:    d.DeviceSN,
				ServiceID:   method,
				ServiceName: method,
				InputJSON:   "",
				OutputJSON:  "",
				Status:      "pending",
				CreatedAt:   time.Now(),
			}
			if paramsJSON, err := json.Marshal(params); err == nil {
				history.InputJSON = string(paramsJSON)
			}
			dao.DB.Create(history)
			
			// 设备需要响应到 /sys/${pk}/${dn}/rrpc/response/${messageId}
			// 这里由设备端处理响应，云端通过轮询或回调获取结果
			_ = messageId // 用于响应匹配
		})

	// 飞燕标准 Topic：设备动态注册 /sys/{ProductKey}/{DeviceName}/thing/register
	_ = emqx.Subscribe("/sys/+/+/thing/register", func(topic string, payload []byte) {
		segs := strings.Split(strings.Trim(topic, "/"), "/")
		if len(segs) < 4 {
			return
		}
		productKey := segs[1]
		deviceName := segs[2]
		var body map[string]interface{}
		if err := json.Unmarshal(payload, &body); err != nil {
			return
		}
		
		var d model.Device
		if err := dao.DB.Where("product_key = ? AND device_name = ?", productKey, deviceName).First(&d).Error; err != nil {
			// 设备不存在，可能是预注册模式，创建新设备
			var p model.Product
			if err := dao.DB.Where("product_key = ?", productKey).First(&p).Error; err != nil {
				return
			}
			// 生成随机 DeviceSecret
			deviceSecret := fmt.Sprintf("%s%08d", productKey[:8], rand.Intn(100000000))
			d = model.Device{
				DeviceName:    deviceName,
				DeviceSN:      fmt.Sprintf("%s_%s", productKey, deviceName),
				DeviceSecret:  deviceSecret,
				ProductKey:    productKey,
				ProductID:     p.ID,
				OwnerID:       1, // 系统用户
				Status:        1,
				NodeType:      "device",
				BindMode:      "product_secret",
			}
			if err := dao.DB.Create(&d).Error; err != nil {
				return
			}
		}
		
		// 返回设备密钥
		response := map[string]interface{}{
			"id":      body["id"],
			"code":    200,
			"message": "success",
			"data": map[string]interface{}{
				"deviceSecret": d.DeviceSecret,
			},
		}
		responseJSON, _ := json.Marshal(response)
		replyTopic := fmt.Sprintf("/sys/%s/%s/thing/register_reply", productKey, deviceName)
		emqx.Publish(replyTopic, response)
		logic.LogMqttMessageAsync(d.DeviceSN, replyTopic, "down", string(responseJSON))
	})

	log.Println("[EMQX] subscribed Feiyan-style topics: /sys/+/+/prop/post, /sys/+/+/thing/event/property/batch_post, /sys/+/+/thing/event/property/history_post, /ota/device/inform/+/+, /sys/+/+/event/post, /sys/+/+/service/+/reply, /sys/+/+/rrpc/request/+, /sys/+/+/thing/register")
	}

	sc := svc.NewServiceContext(&c)
	sc.EMQX = emqx

	// 设置设备数据上报回调，广播到 WebSocket 客户端
	logic.OnDeviceData = handler.BroadcastDeviceData

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 速率限制：全局 100 req/min, 登录/发验证码 10 req/min
	r.Use(middleware.RateLimitMiddleware(100, time.Minute))

	// CORS
	origins := sc.Config.CORSOrigins
	if origins == "" {
		origins = "http://localhost:8088,http://localhost:5173"
	}
	r.Use(func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")
		allowOrigin := "*"
		if origins != "" && origins != "*" {
			// 白名单模式：检查请求 Origin 是否在白名单中
			allowed := false
			for _, o := range strings.Split(origins, ",") {
				o = strings.TrimSpace(o)
				if o == origin || o == "*" {
					allowed = true
					break
				}
			}
			if !allowed {
				allowOrigin = ""
			} else {
				allowOrigin = origin
			}
		}
		ctx.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
		ctx.Next()
	})

	handler.RegisterHandlers(r, sc)

	// 前端静态资源（若存在）
	if _, err := os.Stat("frontend/dist"); err == nil {
		r.Static("/static", "frontend/dist")
		r.NoRoute(func(ctx *gin.Context) {
			if strings.HasPrefix(ctx.Request.URL.Path, "/api") {
				ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "not found"})
				return
			}
			ctx.File("frontend/dist/index.html")
		})
	} else {
		r.NoRoute(func(ctx *gin.Context) {
			if strings.HasPrefix(ctx.Request.URL.Path, "/api") {
				ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "not found"})
				return
			}
			ctx.JSON(http.StatusOK, gin.H{"msg": "eiot api server", "time": time.Now().Format("2006-01-02 15:04:05")})
		})
	}

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	log.Printf("[EIOT] listening on %s", addr)
	srv := &http.Server{Addr: addr, Handler: r}

	// 优雅关机
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[EIOT] shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[EIOT] forced shutdown: %v", err)
	}
	log.Println("[EIOT] server stopped")
}
