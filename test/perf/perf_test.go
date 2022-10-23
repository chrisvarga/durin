package perf

import "testing"

func BenchmarkPerfSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		PerfSet()
	}
}

func BenchmarkPerfGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		PerfGet()
	}
}
