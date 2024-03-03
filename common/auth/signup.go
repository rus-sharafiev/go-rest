package auth

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rus-sharafiev/go-rest/common/db"
	"github.com/rus-sharafiev/go-rest/common/exception"
	"github.com/rus-sharafiev/go-rest/common/jwt"
	"github.com/rus-sharafiev/go-rest/common/localization"
	"github.com/rus-sharafiev/go-rest/common/mail"
	"golang.org/x/crypto/pbkdf2"
)

// -- Sign Up ---------------------------------------------------------------------
type signUp struct {
	db *db.Postgres
}

func (c signUp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		exception.MethodNotAllowed(w)
		return
	}

	var signUpDto SignUpDto
	json.NewDecoder(r.Body).Decode(&signUpDto)

	// Check recap
	if captcha := signUpDto.Grecaptcha; captcha != nil {

		if len(*captcha) == 0 {
			exception.BadRequestFields(w, map[string]string{
				"grecaptcha": localization.SelectString(r, localization.Langs{
					En: "Confirm that you are not a robot 🤖",
					Ru: "Подтвердите что вы не робот 🤖",
				}),
			})
			return

		} else {
			resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", url.Values{
				"secret":   {os.Getenv("reCAPTCHA_SECRET")},
				"response": {*captcha},
			})
			if err != nil {
				exception.InternalServerError(w, err)
				return
			}
			defer resp.Body.Close()

			var recaptchaResponse ReCaptchaResponse
			if err = json.NewDecoder(resp.Body).Decode(&recaptchaResponse); err != nil {
				exception.InternalServerError(w, err)
				return
			}

			fmt.Println(recaptchaResponse)

			if !recaptchaResponse.Success {
				exception.BadRequestFields(w, map[string]string{
					"grecaptcha": localization.SelectString(r, localization.Langs{
						En: "Google thinks you're a robot 🤷‍♂️",
						Ru: "Google считает что ты робот 🤷‍♂️",
					}),
				})
				return
			}

		}
	}

	checkEmailQuery := `SELECT "id" FROM users WHERE "email" = $1`
	if err := c.db.QueryRow(&checkEmailQuery, signUpDto.Email).Scan(); err != pgx.ErrNoRows {
		exception.BadRequestFields(w, map[string]string{
			"email": localization.SelectString(r, localization.Langs{
				En: "Email already exists",
				Ru: "Email уже существует",
			}),
		})
		return
	}

	code := rand.Intn(899999) + 100000
	if err := mail.SendCode(signUpDto.Email, code); err != nil {
		exception.InternalServerError(w, fmt.Errorf("mail server error: %v", err))
	}

	id, err := uuid.NewRandom()
	if err != nil {
		exception.InternalServerError(w, err)
		return
	}

	signUpData := SignUpData{
		Email:    signUpDto.Email,
		Password: signUpDto.Password,
		Code:     code,
	}

	signUpDataJson, err := json.Marshal(signUpData)
	if err != nil {
		exception.InternalServerError(w, err)
		return
	}

	// Write user data to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	if err := rdb.SetNX(context.Background(), id.String(), string(signUpDataJson), 2*time.Minute).Err(); err != nil {
		fmt.Println(err)
		exception.InternalServerError(w, err)
		return
	}

	// Set cookie
	cookie := &http.Cookie{
		Name:   "signup-id",
		Value:  id.String(),
		Path:   "/api/auth/signup/verify",
		MaxAge: 120,
	}
	http.SetCookie(w, cookie)

	successMessage := Message{
		StatusCode: http.StatusOK,
		Message: localization.SelectString(r, localization.Langs{
			En: "Message with confirmation code has been sent successfully",
			Ru: "Письмо с кодом подтверждения успешно отправлено",
		}),
	}

	// OK response
	json.NewEncoder(w).Encode(&successMessage)
}

var SignUp = &signUp{db: db.NewConnection()}

// -- Verify Signup ---------------------------------------------------------------
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
