package token

import "testing"

func BenchmarkParse1(b *testing.B) {
	s := "set foo bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar"
	for i := 0; i < b.N; i++ {
		Parse1(s)
	}
}

func BenchmarkParse2(b *testing.B) {
	s := "set foo bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar"
	for i := 0; i < b.N; i++ {
		Parse2(s)
	}
}

func BenchmarkParse3(b *testing.B) {
	s := "set foo bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar bar"
	for i := 0; i < b.N; i++ {
		Parse3(s)
	}
}
