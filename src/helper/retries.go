package helper

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Function signature of retryable function
type RetryableFunc func() error

// Function signature of retryable function with data
type RetryableFuncWithData[T any] func() (T, error)

// Default timer is a wrapper around time.After
type timerImpl struct{}

func (t *timerImpl) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func Do(retryableFunc RetryableFunc, opts ...Option) error {
	retryableFuncWithData := func() (any, error) {
		return nil, retryableFunc()
	}

	_, err := DoWithData(retryableFuncWithData, opts...)
	return err
}

func DoWithData[T any](retryableFunc RetryableFuncWithData[T], opts ...Option) (T, error) {
	var n uint
	var emptyT T

	// default
	config := newDefaultRetryConfig()

	// apply opts
	for _, opt := range opts {
		opt(config)
	}

	if err := config.context.Err(); err != nil {
		return emptyT, err
	}

	// Setting attempts to 0 means we'll retry until we succeed
	var lastErr error
	if config.attempts == 0 {
		for {
			t, err := retryableFunc()
			if err == nil {
				return t, nil
			}

			if !IsRecoverable(err) {
				return emptyT, err
			}

			if !config.retryIf(err) {
				return emptyT, err
			}

			lastErr = err

			n++
			config.onRetry(n, err)
			select {
			case <-config.timer.After(delay(config, n, err)):
			case <-config.context.Done():
				if config.wrapContextErrorWithLastError {
					return emptyT, Error{config.context.Err(), lastErr}
				}
				return emptyT, config.context.Err()
			}
		}
	}

	errorLog := Error{}

	attemptsForError := make(map[error]uint, len(config.attemptsForError))
	for err, attempts := range config.attemptsForError {
		attemptsForError[err] = attempts
	}

	shouldRetry := true
	for shouldRetry {
		t, err := retryableFunc()
		if err == nil {
			return t, nil
		}

		errorLog = append(errorLog, unpackUnrecoverable(err))

		if !config.retryIf(err) {
			break
		}

		config.onRetry(n, err)

		for errToCheck, attempts := range attemptsForError {
			if errors.Is(err, errToCheck) {
				attempts--
				attemptsForError[errToCheck] = attempts
				shouldRetry = shouldRetry && attempts > 0
			}
		}

		// if this is last attempt - don't wait
		if n == config.attempts-1 {
			break
		}

		select {
		case <-config.timer.After(delay(config, n, err)):
		case <-config.context.Done():
			if config.lastErrorOnly {
				return emptyT, config.context.Err()
			}

			return emptyT, append(errorLog, config.context.Err())
		}

		n++
		shouldRetry = shouldRetry && n < config.attempts
	}

	if config.lastErrorOnly {
		return emptyT, errorLog.Unwrap()
	}
	return emptyT, errorLog
}

func newDefaultRetryConfig() *Config {
	return &Config{
		attempts:         uint(10),
		attemptsForError: make(map[error]uint),
		delay:            100 * time.Millisecond,
		maxJitter:        100 * time.Millisecond,
		onRetry:          func(n uint, err error) {},
		retryIf:          IsRecoverable,
		delayType:        CombineDelay(BackOffDelay, RandomDelay),
		lastErrorOnly:    false,
		context:          context.Background(),
		timer:            &timerImpl{},
	}
}

// Error type represents list of errors in retry
type Error []error

// Error method return string representation of Error
// It is an implementation of error interface
func (e Error) Error() string {
	logWithNumber := make([]string, len(e))
	for i, l := range e {
		if l != nil {
			logWithNumber[i] = fmt.Sprintf("#%d: %s", i+1, l.Error())
		}
	}

	return fmt.Sprintf("All attempts fail:\n%s", strings.Join(logWithNumber, "\n"))
}

func (e Error) Is(target error) bool {
	for _, v := range e {
		if errors.Is(v, target) {
			return true
		}
	}
	return false
}

func (e Error) As(target interface{}) bool {
	for _, v := range e {
		if errors.As(v, target) {
			return true
		}
	}
	return false
}

/*
Unwrap the last error for compatibility with `errors.Unwrap()`.
When you need to unwrap all errors, you should use `WrappedErrors()` instead.

	err := Do(
		func() error {
			return errors.New("original error")
		},
		Attempts(1),
	)

	fmt.Println(errors.Unwrap(err)) # "original error" is printed

Added in version 4.2.0.
*/
func (e Error) Unwrap() error {
	return e[len(e)-1]
}

// WrappedErrors returns the list of errors that this Error is wrapping.
// It is an implementation of the `errwrap.Wrapper` interface
// in package [errwrap](https://github.com/hashicorp/errwrap) so that
// `retry.Error` can be used with that library.
func (e Error) WrappedErrors() []error {
	return e
}

type unrecoverableError struct {
	error
}

func (e unrecoverableError) Unwrap() error {
	return e.error
}

// Unrecoverable wraps an error in `unrecoverableError` struct
func Unrecoverable(err error) error {
	return unrecoverableError{err}
}

// IsRecoverable checks if error is an instance of `unrecoverableError`
func IsRecoverable(err error) bool {
	return !errors.Is(err, unrecoverableError{})
}

// Adds support for errors.Is usage on unrecoverableError
func (unrecoverableError) Is(err error) bool {
	_, isUnrecoverable := err.(unrecoverableError)
	return isUnrecoverable
}

func unpackUnrecoverable(err error) error {
	if unrecoverable, isUnrecoverable := err.(unrecoverableError); isUnrecoverable {
		return unrecoverable.error
	}

	return err
}

func delay(config *Config, n uint, err error) time.Duration {
	delayTime := config.delayType(n, err, config)
	if config.maxDelay > 0 && delayTime > config.maxDelay {
		delayTime = config.maxDelay
	}

	return delayTime
}
