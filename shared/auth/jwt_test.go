package auth

import (
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		email   string
		secret  string
		expiry  time.Duration
		wantErr bool
	}{
		{
			name:    "valid token generation",
			userID:  "user-123",
			email:   "test@example.com",
			secret:  "my-secret-key",
			expiry:  time.Hour,
			wantErr: false,
		},
		{
			name:    "empty user ID",
			userID:  "",
			email:   "test@example.com",
			secret:  "my-secret-key",
			expiry:  time.Hour,
			wantErr: false, // JWT allows empty claims
		},
		{
			name:    "short expiry",
			userID:  "user-123",
			email:   "test@example.com",
			secret:  "my-secret-key",
			expiry:  time.Second,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.userID, tt.email, tt.secret, tt.expiry)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("GenerateToken() returned empty token")
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	secret := "my-secret-key"
	userID := "user-123"
	email := "test@example.com"

	t.Run("valid token", func(t *testing.T) {
		token, err := GenerateToken(userID, email, secret, time.Hour)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		claims, err := ValidateToken(token, secret)
		if err != nil {
			t.Errorf("ValidateToken() error = %v", err)
			return
		}

		if claims.UserID != userID {
			t.Errorf("ValidateToken() UserID = %v, want %v", claims.UserID, userID)
		}
		if claims.Email != email {
			t.Errorf("ValidateToken() Email = %v, want %v", claims.Email, email)
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		token, err := GenerateToken(userID, email, secret, time.Hour)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		_, err = ValidateToken(token, "wrong-secret")
		if err == nil {
			t.Error("ValidateToken() should return error for wrong secret")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		token, err := GenerateToken(userID, email, secret, -time.Hour)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		_, err = ValidateToken(token, secret)
		if err == nil {
			t.Error("ValidateToken() should return error for expired token")
		}
	})

	t.Run("malformed token", func(t *testing.T) {
		_, err := ValidateToken("not-a-valid-token", secret)
		if err == nil {
			t.Error("ValidateToken() should return error for malformed token")
		}
	})

	t.Run("empty token", func(t *testing.T) {
		_, err := ValidateToken("", secret)
		if err == nil {
			t.Error("ValidateToken() should return error for empty token")
		}
	})
}

func TestGenerateRefreshToken(t *testing.T) {
	secret := "my-secret-key"
	userID := "user-123"

	t.Run("valid refresh token generation", func(t *testing.T) {
		token, err := GenerateRefreshToken(userID, secret, 7*24*time.Hour)
		if err != nil {
			t.Errorf("GenerateRefreshToken() error = %v", err)
			return
		}
		if token == "" {
			t.Error("GenerateRefreshToken() returned empty token")
		}
	})

	t.Run("refresh token validation", func(t *testing.T) {
		token, err := GenerateRefreshToken(userID, secret, 7*24*time.Hour)
		if err != nil {
			t.Fatalf("Failed to generate refresh token: %v", err)
		}

		extractedUserID, err := ValidateRefreshToken(token, secret)
		if err != nil {
			t.Errorf("ValidateRefreshToken() error = %v", err)
			return
		}

		if extractedUserID != userID {
			t.Errorf("ValidateRefreshToken() UserID = %v, want %v", extractedUserID, userID)
		}
	})
}

func TestTokenRoundTrip(t *testing.T) {
	secret := "my-secret-key-for-testing"
	testCases := []struct {
		userID string
		email  string
	}{
		{"user-1", "user1@example.com"},
		{"user-2", "user2@example.com"},
		{"", ""}, // Edge case: empty values
		{"special-chars-!@#$%", "special+email@example.com"},
	}

	for _, tc := range testCases {
		t.Run("roundtrip_"+tc.userID, func(t *testing.T) {
			token, err := GenerateToken(tc.userID, tc.email, secret, time.Hour)
			if err != nil {
				t.Fatalf("GenerateToken() error = %v", err)
			}

			claims, err := ValidateToken(token, secret)
			if err != nil {
				t.Fatalf("ValidateToken() error = %v", err)
			}

			if claims.UserID != tc.userID {
				t.Errorf("UserID mismatch: got %v, want %v", claims.UserID, tc.userID)
			}
			if claims.Email != tc.email {
				t.Errorf("Email mismatch: got %v, want %v", claims.Email, tc.email)
			}
		})
	}
}

