package auth

import (
	"encoding/json"
	"net/http"

	"github.com/rus-sharafiev/go-rest-common/db"
	"github.com/rus-sharafiev/go-rest-common/exception"
	"github.com/rus-sharafiev/go-rest-common/localization"
)

type logOut struct {
	db *db.Postgres
}

func (c logOut) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		exception.MethodNotAllowed(w)
		return
	}

	userId, _ := Headers(r)
	if len(userId) == 0 {
		exception.Unauthorized(w)
		return
	}

	// Delete session
	if fingerprint := r.Header.Get("Fingerprint"); len(fingerprint) != 0 {
		query := `DELETE FROM sessions WHERE "fingerprint" = $1;`
		if _, err := c.db.Query(&query, fingerprint); err != nil {
			exception.InternalServerError(w, err)
			return
		}
	}

	// Remove refresh token
	cookie := &http.Cookie{
		Name:   "refresh-token",
		Value:  "",
		Path:   "/api/auth/refresh",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)

	successMessage := Message{
		StatusCode: http.StatusOK,
		Message: localization.SelectString(r, localization.Langs{
			En: "Logged out successfully",
			Ru: "Выход из системы",
		}),
	}

	// OK response
	json.NewEncoder(w).Encode(&successMessage)
}

var LogOut = &logOut{db: &db.Instance}
