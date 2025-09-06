package napcat_go_sdk

import (
	"fmt"
)

type SendMsg interface {
	SendWebSocketMsg() interface{}
	SendHttpMsg() interface{}
}

type ReceiveResponseMessage interface {
	ReceiveResponseMessage(bytes []byte) error
}

func SingleTextMessage(text *string, user *string, ws *WebSocketClient) {
	msg := Message[any]{
		Action: "send_private_msg",
		Params: SendMsgContent{
			UserGroupId: UserGroupId{UserId: user},
			Messages: []Msg{
				{
					Type: "text",
					Data: MsgData{Text: text},
				},
			},
		},
	}
	ws.SendMessage(msg)
}

func SingleGroupMessage(text *string, group *string, ws *WebSocketClient) {
	msg := Message[any]{
		Action: "send_group_msg",
		Params: SendMsgContent{
			UserGroupId: UserGroupId{GroupId: group},
			Messages: []Msg{
				{
					Type: "text",
					Data: MsgData{Text: text},
				},
			},
		},
	}
	ws.SendMessage(msg)

}

func NewMessageGroupInform(title *string, nickname *string, group *string, id *string) {
	// var api_host, port, rebuild_title_host string
	// api_host = utils.GetConfig("API_HOST", "")
	// port = utils.GetConfig("PORT", "")
	// rebuild_title_host = fmt.Sprintf("http://%s:%s/rebuild_title/", api_host, port)
	ws, _ := GetExistWSClient()
	text := fmt.Sprintf("【翻旧账】用户【%s】向翻旧账推送名为【%s】的聊天记录。 ", *nickname, *title)
	SingleGroupMessage(&text, group, ws)
}

func RebuildTitleInform(title *string, group *string, username *string) {
	ws, _ := GetExistWSClient()
	text := fmt.Sprintf("【翻旧账】用户【%s】发起了【%s】对话的重命名", *username, *title)
	SingleGroupMessage(&text, group, ws)
}

func PushQQInform(title *string, group *string, username *string) {
	ws, _ := GetExistWSClient()
	text := fmt.Sprintf("【翻旧账】用户【%s】向您推送了【%s】对话", *username, *title)
	SingleGroupMessage(&text, group, ws)
}
