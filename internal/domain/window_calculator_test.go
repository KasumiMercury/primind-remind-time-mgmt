package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/KasumiMercury/primind-remind-time-mgmt/internal/domain"
)

func TestSlideWindowWidthCalculator_EmptyTimes(t *testing.T) {
	calc := domain.NewSlideWindowWidthCalculator()

	result := calc.CalculateSlideWindowWidths([]time.Time{}, domain.TypeNear)

	assert.Nil(t, result)
}

// === Single Reminder Tests (TargetAt only) ===

func TestSlideWindowWidthCalculator_SingleReminder_StrictTypes(t *testing.T) {
	calc := domain.NewSlideWindowWidthCalculator()
	now := time.Now()

	tests := []struct {
		name     string
		taskType domain.Type
		expected time.Duration
	}{
		{
			name:     "short type gets 2 min (strict TargetAt)",
			taskType: domain.TypeShort,
			expected: 2 * time.Minute,
		},
		{
			name:     "scheduled type gets 2 min (strict TargetAt)",
			taskType: domain.TypeScheduled,
			expected: 2 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateSlideWindowWidths([]time.Time{now}, tt.taskType)
			assert.Equal(t, tt.expected, result[now].Duration())
		})
	}
}

func TestSlideWindowWidthCalculator_SingleReminder_FlexibleTypes(t *testing.T) {
	calc := domain.NewSlideWindowWidthCalculator()
	now := time.Now()

	tests := []struct {
		name     string
		taskType domain.Type
		expected time.Duration
	}{
		{
			name:     "near type gets 5 min (flexible TargetAt)",
			taskType: domain.TypeNear,
			expected: 5 * time.Minute,
		},
		{
			name:     "relaxed type gets 5 min (flexible TargetAt)",
			taskType: domain.TypeRelaxed,
			expected: 5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateSlideWindowWidths([]time.Time{now}, tt.taskType)
			assert.Equal(t, tt.expected, result[now].Duration())
		})
	}
}

// === Two Reminders Tests (Intermediate + TargetAt) ===

func TestSlideWindowWidthCalculator_TwoReminders_TargetAtGetsFixedWidth(t *testing.T) {
	calc := domain.NewSlideWindowWidthCalculator()
	now := time.Now()
	times := []time.Time{now, now.Add(30 * time.Minute)} // 30 min apart

	// near type: TargetAt should be 5 min regardless of interval
	result := calc.CalculateSlideWindowWidths(times, domain.TypeNear)
	assert.Equal(t, 5*time.Minute, result[times[1]].Duration()) // TargetAt

	// short type: TargetAt should be 2 min
	result = calc.CalculateSlideWindowWidths(times, domain.TypeShort)
	assert.Equal(t, 2*time.Minute, result[times[1]].Duration()) // TargetAt
}

func TestSlideWindowWidthCalculator_TwoReminders_IntermediateGets30Percent(t *testing.T) {
	calc := domain.NewSlideWindowWidthCalculator()
	now := time.Now()

	tests := []struct {
		name           string
		intervalToNext time.Duration
		expectedWidth  time.Duration
	}{
		{
			name:           "10 min interval: 30% = 3 min",
			intervalToNext: 10 * time.Minute,
			expectedWidth:  3 * time.Minute,
		},
		{
			name:           "20 min interval: 30% = 6 min",
			intervalToNext: 20 * time.Minute,
			expectedWidth:  6 * time.Minute,
		},
		{
			name:           "33 min 20 sec interval: 30% = 10 min (exactly at max)",
			intervalToNext: 33*time.Minute + 20*time.Second,
			expectedWidth:  10 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			times := []time.Time{
				now,                        // Intermediate
				now.Add(tt.intervalToNext), // TargetAt
			}
			result := calc.CalculateSlideWindowWidths(times, domain.TypeNear)
			assert.Equal(t, tt.expectedWidth, result[times[0]].Duration())
		})
	}
}

func TestSlideWindowWidthCalculator_TwoReminders_IntermediateClampedToMinimum(t *testing.T) {
	calc := domain.NewSlideWindowWidthCalculator()
	now := time.Now()

	// 2 min interval: 30% = 36 sec, should clamp to 1 min minimum
	times := []time.Time{
		now,
		now.Add(2 * time.Minute),
	}

	result := calc.CalculateSlideWindowWidths(times, domain.TypeNear)

	assert.Equal(t, 1*time.Minute, result[times[0]].Duration())  // Intermediate (clamped to min)
	assert.Equal(t, 5*time.Minute, result[times[1]].Duration()) // TargetAt
}

