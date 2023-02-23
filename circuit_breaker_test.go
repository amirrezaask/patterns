package main

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func assert(t *testing.T, expected any, actual any) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("%+v != %+v", expected, actual)
	}
}

func TestCircuitBreaker(t *testing.T) {
	t.Run("circuit breaker should open when reached threshold", func(t *testing.T) {
		cb := NewCircuitBreaker(2, time.Duration(time.Second), 1)
		cb.Call(func() error {
			return errors.New("FAIL")
		})
		cb.Call(func() error {
			return errors.New("FAIL")
		})
		assert(t, StateOpen, cb.state)
	})

	t.Run("circuit breaker should become half open after reaching open timeout", func(t *testing.T) {
		cb := NewCircuitBreaker(1, time.Millisecond*500, 1)
		cb.Call(func() error {
			return errors.New("FAIL")
		})
		assert(t, StateOpen, cb.state)
		time.Sleep(time.Second * 1)
		assert(t, StateHalfOpen, cb.state)
	})

	t.Run("circuit breaker should become half open after reaching open timeout and then with fail should go back to open", func(t *testing.T) {
		cb := NewCircuitBreaker(1, time.Millisecond*500, 1)
		cb.Call(func() error {
			return errors.New("FAIL")
		})
		assert(t, StateOpen, cb.state)
		time.Sleep(time.Second * 1)
		assert(t, StateHalfOpen, cb.state)
		cb.Call(func() error {
			return errors.New("another error")
		})
		assert(t, StateOpen, cb.state)
	})
	t.Run("circuit breaker should become half open after reaching open timeout and then with success should go back to closed", func(t *testing.T) {
		cb := NewCircuitBreaker(1, time.Millisecond*500, 1)
		cb.Call(func() error {
			return errors.New("FAIL")
		})
		assert(t, StateOpen, cb.state)
		time.Sleep(time.Second * 1)
		assert(t, StateHalfOpen, cb.state)
		cb.Call(func() error {
			return nil
		})
		assert(t, StateClosed, cb.state)
	})
}
