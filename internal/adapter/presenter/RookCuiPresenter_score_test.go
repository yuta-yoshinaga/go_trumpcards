//go:build test

package presenter

import "testing"

func TestRookScoreDeltaStr(t *testing.T) {
	for _, tt := range []struct {
		delta int
		want  string
	}{
		{delta: 85, want: "+85"},
		{delta: 0, want: "0"},
		{delta: -80, want: "-80"},
	} {
		if got := rookScoreDeltaStr(tt.delta); got != tt.want {
			t.Errorf("rookScoreDeltaStr(%d) = %q, want %q", tt.delta, got, tt.want)
		}
	}
}
