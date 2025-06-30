package main

import (
	"context"
	"fmt"
	"memento_backend/db"
	"net/http" // 引入 net/http 标准库，用于 HTTP 状态码常量
	"os"
	"path/filepath"
	"sync"

	"github.com/snailllllll/napcat_go_sdk"

	"github.com/gin-gonic/gin"
	"github.com/snailllllll/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	wsClientInstance *napcat_go_sdk.WebSocketClient
	once             sync.Once
	Admin_uin        string
)

func main() {
	// 初始化路由
	utils.LoadConfig()
	router := gin.Default()

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

	token := utils.GetConfig("TOKEN", "")
	Admin_uin = utils.GetConfig("ADMIN_UIN", "")
	text := utils.GetConfig("TEXT", "")
	addr := utils.GetConfig("ADDR", "")
	db_uri := utils.GetConfig("DB_URI", "")
	// 初始化 WebSocket 客户端
	wsClientInstance, err := napcat_go_sdk.GetWebSocketClient(addr, 3001, &token)
	if err != nil {
		fmt.Println("Failed to initialize WebSocket client:", err)
		return
	}
	// 发送bot 登录成功提示
	napcat_go_sdk.SingleTextMessage(&text, &Admin_uin, wsClientInstance)

	// 连接 mongoDB
	err = db.Init(db_uri)
	if err != nil {
		db_connect_success := "MongoDB连接失败"
		napcat_go_sdk.SingleTextMessage(&db_connect_success, &Admin_uin, wsClientInstance)

		return
	}
	

	// 获取具体消息
	router.GET("/messages/:id", func(c *gin.Context) {
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

	// 路由处理函数
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
	// 消息列表接口
	router.GET("/message_list", func(c *gin.Context) {
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
	// 获取全部消息 deprecated
	router.GET("/messages", func(c *gin.Context) {
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
	//重新生成指定 id 的forward_view 的 title 
	router.GET("/rebuild_title/:id", func(c *gin.Context) {
		id := c.Param("id")
		napcat_go_sdk.ProcessForwardViewsToDB(id)
		})
		
	// 启动服务
	napcat_go_sdk.ProcessEmptyTitleForwardViews() //处理没有title的forward_view,处理历史数据用
	router.Run()

}
