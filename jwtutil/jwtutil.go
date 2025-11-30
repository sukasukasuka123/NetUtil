package jwtutil

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	Secret        []byte
	TokenDuration time.Duration
	CookieName    string
}

func NewJWTManager(secret string, duration time.Duration, cookieName string) *JWTManager {
	return &JWTManager{
		Secret:        []byte(secret),
		TokenDuration: duration,
		CookieName:    cookieName,
	}
}

// 生成 JWT
func (m *JWTManager) GenerateToken(claims jwt.MapClaims) (string, error) {
	claims["exp"] = time.Now().Add(m.TokenDuration).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.Secret)
}

// 解析 JWT
func (m *JWTManager) ParseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return m.Secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return token.Claims.(jwt.MapClaims), nil
}

// 写 HttpOnly Cookie
func (m *JWTManager) SetTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     m.CookieName,
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   int(m.TokenDuration.Seconds()),
		Secure:   false, // 生产环境改为 true (https)
		SameSite: http.SameSiteLaxMode,
	})
}

// 从请求读取 Cookie 中的 JWT
func (m *JWTManager) ReadTokenFromCookie(r *http.Request) (string, error) {
	c, err := r.Cookie(m.CookieName)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}
