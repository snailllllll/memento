package napcat_go_sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"memento_backend/db"

	"snail.local/snailllllll/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ForwardMessages [][]ReceiveMessage
var ForwardMessagesView [][]MessageView

type ForwardResponse struct {
	Status  string `json:"status"`
	Retcode int    `json:"retcode"`
	Data    struct {
		Messages []ReceiveMessage `json:"messages"`
	} `json:"data"`
}

func (receiveMessage *ReceiveMessage) ToView() MessageView {
	return MessageView{
		Time:        receiveMessage.Time,
		MessageType: receiveMessage.MessageType,
		Sender:      receiveMessage.Sender,
		RawMessage:  receiveMessage.RawMessage,
	}
}
func (receiveMessage *ReceiveMessage) ISSenderBot() bool {
	// 过滤bot发送的信息和心跳包
	return receiveMessage.SelfId == receiveMessage.Sender.UserId || receiveMessage.MetaEventType == "heartbeat"
}
func (receiveMessage *ReceiveMessage) ParseMessage() {
	// 解析消息：存储图片和表情
	if receiveMessage.ISSenderBot() {
		return
	}

	//fmt.Printf("原始文本: %v\n", receiveMessage.RawMessage)
	// 遍历receiveMessage.Message
	//
	for _, msg := range receiveMessage.Message {
		fmt.Printf("消息类型: %v\n", msg.Type)
		if msg.Type == "image" {
			//fmt.Printf("消息内容: %v\n", msg.Data)
			// 获取图片真实url
			url := msg.Data.Url
			filename := msg.Data.File
			// 从 url 下载filename的图片
			err := utils.DownloadImageFromURL(url, filename)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
		// 处理合并转发消息
		// 解析转发消息中的图片
		if msg.Type == "forward" {
			msg_id := msg.Data.Id
			send_msg := Message[any]{
				Action: GET_FORWARD_MESSAGE,
				Params: MessageId{MessageId: msg_id},
			}
			response, _ := wsClientInstance.SendMessage(send_msg)
			var forward_response ForwardResponse
			err := json.Unmarshal([]byte(response), &forward_response)
			if err != nil {
				fmt.Printf("解析JSON失败: %v", err)
				continue
			}
			var forward_views []MessageView
			origin_message_record, _ := SaveReceiveMessagesToDB(forward_response.Data.Messages)
			for _, msg := range forward_response.Data.Messages {
				msg.ParseMessage()
				forward_views = append(forward_views, msg.ToView())
			}
			view_record, _ := SaveMessageViewsToDB(forward_views)
			// 保存消息记录发送人 
			InsertSender(view_record,receiveMessage.Sender.Nickname)
			// 保存消息和视图的关联关系
			collection := db.Collection("message_db", "message_relations")
			doc := map[string]interface{}{
				"message_record": origin_message_record,
				"view_record":    view_record,
				"created_at":     time.Now(),
			}
			_, err = collection.InsertOne(context.Background(), doc)
			if err != nil {
				log.Printf("保存消息关联关系失败: %v", err)
			}

			// 生成标题
			go ProcessForwardViewsToDB(view_record)
		}
	}

}

// 保存消息视图切片到数据库
func SaveMessageViewsToDB(messageViews []MessageView) (string, error) {
	collection := db.Collection("message_db", "forward_views")
	// 将整个切片作为单个文档插入
	doc := map[string]interface{}{
		"messages": messageViews,
		"count":    len(messageViews),
	}
	result, err := collection.InsertOne(context.Background(), doc)
	if err != nil {
		log.Printf("保存MessageViews失败: %v", err)
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

// 保存接收消息切片到数据库
func SaveReceiveMessagesToDB(messages []ReceiveMessage) (string, error) {
	collection := db.Collection("message_db", "forward_messages")
	// 将整个切片作为单个文档插入
	doc := map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	}
	result, err := collection.InsertOne(context.Background(), doc)
	if err != nil {
		log.Printf("保存ReceiveMessages失败: %v", err)
		return "", err
	}
	// 将ObjectID转换为字符串返回
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

// 辅助函数：将任意类型的切片转换为[]interface{}类型
// 用于MongoDB批量插入操作前的类型转换
func toInterfaceSlice[T any](slice []T) []interface{} {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		result[i] = v
	}
	return result
}

//向MessageViews注入发送人信息
func InsertSender(forward_id string, sender string) {
	// 获取forward_views数据
	collection := db.Collection("message_db", "forward_views")
	var fv ForwardView
	id, err := primitive.ObjectIDFromHex(forward_id)
	if err != nil {
		fmt.Printf("invalid forward ID: %v", err)
		return
	}

	if err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&fv); err != nil {
		fmt.Printf("forward view not found: %v", err)
		return
	}

	// 更新数据库中的sender
	update := bson.M{"$set": bson.M{"sender": sender}}
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": id}, update)
	if err != nil {
		fmt.Printf("failed to update forward view sender: %v", err)
		return
	}
	fmt.Printf("更新sender成功: %v", sender)
}

// 提取MessageViews的消息并转换为json，生成幽默标题并更新到数据库
func ProcessForwardViewsToDB(forward_id string) (string, error) {

	// 获取forward_views数据
	collection := db.Collection("message_db", "forward_views")
	var forwardView bson.M
	id, err := primitive.ObjectIDFromHex(forward_id)
	if err != nil {
		return "", fmt.Errorf("invalid forward ID")
	}

	if err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&forwardView); err != nil {
		return "", fmt.Errorf("forward view not found")
	}

	// 将forward_views转换为JSON并发送到API
	jsonData, err := json.Marshal(forwardView)
	if err != nil {
		return "", fmt.Errorf("failed to marshal forward view")
	}
    // TODO：发送请求到API，url 从配置文件中读取
	resp, err := http.Post("http://14.103.138.175:8888/api/qq-humor-title", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to send request to humor title API")
	}
	defer resp.Body.Close()

	// 解析API响应
	var result struct {
		Success bool   `json:"success"`
		Title   string `json:"title"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode API response")
	}

	if !result.Success {
		return "", fmt.Errorf("humor title API returned failure")
	}

	// 更新数据库中的title
	update := bson.M{"$set": bson.M{"title": result.Title}}
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": id}, update)
	if err != nil {
		return "", fmt.Errorf("failed to update forward view title")
	}

	senderStr := forwardView["sender"].(string)
	groupStr  := utils.GetConfig("INFORM_GROUP", "")
	NewMessageGroupInform(&result.Title, &senderStr, &groupStr,&forward_id)
	return result.Title, nil
}

// ForwardView 定义转发视图结构
type ForwardView struct {
	ID    primitive.ObjectID `bson:"_id"`
	Title string             `bson:"title"`
}

// ProcessEmptyTitleForwardViews 处理所有title为空的forward_views
func ProcessEmptyTitleForwardViews() error {
	// 初始化collection
	fmt.Printf("开始更新title=====================================================\n")

	collection := db.Collection("message_db", "forward_views")

	// 获取所有title为空的forward_views
	filter := bson.M{"$or": []bson.M{
		{"title": bson.M{"$exists": false}},
	}}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return fmt.Errorf("failed to find forward views with empty title")
	}
	fmt.Printf("Found %d forward views with empty title\n", cursor.RemainingBatchLength())
	defer cursor.Close(context.Background())

	// 遍历处理每个forward_view
	for cursor.Next(context.Background()) {
		var fv ForwardView
		if err := cursor.Decode(&fv); err != nil {
			return fmt.Errorf("failed to decode forward view")
		}
		fmt.Printf("Processing forward view %s\n", fv.ID.Hex())
		if _, err := ProcessForwardViewsToDB(fv.ID.Hex()); err != nil {
			return fmt.Errorf("failed to process forward view: %v", err)
		}
	}

	if err := cursor.Err(); err != nil {
		return fmt.Errorf("cursor error: %v", err)
	}

	return nil
}