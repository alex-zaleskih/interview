package auth

import (
	"testing"
	"time"
)

func TestSessions_IssueAndVerify(t *testing.T) {
	s := NewSessions([]byte("secret"), time.Hour)

	userID, ok := s.Verify(s.Issue("652f1a000000000000000001"))

	if !ok || userID != "652f1a000000000000000001" {
		t.Errorf("Verify() = (%q, %v), want (%q, true)", userID, ok, "652f1a000000000000000001")
	}
}

func TestSessions_RejectsInvalidTokens(t *testing.T) {
	s := NewSessions([]byte("secret"), time.Hour)
	valid := s.Issue("user-1")

	expired := NewSessions([]byte("secret"), time.Hour)
	expired.now = func() time.Time { return time.Now().Add(-2 * time.Hour) }

	tests := []struct {
		name  string
		token string
	}{
		{name: "empty", token: ""},
		{name: "garbage", token: "not-a-token"},
		{name: "tampered user id", token: "user-2" + valid[len("user-1"):]},
		{name: "signed with another secret", token: NewSessions([]byte("other"), time.Hour).Issue("user-1")},
		{name: "expired", token: expired.Issue("user-1")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if userID, ok := s.Verify(tt.token); ok {
				t.Errorf("Verify(%q) = (%q, true), want rejection", tt.token, userID)
			}
		})
	}
}
