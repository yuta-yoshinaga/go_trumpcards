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

const shengJiMockOutput = `{"phase":0}`

func TestNewShengJiInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockShengJiPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "ShengJiInteractor: g must not be nil", func() {
			usecase.NewShengJiInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockShengJiGame)
		assert.PanicsWithValue(t, "ShengJiInteractor: gp must not be nil", func() {
			usecase.NewShengJiInteractor(gameMock, nil)
		})
	})
}

func TestShengJiInteractor_Reset(t *testing.T) {
	pMock := new(presenter.MockShengJiPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(shengJiMockOutput)
	gameMock := new(interfaces.MockShengJiGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.ShengJiPhaseDeclare)
	gameMock.On("IsHumanTurn").Return(true)

	si := usecase.NewShengJiInteractor(gameMock, pMock)
	assert.Equal(t, shengJiMockOutput, si.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestShengJiInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockShengJiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(shengJiMockOutput)
		gameMock := new(interfaces.MockShengJiGame)
		cfg := domain.DefaultShengJiConfig()
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.ShengJiPhaseDeclare)
		gameMock.On("IsHumanTurn").Return(true)

		si := usecase.NewShengJiInteractor(gameMock, pMock)
		assert.Equal(t, shengJiMockOutput, si.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config never reaches the game", func(t *testing.T) {
		pMock := new(presenter.MockShengJiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(shengJiMockOutput)
		gameMock := new(interfaces.MockShengJiGame)

		si := usecase.NewShengJiInteractor(gameMock, pMock)
		assert.Equal(t, shengJiMockOutput, si.ResetWithConfig(domain.ShengJiConfig{CpuDifficulty: 9}))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
		gameMock.AssertNotCalled(t, "Reset")
	})
}

// **どのアクションも currentIdx の席として実行される。**
func TestShengJiInteractor_UsesTheCurrentSeat(t *testing.T) {
	newMocks := func() (*presenter.MockShengJiPresenter, *interfaces.MockShengJiGame) {
		pMock := new(presenter.MockShengJiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(shengJiMockOutput)
		gameMock := new(interfaces.MockShengJiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.ShengJiPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("GetCurrentPlayerIdx").Return(2)
		return pMock, gameMock
	}

	// **0 はパス。**省略ではなく、意味のある宣言。
	t.Run("declare passes through, including the pass", func(t *testing.T) {
		pMock, gameMock := newMocks()
		gameMock.On("Declare", 2, domain.CardDesignHeart).Return(nil)
		gameMock.On("Declare", 2, domain.ShengJiNoTrump).Return(nil)

		si := usecase.NewShengJiInteractor(gameMock, pMock)
		assert.Equal(t, shengJiMockOutput, si.Declare(domain.CardDesignHeart))
		gameMock.AssertCalled(t, "Declare", 2, domain.CardDesignHeart)
		assert.Equal(t, shengJiMockOutput, si.Declare(domain.ShengJiNoTrump))
		gameMock.AssertCalled(t, "Declare", 2, domain.ShengJiNoTrump)
	})

	t.Run("bury", func(t *testing.T) {
		pMock, gameMock := newMocks()
		idxs := []int{0, 1, 2, 3, 4, 5, 6, 7}
		gameMock.On("BuryKitty", 2, idxs).Return(nil)

		si := usecase.NewShengJiInteractor(gameMock, pMock)
		assert.Equal(t, shengJiMockOutput, si.BuryKitty(idxs))
		gameMock.AssertCalled(t, "BuryKitty", 2, idxs)
	})

	t.Run("play", func(t *testing.T) {
		pMock, gameMock := newMocks()
		gameMock.On("Play", 2, []int{0, 1}).Return(nil)

		si := usecase.NewShengJiInteractor(gameMock, pMock)
		assert.Equal(t, shengJiMockOutput, si.Play([]int{0, 1}))
		gameMock.AssertCalled(t, "Play", 2, []int{0, 1})
	})
}

func TestShengJiInteractor_BlockedWhenNotTheHumansTurn(t *testing.T) {
	pMock := new(presenter.MockShengJiPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(shengJiMockOutput)
	gameMock := new(interfaces.MockShengJiGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	si := usecase.NewShengJiInteractor(gameMock, pMock)
	assert.Equal(t, shengJiMockOutput, si.Declare(domain.CardDesignHeart))
	assert.Equal(t, shengJiMockOutput, si.BuryKitty([]int{0}))
	assert.Equal(t, shengJiMockOutput, si.Play([]int{0}))
	gameMock.AssertNotCalled(t, "Declare", mock.Anything, mock.Anything)
	gameMock.AssertNotCalled(t, "BuryKitty", mock.Anything, mock.Anything)
	gameMock.AssertNotCalled(t, "Play", mock.Anything, mock.Anything)
}

func TestShengJiInteractor_DomainErrorIsPresented(t *testing.T) {
	wantErr := errors.New("you must follow the led suit while you hold it")
	pMock := new(presenter.MockShengJiPresenter)
	pMock.On("Output", mock.Anything, wantErr).Return(shengJiMockOutput)
	gameMock := new(interfaces.MockShengJiGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetCurrentPlayerIdx").Return(0)
	gameMock.On("Play", 0, []int{1}).Return(wantErr)

	si := usecase.NewShengJiInteractor(gameMock, pMock)
	assert.Equal(t, shengJiMockOutput, si.Play([]int{1}))
	pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
	gameMock.AssertNotCalled(t, "CpuPlay")
}

func TestShengJiInteractor_NextHand(t *testing.T) {
	t.Run("deals the next hand", func(t *testing.T) {
		pMock := new(presenter.MockShengJiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(shengJiMockOutput)
		gameMock := new(interfaces.MockShengJiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.ShengJiPhaseDeclare)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("NextHand").Return(nil)

		si := usecase.NewShengJiInteractor(gameMock, pMock)
		assert.Equal(t, shengJiMockOutput, si.NextHand())
		gameMock.AssertCalled(t, "NextHand")
	})

	t.Run("blocked once the game is over", func(t *testing.T) {
		pMock := new(presenter.MockShengJiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(shengJiMockOutput)
		gameMock := new(interfaces.MockShengJiGame)
		gameMock.On("GetGameEndFlag").Return(true)

		si := usecase.NewShengJiInteractor(gameMock, pMock)
		assert.Equal(t, shengJiMockOutput, si.NextHand())
		gameMock.AssertNotCalled(t, "NextHand")
	})

	t.Run("a domain error is presented", func(t *testing.T) {
		wantErr := errors.New("the hand is still in play")
		pMock := new(presenter.MockShengJiPresenter)
		pMock.On("Output", mock.Anything, wantErr).Return(shengJiMockOutput)
		gameMock := new(interfaces.MockShengJiGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextHand").Return(wantErr)

		si := usecase.NewShengJiInteractor(gameMock, pMock)
		assert.Equal(t, shengJiMockOutput, si.NextHand())
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

// **CPU ループは人間の手番・局の終わり・ゲーム終了で止まる。**
func TestShengJiInteractor_RunCpuTurnsStops(t *testing.T) {
	t.Run("at the human's turn", func(t *testing.T) {
		pMock := new(presenter.MockShengJiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(shengJiMockOutput)
		gameMock := new(interfaces.MockShengJiGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.ShengJiPhasePlay)
		gameMock.On("IsHumanTurn").Return(false).Times(4)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("CpuPlay").Return()

		si := usecase.NewShengJiInteractor(gameMock, pMock)
		si.Reset()
		gameMock.AssertNumberOfCalls(t, "CpuPlay", 4)
	})

	// **局が終わったら人間が次局を始めるまで止まる。**続けると精算画面が見えない。
	t.Run("at the end of a hand", func(t *testing.T) {
		pMock := new(presenter.MockShengJiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(shengJiMockOutput)
		gameMock := new(interfaces.MockShengJiGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.ShengJiPhaseHandEnd)

		si := usecase.NewShengJiInteractor(gameMock, pMock)
		si.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("at game end", func(t *testing.T) {
		pMock := new(presenter.MockShengJiPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(shengJiMockOutput)
		gameMock := new(interfaces.MockShengJiGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		si := usecase.NewShengJiInteractor(gameMock, pMock)
		si.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

func TestShengJiInteractor_GetConfigAndActionLog(t *testing.T) {
	cfg := domain.DefaultShengJiConfig()
	pMock := new(presenter.MockShengJiPresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return(`[]`)
	gameMock := new(interfaces.MockShengJiGame)
	gameMock.On("GetConfig").Return(cfg)

	si := usecase.NewShengJiInteractor(gameMock, pMock)
	assert.Equal(t, cfg, si.GetConfig())
	assert.Equal(t, `[]`, si.ActionLog())
}

func TestShengJiInteractor_SnapshotAndRestore(t *testing.T) {
	pMock := new(presenter.MockShengJiPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(shengJiMockOutput)

	g := domain.NewDefaultShengJi()
	g.Reset()
	si := usecase.NewShengJiInteractor(g, pMock)
	data, err := si.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreShengJiInteractor(data, pMock)
	assert.NoError(t, err)
	assert.Equal(t, g.GetConfig(), restored.GetConfig())

	_, err = usecase.RestoreShengJiInteractor([]byte(`{`), pMock)
	assert.Error(t, err)
}
