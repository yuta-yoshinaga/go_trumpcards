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

func TestNewPokerInteractor(t *testing.T) {
	mg := new(interfaces.MockPokerGame)
	mp := new(presenter.MockPokerPresenter)
	pi := usecase.NewPokerInteractor(mg, mp)
	assert.NotNil(t, pi)
}

func TestNewPokerInteractor_NilGame(t *testing.T) {
	mp := new(presenter.MockPokerPresenter)
	assert.PanicsWithValue(t, "PokerInteractor: p must not be nil", func() {
		usecase.NewPokerInteractor(nil, mp)
	})
}

func TestNewPokerInteractor_NilPresenter(t *testing.T) {
	mg := new(interfaces.MockPokerGame)
	assert.PanicsWithValue(t, "PokerInteractor: pp must not be nil", func() {
		usecase.NewPokerInteractor(mg, nil)
	})
}

func TestPokerInteractor_Reset(t *testing.T) {
	mg := new(interfaces.MockPokerGame)
	mp := new(presenter.MockPokerPresenter)
	pi := usecase.NewPokerInteractor(mg, mp)

	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset output")

	result := pi.Reset()
	assert.Equal(t, "reset output", result)
	mg.AssertCalled(t, "Reset")
}

func TestPokerInteractor_Reset_Error(t *testing.T) {
	mg := new(interfaces.MockPokerGame)
	mp := new(presenter.MockPokerPresenter)
	pi := usecase.NewPokerInteractor(mg, mp)

	err := errors.New("reset failed")
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := pi.Reset()
	assert.Equal(t, "error output", result)
}

func TestPokerInteractor_GetConfig(t *testing.T) {
	mg := new(interfaces.MockPokerGame)
	mp := new(presenter.MockPokerPresenter)
	pi := usecase.NewPokerInteractor(mg, mp)

	expected := domain.PokerConfig{InitChips: 1000, Ante: 10, MinBet: 10, CpuCount: 2, JokerCount: 1}
	mg.On("GetConfig").Return(expected)

	result := pi.GetConfig()
	assert.Equal(t, expected, result)
	mg.AssertCalled(t, "GetConfig")
}

func TestPokerInteractor_ResetWithConfig(t *testing.T) {
	mg := new(interfaces.MockPokerGame)
	mp := new(presenter.MockPokerPresenter)
	pi := usecase.NewPokerInteractor(mg, mp)

	cfg := domain.PokerConfig{InitChips: 2000, Ante: 20, MinBet: 20, CpuCount: 2, JokerCount: 1}
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset with config output")

	result := pi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "reset with config output", result)
	mg.AssertCalled(t, "SetConfig", cfg)
	mg.AssertCalled(t, "Reset")
}

func TestPokerInteractor_ResetWithConfig_Error(t *testing.T) {
	mg := new(interfaces.MockPokerGame)
	mp := new(presenter.MockPokerPresenter)
	pi := usecase.NewPokerInteractor(mg, mp)

	cfg := domain.PokerConfig{InitChips: 1000, Ante: 10, MinBet: 10, CpuCount: 3, JokerCount: 0}
	err := errors.New("reset failed")
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := pi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "error output", result)
}

func TestPokerInteractor_ResetWithConfig_WithProfile(t *testing.T) {
	mg := new(interfaces.MockPokerGame)
	mp := new(presenter.MockPokerPresenter)
	pi := usecase.NewPokerInteractor(mg, mp)

	cfg := domain.PokerConfig{InitChips: 2000, Ante: 20, MinBet: 20, CpuCount: 2, JokerCount: 1}
	profileData := []byte(`{"gamesPlayed":3}`)
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mg.On("ImportProfile", profileData).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("with profile output")

	result := pi.ResetWithConfig(cfg, profileData)
	assert.Equal(t, "with profile output", result)
	mg.AssertCalled(t, "ImportProfile", profileData)
}

func TestPokerInteractor_Action(t *testing.T) {
	mg := new(interfaces.MockPokerGame)
	mp := new(presenter.MockPokerPresenter)
	pi := usecase.NewPokerInteractor(mg, mp)

	mg.On("PlayerAction", domain.PokerActionCheck, 0, 0).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("action output")

	result := pi.Action(domain.PokerActionCheck, 0, 0)
	assert.Equal(t, "action output", result)
	mg.AssertCalled(t, "PlayerAction", domain.PokerActionCheck, 0, 0)
}

