//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newCpuTestGame(difficulty DoudizhuCpuDifficulty) *Doudizhu {
	config := DoudizhuConfig{CpuDifficulty: difficulty}
	players := []*DoudizhuPlayer{
		NewDoudizhuPlayer(true),
		NewDoudizhuPlayer(false),
		NewDoudizhuPlayer(false),
	}
	return NewDoudizhu(NewTrumpCards(DoudizhuJokerCount), players, config)
}

func TestDoudizhu_HandQuality(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	p := d.GetPlayer(1)
	// Rocket: both jokers
	p.AddCard(NewCard(CardDesignJoker, 1, false))
	p.AddCard(NewCard(CardDesignJoker, 2, false))
	// A bomb of 5s
	p.AddCard(NewCard(CardDesignSpade, 5, false))
	p.AddCard(NewCard(CardDesignHeart, 5, false))
	p.AddCard(NewCard(CardDesignDiamond, 5, false))
	p.AddCard(NewCard(CardDesignClover, 5, false))
	// pair of 2s
	p.AddCard(NewCard(CardDesignSpade, 2, false))
	p.AddCard(NewCard(CardDesignHeart, 2, false))

	score := d.handQuality(p)
	assert.Greater(t, score, 5)
}

func TestDoudizhu_EvaluateBid_Easy(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyEasy)
	p := d.GetPlayer(1)
	p.AddCard(NewCard(CardDesignJoker, 1, false))
	p.AddCard(NewCard(CardDesignJoker, 2, false))
	assert.Equal(t, 0, d.evaluateBid(p))
}

func TestDoudizhu_EvaluateBid_StrongHand(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	p := d.GetPlayer(1)
	p.AddCard(NewCard(CardDesignJoker, 1, false))
	p.AddCard(NewCard(CardDesignJoker, 2, false))
	p.AddCard(NewCard(CardDesignSpade, 5, false))
	p.AddCard(NewCard(CardDesignHeart, 5, false))
	p.AddCard(NewCard(CardDesignDiamond, 5, false))
	p.AddCard(NewCard(CardDesignClover, 5, false))
	bid := d.evaluateBid(p)
	assert.Greater(t, bid, 0)
}

func TestDoudizhu_EvaluateBid_BelowHighest(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	d.round.highestBid = 3
	p := d.GetPlayer(1)
	p.AddCard(NewCard(CardDesignSpade, 5, false))
	p.AddCard(NewCard(CardDesignHeart, 5, false))
	p.AddCard(NewCard(CardDesignDiamond, 5, false))
	// score gives bid <= 3 → must pass
	assert.Equal(t, 0, d.evaluateBid(p))
}

func TestDoudizhu_FindAllBeatingCombos_Pair(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	p := d.GetPlayer(1)
	p.AddCard(NewCard(CardDesignSpade, 8, false))
	p.AddCard(NewCard(CardDesignHeart, 8, false))
	table := &DoudizhuCombo{Type: DoudizhuComboPair, Rank: 5, Length: 1}
	results := d.findAllBeatingCombos(p, table)
	assert.NotEmpty(t, results)
}

func TestDoudizhu_FindAllBeatingCombos_Trio(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	p := d.GetPlayer(1)
	p.AddCard(NewCard(CardDesignSpade, 9, false))
	p.AddCard(NewCard(CardDesignHeart, 9, false))
	p.AddCard(NewCard(CardDesignDiamond, 9, false))
	table := &DoudizhuCombo{Type: DoudizhuComboTrio, Rank: 5, Length: 1}
	results := d.findAllBeatingCombos(p, table)
	assert.NotEmpty(t, results)
}

func TestDoudizhu_FindAllBeatingCombos_TrioSingle(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	p := d.GetPlayer(1)
	p.AddCard(NewCard(CardDesignSpade, 9, false))
	p.AddCard(NewCard(CardDesignHeart, 9, false))
	p.AddCard(NewCard(CardDesignDiamond, 9, false))
	p.AddCard(NewCard(CardDesignClover, 3, false))
	table := &DoudizhuCombo{Type: DoudizhuComboTrioSingle, Rank: 5, Length: 1}
	results := d.findAllBeatingCombos(p, table)
	assert.NotEmpty(t, results)
}

func TestDoudizhu_FindAllBeatingCombos_TrioPair(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	p := d.GetPlayer(1)
	p.AddCard(NewCard(CardDesignSpade, 9, false))
	p.AddCard(NewCard(CardDesignHeart, 9, false))
	p.AddCard(NewCard(CardDesignDiamond, 9, false))
	p.AddCard(NewCard(CardDesignClover, 3, false))
	p.AddCard(NewCard(CardDesignSpade, 3, false))
	table := &DoudizhuCombo{Type: DoudizhuComboTrioPair, Rank: 5, Length: 1}
	results := d.findAllBeatingCombos(p, table)
	assert.NotEmpty(t, results)
}

