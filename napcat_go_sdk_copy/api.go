package napcat_go_sdk
import "fmt"
import 	"snail.local/snailllllll/utils"


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

func NewMessageGroupInform(title *string,nickname *string,group *string,id *string){
	var api_host,port,rebuild_title_host string
	api_host = utils.GetConfig("API_HOST","")
	port = utils.GetConfig("PORT","")
	rebuild_title_host=fmt.Sprintf("http://%s:%s/rebuild_title/",api_host,port)
	ws,_:=GetExistWSClient()
	text := fmt.Sprintf("【翻旧账】用户【%s】推送了新的聊天记录到翻旧账平台,系统已接入 Ai 自动命名为【%s】。  可以点击以下链接快速发起标题重建%s%s", *nickname, *title,	rebuild_title_host,*id)
	SingleGroupMessage(&text, group, ws)
}

func RebuildTitleInform(title *string,group *string){
	ws,_:=GetExistWSClient()
	text := fmt.Sprintf("【翻旧账】发起了【%s】对话的重命名", *title, )
	SingleGroupMessage(&text, group, ws)
}
func RebuildTitlelock(title *string,group *string){
	ws,_:=GetExistWSClient()
	text := fmt.Sprintf("【翻旧账】【%s】对话正在生成新命名,请耐心等待！", *title, )
	SingleGroupMessage(&text, group, ws)
}

