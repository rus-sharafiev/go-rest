package chat

import (
	"database/sql"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
	"github.com/rus-sharafiev/go-rest-common/exception"
	"github.com/rus-sharafiev/go-rest/auth"
)

// -- UPGRADE TO WEBSOCKET CONNECTION ---------------------------------------------

func (c controller) upgradeToWebsocket(w http.ResponseWriter, r *http.Request) {
	userIdStr, _ := auth.Headers(r)
	if len(userIdStr) == 0 {
		exception.Unauthorized(w)
		return
	}
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		exception.InternalServerError(w, err)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		exception.InternalServerError(w, err)
		return
	}

	conn.SetCloseHandler(func(code int, text string) error {
		if err := conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(code, text),
		); err != nil {
			return err
		}

		fmt.Printf("User %v, close ws connection with code: %v\n", userId, code)
		delete(c.clientWsConns, userId)
		return nil
	})

	if clientConn := c.clientWsConns[userId]; clientConn != nil {
		clientConn.Close()
	}

	c.clientWsConns[userId] = conn
	fmt.Printf("User %v has opened ws connection\n", userId)

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				conn.Close()
			}
			return
		}

		query := `
			WITH chat AS (
				INSERT INTO chats (users)
				VALUES (@users)
				ON CONFLICT (users) DO 
				UPDATE SET "updatedAt" = CURRENT_TIMESTAMP
				RETURNING *
			)
			INSERT INTO messages ("chatId", "from", "to", "message")
			VALUES ((SELECT id FROM chat), @from, @to, @message)
			RETURNING "id", "createdAt";
		`
		usrsArr := []int{userId, *msg.To}
		slices.Sort(usrsArr)
		args := pgx.NamedArgs{
			"from":    &userId,
			"to":      msg.To,
			"message": msg.Message,
			"users":   &usrsArr,
		}

		var msgId sql.NullString
		var msgCreatedAt sql.NullTime
		if err := c.db.QueryRow(&query, &args).Scan(&msgId, &msgCreatedAt); err != nil {
			exception.WsError(conn, err)
			return
		} else if msgId.Valid && msgCreatedAt.Valid {
			msg.ID = &msgId.String
			msg.CreatedAt = &msgCreatedAt.Time
		}
		conn.WriteJSON(MessageStatus{Status: "sent", ID: msg.ID, Time: msg.CreatedAt})

		if recipientConn := c.clientWsConns[*msg.To]; recipientConn != nil {
			query := `
				WITH m AS (
					UPDATE messages SET "deliveredAt" = CURRENT_TIMESTAMP
					WHERE "id" = $1
					RETURNING *
				)
				SELECT row_to_json(row)
				FROM (
					SELECT m.*, row_to_json(s) AS from, row_to_json(r) AS to
					FROM m

						LEFT JOIN (
							SELECT id, "email", "firstName", "lastName", "avatar" 
							FROM users
						) s
						ON s.id = m.from

						LEFT JOIN (
							SELECT id, "email", "firstName", "lastName", "avatar" 
							FROM users
						) r
						ON r.id = m.to
				) row;
			`
			deliveryTime := time.Now()
			if err := c.db.MessageJsonString(recipientConn, &query, msg.ID); err != nil {
				if websocket.IsUnexpectedCloseError(err) {
					fmt.Printf("error: %v\nws/services.go on line 62", err)
					conn.WriteJSON(MessageStatus{Status: "error", ID: msg.ID, Time: &deliveryTime})
					return
				}
			}

			conn.WriteJSON(MessageStatus{Status: "delivered", ID: msg.ID, Time: &deliveryTime})
		}
	}
}

// -- GET ALL CHATS ---------------------------------------------------------------

func (c controller) findAllChats(w http.ResponseWriter, r *http.Request) {
	userId, _ := auth.Headers(r)
	if len(userId) == 0 {
		exception.Unauthorized(w)
		return
	}

	limit := 30
	offset := 0

	if perPageStr := r.URL.Query().Get("per_page"); len(perPageStr) != 0 {
		if perPage, err := strconv.Atoi(perPageStr); err == nil {
			limit = perPage
		}
	}

	if pageStr := r.URL.Query().Get("page"); len(pageStr) != 0 {
		if page, err := strconv.Atoi(pageStr); err == nil {
			offset = limit * page
		}
	}

	query := `
		SELECT jsonb_agg(rows)
		FROM (
			SELECT *
			FROM chats
    		WHERE @userId = ANY (users)
			GROUP BY id
			ORDER BY "updatedAt" DESC
			LIMIT @limit
			OFFSET @offset
		) rows;
	`
	args := pgx.NamedArgs{
		"userId": &userId,
		"limit":  &limit,
		"offset": &offset,
	}
	c.db.WriteJsonString(w, &query, &args)
}

// -- GET ONE CHAT ----------------------------------------------------------------

func (c controller) findOneChat(chatId string, w http.ResponseWriter, r *http.Request) {
	userId, _ := auth.Headers(r)
	if len(userId) == 0 {
		exception.Unauthorized(w)
		return
	}

	query := `
		SELECT row_to_json(row)
		FROM (
			SELECT *
			FROM chats
    		WHERE id = @chatId AND @userId = ANY (users)
		) row;
	`
	args := pgx.NamedArgs{
		"userId": &userId,
		"chatId": &chatId,
	}
	c.db.WriteJsonString(w, &query, &args)
}

// -- DELETE CHAT -----------------------------------------------------------------

func (c controller) deleteChat(chatId string, w http.ResponseWriter, r *http.Request) {
	userId, _ := auth.Headers(r)
	if len(userId) == 0 {
		exception.Unauthorized(w)
		return
	}

	query := `		 
		WITH del AS (
			DELETE FROM chats
			WHERE id = @chatId AND @userId = ANY (users)
			RETURNING *
		)
		SELECT row_to_json(row)
		FROM (SELECT * FROM del) row;
	`
	args := pgx.NamedArgs{
		"userId": &userId,
		"chatId": &chatId,
	}
	c.db.WriteJsonString(w, &query, &args)
}

// -- GET ALL MESSAGES IN CHAT ----------------------------------------------------

func (c controller) findAllMessages(w http.ResponseWriter, r *http.Request, chatId string) {
	userId, _ := auth.Headers(r)
	if len(userId) == 0 {
		exception.Unauthorized(w)
		return
	}

	query := `
		SELECT jsonb_agg(rows)
		FROM (
			SELECT m.*, row_to_json(s) AS from, row_to_json(r) AS to
			FROM messages m

				LEFT JOIN (
					SELECT id, "email", "firstName", "lastName", "avatar" 
					FROM users
				) s
				ON s.id = m.from

				LEFT JOIN (
					SELECT id, "email", "firstName", "lastName", "avatar" 
					FROM users
				) r
				ON r.id = m.to

			WHERE m."chatId" = $1
			GROUP BY m.id, s.*, r.*
			ORDER BY m."createdAt" DESC
			LIMIT 30
			OFFSET 0
		) rows;
	`
	c.db.WriteJsonString(w, &query, chatId)
}
