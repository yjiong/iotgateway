package devservice

import (
	"fmt"
	"net"
	"time"
)

// MsgUp ...
type MsgUp interface {
	// Handle the message and optionally return a response message.
	EncodeAutoup(data map[string]interface{}) error
}

// PackHandler ..
type PackHandler interface {
	PackageParser([]byte) ([]byte, error)
}

// ListenAndServe binds to the given address and serve requests forever.
func ListenAndServe(n, addr string, mu MsgUp) error {
	uaddr, err := net.ResolveUDPAddr(n, addr)
	if err != nil {
		return err
	}

	l, err := net.ListenUDP(n, uaddr)
	if err != nil {
		return err
	}
	return Server(l, mu)
}

// Server processes incoming UDP packets on the given listener, and processes
// these requests forever  until the listener is closed
func Server(listener *net.UDPConn, mu MsgUp) error {
	buf := make([]byte, 30)
	for {
		//nr, addr, err := listener.ReadFromUDP(buf)
		nr, _, err := listener.ReadFromUDP(buf)
		if err != nil {
			if neterr, ok := err.(net.Error); ok && (neterr.Temporary() || neterr.Timeout()) {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			return err
		}
		tmp := make([]byte, nr)
		copy(tmp, buf)
		go handlePacket(tmp, mu, gethp())
	}
}

func gethp() (hp PackHandler) {
	return hp
}

func handlePacket(data []byte, mu MsgUp, hp PackHandler) {
	mu.EncodeAutoup(map[string]interface{}{
		"receiveAleanDate": fmt.Sprintf("% x", data),
	})
}
