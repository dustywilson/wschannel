package wschannel

import (
	"code.google.com/p/go.net/websocket"
	"strings"
	"time"
)

type Service struct {
	enabled          bool
	statusChan       chan bool
	shutdownComplete chan bool
	bulkCloser       chan bool
	sessions         map[string]*Session
	sessionAdd       chan *Session
	sessionDel       chan *Session
}

func NewService() *Service {
	sv := new(Service)
	sv.enabled = true
	sv.statusChan = make(chan bool)
	sv.shutdownComplete = make(chan bool)
	sv.bulkCloser = make(chan bool)
	sv.sessions = make(map[string]*Session)
	sv.sessionAdd = make(chan *Session, 100)
	sv.sessionDel = make(chan *Session, 100)
	go sv.loop()
	return sv
}

func (sv *Service) loop() {
	for {
		select {
		case state := <-sv.statusChan:
			sv.enabled = state
			if !state {
				sv.shutdown()
			}
		case ss := <-sv.sessionAdd:
			sv.sessions[ss.id] = ss
			go ss.loop()
		case ss := <-sv.sessionDel:
			delete(sv.sessions, ss.id)
		}
	}
}

func (sv *Service) Handler(pathPrefix string) websocket.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		sv.handler(pathPrefix, ws)
	})
}

func (sv *Service) handler(pathPrefix string, ws *websocket.Conn) {
	if !sv.enabled {
		websocket.JSON.Send(ws, MessageServiceTemporarilyOffline)
		return
	}
	pathParts := strings.Split(strings.Replace(ws.Config().Location.Path, pathPrefix+"/", "", 1), "/")
	if len(pathParts) < 2 {
		return
	}
	sessionId := pathParts[0]
	connectionId := pathParts[1]
	ss := sv.GetSession(sessionId)
	if ss == nil {
		return
	}
	cn := ss.GetConnection(connectionId)
	if cn != nil {
		cn.reconnected <- true
		websocket.JSON.Send(ws, NewSettingMessage("connection", "resumed"))
	} else {
		cn = ss.NewConnection(connectionId)
		websocket.JSON.Send(ws, NewSettingMessage("connection", "new"))
	}
	websocket.JSON.Send(ws, NewSettingMessage("defaultReconnectDelay", time.Duration(time.Second*5).Nanoseconds()))
	cn.ws = ws
	cn.run()
}

func (sv *Service) Shutdown() chan bool {
	if sv.enabled {
		sv.statusChan <- false
		return sv.shutdownComplete
	}
	return nil
}

func (sv *Service) shutdown() {
	close(sv.bulkCloser)
	close(sv.shutdownComplete)
}

func (sv *Service) GetSessions() []string {
	sessions := make([]string, len(sv.sessions))
	i := 0
	for id, _ := range sv.sessions {
		sessions[i] = id
		i++
	}
	return sessions
}

func (sv *Service) GetSession(id string) *Session {
	if ss, ok := sv.sessions[id]; ok {
		return ss
	}
	return nil
}

func (sv *Service) NewSession(id string) *Session {
	ss := new(Session)
	ss.init(id)
	ss.sv = sv
	sv.sessionAdd <- ss
	return ss
}
