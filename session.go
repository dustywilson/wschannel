package wschannel

type Session struct {
	sv            *Service
	id            string
	kill          chan bool
	C             chan interface{}
	connections   map[string]*Connection
	connectionAdd chan *Connection
	connectionDel chan *Connection
}

func (ss *Session) init(id string) {
	ss.id = id
	ss.kill = make(chan bool)
	ss.C = make(chan interface{}, 100)
	ss.connections = make(map[string]*Connection)
	ss.connectionAdd = make(chan *Connection, 100)
	ss.connectionDel = make(chan *Connection, 100)
}

func (ss *Session) loop() {
	for {
		select {
		case <-ss.kill:
			ss.sv.sessionDel <- ss
			for _, cn := range ss.connections {
				go func(ss *Session, cn *Connection) {
					delete(ss.connections, cn.id)
					cn.Kill()
				}(ss, cn)
			}
			return
		case cn := <-ss.connectionAdd:
			ss.connections[cn.id] = cn
		case cn := <-ss.connectionDel:
			delete(ss.connections, cn.id)
		case message := <-ss.C:
			for _, cn := range ss.connections {
				if len(cn.C) < cap(cn.C) {
					cn.C <- message
				}
			}
		}
	}
}

func (ss *Session) Id() string {
	return ss.id
}

func (ss *Session) Kill() {
	close(ss.kill)
}

func (ss *Session) Send(message interface{}) {
	ss.C <- message
}

func (ss *Session) GetConnections() []string {
	connections := make([]string, len(ss.connections))
	i := 0
	for id, _ := range ss.connections {
		connections[i] = id
		i++
	}
	return connections
}

func (ss *Session) GetConnection(id string) *Connection {
	if cn, ok := ss.connections[id]; ok {
		return cn
	}
	return nil
}

func (ss *Session) NewConnection(id string) *Connection {
	cn := new(Connection)
	cn.init(id)
	cn.ss = ss
	ss.connectionAdd <- cn
	return cn
}
