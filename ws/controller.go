package ws

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rus-sharafiev/go-rest-common/db"
	"github.com/rus-sharafiev/go-rest-common/exception"
	"github.com/rus-sharafiev/go-rest/auth"
)

type controller struct {
	db          *db.Postgres
	clientConns map[string]*websocket.Conn
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // for dev
}

func (c controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userId, _ := auth.Headers(r)
	if len(userId) == 0 {
		exception.Unauthorized(w)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		exception.InternalServerError(w, err)
		return
	}

	if clientConn := c.clientConns[userId]; clientConn != nil {
		clientConn.Close()
	}

	c.clientConns[userId] = conn

	c.sendUnreadMessages(conn, &userId)

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				fmt.Println(err)
				conn.Close()
				delete(c.clientConns, userId)
			}
			return
		}
		msg.From = &userId
		c.handleMessage(&msg, conn)
	}

}

var Controller = controller{db: &db.Instance, clientConns: make(map[string]*websocket.Conn)}
