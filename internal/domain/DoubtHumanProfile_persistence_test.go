//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoubtHumanProfile_ExportImport_RoundTrip(t *testing.T) {
	p := &DoubtHumanProfile{
		DoubtCorrect:    3,
		DoubtTotal:      7,
		GamesPlayed:     5,
		HesitationCount: 4,
		HesitationMean:  1200.0,
		HesitationM2:    180000.0,
	}
	p.BluffsByBracket[0] = struct{ Bluffs, Total int }{2, 5}
	p.BluffsByBracket[1] = struct{ Bluffs, Total int }{3, 8}
	p.BluffsByBracket[2] = struct{ Bluffs, Total int }{1, 2}

	data := p.Export()
	p2 := &DoubtHumanProfile{}
	p2.Import(data)

	assert.Equal(t, p.BluffsByBracket, p2.BluffsByBracket)
	assert.Equal(t, p.DoubtCorrect, p2.DoubtCorrect)
	assert.Equal(t, p.DoubtTotal, p2.DoubtTotal)
	assert.Equal(t, p.GamesPlayed, p2.GamesPlayed)
	assert.Equal(t, p.HesitationCount, p2.HesitationCount)
	assert.Equal(t, p.HesitationMean, p2.HesitationMean)
	assert.Equal(t, p.HesitationM2, p2.HesitationM2)
}

func TestDoubtHumanProfile_ExportImport_JSON_RoundTrip(t *testing.T) {
	p := &DoubtHumanProfile{GamesPlayed: 3, DoubtCorrect: 2, DoubtTotal: 5}
	data := p.Export()
	jsonBytes, err := json.Marshal(data)
	require.NoError(t, err)

	parsed, err := ImportDoubtHumanProfileJSON(jsonBytes)
	require.NoError(t, err)

	p2 := &DoubtHumanProfile{}
	p2.Import(parsed)

	assert.Equal(t, p.GamesPlayed, p2.GamesPlayed)
	assert.Equal(t, p.DoubtCorrect, p2.DoubtCorrect)
}

func TestImportDoubtHumanProfileJSON_InvalidJSON(t *testing.T) {
	_, err := ImportDoubtHumanProfileJSON([]byte("invalid"))
	assert.Error(t, err)
}

func newDoubtForMetaAITest(metaAI bool) *Doubt {
	tc := NewTrumpCards(0)
	players := []*DoubtPlayer{
		NewDoubtPlayer(true),
		NewDoubtPlayer(false),
		NewDoubtPlayer(false),
		NewDoubtPlayer(false),
	}
	game := NewDoubt(tc, players)
	cfg := DefaultDoubtConfig()
	cfg.CpuMetaAI = metaAI
	game.SetConfig(cfg)
	game.Reset()
	return game
}

func TestDoubt_ExportProfile_NilWhenNoProfile(t *testing.T) {
	game := newDoubtForMetaAITest(false)
	assert.Nil(t, game.ExportProfile())
}

func TestDoubt_ExportProfile_ReturnsData(t *testing.T) {
	game := newDoubtForMetaAITest(true)
	profile := game.ExportProfile()
	assert.NotNil(t, profile)
}

func TestDoubt_ImportProfile_ValidJSON(t *testing.T) {
	game := newDoubtForMetaAITest(true)
	profileData := DoubtHumanProfileData{GamesPlayed: 3, DoubtCorrect: 2, DoubtTotal: 5}
	jsonBytes, _ := json.Marshal(profileData)

	err := game.ImportProfile(jsonBytes)
	assert.NoError(t, err)
	assert.Equal(t, 3, game.GetHumanProfile().GamesPlayed)
}

func TestDoubt_ImportProfile_EmptyBytes(t *testing.T) {
	game := newDoubtForMetaAITest(false)
	assert.NoError(t, game.ImportProfile(nil))
	assert.NoError(t, game.ImportProfile([]byte{}))
}

func TestDoubt_ImportProfile_InvalidJSON(t *testing.T) {
	game := newDoubtForMetaAITest(false)
	assert.Error(t, game.ImportProfile([]byte("invalid")))
}