func TestSlideWindowWidthCalculator_TwoReminders_IntermediateClampedToMaximum(t *testing.T) {
	calc := domain.NewSlideWindowWidthCalculator()
	now := time.Now()

	// 60 min interval: 30% = 18 min > 10 min max, should clamp to 10 min
	times := []time.Time{
		now,
		now.Add(60 * time.Minute),
	}

	result := calc.CalculateSlideWindowWidths(times, domain.TypeRelaxed)

	assert.Equal(t, 10*time.Minute, result[times[0]].Duration()) // Intermediate (clamped to max)
	assert.Equal(t, 5*time.Minute, result[times[1]].Duration())  // TargetAt
}

// === Three Reminders Tests ===

func TestSlideWindowWidthCalculator_ThreeReminders_MixedIntervals(t *testing.T) {
	calc := domain.NewSlideWindowWidthCalculator()
	now := time.Now()
	times := []time.Time{
		now,                        // Intermediate: interval to next = 20 min, 30% = 6 min
		now.Add(20 * time.Minute),  // Intermediate: interval to next = 40 min, 30% = 12 min -> clamp to 10 min
		now.Add(60 * time.Minute),  // TargetAt: 5 min (near type)
	}

	result := calc.CalculateSlideWindowWidths(times, domain.TypeNear)

	assert.Equal(t, 6*time.Minute, result[times[0]].Duration())  // 20 * 0.30 = 6
	assert.Equal(t, 10*time.Minute, result[times[1]].Duration()) // 40 * 0.30 = 12 -> 10 max
	assert.Equal(t, 5*time.Minute, result[times[2]].Duration())  // TargetAt
}

func TestSlideWindowWidthCalculator_ThreeReminders_StrictType(t *testing.T) {
	calc := domain.NewSlideWindowWidthCalculator()
	now := time.Now()
	times := []time.Time{
		now,                        // Intermediate: interval to next = 10 min, 30% = 3 min
		now.Add(10 * time.Minute),  // Intermediate: interval to next = 20 min, 30% = 6 min -> clamp to 5 min (short)
		now.Add(30 * time.Minute),  // TargetAt: 2 min (short type)
	}

	result := calc.CalculateSlideWindowWidths(times, domain.TypeShort)

	assert.Equal(t, 3*time.Minute, result[times[0]].Duration()) // 10 * 0.30 = 3
	assert.Equal(t, 5*time.Minute, result[times[1]].Duration()) // 20 * 0.30 = 6 -> clamp to 5 (short max)
	assert.Equal(t, 2*time.Minute, result[times[2]].Duration()) // TargetAt (short)
}

// === Four Reminders Tests ===

func TestSlideWindowWidthCalculator_FourReminders(t *testing.T) {
	calc := domain.NewSlideWindowWidthCalculator()
	now := time.Now()
	times := []time.Time{
		now,                        // Intermediate: interval = 5 min, 30% = 1.5 min
		now.Add(5 * time.Minute),   // Intermediate: interval = 10 min, 30% = 3 min
		now.Add(15 * time.Minute),  // Intermediate: interval = 15 min, 30% = 4.5 min
		now.Add(30 * time.Minute),  // TargetAt: 2 min (short type)
	}

	result := calc.CalculateSlideWindowWidths(times, domain.TypeShort)

	assert.Equal(t, 1*time.Minute+30*time.Second, result[times[0]].Duration()) // 5 * 0.30 = 1.5
	assert.Equal(t, 3*time.Minute, result[times[1]].Duration())                 // 10 * 0.30 = 3
	assert.Equal(t, 4*time.Minute+30*time.Second, result[times[2]].Duration()) // 15 * 0.30 = 4.5
	assert.Equal(t, 2*time.Minute, result[times[3]].Duration())                 // TargetAt
}

// === Unsorted Input Tests ===

func TestSlideWindowWidthCalculator_UnsortedTimes_CorrectlyIdentifiesTargetAt(t *testing.T) {
	calc := domain.NewSlideWindowWidthCalculator()
	now := time.Now()
	// Input times are NOT sorted - should still correctly identify chronological last as TargetAt
	times := []time.Time{
		now.Add(60 * time.Minute), // This is chronologically last = TargetAt
		now,                        // This is chronologically first
		now.Add(30 * time.Minute),  // This is middle
	}

	result := calc.CalculateSlideWindowWidths(times, domain.TypeNear)

	// First (chronologically): interval to next = 30 min, 30% = 9 min
	assert.Equal(t, 9*time.Minute, result[now].Duration())

	// Middle: interval to next = 30 min, 30% = 9 min
	assert.Equal(t, 9*time.Minute, result[now.Add(30*time.Minute)].Duration())

	// TargetAt (60 min): should be 5 min (near type)
	assert.Equal(t, 5*time.Minute, result[now.Add(60*time.Minute)].Duration())
}

// === CalculateSingleSlideWindowWidth Tests ===

