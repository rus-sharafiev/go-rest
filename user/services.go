package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/rus-sharafiev/go-rest/common/auth"
	"github.com/rus-sharafiev/go-rest/common/exception"
)

// -- CREATE ----------------------------------------------------------------------

func (c *controller) create(w http.ResponseWriter, r *http.Request) {
	if _, role := auth.Headers(r); role != "ADMIN" {
		exception.Forbidden(w)
		return
	}

	var payload CreateDto
	json.NewDecoder(r.Body).Decode(&payload)

	query := `
		WITH u AS (
			INSERT INTO users ("email", "firstName")
			VALUES (@email, @firstName)
			RETURNING "id"
		)
		INSERT INTO passwords ("userId", "passwordHash")
		SELECT u."id", @hash
		FROM u;
	`
	args := pgx.NamedArgs{
		"email":     payload.Email,
		"firstName": payload.FirstName,
		"hash":      payload.PasswordHash,
	}

	c.db.WriteJsonString(w, &query, args)
}

// -- FIND ALL --------------------------------------------------------------------

func (c *controller) findAll(w http.ResponseWriter, r *http.Request) {
	if _, role := auth.Headers(r); role != "ADMIN" {
		exception.Forbidden(w)
		return
	}

	query := `
		SELECT jsonb_agg(row)
		FROM (
			SELECT *
			FROM users
			GROUP BY id
			ORDER BY id
		) row;
	`
	c.db.WriteJsonString(w, &query)
}

// -- FIND ONE --------------------------------------------------------------------

func (c *controller) findOne(id string, w http.ResponseWriter, r *http.Request) {
	if userId, role := auth.Headers(r); role != "ADMIN" || userId != id {
		if len(userId) == 0 {
			exception.Unauthorized(w)
			return
		}

		exception.Forbidden(w)
		return
	}

	query := `
		SELECT row_to_json(row)
		FROM (
			SELECT *
			FROM users
			WHERE "id" = $1
		) row; 
	`
	c.db.WriteJsonString(w, &query, id)
}

// -- UPDATE ----------------------------------------------------------------------

func (c *controller) update(id string, w http.ResponseWriter, r *http.Request) {
	if userId, role := auth.Headers(r); role != "ADMIN" || userId != id {
		if len(userId) == 0 {
			exception.Unauthorized(w)
			return
		}

		exception.Forbidden(w)
		return
	}

	var payload UpdateDto
	json.NewDecoder(r.Body).Decode(&payload)

	var currentAvatar sql.NullString
	avatarQuery := "SELECT avatar FROM users WHERE id = $1;"
	c.db.QueryRow(&avatarQuery, id).Scan(&currentAvatar)

	if currentAvatar.Valid && currentAvatar.String != *payload.Avatar {
		if err := os.Remove(currentAvatar.String); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				fmt.Println(err)
			}
		}
	}

	query := `
		WITH u AS (
			UPDATE users SET ("firstName", "lastName", "phone", "avatar", "updatedAt") = 
				(@firstName, @lastName, @phone, @avatar, CURRENT_TIMESTAMP)
			WHERE "id" = @id
			RETURNING *
		)
		SELECT row_to_json(row)
		FROM (SELECT * FROM u) row;
	`
	args := pgx.NamedArgs{
		"id":        id,
		"firstName": payload.FirstName,
		"lastName":  payload.LastName,
		"phone":     payload.Phone,
		"avatar":    payload.Avatar,
	}

	c.db.WriteJsonString(w, &query, args)
}

// -- DELETE ----------------------------------------------------------------------

func (c *controller) delete(id string, w http.ResponseWriter, r *http.Request) {
	if _, role := auth.Headers(r); role != "ADMIN" {
		exception.Forbidden(w)
		return
	}

	query := `
		WITH u AS (
			DELETE FROM users
			WHERE id = $1
			RETURNING *
		)
		SELECT row_to_json(row)
		FROM (SELECT * FROM u) row;
	`
	c.db.WriteJsonString(w, &query, id)
}
