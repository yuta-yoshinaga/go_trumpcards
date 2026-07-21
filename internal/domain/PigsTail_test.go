//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPigsTail() (*PigsTail, []*PigsTailPlayer) {
	players := []*PigsTailPlayer{
		NewPigsTailPlayer(true),
		NewPigsTailPlayer(false),
		NewPigsTailPlayer(false),
		NewPigsTailPlayer(false),
	}
	pt := NewPigsTail(NewTrumpCards(0), players)
	return pt, players
}

func TestNewPigsTail(t *testing.T) {
	pt, _ := newTestPigsTail()
	assert.False(t, pt.GetGameEndFlag())
	assert.Equal(t, -1, pt.GetLoserIdx())
	assert.Nil(t, pt.GetLastDrawCard())
	assert.False(t, pt.GetLastPenalty())
	assert.Equal(t, 4, pt.GetPlayerCnt())
	assert.Equal(t, 0, pt.GetCurrentTurn())
}

func TestPigsTail_Reset(t *testing.T) {
	pt, _ := newTestPigsTail()
	pt.Reset()

	assert.False(t, pt.GetGameEndFlag())
	assert.Equal(t, -1, pt.GetLoserIdx())
	assert.Equal(t, 52, pt.GetCircleCount())
	assert.Empty(t, pt.GetCenter())
	assert.Nil(t, pt.GetCenterTopCard())
}

func TestPigsTail_Reset_ClearsPlayerHands(t *testing.T) {
	pt, players := newTestPigsTail()
	players[0].AddCard(NewCard(CardDesignSpade, 1, false))
	pt.Reset()

	for i := 0; i < pt.GetPlayerCnt(); i++ {
		assert.Equal(t, 0, pt.GetPlayer(i).GetCardsSize())
	}
}

func TestPigsTail_PlayerAction_Basic(t *testing.T) {
	pt, _ := newTestPigsTail()
	pt.Reset()

	// Find the human player's turn
	for !pt.IsHumanTurn() && !pt.GetGameEndFlag() {
		err := pt.CpuAction()
		require.NoError(t, err)
	}
	if pt.GetGameEndFlag() {
		return
	}

	err := pt.PlayerAction(0)
	assert.NoError(t, err)
	assert.NotNil(t, pt.GetLastDrawCard())
	assert.NotNil(t, pt.GetHumanAction())
}

func TestPigsTail_PlayerAction_GameEnded(t *testing.T) {
	pt, _ := newTestPigsTail()
	pt.SetGameEndFlag(true)

	err := pt.PlayerAction(0)
	assert.ErrorIs(t, err, ErrGameEnded)
}

func TestPigsTail_PlayerAction_NotHumanTurn(t *testing.T) {
	pt, _ := newTestPigsTail()
	pt.Reset()

	// Set current turn to a CPU player
	for i, p := range []*PigsTailPlayer{pt.GetPlayer(0), pt.GetPlayer(1), pt.GetPlayer(2), pt.GetPlayer(3)} {
		if !p.GetIsHuman() {
			pt.SetCurrentTurn(i)
			break
		}
	}

	err := pt.PlayerAction(0)
	assert.ErrorIs(t, err, ErrNotHumanTurn)
}

func TestPigsTail_CpuAction_Basic(t *testing.T) {
	pt, _ := newTestPigsTail()
	pt.Reset()

	// Find a CPU player's turn
	for i := 0; i < pt.GetPlayerCnt(); i++ {
		if !pt.GetPlayer(i).GetIsHuman() {
			pt.SetCurrentTurn(i)
			break
		}
	}

	err := pt.CpuAction()
	assert.NoError(t, err)
	assert.NotNil(t, pt.GetCpuActions())
	assert.Len(t, pt.GetCpuActions(), 1)
}

func TestPigsTail_CpuAction_GameEnded(t *testing.T) {
	pt, _ := newTestPigsTail()
	pt.SetGameEndFlag(true)

	err := pt.CpuAction()
	assert.ErrorIs(t, err, ErrGameEnded)
}

func TestPigsTail_CpuAction_IsHumanTurn(t *testing.T) {
	pt, _ := newTestPigsTail()
	pt.Reset()

	// Set current turn to the human player
	for i := 0; i < pt.GetPlayerCnt(); i++ {
		if pt.GetPlayer(i).GetIsHuman() {
			pt.SetCurrentTurn(i)
			break
		}
	}

	err := pt.CpuAction()
	assert.ErrorIs(t, err, ErrNotHumanTurn)
}

