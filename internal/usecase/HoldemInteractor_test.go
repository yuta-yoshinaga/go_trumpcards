package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestNewHoldemInteractor(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)
	assert.NotNil(t, hi)
}

func TestNewHoldemInteractor_NilGame(t *testing.T) {
	mp := new(presenter.MockHoldemPresenter)
	assert.Panics(t, func() {
		NewHoldemInteractor(nil, mp)
	})
}

func TestNewHoldemInteractor_NilPresenter(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	assert.Panics(t, func() {
		NewHoldemInteractor(mg, nil)
	})
}

func TestHoldemInteractor_Reset(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset output")

	result := hi.Reset()
	assert.Equal(t, "reset output", result)
	mg.AssertCalled(t, "Reset")
}

func TestHoldemInteractor_Reset_Error(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	err := errors.New("reset failed")
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := hi.Reset()
	assert.Equal(t, "error output", result)
}

func TestHoldemInteractor_ResetWithConfig_Error(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	cfg := domain.HoldemConfig{SmallBlind: 10, BigBlind: 20, InitChips: 2000, BlindLevelHands: 10}
	err := errors.New("reset failed")
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := hi.ResetWithConfig(cfg)
	assert.Equal(t, "error output", result)
}

func TestHoldemInteractor_ResetWithConfig(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	cfg := domain.HoldemConfig{SmallBlind: 10, BigBlind: 20, InitChips: 2000, BlindLevelHands: 10}
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset with config output")

	result := hi.ResetWithConfig(cfg)
	assert.Equal(t, "reset with config output", result)
	mg.AssertCalled(t, "SetConfig", cfg)
	mg.AssertCalled(t, "Reset")
}

func TestHoldemInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	mp.On("Output", mg, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")
	cfg := domain.HoldemConfig{SmallBlind: 0, BigBlind: 10, BlindLevelHands: 10}
	result := hi.ResetWithConfig(cfg)
	assert.Equal(t, "validation error", result)
	mg.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestHoldemInteractor_Action(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	mg.On("PlayerAction", domain.HoldemActionCheck, 0, 0).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("action output")

	result := hi.Action(domain.HoldemActionCheck, 0, 0)
	assert.Equal(t, "action output", result)
	mg.AssertCalled(t, "PlayerAction", domain.HoldemActionCheck, 0, 0)
}

func TestHoldemInteractor_GetConfig(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	cfg := domain.DefaultHoldemConfig()
	mg.On("GetConfig").Return(cfg)

	result := hi.GetConfig()
	assert.Equal(t, cfg, result)
	mg.AssertCalled(t, "GetConfig")
}

func TestHoldemInteractor_ResetWithConfig_TableSizeChange(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	cfg := domain.DefaultHoldemConfig()
	cfg.TableSize = domain.HoldemTableSize6
	mg.On("GetPlayerCnt").Return(4)
	mg.On("Resize", mock.MatchedBy(func(players []*domain.HoldemPlayer) bool {
		return len(players) == 6 && players[0].GetIsHuman()
	})).Return()
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("resize output")

	result := hi.ResetWithConfig(cfg)
	assert.Equal(t, "resize output", result)
	mg.AssertCalled(t, "Resize", mock.Anything)
}

func TestHoldemInteractor_ResetWithConfig_SameTableSize(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	cfg := domain.DefaultHoldemConfig()
	cfg.TableSize = domain.HoldemTableSize4
	mg.On("GetPlayerCnt").Return(4)
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("no resize output")

	result := hi.ResetWithConfig(cfg)
	assert.Equal(t, "no resize output", result)
	mg.AssertNotCalled(t, "Resize", mock.Anything)
}

func TestHoldemInteractor_ResetWithConfig_TableSizeZero(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	cfg := domain.DefaultHoldemConfig()
	cfg.TableSize = 0 // not set, should skip resize
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("zero output")

	result := hi.ResetWithConfig(cfg)
	assert.Equal(t, "zero output", result)
	mg.AssertNotCalled(t, "Resize", mock.Anything)
}

func TestHoldemInteractor_ActionLog(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	mp.On("ActionLogOutput", mg).Return(`{"entries":[]}`)

	hi := NewHoldemInteractor(mg, mp)
	result := hi.ActionLog()
	assert.Equal(t, `{"entries":[]}`, result)
	mp.AssertExpectations(t)
}

func TestHoldemInteractor_Action_Error(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	err := errors.New("test error")
	mg.On("PlayerAction", domain.HoldemActionBet, 50, 0).Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := hi.Action(domain.HoldemActionBet, 50, 0)
	assert.Equal(t, "error output", result)
}

func TestHoldemInteractor_Rebuy(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	mg.On("Rebuy").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("rebuy output")

	result := hi.Rebuy()
	assert.Equal(t, "rebuy output", result)
	mg.AssertCalled(t, "Rebuy")
}

func TestHoldemInteractor_Rebuy_Error(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	err := errors.New("rebuy failed")
	mg.On("Rebuy").Return(err)
	mp.On("Output", mg, err).Return("rebuy error output")

	result := hi.Rebuy()
	assert.Equal(t, "rebuy error output", result)
}

func TestHoldemInteractor_SkipRebuy(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	mg.On("SkipRebuy").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("skip rebuy output")

	result := hi.SkipRebuy()
	assert.Equal(t, "skip rebuy output", result)
	mg.AssertCalled(t, "SkipRebuy")
}

func TestHoldemInteractor_SkipRebuy_Error(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	err := errors.New("skip rebuy failed")
	mg.On("SkipRebuy").Return(err)
	mp.On("Output", mg, err).Return("skip rebuy error output")

	result := hi.SkipRebuy()
	assert.Equal(t, "skip rebuy error output", result)
}

func TestHoldemInteractor_Addon(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	mg.On("Addon").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("addon output")

	result := hi.Addon()
	assert.Equal(t, "addon output", result)
	mg.AssertCalled(t, "Addon")
}

func TestHoldemInteractor_Addon_Error(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	err := errors.New("addon failed")
	mg.On("Addon").Return(err)
	mp.On("Output", mg, err).Return("addon error output")

	result := hi.Addon()
	assert.Equal(t, "addon error output", result)
}

func TestHoldemInteractor_SkipAddon(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	mg.On("SkipAddon").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("skip addon output")

	result := hi.SkipAddon()
	assert.Equal(t, "skip addon output", result)
	mg.AssertCalled(t, "SkipAddon")
}

func TestHoldemInteractor_SkipAddon_Error(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	err := errors.New("skip addon failed")
	mg.On("SkipAddon").Return(err)
	mp.On("Output", mg, err).Return("skip addon error output")

	result := hi.SkipAddon()
	assert.Equal(t, "skip addon error output", result)
}

func TestHoldemInteractor_Muck(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	mg.On("Muck").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("muck output")

	result := hi.Muck()
	assert.Equal(t, "muck output", result)
	mg.AssertCalled(t, "Muck")
}

func TestHoldemInteractor_Muck_Error(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	err := errors.New("muck failed")
	mg.On("Muck").Return(err)
	mp.On("Output", mg, err).Return("muck error output")

	result := hi.Muck()
	assert.Equal(t, "muck error output", result)
}

func TestHoldemInteractor_ShowHand(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	mg.On("ShowHand").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("show hand output")

	result := hi.ShowHand()
	assert.Equal(t, "show hand output", result)
	mg.AssertCalled(t, "ShowHand")
}

func TestHoldemInteractor_ShowHand_Error(t *testing.T) {
	mg := new(interfaces.MockHoldemGame)
	mp := new(presenter.MockHoldemPresenter)
	hi := NewHoldemInteractor(mg, mp)

	err := errors.New("show hand failed")
	mg.On("ShowHand").Return(err)
	mp.On("Output", mg, err).Return("show hand error output")

	result := hi.ShowHand()
	assert.Equal(t, "show hand error output", result)
}
