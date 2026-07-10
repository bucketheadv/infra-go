package retryx

import (
	"context"
	"runtime"
	"time"
)

// Config 重试配置。
type Config struct {
	// MaxAttempts 最大尝试次数（含首次）；<= 0 时默认为 3。
	MaxAttempts int
	// Backoff 退避策略；nil 时不等待。
	Backoff Backoff
	// RetryIf 判断是否应重试；nil 时任意非 nil 错误均重试。
	RetryIf func(err error) bool
}

// DefaultConfig 返回默认重试配置（3 次、无退避）。
func DefaultConfig() Config {
	return Config{MaxAttempts: 3}
}

// Do 在 cfg 限制下执行 fn，失败时按退避策略重试。
// ctx 为 nil 时视为 context.Background()。
func Do(ctx context.Context, cfg Config, fn func() error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	maxAttempts := cfg.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	retryIf := cfg.RetryIf
	if retryIf == nil {
		retryIf = func(err error) bool { return err != nil }
	}
	backoff := cfg.Backoff

	var err error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err = fn()
		if err == nil {
			return nil
		}
		if !retryIf(err) || attempt == maxAttempts-1 {
			return err
		}
		if backoff != nil {
			if waitErr := sleep(ctx, backoff.Next(attempt)); waitErr != nil {
				return waitErr
			}
		}
	}
	return err
}

func sleep(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		// Fixed(0) 等零等待场景让出调度，避免纯忙等占满 CPU。
		runtime.Gosched()
		return ctx.Err()
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
