package handler

// Handler defines the interface of a handler backend.
type Handler interface {
	Close() error                                          // closes the handler
	SendDataUp(interface {}) error                // send data-up payload
	DataDownChan() chan DataDownPayload                    // returns DataDownPayload channel
}
