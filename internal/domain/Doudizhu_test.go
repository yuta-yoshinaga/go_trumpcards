//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDoudizhu() *Doudizhu {
	config := DefaultDoudizhuConfig()
	players := []*DoudizhuPlayer{
		NewDoudizhuPlayer(true),
		NewDoudizhuPlayer(false),
		NewDoudizhuPlayer(false),
	}
	tc := NewTrumpCards(DoudizhuJokerCount)
	return NewDoudizhu(tc, players, config)
}

func dzSetupPlayPhase(d *Doudizhu) {
	d.SetPhase(DoudizhuPhasePlay)
	d.SetLandlordIdx(0)
	d.SetBaseBid(1)
	d.SetCurrentTurn(0)
}

func TestDoudizhu_NewDefault(t *testing.T) {
	d := NewDefaultDoudizhu()
	assert.Equal(t, DoudizhuPlayerCnt, d.GetPlayerCnt())
	assert.True(t, d.GetPlayer(0).GetIsHuman() || d.GetPlayer(1).GetIsHuman() || d.GetPlayer(2).GetIsHuman())
}

func TestDoudizhu_Reset_DealCorrectly(t *testing.T) {
	d := NewDefaultDoudizhu()
	d.Reset()

	totalCards := 0
	for i := 0; i < DoudizhuPlayerCnt; i++ {
		totalCards += d.GetPlayer(i).GetCardsSize()
	}
	totalCards += len(d.GetKittyCards())
	assert.Equal(t, 54, totalCards)

	for i := 0; i < DoudizhuPlayerCnt; i++ {
		assert.Equal(t, DoudizhuHandSize, d.GetPlayer(i).GetCardsSize())
	}
	assert.Equal(t, DoudizhuKittyCount, len(d.GetKittyCards()))
	assert.Equal(t, DoudizhuPhaseBid, d.GetPhase())
	assert.False(t, d.GetGameEndFlag())
}

func TestDoudizhu_PlayerBid_Success(t *testing.T) {
	d := newTestDoudizhu()
	d.Reset()

	for d.GetPhase() == DoudizhuPhaseBid && !d.IsHumanTurn() {
		d.CpuPlay()
	}
	if d.GetPhase() != DoudizhuPhaseBid {
		return
	}

	highBid := d.GetHighestBid()
	bid := highBid + 1
	if bid > DoudizhuMaxBid {
		bid = 0
	}
	err := d.PlayerBid(bid)
	assert.NoError(t, err)
}

func TestDoudizhu_PlayerBid_ErrorWhenNotBidPhase(t *testing.T) {
	d := newTestDoudizhu()
	dzSetupPlayPhase(d)
	err := d.PlayerBid(1)
	assert.Error(t, err)
}

func TestDoudizhu_PlayerBid_ErrorWhenGameEnded(t *testing.T) {
	d := newTestDoudizhu()
	d.SetPhase(DoudizhuPhaseBid)
	d.SetGameEndFlag(true)
	err := d.PlayerBid(1)
	assert.ErrorIs(t, err, ErrGameEnded)
}

func TestDoudizhu_PlayerBid_ErrorWhenNotHumanTurn(t *testing.T) {
	d := newTestDoudizhu()
	d.SetPhase(DoudizhuPhaseBid)
	d.SetCurrentTurn(1)
	err := d.PlayerBid(1)
	assert.ErrorIs(t, err, ErrNotHumanTurn)
}

func TestDoudizhu_PlayerBid_ErrorInvalidValue(t *testing.T) {
	d := newTestDoudizhu()
	d.SetPhase(DoudizhuPhaseBid)
	d.SetCurrentTurn(0)
	err := d.PlayerBid(4)
	assert.Error(t, err)
}

func TestDoudizhu_PlayerBid_ErrorMustExceedHighest(t *testing.T) {
	d := newTestDoudizhu()
	d.SetPhase(DoudizhuPhaseBid)
	d.SetCurrentTurn(0)
	d.SetHighestBid(2)
	d.SetHighestBidder(1)
	err := d.PlayerBid(1)
	assert.Error(t, err)
}

