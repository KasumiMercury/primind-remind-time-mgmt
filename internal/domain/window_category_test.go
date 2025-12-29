package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/domain"
)

func TestGetTargetAtWindowWidth(t *testing.T) {
	tests := []struct {
		name     string
		taskType domain.Type
		expected time.Duration
	}{
		{
			name:     "short gets 2 min",
			taskType: domain.TypeShort,
			expected: 2 * time.Minute,
		},
		{
			name:     "scheduled gets 2 min",
			taskType: domain.TypeScheduled,
			expected: 2 * time.Minute,
		},
		{
			name:     "near gets 5 min",
			taskType: domain.TypeNear,
			expected: 5 * time.Minute,
		},
		{
			name:     "relaxed gets 5 min",
			taskType: domain.TypeRelaxed,
			expected: 5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domain.GetTargetAtWindowWidth(tt.taskType)
			assert.Equal(t, tt.expected, result.Duration())
		})
	}
}

func TestGetIntermediateWindowWidth(t *testing.T) {
	tests := []struct {
		name           string
		taskType       domain.Type
		intervalToNext time.Duration
		expected       time.Duration
	}{
		// short type: max 5 min
		{
			name:           "short: 10 min interval -> 30% = 3 min (within max)",
			taskType:       domain.TypeShort,
			intervalToNext: 10 * time.Minute,
			expected:       3 * time.Minute,
		},
		{
			name:           "short: 20 min interval -> 30% = 6 min -> clamp to 5 min",
			taskType:       domain.TypeShort,
			intervalToNext: 20 * time.Minute,
			expected:       5 * time.Minute,
		},
		{
			name:           "short: 60 min interval -> 30% = 18 min -> clamp to 5 min",
			taskType:       domain.TypeShort,
			intervalToNext: 60 * time.Minute,
			expected:       5 * time.Minute,
		},
		{
			name:           "short: 2 min interval -> 30% = 36 sec -> clamp to 1 min",
			taskType:       domain.TypeShort,
			intervalToNext: 2 * time.Minute,
			expected:       1 * time.Minute,
		},
		// scheduled type: max 10 min
		{
			name:           "scheduled: 20 min interval -> 30% = 6 min (within max)",
			taskType:       domain.TypeScheduled,
			intervalToNext: 20 * time.Minute,
			expected:       6 * time.Minute,
		},
		{
			name:           "scheduled: 60 min interval -> 30% = 18 min -> clamp to 10 min",
			taskType:       domain.TypeScheduled,
			intervalToNext: 60 * time.Minute,
			expected:       10 * time.Minute,
		},
		// near type: max 10 min
		{
			name:           "near: 20 min interval -> 30% = 6 min (within max)",
			taskType:       domain.TypeNear,
			intervalToNext: 20 * time.Minute,
			expected:       6 * time.Minute,
		},
		{
			name:           "near: 60 min interval -> 30% = 18 min -> clamp to 10 min",
			taskType:       domain.TypeNear,
			intervalToNext: 60 * time.Minute,
			expected:       10 * time.Minute,
		},
		// relaxed type: max 10 min
		{
			name:           "relaxed: 20 min interval -> 30% = 6 min (within max)",
			taskType:       domain.TypeRelaxed,
			intervalToNext: 20 * time.Minute,
			expected:       6 * time.Minute,
		},
		{
			name:           "relaxed: 60 min interval -> 30% = 18 min -> clamp to 10 min",
			taskType:       domain.TypeRelaxed,
			intervalToNext: 60 * time.Minute,
			expected:       10 * time.Minute,
		},
		// Minimum clamp for non-short types
		{
			name:           "near: 2 min interval -> 30% = 36 sec -> clamp to 1 min",
			taskType:       domain.TypeNear,
			intervalToNext: 2 * time.Minute,
			expected:       1 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domain.GetIntermediateWindowWidth(tt.taskType, tt.intervalToNext)
			assert.Equal(t, tt.expected, result.Duration())
		})
	}
}
