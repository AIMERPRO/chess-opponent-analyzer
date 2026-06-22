package auth

import (
	"errors"
	"testing"

	"github.com/AIMERPRO/chess-opponent-analyzer/internal/core/apperrors"
)

// validateCase is a single row of a validation table test.
type validateCase struct {
	name        string
	req         validatable
	wantError   bool
	expectedErr error
}

// runValidateCases runs a table of cases: a single place for the loop and
// assertions so they are not duplicated across tests.
func runValidateCases(t *testing.T, tests []validateCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantError)
			}
			if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
				t.Errorf("Validate() error = %v, expected %v", err, tt.expectedErr)
			}
		})
	}
}

func TestLoginRequestDTO_Validate(t *testing.T) {
	runValidateCases(t, []validateCase{
		{
			name:        "valid login",
			req:         LoginRequestDTO{Username: "test123", Password: "test1234", DeviceID: "device-1"},
			wantError:   false,
			expectedErr: nil,
		},
		{
			name:        "empty username",
			req:         LoginRequestDTO{Username: "", Password: "test1234", DeviceID: "device-1"},
			wantError:   true,
			expectedErr: apperrors.ErrInvalidInput,
		},
		{
			name:        "short username",
			req:         LoginRequestDTO{Username: "test", Password: "test1234", DeviceID: "device-1"},
			wantError:   true,
			expectedErr: apperrors.ErrInvalidInput,
		},
		{
			name:        "empty password",
			req:         LoginRequestDTO{Username: "test123", Password: "", DeviceID: "device-1"},
			wantError:   true,
			expectedErr: apperrors.ErrInvalidInput,
		},
		{
			name:        "short password",
			req:         LoginRequestDTO{Username: "test123", Password: "test1", DeviceID: "device-1"},
			wantError:   true,
			expectedErr: apperrors.ErrInvalidInput,
		},
		{
			name:        "empty deviceID",
			req:         LoginRequestDTO{Username: "test123", Password: "test1234", DeviceID: ""},
			wantError:   true,
			expectedErr: apperrors.ErrInvalidInput,
		},
	})
}

func TestRegisterRequestDTO_Validate(t *testing.T) {
	runValidateCases(t, []validateCase{
		{
			name:        "valid register",
			req:         RegisterRequestDTO{Username: "test123", Password: "test1234", LichessUsername: "magnus", DeviceID: "device-1"},
			wantError:   false,
			expectedErr: nil,
		},
		{
			name:        "empty username",
			req:         RegisterRequestDTO{Username: "", Password: "test1234", LichessUsername: "magnus", DeviceID: "device-1"},
			wantError:   true,
			expectedErr: apperrors.ErrInvalidInput,
		},
		{
			name:        "short username",
			req:         RegisterRequestDTO{Username: "test", Password: "test1234", LichessUsername: "magnus", DeviceID: "device-1"},
			wantError:   true,
			expectedErr: apperrors.ErrInvalidInput,
		},
		{
			name:        "empty password",
			req:         RegisterRequestDTO{Username: "test123", Password: "", LichessUsername: "magnus", DeviceID: "device-1"},
			wantError:   true,
			expectedErr: apperrors.ErrInvalidInput,
		},
		{
			name:        "short password",
			req:         RegisterRequestDTO{Username: "test123", Password: "test1", LichessUsername: "magnus", DeviceID: "device-1"},
			wantError:   true,
			expectedErr: apperrors.ErrInvalidInput,
		},
		{
			name:        "empty lichess username",
			req:         RegisterRequestDTO{Username: "test123", Password: "test1234", LichessUsername: "", DeviceID: "device-1"},
			wantError:   true,
			expectedErr: apperrors.ErrInvalidInput,
		},
		{
			name:        "empty deviceID",
			req:         RegisterRequestDTO{Username: "test123", Password: "test1234", LichessUsername: "magnus", DeviceID: ""},
			wantError:   true,
			expectedErr: apperrors.ErrInvalidInput,
		},
	})
}

func TestUpdateUserDTO_Validate(t *testing.T) {
	runValidateCases(t, []validateCase{
		{
			name:        "valid username",
			req:         UpdateUserDTO{Username: strPtr("test123")},
			wantError:   false,
			expectedErr: nil,
		},
		{
			name:        "nil username is allowed",
			req:         UpdateUserDTO{Username: nil},
			wantError:   false,
			expectedErr: nil,
		},
		{
			name:        "short username",
			req:         UpdateUserDTO{Username: strPtr("test")},
			wantError:   true,
			expectedErr: apperrors.ErrInvalidInput,
		},
	})
}

func TestTokenRequestDTO_Validate(t *testing.T) {
	runValidateCases(t, []validateCase{
		{
			name:        "valid refresh token",
			req:         TokenRequestDTO{RefreshToken: "some-refresh-token"},
			wantError:   false,
			expectedErr: nil,
		},
		{
			name:        "empty refresh token",
			req:         TokenRequestDTO{RefreshToken: ""},
			wantError:   true,
			expectedErr: apperrors.ErrInvalidInput,
		},
	})
}

func strPtr(s string) *string {
	return &s
}
