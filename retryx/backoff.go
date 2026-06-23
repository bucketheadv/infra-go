package retryx

import "time"

// Backoff 根据失败次数（从 0 开始）返回下次重试前的等待时长。
type Backoff interface {
	Next(attempt int) time.Duration
}

type fixedBackoff struct {
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
	initial time.Duration
	max     time.Duration
}

// Exponential 返回指数退避（倍率 2）；initial 为首次等待时长，max 为上限。
func Exponential(initial, max time.Duration) Backoff {
	if initial <= 0 {
		initial = time.Millisecond
	}
	if max <= 0 {
		max = initial
	}
	return exponentialBackoff{initial: initial, max: max}
}

func (b exponentialBackoff) Next(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	delay := b.initial
	for i := 0; i < attempt; i++ {
		if delay >= b.max {
			return b.max
		}
		if delay > b.max/2 {
			return b.max
		}
		delay *= 2
	}
	if delay > b.max {
		return b.max
	}
	return delay
}
