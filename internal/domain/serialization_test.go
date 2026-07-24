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

func TestBlackJack_JSONRoundTrip(t *testing.T) {
	bj := NewDefaultBlackJack()

	data, err := json.Marshal(bj)
	require.NoError(t, err)

	var got BlackJack
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, bj.GetPhase(), got.GetPhase())
	assert.Equal(t, bj.GetGameEndFlag(), got.GetGameEndFlag())
	assert.Equal(t, bj.GetDeckCount(), got.GetDeckCount())
}

func TestPoker_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	players := []*PokerPlayer{
		NewPokerPlayer(true, PokerStyleBalanced),
		NewPokerPlayer(false, PokerStyleConservative),
	}
	p := NewPoker(tc, players, DefaultPokerConfig())

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var got Poker
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, p.GetPhase(), got.GetPhase())
	assert.Equal(t, p.GetGameEndFlag(), got.GetGameEndFlag())
}

func TestHoldem_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	players := []*HoldemPlayer{
		NewHoldemPlayer(true, HoldemStyleTAG),
		NewHoldemPlayer(false, HoldemStyleLAP),
	}
	h := NewHoldem(tc, players, DefaultHoldemConfig())

	data, err := json.Marshal(h)
	require.NoError(t, err)

	var got Holdem
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, h.GetPhase(), got.GetPhase())
	assert.Equal(t, h.GetGameEndFlag(), got.GetGameEndFlag())
}

func TestOmaha_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	players := []*OmahaPlayer{
		NewOmahaPlayer(true, HoldemStyleTAG),
		NewOmahaPlayer(false, HoldemStyleLAP),
	}
	o := NewOmaha(tc, players, DefaultOmahaConfig())

	data, err := json.Marshal(o)
	require.NoError(t, err)

	var got Omaha
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, o.GetPhase(), got.GetPhase())
	assert.Equal(t, o.GetGameEndFlag(), got.GetGameEndFlag())
}

func TestShortDeck_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	players := []*ShortDeckPlayer{
		NewShortDeckPlayer(true, HoldemStyleTAG),
		NewShortDeckPlayer(false, HoldemStyleLAP),
	}
	sd := NewShortDeck(tc, players, DefaultShortDeckConfig())

	data, err := json.Marshal(sd)
	require.NoError(t, err)

	var got ShortDeck
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, sd.GetPhase(), got.GetPhase())
	assert.Equal(t, sd.GetGameEndFlag(), got.GetGameEndFlag())
}

func TestIndianPoker_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	players := []*IndianPokerPlayer{
		NewIndianPokerPlayer(true, HoldemStyleTAG),
		NewIndianPokerPlayer(false, HoldemStyleLAP),
	}
	ip := NewIndianPoker(tc, players, DefaultIndianPokerConfig())

	data, err := json.Marshal(ip)
	require.NoError(t, err)

	var got IndianPoker
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, ip.GetPhase(), got.GetPhase())
	assert.Equal(t, ip.GetGameEndFlag(), got.GetGameEndFlag())
}

func TestVideoPoker_JSONRoundTrip(t *testing.T) {
	vp := NewDefaultVideoPoker()
	vp.SetHandName("Wild Royal Flush")
	vp.SetHandKey("wildRoyalFlush")

	data, err := json.Marshal(vp)
	require.NoError(t, err)

	var got VideoPoker
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, vp.GetPhase(), got.GetPhase())
	assert.Equal(t, vp.GetGameEndFlag(), got.GetGameEndFlag())
	assert.Equal(t, vp.GetHandName(), got.GetHandName())
	assert.Equal(t, vp.GetHandKey(), got.GetHandKey())
}

func TestHearts_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	players := []*HeartsPlayer{
		NewHeartsPlayer(true),
		NewHeartsPlayer(false),
		NewHeartsPlayer(false),
		NewHeartsPlayer(false),
	}
	h := NewHearts(tc, players, DefaultHeartsConfig())

	data, err := json.Marshal(h)
	require.NoError(t, err)

	var got Hearts
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, h.GetPhase(), got.GetPhase())
	assert.Equal(t, h.GetGameEndFlag(), got.GetGameEndFlag())
}

func TestSpades_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	players := []*SpadesPlayer{
		NewSpadesPlayer(true),
		NewSpadesPlayer(false),
		NewSpadesPlayer(false),
		NewSpadesPlayer(false),
	}
	s := NewSpades(tc, players, DefaultSpadesConfig())

	data, err := json.Marshal(s)
	require.NoError(t, err)

	var got Spades
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, s.GetPhase(), got.GetPhase())
	assert.Equal(t, s.GetGameEndFlag(), got.GetGameEndFlag())
}

func TestEuchre_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	players := []*EuchrePlayer{
		NewEuchrePlayer(true, 0),
		NewEuchrePlayer(false, 1),
		NewEuchrePlayer(false, 0),
		NewEuchrePlayer(false, 1),
	}
	e := NewEuchre(tc, players, DefaultEuchreConfig())

	data, err := json.Marshal(e)
	require.NoError(t, err)

	var got Euchre
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, e.GetPhase(), got.GetPhase())
	assert.Equal(t, e.GetGameEndFlag(), got.GetGameEndFlag())
}

