package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

// Init 初始化MongoDB连接
func Init(uri string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("MongoDB连接失败: %v", err)
	}

	// 验证连接是否成功
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("MongoDB连接验证失败: %v", err)
	}

	Client = client
	fmt.Println("成功连接到MongoDB!")
	return nil
}

// Collection 获取MongoDB集合处理器
func Collection(dbName, colName string) *mongo.Collection {
	return Client.Database(dbName).Collection(colName)
}
