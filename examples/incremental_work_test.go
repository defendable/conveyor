package examples

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	numIncrementalWork10000 = 499999995000
	numIncrementalWork1000  = 499999500
	numIncremetnalWork100   = 499950
)

func BenchmarkTestIncrementalWork_100_100_100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(100, 100, 100)
		assert.Equal(b, numIncremetnalWork100, result)
	}
}

func BenchmarkTestIncrementalWork_100_50_50(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(100, 50, 50)
		assert.Equal(b, numIncremetnalWork100, result)
	}
}

func BenchmarkTestIncrementalWork_100_25_25(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(100, 25, 25)
		assert.Equal(b, numIncremetnalWork100, result)
	}
}

func BenchmarkTestIncrementalWork_100_8_8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(100, 8, 8)
		assert.Equal(b, numIncremetnalWork100, result)
	}
}

func BenchmarkTestIncrementalWork_100_4_4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(100, 4, 4)
		assert.Equal(b, numIncremetnalWork100, result)
	}
}

func BenchmarkTestIncrementalWork_100_2_2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(100, 2, 2)
		assert.Equal(b, numIncremetnalWork100, result)
	}
}

func BenchmarkTestIncrementalWork_100_1_1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(100, 1, 1)
		assert.Equal(b, numIncremetnalWork100, result)
	}
}

func BenchmarkTestIncrementalWork_1000_100_100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(1000, 100, 100)
		assert.Equal(b, numIncrementalWork1000, result)
	}
}

func BenchmarkTestIncrementalWork_1000_50_50(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(1000, 50, 50)
		assert.Equal(b, numIncrementalWork1000, result)
	}
}

func BenchmarkTestIncrementalWork_1000_25_25(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(1000, 25, 25)
		assert.Equal(b, numIncrementalWork1000, result)
	}
}

func BenchmarkTestIncrementalWork_1000_8_8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(1000, 8, 8)
		assert.Equal(b, numIncrementalWork1000, result)
	}
}

func BenchmarkTestIncrementalWork_1000_4_4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(1000, 4, 4)
		assert.Equal(b, numIncrementalWork1000, result)
	}
}

func BenchmarkTestIncrementalWork_1000_2_2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(1000, 2, 2)
		assert.Equal(b, numIncrementalWork1000, result)
	}
}

func BenchmarkTestIncrementalWork_1000_1_1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(1000, 1, 1)
		assert.Equal(b, numIncrementalWork1000, result)
	}
}

func BenchmarkTestIncrementalWork_10000_100_200(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(10000, 100, 200)
		assert.Equal(b, numIncrementalWork10000, result)
	}
}

func BenchmarkTestIncrementalWork_10000_50_100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(10000, 50, 100)
		assert.Equal(b, numIncrementalWork10000, result)
	}
}

func BenchmarkTestIncrementalWork_10000_25_50(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(10000, 25, 50)
		assert.Equal(b, numIncrementalWork10000, result)
	}
}

func BenchmarkTestIncrementalWork_10000_8_16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(10000, 8, 16)
		assert.Equal(b, numIncrementalWork10000, result)
	}
}

func BenchmarkTestIncrementalWork_10000_4_8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(10000, 4, 8)
		assert.Equal(b, numIncrementalWork10000, result)
	}
}

func BenchmarkTestIncrementalWork_10000_2_4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(10000, 2, 4)
		assert.Equal(b, numIncrementalWork10000, result)
	}
}

func BenchmarkTestIncrementalWork_10000_1_2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := incrementalWork(10000, 1, 2)
		assert.Equal(b, numIncrementalWork10000, result)
	}
}
