package handler

// Handler defines the interface of a handler backend.
type Handler interface {
	Close() error                       // 断开mqtt连接
	SendDataUp(interface{}) error       // 发送上行数据
	SendSerDataUp([]byte) error         // 发送上行数据
	DataDownChan() chan DataDownPayload // 返回订阅到的消息数据channel
	IsConnected() bool
}
