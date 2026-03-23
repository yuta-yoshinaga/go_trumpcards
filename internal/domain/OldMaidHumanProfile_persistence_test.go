//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOldMaidHumanProfile_ExportImport_RoundTrip(t *testing.T) {
	p := &OldMaidHumanProfile{
		PositionBuckets: [3]int{5, 3, 7},
		TotalPicks:      15,
		ShuffleCount:    2,
		DrawCount:       10,
		GamesPlayed:     4,
	}

	data := p.Export()
	p2 := &OldMaidHumanProfile{}
	p2.Import(data)

	assert.Equal(t, p.PositionBuckets, p2.PositionBuckets)
	assert.Equal(t, p.TotalPicks, p2.TotalPicks)
	assert.Equal(t, p.ShuffleCount, p2.ShuffleCount)
	assert.Equal(t, p.DrawCount, p2.DrawCount)
	assert.Equal(t, p.GamesPlayed, p2.GamesPlayed)
}

func TestOldMaidHumanProfile_ExportImport_JSON_RoundTrip(t *testing.T) {
	p := &OldMaidHumanProfile{GamesPlayed: 3, TotalPicks: 8}

	data := p.Export()
	jsonBytes, err := json.Marshal(data)
	require.NoError(t, err)

	parsed, err := ImportOldMaidHumanProfileJSON(jsonBytes)
	require.NoError(t, err)

	p2 := &OldMaidHumanProfile{}
	p2.Import(parsed)

	assert.Equal(t, p.GamesPlayed, p2.GamesPlayed)
	assert.Equal(t, p.TotalPicks, p2.TotalPicks)
}

func TestImportOldMaidHumanProfileJSON_InvalidJSON(t *testing.T) {
	_, err := ImportOldMaidHumanProfileJSON([]byte("invalid"))
	assert.Error(t, err)
}

func newOldMaidForMetaAITest(metaAI bool) *OldMaid {
	tc := NewTrumpCards(0)
	players := []*OldMaidPlayer{
		NewOldMaidPlayer(true),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
		NewOldMaidPlayer(false),
	}
	om := NewOldMaid(tc, players)
	cfg := DefaultOldMaidConfig()
	cfg.CpuMetaAI = metaAI
	om.SetConfig(cfg)
	om.Reset()
	return om
}

func TestOldMaid_ExportProfile_NilWhenNoProfile(t *testing.T) {
	om := newOldMaidForMetaAITest(false)
	assert.Nil(t, om.ExportProfile())
}

func TestOldMaid_ExportProfile_ReturnsData(t *testing.T) {
	om := newOldMaidForMetaAITest(true)
	assert.NotNil(t, om.ExportProfile())
}

func TestOldMaid_ImportProfile_ValidJSON(t *testing.T) {
	om := newOldMaidForMetaAITest(true)
	profileData := OldMaidHumanProfileData{GamesPlayed: 3, TotalPicks: 10}
	jsonBytes, _ := json.Marshal(profileData)

	err := om.ImportProfile(jsonBytes)
	assert.NoError(t, err)
	assert.Equal(t, 3, om.GetHumanProfile().GamesPlayed)
}

func TestOldMaid_ImportProfile_EmptyBytes(t *testing.T) {
	om := newOldMaidForMetaAITest(false)
	assert.NoError(t, om.ImportProfile(nil))
	assert.NoError(t, om.ImportProfile([]byte{}))
}

func TestOldMaid_ImportProfile_InvalidJSON(t *testing.T) {
	om := newOldMaidForMetaAITest(false)
	assert.Error(t, om.ImportProfile([]byte("invalid")))
}
