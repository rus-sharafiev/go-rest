package ws

import (
	"database/sql"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
	"github.com/rus-sharafiev/go-rest-common/exception"
)

func (c controller) handleMessage(msg *Message, senderConn *websocket.Conn) {

	query := `
		INSERT INTO messages ("from", "to", "msg")
		VALUES (@from, @to, @message)
		RETURNING "id", "createdAt";
	`
	args := pgx.NamedArgs{
		"from":    msg.From,
		"to":      msg.To,
		"message": msg.Message,
	}

	var msgId sql.NullString
	var createdAt sql.NullTime
	if err := c.db.QueryRow(&query, &args).Scan(&msgId, &createdAt); err != nil {
		exception.WsError(senderConn, err)
		return
	} else if msgId.Valid && createdAt.Valid {
		msg.ID = &msgId.String
		msg.CreatedAt = &createdAt.Time
	}

	if recipientConn := c.clientConns[*msg.To]; recipientConn != nil {
		if err := recipientConn.WriteJSON(msg); err != nil {
			exception.WsError(senderConn, err)
			return
		}

		senderConn.WriteJSON(MessageStatus{Status: "success", ID: msg.ID})
	}
}

func (c controller) sendUnreadMessages(conn *websocket.Conn, userId *string) {
	query := `
		SELECT jsonb_agg(row)
		FROM (
			SELECT *
			FROM messages
			WHERE "to" = $1 AND "readAt" IS NULL
		) row;
	`
	c.db.MessageJsonString(conn, &query, userId)
}
