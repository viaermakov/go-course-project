package limiter

import (
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"sync/atomic"
	"testing"
	"time"
)

type inputData struct {
	maxRequests       int
	totalRequests     int
	interval          time.Duration
	wantExecutedTasks int
	after             time.Duration
}

func TestLimiter_NumberOfRequests(t *testing.T) {
	defer goleak.VerifyNone(t)

	tests := []struct {
		name      string
		inputData inputData
	}{
		{
			"should run first tasks immediately",
			inputData{
				maxRequests:       2,
				totalRequests:     10,
				interval:          1 * time.Second,
				after:             10 * time.Millisecond,
				wantExecutedTasks: 2,
			},
		},
		{
			"should run 6 tasks after 1sec",
			inputData{
				maxRequests:       2,
				totalRequests:     10,
				interval:          500 * time.Millisecond,
				after:             1050 * time.Millisecond,
				wantExecutedTasks: 6,
			},
		},
		{
			"should run all tasks at once",
			inputData{
				maxRequests:       10,
				totalRequests:     6,
				interval:          500 * time.Millisecond,
				after:             10 * time.Millisecond,
				wantExecutedTasks: 6,
			},
		},
	}

	for _, test := range tests {
		test := test
		var gotSuccessfulRequests atomic.Int64

		limiter, cancel := NewLimiter(test.inputData.interval, uint(test.inputData.maxRequests))

		go func() {
			_ = time.AfterFunc(test.inputData.after, func() {
				require.Equal(t, int64(test.inputData.wantExecutedTasks), gotSuccessfulRequests.Load())
			})
		}()

		for range test.inputData.totalRequests {
			limiter.Wait()

			go func() {
				gotSuccessfulRequests.Add(1)
			}()
		}

		cancel()
		time.Sleep(100 * time.Millisecond)
	}
}
