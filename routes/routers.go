package routes

import (
	"context"
	"fmt"
	"memento_backend/db"
	"memento_backend/middleware"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"snail.local/snailllllll/napcat_go_sdk"
	"snail.local/snailllllll/verification"
)

// SetupRoutes 配置所有路由
func SetupRoutes(router *gin.Engine, userService *db.UserService, verificationService *verification.VerificationCodeService, wsClient *napcat_go_sdk.WebSocketClient) {
	// CORS 跨域中间件：允许跨域请求
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	// 基础路由
	setupBasicRoutes(router)

	// 消息相关路由
	setupMessageRoutes(router)

	// 用户管理路由
	setupUserRoutes(router, userService)

	// 工具路由
	setupToolRoutes(router, verificationService)
}

// 基础路由
func setupBasicRoutes(router *gin.Engine) {
	// 根路由
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World!",
		})
	})

	// 图片查看接口
	router.GET("/pic/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		filePath := filepath.Join(".", "pics", filename)

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}

		// Serve the file
		c.File(filePath)
	})
}

// 消息相关路由
func setupMessageRoutes(router *gin.Engine) {
	// 创建需要鉴权的接口组
	authGroup := router.Group("")
	authGroup.Use(middleware.AuthMiddleware())
	{
		// 获取具体消息（需要鉴权）
		authGroup.GET("/messages/:id", func(c *gin.Context) {
			// url 参数
			id, err := primitive.ObjectIDFromHex(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message ID"})
				return
			}

			collection := db.Collection("message_db", "forward_views")
			var message bson.M
			if err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&message); err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
				return
			}

			c.JSON(http.StatusOK, message)
		})

		// 获取消息列表（需要鉴权）
		authGroup.GET("/message_list", func(c *gin.Context) {
			collection := db.Collection("message_db", "forward_views")
			cursor, err := collection.Find(context.Background(), bson.M{})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query messages"})
				return
			}

			var messages []bson.M
			if err := cursor.All(context.Background(), &messages); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode messages"})
				return
			}

			// 只返回title和id字段
			var result []map[string]interface{}
			for i, msg := range messages {
				title, ok := msg["title"] // 使用双返回值形式获取map值，可以安全判断字段是否存在
				if !ok || title == nil || title == "" {
					title = fmt.Sprintf("聊天%d", i+1) //未没有标题的聊天使用序号作为标题
				}
				result = append(result, map[string]interface{}{
					"id":    msg["_id"],
					"title": title,
				})
			}

			c.JSON(http.StatusOK, result)
		})

		// 获取全部消息 deprecated（需要鉴权）
		authGroup.GET("/messages", func(c *gin.Context) {
			collection := db.Collection("message_db", "forward_views")
			cursor, err := collection.Find(context.Background(), bson.M{})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query messages"})
				return
			}

			var messages []bson.M
			if err := cursor.All(context.Background(), &messages); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode messages"})
				return
			}

			c.JSON(http.StatusOK, messages)
		})
	}
}

// 用户管理路由
func setupUserRoutes(router *gin.Engine, userService *db.UserService) {
	// 创建用户
	router.POST("/users", func(c *gin.Context) {
		var user db.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := userService.CreateUser(c.Request.Context(), &user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "用户创建成功",
			"user":    user,
		})
	})

	// 获取所有用户
	router.GET("/users", func(c *gin.Context) {
		users, err := userService.GetAllUsers(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"users": users,
			"count": len(users),
		})
	})

	// 根据用户名获取用户
	router.GET("/users/name/:name", func(c *gin.Context) {
		name := c.Param("name")
		user, err := userService.GetUserByName(c.Request.Context(), name)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	})

	// 根据ID获取用户
	router.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		user, err := userService.GetUserByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	})

	// 更新用户
	router.PUT("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		var updateData map[string]interface{}
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := userService.UpdateUser(c.Request.Context(), id, updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "用户更新成功"})
	})

	// 删除用户
	router.DELETE("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		if err := userService.DeleteUser(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "用户删除成功"})
	})
}

// 工具路由
func setupToolRoutes(router *gin.Engine, verificationService *verification.VerificationCodeService) {
	// 申请验证码（公开接口）
	router.POST("/verification/code", func(c *gin.Context) {
		var request struct {
			Name    string `json:"name" binding:"required"`
			Channel string `json:"channel" binding:"required,oneof=qq phone"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		// 生成验证码
		code, err := verificationService.GenerateVerificationCode(c.Request.Context(), request.Name, request.Channel)
		if err != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "操作过于频繁，请稍后再试",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "验证码已生成",
			"name":    request.Name,
			"channel": request.Channel,
			"code":    code, // 实际生产环境中应该不返回验证码，只返回成功消息
		})
	})

	// 验证验证码
	router.POST("/verification/verify", func(c *gin.Context) {
		var request struct {
			Name    string `json:"name" binding:"required"`
			Channel string `json:"channel" binding:"required,oneof=qq phone"`
			Code    string `json:"code" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		// 校验验证码
		isValid, err := verificationService.VerifyCode(c.Request.Context(), request.Name, request.Channel, request.Code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "验证过程出错",
				"valid":   false,
				"error":   err.Error(),
			})
			return
		}

		if isValid {
			// 验证成功，生成用户token
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			tokenService := db.NewTokenService()
			userToken, err := tokenService.GenerateUserToken(ctx, request.Name, request.Channel)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "生成token失败",
					"valid":   true,
					"error":   err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "验证成功",
				"valid":   true,
				"token":   userToken,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "验证码无效或已过期",
				"valid":   false,
			})
		}
	})

	// 创建需要鉴权的接口组
	authGroup := router.Group("")
	authGroup.Use(middleware.AuthMiddleware())
	{
		// 重新生成指定 id 的forward_view 的 title（需要鉴权）
		authGroup.GET("/rebuild_title/:id", func(c *gin.Context) {
			id := c.Param("id")
			napcat_go_sdk.Rebuild_title(id)
		})

		// 推送消息到QQ（需要鉴权）
		authGroup.GET("/push_to_qq/:id", func(c *gin.Context) {
			id := c.Param("id")
			napcat_go_sdk.PushMessageViewToQQ(id)
		})
	}
}