func TestPigsTail_SuitMatchPenalty(t *testing.T) {
	pt, _ := newTestPigsTail()
	pt.Reset()

	penaltyOccurred := false
	for i := 0; i < 1000; i++ {
		pt2, _ := newTestPigsTail()
		pt2.Reset()
		// Play until penalty or 20 turns
		for turn := 0; turn < 20 && !pt2.GetGameEndFlag(); turn++ {
			if pt2.IsHumanTurn() {
				_ = pt2.PlayerAction(0)
			} else {
				_ = pt2.CpuAction()
			}
			if pt2.GetLastPenalty() {
				penaltyOccurred = true
				break
			}
		}
		if penaltyOccurred {
			break
		}
	}
	assert.True(t, penaltyOccurred, "penalty should occur at least once in 1000 games")
}

func TestPigsTail_GameEndCondition(t *testing.T) {
	pt, _ := newTestPigsTail()
	pt.Reset()

	for !pt.GetGameEndFlag() {
		if pt.IsHumanTurn() {
			_ = pt.PlayerAction(0)
		} else {
			_ = pt.CpuAction()
		}
	}

	assert.True(t, pt.GetGameEndFlag())
	assert.GreaterOrEqual(t, pt.GetLoserIdx(), 0)
	assert.Less(t, pt.GetLoserIdx(), pt.GetPlayerCnt())
	assert.Equal(t, 0, pt.GetCircleCount())
}

func TestPigsTail_LoserHasMostCards(t *testing.T) {
	for i := 0; i < 10; i++ {
		pt, _ := newTestPigsTail()
		pt.Reset()

		for !pt.GetGameEndFlag() {
			if pt.IsHumanTurn() {
				_ = pt.PlayerAction(0)
			} else {
				_ = pt.CpuAction()
			}
		}

		loserCards := pt.GetPlayer(pt.GetLoserIdx()).GetCardsSize()
		for j := 0; j < pt.GetPlayerCnt(); j++ {
			assert.LessOrEqual(t, pt.GetPlayer(j).GetCardsSize(), loserCards,
				"loser should have the most cards")
		}
	}
}

func TestPigsTail_GetPlayer_OutOfBounds(t *testing.T) {
	pt, _ := newTestPigsTail()
	assert.Nil(t, pt.GetPlayer(-1))
	assert.Nil(t, pt.GetPlayer(100))
}

func TestPigsTail_SetConfig(t *testing.T) {
	pt, _ := newTestPigsTail()
	cfg := PigsTailConfig{CpuHesitationEnabled: true}
	pt.SetConfig(cfg)
	assert.Equal(t, cfg, pt.GetConfig())
}

func TestPigsTail_Reset_RebuildsRosterFromConfig(t *testing.T) {
	pt, _ := newTestPigsTail()
	for _, count := range []int{PigsTailMinPlayers, 3, PigsTailMaxPlayers} {
		pt.SetConfig(PigsTailConfig{PlayerCount: count})
		pt.Reset()
		assert.Equal(t, count, pt.GetPlayerCnt())
		// 常に人間がちょうど1人。
		humans := 0
		for i := 0; i < pt.GetPlayerCnt(); i++ {
			if pt.GetPlayer(i).GetIsHuman() {
				humans++
			}
		}
		assert.Equal(t, 1, humans)
	}
}

func TestPigsTail_Reset_ClampsOutOfRangePlayerCount(t *testing.T) {
	pt, _ := newTestPigsTail()
	pt.SetConfig(PigsTailConfig{PlayerCount: 99})
	pt.Reset()
	assert.Equal(t, PigsTailMaxPlayers, pt.GetPlayerCnt())
}

func TestPigsTail_CpuHesitation(t *testing.T) {
	pt, _ := newTestPigsTail()
	pt.SetConfig(PigsTailConfig{CpuHesitationEnabled: true, PlayerCount: PigsTailPlayerCnt})
	pt.Reset()

	// Find a CPU player's turn
	for i := 0; i < pt.GetPlayerCnt(); i++ {
		if !pt.GetPlayer(i).GetIsHuman() {
			pt.SetCurrentTurn(i)
			break
		}
	}

	err := pt.CpuAction()
	require.NoError(t, err)
	assert.Greater(t, pt.GetCpuActions()[0].HesitationMs, 0)
}

func TestPigsTail_ActionLog(t *testing.T) {
	pt, _ := newTestPigsTail()
	pt.Reset()
	assert.Empty(t, pt.GetActionLog())

	// Play one turn
	if pt.IsHumanTurn() {
		_ = pt.PlayerAction(0)
	} else {
		_ = pt.CpuAction()
	}
	assert.NotEmpty(t, pt.GetActionLog())
}

