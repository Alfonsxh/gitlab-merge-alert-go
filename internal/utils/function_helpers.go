package utils

import (
	"fmt"
	"time"

	"gitlab-merge-alert-go/pkg/logger"
)

// Result 函数执行结果
type Result[T any] struct {
	Value T
	Error error
}

// ExecuteWithTimeout 带超时执行函数
func ExecuteWithTimeout[T any](timeout time.Duration, fn func() (T, error)) *Result[T] {
	result := make(chan *Result[T], 1)

	go func() {
		value, err := fn()
		result <- &Result[T]{Value: value, Error: err}
	}()

	select {
	case res := <-result:
		return res
	case <-time.After(timeout):
		var zero T
		return &Result[T]{
			Value: zero,
			Error: fmt.Errorf("操作超时，超过 %v", timeout),
		}
	}
}

// ExecuteWithRetry 带重试执行函数
func ExecuteWithRetry[T any](maxRetries int, retryDelay time.Duration, fn func() (T, error)) (T, error) {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			logger.GetLogger().Warnf("第 %d 次重试执行函数", i)
			time.Sleep(retryDelay)
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err
		logger.GetLogger().Warnf("函数执行失败 (尝试 %d/%d): %v", i+1, maxRetries+1, err)
	}

	var zero T
	return zero, fmt.Errorf("函数执行失败，已重试 %d 次: %w", maxRetries, lastErr)
}

// ExecuteWithLogging 带日志记录执行函数
func ExecuteWithLogging[T any](operation string, fn func() (T, error)) (T, error) {
	start := time.Now()
	logger.GetLogger().Debugf("开始执行: %s", operation)

	result, err := fn()
	duration := time.Since(start)

	if err != nil {
		logger.GetLogger().Errorf("执行失败: %s (耗时: %v, 错误: %v)", operation, duration, err)
	} else {
		logger.GetLogger().Debugf("执行成功: %s (耗时: %v)", operation, duration)
	}

	return result, err
}

// BatchExecute 批量执行函数
func BatchExecute[T any](items []T, fn func(T) error) []error {
	var errors []error

	for i, item := range items {
		if err := fn(item); err != nil {
			errors = append(errors, fmt.Errorf("项目 %d 执行失败: %w", i, err))
		} else {
			errors = append(errors, nil)
		}
	}

	return errors
}

// BatchExecuteParallel 并行批量执行函数
func BatchExecuteParallel[T any](items []T, fn func(T) error, maxConcurrency int) []error {
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	semaphore := make(chan struct{}, maxConcurrency)
	results := make([]error, len(items))
	done := make(chan struct{})

	for i, item := range items {
		go func(index int, item T) {
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			results[index] = fn(item)
			if index == len(items)-1 {
				close(done)
			}
		}(i, item)
	}

	<-done // 等待所有任务完成
	return results
}

// SafeExecute 安全执行函数（捕获panic）
func SafeExecute[T any](fn func() (T, error)) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.GetLogger().Errorf("函数执行时发生panic: %v", r)
			err = fmt.Errorf("函数执行时发生panic: %v", r)
		}
	}()

	return fn()
}

// ConditionalExecute 条件执行函数
func ConditionalExecute[T any](condition bool, fn func() (T, error)) (T, error) {
	if !condition {
		var zero T
		return zero, fmt.Errorf("执行条件不满足")
	}
	return fn()
}

// CacheExecute 带缓存执行函数
type CacheEntry[T any] struct {
	Value     T
	ExpiresAt time.Time
}

var cache = make(map[string]*CacheEntry[interface{}])

func CacheExecute[T any](key string, ttl time.Duration, fn func() (T, error)) (T, error) {
	// 检查缓存
	if entry, exists := cache[key]; exists {
		if time.Now().Before(entry.ExpiresAt) {
			if value, ok := entry.Value.(T); ok {
				logger.GetLogger().Debugf("缓存命中: %s", key)
				return value, nil
			}
		}
		// 缓存过期，删除
		delete(cache, key)
	}

	// 执行函数
	result, err := fn()
	if err != nil {
		return result, err
	}

	// 存储到缓存
	cache[key] = &CacheEntry[interface{}]{
		Value:     interface{}(result),
		ExpiresAt: time.Now().Add(ttl),
	}

	logger.GetLogger().Debugf("缓存存储: %s", key)
	return result, nil
}
