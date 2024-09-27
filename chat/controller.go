package chat

import (
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/rus-sharafiev/go-rest-common/db"
	"github.com/rus-sharafiev/go-rest-common/exception"
)

type controller struct {
	db            *db.Postgres
	clientWsConns map[int]*websocket.Conn
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // for dev
}

func (c controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	switch parts := strings.Split(path, "/")[2:]; len(parts) {

	// chats/
	case 0:
		switch r.Method {
		case http.MethodGet:
			c.findAllChats(w, r)
		default:
			exception.MethodNotAllowed(w)
		}

	// chats/{chatId}/
	case 1:
		id := parts[0]
		if id == "ws" {

			c.upgradeToWebsocket(w, r)

		} else {

			switch r.Method {
			case http.MethodGet:
				c.findOneChat(id, w, r)
			// case http.MethodPatch:
			// 	c.update(id, w, r)
			case http.MethodDelete:
				c.deleteChat(id, w, r)
			default:
				exception.MethodNotAllowed(w)
			}
		}

	// chats/{chatId}/messages/
	case 2:
		if parts[1] == "messages" {
			id := parts[0]
			switch r.Method {
			case http.MethodGet:
				c.findAllMessages(w, r, id)
			default:
				exception.MethodNotAllowed(w)
			}
		} else {
			http.NotFound(w, r)
		}

	// chats/{chatId}/messages/{messageId}
	case 3:
		if parts[1] == "messages" {
			id := parts[3]
			switch r.Method {
			case http.MethodGet:
				c.findAllMessages(w, r, id)
			// case http.MethodPatch:
			// 	c.update(id, w, r)
			// case http.MethodDelete:
			// 	c.delete(id, w, r)
			default:
				exception.MethodNotAllowed(w)
			}
		} else {
			http.NotFound(w, r)
		}

	default:
		http.NotFound(w, r)
	}

}

var Controller = controller{db: &db.Instance, clientWsConns: make(map[int]*websocket.Conn)}
