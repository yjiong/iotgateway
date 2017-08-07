package handler

// DataUpPayload represents a data-up payload.
type Header struct {
	Devid		string		`json:"_devid"`
	Model		string		`json:"_model"`
	Version	string		`json:"_version"`
	Runstate	string		`json:"_runstate"`
}

type Request struct {
	Cmd				string				`json:"cmd"`
	Data			interface {}		`json:"data"`
	Timestamp		int64    			`json:"timestamp"`
}

type Response struct {
	Data			interface {}		`json:"data"`
	Statuscode		bool				`json:"statuscode"`
	Timestamp		int64    			`json:"timestamp"`
}

type DataupPayload struct {
	Header 		Header			`json:"header"`
	Msgtype		string			`json:"msgtype"`
	Request		Request		`json:"request"`
}

// DataDownPayload represents a data-down payload.
type DataDownPayload struct {
	Pj interface {}
}

