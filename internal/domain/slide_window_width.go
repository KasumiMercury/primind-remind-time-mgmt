package domain

import (
	"errors"
	"time"
)

type SlideWindowWidth struct {
	duration time.Duration
}

const (
	MinSlideWindowWidth = 1 * time.Minute
	MaxSlideWindowWidth = 10 * time.Minute

	WindowWidthShort = 2 * time.Minute
	WindowWidthBase  = 5 * time.Minute

	IntermediateMaxWidthShort = 5 * time.Minute

	IntermediateIntervalRatio = 0.30
)

var (
	ErrSlideWindowWidthTooSmall = errors.New("slide window width must be at least 1 minute")
	ErrSlideWindowWidthTooLarge = errors.New("slide window width must not exceed 10 minutes")
)

func NewSlideWindowWidth(d time.Duration) (SlideWindowWidth, error) {
	if d < MinSlideWindowWidth {
		return SlideWindowWidth{}, ErrSlideWindowWidthTooSmall
	}

	if d > MaxSlideWindowWidth {
		return SlideWindowWidth{}, ErrSlideWindowWidthTooLarge
	}

	return SlideWindowWidth{duration: d}, nil
}

func MustSlideWindowWidth(d time.Duration) SlideWindowWidth {
	w, err := NewSlideWindowWidth(d)
	if err != nil {
		panic(err)
	}

	return w
}

func SlideWindowWidthFromSeconds(seconds int32) (SlideWindowWidth, error) {
	return NewSlideWindowWidth(time.Duration(seconds) * time.Second)
}

func (w SlideWindowWidth) Duration() time.Duration {
	return w.duration
}

func (w SlideWindowWidth) Seconds() int32 {
	return int32(w.duration / time.Second) // #nosec G115
}

func (w SlideWindowWidth) IsZero() bool {
	return w.duration == 0
}
