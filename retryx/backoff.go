package retryx

import (
	"math/rand/v2"
	"time"
)

// Backoff 根据失败次数（从 0 开始）返回下次重试前的等待时长。
type Backoff interface {
	Next(attempt int) time.Duration
}

type fixedBackoff struct {
	// delay 固定等待时长。
	delay time.Duration
}

// Fixed 返回固定间隔退避；delay <= 0 时不等待。
func Fixed(delay time.Duration) Backoff {
	return fixedBackoff{delay: delay}
}

func (b fixedBackoff) Next(attempt int) time.Duration {
	return b.delay
}

type exponentialBackoff struct {
	// initial 首次等待时长。
	initial time.Duration
	// max 等待时长上限。
	max time.Duration
	// jitter 为 true 时在 [delay/2, delay] 内随机。
	jitter bool
}

// Exponential 返回指数退避（倍率 2）；initial 为首次等待时长，max 为上限。
func Exponential(initial, max time.Duration) Backoff {
	return ExponentialWithJitter(initial, max, false)
}

// ExponentialWithJitter 返回指数退避；jitter 为 true 时在 [delay/2, delay] 内随机。
func ExponentialWithJitter(initial, max time.Duration, jitter bool) Backoff {
	if initial <= 0 {
		initial = time.Millisecond
	}
	if max <= 0 {
		max = initial
	}
	return exponentialBackoff{initial: initial, max: max, jitter: jitter}
}

func (b exponentialBackoff) Next(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	delay := b.initial
	for i := 0; i < attempt; i++ {
		if delay >= b.max {
			delay = b.max
			break
		}
		if delay > b.max/2 {
			delay = b.max
			break
		}
		delay *= 2
	}
	if delay > b.max {
		delay = b.max
	}
	if !b.jitter || delay <= 0 {
		return delay
	}
	half := delay / 2
	return half + time.Duration(rand.Int64N(int64(delay-half+1)))
}
