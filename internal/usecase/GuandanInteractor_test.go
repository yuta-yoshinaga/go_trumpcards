//go:build test

package usecase_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

const guandanMockOutput = `{"phase":1}`

func TestNewGuandanInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockGuandanPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "GuandanInteractor: g must not be nil", func() {
			usecase.NewGuandanInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockGuandanGame)
		assert.PanicsWithValue(t, "GuandanInteractor: gp must not be nil", func() {
			usecase.NewGuandanInteractor(gameMock, nil)
		})
	})
}

func TestGuandanInteractor_Reset(t *testing.T) {
	pMock := new(presenter.MockGuandanPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(guandanMockOutput)
	gameMock := new(interfaces.MockGuandanGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.GuandanPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	gi := usecase.NewGuandanInteractor(gameMock, pMock)
	assert.Equal(t, guandanMockOutput, gi.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestGuandanInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockGuandanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(guandanMockOutput)
		gameMock := new(interfaces.MockGuandanGame)
		cfg := domain.DefaultGuandanConfig()
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.GuandanPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		gi := usecase.NewGuandanInteractor(gameMock, pMock)
		assert.Equal(t, guandanMockOutput, gi.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config never reaches the game", func(t *testing.T) {
		pMock := new(presenter.MockGuandanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(guandanMockOutput)
		gameMock := new(interfaces.MockGuandanGame)

		gi := usecase.NewGuandanInteractor(gameMock, pMock)
		assert.Equal(t, guandanMockOutput, gi.ResetWithConfig(domain.GuandanConfig{CpuDifficulty: 9}))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
		gameMock.AssertNotCalled(t, "Reset")
	})
}

// **どのアクションも currentIdx の席として実行される。**
func TestGuandanInteractor_UsesTheCurrentSeat(t *testing.T) {
	newMocks := func() (*presenter.MockGuandanPresenter, *interfaces.MockGuandanGame) {
		pMock := new(presenter.MockGuandanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(guandanMockOutput)
		gameMock := new(interfaces.MockGuandanGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.GuandanPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("GetCurrentPlayerIdx").Return(2)
		return pMock, gameMock
	}

	t.Run("play", func(t *testing.T) {
		pMock, gameMock := newMocks()
		gameMock.On("PlayCards", 2, []int{0, 1, 2}).Return(nil)

		gi := usecase.NewGuandanInteractor(gameMock, pMock)
		assert.Equal(t, guandanMockOutput, gi.PlayCards([]int{0, 1, 2}))
		gameMock.AssertCalled(t, "PlayCards", 2, []int{0, 1, 2})
	})

	t.Run("pass", func(t *testing.T) {
		pMock, gameMock := newMocks()
		gameMock.On("Pass", 2).Return(nil)

		gi := usecase.NewGuandanInteractor(gameMock, pMock)
		assert.Equal(t, guandanMockOutput, gi.Pass())
		gameMock.AssertCalled(t, "Pass", 2)
	})

	t.Run("return tribute", func(t *testing.T) {
		pMock, gameMock := newMocks()
		gameMock.On("ReturnTribute", 2, 5).Return(nil)

		gi := usecase.NewGuandanInteractor(gameMock, pMock)
		assert.Equal(t, guandanMockOutput, gi.ReturnTribute(5))
		gameMock.AssertCalled(t, "ReturnTribute", 2, 5)
	})
}

func TestGuandanInteractor_BlockedWhenNotTheHumansTurn(t *testing.T) {
	pMock := new(presenter.MockGuandanPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(guandanMockOutput)
	gameMock := new(interfaces.MockGuandanGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	gi := usecase.NewGuandanInteractor(gameMock, pMock)
	assert.Equal(t, guandanMockOutput, gi.PlayCards([]int{0}))
	assert.Equal(t, guandanMockOutput, gi.Pass())
	assert.Equal(t, guandanMockOutput, gi.ReturnTribute(0))
	gameMock.AssertNotCalled(t, "PlayCards", mock.Anything, mock.Anything)
	gameMock.AssertNotCalled(t, "Pass", mock.Anything)
	gameMock.AssertNotCalled(t, "ReturnTribute", mock.Anything, mock.Anything)
}

func TestGuandanInteractor_DomainErrorIsPresented(t *testing.T) {
	wantErr := errors.New("that beats nothing on the table")
	pMock := new(presenter.MockGuandanPresenter)
	pMock.On("Output", mock.Anything, wantErr).Return(guandanMockOutput)
	gameMock := new(interfaces.MockGuandanGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetCurrentPlayerIdx").Return(0)
	gameMock.On("PlayCards", 0, []int{1}).Return(wantErr)

	gi := usecase.NewGuandanInteractor(gameMock, pMock)
	assert.Equal(t, guandanMockOutput, gi.PlayCards([]int{1}))
	pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
	gameMock.AssertNotCalled(t, "CpuPlay")
}

func TestGuandanInteractor_NextHand(t *testing.T) {
	t.Run("deals the next hand", func(t *testing.T) {
		pMock := new(presenter.MockGuandanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(guandanMockOutput)
		gameMock := new(interfaces.MockGuandanGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.GuandanPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("NextHand").Return(nil)

		gi := usecase.NewGuandanInteractor(gameMock, pMock)
		assert.Equal(t, guandanMockOutput, gi.NextHand())
		gameMock.AssertCalled(t, "NextHand")
	})

	t.Run("blocked once the game is over", func(t *testing.T) {
		pMock := new(presenter.MockGuandanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(guandanMockOutput)
		gameMock := new(interfaces.MockGuandanGame)
		gameMock.On("GetGameEndFlag").Return(true)

		gi := usecase.NewGuandanInteractor(gameMock, pMock)
		assert.Equal(t, guandanMockOutput, gi.NextHand())
		gameMock.AssertNotCalled(t, "NextHand")
	})

	t.Run("a domain error is presented", func(t *testing.T) {
		wantErr := errors.New("the hand is still in play")
		pMock := new(presenter.MockGuandanPresenter)
		pMock.On("Output", mock.Anything, wantErr).Return(guandanMockOutput)
		gameMock := new(interfaces.MockGuandanGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextHand").Return(wantErr)

		gi := usecase.NewGuandanInteractor(gameMock, pMock)
		assert.Equal(t, guandanMockOutput, gi.NextHand())
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

// **CPU ループは人間の手番・局の終わり・ゲーム終了で止まる。**
func TestGuandanInteractor_RunCpuTurnsStops(t *testing.T) {
	t.Run("at the human's turn", func(t *testing.T) {
		pMock := new(presenter.MockGuandanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(guandanMockOutput)
		gameMock := new(interfaces.MockGuandanGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.GuandanPhasePlay)
		gameMock.On("IsHumanTurn").Return(false).Times(4)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("CpuPlay").Return()

		gi := usecase.NewGuandanInteractor(gameMock, pMock)
		gi.Reset()
		gameMock.AssertNumberOfCalls(t, "CpuPlay", 4)
	})

	// **局が終わったら人間が次局を始めるまで止まる。**続けると精算画面が見えない。
	t.Run("at the end of a hand", func(t *testing.T) {
		pMock := new(presenter.MockGuandanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(guandanMockOutput)
		gameMock := new(interfaces.MockGuandanGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.GuandanPhaseHandEnd)

		gi := usecase.NewGuandanInteractor(gameMock, pMock)
		gi.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("at game end", func(t *testing.T) {
		pMock := new(presenter.MockGuandanPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(guandanMockOutput)
		gameMock := new(interfaces.MockGuandanGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		gi := usecase.NewGuandanInteractor(gameMock, pMock)
		gi.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

func TestGuandanInteractor_GetConfigAndActionLog(t *testing.T) {
	cfg := domain.DefaultGuandanConfig()
	pMock := new(presenter.MockGuandanPresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return(`[]`)
	gameMock := new(interfaces.MockGuandanGame)
	gameMock.On("GetConfig").Return(cfg)

	gi := usecase.NewGuandanInteractor(gameMock, pMock)
	assert.Equal(t, cfg, gi.GetConfig())
	assert.Equal(t, `[]`, gi.ActionLog())
}

func TestGuandanInteractor_SnapshotAndRestore(t *testing.T) {
	pMock := new(presenter.MockGuandanPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(guandanMockOutput)

	g := domain.NewDefaultGuandan()
	g.Reset()
	gi := usecase.NewGuandanInteractor(g, pMock)
	data, err := gi.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreGuandanInteractor(data, pMock)
	assert.NoError(t, err)
	assert.Equal(t, g.GetConfig(), restored.GetConfig())

	_, err = usecase.RestoreGuandanInteractor([]byte(`{`), pMock)
	assert.Error(t, err)
}

// **役の下読みはここを通る** (#5734)。presenter の CheckOutput に素通しする。
func TestGuandanInteractor_Check(t *testing.T) {
	pMock := new(presenter.MockGuandanPresenter)
	pMock.On("CheckOutput", mock.Anything, mock.Anything).Return("combo")
	gameMock := new(interfaces.MockGuandanGame)

	gi := usecase.NewGuandanInteractor(gameMock, pMock)
	assert.Equal(t, "combo", gi.Check([]int{0, 1}))
	pMock.AssertCalled(t, "CheckOutput", gameMock, []int{0, 1})
}
