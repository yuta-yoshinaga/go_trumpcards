//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

func TestTichuPlayerAccessors(t *testing.T) {
	p := NewTichuPlayer(true)
	if !p.GetIsHuman() {
		t.Error("should be human")
	}
	if p.GetRank() != 0 || p.GetDeclType() != TichuDeclNone {
		t.Error("fresh player should have rank 0 and no declaration")
	}
	p.SetRank(2)
	p.SetDeclType(TichuDeclGrand)
	if p.GetRank() != 2 || p.GetDeclType() != TichuDeclGrand {
		t.Error("setters failed")
	}
	p.AddCollected([]*Card{tcNorm(5, CardDesignSpade), tcDragon()})
	if TichuCardsPoints(p.GetCollected()) != 30 {
		t.Errorf("collected points = %d, want 30", TichuCardsPoints(p.GetCollected()))
	}
}

func TestTichuPlayerSortByStrength(t *testing.T) {
	p := NewTichuPlayer(false)
	p.AddCard(tcDragon())
	p.AddCard(tcNorm(3, CardDesignSpade))
	p.AddCard(tcMahjong())
	p.AddCard(tcNorm(13, CardDesignHeart))
	p.SortCardsByStrength()
	// weakest first: mahjong, 3, king, dragon
	if tichuSpecialKind(p.GetCard(0)) != TichuMahjong {
		t.Error("mahjong should sort first")
	}
	if tichuSpecialKind(p.GetCard(p.GetCardsSize()-1)) != TichuDragon {
		t.Error("dragon should sort last")
	}
}

func TestTichuPlayerJSON(t *testing.T) {
	p := NewTichuPlayer(true)
	p.AddCard(tcNorm(7, CardDesignSpade))
	p.SetRank(1)
	p.SetDeclType(TichuDeclTichu)
	p.AddCollected([]*Card{tcNorm(10, CardDesignHeart)})

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var p2 TichuPlayer
	if err := json.Unmarshal(data, &p2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p2.GetRank() != 1 || p2.GetDeclType() != TichuDeclTichu {
		t.Error("round-trip lost rank/declaration")
	}
	if p2.GetCardsSize() != 1 || len(p2.GetCollected()) != 1 {
		t.Error("round-trip lost cards/collected")
	}

	// nil GamePlayer fallback
	var p3 TichuPlayer
	if err := json.Unmarshal([]byte(`{}`), &p3); err != nil {
		t.Fatalf("empty unmarshal: %v", err)
	}
	if p3.GamePlayer == nil {
		t.Error("GamePlayer should default to non-nil")
	}
}
