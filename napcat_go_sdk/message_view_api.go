package napcat_go_sdk

import (
	"context"
	"memento_backend/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetMessageViewTitle 根据指定的message ID获取对应的title
func GetMessageViewTitle(id string) (string, error) {
	// 将字符串ID转换为ObjectID
	hexID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return "", err
	}

	// 从数据库查询消息
	collection := db.Collection("message_db", "forward_views")
	var message bson.M
	if err := collection.FindOne(context.Background(), bson.M{"_id": hexID}).Decode(&message); err != nil {
		return "", err
	}

	// 获取title字段
	title, ok := message["title"].(string)
	if !ok {
		title = "" // 字段不存在或类型不匹配时的默认值
	}

	return title, nil
}
