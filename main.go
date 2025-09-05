package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"memento_backend/db"

	"snail.local/snailllllll/verification"

	"github.com/gin-gonic/gin"
	"snail.local/snailllllll/napcat_go_sdk"
	"snail.local/snailllllll/routes"
	"snail.local/snailllllll/utils"
)

var (
	wsClientInstance *napcat_go_sdk.WebSocketClient
	once             sync.Once
	Admin_uin        string
)

func main() {
	// 初始化配置
	utils.LoadConfig()

	// 创建路由引擎
	router := gin.Default()

	// 初始化 WebSocket 客户端
	wsClientInstance, err := napcat_go_sdk.GetWebSocketClient(utils.Config.Addr, 3001, &utils.Config.Token)
	if err != nil {
		fmt.Println("Failed to initialize WebSocket client:", err)
		return
	}

	// 发送bot 登录成功提示
	text := utils.GetConfig("TEXT", "")
	napcat_go_sdk.SingleTextMessage(&text, &utils.Config.AdminUIN, wsClientInstance)

	// 连接 mongoDB
	err = db.Init(utils.Config.DBURI)
	if err != nil {
		db_connect_success := "MongoDB连接失败"
		napcat_go_sdk.SingleTextMessage(&db_connect_success, &utils.Config.AdminUIN, wsClientInstance)
		return
	}

	// 初始化用户服务并创建索引
	userService := db.NewUserService()
	ctx := context.Background()
	if err := userService.CreateIndexes(ctx); err != nil {
		fmt.Printf("创建用户索引失败: %v\n", err)
	}

	// 创建token服务并初始化索引
	tokenService := db.NewTokenService()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := tokenService.CreateIndexes(ctx); err != nil {
		fmt.Printf("创建token索引失败: %v\n", err)
	}

	// 初始化验证码服务并创建索引
	verificationService := verification.NewVerificationCodeService()
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := verificationService.CreateIndexes(ctx); err != nil {
		fmt.Printf("创建验证码索引失败: %v\n", err)
	}

	// 设置所有路由
	routes.SetupRoutes(router, userService, verificationService, wsClientInstance)

	// 处理历史数据
	napcat_go_sdk.ProcessEmptyTitleForwardViews() //处理没有title的forward_view,处理历史数据用

	// 启动服务
	router.Run()
}
