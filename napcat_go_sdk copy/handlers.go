package napcat_go_sdk



type BaseHandler struct{}

func (h *BaseHandler) HandleMessage(receiveMessage *ReceiveMessage) {
	go receiveMessage.ParseMessage()
}
