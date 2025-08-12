package napcat_go_sdk

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
