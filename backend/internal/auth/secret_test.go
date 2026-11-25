package auth

import (
	"strings"
	"testing"
)

func TestSecretFromEnv(t *testing.T) {
	tests := []struct {
		name      string
		appEnv    string
		secret    string
		wantErr   string
		wantWarn  bool
		wantValue string
	}{
		{
			name:    "production missing secret",
			appEnv:  "production",
			secret:  "",
			wantErr: "JWT_SECRET is required when APP_ENV=production",
		},
		{
			name:    "production short secret",
			appEnv:  "production",
			secret:  "short-secret",
			wantErr: "JWT_SECRET must be at least 32 characters when APP_ENV=production",
		},
		{
			name:      "production valid secret",
			appEnv:    "production",
			secret:    "this-is-a-valid-production-secret-32+",
			wantValue: "this-is-a-valid-production-secret-32+",
		},
		{
			name:      "development missing secret uses dev default",
			appEnv:    "development",
			secret:    "",
			wantValue: devJWTSecret,
		},
		{
			name:      "development explicit secret",
			appEnv:    "development",
			secret:    "local-dev-secret-that-is-long-enough",
			wantValue: "local-dev-secret-that-is-long-enough",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_ENV", tt.appEnv)
			t.Setenv("JWT_SECRET", tt.secret)

			got, err := SecretFromEnv()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.wantValue {
				t.Fatalf("expected secret %q, got %q", tt.wantValue, string(got))
			}
			if tt.name == "production valid secret" && len(got) < minJWTSecretLength {
				t.Fatalf("expected secret length >= %d", minJWTSecretLength)
			}
			_ = strings.TrimSpace // keep import used if adjusted
		})
	}
}
