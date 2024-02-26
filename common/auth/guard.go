package auth

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/rus-sharafiev/go-rest/common/jwt"
)

func Guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if auth := r.Header.Get("Authorization"); len(auth) != 0 {
			token := strings.Split(auth, " ")[1]

			claims, err := jwt.Validate(token)
			if err == nil {
				r.Header.Add("userID", strconv.Itoa(claims.UserId))
				r.Header.Add("userAccess", claims.UserAccess)
			} else {
				r.Header.Del("userID")
				r.Header.Del("userAccess")
			}

		}

		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func Headers(r *http.Request) (string, string) {
	userID := r.Header.Get("userID")
	userAccess := r.Header.Get("userAccess")

	return userID, userAccess
}
