package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWT(t *testing.T) {
	issuer := New("test-secret")
	token, err := issuer.Issue(42, "用户名", 1)
	if err != nil {
		t.Fatal(err)
	}
	claims, ok := issuer.Bearer("Bearer " + token)
	if !ok || claims.ID != 42 || claims.Name != "用户名" || claims.Mode != 1 {
		t.Fatalf("声明 = %#v", claims)
	}
	left := time.Until(claims.ExpiresAt.Time)
	if left > 70*24*time.Hour+time.Hour || left < 70*24*time.Hour-time.Hour-time.Second {
		t.Fatalf("有效期 = %s", left)
	}
	if _, ok := New("other-secret").Bearer("Bearer " + token); ok {
		t.Fatal("接受了其他密钥的令牌")
	}
	if _, ok := issuer.Bearer("bearer " + token); ok {
		t.Fatal("接受了小写前缀")
	}
	expired, err := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{ID: 42, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Second))}}).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := issuer.Bearer("Bearer " + expired); ok {
		t.Fatal("接受了过期令牌")
	}
}
