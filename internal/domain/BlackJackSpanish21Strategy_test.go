//go:build test

package domain_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func s21Hand(values ...int) *domain.BlackJackHand {
	h := domain.NewBlackJackHand()
	for i, v := range values {
		design := domain.CardDesignSpade
		if i%2 == 1 {
			design = domain.CardDesignHeart
		}
		h.AddCard(domain.NewCard(design, v, false))
	}
	return h
}

func up(v int) *domain.Card { return domain.NewCard(domain.CardDesignClover, v, false) }

// TestSpanish21Strategy_DiffersFromStandard pins the cells where the 48-card
// deck changes the right play (#4705).
//
// **標準表と同じ答えしか固定しないテストは、配線が外れていても通る。**
// ここは「標準ならこう、スパニッシュ21ならこう」と両方を突き合わせる。
func TestSpanish21Strategy_DiffersFromStandard(t *testing.T) {
	tests := []struct {
		name     string
		hand     *domain.BlackJackHand
		upcard   *domain.Card
		standard domain.BJSuggestedAction
		spanish  domain.BJSuggestedAction
	}{
		{
			// 10 が抜けているぶんバストしにくいので、止まらず引く。
			"hard 12 vs 4: stand becomes hit",
			s21Hand(9, 3), up(4), domain.BJSuggestStand, domain.BJSuggestHit,
		},
		{
			"hard 13 vs 6: stand becomes hit",
			s21Hand(9, 4), up(6), domain.BJSuggestStand, domain.BJSuggestHit,
		},
		{
			"hard 14 vs 3: stand becomes hit",
			s21Hand(9, 5), up(3), domain.BJSuggestStand, domain.BJSuggestHit,
		},
		{
			// 11 に対して 10 を引ける確率が下がるので、高いアップカードには倍賭けしない。
			"hard 11 vs 9: double becomes hit",
			s21Hand(7, 4), up(9), domain.BJSuggestDouble, domain.BJSuggestHit,
		},
		{
			"hard 10 vs 9: double becomes hit",
			s21Hand(6, 4), up(9), domain.BJSuggestDouble, domain.BJSuggestHit,
		},
		{
			"soft 13 vs 5: double becomes hit",
			s21Hand(1, 2), up(5), domain.BJSuggestDouble, domain.BJSuggestHit,
		},
		{
			"hard 17 vs A: stand becomes surrender",
			s21Hand(9, 8), up(1), domain.BJSuggestStand, domain.BJSuggestSurrender,
		},
		{
			"pair 4,4 vs 5: split becomes hit",
			s21Hand(4, 4), up(5), domain.BJSuggestSplit, domain.BJSuggestHit,
		},
		{
			"pair 6,6 vs 3: split becomes hit",
			s21Hand(6, 6), up(3), domain.BJSuggestSplit, domain.BJSuggestHit,
		},
		{
			"pair 2,2 vs 8: hit becomes split",
			s21Hand(2, 2), up(8), domain.BJSuggestHit, domain.BJSuggestSplit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStandard := domain.GetBasicStrategyAction(tt.hand, tt.upcard, false)
			gotSpanish := domain.GetSpanish21StrategyAction(tt.hand, tt.upcard)
			assert.Equal(t, tt.standard, gotStandard, "standard deck advice changed unexpectedly")
			assert.Equal(t, tt.spanish, gotSpanish, "Spanish 21 advice")
		})
	}
}

// TestGetVariantStrategyAction_Dispatches confirms the variant actually reaches
// the table — the wiring the issue is about.
func TestGetVariantStrategyAction_Dispatches(t *testing.T) {
	hand := s21Hand(9, 3) // hard 12
	upcard := up(4)

	// スパニッシュ21ならヒット、それ以外は標準表のスタンド。
	assert.Equal(t, domain.BJSuggestHit,
		domain.GetVariantStrategyAction(hand, upcard, false, domain.BJVariantSpanish21))
	assert.Equal(t, domain.BJSuggestStand,
		domain.GetVariantStrategyAction(hand, upcard, false, domain.BJVariantStandard))
}

// TestSpanish21Game_SuggestionUsesTheVariantTable drives the real game object so
// the config → suggestion path is covered, not just the table function.
func TestSpanish21Game_SuggestionUsesTheVariantTable(t *testing.T) {
	forVariant := func(newGame func() *domain.BlackJack) domain.BJSuggestedAction {
		bj := newGame()
		bj.ToggleHint()
		// ハード12 対 ディーラー4。
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignClover, 4, false))
		bj.GetDealer().AddCard(domain.NewCard(domain.CardDesignDiamond, 5, false))
		bj.SetPhase(domain.BJPhaseAction)
		return bj.GetBasicStrategySuggestion()
	}

	assert.Equal(t, domain.BJSuggestHit, forVariant(domain.NewSpanish21BlackJack),
		"Spanish 21 should hit hard 12 vs 4")
	assert.Equal(t, domain.BJSuggestStand, forVariant(domain.NewDefaultBlackJack),
		"standard blackjack should still stand on hard 12 vs 4")
}

// TestSpanish21Strategy_CoversEveryCell is a shape guard: every (hand, upcard)
// the table can be asked about must return a real action.
func TestSpanish21Strategy_CoversEveryCell(t *testing.T) {
	upcards := []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 1}
	checked := 0
	for _, u := range upcards {
		for total := 5; total <= 20; total++ {
			var hand *domain.BlackJackHand
			switch {
			case total <= 11:
				hand = s21Hand(2, total-2)
			default:
				hand = s21Hand(9, total-9)
			}
			got := domain.GetSpanish21StrategyAction(hand, up(u))
			assert.NotEqual(t, domain.BJSuggestNone, got, fmt.Sprintf("hard %d vs %d", total, u))
			checked++
		}
		for soft := 13; soft <= 20; soft++ {
			got := domain.GetSpanish21StrategyAction(s21Hand(1, soft-11), up(u))
			assert.NotEqual(t, domain.BJSuggestNone, got, fmt.Sprintf("soft %d vs %d", soft, u))
			checked++
		}
		for pv := 1; pv <= 10; pv++ {
			got := domain.GetSpanish21StrategyAction(s21Hand(pv, pv), up(u))
			assert.NotEqual(t, domain.BJSuggestNone, got, fmt.Sprintf("pair %d vs %d", pv, u))
			checked++
		}
	}
	// **走査が壊れて0件になっても通らないようにする。**
	assert.GreaterOrEqual(t, checked, 300, "cells checked")
}
