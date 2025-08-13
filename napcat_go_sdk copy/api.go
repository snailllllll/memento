package napcat_go_sdk
import "fmt"
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

func NewMessageGroupInform(title *string,nickname *string,group *string){
	ws,_:=GetExistWSClient()
	text := fmt.Sprintf("【翻旧账】用户【%s】推送了新的聊天记录到翻旧账平台,系统已接入 Ai 自动命名为【%s】", *nickname, *title)
	SingleGroupMessage(&text, group, ws)
}