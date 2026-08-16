package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPineapplePlayer(t *testing.T) {
	p := NewPineapplePlayer(true, HoldemStyleTAG)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, HoldemStyleTAG, p.GetPlayStyle())
	assert.Equal(t, "TAG", p.GetPlayStyleName())
	assert.Equal(t, 0, p.GetCardsSize())
}

func TestPineapplePlayer_HUDStats(t *testing.T) {
	p := NewPineapplePlayer(false, HoldemStyleLAG)

	assert.Equal(t, 0, p.GetVPIP())
	assert.Equal(t, 0, p.GetPFR())
	assert.Equal(t, 0, p.GetThreeBet())
	assert.Equal(t, "-", p.GetAFDisplay())

	p.IncrementTotalHands()
	p.IncrementTotalHands()
	p.IncrementVPIP()
	p.IncrementPFR()
	assert.Equal(t, 50, p.GetVPIP())
	assert.Equal(t, 50, p.GetPFR())

	p.IncrementThreeBetOpportunity()
	p.IncrementThreeBetOpportunity()
	p.IncrementThreeBet()
	assert.Equal(t, 50, p.GetThreeBet())

	p.IncrementPostFlopBetRaise()
	p.IncrementPostFlopBetRaise()
	assert.Equal(t, "∞", p.GetAFDisplay())

	p.IncrementPostFlopCall()
	assert.Equal(t, "2.0", p.GetAFDisplay())
}

func TestPineapplePlayer_EvalBestHand(t *testing.T) {
	t.Run("2 hole cards + 5 community = standard holdem eval", func(t *testing.T) {
		p := NewPineapplePlayer(true, HoldemStyleTAG)
		p.AddCard(NewCard(CardDesignSpade, 1, false))  // Ace spades
		p.AddCard(NewCard(CardDesignSpade, 13, false)) // King spades

		comm := []*Card{
			NewCard(CardDesignSpade, 12, false), // Q spades
			NewCard(CardDesignSpade, 11, false), // J spades
			NewCard(CardDesignSpade, 10, false), // 10 spades
			NewCard(CardDesignHeart, 2, false),  // 2 hearts
			NewCard(CardDesignHeart, 3, false),  // 3 hearts
		}
		rank := p.EvalBestHand(comm)
		assert.Equal(t, PokerHandRoyalFlush, rank)
		assert.NotNil(t, p.GetBestHand())
	})

	t.Run("3 hole cards + 3 community = best 5 of 6", func(t *testing.T) {
		p := NewPineapplePlayer(true, HoldemStyleTAG)
		p.AddCard(NewCard(CardDesignSpade, 1, false))  // Ace spades
		p.AddCard(NewCard(CardDesignSpade, 13, false)) // King spades
		p.AddCard(NewCard(CardDesignHeart, 2, false))  // 2 hearts (weak)

		comm := []*Card{
			NewCard(CardDesignSpade, 12, false), // Q spades
			NewCard(CardDesignSpade, 11, false), // J spades
			NewCard(CardDesignSpade, 10, false), // 10 spades
		}
		rank := p.EvalBestHand(comm)
		assert.Equal(t, PokerHandRoyalFlush, rank)
	})

	t.Run("less than 5 total cards returns HighCard", func(t *testing.T) {
		p := NewPineapplePlayer(true, HoldemStyleTAG)
		p.AddCard(NewCard(CardDesignSpade, 1, false))
		p.AddCard(NewCard(CardDesignSpade, 13, false))

		rank := p.EvalBestHand([]*Card{NewCard(CardDesignHeart, 2, false)})
		assert.Equal(t, PokerHandHighCard, rank)
		assert.Nil(t, p.GetBestHand())
	})
}

