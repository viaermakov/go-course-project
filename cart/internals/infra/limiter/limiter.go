package limiter

import (
	"context"
	"time"
)

// Limiter represents a rate limiter that allows a maximum number of requests to be executed within a given time period.
// Usage example:
//
//	l := NewLimiter(1*time.Second, 5) // Creates a limiter that allows 5 requests per second
//	l.Wait()                          // Blocks until a request can be executed
type Limiter struct {
	maxReq   uint
	throttle chan struct{}
}

func NewLimiter(t time.Duration, maxReq uint) (*Limiter, func()) {
	l := &Limiter{
		maxReq:   maxReq,
		throttle: make(chan struct{}, maxReq),
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		// don't wait for t to run first requests on initialization
		l.runBatch()
		timer := time.NewTicker(t)

		for range timer.C {
			select {
			case <-ctx.Done():
				timer.Stop()
				close(l.throttle)
				return // stop making requests if limiter was stopped using cancel()
			default:
				l.runBatch()
			}
		}
	}()

	return l, cancel
}

// Wait blocks until a request can be executed within the rate limit.
func (l *Limiter) Wait() {
	<-l.throttle
}

// runBatch refills the throttle channel to represent the maximum number of allowed requests.
func (l *Limiter) runBatch() {
	for range l.maxReq {
		l.throttle <- struct{}{}
	}
}
