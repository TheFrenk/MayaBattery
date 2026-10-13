package estimate

import (
	"math"
	"testing"
	"time"
)

func TestRemaining(t *testing.T) {
	tr := New(6 * time.Hour)
	start := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	for i := 0; i <= 6; i++ {
		tr.Add(start.Add(time.Duration(i)*10*time.Minute), 80-i, false)
	}
	rem, ok := tr.Remaining()
	if !ok || math.Abs(rem.Hours()-74.0/6.0) > 0.01 {
		t.Fatalf("got %v, %v", rem, ok)
	}
}

func TestChargingResets(t *testing.T) {
	tr := New(6 * time.Hour)
	now := time.Now()
	tr.Add(now, 50, false)
	tr.Add(now.Add(time.Minute), 50, true)
	if _, ok := tr.Remaining(); ok {
		t.Error("при зарядке оценки быть не должно")
	}
}

func TestFormat(t *testing.T) {
	if got := Format(14*time.Hour + 20*time.Minute); got != "14 ч 20 мин" {
		t.Errorf("got %q", got)
	}
	if got := Format(45 * time.Minute); got != "45 мин" {
		t.Errorf("got %q", got)
	}
}
