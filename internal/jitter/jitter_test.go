package jitter_test

import (
	"testing"
	"time"

	"github.com/driftwatch/driftwatch/internal/jitter"
)

func TestNew_PanicsOnZeroFactor(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero factor")
		}
	}()
	jitter.New(0)
}

func TestNew_PanicsOnNegativeFactor(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for negative factor")
		}
	}()
	jitter.New(-0.5)
}

func TestNew_PanicsOnFactorAboveOne(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for factor > 1")
		}
	}()
	jitter.New(1.1)
}

func TestApply_ResultWithinExpectedRange(t *testing.T) {
	t.Parallel()
	j := jitter.New(0.25)
	base := 100 * time.Millisecond

	for i := 0; i < 100; i++ {
		got := j.Apply(base)
		if got < base {
			t.Fatalf("Apply returned %v, want >= %v", got, base)
		}
		max := base + time.Duration(float64(base)*0.25)
		if got > max {
			t.Fatalf("Apply returned %v, want <= %v", got, max)
		}
	}
}

func TestApply_FactorOne_DoubleAtMost(t *testing.T) {
	t.Parallel()
	j := jitter.New(1.0)
	base := 50 * time.Millisecond

	for i := 0; i < 50; i++ {
		got := j.Apply(base)
		if got < base || got > 2*base {
			t.Fatalf("Apply(%v) = %v, want in [%v, %v]", base, got, base, 2*base)
		}
	}
}

func TestApplyRange_ResultWithinBounds(t *testing.T) {
	t.Parallel()
	j := jitter.New(0.5)
	min := 10 * time.Millisecond
	max := 50 * time.Millisecond

	for i := 0; i < 100; i++ {
		got := j.ApplyRange(min, max)
		if got < min || got >= max {
			t.Fatalf("ApplyRange returned %v, want in [%v, %v)", got, min, max)
		}
	}
}

func TestApplyRange_PanicsWhenMinEqualsMax(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when min == max")
		}
	}()
	j := jitter.New(0.5)
	j.ApplyRange(10*time.Millisecond, 10*time.Millisecond)
}
