package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rus-sharafiev/go-rest/common/db"
	"github.com/rus-sharafiev/go-rest/common/exception"
	"github.com/rus-sharafiev/go-rest/common/jwt"
	"github.com/rus-sharafiev/go-rest/common/localization"
	"golang.org/x/crypto/pbkdf2"
)

type verifySignup struct {
	db *db.Postgres
}

func (c verifySignup) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		exception.MethodNotAllowed(w)
		return
	}

	var signUpCode SignUpCode
	json.NewDecoder(r.Body).Decode(&signUpCode)

	// Get signup id from cookie
	signupIdCookie, err := r.Cookie("signup-id")
	if err != nil {
		exception.InternalServerError(w, err)
		return
	}

	// Get signup data from Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	signUpDataJson, err := rdb.Get(context.Background(), signupIdCookie.Value).Result()
	if err == redis.Nil {
		exception.BadRequestFields(w, map[string]string{
			"code": localization.SelectString(r, localization.Langs{
				En: "Verification code has expired",
				Ru: "Срок действия кода подтверждения истек",
			}),
		})
		return
	} else if err != nil {
		exception.InternalServerError(w, err)
		return
	}

	var signUpData SignUpData
	if err := json.Unmarshal([]byte(signUpDataJson), &signUpData); err != nil {
		exception.InternalServerError(w, err)
		return
	}

	go rdb.Del(context.Background(), signupIdCookie.Value)

	// Verify code
	if signUpCode.Code != signUpData.Code {
		exception.BadRequestFields(w, map[string]string{
			"code": localization.SelectString(r, localization.Langs{
				En: "incorrect verification code",
				Ru: "Не верный код подтверждения",
			}),
		})
		return
	}

	// Register new user
	salt := make([]byte, 16)
	rand.Read(salt)
	hash := pbkdf2.Key([]byte(signUpData.Password), salt, 4096, 32, sha1.New)

	var hashedPassword strings.Builder
	hashedPassword.WriteString(base64.StdEncoding.EncodeToString(hash))
	hashedPassword.WriteString(".")
	hashedPassword.WriteString(base64.StdEncoding.EncodeToString(salt))

	createUserQuery := `
		WITH u AS (
			INSERT INTO users ("email")
			VALUES ($1)
			RETURNING *
		), p AS (
			INSERT INTO passwords ("userId", "passwordHash")
			SELECT u."id", $2
			FROM u
		)
		SELECT * FROM u; 
	`
	rows, _ := c.db.Query(&createUserQuery, signUpData.Email, hashedPassword.String())
	userData, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[UserData])
	if err != nil {
		exception.InternalServerError(w, err)
		return
	}

	token, err := jwt.GenerateAccessToken(*userData.ID, *userData.Access)
	if err != nil {
		exception.InternalServerError(w, err)
		return
	}

	result := LoginResult{
		User:        userData,
		AccessToken: token,
	}

	// Set cookie with refresh token
	if err := jwt.SetRefreshToken(*userData.ID, *userData.Access, w); err != nil {
		exception.InternalServerError(w, err)
		return
	}

	// OK response
	json.NewEncoder(w).Encode(&result)
}

var VerifySignup = &verifySignup{db: db.NewConnection()}
