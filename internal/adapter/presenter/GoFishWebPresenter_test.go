//go:build test

package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupGoFishMock() *interfaces.MockGoFishGame {
	m := new(interfaces.MockGoFishGame)
	m.On("GetPlayerCnt").Return(4)
	m.On("GetKnownRanks").Return(map[int][]int{}).Maybe()
	m.On("GetCurrentTurn").Return(0)
	m.On("GetPhase").Return(domain.GoFishPhasePlay)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetTurnNumber").Return(1)
	m.On("GetDeckRemaining").Return(32)
	m.On("GetConfig").Return(domain.DefaultGoFishConfig())
	m.On("GetLastAskPlayerIdx").Return(-1)
	m.On("GetLastAskTargetIdx").Return(-1)
	m.On("GetLastAskRank").Return(0)
	m.On("GetLastAskSuccess").Return(false)
	m.On("GetLastCardsReceived").Return(([]*domain.Card)(nil))
	m.On("GetLastDrawnCard").Return((*domain.Card)(nil))
	m.On("GetLastBookFormed").Return(false)
	m.On("GetLastBookRank").Return(0)
	m.On("GetCpuActions").Return(([]*domain.GoFishCpuAction)(nil))
	m.On("GetHumanAction").Return((*domain.GoFishCpuAction)(nil))
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("IsHumanTurn").Return(true)

	for i := range 4 {
		p := domain.NewGoFishPlayer(i == 0)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		m.On("GetPlayer", i).Return(p)
	}
	return m
}

func TestGoFishWebPresenter_Output_Initial(t *testing.T) {
	p := new(presenter.GoFishWebPresenter)
	m := setupGoFishMock()

	result := p.Output(m, nil)
	var out controller.GoFishWebOutput
	require.NoError(t, json.Unmarshal([]byte(result), &out))

	assert.Len(t, out.Players, 4)
	assert.Equal(t, 0, out.CurrentTurn)
	assert.False(t, out.GameEndFlag)
	assert.Equal(t, 32, out.DeckRemaining)
	assert.Equal(t, 1, out.TurnNumber)
	assert.Nil(t, out.LastAsk)
	assert.Empty(t, out.CpuActions)
	assert.Nil(t, out.HumanAction)
	assert.Empty(t, out.Message)
}

func TestGoFishWebPresenter_Output_WithError(t *testing.T) {
	p := new(presenter.GoFishWebPresenter)
	m := setupGoFishMock()

	result := p.Output(m, domain.ErrGoFishInvalidRank)
	var out controller.GoFishWebOutput
	require.NoError(t, json.Unmarshal([]byte(result), &out))

	assert.Contains(t, out.Message, "you do not hold that rank")
}

func TestGoFishWebPresenter_Output_GameEnd_HumanWin(t *testing.T) {
	p := new(presenter.GoFishWebPresenter)
	m := new(interfaces.MockGoFishGame)
	m.On("GetPlayerCnt").Return(4)
	m.On("GetKnownRanks").Return(map[int][]int{}).Maybe()
	m.On("GetCurrentTurn").Return(0)
	m.On("GetPhase").Return(domain.GoFishPhaseGameEnd)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerIdx").Return(0)
	m.On("GetTurnNumber").Return(20)
	m.On("GetDeckRemaining").Return(0)
	m.On("GetConfig").Return(domain.DefaultGoFishConfig())
	m.On("GetLastAskPlayerIdx").Return(-1)
	m.On("GetCpuActions").Return(([]*domain.GoFishCpuAction)(nil))
	m.On("GetHumanAction").Return((*domain.GoFishCpuAction)(nil))

	humanPlayer := domain.NewGoFishPlayer(true)
	humanPlayer.AddBook([]*domain.Card{
		domain.NewCard(1, 1, false), domain.NewCard(2, 1, false),
		domain.NewCard(3, 1, false), domain.NewCard(4, 1, false),
	})
	m.On("GetPlayer", 0).Return(humanPlayer)
	for i := 1; i < 4; i++ {
		m.On("GetPlayer", i).Return(domain.NewGoFishPlayer(false))
	}

	result := p.Output(m, nil)
	var out controller.GoFishWebOutput
	require.NoError(t, json.Unmarshal([]byte(result), &out))

	assert.True(t, out.GameEndFlag)
	assert.Equal(t, 0, out.WinnerIdx)
	assert.Equal(t, "gofish.result.humanWin", out.MessageCode)
}

func TestGoFishWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.GoFishWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockGoFishGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "ask_hit", Detail: "P0 asked P1 for rank 3"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"actionType":"ask_hit"`)
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockGoFishGame)
		m.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
	})
}

// The CUI has always shown these; the web rebuilt them client-side from lastAsk,
// which meant a reload erased the table's memory mid-game (#6312).
func TestGoFishWebPresenter_SendsKnownRanks(t *testing.T) {
	m := setupGoFishMock()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetKnownRanks")
	m.On("GetKnownRanks").Return(map[int][]int{0: {3, 7}, 2: {11}}).Maybe()

	p := new(presenter.GoFishWebPresenter)
	var out controller.GoFishWebOutput
	assert.NoError(t, json.Unmarshal([]byte(p.Output(m, nil)), &out))

	assert.Equal(t, []int{3, 7}, out.Players[0].KnownRanks)
	assert.Equal(t, []int{11}, out.Players[2].KnownRanks)
	// A seat nobody has learned anything about carries none -- not the previous
	// seat's list, which is what an off-by-one in the loop would produce.
	assert.Empty(t, out.Players[1].KnownRanks)
}
