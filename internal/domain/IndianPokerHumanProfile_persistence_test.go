//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndianPokerHumanProfile_ExportImport_RoundTrip(t *testing.T) {
	p := &IndianPokerHumanProfile{
		FoldToBetCount:  2,
		FoldToBetTotal:  8,
		GamesPlayed:     3,
		HesitationCount: 5,
		HesitationMean:  1200.0,
		HesitationM2:    180000.0,
	}
	p.AggressiveByBracket[0] = struct{ Aggressive, Total int }{1, 4}
	p.AggressiveByBracket[1] = struct{ Aggressive, Total int }{3, 6}
	p.AggressiveByBracket[2] = struct{ Aggressive, Total int }{2, 3}

	data := p.Export()
	p2 := &IndianPokerHumanProfile{}
	p2.Import(data)

	assert.Equal(t, p.AggressiveByBracket, p2.AggressiveByBracket)
	assert.Equal(t, p.FoldToBetCount, p2.FoldToBetCount)
	assert.Equal(t, p.FoldToBetTotal, p2.FoldToBetTotal)
	assert.Equal(t, p.GamesPlayed, p2.GamesPlayed)
	assert.Equal(t, p.HesitationCount, p2.HesitationCount)
	assert.Equal(t, p.HesitationMean, p2.HesitationMean)
	assert.Equal(t, p.HesitationM2, p2.HesitationM2)
}

func TestIndianPokerHumanProfile_ExportImport_JSON_RoundTrip(t *testing.T) {
	p := &IndianPokerHumanProfile{GamesPlayed: 2, FoldToBetCount: 1, FoldToBetTotal: 4}

	data := p.Export()
	jsonBytes, err := json.Marshal(data)
	require.NoError(t, err)

	parsed, err := ImportIndianPokerHumanProfileJSON(jsonBytes)
	require.NoError(t, err)

	p2 := &IndianPokerHumanProfile{}
	p2.Import(parsed)

	assert.Equal(t, p.GamesPlayed, p2.GamesPlayed)
	assert.Equal(t, p.FoldToBetCount, p2.FoldToBetCount)
}

func TestImportIndianPokerHumanProfileJSON_InvalidJSON(t *testing.T) {
	_, err := ImportIndianPokerHumanProfileJSON([]byte("invalid"))
	assert.Error(t, err)
}

func TestIndianPoker_ExportProfile_NilWhenNoProfile(t *testing.T) {
	cfg := DefaultIndianPokerConfig()
	cfg.CpuMetaAI = false
	ip := NewIndianPoker(cfg)
	_ = ip.Reset()
	assert.Nil(t, ip.ExportProfile())
}

func TestIndianPoker_ExportProfile_ReturnsData(t *testing.T) {
	cfg := DefaultIndianPokerConfig()
	cfg.CpuMetaAI = true
	ip := NewIndianPoker(cfg)
	_ = ip.Reset()
	assert.NotNil(t, ip.ExportProfile())
}

func TestIndianPoker_ImportProfile_ValidJSON(t *testing.T) {
	cfg := DefaultIndianPokerConfig()
	cfg.CpuMetaAI = true
	ip := NewIndianPoker(cfg)
	_ = ip.Reset()

	profileData := IndianPokerHumanProfileData{GamesPlayed: 4, FoldToBetCount: 3, FoldToBetTotal: 7}
	jsonBytes, _ := json.Marshal(profileData)

	err := ip.ImportProfile(jsonBytes)
	assert.NoError(t, err)
	assert.Equal(t, 4, ip.GetHumanProfile().GamesPlayed)
}

func TestIndianPoker_ImportProfile_EmptyBytes(t *testing.T) {
	ip := NewIndianPoker(DefaultIndianPokerConfig())
	assert.NoError(t, ip.ImportProfile(nil))
	assert.NoError(t, ip.ImportProfile([]byte{}))
}

func TestIndianPoker_ImportProfile_InvalidJSON(t *testing.T) {
	ip := NewIndianPoker(DefaultIndianPokerConfig())
	assert.Error(t, ip.ImportProfile([]byte("invalid")))
}
