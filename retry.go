package main

import (
	"fmt"
	"math"
	"time"
)

var RetryExhaustedError error = fmt.Errorf("Function never succeeded in Retry")

func RetryExhausted(e error) bool {
	_, ok := e.(RetryExhaustedError)
	return ok
}

// Retriable takes 0-indexed iteration and returns (done, error)
type Retryable func(uint) (bool, error)

type RetryWait func(uint) (time.Duration, error)

func ExponentialBackoff(uint i) (time.Duration, error) {
	return math.Pow(2, i) * time.Second, nil
}

func LinearBackoff(uint i, unit time.Duration) (time.Duration, error) {
	return i * unit, nil
}

func LinearBackoffSecond(uint i) (time.Duration, error) {
	return LinearBackoff(i, time.Second)
}

func LinearBackoffMillisecond(uint i) (time.Duration, error) {
	return LinearBackoff(i, time.Millisecond)
}

func MaxTries(maxIterations uint, r Retryable) Retryable {
	return func(i uint) (time.Duration, error) {
		if i > maxIterations {
			return nil, RetryExhaustedError
		}
		return r(i)
	}
}

func MaxInterval(i time.Duration, r Retryable) Retryable {
	d, err := r(i)
	if d > i {
		return i, err
	}
	return d, err
}

func MinInterval(i time.Duration, r Retryable) Retryable {
	d, err := r(i)
	if d < i {
		return i, err
	}
	return d, err
}

type Retrier struct {
	waitF  RetryWait
	f      Retryable
	errors []error
}

func NewRetrier(waitF RetryWait, f Retryable) *Retrier {
	return &Retrier{
		waitF,
		f,
	}
}

// Do retriable function r. If r returns an error, add it to the error list for
// the iteration.
// Stops iteration if r returns done
func (r *Retrier) Do() error {
	for i := 0; ; i++ {
		sleep, err := r.waitF(i)
		if err != nil {
			return err
		}
		time.Sleep(sleep)
		done, err := r.f(i)
		r.errors = append(r.errors, err)
		if done {
			return err
		}
	}
}
