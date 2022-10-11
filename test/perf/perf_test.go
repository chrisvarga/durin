package token

import "testing"

func BenchmarkPerfDurin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		PerfDurin()
	}
}
func BenchmarkPerfErebor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		PerfErebor()
	}
}
func BenchmarkPerfRedis(b *testing.B) {
	for i := 0; i < b.N; i++ {
		PerfRedis()
	}
}
