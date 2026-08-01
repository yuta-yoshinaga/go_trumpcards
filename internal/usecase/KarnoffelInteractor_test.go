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

const karnoffelMockOutput = `{"phase":0}`

func TestNewKarnoffelInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockKarnoffelPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "KarnoffelInteractor: g must not be nil", func() {
			usecase.NewKarnoffelInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockKarnoffelGame)
		assert.PanicsWithValue(t, "KarnoffelInteractor: gp must not be nil", func() {
			usecase.NewKarnoffelInteractor(gameMock, nil)
		})
	})
}

func TestKarnoffelInteractor_Reset(t *testing.T) {
	pMock := new(presenter.MockKarnoffelPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(karnoffelMockOutput)
	gameMock := new(interfaces.MockKarnoffelGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KarnoffelPhasePlay)
	gameMock.On("IsHumanTurn").Return(true)

	ki := usecase.NewKarnoffelInteractor(gameMock, pMock)
	assert.Equal(t, karnoffelMockOutput, ki.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestKarnoffelInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockKarnoffelPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(karnoffelMockOutput)
		gameMock := new(interfaces.MockKarnoffelGame)
		cfg := domain.DefaultKarnoffelConfig()
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.KarnoffelPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ki := usecase.NewKarnoffelInteractor(gameMock, pMock)
		assert.Equal(t, karnoffelMockOutput, ki.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config never reaches the game", func(t *testing.T) {
		pMock := new(presenter.MockKarnoffelPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(karnoffelMockOutput)
		gameMock := new(interfaces.MockKarnoffelGame)

		ki := usecase.NewKarnoffelInteractor(gameMock, pMock)
		assert.Equal(t, karnoffelMockOutput, ki.ResetWithConfig(domain.KarnoffelConfig{TargetHands: 99}))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
		gameMock.AssertNotCalled(t, "Reset")
	})
}

func TestKarnoffelInteractor_PlayCard(t *testing.T) {
	t.Run("uses the current seat", func(t *testing.T) {
		pMock := new(presenter.MockKarnoffelPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(karnoffelMockOutput)
		gameMock := new(interfaces.MockKarnoffelGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("GetPhase").Return(domain.KarnoffelPhaseHandEnd)
		gameMock.On("GetCurrentPlayerIdx").Return(2)
		gameMock.On("PlayCard", 2, 3).Return(nil)

		ki := usecase.NewKarnoffelInteractor(gameMock, pMock)
		assert.Equal(t, karnoffelMockOutput, ki.PlayCard(3))
		gameMock.AssertCalled(t, "PlayCard", 2, 3)
	})

	t.Run("blocked when it is not the human's turn", func(t *testing.T) {
		pMock := new(presenter.MockKarnoffelPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(karnoffelMockOutput)
		gameMock := new(interfaces.MockKarnoffelGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ki := usecase.NewKarnoffelInteractor(gameMock, pMock)
		assert.Equal(t, karnoffelMockOutput, ki.PlayCard(0))
		gameMock.AssertNotCalled(t, "PlayCard", mock.Anything, mock.Anything)
	})

	// **第 1 トリックのリードに悪魔は使えない。**その拒否が伝わること。
	t.Run("a domain error is presented", func(t *testing.T) {
		wantErr := errors.New("the devil cannot lead the first trick")
		pMock := new(presenter.MockKarnoffelPresenter)
		pMock.On("Output", mock.Anything, wantErr).Return(karnoffelMockOutput)
		gameMock := new(interfaces.MockKarnoffelGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("GetCurrentPlayerIdx").Return(0)
		gameMock.On("PlayCard", 0, 0).Return(wantErr)

		ki := usecase.NewKarnoffelInteractor(gameMock, pMock)
		assert.Equal(t, karnoffelMockOutput, ki.PlayCard(0))
		pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

func TestKarnoffelInteractor_NextHand(t *testing.T) {
	t.Run("deals again", func(t *testing.T) {
		pMock := new(presenter.MockKarnoffelPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(karnoffelMockOutput)
		gameMock := new(interfaces.MockKarnoffelGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextHand").Return(nil)
		gameMock.On("GetPhase").Return(domain.KarnoffelPhasePlay)
		gameMock.On("IsHumanTurn").Return(true)

		ki := usecase.NewKarnoffelInteractor(gameMock, pMock)
		assert.Equal(t, karnoffelMockOutput, ki.NextHand())
		gameMock.AssertCalled(t, "NextHand")
	})

	t.Run("blocked after the game", func(t *testing.T) {
		pMock := new(presenter.MockKarnoffelPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(karnoffelMockOutput)
		gameMock := new(interfaces.MockKarnoffelGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ki := usecase.NewKarnoffelInteractor(gameMock, pMock)
		assert.Equal(t, karnoffelMockOutput, ki.NextHand())
		gameMock.AssertNotCalled(t, "NextHand")
	})

	t.Run("an error is presented", func(t *testing.T) {
		wantErr := errors.New("the hand is still in progress")
		pMock := new(presenter.MockKarnoffelPresenter)
		pMock.On("Output", mock.Anything, wantErr).Return(karnoffelMockOutput)
		gameMock := new(interfaces.MockKarnoffelGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextHand").Return(wantErr)

		ki := usecase.NewKarnoffelInteractor(gameMock, pMock)
		assert.Equal(t, karnoffelMockOutput, ki.NextHand())
		pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
	})
}

// **CPU ループは人間の手番と局終了で止まる。**
func TestKarnoffelInteractor_RunCpuTurnsStops(t *testing.T) {
	t.Run("at the human's turn", func(t *testing.T) {
		pMock := new(presenter.MockKarnoffelPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(karnoffelMockOutput)
		gameMock := new(interfaces.MockKarnoffelGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.KarnoffelPhasePlay)
		gameMock.On("IsHumanTurn").Return(false).Times(3)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("CpuPlay").Return()

		ki := usecase.NewKarnoffelInteractor(gameMock, pMock)
		ki.Reset()
		gameMock.AssertNumberOfCalls(t, "CpuPlay", 3)
	})

	t.Run("at the settlement", func(t *testing.T) {
		pMock := new(presenter.MockKarnoffelPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(karnoffelMockOutput)
		gameMock := new(interfaces.MockKarnoffelGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.KarnoffelPhaseHandEnd)

		ki := usecase.NewKarnoffelInteractor(gameMock, pMock)
		ki.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

func TestKarnoffelInteractor_GetConfigAndActionLog(t *testing.T) {
	cfg := domain.DefaultKarnoffelConfig()
	pMock := new(presenter.MockKarnoffelPresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return(`[]`)
	gameMock := new(interfaces.MockKarnoffelGame)
	gameMock.On("GetConfig").Return(cfg)

	ki := usecase.NewKarnoffelInteractor(gameMock, pMock)
	assert.Equal(t, cfg, ki.GetConfig())
	assert.Equal(t, `[]`, ki.ActionLog())
}

func TestKarnoffelInteractor_SnapshotAndRestore(t *testing.T) {
	pMock := new(presenter.MockKarnoffelPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(karnoffelMockOutput)

	g := domain.NewDefaultKarnoffel()
	g.Reset()
	ki := usecase.NewKarnoffelInteractor(g, pMock)
	data, err := ki.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreKarnoffelInteractor(data, pMock)
	assert.NoError(t, err)
	assert.Equal(t, g.GetConfig(), restored.GetConfig())

	_, err = usecase.RestoreKarnoffelInteractor([]byte(`{`), pMock)
	assert.Error(t, err)
}
