package domain

import (
	"sort"
	"time"
)

type SlideWindowWidthCalculator struct{}

func NewSlideWindowWidthCalculator() *SlideWindowWidthCalculator {
	return &SlideWindowWidthCalculator{}
}

// 1. Sort times
//
// 2. TargetAt (last reminder) gets fixed width:
//   - 2 minutes for short/scheduled types
//   - 5 minutes for near/relaxed types
//
// 3. Intermediate reminders get:
//
//   - 30% of interval to NEXT reminder
//
//     max width clamped per task type:
//
//   - 5 minutes for short type
//
//   - 10 minutes for other types
func (c *SlideWindowWidthCalculator) CalculateSlideWindowWidths(
	times []time.Time,
	taskType Type,
) map[time.Time]SlideWindowWidth {
	if len(times) == 0 {
		return nil
	}

	sortedTimes := make([]time.Time, len(times))
	copy(sortedTimes, times)
	sort.Slice(sortedTimes, func(i, j int) bool {
		return sortedTimes[i].Before(sortedTimes[j])
	})

	result := make(map[time.Time]SlideWindowWidth, len(times))
	lastIndex := len(sortedTimes) - 1

	for i, t := range sortedTimes {
		if i == lastIndex {
			// TargetAt (last reminder)
			result[t] = GetTargetAtWindowWidth(taskType)
		} else {
			// Intermediate reminder: 30% of interval to next reminder, clamped per task type
			result[t] = c.calculateIntermediateWidth(sortedTimes, i, taskType)
		}
	}

	return result
}

// CalculateSingleSlideWindowWidth calculates slide window width for a single reminder time.
func (c *SlideWindowWidthCalculator) CalculateSingleSlideWindowWidth(taskType Type) SlideWindowWidth {
	return GetTargetAtWindowWidth(taskType)
}

func (c *SlideWindowWidthCalculator) calculateIntermediateWidth(times []time.Time, idx int, taskType Type) SlideWindowWidth {
	if idx >= len(times)-1 {
		return MustSlideWindowWidth(MinSlideWindowWidth)
	}

	intervalToNext := times[idx+1].Sub(times[idx])

	return GetIntermediateWindowWidth(taskType, intervalToNext)
}
