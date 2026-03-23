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
		HesitationMean:  2000.0,
		HesitationM2:    300000.0,
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
	p := &DoubtHumanProfile{GamesPlayed: 4, DoubtCorrect: 2, DoubtTotal: 6}

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

func TestDoubt_ExportProfile_NilWhenNoProfile(t *testing.T) {
	cfg := DefaultDoubtConfig()
	cfg.CpuMetaAI = false
	game := NewDoubt(cfg)
	game.Reset()
	assert.Nil(t, game.ExportProfile())
}

func TestDoubt_ExportProfile_ReturnsData(t *testing.T) {
	cfg := DefaultDoubtConfig()
	cfg.CpuMetaAI = true
	game := NewDoubt(cfg)
	game.Reset()
	profile := game.ExportProfile()
	assert.NotNil(t, profile)
}

func TestDoubt_ImportProfile_ValidJSON(t *testing.T) {
	cfg := DefaultDoubtConfig()
	cfg.CpuMetaAI = true
	game := NewDoubt(cfg)
	game.Reset()

	profileData := DoubtHumanProfileData{GamesPlayed: 3, DoubtCorrect: 2, DoubtTotal: 5}
	jsonBytes, _ := json.Marshal(profileData)

	err := game.ImportProfile(jsonBytes)
	assert.NoError(t, err)
	assert.Equal(t, 3, game.GetHumanProfile().GamesPlayed)
}

func TestDoubt_ImportProfile_EmptyBytes(t *testing.T) {
	cfg := DefaultDoubtConfig()
	game := NewDoubt(cfg)
	assert.NoError(t, game.ImportProfile(nil))
	assert.NoError(t, game.ImportProfile([]byte{}))
}

func TestDoubt_ImportProfile_InvalidJSON(t *testing.T) {
	game := NewDoubt(DefaultDoubtConfig())
	assert.Error(t, game.ImportProfile([]byte("invalid")))
}