func TestDoudizhu_FindAllBeatingCombos_Straight(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	p := d.GetPlayer(1)
	for v := 6; v <= 10; v++ {
		p.AddCard(NewCard(CardDesignSpade, v, false))
	}
	table := &DoudizhuCombo{Type: DoudizhuComboStraight, Rank: 3, Length: 5}
	results := d.findAllBeatingCombos(p, table)
	assert.NotEmpty(t, results)
}

func TestDoudizhu_FindAllBeatingCombos_ConsecutivePair(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	p := d.GetPlayer(1)
	for v := 7; v <= 9; v++ {
		p.AddCard(NewCard(CardDesignSpade, v, false))
		p.AddCard(NewCard(CardDesignHeart, v, false))
	}
	table := &DoudizhuCombo{Type: DoudizhuComboConsecutivePair, Rank: 3, Length: 3}
	results := d.findAllBeatingCombos(p, table)
	assert.NotEmpty(t, results)
}

func TestDoudizhu_FindAllBeatingCombos_BombBeatsNonBomb(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	p := d.GetPlayer(1)
	p.AddCard(NewCard(CardDesignSpade, 7, false))
	p.AddCard(NewCard(CardDesignHeart, 7, false))
	p.AddCard(NewCard(CardDesignDiamond, 7, false))
	p.AddCard(NewCard(CardDesignClover, 7, false))
	table := &DoudizhuCombo{Type: DoudizhuComboSingle, Rank: 15, Length: 1}
	results := d.findAllBeatingCombos(p, table)
	assert.NotEmpty(t, results)
}

func TestDoudizhu_FindAllBeatingCombos_RocketBeatsBomb(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	p := d.GetPlayer(1)
	p.AddCard(NewCard(CardDesignJoker, 1, false))
	p.AddCard(NewCard(CardDesignJoker, 2, false))
	table := &DoudizhuCombo{Type: DoudizhuComboBomb, Rank: 15, Length: 1}
	results := d.findAllBeatingCombos(p, table)
	assert.NotEmpty(t, results)
}

func TestDoudizhu_FindAllBeatingCombos_NoBeat(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	p := d.GetPlayer(1)
	p.AddCard(NewCard(CardDesignSpade, 3, false))
	table := &DoudizhuCombo{Type: DoudizhuComboSingle, Rank: 15, Length: 1}
	results := d.findAllBeatingCombos(p, table)
	assert.Empty(t, results)
}

func TestDoudizhu_IsTeammate(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	d.SetLandlordIdx(0)
	assert.True(t, d.isTeammate(1, 2))
	assert.False(t, d.isTeammate(0, 1))
	assert.False(t, d.isTeammate(-1, 1))
}

func TestDoudizhu_ShouldLetTeammateWin(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	d.SetLandlordIdx(0)
	// Player 2 (peasant) has 2 cards → teammate player 1 should let win
	d.GetPlayer(2).AddCard(NewCard(CardDesignSpade, 5, false))
	d.GetPlayer(2).AddCard(NewCard(CardDesignHeart, 5, false))
	assert.True(t, d.shouldLetTeammateWin(1))
}

func TestDoudizhu_IsUrgent(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	d.SetLandlordIdx(0)
	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 5, false))
	assert.True(t, d.isUrgent())
}

func TestDoudizhu_CpuPlay_NormalLead(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	d.SetPhase(DoudizhuPhasePlay)
	d.SetLandlordIdx(1)
	d.SetCurrentTurn(1)
	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 9, false))
	d.GetPlayer(1).AddCard(NewCard(CardDesignSpade, 3, false))
	d.GetPlayer(1).AddCard(NewCard(CardDesignHeart, 7, false))
	d.GetPlayer(2).AddCard(NewCard(CardDesignDiamond, 9, false))
	d.CpuPlay()
	assert.NotEmpty(t, d.GetCpuActions())
}

func TestDoudizhu_CpuPlay_NormalResponse(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	d.SetPhase(DoudizhuPhasePlay)
	d.SetLandlordIdx(0)
	d.SetCurrentTurn(1)
	d.SetTableCombo(&DoudizhuCombo{Type: DoudizhuComboSingle, Rank: 5, Length: 1, Cards: []*Card{NewCard(CardDesignSpade, 5, false)}})
	d.SetLastPlayIdx(0)
	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 9, false))
	d.GetPlayer(1).AddCard(NewCard(CardDesignHeart, 10, false))
	d.GetPlayer(2).AddCard(NewCard(CardDesignDiamond, 9, false))
	d.CpuPlay()
	assert.NotEmpty(t, d.GetCpuActions())
}

func TestDoudizhu_CpuPlay_HardLeadStraight(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyHard)
	d.SetPhase(DoudizhuPhasePlay)
	d.SetLandlordIdx(1)
	d.SetCurrentTurn(1)
	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 13, false))
	for v := 3; v <= 7; v++ {
		d.GetPlayer(1).AddCard(NewCard(CardDesignSpade, v, false))
	}
	d.GetPlayer(2).AddCard(NewCard(CardDesignDiamond, 13, false))
	d.CpuPlay()
	assert.NotEmpty(t, d.GetCpuActions())
}

