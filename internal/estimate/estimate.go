package estimate

import (
	"fmt"
	"time"
)

const (
	minSpan = 10 * time.Minute
	minDrop = 2
	maxGap  = 15 * time.Minute
)

type sample struct {
	t time.Time
	p int
}

type Tracker struct {
	samples []sample
	window  time.Duration
}

func New(window time.Duration) *Tracker { return &Tracker{window: window} }

func (t *Tracker) Add(now time.Time, percent int, charging bool) {
	if charging {
		t.samples = nil
		return
	}
	if n := len(t.samples); n > 0 {
		last := t.samples[n-1]
		if percent > last.p || now.Sub(last.t) > maxGap {
			t.samples = nil
		}
	}
	t.samples = append(t.samples, sample{now, percent})

	cut := now.Add(-t.window)
	i := 0
	for i < len(t.samples)-1 && t.samples[i].t.Before(cut) {
		i++
	}
	t.samples = t.samples[i:]
}

func (t *Tracker) Remaining() (time.Duration, bool) {
	if len(t.samples) < 2 {
		return 0, false
	}
	first, last := t.samples[0], t.samples[len(t.samples)-1]
	span := last.t.Sub(first.t)
	drop := first.p - last.p
	if span < minSpan || drop < minDrop {
		return 0, false
	}
	hoursPerPercent := span.Hours() / float64(drop)
	return time.Duration(hoursPerPercent * float64(last.p) * float64(time.Hour)), true
}

func Format(d time.Duration) string {
	m := int(d.Round(time.Minute).Minutes())
	h, mm := m/60, m%60
	switch {
	case h > 99:
		return ">99 ч"
	case h == 0:
		return fmt.Sprintf("%d мин", mm)
	default:
		return fmt.Sprintf("%d ч %02d мин", h, mm)
	}
}
