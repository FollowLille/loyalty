package auth

import (
	"testing"
	"time"

	"github.com/FollowLille/loyalty/internal/config"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	type args struct {
		username string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid_username",
			args: args{
				username: "test_user",
			},
			wantErr: false,
		},
		{
			name: "empty_username",
			args: args{
				username: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.SuperSecretKey = "test_secret" // Установим тестовый ключ
			token, err := GenerateToken(tt.args.username)
			if tt.wantErr {
				assert.Error(t, err, "GenerateToken() should return error for test case: %v", tt.name)
			} else {
				assert.NoError(t, err, "GenerateToken() failed for test case: %v", tt.name)
				assert.NotEmpty(t, token, "Token should not be empty for test case: %v", tt.name)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	type args struct {
		tokenStr string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "valid_token",
			args: args{
				tokenStr: func() string {
					config.SuperSecretKey = "test_secret"
					token, _ := GenerateToken("test_user")
					return token
				}(),
			},
			want:    "test_user",
			wantErr: false,
		},
		{
			name: "expired_token",
			args: args{
				tokenStr: func() string {
					token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
						"username": "expired_user",
						"exp":      time.Now().Add(-time.Hour).Unix(),
					})
					tokenStr, _ := token.SignedString([]byte("test_secret"))
					return tokenStr
				}(),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "invalid_signature",
			args: args{
				tokenStr: func() string {
					token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
						"username": "invalid_user",
						"exp":      time.Now().Add(time.Hour).Unix(),
					})
					tokenStr, _ := token.SignedString([]byte("wrong_secret"))
					return tokenStr
				}(),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "malformed_token",
			args: args{
				tokenStr: "malformed.token.string",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "missing_username",
			args: args{
				tokenStr: func() string {
					token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
						"exp": time.Now().Add(time.Hour).Unix(),
					})
					tokenStr, _ := token.SignedString([]byte("test_secret"))
					return tokenStr
				}(),
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.SuperSecretKey = "test_secret" // Установим тестовый ключ
			username, err := ValidateToken(tt.args.tokenStr)
			if tt.wantErr {
				assert.Error(t, err, "ValidateToken() should return error for test case: %v", tt.name)
				assert.Empty(t, username, "Username should be empty for test case: %v", tt.name)
			} else {
				assert.NoError(t, err, "ValidateToken() failed for test case: %v", tt.name)
				assert.Equal(t, tt.want, username, "Expected and actual usernames do not match for test case: %v", tt.name)
			}
		})
	}
}
