package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"memento_backend/db"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware 鉴权中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// 检查token格式 (Bearer token)
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token required",
			})
			c.Abort()
			return
		}

		// 使用TokenService验证token
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tokenService := db.NewTokenService()
		isValid, userToken, err := tokenService.ValidateToken(ctx, token)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Token验证失败",
			})
			c.Abort()
			return
		}

		if !isValid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// 将token和用户信息保存到上下文中，后续处理函数可以使用
		c.Set("token", token)
		c.Set("username", userToken.Name)
		c.Set("channel", userToken.Channel)
		c.Next()
	}
}

// GetUsernameFromContext 从gin.Context获取用户名
func GetUsernameFromContext(c *gin.Context) (string, error) {
	if username, exists := c.Get("username"); exists {
		if name, ok := username.(string); ok {
			return name, nil
		}
	}
	return "", errors.New("无法从context获取用户名")
}

// GetChannelFromContext 从gin.Context获取渠道
func GetChannelFromContext(c *gin.Context) (string, error) {
	if channel, exists := c.Get("channel"); exists {
		if ch, ok := channel.(string); ok {
			return ch, nil
		}
	}
	return "", errors.New("无法从context获取渠道")
}

// RequireAuth 需要鉴权的接口组
func RequireAuth(router *gin.RouterGroup) *gin.RouterGroup {
	authGroup := router.Group("")
	authGroup.Use(AuthMiddleware())
	return authGroup
}
