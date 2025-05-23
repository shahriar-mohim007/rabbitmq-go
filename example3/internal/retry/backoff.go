package retry

import (
	"math"
	"time"
)

func ExponentialBackoff(attempt int) time.Duration {
	baseDelay := 1 * time.Second
	return time.Duration(math.Pow(2, float64(attempt))) * baseDelay
}
