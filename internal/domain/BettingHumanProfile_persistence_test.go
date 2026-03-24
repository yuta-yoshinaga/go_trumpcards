//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBettingHumanProfile_ExportImport_RoundTrip(t *testing.T) {
	p := &BettingHumanProfile{
		FoldToBetCount:  3,
		FoldToBetTotal:  10,
		GamesPlayed:     5,
		HesitationCount: 4,
		HesitationMean:  1500.0,
		HesitationM2:    250000.0,
	}
	p.AggressiveByBracket[0] = struct{ Aggressive, Total int }{2, 5}
	p.AggressiveByBracket[1] = struct{ Aggressive, Total int }{4, 8}
	p.AggressiveByBracket[2] = struct{ Aggressive, Total int }{1, 3}

	data := p.Export()
	p2 := &BettingHumanProfile{}
	p2.Import(data)

	assert.Equal(t, p.AggressiveByBracket, p2.AggressiveByBracket)
	assert.Equal(t, p.FoldToBetCount, p2.FoldToBetCount)
	assert.Equal(t, p.FoldToBetTotal, p2.FoldToBetTotal)
	assert.Equal(t, p.GamesPlayed, p2.GamesPlayed)
	assert.Equal(t, p.HesitationCount, p2.HesitationCount)
	assert.Equal(t, p.HesitationMean, p2.HesitationMean)
	assert.Equal(t, p.HesitationM2, p2.HesitationM2)
}

func TestBettingHumanProfile_ExportImport_JSON_RoundTrip(t *testing.T) {
	p := &BettingHumanProfile{GamesPlayed: 3, FoldToBetCount: 2, FoldToBetTotal: 5}
	p.AggressiveByBracket[1] = struct{ Aggressive, Total int }{3, 6}

	data := p.Export()
	jsonBytes, err := json.Marshal(data)
	require.NoError(t, err)

	parsed, err := ImportBettingHumanProfileJSON(jsonBytes)
	require.NoError(t, err)

	p2 := &BettingHumanProfile{}
	p2.Import(parsed)

	assert.Equal(t, p.GamesPlayed, p2.GamesPlayed)
	assert.Equal(t, p.FoldToBetCount, p2.FoldToBetCount)
	assert.Equal(t, p.AggressiveByBracket[1].Aggressive, p2.AggressiveByBracket[1].Aggressive)
}

func TestImportBettingHumanProfileJSON_InvalidJSON(t *testing.T) {
	_, err := ImportBettingHumanProfileJSON([]byte("invalid"))
	assert.Error(t, err)
}

func TestPoker_ExportProfile_NilWhenNoProfile(t *testing.T) {
	pk := newPokerForMetaAITest(false)
	assert.Nil(t, pk.ExportProfile())
}

func TestPoker_ExportProfile_ReturnsData(t *testing.T) {
	pk := newPokerForMetaAITest(true)
	profile := pk.ExportProfile()
	assert.NotNil(t, profile)
	data, ok := profile.(*BettingHumanProfileData)
	assert.True(t, ok)
	assert.Equal(t, 0, data.GamesPlayed)
}

func TestPoker_ImportProfile_ValidJSON(t *testing.T) {
	pk := newPokerForMetaAITest(true)
	profileData := BettingHumanProfileData{GamesPlayed: 3, FoldToBetCount: 2, FoldToBetTotal: 5}
	jsonBytes, _ := json.Marshal(profileData)

	err := pk.ImportProfile(jsonBytes)
	assert.NoError(t, err)
	assert.Equal(t, 3, pk.GetHumanProfile().GamesPlayed)
	assert.Equal(t, 2, pk.GetHumanProfile().FoldToBetCount)
}

func TestPoker_ImportProfile_EmptyBytes(t *testing.T) {
	pk := newPokerForMetaAITest(true)
	err := pk.ImportProfile(nil)
	assert.NoError(t, err)

	err = pk.ImportProfile([]byte{})
	assert.NoError(t, err)
}

func TestPoker_ImportProfile_InvalidJSON(t *testing.T) {
	pk := newPokerForMetaAITest(true)
	err := pk.ImportProfile([]byte("invalid"))
	assert.Error(t, err)
}

func TestHoldem_ExportImportProfile(t *testing.T) {
	h := newHoldemForMetaAITest(true)
	assert.NotNil(t, h.ExportProfile())

	profileData := BettingHumanProfileData{GamesPlayed: 5, FoldToBetCount: 3, FoldToBetTotal: 8}
	jsonBytes, _ := json.Marshal(profileData)
	assert.NoError(t, h.ImportProfile(jsonBytes))
	assert.Equal(t, 5, h.GetHumanProfile().GamesPlayed)
}

func TestHoldem_ExportProfile_NilWhenDisabled(t *testing.T) {
	h := newHoldemForMetaAITest(false)
	assert.Nil(t, h.ExportProfile())
}

func TestOmaha_ExportImportProfile(t *testing.T) {
	o := newOmahaForMetaAITest(true)
	assert.NotNil(t, o.ExportProfile())

	profileData := BettingHumanProfileData{GamesPlayed: 4, FoldToBetCount: 2, FoldToBetTotal: 6}
	jsonBytes, _ := json.Marshal(profileData)
	assert.NoError(t, o.ImportProfile(jsonBytes))
	assert.Equal(t, 4, o.GetHumanProfile().GamesPlayed)
}

func TestOmaha_ExportProfile_NilWhenDisabled(t *testing.T) {
	o := newOmahaForMetaAITest(false)
	assert.Nil(t, o.ExportProfile())
}

func newPokerForMetaAITest(metaAI bool) *Poker {
	tc := NewTrumpCards(0)
	players := []*PokerPlayer{
		NewPokerPlayer(true, PokerStyleBalanced),
		NewPokerPlayer(false, PokerStyleConservative),
		NewPokerPlayer(false, PokerStyleAggressive),
		NewPokerPlayer(false, PokerStyleBluffer),
	}
	for _, pl := range players {
		pl.SetChips(1000)
	}
	cfg := DefaultPokerConfig()
	cfg.CpuMetaAI = metaAI
	pk := NewPoker(tc, players, cfg)
	_ = pk.Reset()
	return pk
}

func newHoldemForMetaAITest(metaAI bool) *Holdem {
	tc := NewTrumpCards(0)
	players := []*HoldemPlayer{
		NewHoldemPlayer(true, HoldemStyleTAG),
		NewHoldemPlayer(false, HoldemStyleLAP),
		NewHoldemPlayer(false, HoldemStyleTAP),
		NewHoldemPlayer(false, HoldemStyleLAG),
	}
	for _, pl := range players {
		pl.SetChips(1000)
	}
	cfg := DefaultHoldemConfig()
	cfg.CpuMetaAI = metaAI
	h := NewHoldem(tc, players, cfg)
	_ = h.Reset()
	return h
}

func newOmahaForMetaAITest(metaAI bool) *Omaha {
	tc := NewTrumpCards(0)
	players := []*OmahaPlayer{
		NewOmahaPlayer(true, HoldemStyleTAG),
		NewOmahaPlayer(false, HoldemStyleLAP),
		NewOmahaPlayer(false, HoldemStyleTAP),
		NewOmahaPlayer(false, HoldemStyleLAG),
	}
	for _, pl := range players {
		pl.SetChips(1000)
	}
	cfg := DefaultOmahaConfig()
	cfg.CpuMetaAI = metaAI
	o := NewOmaha(tc, players, cfg)
	_ = o.Reset()
	return o
}
