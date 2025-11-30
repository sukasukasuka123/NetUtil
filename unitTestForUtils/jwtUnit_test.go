package unitTestForUtils

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"NetUtil/jwtutil"
)

// 基础：创建一个测试用的 JWTManager
func newTestManager() *jwtutil.JWTManager {
	return jwtutil.NewJWTManager("test-secret", time.Minute, "jwt_token")
}

// 1. 测试生成与解析 Token
func TestGenerateAndParseToken(t *testing.T) {
	m := newTestManager()

	claims := map[string]interface{}{
		"user_id": 123,
		"role":    "admin",
	}

	token, err := m.GenerateToken(claims)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatalf("expected token, got empty string")
	}

	parsed, err := m.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	if parsed["user_id"] != float64(123) { // MapClaims 默认是 float64
		t.Errorf("user_id mismatch, expected 123 got %v", parsed["user_id"])
	}
	if parsed["role"] != "admin" {
		t.Errorf("role mismatch, expected admin got %v", parsed["role"])
	}

	_, hasExp := parsed["exp"]
	if !hasExp {
		t.Errorf("exp should exist in claims")
	}
}

// 2. 测试解析非法 Token
func TestParseInvalidToken(t *testing.T) {
	m := newTestManager()

	_, err := m.ParseToken("not-a-valid-token")
	if err == nil {
		t.Fatalf("expected error when parsing invalid token, got nil")
	}
}

// 3. 测试写 Cookie
func TestSetTokenCookie(t *testing.T) {
	m := newTestManager()

	w := httptest.NewRecorder()
	m.SetTokenCookie(w, "mock-token")

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected cookie to be set")
	}

	c := cookies[0]
	if c.Name != m.CookieName {
		t.Errorf("cookie name mismatch: got %s", c.Name)
	}
	if c.Value != "mock-token" {
		t.Errorf("cookie value mismatch: got %s", c.Value)
	}
	if !c.HttpOnly {
		t.Errorf("cookie HttpOnly should be true")
	}
}

// 4. 测试从请求读取 token
func TestReadTokenFromCookie(t *testing.T) {
	m := newTestManager()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  m.CookieName,
		Value: "cookie-token",
	})

	token, err := m.ReadTokenFromCookie(r)
	if err != nil {
		t.Fatalf("ReadTokenFromCookie failed: %v", err)
	}
	if token != "cookie-token" {
		t.Errorf("expected 'cookie-token', got %s", token)
	}
}

// 5. 测试读取不存在 Cookie
func TestReadTokenFromCookie_Missing(t *testing.T) {
	m := newTestManager()

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := m.ReadTokenFromCookie(r)
	if err == nil {
		t.Fatalf("expected error for missing cookie, got nil")
	}
}
func TestJwtSuite(t *testing.T) {
	t.Run("GenerateAndParseToken", TestGenerateAndParseToken)
	t.Run("ParseInvalidToken", TestParseInvalidToken)
	t.Run("SetTokenCookie", TestSetTokenCookie)
	t.Run("ReadTokenFromCookie", TestReadTokenFromCookie)
	t.Run("ReadTokenFromCookie_Missing", TestReadTokenFromCookie_Missing)
}
