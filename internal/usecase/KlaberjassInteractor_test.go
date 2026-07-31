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

const klaberjassMockOutput = `{"phase":0}`

// kjPlayable stubs a game that is sitting on the human's turn.
func kjPlayable() (*interfaces.MockKlaberjassGame, *presenter.MockKlaberjassPresenter) {
	pMock := new(presenter.MockKlaberjassPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(klaberjassMockOutput)
	gameMock := new(interfaces.MockKlaberjassGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetPhase").Return(domain.KlaberjassPhaseHandEnd)
	return gameMock, pMock
}

func TestNewKlaberjassInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockKlaberjassPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "KlaberjassInteractor: g must not be nil", func() {
			usecase.NewKlaberjassInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockKlaberjassGame)
		assert.PanicsWithValue(t, "KlaberjassInteractor: gp must not be nil", func() {
			usecase.NewKlaberjassInteractor(gameMock, nil)
		})
	})
}

func TestKlaberjassInteractor_Reset(t *testing.T) {
	pMock := new(presenter.MockKlaberjassPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(klaberjassMockOutput)
	gameMock := new(interfaces.MockKlaberjassGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KlaberjassPhaseBidTurnUp)
	gameMock.On("IsHumanTurn").Return(true)

	ki := usecase.NewKlaberjassInteractor(gameMock, pMock)
	assert.Equal(t, klaberjassMockOutput, ki.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestKlaberjassInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockKlaberjassPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(klaberjassMockOutput)
		gameMock := new(interfaces.MockKlaberjassGame)
		cfg := domain.KlaberjassConfig{TargetScore: 300, AllowSchmeiss: false}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.KlaberjassPhaseBidTurnUp)
		gameMock.On("IsHumanTurn").Return(true)

		ki := usecase.NewKlaberjassInteractor(gameMock, pMock)
		assert.Equal(t, klaberjassMockOutput, ki.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config never reaches the game", func(t *testing.T) {
		pMock := new(presenter.MockKlaberjassPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(klaberjassMockOutput)
		gameMock := new(interfaces.MockKlaberjassGame)

		ki := usecase.NewKlaberjassInteractor(gameMock, pMock)
		assert.Equal(t, klaberjassMockOutput, ki.ResetWithConfig(domain.KlaberjassConfig{TargetScore: 0}))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
		gameMock.AssertNotCalled(t, "Reset")
	})
}

// **ビッド系は bidIdx、プレイは currentIdx。**取り違えると常に相手の席で
// 動こうとして弾かれる。
func TestKlaberjassInteractor_UsesTheRightSeatPerPhase(t *testing.T) {
	t.Run("bidding uses the bid seat", func(t *testing.T) {
		gameMock, pMock := kjPlayable()
		gameMock.On("GetBidPlayerIdx").Return(1)
		gameMock.On("AcceptTrump", 1).Return(nil)
		gameMock.On("CallTrump", 1, domain.CardDesignHeart).Return(nil)
		gameMock.On("Pass", 1).Return(nil)
		gameMock.On("Schmeiss", 1).Return(nil)
		gameMock.On("AnswerSchmeiss", 1, true).Return(nil)

		ki := usecase.NewKlaberjassInteractor(gameMock, pMock)
		ki.AcceptTrump()
		ki.CallTrump(domain.CardDesignHeart)
		ki.Pass()
		ki.Schmeiss()
		ki.AnswerSchmeiss(true)

		gameMock.AssertCalled(t, "AcceptTrump", 1)
		gameMock.AssertCalled(t, "CallTrump", 1, domain.CardDesignHeart)
		gameMock.AssertCalled(t, "Pass", 1)
		gameMock.AssertCalled(t, "Schmeiss", 1)
		gameMock.AssertCalled(t, "AnswerSchmeiss", 1, true)
	})

	t.Run("playing uses the current seat", func(t *testing.T) {
		gameMock, pMock := kjPlayable()
		gameMock.On("GetCurrentPlayerIdx").Return(0)
		gameMock.On("PlayCard", 0, 3).Return(nil)

		ki := usecase.NewKlaberjassInteractor(gameMock, pMock)
		assert.Equal(t, klaberjassMockOutput, ki.PlayCard(3))
		gameMock.AssertCalled(t, "PlayCard", 0, 3)
	})
}

func TestKlaberjassInteractor_BlockedWhenNotTheHumansTurn(t *testing.T) {
	pMock := new(presenter.MockKlaberjassPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(klaberjassMockOutput)
	gameMock := new(interfaces.MockKlaberjassGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	ki := usecase.NewKlaberjassInteractor(gameMock, pMock)
	assert.Equal(t, klaberjassMockOutput, ki.PlayCard(0))
	assert.Equal(t, klaberjassMockOutput, ki.AcceptTrump())
	gameMock.AssertNotCalled(t, "PlayCard", mock.Anything, mock.Anything)
	gameMock.AssertNotCalled(t, "AcceptTrump", mock.Anything)
}

func TestKlaberjassInteractor_DomainErrorIsPresented(t *testing.T) {
	wantErr := errors.New("that card may not be played")
	pMock := new(presenter.MockKlaberjassPresenter)
	pMock.On("Output", mock.Anything, wantErr).Return(klaberjassMockOutput)
	gameMock := new(interfaces.MockKlaberjassGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetCurrentPlayerIdx").Return(0)
	gameMock.On("PlayCard", 0, 1).Return(wantErr)

	ki := usecase.NewKlaberjassInteractor(gameMock, pMock)
	assert.Equal(t, klaberjassMockOutput, ki.PlayCard(1))
	pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
	// エラーで止まったら CPU は動かさない。
	gameMock.AssertNotCalled(t, "CpuPlay")
}

func TestKlaberjassInteractor_NextDeal(t *testing.T) {
	t.Run("deals again", func(t *testing.T) {
		pMock := new(presenter.MockKlaberjassPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(klaberjassMockOutput)
		gameMock := new(interfaces.MockKlaberjassGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextDeal").Return(nil)
		gameMock.On("GetPhase").Return(domain.KlaberjassPhaseBidTurnUp)
		gameMock.On("IsHumanTurn").Return(true)

		ki := usecase.NewKlaberjassInteractor(gameMock, pMock)
		assert.Equal(t, klaberjassMockOutput, ki.NextDeal())
		gameMock.AssertCalled(t, "NextDeal")
	})

	t.Run("blocked after the game ends", func(t *testing.T) {
		pMock := new(presenter.MockKlaberjassPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(klaberjassMockOutput)
		gameMock := new(interfaces.MockKlaberjassGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ki := usecase.NewKlaberjassInteractor(gameMock, pMock)
		assert.Equal(t, klaberjassMockOutput, ki.NextDeal())
		gameMock.AssertNotCalled(t, "NextDeal")
	})

	t.Run("an error is presented", func(t *testing.T) {
		wantErr := errors.New("the deal is still in progress")
		pMock := new(presenter.MockKlaberjassPresenter)
		pMock.On("Output", mock.Anything, wantErr).Return(klaberjassMockOutput)
		gameMock := new(interfaces.MockKlaberjassGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextDeal").Return(wantErr)

		ki := usecase.NewKlaberjassInteractor(gameMock, pMock)
		assert.Equal(t, klaberjassMockOutput, ki.NextDeal())
		pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
	})
}

// **CPU ループは人間の手番と精算で止まる。**止まらないと操作できない。
func TestKlaberjassInteractor_RunCpuTurnsStops(t *testing.T) {
	t.Run("at the human's turn", func(t *testing.T) {
		pMock := new(presenter.MockKlaberjassPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(klaberjassMockOutput)
		gameMock := new(interfaces.MockKlaberjassGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.KlaberjassPhasePlay)
		gameMock.On("IsHumanTurn").Return(false).Times(3)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("CpuPlay").Return()

		ki := usecase.NewKlaberjassInteractor(gameMock, pMock)
		ki.Reset()
		gameMock.AssertNumberOfCalls(t, "CpuPlay", 3)
	})

	t.Run("at the settlement", func(t *testing.T) {
		pMock := new(presenter.MockKlaberjassPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(klaberjassMockOutput)
		gameMock := new(interfaces.MockKlaberjassGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.KlaberjassPhaseHandEnd)

		ki := usecase.NewKlaberjassInteractor(gameMock, pMock)
		ki.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

func TestKlaberjassInteractor_GetConfigAndActionLog(t *testing.T) {
	cfg := domain.DefaultKlaberjassConfig()
	pMock := new(presenter.MockKlaberjassPresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return(`[]`)
	gameMock := new(interfaces.MockKlaberjassGame)
	gameMock.On("GetConfig").Return(cfg)

	ki := usecase.NewKlaberjassInteractor(gameMock, pMock)
	assert.Equal(t, cfg, ki.GetConfig())
	assert.Equal(t, `[]`, ki.ActionLog())
}

func TestKlaberjassInteractor_SnapshotAndRestore(t *testing.T) {
	pMock := new(presenter.MockKlaberjassPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(klaberjassMockOutput)

	g := domain.NewDefaultKlaberjass()
	g.Reset()
	ki := usecase.NewKlaberjassInteractor(g, pMock)
	data, err := ki.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreKlaberjassInteractor(data, pMock)
	assert.NoError(t, err)
	assert.Equal(t, g.GetConfig(), restored.GetConfig())

	_, err = usecase.RestoreKlaberjassInteractor([]byte(`{`), pMock)
	assert.Error(t, err)
}
