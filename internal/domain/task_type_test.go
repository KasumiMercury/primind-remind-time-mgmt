package domain

import (
	"errors"
	"testing"
)

func TestNewType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Type
		wantErr error
	}{
		{
			name:    "valid short type",
			input:   "short",
			want:    TypeShort,
			wantErr: nil,
		},
		{
			name:    "valid near type",
			input:   "near",
			want:    TypeNear,
			wantErr: nil,
		},
		{
			name:    "valid relaxed type",
			input:   "relaxed",
			want:    TypeRelaxed,
			wantErr: nil,
		},
		{
			name:    "valid scheduled type",
			input:   "scheduled",
			want:    TypeScheduled,
			wantErr: nil,
		},
		{
			name:    "invalid type returns error",
			input:   "invalid",
			want:    "",
			wantErr: ErrInvalidTaskType,
		},
		{
			name:    "empty string returns error",
			input:   "",
			want:    "",
			wantErr: ErrInvalidTaskType,
		},
		{
			name:    "uppercase type returns error",
			input:   "SHORT",
			want:    "",
			wantErr: ErrInvalidTaskType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewType(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("NewType(%q) expected error, got nil", tt.input)

					return
				}

				if !errors.Is(err, tt.wantErr) {
					t.Errorf("NewType(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				}

				return
			}

			if err != nil {
				t.Errorf("NewType(%q) unexpected error: %v", tt.input, err)

				return
			}

			if got != tt.want {
				t.Errorf("NewType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