func TestSlideWindowWidthCalculator_CalculateSingleSlideWindowWidth(t *testing.T) {
	calc := domain.NewSlideWindowWidthCalculator()

	tests := []struct {
		name     string
		taskType domain.Type
		expected time.Duration
	}{
		{
			name:     "short type gets TargetAt width (2 min)",
			taskType: domain.TypeShort,
			expected: 2 * time.Minute,
		},
		{
			name:     "scheduled type gets TargetAt width (2 min)",
			taskType: domain.TypeScheduled,
			expected: 2 * time.Minute,
		},
		{
			name:     "near type gets TargetAt width (5 min)",
			taskType: domain.TypeNear,
			expected: 5 * time.Minute,
		},
		{
			name:     "relaxed type gets TargetAt width (5 min)",
			taskType: domain.TypeRelaxed,
			expected: 5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateSingleSlideWindowWidth(tt.taskType)
			assert.Equal(t, tt.expected, result.Duration())
		})
	}
}

// === Edge Cases ===

func TestSlideWindowWidthCalculator_VeryCloseIntervals(t *testing.T) {
	calc := domain.NewSlideWindowWidthCalculator()
	now := time.Now()

	// 1 min interval: 30% = 18 sec, should clamp to 1 min minimum
	times := []time.Time{
		now,
		now.Add(1 * time.Minute),
	}

	result := calc.CalculateSlideWindowWidths(times, domain.TypeRelaxed)

	assert.Equal(t, 1*time.Minute, result[times[0]].Duration())  // Clamped to min
	assert.Equal(t, 5*time.Minute, result[times[1]].Duration())  // TargetAt
}

func TestSlideWindowWidthCalculator_VeryLongIntervals(t *testing.T) {
	calc := domain.NewSlideWindowWidthCalculator()
	now := time.Now()

	// 120 min interval: 30% = 36 min > 10 min max, should clamp to 10 min
	times := []time.Time{
		now,
		now.Add(120 * time.Minute),
	}

	result := calc.CalculateSlideWindowWidths(times, domain.TypeNear)

	assert.Equal(t, 10*time.Minute, result[times[0]].Duration()) // Clamped to max
	assert.Equal(t, 5*time.Minute, result[times[1]].Duration())  // TargetAt
}

func TestSlideWindowWidthCalculator_ExactlyAtBoundary(t *testing.T) {
	calc := domain.NewSlideWindowWidthCalculator()
	now := time.Now()

	// Exactly at minimum boundary: 3 min 20 sec interval, 30% = 1 min exactly
	times := []time.Time{
		now,
		now.Add(3*time.Minute + 20*time.Second),
	}

	result := calc.CalculateSlideWindowWidths(times, domain.TypeNear)

	assert.Equal(t, 1*time.Minute, result[times[0]].Duration()) // Exactly at min
	assert.Equal(t, 5*time.Minute, result[times[1]].Duration()) // TargetAt
}

// === Short Type Intermediate Max Tests ===

func TestSlideWindowWidthCalculator_ShortType_IntermediateClampedTo5Min(t *testing.T) {
	calc := domain.NewSlideWindowWidthCalculator()
	now := time.Now()

	// 60 min interval: 30% = 18 min > 5 min max for short, should clamp to 5 min
	times := []time.Time{
		now,
		now.Add(60 * time.Minute),
	}

	result := calc.CalculateSlideWindowWidths(times, domain.TypeShort)

	assert.Equal(t, 5*time.Minute, result[times[0]].Duration())  // Clamped to 5 min (short max)
	assert.Equal(t, 2*time.Minute, result[times[1]].Duration())  // TargetAt (short)
}

func TestSlideWindowWidthCalculator_ShortVsOtherTypes_IntermediateMaxDifference(t *testing.T) {
	calc := domain.NewSlideWindowWidthCalculator()
	now := time.Now()

	// 60 min interval: 30% = 18 min
	// short should clamp to 5 min, others should clamp to 10 min
	times := []time.Time{
		now,
		now.Add(60 * time.Minute),
	}

	// short type: intermediate max = 5 min
	resultShort := calc.CalculateSlideWindowWidths(times, domain.TypeShort)
	assert.Equal(t, 5*time.Minute, resultShort[times[0]].Duration())

	// scheduled type: intermediate max = 10 min
	resultScheduled := calc.CalculateSlideWindowWidths(times, domain.TypeScheduled)
	assert.Equal(t, 10*time.Minute, resultScheduled[times[0]].Duration())

	// near type: intermediate max = 10 min
	resultNear := calc.CalculateSlideWindowWidths(times, domain.TypeNear)
	assert.Equal(t, 10*time.Minute, resultNear[times[0]].Duration())

	// relaxed type: intermediate max = 10 min
	resultRelaxed := calc.CalculateSlideWindowWidths(times, domain.TypeRelaxed)
	assert.Equal(t, 10*time.Minute, resultRelaxed[times[0]].Duration())
}