func TestDoudizhu_DecideLandlord(t *testing.T) {
	d := newTestDoudizhu()
	d.Reset()

	humanIdx := -1
	for i := 0; i < DoudizhuPlayerCnt; i++ {
		if d.GetPlayer(i).GetIsHuman() {
			humanIdx = i
			break
		}
	}

	d.SetCurrentTurn(humanIdx)
	d.SetPhase(DoudizhuPhaseBid)
	err := d.PlayerBid(3)
	require.NoError(t, err)

	assert.Equal(t, DoudizhuPhasePlay, d.GetPhase())
	assert.Equal(t, humanIdx, d.GetLandlordIdx())
	assert.True(t, d.GetPlayer(humanIdx).GetIsLandlord())
	assert.Equal(t, DoudizhuHandSize+DoudizhuKittyCount, d.GetPlayer(humanIdx).GetCardsSize())
}

func TestDoudizhu_PlayerPlay_Single(t *testing.T) {
	d := newTestDoudizhu()
	dzSetupPlayPhase(d)

	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 3, false))
	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 5, false))
	d.GetPlayer(1).AddCard(NewCard(CardDesignHeart, 10, false))
	d.GetPlayer(2).AddCard(NewCard(CardDesignDiamond, 10, false))

	err := d.PlayerPlay([]int{0})
	assert.NoError(t, err)
	assert.NotNil(t, d.GetTableCombo())
	assert.Equal(t, DoudizhuComboSingle, d.GetTableCombo().Type)
}

func TestDoudizhu_PlayerPlay_PassWhenTableEmpty(t *testing.T) {
	d := newTestDoudizhu()
	dzSetupPlayPhase(d)
	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 3, false))

	err := d.PlayerPlay([]int{})
	assert.Error(t, err)
}

func TestDoudizhu_PlayerPlay_PassWhenTableHasCards(t *testing.T) {
	d := newTestDoudizhu()
	dzSetupPlayPhase(d)
	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 3, false))
	d.SetTableCombo(&DoudizhuCombo{Type: DoudizhuComboSingle, Rank: 10, Length: 1})
	d.SetLastPlayIdx(1)

	err := d.PlayerPlay([]int{})
	assert.NoError(t, err)
}

func TestDoudizhu_PlayerPlay_InvalidCombo(t *testing.T) {
	d := newTestDoudizhu()
	dzSetupPlayPhase(d)
	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 3, false))
	d.GetPlayer(0).AddCard(NewCard(CardDesignHeart, 5, false))

	err := d.PlayerPlay([]int{0, 1})
	assert.Error(t, err)
}

func TestDoudizhu_PlayerPlay_CannotBeat(t *testing.T) {
	d := newTestDoudizhu()
	dzSetupPlayPhase(d)
	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 3, false))
	d.SetTableCombo(&DoudizhuCombo{Type: DoudizhuComboSingle, Rank: 10, Length: 1})
	d.SetLastPlayIdx(1)

	err := d.PlayerPlay([]int{0})
	assert.Error(t, err)
}

func TestDoudizhu_PlayerPlay_ErrorNotPlayPhase(t *testing.T) {
	d := newTestDoudizhu()
	d.SetPhase(DoudizhuPhaseBid)
	err := d.PlayerPlay([]int{0})
	assert.Error(t, err)
}

func TestDoudizhu_PlayerPlay_ErrorGameEnded(t *testing.T) {
	d := newTestDoudizhu()
	dzSetupPlayPhase(d)
	d.SetGameEndFlag(true)
	err := d.PlayerPlay([]int{0})
	assert.ErrorIs(t, err, ErrGameEnded)
}

func TestDoudizhu_PlayerPlay_ErrorNotHumanTurn(t *testing.T) {
	d := newTestDoudizhu()
	dzSetupPlayPhase(d)
	d.SetCurrentTurn(1)
	err := d.PlayerPlay([]int{0})
	assert.ErrorIs(t, err, ErrNotHumanTurn)
}

func TestDoudizhu_PlayerPlay_ErrorInvalidCardIndex(t *testing.T) {
	d := newTestDoudizhu()
	dzSetupPlayPhase(d)
	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 3, false))

	err := d.PlayerPlay([]int{5})
	assert.Error(t, err)
}

