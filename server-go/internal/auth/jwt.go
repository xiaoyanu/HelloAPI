package auth

import (
	"crypto/rand"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Mode int    `json:"mode"`
	jwt.RegisteredClaims
}

type JWT struct{ secret []byte }

func New(secret string) *JWT {
	if secret == "" {
		return &JWT{secret: []byte(rand.Text())}
	}
	return &JWT{secret: []byte(secret)}
}

func (j *JWT) Issue(id int, name string, mode int) (string, error) {
	claims := Claims{
		ID: id, Name: name, Mode: mode,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().AddDate(0, 0, 70))},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(j.secret)
}

func (j *JWT) Bearer(header string) (Claims, bool) {
	raw, ok := strings.CutPrefix(header, "Bearer ")
	if !ok {
		return Claims{}, false
	}
	var claims Claims
	token, err := jwt.ParseWithClaims(raw, &claims, func(*jwt.Token) (any, error) { return j.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuedAt())
	if err != nil || !token.Valid {
		return Claims{}, false
	}
	return claims, true
}