func TestPineapplePlayer_GetComparisonCards(t *testing.T) {
	p := NewPineapplePlayer(true, HoldemStyleTAG)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignSpade, 13, false))

	comm := []*Card{
		NewCard(CardDesignSpade, 12, false), NewCard(CardDesignSpade, 11, false), NewCard(CardDesignSpade, 10, false),
		NewCard(CardDesignHeart, 2, false), NewCard(CardDesignHeart, 3, false),
	}
	p.EvalBestHand(comm)

	cards := p.GetComparisonCards()
	assert.Equal(t, 5, len(cards))
	// Verify it's a copy
	cards[0] = nil
	assert.NotNil(t, p.GetComparisonCards()[0])
}

func TestPineapplePlayer_JSON(t *testing.T) {
	p := NewPineapplePlayer(true, HoldemStyleLAG)
	p.SetChips(500)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignSpade, 13, false))
	p.IncrementTotalHands()
	p.IncrementVPIP()

	data, err := json.Marshal(p)
	assert.NoError(t, err)

	p2 := &PineapplePlayer{}
	err = json.Unmarshal(data, p2)
	assert.NoError(t, err)

	assert.True(t, p2.GetIsHuman())
	assert.Equal(t, HoldemStyleLAG, p2.GetPlayStyle())
	assert.Equal(t, 500, p2.GetChips())
	assert.Equal(t, 1, p2.GetTotalHands())
	assert.Equal(t, 1, p2.GetVPIPCount())
	assert.Equal(t, 2, p2.GetCardsSize())
}

// **表示のために状態を書き換えない。** EvalBestHand を描画のたびに呼ぶと
// handRank / bestHand が動くので、途中経過の表示には PeekBestHand を使う
// (#5488、Omaha の #4680 と同じ形)。
func TestPineapplePlayer_PeekBestHandLeavesTheStateAlone(t *testing.T) {
	pp := NewPineapplePlayer(true, HoldemStyleTAG)
	pp.AddCard(NewCard(CardDesignSpade, 14, false))
	pp.AddCard(NewCard(CardDesignSpade, 13, false))
	pp.AddCard(NewCard(CardDesignHeart, 2, false))
	board := []*Card{
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignSpade, 11, false),
		NewCard(CardDesignSpade, 10, false),
	}

	beforeRank, beforeBest := pp.GetHandRank(), pp.GetComparisonCards()

	rank, best := pp.PeekBestHand(board)
	if rank != PokerHandStraightFlush {
		t.Errorf("PeekBestHand rank = %d, want straight flush (%d)", rank, PokerHandStraightFlush)
	}
	if len(best) != 5 {
		t.Errorf("PeekBestHand returned %d cards, want 5", len(best))
	}
	if pp.GetHandRank() != beforeRank {
		t.Errorf("handRank changed to %d (was %d) — Peek must not record", pp.GetHandRank(), beforeRank)
	}
	if len(pp.GetComparisonCards()) != len(beforeBest) {
		t.Error("bestHand changed — Peek must not record")
	}

	// EvalBestHand は同じ答えを返し、そのときだけ記録する。ここがずれると
	// 「表示とショーダウンで役が違う」になる。
	if got := pp.EvalBestHand(board); got != rank {
		t.Errorf("EvalBestHand = %d, PeekBestHand = %d — they must agree", got, rank)
	}
	if pp.GetHandRank() != rank {
		t.Errorf("EvalBestHand did not record: handRank = %d, want %d", pp.GetHandRank(), rank)
	}
}

// 5 枚に満たないうちは役を確定させない。ここで確定した組を返すと、フロップ前に
// 「ハイカード」以上の役が画面に出る。
func TestPineapplePlayer_PeekBestHandIsUndecidedBeforeFive(t *testing.T) {
	pp := NewPineapplePlayer(true, HoldemStyleTAG)
	pp.AddCard(NewCard(CardDesignSpade, 14, false))
	pp.AddCard(NewCard(CardDesignHeart, 9, false))

	rank, best := pp.PeekBestHand(nil)
	if rank != PokerHandHighCard || best != nil {
		t.Errorf("PeekBestHand(nil) = (%d, %v), want (high card, nil)", rank, best)
	}
}