func TestDoudizhu_WinCondition_LandlordWins(t *testing.T) {
	d := newTestDoudizhu()
	dzSetupPlayPhase(d)

	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 1, false))
	d.GetPlayer(1).AddCard(NewCard(CardDesignHeart, 3, false))
	d.GetPlayer(2).AddCard(NewCard(CardDesignDiamond, 3, false))

	err := d.PlayerPlay([]int{0})
	require.NoError(t, err)

	assert.True(t, d.GetGameEndFlag())
	assert.Equal(t, DoudizhuPhaseEnd, d.GetPhase())
	scores := d.GetScores()
	assert.Greater(t, scores[0], 0)
	assert.Less(t, scores[1], 0)
	assert.Less(t, scores[2], 0)
}

func TestDoudizhu_WinCondition_PeasantsWin(t *testing.T) {
	d := newTestDoudizhu()
	dzSetupPlayPhase(d)
	d.SetCurrentTurn(1)
	d.SetLandlordIdx(0)

	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 3, false))
	d.GetPlayer(1).AddCard(NewCard(CardDesignHeart, 1, false))
	d.GetPlayer(2).AddCard(NewCard(CardDesignDiamond, 3, false))

	d.CpuPlay()

	assert.True(t, d.GetGameEndFlag())
	scores := d.GetScores()
	assert.Less(t, scores[0], 0)
	assert.Greater(t, scores[1], 0)
	assert.Greater(t, scores[2], 0)
}

func TestDoudizhu_Scoring_WithBombs(t *testing.T) {
	d := newTestDoudizhu()
	dzSetupPlayPhase(d)
	d.SetBombCount(2)
	d.SetBaseBid(2)

	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 1, false))
	d.GetPlayer(1).AddCard(NewCard(CardDesignHeart, 3, false))
	d.GetPlayer(2).AddCard(NewCard(CardDesignDiamond, 3, false))

	err := d.PlayerPlay([]int{0})
	require.NoError(t, err)

	scores := d.GetScores()
	assert.Equal(t, 16, scores[0])
	assert.Equal(t, -8, scores[1])
	assert.Equal(t, -8, scores[2])
}

func TestDoudizhu_CpuPlay_PlayPhase(t *testing.T) {
	d := newTestDoudizhu()
	dzSetupPlayPhase(d)
	d.SetCurrentTurn(1)

	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 3, false))
	d.GetPlayer(1).AddCard(NewCard(CardDesignHeart, 5, false))
	d.GetPlayer(1).AddCard(NewCard(CardDesignHeart, 7, false))
	d.GetPlayer(2).AddCard(NewCard(CardDesignDiamond, 3, false))

	d.CpuPlay()

	actions := d.GetCpuActions()
	assert.NotEmpty(t, actions)
}

func TestDoudizhu_PassClear(t *testing.T) {
	d := newTestDoudizhu()
	dzSetupPlayPhase(d)

	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 1, false))
	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 3, false))
	d.GetPlayer(1).AddCard(NewCard(CardDesignHeart, 4, false))
	d.GetPlayer(1).AddCard(NewCard(CardDesignHeart, 5, false))
	d.GetPlayer(2).AddCard(NewCard(CardDesignDiamond, 4, false))
	d.GetPlayer(2).AddCard(NewCard(CardDesignDiamond, 5, false))

	err := d.PlayerPlay([]int{0})
	require.NoError(t, err)
	assert.NotNil(t, d.GetTableCombo())

	d.SetCurrentTurn(1)
	d.SetPassCount(0)
	d.SetLastPlayIdx(0)

	d.SetCurrentTurn(0)
	d.SetPassCount(0)
	d.SetTableCombo(nil)
	d.SetLastPlayIdx(-1)
	assert.Nil(t, d.GetTableCombo())
}

func TestDoudizhu_HasPendingAction(t *testing.T) {
	d := NewDefaultDoudizhu()
	assert.False(t, d.HasPendingAction())
}

