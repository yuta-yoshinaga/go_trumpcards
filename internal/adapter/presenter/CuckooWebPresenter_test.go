//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupCuckooWebMock() (*interfaces.MockCuckooGame, []*domain.CuckooPlayer) {
	m := new(interfaces.MockCuckooGame)
	players := makeCuckooPlayers()
	players[0].SetCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	for i := 1; i < 4; i++ {
		players[i].SetCard(domain.NewCard(domain.CardDesignHeart, 9, false))
	}
	m.On("GetPhase").Return(domain.CuckooPhaseTurn)
	m.On("GetRoundNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(3)
	m.On("GetStockCount").Return(40)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetPendingSwapFrom").Return(-1)
	m.On("GetPendingSwapTo").Return(-1)
	m.On("GetRoundLowest").Return(-1)
	m.On("GetRoundLosers").Return([]int{})
	m.On("GetConfig").Return(domain.DefaultCuckooConfig())
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
		m.On("IsKingRevealed", i).Return(false)
	}
	return m, players
}

func parseCuckooOutput(t *testing.T, s string) *controller.CuckooWebOutput {
	t.Helper()
	var out controller.CuckooWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return &out
}

func TestCuckooWebPresenter_Output(t *testing.T) {
	p := new(presenter.CuckooWebPresenter)

	t.Run("turn phase hides opponents", func(t *testing.T) {
		m, _ := setupCuckooWebMock()
		out := parseCuckooOutput(t, p.Output(m, nil))
		assert.Equal(t, int(domain.CuckooPhaseTurn), out.Phase)
		assert.Len(t, out.Players, 4)
		assert.NotNil(t, out.Players[0].Card) // human card shown
		assert.Nil(t, out.Players[1].Card)    // opponent hidden
		assert.True(t, out.Players[0].IsCurrentTurn)
		assert.Equal(t, "cuckoo.turnPhase", out.MessageCode)
		assert.Equal(t, -1, out.RoundLowest) // undecided outside round end
	})

	t.Run("error message", func(t *testing.T) {
		m, _ := setupCuckooWebMock()
		out := parseCuckooOutput(t, p.Output(m, errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
	})

	t.Run("round end reveals all and lowest value", func(t *testing.T) {
		m, _ := setupCuckooWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CuckooPhaseRoundEnd)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundLowest")
		m.On("GetRoundLowest").Return(5)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundLosers")
		m.On("GetRoundLosers").Return([]int{0})
		out := parseCuckooOutput(t, p.Output(m, nil))
		assert.NotNil(t, out.Players[1].Card)
		assert.Equal(t, "cuckoo.roundEnd", out.MessageCode)
		assert.Equal(t, 5, out.RoundLowest) // populated at round end
		assert.Equal(t, []int{0}, out.RoundLosers)
	})

	t.Run("game end human win", func(t *testing.T) {
		m, _ := setupCuckooWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)
		out := parseCuckooOutput(t, p.Output(m, nil))
		assert.True(t, out.GameEndFlag)
		assert.Equal(t, "cuckoo.result.humanWin", out.MessageCode)
	})

	t.Run("game end cpu win", func(t *testing.T) {
		m, _ := setupCuckooWebMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(2)
		out := parseCuckooOutput(t, p.Output(m, nil))
		assert.Equal(t, "cuckoo.result.cpuWin", out.MessageCode)
		assert.Equal(t, "2", out.MessageParams["cpuId"])
	})
}

func TestCuckooWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.CuckooWebPresenter)
	m := new(interfaces.MockCuckooGame)
	entries := []*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "swap", Detail: "swap"},
	}
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return(entries)
	assert.Contains(t, p.ActionLogOutput(m), "swap")
}
