package presenter

import "testing"

func TestEncodeIndexedScores(t *testing.T) {
	tests := []struct {
		name  string
		n     int
		score func(int) (int, bool)
		want  string
	}{
		{
			name: "skips seats and joins scores",
			n:    4,
			score: func(i int) (int, bool) {
				return []int{12, 0, -3, 8}[i], i != 1
			},
			want: "0:12,2:-3,3:8",
		},
		{
			name:  "zero players",
			n:     0,
			score: func(int) (int, bool) { return 0, true },
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := encodeIndexedScores(tt.n, tt.score); got != tt.want {
				t.Fatalf("encodeIndexedScores() = %q, want %q", got, tt.want)
			}
		})
	}
}
