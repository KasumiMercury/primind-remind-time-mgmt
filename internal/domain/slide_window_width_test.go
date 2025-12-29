package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/domain"
)

func TestNewSlideWindowWidthSuccess(t *testing.T) {
	tests := []struct {
		name            string
		duration        time.Duration
		expectedSeconds int32
	}{
		{
			name:            "minimum valid width (1 minute)",
			duration:        1 * time.Minute,
			expectedSeconds: 60,
		},
		{
			name:            "maximum valid width (10 minutes)",
			duration:        10 * time.Minute,
			expectedSeconds: 600,
		},
		{
			name:            "strict width (2 minutes)",
			duration:        2 * time.Minute,
			expectedSeconds: 120,
		},
		{
			name:            "5 minutes",
			duration:        5 * time.Minute,
			expectedSeconds: 300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, err := domain.NewSlideWindowWidth(tt.duration)

			require.NoError(t, err)
			assert.Equal(t, tt.duration, width.Duration())
			assert.Equal(t, tt.expectedSeconds, width.Seconds())
			assert.False(t, width.IsZero())
		})
	}
}

func TestNewSlideWindowWidthError(t *testing.T) {
	tests := []struct {
		name        string
		duration    time.Duration
		expectedErr error
	}{
		{
			name:        "too small (30 seconds)",
			duration:    30 * time.Second,
			expectedErr: domain.ErrSlideWindowWidthTooSmall,
		},
		{
			name:        "zero duration",
			duration:    0,
			expectedErr: domain.ErrSlideWindowWidthTooSmall,
		},
		{
			name:        "negative duration",
			duration:    -1 * time.Minute,
			expectedErr: domain.ErrSlideWindowWidthTooSmall,
		},
		{
			name:        "too large (11 minutes)",
			duration:    11 * time.Minute,
			expectedErr: domain.ErrSlideWindowWidthTooLarge,
		},
		{
			name:        "way too large (30 minutes)",
			duration:    30 * time.Minute,
			expectedErr: domain.ErrSlideWindowWidthTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewSlideWindowWidth(tt.duration)

			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestSlideWindowWidthFromSecondsSuccess(t *testing.T) {
	tests := []struct {
		name             string
		seconds          int32
		expectedDuration time.Duration
	}{
		{
			name:             "minimum (60 seconds)",
			seconds:          60,
			expectedDuration: 1 * time.Minute,
		},
		{
			name:             "maximum (600 seconds)",
			seconds:          600,
			expectedDuration: 10 * time.Minute,
		},
		{
			name:             "medium (300 seconds)",
			seconds:          300,
			expectedDuration: 5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, err := domain.SlideWindowWidthFromSeconds(tt.seconds)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedDuration, width.Duration())
			assert.Equal(t, tt.seconds, width.Seconds())
		})
	}
}

func TestSlideWindowWidthFromSecondsError(t *testing.T) {
	tests := []struct {
		name    string
		seconds int32
	}{
		{
			name:    "too small (30 seconds)",
			seconds: 30,
		},
		{
			name:    "zero",
			seconds: 0,
		},
		{
			name:    "too large (601 seconds)",
			seconds: 601,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.SlideWindowWidthFromSeconds(tt.seconds)

			require.Error(t, err)
		})
	}
}

func TestMustSlideWindowWidthSuccess(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
	}{
		{
			name:     "valid minimum",
			duration: 1 * time.Minute,
		},
		{
			name:     "valid maximum",
			duration: 10 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				width := domain.MustSlideWindowWidth(tt.duration)
				assert.Equal(t, tt.duration, width.Duration())
			})
		})
	}
}

func TestMustSlideWindowWidthPanic(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
	}{
		{
			name:     "too small",
			duration: 30 * time.Second,
		},
		{
			name:     "too large",
			duration: 11 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() {
				domain.MustSlideWindowWidth(tt.duration)
			})
		})
	}
}

func TestSlideWindowWidthIsZero(t *testing.T) {
	tests := []struct {
		name     string
		width    domain.SlideWindowWidth
		expected bool
	}{
		{
			name:     "zero value",
			width:    domain.SlideWindowWidth{},
			expected: true,
		},
		{
			name:     "valid value",
			width:    domain.MustSlideWindowWidth(5 * time.Minute),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.width.IsZero())
		})
	}
}