func TestNapoleon_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	players := []*NapoleonPlayer{
		NewNapoleonPlayer(true),
		NewNapoleonPlayer(false),
		NewNapoleonPlayer(false),
		NewNapoleonPlayer(false),
		NewNapoleonPlayer(false),
	}
	n := NewNapoleon(tc, players, DefaultNapoleonConfig())

	data, err := json.Marshal(n)
	require.NoError(t, err)

	var got Napoleon
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, n.GetPhase(), got.GetPhase())
	assert.Equal(t, n.GetGameEndFlag(), got.GetGameEndFlag())
}

func TestOldMaid_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	players := []*OldMaidPlayer{
		NewOldMaidPlayer(true),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
	}
	om := NewOldMaid(tc, players)

	data, err := json.Marshal(om)
	require.NoError(t, err)

	var got OldMaid
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, om.GetGameEndFlag(), got.GetGameEndFlag())
	assert.Equal(t, om.GetCurrentTurn(), got.GetCurrentTurn())
}

func TestDoubt_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	players := []*DoubtPlayer{
		NewDoubtPlayer(true),
		NewDoubtPlayer(false),
	}
	d := NewDoubt(tc, players)

	data, err := json.Marshal(d)
	require.NoError(t, err)

	var got Doubt
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, d.GetPhase(), got.GetPhase())
	assert.Equal(t, d.GetGameEndFlag(), got.GetGameEndFlag())
}

func TestDaifugo_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	players := []*DaifugoPlayer{
		NewDaifugoPlayer(true),
		NewDaifugoPlayer(false),
		NewDaifugoPlayer(false),
	}
	dg := NewDaifugo(tc, players, DefaultDaifugoConfig())

	data, err := json.Marshal(dg)
	require.NoError(t, err)

	var got Daifugo
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, dg.GetGameEndFlag(), got.GetGameEndFlag())
	assert.Equal(t, dg.GetCurrentTurn(), got.GetCurrentTurn())
}

func TestSevens_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	players := []*SevensPlayer{
		NewSevensPlayer(true),
		NewSevensPlayer(false),
		NewSevensPlayer(false),
	}
	s := NewSevens(tc, players, DefaultSevensConfig())

	data, err := json.Marshal(s)
	require.NoError(t, err)

	var got Sevens
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, s.GetGameEndFlag(), got.GetGameEndFlag())
	assert.Equal(t, s.GetCurrentTurn(), got.GetCurrentTurn())
}

func TestCrazyEights_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	players := []*CrazyEightsPlayer{
		NewCrazyEightsPlayer(true),
		NewCrazyEightsPlayer(false),
	}
	ce := NewCrazyEights(tc, players, DefaultCrazyEightsConfig())

	data, err := json.Marshal(ce)
	require.NoError(t, err)

	var got CrazyEights
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, ce.GetPhase(), got.GetPhase())
	assert.Equal(t, ce.GetGameEndFlag(), got.GetGameEndFlag())
}

func TestKlondike_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	k := NewKlondike(tc)

	data, err := json.Marshal(k)
	require.NoError(t, err)

	var got Klondike
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, k.GetPhase(), got.GetPhase())
}

func TestFreeCell_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	f := NewFreeCell(tc)

	data, err := json.Marshal(f)
	require.NoError(t, err)

	var got FreeCell
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, f.GetPhase(), got.GetPhase())
}

func TestSpider_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	s := NewSpider(tc)

	data, err := json.Marshal(s)
	require.NoError(t, err)

	var got Spider
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, s.GetPhase(), got.GetPhase())
}

func TestPyramid_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	p := NewPyramid(tc)

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var got Pyramid
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, p.GetPhase(), got.GetPhase())
}

func TestMemory_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	players := []*MemoryPlayer{
		NewMemoryPlayer(true),
		NewMemoryPlayer(false),
	}
	m := NewMemory(tc, players, DefaultMemoryConfig())

	data, err := json.Marshal(m)
	require.NoError(t, err)

	var got Memory
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, m.GetPhase(), got.GetPhase())
	assert.Equal(t, m.GetGameEndFlag(), got.GetGameEndFlag())
}

func TestGinRummy_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	players := []*GinRummyPlayer{
		NewGinRummyPlayer(true),
		NewGinRummyPlayer(false),
	}
	gr := NewGinRummy(tc, players, DefaultGinRummyConfig())

	data, err := json.Marshal(gr)
	require.NoError(t, err)

	var got GinRummy
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, gr.GetPhase(), got.GetPhase())
	assert.Equal(t, gr.GetGameEndFlag(), got.GetGameEndFlag())
}

func TestCribbage_JSONRoundTrip(t *testing.T) {
	tc := NewTrumpCards(0)
	players := []*CribbagePlayer{
		NewCribbagePlayer(true),
		NewCribbagePlayer(false),
	}
	cr := NewCribbage(tc, players, DefaultCribbageConfig())

	data, err := json.Marshal(cr)
	require.NoError(t, err)

	var got Cribbage
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, cr.GetPhase(), got.GetPhase())
	assert.Equal(t, cr.GetGameEndFlag(), got.GetGameEndFlag())
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
