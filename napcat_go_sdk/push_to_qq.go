package napcat_go_sdk

import (
	"context"
	"errors"
	"fmt"
	"log"
	"memento_backend/db"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"snail.local/snailllllll/utils"
)

func PushMessageViewToQQ(forward_id string) error {
	log.Printf("开始推送消息到QQ，forward_id: %s", forward_id)

	// 使用GetMessageViewTitle方法获取title
	title, err := GetMessageViewTitle(forward_id)
	if err != nil {
		log.Printf("获取消息标题失败: %v", err)
		return fmt.Errorf("获取消息标题失败: %w", err)
	}

	// 从关系表中获取原始消息
	collection := db.Collection("message_db", "message_relations")

	// 验证forward_id格式
	if _, err := primitive.ObjectIDFromHex(forward_id); err != nil {
		log.Printf("警告: forward_id格式无效: %s", forward_id)
	}

	// 确保使用字符串类型查询
	filter := bson.M{"view_record": forward_id}
	var relation bson.M

	log.Printf("执行MongoDB查询: 集合=message_relations, 条件=%+v", filter)

	// 设置查询超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 添加查询时间测量
	startTime := time.Now()
	err = collection.FindOne(ctx, filter).Decode(&relation)
	elapsed := time.Since(startTime)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Printf("未找到匹配文档: view_record=%s (查询耗时: %v)", forward_id, elapsed)

			// 添加额外诊断
			count, _ := collection.CountDocuments(ctx, bson.M{})
			log.Printf("诊断: 集合中共有%d个文档", count)

			// 检查是否存在类似文档
			partialFilter := bson.M{"view_record": bson.M{"$regex": primitive.Regex{Pattern: forward_id[:12], Options: ""}}}
			partialCount, _ := collection.CountDocuments(ctx, partialFilter)
			log.Printf("诊断: 部分匹配(%s...)的文档数: %d", forward_id[:12], partialCount)
		} else {
			log.Printf("查询失败: %v (查询耗时: %v)", err, elapsed)
		}
		return fmt.Errorf("查询消息关系失败: %w", err)
	} else {
		log.Printf("查询成功: 找到文档ID=%v (查询耗时: %v)", relation["_id"], elapsed)

		// 验证结果类型
		if viewRecord, ok := relation["view_record"].(string); ok {
			log.Printf("验证: view_record字段类型为string, 值=%s", viewRecord)
		} else {
			log.Printf("警告: view_record字段类型为%T, 值=%v", relation["view_record"], relation["view_record"])
		}
	}

	// 获取 message_record 的值
	messageRecord, ok := relation["message_record"].(string)
	if !ok {
		log.Println("message_record 字段不存在或类型错误")
		return errors.New("message_record 字段不存在或类型错误")
	}

	// 从 forward_messages 库中读取原始记录
	forwardCollection := db.Collection("message_db", "forward_messages")
	var result struct {
		Messages []ReceiveMessage `bson:"messages"`
		Count    int              `bson:"count"`
	}
	objID, err := primitive.ObjectIDFromHex(messageRecord)
	if err != nil {
		log.Printf("转换 ObjectID 失败: %v", err)
		return fmt.Errorf("转换 ObjectID 失败: %w", err)
	}

	err = forwardCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&result)
	if err != nil {
		log.Printf("查找原始消息失败: %v", err)
		return fmt.Errorf("查找原始消息失败: %w", err)
	}

	// 现在可以使用 result.Messages 和 result.Count
	prasemessages(result.Messages)
	log.Printf("获取到 %d 条原始消息", result.Count)

	Send_forward_message_to_group(result.Messages, title)
	return nil
	return nil
}

func prasemessages(result []ReceiveMessage) {
	var api_host, port, pic_host string
	api_host = utils.GetConfig("API_HOST", "")
	port = utils.GetConfig("PORT", "")
	pic_host = fmt.Sprintf("http://%s:%s/pic/", api_host, port)
	// 遍历ReceiveMessage，将图片消息的Url替换为翻旧账地址
	for i := range result {
		for j := range result[i].Message {
			msg := &result[i].Message[j]
			if msg.Type == "image" {
				msg.Data.Url = pic_host + msg.Data.File
			}
		}
	}
}
func Send_forward_message_to_group(result []ReceiveMessage, title string) {
	ws, _ := GetExistWSClient()
	group := utils.GetConfig("INFORM_GROUP", "")
	promt := "翻旧账推送"
	summary := "summary"
	source := title
	send_forward_message(result, group, promt, summary, source, ws)

}
func send_forward_message(result []ReceiveMessage, group, promt, summary, source string, ws *WebSocketClient) {
	msg := Message[any]{
		Action: "send_forward_msg",
		Params: ForwardMsgContent{
			UserGroupId: UserGroupId{GroupId: &group},
			Messages:    convertToNodeMsgList(result),
			News: []struct {
				Text *string `json:"text"`
			}{}, // 初始化为空切片
			Prompt:  promt,   // 新增字段
			Summary: summary, // 新增字段
			Source:  source,  // 新增字段
		},
	}
	ws.SendMessage(msg)
}

// 将ReceiveMessage转换为节点消息
func convertToNodeMsg(receiveMsg ReceiveMessage) Msg {
	// 将发送者用户ID转换为字符串指针
	userIdStr := strconv.Itoa(receiveMsg.Sender.UserId)

	// 创建MsgData结构
	msgData := MsgData{
		UserId:   &userIdStr,
		NickName: &receiveMsg.Sender.Nickname,
		Content:  convertMessageListToMsg(receiveMsg.Message),
	}

	return Msg{
		Type: "node",
		Data: msgData,
	}
}

// 辅助函数：将[]MessageList转换为[]Msg
func convertMessageListToMsg(messageList []MessageList) []Msg {
	var msgs []Msg
	for _, msgItem := range messageList {
		msg := Msg{
			Type: msgItem.Type,
			Data: MsgData{
				Text:     &msgItem.Data.Text,
				File:     &msgItem.Data.File,
				Url:      &msgItem.Data.Url,
				FileSize: &msgItem.Data.FileSize,
			},
		}
		//fmt.Println(msgItem.Data.Url)
		msgs = append(msgs, msg)
	}
	return msgs
}

// 将多个ReceiveMessage转换为节点消息列表
func convertToNodeMsgList(receiveMsgs []ReceiveMessage) []Msg {
	var msgList []Msg
	for _, msg := range receiveMsgs {
		msgList = append(msgList, convertToNodeMsg(msg))
	}
	return msgList
}
