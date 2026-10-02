//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalabresella_GetHighestBidDuringAuction(t *testing.T) {
	g := NewDefaultCalabresella()
	g.Reset()
	g.phase = CalabresellaPhaseBid

	assert.Equal(t, CalabresellaBidNone, g.GetHighestBid(), "all players have passed")
	assert.Equal(t, CalabresellaBidNone, g.GetWinningBid(), "winning bid remains unset during the auction")
	assert.True(t, g.isBidLegal(CalabresellaBidChiamo), "chiamo is legal when there is no bid")

	g.bids = [CalabresellaPlayerCnt]CalabresellaBid{
		CalabresellaBidChiamo,
		CalabresellaBidNone,
		CalabresellaBidNone,
	}
	assert.Equal(t, CalabresellaBidChiamo, g.GetHighestBid())
	assert.False(t, g.isBidLegal(CalabresellaBidChiamo), "chiamo cannot match the current highest bid")
	assert.Equal(t, CalabresellaBidNone, g.GetWinningBid(), "winning bid remains unset during the auction")

	g.bids[1] = CalabresellaBidSolo
	assert.Equal(t, CalabresellaBidSolo, g.GetHighestBid(), "solo is higher than chiamo")
	assert.Equal(t, CalabresellaBidNone, g.GetWinningBid(), "winning bid remains unset during the auction")
}

func TestCalabresella_FinalTrickMatchEndKeepsRoundResultAndScoresOnce(t *testing.T) {
	g := NewDefaultCalabresella()
	g.Reset()
	cfg := g.GetConfig()
	cfg.TargetPoints = 3
	g.SetConfig(cfg)
	g.SetPlayerScores([CalabresellaPlayerCnt]int{1, 0, 0})
	g.SetSoloistIdx(0)
	g.SetWinningBid(CalabresellaBidChiamo)
	g.SetTrickNumber(CalabresellaTrickCount)
	g.SetRoundThirds([CalabresellaPlayerCnt]int{17, 0, 0})
	g.SetPhase(CalabresellaPhaseTrickEnd)
	g.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 3, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 2, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 1, false)},
	}

	g.ResolveTrick()

	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, CalabresellaPhaseGameEnd, g.GetPhase())
	assert.True(t, g.GetSoloistWon())
	assert.Equal(t, [CalabresellaPlayerCnt]int{2, -1, -1}, g.GetRoundScoreChanges())
	settledScores := g.GetPlayerScores()
	assert.Equal(t, [CalabresellaPlayerCnt]int{3, -1, -1}, settledScores)

	g.ScoreRound()
	assert.Equal(t, settledScores, g.GetPlayerScores())
	assert.Equal(t, [CalabresellaPlayerCnt]int{2, -1, -1}, g.GetRoundScoreChanges())
}