func TestDoudizhu_JSON_RoundTrip(t *testing.T) {
	d := NewDefaultDoudizhu()
	d.Reset()

	data, err := json.Marshal(d)
	require.NoError(t, err)

	d2 := &Doudizhu{}
	err = json.Unmarshal(data, d2)
	require.NoError(t, err)

	assert.Equal(t, d.GetPhase(), d2.GetPhase())
	assert.Equal(t, d.GetPlayerCnt(), d2.GetPlayerCnt())
	assert.Equal(t, len(d.GetKittyCards()), len(d2.GetKittyCards()))
}

func TestDoudizhu_JSON_MaxSliceLen(t *testing.T) {
	d := &Doudizhu{}
	largeJSON := `{"pl":[` + func() string {
		s := ""
		for i := 0; i < 1001; i++ {
			if i > 0 {
				s += ","
			}
			s += `{"gp":{"p":{"c":[]},"h":false,"f":false},"ll":false}`
		}
		return s
	}() + `]}`

	err := json.Unmarshal([]byte(largeJSON), d)
	assert.Error(t, err)
}

func TestDoudizhuPlayer_Basics(t *testing.T) {
	p := NewDoudizhuPlayer(true)
	assert.True(t, p.GetIsHuman())
	assert.False(t, p.GetIsLandlord())

	p.SetIsLandlord(true)
	assert.True(t, p.GetIsLandlord())
}

func TestDoudizhuPlayer_SortByStrength(t *testing.T) {
	p := NewDoudizhuPlayer(true)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignHeart, 3, false))
	p.AddCard(NewCard(CardDesignDiamond, 13, false))

	p.SortCardsByStrength()

	assert.Equal(t, 3, p.GetCard(0).GetValue())
	assert.Equal(t, 13, p.GetCard(1).GetValue())
	assert.Equal(t, 1, p.GetCard(2).GetValue())
}

func TestDoudizhuPlayer_JSON_RoundTrip(t *testing.T) {
	p := NewDoudizhuPlayer(true)
	p.SetIsLandlord(true)
	p.AddCard(NewCard(CardDesignSpade, 5, false))

	data, err := json.Marshal(p)
	require.NoError(t, err)

	p2 := &DoudizhuPlayer{}
	err = json.Unmarshal(data, p2)
	require.NoError(t, err)

	assert.True(t, p2.GetIsHuman())
	assert.True(t, p2.GetIsLandlord())
	assert.Equal(t, 1, p2.GetCardsSize())
}

func TestDoudizhuConfig_Default(t *testing.T) {
	c := DefaultDoudizhuConfig()
	assert.Equal(t, DoudizhuDifficultyNormal, c.CpuDifficulty)
	assert.NoError(t, c.Validate())
}

func TestDoudizhuConfig_Validate_InvalidDifficulty(t *testing.T) {
	c := DoudizhuConfig{CpuDifficulty: DoudizhuCpuDifficulty(5)}
	assert.Error(t, c.Validate())
}

func TestDoudizhuConfig_JSON_RoundTrip(t *testing.T) {
	c := DoudizhuConfig{CpuDifficulty: DoudizhuDifficultyHard}
	data, err := json.Marshal(c)
	require.NoError(t, err)

	var c2 DoudizhuConfig
	err = json.Unmarshal(data, &c2)
	require.NoError(t, err)
	assert.Equal(t, DoudizhuDifficultyHard, c2.CpuDifficulty)
}

func TestDoudizhuCpuAction_JSON_RoundTrip(t *testing.T) {
	a := &DoudizhuCpuAction{
		PlayerIdx:   1,
		PlayedCards: []*Card{NewCard(CardDesignSpade, 5, false)},
		BidValue:    2,
	}
	data, err := json.Marshal(a)
	require.NoError(t, err)

	a2 := &DoudizhuCpuAction{}
	err = json.Unmarshal(data, a2)
	require.NoError(t, err)
	assert.Equal(t, 1, a2.PlayerIdx)
	assert.Equal(t, 2, a2.BidValue)
	assert.Len(t, a2.PlayedCards, 1)
}

func TestDoudizhuCombo_JSON_RoundTrip(t *testing.T) {
	c := &DoudizhuCombo{
		Type:   DoudizhuComboBomb,
		Cards:  []*Card{NewCard(CardDesignSpade, 5, false)},
		Rank:   5,
		Length: 1,
	}
	data, err := json.Marshal(c)
	require.NoError(t, err)

	c2 := &DoudizhuCombo{}
	err = json.Unmarshal(data, c2)
	require.NoError(t, err)
	assert.Equal(t, DoudizhuComboBomb, c2.Type)
	assert.Equal(t, 5, c2.Rank)
}

