package main

import (
	"fmt"
	"time"
)

const (
	StateOpen int8 = iota
	StateClosed
	StateHalfOpen
)

type CircuitBreaker struct {
	state                 int8
	closedFailedCounter   int64
	halfOpenThreshold     int64
	halfOpenSuccess       int64
	closedFailedThreshold int64
	openTimeout           time.Duration
}

func NewCircuitBreaker(closedToOpenFailureThreshold int64, openTimeout time.Duration, halfOpenThreshold int64) *CircuitBreaker {
	return &CircuitBreaker{
		state:                 StateClosed,
		closedFailedCounter:   0,
		closedFailedThreshold: closedToOpenFailureThreshold,
		openTimeout:           openTimeout,
		halfOpenThreshold:     halfOpenThreshold,
		halfOpenSuccess:       0,
	}
}

func (c *CircuitBreaker) Call(f func() error) error {
	switch c.state {
	case StateOpen:
		return fmt.Errorf("circuit breaker is not closed")
	case StateClosed:
		err := f()
		if err == nil {
			return nil
		}
		c.closedFailedCounter++
		if c.closedFailedCounter >= c.closedFailedThreshold {
			c.state = StateOpen
			go func(c *CircuitBreaker) {
				timer := time.NewTimer(c.openTimeout)
				<-timer.C
				c.state = StateHalfOpen
			}(c)
		}
		return err

	case StateHalfOpen:
		err := f()
		if err != nil {
			c.state = StateOpen
			go func(c *CircuitBreaker) {
				timer := time.NewTimer(c.openTimeout)
				<-timer.C
				c.state = StateHalfOpen
			}(c)
			return err
		}
		c.halfOpenSuccess++
		if c.halfOpenSuccess >= c.halfOpenThreshold {
			c.state = StateClosed
		}
		return nil
	default:
		panic("unexpected state in circuit breaker")
	}
}
