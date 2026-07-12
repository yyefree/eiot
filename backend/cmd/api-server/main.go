package main

import (
	"eiot/internal/dao"
	"eiot/internal/handler"
	"eiot/internal/logic"
	"eiot/internal/svc"
	"eiot/pkg/cache"
	"eiot/pkg/config"
	"eiot/pkg/mqtt"
	"eiot/pkg/tsdb"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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
		log.Printf("[WARN] MySQL init failed: %v", err)
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
			logic.HandleDeviceReportByPKAndName(productKey, deviceName, params)
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
		log.Println("[EMQX] subscribed Feiyan-style topics: /sys/+/+/prop/post, /ota/device/inform/+/+")
	}

	sc := svc.NewServiceContext(&c)
	sc.EMQX = emqx

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// CORS
	origins := sc.Config.CORSOrigins
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
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
