package wschannel

import (
	"code.google.com/p/go.net/websocket"
	"io"
	"log"
	"time"
)

type Connection struct {
	ss           *Session
	id           string
	kill         chan bool
	killed       bool
	C            chan interface{}
	disconnected chan bool
	reconnected  chan bool
	ws           *websocket.Conn
}

func (cn *Connection) init(id string) {
	cn.id = id
	cn.kill = make(chan bool)
	cn.C = make(chan interface{}, 100)
	cn.disconnected = make(chan bool)
	cn.reconnected = make(chan bool)
}

func (cn *Connection) loop() {
	for {
		select {
		case <-cn.ss.sv.bulkCloser:
			websocket.JSON.Send(cn.ws, MessageServiceRestarting)
			return
		case <-cn.ss.kill:
			websocket.JSON.Send(cn.ws, MessageSessionEnded)
			cn.Kill()
			return
		case <-cn.kill:
			return
		case <-cn.reconnected:
			continue
		case <-cn.disconnected:
			killTimer := time.NewTimer(time.Minute * 5) // TODO: should probably be configurable
		Disconnected:
			for {
				select {
				case <-cn.ss.sv.bulkCloser:
					websocket.JSON.Send(cn.ws, MessageServiceRestarting)
					return
				case <-cn.ss.kill:
					websocket.JSON.Send(cn.ws, MessageSessionEnded)
					cn.Kill()
					return
				case <-cn.kill:
					return
				case <-cn.disconnected:
					continue Disconnected
				case <-cn.reconnected:
					break Disconnected
				case <-killTimer.C:
					return
				}
			}
		case message := <-cn.C:
			err := websocket.JSON.Send(cn.ws, message)
			if err != nil {
				cn.C <- message
				cn.disconnected <- true
			}
		}
	}
}

func (cn *Connection) run() {
	go cn.readLoop()
	cn.loop()
	cn.Kill()
	cn.ss.connectionDel <- cn
}

func (cn *Connection) readLoop() {
	zot := make([]byte, 1)
	for {
		// we only read so we can test for EOF
		n, err := cn.ws.Read(zot)
		if n == 0 || err == io.EOF || err == io.ErrUnexpectedEOF {
			if cn.ws == nil {
				return
			} else {
				cn.disconnected <- true
				return
			}
			return
		} else if err != nil {
			log.Printf("wschannel cn.readLoop Error: %s\n", err)
		}
	}
}

func (cn *Connection) Id() string {
	return cn.id
}

func (cn *Connection) Kill() {
	if !cn.killed {
		cn.killed = true
		close(cn.kill)
		cn.ss.connectionDel <- cn
		cn.ws.Close()
		cn.ws = nil
	}
}

func (cn *Connection) Send(message interface{}) {
	cn.C <- message
}