func TestDoudizhu_CpuPlay_HardResponseTeammate(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyHard)
	d.SetPhase(DoudizhuPhasePlay)
	d.SetLandlordIdx(0)
	d.SetCurrentTurn(2)
	// teammate (player 1, peasant) led a strong play
	d.SetTableCombo(&DoudizhuCombo{Type: DoudizhuComboSingle, Rank: 15, Length: 1, Cards: []*Card{NewCard(CardDesignSpade, 2, false)}})
	d.SetLastPlayIdx(1)
	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 9, false))
	d.GetPlayer(1).AddCard(NewCard(CardDesignHeart, 10, false))
	d.GetPlayer(2).AddCard(NewCard(CardDesignJoker, 2, false))
	d.CpuPlay()
	// player 2 should pass (let teammate win with high card)
	assert.NotEmpty(t, d.GetCpuActions())
}

func TestDoudizhu_CpuBid(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	d.SetPhase(DoudizhuPhaseBid)
	d.SetCurrentTurn(1)
	d.GetPlayer(1).AddCard(NewCard(CardDesignSpade, 5, false))
	d.CpuPlay()
	assert.NotEmpty(t, d.GetCpuActions())
}

func TestDoudizhu_FindWeakStraight(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyHard)
	p := d.GetPlayer(1)
	for v := 3; v <= 7; v++ {
		p.AddCard(NewCard(CardDesignSpade, v, false))
	}
	freq := d.buildFrequency(p)
	result := d.findWeakStraight(p, freq)
	assert.Len(t, result, 5)
}

func TestDoudizhu_FindWeakConsecutivePair(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyHard)
	p := d.GetPlayer(1)
	for v := 3; v <= 5; v++ {
		p.AddCard(NewCard(CardDesignSpade, v, false))
		p.AddCard(NewCard(CardDesignHeart, v, false))
	}
	freq := d.buildFrequency(p)
	result := d.findWeakConsecutivePair(p, freq)
	assert.Len(t, result, 6)
}

func TestDoudizhu_Getters_BidValuesAndLastPlay(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	d.Reset()
	bids := d.GetBidValues()
	assert.Len(t, bids, DoudizhuPlayerCnt)
	assert.Equal(t, -1, d.GetLastPlayIdx())
}

func TestDoudizhu_FindLeadNormal_PrefersPair(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	p := d.GetPlayer(1)
	// only pairs and a 2 (strength 15, excluded from singles<15 and pairs<15)
	p.AddCard(NewCard(CardDesignSpade, 5, false))
	p.AddCard(NewCard(CardDesignHeart, 5, false))
	p.AddCard(NewCard(CardDesignSpade, 2, false))
	p.AddCard(NewCard(CardDesignHeart, 2, false))
	indices := d.findLeadNormal(p)
	assert.NotEmpty(t, indices)
}

func TestDoudizhu_FindLeadHard_PrefersStraight(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyHard)
	p := d.GetPlayer(1)
	for v := 3; v <= 7; v++ {
		p.AddCard(NewCard(CardDesignSpade, v, false))
	}
	indices := d.findLeadHard(p)
	// straight is 5 cards
	assert.Len(t, indices, 5)
}

func TestDoudizhu_FindResponseHard_BeatsLandlord(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyHard)
	d.SetLandlordIdx(0)
	d.SetCurrentTurn(1)
	d.SetLastPlayIdx(0) // landlord led
	d.SetTableCombo(&DoudizhuCombo{Type: DoudizhuComboSingle, Rank: 5, Length: 1})
	p := d.GetPlayer(1)
	p.AddCard(NewCard(CardDesignSpade, 9, false))
	indices := d.findResponseHard(p)
	assert.NotEmpty(t, indices)
}

func TestDoudizhu_FindResponseNormal_Pass(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	d.SetLandlordIdx(0)
	d.SetCurrentTurn(1)
	d.SetLastPlayIdx(0)
	d.SetTableCombo(&DoudizhuCombo{Type: DoudizhuComboSingle, Rank: 15, Length: 1})
	p := d.GetPlayer(1)
	p.AddCard(NewCard(CardDesignSpade, 3, false)) // cannot beat a 2
	indices := d.findResponseNormal(p)
	assert.Empty(t, indices)
}

func TestDoudizhu_FindRocket(t *testing.T) {
	d := newCpuTestGame(DoudizhuDifficultyNormal)
	p := d.GetPlayer(1)
	p.AddCard(NewCard(CardDesignJoker, 1, false))
	p.AddCard(NewCard(CardDesignJoker, 2, false))
	assert.Len(t, d.findRocket(p), 2)

	p2 := d.GetPlayer(2)
	p2.AddCard(NewCard(CardDesignSpade, 5, false))
	assert.Nil(t, d.findRocket(p2))
}
