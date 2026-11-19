package captcha

import (
	"bytes"
	"context"
	"crypto/rand"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mojocn/base64Captcha"
)

const (
	cookieName = "JSESSIONID"
	sessionTTL = 30 * time.Minute
	alphabet   = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
)

type entry struct {
	code    string
	expires time.Time
}

type Sessions struct {
	mu      sync.Mutex
	entries map[string]entry
	driver  *base64Captcha.DriverString
}

func New() *Sessions {
	return &Sessions{
		entries: make(map[string]entry),
		driver:  base64Captcha.NewDriverString(32, 120, 2, base64Captcha.OptionShowHollowLine, 4, alphabet, nil, nil, []string{"wqy-microhei.ttc"}),
	}
}

func (s *Sessions) Issue(w http.ResponseWriter, r *http.Request) error {
	var code [4]byte
	limit := big.NewInt(int64(len(alphabet)))
	for i := range code {
		value, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return err
		}
		code[i] = alphabet[value.Int64()]
	}
	picture, err := s.driver.DrawCaptcha(string(code[:]))
	if err != nil {
		return err
	}
	var raw bytes.Buffer
	if _, err := picture.WriteTo(&raw); err != nil {
		return err
	}

	now := time.Now()
	id := sessionID(r)
	s.mu.Lock()
	if existing, ok := s.entries[id]; !ok || !now.Before(existing.expires) {
		delete(s.entries, id)
		id = newSessionID(w)
	}
	s.entries[id] = entry{code: string(code[:]), expires: now.Add(sessionTTL)}
	s.mu.Unlock()

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", time.Unix(0, 0).UTC().Format(http.TimeFormat))
	_, _ = w.Write(raw.Bytes())
	return nil
}

// Verify 原子检查和清除验证码。注册输错时可继续使用当前验证码，登录输错时清除。
func (s *Sessions) Verify(w http.ResponseWriter, r *http.Request, code string, clearOnFailure bool) bool {
	id := sessionID(r)
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.entries[id]
	if !ok || !now.Before(existing.expires) {
		delete(s.entries, id)
		s.entries[newSessionID(w)] = entry{expires: now.Add(sessionTTL)}
		return false
	}
	code = strings.TrimFunc(code, func(r rune) bool { return r <= 0x20 })
	valid := existing.code != "" && strings.ToLower(existing.code) == strings.ToLower(code)
	if valid || clearOnFailure {
		existing.code = ""
	}
	existing.expires = now.Add(sessionTTL)
	s.entries[id] = existing
	return valid
}

func newSessionID(w http.ResponseWriter) string {
	id := rand.Text()
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: id, Path: "/", HttpOnly: true})
	return id
}

func (s *Sessions) Cleanup(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.mu.Lock()
			for id, session := range s.entries {
				if !now.Before(session.expires) {
					delete(s.entries, id)
				}
			}
			s.mu.Unlock()
		}
	}
}

func sessionID(r *http.Request) string {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}
