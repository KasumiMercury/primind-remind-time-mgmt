package domain

import "time"

// GetTargetAtWindowWidth returns the window width for the TargetAt (last) reminder.
// Uses switch statement to determine width based on task type:
//   - short/scheduled: 2 minutes
//   - near/relaxed: 5 minutes
func GetTargetAtWindowWidth(taskType Type) SlideWindowWidth {
	switch taskType {
	case TypeShort, TypeScheduled:
		return MustSlideWindowWidth(WindowWidthShort)
	case TypeNear, TypeRelaxed:
		return MustSlideWindowWidth(WindowWidthBase)
	default:
		return MustSlideWindowWidth(WindowWidthBase)
	}
}

func GetIntermediateWindowWidth(taskType Type, intervalToNext time.Duration) SlideWindowWidth {
	// Calculate 30% of interval
	rawWidth := time.Duration(float64(intervalToNext) * IntermediateIntervalRatio)

	// Determine max width based on task type
	var maxWidth time.Duration
	switch taskType {
	case TypeShort:
		maxWidth = IntermediateMaxWidthShort
	default:
		maxWidth = MaxSlideWindowWidth
	}

	// Clamp
	if rawWidth < MinSlideWindowWidth {
		return MustSlideWindowWidth(MinSlideWindowWidth)
	}
	if rawWidth > maxWidth {
		return MustSlideWindowWidth(maxWidth)
	}
	return MustSlideWindowWidth(rawWidth)
}
