package roundrobin

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/matryer/is"
)

func TestRoundRobin(t *testing.T) {
	is := is.New(t)
	var a, b, c, d int64
	options := []string{"a", "b", "c", "d"}

	rr := New(options)
	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			pick := rr.Pick()
			switch pick {
			case "a":
				atomic.AddInt64(&a, 1)
			case "b":
				atomic.AddInt64(&b, 1)
			case "c":
				atomic.AddInt64(&c, 1)
			case "d":
				atomic.AddInt64(&d, 1)
			default:
				t.Error("invalid pick:", pick)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	for _, n := range []int64{a, b, c, d} {
		requireWithinRange(t, n, 24, 26)
	}
	is.Equal(int64(100), a+b+c+d)
}

func requireWithinRange(t *testing.T, n, min, max int64) {
	t.Helper()
	is := is.New(t)
	is.True(n >= min) // n should be at least min
	is.True(n <= max) // n should be at most max
}