func TestPigsTail_JSONRoundTrip(t *testing.T) {
	pt, _ := newTestPigsTail()
	pt.Reset()

	// Play a few turns
	for i := 0; i < 5 && !pt.GetGameEndFlag(); i++ {
		if pt.IsHumanTurn() {
			_ = pt.PlayerAction(0)
		} else {
			_ = pt.CpuAction()
		}
	}

	data, err := json.Marshal(pt)
	require.NoError(t, err)

	var restored PigsTail
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, pt.GetGameEndFlag(), restored.GetGameEndFlag())
	assert.Equal(t, pt.GetCurrentTurn(), restored.GetCurrentTurn())
	assert.Equal(t, pt.GetLoserIdx(), restored.GetLoserIdx())
	assert.Equal(t, pt.GetCircleCount(), restored.GetCircleCount())
	assert.Equal(t, len(pt.GetCenter()), len(restored.GetCenter()))
	assert.Equal(t, pt.GetPlayerCnt(), restored.GetPlayerCnt())
}

func TestPigsTail_UnmarshalJSON_NilFields(t *testing.T) {
	data := []byte(`{"tc":null,"ce":null,"pl":null,"al":null}`)
	var pt PigsTail
	err := json.Unmarshal(data, &pt)
	require.NoError(t, err)
	assert.NotNil(t, pt.center)
	assert.NotNil(t, pt.players)
	assert.NotNil(t, pt.actionLog)
}

func TestPigsTail_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var pt PigsTail
	err := json.Unmarshal([]byte(`{invalid`), &pt)
	assert.Error(t, err)
}

func TestPigsTail_UnmarshalJSON_OversizedArray(t *testing.T) {
	// Create a JSON with too many players
	players := make([]json.RawMessage, 1001)
	for i := range players {
		players[i] = []byte(`{"gp":{"p":{"c":[]},"h":false,"f":false}}`)
	}
	playersJSON, _ := json.Marshal(players)
	data := []byte(`{"pl":` + string(playersJSON) + `}`)

	var pt PigsTail
	err := json.Unmarshal(data, &pt)
	assert.Error(t, err)
}

func TestPigsTailCpuAction_JSONRoundTrip(t *testing.T) {
	action := &PigsTailCpuAction{
		DrawPlayerIdx: 1,
		DrawnCard:     NewCard(CardDesignHeart, 10, false),
		PenaltyFlag:   true,
		PenaltyCount:  5,
		HesitationMs:  300,
	}
	data, err := json.Marshal(action)
	require.NoError(t, err)

	var restored PigsTailCpuAction
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)
	assert.Equal(t, action.DrawPlayerIdx, restored.DrawPlayerIdx)
	assert.Equal(t, action.PenaltyFlag, restored.PenaltyFlag)
	assert.Equal(t, action.PenaltyCount, restored.PenaltyCount)
	assert.Equal(t, action.HesitationMs, restored.HesitationMs)
}

func TestPigsTailCpuAction_UnmarshalJSON_Invalid(t *testing.T) {
	var a PigsTailCpuAction
	err := json.Unmarshal([]byte(`{invalid`), &a)
	assert.Error(t, err)
}

func TestCardShortStr(t *testing.T) {
	tests := []struct {
		name string
		card *Card
		want string
	}{
		{"spade ace", NewCard(CardDesignSpade, 1, false), "S1"},
		{"heart 5", NewCard(CardDesignHeart, 5, false), "H5"},
		{"diamond king", NewCard(CardDesignDiamond, 13, false), "D13"},
		{"clover 10", NewCard(CardDesignClover, 10, false), "C10"},
		{"nil card", nil, "?"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, cardShortStr(tt.card))
		})
	}
}

func TestPigsTail_PlayerAction_ClearsAndResetseCpuActions(t *testing.T) {
	pt, _ := newTestPigsTail()
	pt.Reset()

	// Run some CPU turns first
	for !pt.IsHumanTurn() && !pt.GetGameEndFlag() {
		_ = pt.CpuAction()
	}
	if pt.GetGameEndFlag() {
		return
	}

	// There should be some CPU actions recorded
	cpuActionsBefore := pt.GetCpuActions()
	if len(cpuActionsBefore) == 0 {
		// Human is first, skip this test
		return
	}

	// Human action should clear CPU actions
	_ = pt.PlayerAction(0)
	assert.Nil(t, pt.GetCpuActions(), "CPU actions should be cleared after human action")
}