func TestPokerInteractor_Action_Error(t *testing.T) {
	mg := new(interfaces.MockPokerGame)
	mp := new(presenter.MockPokerPresenter)
	pi := usecase.NewPokerInteractor(mg, mp)

	err := errors.New("action failed")
	mg.On("PlayerAction", domain.PokerActionBet, 50, 0).Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := pi.Action(domain.PokerActionBet, 50, 0)
	assert.Equal(t, "error output", result)
}

func TestPokerInteractor_Exchange(t *testing.T) {
	mg := new(interfaces.MockPokerGame)
	mp := new(presenter.MockPokerPresenter)
	pi := usecase.NewPokerInteractor(mg, mp)

	indices := []int{0, 2, 4}
	mg.On("PlayerExchange", indices).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("exchange output")

	result := pi.Exchange(indices)
	assert.Equal(t, "exchange output", result)
	mg.AssertCalled(t, "PlayerExchange", indices)
}

func TestPokerInteractor_Exchange_Error(t *testing.T) {
	mg := new(interfaces.MockPokerGame)
	mp := new(presenter.MockPokerPresenter)
	pi := usecase.NewPokerInteractor(mg, mp)

	indices := []int{1}
	err := errors.New("exchange failed")
	mg.On("PlayerExchange", indices).Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := pi.Exchange(indices)
	assert.Equal(t, "error output", result)
}

func TestPokerInteractor_Stand(t *testing.T) {
	mg := new(interfaces.MockPokerGame)
	mp := new(presenter.MockPokerPresenter)
	pi := usecase.NewPokerInteractor(mg, mp)

	mg.On("PlayerStand").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("stand output")

	result := pi.Stand()
	assert.Equal(t, "stand output", result)
	mg.AssertCalled(t, "PlayerStand")
}

func TestPokerInteractor_Stand_Error(t *testing.T) {
	mg := new(interfaces.MockPokerGame)
	mp := new(presenter.MockPokerPresenter)
	pi := usecase.NewPokerInteractor(mg, mp)

	err := errors.New("stand failed")
	mg.On("PlayerStand").Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := pi.Stand()
	assert.Equal(t, "error output", result)
}

func TestPokerInteractor_ActionLog(t *testing.T) {
	mg := new(interfaces.MockPokerGame)
	mp := new(presenter.MockPokerPresenter)
	mp.On("ActionLogOutput", mg).Return(`{"entries":[]}`)

	pi := usecase.NewPokerInteractor(mg, mp)
	result := pi.ActionLog()
	assert.Equal(t, `{"entries":[]}`, result)
	mp.AssertExpectations(t)
}

func TestPokerInteractor_Odds(t *testing.T) {
	mg := new(interfaces.MockPokerGame)
	mp := new(presenter.MockPokerPresenter)
	pi := usecase.NewPokerInteractor(mg, mp)

	odds := []domain.PokerDrawOdds{
		{HandRank: 0, HandName: "High Card", Probability: 0.5, Count: 1, Total: 2},
	}
	indices := []int{0, 1}
	mg.On("CalcDrawOdds", indices).Return(odds, nil)
	mp.On("OutputWithOdds", mg, mock.Anything, odds).Return("odds output")

	result := pi.Odds(indices)
	assert.Equal(t, "odds output", result)
	mg.AssertCalled(t, "CalcDrawOdds", indices)
}

func TestPokerInteractor_Odds_Error(t *testing.T) {
	mg := new(interfaces.MockPokerGame)
	mp := new(presenter.MockPokerPresenter)
	pi := usecase.NewPokerInteractor(mg, mp)

	err := errors.New("odds failed")
	indices := []int{0}
	mg.On("CalcDrawOdds", indices).Return([]domain.PokerDrawOdds(nil), err)
	mp.On("Output", mg, err).Return("error output")

	result := pi.Odds(indices)
	assert.Equal(t, "error output", result)
}

func TestPokerInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	ppMock := new(presenter.MockPokerPresenter)
	gameMock := new(interfaces.MockPokerGame)
	ppMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

	pi := usecase.NewPokerInteractor(gameMock, ppMock)
	cfg := domain.DefaultPokerConfig()
	cfg.CpuCount = 0
	result := pi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "validation error", result)
	gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
}
