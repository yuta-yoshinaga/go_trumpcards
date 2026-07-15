//go:build test

package presenter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func TestGoFishCuiPresenter_Output_Initial(t *testing.T) {
	p := new(presenter.GoFishCuiPresenter)
	m := setupGoFishMock()

	result := p.Output(m, nil)
	assert.Contains(t, result, "Go Fish")
	assert.Contains(t, result, "山札: 32枚")
	assert.Contains(t, result, "のターン")
}

func TestGoFishCuiPresenter_Output_KnownRanks(t *testing.T) {
	p := new(presenter.GoFishCuiPresenter)
	m := setupGoFishMock()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetKnownRanks")
	m.On("GetKnownRanks").Return(map[int][]int{1: {1, 13}})
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayer")
	human := domain.NewGoFishPlayer(true)
	human.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false)) // human also holds the Ace
	m.On("GetPlayer", 0).Return(human)
	for i := 1; i < 4; i++ {
		m.On("GetPlayer", i).Return(domain.NewGoFishPlayer(false))
	}

	result := p.Output(m, nil)
	assert.Contains(t, result, "既知ランク")
	// Ranks render as card-face labels: Ace starred (human holds it), King as K.
	assert.Contains(t, result, "A* K")
	assert.NotContains(t, result, "13")
}

func TestGoFishCuiPresenter_Output_GameEnd(t *testing.T) {
	p := new(presenter.GoFishCuiPresenter)
	m := new(interfaces.MockGoFishGame)
	m.On("GetPlayerCnt").Return(4)
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
	m.On("GetPlayer", 0).Return(humanPlayer)
	for i := 1; i < 4; i++ {
		m.On("GetPlayer", i).Return(domain.NewGoFishPlayer(false))
	}

	result := p.Output(m, nil)
	assert.Contains(t, result, "あなたの勝ち")
}

func TestGoFishCuiPresenter_Output_Error(t *testing.T) {
	p := new(presenter.GoFishCuiPresenter)
	m := setupGoFishMock()

	result := p.Output(m, domain.ErrGoFishInvalidRank)
	assert.Contains(t, result, "you do not hold that rank")
}

// TestGoFishCuiPresenter_Output_AskSuccess exercises the last-ask success
// branch and the optional bookFormed line.
func TestGoFishCuiPresenter_Output_AskSuccess(t *testing.T) {
	p := new(presenter.GoFishCuiPresenter)
	m := setupGoFishMock()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastAskPlayerIdx")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastAskTargetIdx")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastAskRank")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastAskSuccess")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastCardsReceived")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastBookFormed")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastBookRank")
	m.On("GetLastAskPlayerIdx").Return(0)
	m.On("GetLastAskTargetIdx").Return(1)
	m.On("GetLastAskRank").Return(13)
	m.On("GetLastAskSuccess").Return(true)
	m.On("GetLastCardsReceived").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, false),
	})
	m.On("GetLastBookFormed").Return(true)
	m.On("GetLastBookRank").Return(1)

	result := p.Output(m, nil)
	// King asked, Ace booked → both render as card-face labels, not 13/1.
	assert.Contains(t, result, "ランク K")
	assert.Contains(t, result, "1枚もらった")
	assert.Contains(t, result, "ブック完成")
	assert.NotContains(t, result, "ランク 13")
}

// TestGoFishCuiPresenter_Output_AskFail exercises the Go-Fish failure branch
// of the last-ask block.
func TestGoFishCuiPresenter_Output_AskFail(t *testing.T) {
	p := new(presenter.GoFishCuiPresenter)
	m := setupGoFishMock()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastAskPlayerIdx")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastAskTargetIdx")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastAskRank")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastAskSuccess")
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLastCardsReceived")
	m.On("GetLastAskPlayerIdx").Return(0)
	m.On("GetLastAskTargetIdx").Return(1)
	m.On("GetLastAskRank").Return(11)
	m.On("GetLastAskSuccess").Return(false)
	m.On("GetLastCardsReceived").Return(([]*domain.Card)(nil))

	result := p.Output(m, nil)
	assert.Contains(t, result, "Go Fish!")
}

// TestGoFishCuiPresenter_Output_CpuActions exercises the CPU-action loop with
// a success entry and a failure entry that completes a book.
func TestGoFishCuiPresenter_Output_CpuActions(t *testing.T) {
	p := new(presenter.GoFishCuiPresenter)
	m := setupGoFishMock()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCpuActions")
	m.On("GetCpuActions").Return([]*domain.GoFishCpuAction{
		{AskPlayerIdx: 1, AskTargetIdx: 2, AskRank: 5, Success: true, CardsReceived: 2, BookFormed: true, BookRank: 5},
		{AskPlayerIdx: 2, AskTargetIdx: 0, AskRank: 9, Success: false, BookFormed: false},
	})

	result := p.Output(m, nil)
	assert.Contains(t, result, "[CPU]")
	assert.Contains(t, result, "ランク 5")
	assert.Contains(t, result, "ランク 9")
	assert.Contains(t, result, "Go Fish!")
	assert.Contains(t, result, "ブック完成")
}

// TestGoFishCuiPresenter_Output_CpuWins covers the non-human winner path of
// the game-end banner + score entries.
func TestGoFishCuiPresenter_Output_CpuWins(t *testing.T) {
	p := new(presenter.GoFishCuiPresenter)
	m := new(interfaces.MockGoFishGame)
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPhase").Return(domain.GoFishPhaseGameEnd)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetWinnerIdx").Return(2)
	m.On("GetDeckRemaining").Return(0)
	m.On("GetLastAskPlayerIdx").Return(-1)
	m.On("GetCpuActions").Return(([]*domain.GoFishCpuAction)(nil))

	humanPlayer := domain.NewGoFishPlayer(true)
	m.On("GetPlayer", 0).Return(humanPlayer)
	for i := 1; i < 4; i++ {
		m.On("GetPlayer", i).Return(domain.NewGoFishPlayer(false))
	}

	result := p.Output(m, nil)
	assert.Contains(t, result, "CPU 2の勝ち")
}

func TestGoFishCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.GoFishCuiPresenter)
	m := new(interfaces.MockGoFishGame)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "棋譜はありません")
}
