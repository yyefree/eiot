package middleware

import (
	"eiot/pkg/config"
	"eiot/pkg/util"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ctxKey string

const (
	UserIDKey ctxKey = "uid"
	RoleKey   ctxKey = "role"
)

// AuthMiddleware Gin 认证中间件
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未登录"})
			return
		}
		claims, err := util.ParseJWT(strings.TrimPrefix(auth, "Bearer "), cfg.JWTSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Token无效: " + err.Error()})
			return
		}
		uidf, _ := claims["uid"]
		uid := uint(0)
		if uidf != nil {
			if f, ok := uidf.(float64); ok {
				uid = uint(f)
			}
		}
		role, _ := claims["role"].(string)
		c.Set(string(UserIDKey), uid)
		c.Set(string(RoleKey), role)
		c.Next()
	}
}

// AdminMiddleware 管理员中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(string(RoleKey))
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "msg": "仅管理员可操作"})
			return
		}
		c.Next()
	}
}

// UID 从上下文读取用户ID
func UID(c *gin.Context) uint {
	if v, ok := c.Get(string(UserIDKey)); ok {
		if u, ok := v.(uint); ok {
			return u
		}
	}
	return 0
}

// Role 从上下文读取角色
func Role(c *gin.Context) string {
	if v, ok := c.Get(string(RoleKey)); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return "user"
}
