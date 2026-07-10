package retryx

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDoSuccess(t *testing.T) {
	calls := 0
	err := Do(context.Background(), DefaultConfig(), func() error {
		calls++
		return nil
	})
	if err != nil || calls != 1 {
		t.Fatalf("Do() calls=%d err=%v", calls, err)
	}
}

func TestDoNilContext(t *testing.T) {
	err := Do(nil, DefaultConfig(), func() error { return nil })
	if err != nil {
		t.Fatalf("nil context should be treated as Background: %v", err)
	}
}

func TestDoRetryThenSuccess(t *testing.T) {
	errSentinel := errors.New("retry")
	calls := 0
	cfg := Config{
		MaxAttempts: 3,
		Backoff:     Fixed(0),
	}
	err := Do(context.Background(), cfg, func() error {
		calls++
		if calls < 3 {
			return errSentinel
		}
		return nil
	})
	if err != nil || calls != 3 {
		t.Fatalf("Do() calls=%d err=%v", calls, err)
	}
}

func TestDoExhausted(t *testing.T) {
	errSentinel := errors.New("fail")
	calls := 0
	err := Do(context.Background(), Config{MaxAttempts: 2, Backoff: Fixed(0)}, func() error {
		calls++
		return errSentinel
	})
	if !errors.Is(err, errSentinel) || calls != 2 {
		t.Fatalf("Do() calls=%d err=%v", calls, err)
	}
}

func TestDoContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Do(ctx, DefaultConfig(), func() error {
		return errors.New("fail")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Do() err=%v", err)
	}
}

func TestDoRetryIf(t *testing.T) {
	calls := 0
	errPermanent := errors.New("permanent")
	cfg := Config{
		MaxAttempts: 5,
		Backoff:     Fixed(0),
		RetryIf: func(err error) bool {
			return !errors.Is(err, errPermanent)
		},
	}
	err := Do(context.Background(), cfg, func() error {
		calls++
		return errPermanent
	})
	if !errors.Is(err, errPermanent) || calls != 1 {
		t.Fatalf("Do() calls=%d err=%v", calls, err)
	}
}

func TestExponentialBackoff(t *testing.T) {
	b := Exponential(10*time.Millisecond, 50*time.Millisecond)
	if got := b.Next(0); got != 10*time.Millisecond {
		t.Fatalf("Next(0) = %v", got)
	}
	if got := b.Next(1); got != 20*time.Millisecond {
		t.Fatalf("Next(1) = %v", got)
	}
	if got := b.Next(5); got != 50*time.Millisecond {
		t.Fatalf("Next(5) = %v", got)
	}
}
