package fosp

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
	"sync"
)

type server struct {
	database    *database
	connections map[string][]*connection
	connsLock   sync.Mutex
	domain      string
}

func NewServer(dbDriver DatabaseDriver, domain string) *server {
	if dbDriver == nil {
		panic("Cannot initialize server without database")
	}
	s := new(server)
	s.database = NewDatabase(dbDriver, s)
	s.domain = domain
	s.connections = make(map[string][]*connection)
	return s
}

func (s *server) RequestHandler(res http.ResponseWriter, req *http.Request) {
	ws, err := websocket.Upgrade(res, req, nil, 1024, 104)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(res, "Not a WebSocket handshake", 400)
		return
	} else if err != nil {
		log.Println("Error while setting up WebSocket connection :: ", err)
		return
	}
	log.Println("Successfully accepted new connection")
	NewConnection(ws, s)
}

func (s *server) registerConnection(c *connection, remote string) {
	s.connsLock.Lock()
	s.connections[remote] = append(s.connections[remote], c)
	s.connsLock.Unlock()
}

func (s *server) Unregister(c *connection, remote string) {
	s.connsLock.Lock()
	for i, v := range s.connections[remote] {
		if v == c {
			s.connections[remote] = append(s.connections[remote][:i], s.connections[remote][i+1:]...)
			break
		}
	}
	s.connsLock.Unlock()
}

func (s *server) routeNotification(user string, notf *Notification) {
	//log.Printf("Sending notification %v to user %s", notf, user)
	if strings.HasSuffix(user, "@"+s.domain) {
		user_name := strings.TrimSuffix(user, s.domain)
		//log.Printf("Is local user %s", user_name)
		//log.Printf("Connections are %v", s.connections[user_name])
		for _, connection := range s.connections[user_name] {
			//log.Printf("Sending notification on local connection")
			connection.send(notf)
		}
	} else if notf.url.Domain() == s.domain {
		parts := strings.Split(user, "@")
		if len(parts) != 2 {
			panic(user + " is not a valid user identifier")
		}
		remote_domain := parts[1]
		//log.Printf("Is local notification that will be routed to remote server")
		remote_connection, err := s.getOrOpenRemoteConnection(remote_domain)
		if err == nil {
			notf.SetHead("User", user)
			remote_connection.send(notf)
		}
	}
}

func (s *server) forwardRequest(user string, rt RequestType, url *Url, headers map[string]string, body string) (*Response, error) {
	remote_domain := url.Domain()
	headers["User"] = user
	remote_connection, err := s.getOrOpenRemoteConnection(remote_domain)
	if err != nil {
		return nil, err
	}
	resp, err := remote_connection.SendRequest(rt, url, headers, body)
	log.Println("Recieved response from forwarded request")
	if err != nil {
		log.Println("Error occured while forwarding " + err.Error())
		return nil, err
	} else {
		resp.DeleteHead("User")
		return resp, nil
	}
}

func (s *server) getOrOpenRemoteConnection(remote_domain string) (*connection, error) {
	if connections, ok := s.connections["@"+remote_domain]; ok {
		for _, connection := range connections {
			return connection, nil
		}
	}
	return OpenConnection(s, remote_domain)
}

func (s *server) Domain() string {
	if s.domain == "" {
		return "localhost.localdomain"
	} else {
		return s.domain
	}
}
