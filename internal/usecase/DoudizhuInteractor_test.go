package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func newTestDoudizhu() *domain.Doudizhu {
	config := domain.DefaultDoudizhuConfig()
	players := []*domain.DoudizhuPlayer{
		domain.NewDoudizhuPlayer(true),
		domain.NewDoudizhuPlayer(false),
		domain.NewDoudizhuPlayer(false),
	}
	return domain.NewDoudizhu(domain.NewTrumpCards(domain.DoudizhuJokerCount), players, config)
}

func TestNewDoudizhuInteractor_NilGuards(t *testing.T) {
	dgpMock := new(presenter.MockDoudizhuPresenter)
	t.Run("panics when dg is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "DoudizhuInteractor: dg must not be nil", func() {
			usecase.NewDoudizhuInteractor(nil, dgpMock)
		})
	})
	t.Run("panics when dgp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "DoudizhuInteractor: dgp must not be nil", func() {
			usecase.NewDoudizhuInteractor(newTestDoudizhu(), nil)
		})
	})
}

func TestDoudizhuInteractor_RealGame(t *testing.T) {
	mockOutput := `{"players":[]}`
	dgpMock := new(presenter.MockDoudizhuPresenter)
	dgpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	tdi := usecase.NewDoudizhuInteractor(newTestDoudizhu(), dgpMock)

	t.Run("success Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, tdi.Reset())
	})
	t.Run("success Bid", func(t *testing.T) {
		assert.Equal(t, mockOutput, tdi.Bid(0))
	})
	t.Run("success Play with pass", func(t *testing.T) {
		assert.Equal(t, mockOutput, tdi.Play([]int{}))
	})
	t.Run("success ResetWithConfig", func(t *testing.T) {
		assert.Equal(t, mockOutput, tdi.ResetWithConfig(domain.DefaultDoudizhuConfig()))
	})
}

func TestDoudizhuInteractor_ActionLog(t *testing.T) {
	dgpMock := new(presenter.MockDoudizhuPresenter)
	gameMock := new(interfaces.MockDoudizhuGame)
	dgpMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)

	di := usecase.NewDoudizhuInteractor(gameMock, dgpMock)
	assert.Equal(t, `{"entries":[]}`, di.ActionLog())
	dgpMock.AssertExpectations(t)
}

func TestDoudizhuInteractor_MockGame(t *testing.T) {
	mockOutput := `{"players":[]}`
	dgpMock := new(presenter.MockDoudizhuPresenter)
	dgpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockDoudizhuGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("HasPendingAction").Return(false)
	gameMock.On("CpuPlay").Return()
	gameMock.On("PlayerPlay", mock.Anything).Return(nil)
	gameMock.On("PlayerBid", mock.Anything).Return(nil)
	gameMock.On("SetConfig", mock.Anything).Return()

	di := usecase.NewDoudizhuInteractor(gameMock, dgpMock)

	t.Run("Reset calls game.Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, di.Reset())
		gameMock.AssertCalled(t, "Reset")
	})
	t.Run("Bid calls game.PlayerBid", func(t *testing.T) {
		assert.Equal(t, mockOutput, di.Bid(2))
		gameMock.AssertCalled(t, "PlayerBid", 2)
	})
	t.Run("Play calls game.PlayerPlay when human turn", func(t *testing.T) {
		assert.Equal(t, mockOutput, di.Play([]int{0}))
		gameMock.AssertCalled(t, "PlayerPlay", []int{0})
	})
	t.Run("GetConfig delegates", func(t *testing.T) {
		gameMock.On("GetConfig").Return(domain.DefaultDoudizhuConfig())
		assert.Equal(t, domain.DefaultDoudizhuConfig(), di.GetConfig())
	})
	t.Run("ResetWithConfig calls SetConfig then Reset", func(t *testing.T) {
		cfg := domain.DefaultDoudizhuConfig()
		assert.Equal(t, mockOutput, di.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})
}

func TestDoudizhuInteractor_Play_NotHumanTurn(t *testing.T) {
	mockOutput := `{"players":[]}`
	dgpMock := new(presenter.MockDoudizhuPresenter)
	dgpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockDoudizhuGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	di := usecase.NewDoudizhuInteractor(gameMock, dgpMock)
	assert.Equal(t, mockOutput, di.Play([]int{0}))
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything)
}

func TestDoudizhuInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	dgpMock := new(presenter.MockDoudizhuPresenter)
	gameMock := new(interfaces.MockDoudizhuGame)
	dgpMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

	di := usecase.NewDoudizhuInteractor(gameMock, dgpMock)
	cfg := domain.DefaultDoudizhuConfig()
	cfg.CpuDifficulty = domain.DoudizhuCpuDifficulty(-1)
	assert.Equal(t, "validation error", di.ResetWithConfig(cfg))
	gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
}
