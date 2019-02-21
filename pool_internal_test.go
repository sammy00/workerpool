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
		exec := Pool(c.n)

		pool := exec.(*pool)
		if cap(pool.pendings) != c.expect {
			t.Fatalf("#%d failed: got pool size as %d, expect %d",
				i, cap(pool.pendings), c.expect)
		}

		exec.Close()
	}
}

//func TestPool_Close(t *testing.T) {
//	pool := Pool(2)
//}
