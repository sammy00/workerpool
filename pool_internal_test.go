package workerpool

import (
	"runtime"
	"testing"
)

func TestPool(t *testing.T) {
	numCPU := runtime.NumCPU()

	testCases := []struct {
		n      int
		expect int
	}{
		{-2, numCPU},
		{-1, numCPU},
		{0, numCPU},
		{1, 1},
		{2, 2},
	}

	for i, c := range testCases {
		exec, cancel := Pool(c.n)
		defer cancel()

		pool := exec.(pool)
		if cap(pool.in) != c.expect {
			t.Fatalf("#%d failed: got pool size as %d, expect %d",
				i, cap(pool.in), c.expect)
		}
	}
}