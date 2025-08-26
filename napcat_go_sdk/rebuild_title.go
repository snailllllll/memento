package napcat_go_sdk

import (
	"context"
	"memento_backend/db"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"snail.local/snailllllll/utils"
)

func Rebuild_title(id string){
	lockKey := "rebuild_title_" + id
	// 从db.Collection("message_db", "forward_views")查询消息 title
	hex_id,_ := primitive.ObjectIDFromHex(id)
	collection := db.Collection("message_db", "forward_views")
	var message bson.M
	if err := collection.FindOne(context.Background(), bson.M{"_id": hex_id}).Decode(&message); err != nil {
		return
	}
	title, ok := message["title"].(string)
	if !ok {
		title = "" // 字段不存在或类型不匹配时的默认值
	}

	if utils.LockExists(lockKey) {
		RebuildTitlelock(&title, &utils.Config.InformGroup)

		return			
	} else {
		utils.SetLock(lockKey, 300*time.Second)
		defer utils.DeleteLock(lockKey)
		RebuildTitleInform(&title, &utils.Config.InformGroup)
		ProcessForwardViewsToDB(id)
}
}