package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"strings"
	"time"
)

// Sessions issues and verifies signed session tokens.
//
// Token format: <userID>.<expiresAtUnix>.<base64url(HMAC-SHA256(userID.expiresAtUnix))>
type Sessions struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewSessions(secret []byte, ttl time.Duration) *Sessions {
	return &Sessions{secret: secret, ttl: ttl, now: time.Now}
}

func (s *Sessions) TTL() time.Duration {
	return s.ttl
}

func (s *Sessions) Issue(userID string) string {
	payload := userID + "." + strconv.FormatInt(s.now().Add(s.ttl).Unix(), 10)
	return payload + "." + s.sign(payload)
}

// Verify checks the token signature and expiration and returns the user ID stored in it.
func (s *Sessions) Verify(token string) (userID string, ok bool) {
	i := strings.LastIndexByte(token, '.')
	if i < 0 {
		return "", false
	}
	payload, signature := token[:i], token[i+1:]
	if !hmac.Equal([]byte(signature), []byte(s.sign(payload))) {
		return "", false
	}

	userID, expiresAt, found := strings.Cut(payload, ".")
	if !found || userID == "" {
		return "", false
	}
	exp, err := strconv.ParseInt(expiresAt, 10, 64)
	if err != nil || s.now().Unix() >= exp {
		return "", false
	}
	return userID, true
}

func (s *Sessions) sign(payload string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