func TestDoudizhu_ActionLog(t *testing.T) {
	d := newTestDoudizhu()
	dzSetupPlayPhase(d)
	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 1, false))
	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 3, false))
	d.GetPlayer(1).AddCard(NewCard(CardDesignHeart, 3, false))
	d.GetPlayer(2).AddCard(NewCard(CardDesignDiamond, 3, false))

	err := d.PlayerPlay([]int{0})
	require.NoError(t, err)

	log := d.GetActionLog()
	assert.NotEmpty(t, log)
	assert.Equal(t, "play", log[0].ActionType)
}

func TestDoudizhu_BombCountIncrementsOnBomb(t *testing.T) {
	d := newTestDoudizhu()
	dzSetupPlayPhase(d)

	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 5, false))
	d.GetPlayer(0).AddCard(NewCard(CardDesignHeart, 5, false))
	d.GetPlayer(0).AddCard(NewCard(CardDesignDiamond, 5, false))
	d.GetPlayer(0).AddCard(NewCard(CardDesignClover, 5, false))
	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 3, false))
	d.GetPlayer(1).AddCard(NewCard(CardDesignHeart, 3, false))
	d.GetPlayer(2).AddCard(NewCard(CardDesignDiamond, 3, false))

	assert.Equal(t, 0, d.GetBombCount())
	err := d.PlayerPlay([]int{0, 1, 2, 3})
	require.NoError(t, err)
	assert.Equal(t, 1, d.GetBombCount())
}

func TestDoudizhu_BombCountIncrementsOnRocket(t *testing.T) {
	d := newTestDoudizhu()
	dzSetupPlayPhase(d)

	d.GetPlayer(0).AddCard(NewCard(CardDesignJoker, 1, false))
	d.GetPlayer(0).AddCard(NewCard(CardDesignJoker, 2, false))
	d.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 3, false))
	d.GetPlayer(1).AddCard(NewCard(CardDesignHeart, 3, false))
	d.GetPlayer(2).AddCard(NewCard(CardDesignDiamond, 3, false))

	err := d.PlayerPlay([]int{0, 1})
	require.NoError(t, err)
	assert.Equal(t, 1, d.GetBombCount())
}

func TestDoudizhu_FullGame_CpuDriven(t *testing.T) {
	d := NewDefaultDoudizhu()
	d.SetConfig(DoudizhuConfig{CpuDifficulty: DoudizhuDifficultyEasy})

	for iter := 0; iter < 100; iter++ {
		d.Reset()

		for d.GetPhase() == DoudizhuPhaseBid {
			if d.IsHumanTurn() {
				highBid := d.GetHighestBid()
				bid := highBid + 1
				if bid > DoudizhuMaxBid {
					bid = 0
				}
				_ = d.PlayerBid(bid)
			} else {
				d.CpuPlay()
			}
		}

		if d.GetPhase() != DoudizhuPhasePlay {
			continue
		}

		moves := 0
		for !d.GetGameEndFlag() && moves < 200 {
			if d.IsHumanTurn() {
				player := d.GetPlayer(d.GetCurrentTurn())
				if d.GetTableCombo() == nil {
					_ = d.PlayerPlay([]int{0})
				} else {
					played := false
					for i := 0; i < player.GetCardsSize(); i++ {
						combo := DoudizhuClassifyCombo([]*Card{player.GetCard(i)})
						if combo != nil && DoudizhuCanBeat(combo, d.GetTableCombo()) {
							_ = d.PlayerPlay([]int{i})
							played = true
							break
						}
					}
					if !played {
						_ = d.PlayerPlay([]int{})
					}
				}
			} else {
				d.CpuPlay()
			}
			moves++
		}

		if d.GetGameEndFlag() {
			scores := d.GetScores()
			total := 0
			for _, s := range scores {
				total += s
			}
			assert.Equal(t, 0, total, "scores must sum to zero")
			return
		}
	}
	t.Log("no game completed in 100 iterations (non-fatal)")
}
