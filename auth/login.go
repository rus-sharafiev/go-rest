package auth

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/rus-sharafiev/go-rest-common/db"
	"github.com/rus-sharafiev/go-rest-common/exception"
	"github.com/rus-sharafiev/go-rest-common/jwt"
	"github.com/rus-sharafiev/go-rest-common/localization"
	"golang.org/x/crypto/pbkdf2"
)

type logIn struct {
	db *db.Postgres
}

func (c logIn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		exception.MethodNotAllowed(w)
		return
	}
	var logInDto LogInDto
	json.NewDecoder(r.Body).Decode(&logInDto)

	var userPswd UserPswd
	pswdQuery := `
		SELECT u."id", p."passwordHash"
		FROM users u
		LEFT JOIN passwords p
		ON u."id" = p."userId"
		WHERE u."email" = $1;
	`
	if err := c.db.QueryRow(&pswdQuery, logInDto.Email).Scan(&userPswd.UserId, &userPswd.PasswordHash); err != nil {
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

	passwordHashAndSalt := strings.Split(userPswd.PasswordHash, ".")
	passwordFromDb := passwordHashAndSalt[0]
	salt, err := base64.StdEncoding.DecodeString(passwordHashAndSalt[1])
	if err != nil {
		exception.InternalServerError(w, err)
		return
	}

	hash := pbkdf2.Key([]byte(logInDto.Password), salt, 4096, 32, sha1.New)
	providedPassword := base64.StdEncoding.EncodeToString(hash)

	if passwordFromDb != providedPassword {
		exception.BadRequestFields(w, map[string]string{
			"password": localization.SelectString(r, localization.Langs{
				En: "Incorrect password",
				Ru: "Неверный пароль",
			}),
		})
		return
	}

	query := `SELECT * FROM users WHERE "id" = $1;`
	rows, _ := c.db.Query(&query, userPswd.UserId)
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

	if fingerprint := r.Header.Get("Fingerprint"); len(fingerprint) != 0 {

		query := `
			INSERT INTO sessions ("userId", "fingerprint", "userAgent", "ip", "updatedAt") 
			VALUES (@userId, @fingerprint, @userAgent, @ip, CURRENT_TIMESTAMP)
			ON CONFLICT ("fingerprint") DO 
				UPDATE SET ("userId", "userAgent", "ip", "updatedAt") = 
				(EXCLUDED."userId", EXCLUDED."userAgent", EXCLUDED."ip", EXCLUDED."updatedAt");
		`
		args := pgx.NamedArgs{
			"userId":      userData.ID,
			"fingerprint": fingerprint,
			"userAgent":   r.Header.Get("User-Agent"),
			"ip":          strings.Split(r.RemoteAddr, ":")[0],
		}

		if _, err := c.db.Query(&query, args); err != nil {
			exception.InternalServerError(w, err)
			return
		}
	}

	// OK response
	json.NewEncoder(w).Encode(&result)
}

var LogIn = &logIn{db: &db.Instance}
