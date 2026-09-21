//go:build test

package domain

import "testing"

func TestBeziqueMeldKeyAllBranches(t *testing.T) {
	tests := []struct {
		name string
		meld BeziqueMeld
		want string
	}{
		// 通常のマリッジは呼び出し元の if で別コードに振り分けられ、ここにはロイヤルのみ来る。
		{name: "royal marriage", meld: BeziqueMeld{Type: BeziqueMeldMarriage, Points: BeziqueRoyalMarriagePoints}, want: "bezique.meld.royalMarriage"},
		{name: "bezique", meld: BeziqueMeld{Type: BeziqueMeldBezique}, want: "bezique.meld.bezique"},
		{name: "four aces", meld: BeziqueMeld{Type: BeziqueMeldFourAces}, want: "bezique.meld.fourAces"},
		{name: "four kings", meld: BeziqueMeld{Type: BeziqueMeldFourKings}, want: "bezique.meld.fourKings"},
		{name: "four queens", meld: BeziqueMeld{Type: BeziqueMeldFourQueens}, want: "bezique.meld.fourQueens"},
		{name: "four jacks", meld: BeziqueMeld{Type: BeziqueMeldFourJacks}, want: "bezique.meld.fourJacks"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := beziqueMeldKey(tt.meld); got != tt.want {
				t.Errorf("beziqueMeldKey(%+v) = %q, want %q", tt.meld, got, tt.want)
			}
		})
	}
}
