package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rus-sharafiev/go-rest-common/db"
	"github.com/rus-sharafiev/go-rest-common/exception"
	"github.com/rus-sharafiev/go-rest-common/jwt"
	"github.com/rus-sharafiev/go-rest-common/localization"
	"github.com/rus-sharafiev/go-rest-common/mail"
	"golang.org/x/crypto/pbkdf2"
)

// -- Reset -----------------------------------------------------------------------
type resetPassword struct {
	db *db.Postgres
}

func (c resetPassword) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		exception.MethodNotAllowed(w)
		return
	}
	var resetPasswordDto ResetPasswordDto
	json.NewDecoder(r.Body).Decode(&resetPasswordDto)

	var (
		userId     int
		userAccess string
	)
	checkEmailQuery := `SELECT "id", "access" FROM users WHERE "email" = $1`
	if err := c.db.QueryRow(&checkEmailQuery, &resetPasswordDto.Email).Scan(&userId, &userAccess); err != nil {
		if err == pgx.ErrNoRows {
			exception.BadRequestFields(w, map[string]string{
				"email": localization.SelectString(r, localization.Langs{
					En: "Email does not exist",
					Ru: "Email не зарегистрирован",
				}),
			})
		} else {
			exception.InternalServerError(w, err)
		}
		return
	}

	// Generate token
	token, err := jwt.GenerateAccessToken(userId, userAccess)
	if err != nil {
		exception.InternalServerError(w, err)
		return
	}

	// Create uniq redis id
	id, err := uuid.NewRandom()
	if err != nil {
		exception.InternalServerError(w, err)
		return
	}

	// Write token to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	if err := rdb.SetNX(context.Background(), id.String(), token, 5*time.Minute).Err(); err != nil {
		fmt.Println(err)
		exception.InternalServerError(w, err)
		return
	}

	// Set cookie
	cookie := &http.Cookie{
		Name:   "reset-password-id",
		Value:  id.String(),
		Path:   "/api/auth/password/update",
		MaxAge: 350,
	}
	http.SetCookie(w, cookie)

	// Send email with reset link
	link := r.Header.Get("Origin") + "/update-password/?token=" + token
	if err := mail.SendPasswordResetLink(resetPasswordDto.Email, link); err != nil {
		exception.InternalServerError(w, fmt.Errorf("mail server error: %v", err))
		return
	}

	successMessage := Message{
		StatusCode: http.StatusOK,
		Message: localization.SelectString(r, localization.Langs{
			En: "Message with reset password link has been sent successfully",
			Ru: "Письмо с ссылкой для сброса пароля успешно отправлено",
		}),
	}

	// OK response
	json.NewEncoder(w).Encode(&successMessage)
}

var ResetPassword = &resetPassword{db: &db.Instance}

// -- Update ----------------------------------------------------------------------
type updatePassword struct {
	db *db.Postgres
}

func (c updatePassword) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		exception.MethodNotAllowed(w)
		return
	}

	var updatePasswordDto UpdatePasswordDto
	json.NewDecoder(r.Body).Decode(&updatePasswordDto)

	// Get signup id from cookie
	resetPasswordIdCookie, err := r.Cookie("reset-password-id")
	if err != nil {
		exception.BadRequestError(w, err)
		return
	}

	// Get signup data from Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	tokenFromRedis, err := rdb.Get(context.Background(), resetPasswordIdCookie.Value).Result()
	if err == redis.Nil {
		exception.BadRequestMessage(w, localization.SelectString(r, localization.Langs{
			En: "The link has expired",
			Ru: "Срок действия ссылки истек",
		}))
		return
	} else if err != nil {
		exception.InternalServerError(w, err)
		return
	}

	defer rdb.Del(context.Background(), resetPasswordIdCookie.Value)

	if updatePasswordDto.Token != tokenFromRedis {
		exception.BadRequestMessage(w, localization.SelectString(r, localization.Langs{
			En: "Invalid token has been provided",
			Ru: "Предоставлен неверный токен",
		}))
	}

	claims, err := jwt.Validate(tokenFromRedis)
	if err != nil {
		exception.BadRequestMessage(w, localization.SelectString(r, localization.Langs{
			En: "Invalid token has been provided",
			Ru: "Предоставлен неверный токен",
		}))
		return
	}

	// Generate password hash
	salt := make([]byte, 16)
	rand.Read(salt)
	hash := pbkdf2.Key([]byte(updatePasswordDto.Password), salt, 4096, 32, sha1.New)

	var hashedPassword strings.Builder
	hashedPassword.WriteString(base64.StdEncoding.EncodeToString(hash))
	hashedPassword.WriteString(".")
	hashedPassword.WriteString(base64.StdEncoding.EncodeToString(salt))

	updatePasswordQuery := `
			UPDATE passwords SET "passwordHash" = $2
			WHERE "userId" = $1;
	`
	if _, err := c.db.Query(&updatePasswordQuery, &claims.UserId, hashedPassword.String()); err != nil {
		exception.InternalServerError(w, err)
		return
	}

	// OK response
	successMessage := Message{
		StatusCode: http.StatusOK,
		Message: localization.SelectString(r, localization.Langs{
			En: "Password changed successfully",
			Ru: "Пароль успешно изменен",
		}),
	}
	json.NewEncoder(w).Encode(&successMessage)
}

var UpdatePassword = &updatePassword{db: &db.Instance}
