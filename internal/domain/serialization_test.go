//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCard_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		card *Card
	}{
		{"spade ace", NewCard(CardDesignSpade, 1, false)},
		{"heart king drawn", NewCard(CardDesignHeart, CardValueMax, true)},
		{"joker", NewCard(CardDesignJoker, 1, false)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.card)
			require.NoError(t, err)

			var got Card
			require.NoError(t, json.Unmarshal(data, &got))

			assert.Equal(t, tt.card.GetDesign(), got.GetDesign())
			assert.Equal(t, tt.card.GetValue(), got.GetValue())
			assert.Equal(t, tt.card.GetDraw(), got.GetDraw())
		})
	}
}

func TestTrumpCards_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(1) // 52 + 1 joker
	tc.Shuffle()
	// Draw a few cards
	tc.DrawCard()
	tc.DrawCard()

	data, err := json.Marshal(tc)
	require.NoError(t, err)

	var got TrumpCards
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, tc.GetTotalCount(), got.GetTotalCount())
	assert.Equal(t, tc.GetRemainingCount(), got.GetRemainingCount())
	// Verify card order is preserved
	for i := 0; i < tc.GetTotalCount(); i++ {
		assert.Equal(t, tc.deck[i].GetDesign(), got.deck[i].GetDesign())
		assert.Equal(t, tc.deck[i].GetValue(), got.deck[i].GetValue())
		assert.Equal(t, tc.deck[i].GetDraw(), got.deck[i].GetDraw())
	}
}

func TestPlayer_JSONRoundTrip(t *testing.T) {
	p := &Player{cards: make([]*Card, 0)}
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignHeart, 13, true))

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var got Player
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, p.GetCardsSize(), got.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		assert.Equal(t, p.GetCard(i).GetDesign(), got.GetCard(i).GetDesign())
		assert.Equal(t, p.GetCard(i).GetValue(), got.GetCard(i).GetValue())
	}
}

func TestGamePlayer_JSONRoundTrip(t *testing.T) {
	gp := NewGamePlayer(true)
	gp.AddCard(NewCard(CardDesignClover, 5, false))
	gp.SetIsFinished(true)

	data, err := json.Marshal(gp)
	require.NoError(t, err)

	var got GamePlayer
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, gp.GetIsHuman(), got.GetIsHuman())
	assert.Equal(t, gp.GetIsFinished(), got.GetIsFinished())
	assert.Equal(t, gp.GetCardsSize(), got.GetCardsSize())
}

func TestRankedGamePlayer_JSONRoundTrip(t *testing.T) {
	rp := NewRankedGamePlayer(false)
	rp.AddCard(NewCard(CardDesignDiamond, 10, false))
	rp.SetRank(2)

	data, err := json.Marshal(rp)
	require.NoError(t, err)

	var got RankedGamePlayer
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, rp.GetIsHuman(), got.GetIsHuman())
	assert.Equal(t, rp.GetRank(), got.GetRank())
	assert.Equal(t, rp.GetCardsSize(), got.GetCardsSize())
}

func TestChipHolder_JSONRoundTrip(t *testing.T) {
	ch := &ChipHolder{}
	ch.SetChips(1000)

	data, err := json.Marshal(ch)
	require.NoError(t, err)

	var got ChipHolder
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, ch.GetChips(), got.GetChips())
}

func TestBaccarat_JSONRoundTrip(t *testing.T) {
	bac := NewDefaultBaccarat()
	// Play a game to populate state
	_ = bac.Bet(100, BaccaratBetPlayer, 10, 10)

	data, err := json.Marshal(bac)
	require.NoError(t, err)

	var got Baccarat
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, bac.GetPhase(), got.GetPhase())
	assert.Equal(t, bac.GetChips(), got.GetChips())
	assert.Equal(t, bac.GetBetAmount(), got.GetBetAmount())
	assert.Equal(t, bac.GetBetType(), got.GetBetType())
	assert.Equal(t, bac.GetResult(), got.GetResult())
	assert.Equal(t, bac.GetPayout(), got.GetPayout())
	assert.Equal(t, bac.GetGameEndFlag(), got.GetGameEndFlag())
	assert.Equal(t, bac.GetHistory(), got.GetHistory())
	assert.Equal(t, bac.GetPlayerPairBet(), got.GetPlayerPairBet())
	assert.Equal(t, bac.GetBankerPairBet(), got.GetBankerPairBet())
	assert.Equal(t, len(bac.GetPlayerHand()), len(got.GetPlayerHand()))
	assert.Equal(t, len(bac.GetBankerHand()), len(got.GetBankerHand()))
	assert.Equal(t, len(bac.GetActionLog()), len(got.GetActionLog()))
	assert.Equal(t, len(bac.GetSideBetResults()), len(got.GetSideBetResults()))
}

func TestBaccarat_JSONRoundTrip_BetPhase(t *testing.T) {
	bac := NewDefaultBaccarat()
	// Don't play — serialize in initial bet phase
	data, err := json.Marshal(bac)
	require.NoError(t, err)

	var got Baccarat
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, BaccaratPhaseBet, got.GetPhase())
	assert.Equal(t, BaccaratDefaultChips, got.GetChips())
	assert.False(t, got.GetGameEndFlag())
}

func TestActionLogEntry_JSONRoundTrip(t *testing.T) {
	entry := &ActionLogEntry{
		TurnNumber: 3,
		PlayerIdx:  0,
		ActionType: "hit",
		Detail:     "Player hits",
		Cards:      []*Card{NewCard(CardDesignSpade, 7, false)},
	}

	data, err := json.Marshal(entry)
	require.NoError(t, err)

	var got ActionLogEntry
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, entry.TurnNumber, got.TurnNumber)
	assert.Equal(t, entry.PlayerIdx, got.PlayerIdx)
	assert.Equal(t, entry.ActionType, got.ActionType)
	assert.Equal(t, entry.Detail, got.Detail)
	require.Len(t, got.Cards, 1)
	assert.Equal(t, entry.Cards[0].GetDesign(), got.Cards[0].GetDesign())
	assert.Equal(t, entry.Cards[0].GetValue(), got.Cards[0].GetValue())
}
