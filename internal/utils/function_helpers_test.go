package utils

import (
	"errors"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Alfonsxh/gitlab-merge-alert-go/pkg/logger"
)

func TestMain(m *testing.M) {
	logger.Init("error")
	code := m.Run()
	os.Exit(code)
}

func TestExecuteWithTimeoutSuccess(t *testing.T) {
	res := ExecuteWithTimeout(time.Second, func() (int, error) {
		return 42, nil
	})

	if res.Error != nil {
		t.Fatalf("expected no error, got %v", res.Error)
	}
	if res.Value != 42 {
		t.Fatalf("expected value 42, got %d", res.Value)
	}
}

func TestExecuteWithTimeoutTimeout(t *testing.T) {
	res := ExecuteWithTimeout(10*time.Millisecond, func() (int, error) {
		time.Sleep(40 * time.Millisecond)
		return 0, nil
	})

	if res.Error == nil {
		t.Fatalf("expected timeout error, got nil")
	}
	if !errors.Is(res.Error, res.Error) && res.Error.Error() == "" {
		t.Fatalf("expected timeout error text, got empty")
	}
}

func TestExecuteWithRetrySuccessAfterRetries(t *testing.T) {
	var attempts int32
	value, err := ExecuteWithRetry[int](2, time.Millisecond, func() (int, error) {
		if atomic.AddInt32(&attempts, 1) < 3 {
			return 0, errors.New("transient error")
		}
		return 99, nil
	})

	if err != nil {
		t.Fatalf("expected success, got err=%v", err)
	}
	if value != 99 {
		t.Fatalf("expected value 99, got %d", value)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestExecuteWithRetryExhausted(t *testing.T) {
	var attempts int32
	_, err := ExecuteWithRetry[int](1, time.Millisecond, func() (int, error) {
		atomic.AddInt32(&attempts, 1)
		return 0, errors.New("still failing")
	})

	if err == nil {
		t.Fatalf("expected error after exhausting retries")
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestExecuteWithLogging(t *testing.T) {
	value, err := ExecuteWithLogging("success", func() (string, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if value != "ok" {
		t.Fatalf("expected value ok, got %s", value)
	}

	_, err = ExecuteWithLogging("failure", func() (string, error) {
		return "", errors.New("boom")
	})
	if err == nil {
		t.Fatalf("expected error from ExecuteWithLogging")
	}
}

func TestBatchExecute(t *testing.T) {
	items := []int{1, 2, 3}
	errs := BatchExecute(items, func(v int) error {
		if v == 2 {
			return errors.New("bad item")
		}
		return nil
	})

	if len(errs) != 3 {
		t.Fatalf("expected 3 results, got %d", len(errs))
	}
	if errs[0] != nil {
		t.Fatalf("expected first item nil error, got %v", errs[0])
	}
	if errs[1] == nil || errs[1].Error() != "项目 1 执行失败: bad item" {
		t.Fatalf("unexpected error for second item: %v", errs[1])
	}
}

func TestBatchExecuteParallel(t *testing.T) {
	items := []int{0, 1, 2, 3}
	start := time.Now()
	errs := BatchExecuteParallel(items, func(v int) error {
		time.Sleep(time.Duration(v+1) * 5 * time.Millisecond)
		if v%2 == 0 {
			return errors.New("even error")
		}
		return nil
	}, 2)

	if len(errs) != len(items) {
		t.Fatalf("expected %d results, got %d", len(items), len(errs))
	}
	if errs[0] == nil || errs[0].Error() != "even error" {
		t.Fatalf("expected error for first item, got %v", errs[0])
	}
	if errs[1] != nil {
		t.Fatalf("expected nil error for second item, got %v", errs[1])
	}

	elapsed := time.Since(start)
	if elapsed < 20*time.Millisecond {
		t.Fatalf("expected parallel execution with throttling, elapsed=%v", elapsed)
	}
}

func TestSafeExecute(t *testing.T) {
	value, err := SafeExecute(func() (int, error) {
		return 7, nil
	})
	if err != nil || value != 7 {
		t.Fatalf("expected success result, got value=%d err=%v", value, err)
	}

	_, err = SafeExecute(func() (int, error) {
		panic("boom")
	})
	if err == nil {
		t.Fatalf("expected panic to be captured as error")
	}
}

func TestConditionalExecute(t *testing.T) {
	value, err := ConditionalExecute(true, func() (string, error) {
		return "ok", nil
	})
	if err != nil || value != "ok" {
		t.Fatalf("expected successful execution, got value=%s err=%v", value, err)
	}

	_, err = ConditionalExecute(false, func() (string, error) {
		return "should not run", nil
	})
	if err == nil {
		t.Fatalf("expected error when condition is false")
	}
}

func TestCacheExecuteCachesValue(t *testing.T) {
	cache = make(map[string]*CacheEntry[interface{}])
	var calls int32

	first, err := CacheExecute("key", time.Minute, func() (int, error) {
		return int(atomic.AddInt32(&calls, 1)), nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if first != 1 {
		t.Fatalf("unexpected value %d", first)
	}

	second, err := CacheExecute("key", time.Minute, func() (int, error) {
		return int(atomic.AddInt32(&calls, 1)), nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if second != 1 {
		t.Fatalf("expected cached value 1, got %d", second)
	}
	if calls != 1 {
		t.Fatalf("expected function to run once, ran %d times", calls)
	}
}

func TestCacheExecuteExpires(t *testing.T) {
	cache = make(map[string]*CacheEntry[interface{}])
	var calls int32

	_, err := CacheExecute("expiring", -time.Second, func() (int, error) {
		return int(atomic.AddInt32(&calls, 1)), nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = CacheExecute("expiring", time.Minute, func() (int, error) {
		return int(atomic.AddInt32(&calls, 1)), nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected function to run twice due to expiry, ran %d times", calls)
	}
}
