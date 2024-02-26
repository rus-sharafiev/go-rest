package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rus-sharafiev/go-rest/common/db"
	"github.com/rus-sharafiev/go-rest/common/exception"
	"github.com/rus-sharafiev/go-rest/common/localization"
	"github.com/rus-sharafiev/go-rest/common/mail"
)

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
