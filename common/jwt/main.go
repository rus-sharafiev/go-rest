package jwt

import (
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserId     int    `json:"userId"`
	UserAccess string `json:"userAccess"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(id int, userAccess string) (string, error) {
	claims := Claims{
		id,
		userAccess,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_KEY")))
}

func SetRefreshToken(id int, userAccess string, w http.ResponseWriter) error {
	claims := Claims{
		id,
		userAccess,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshToken, err := token.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:   "refresh-token",
		Value:  "Bearer " + refreshToken,
		Path:   "/api/auth/refresh",
		MaxAge: 0,
	}
	http.SetCookie(w, cookie)

	return nil
}

func Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_KEY")), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, nil
	}
}
