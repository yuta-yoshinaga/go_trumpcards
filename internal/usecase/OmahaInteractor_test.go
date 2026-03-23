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

func TestNewOmahaInteractor(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)
	assert.NotNil(t, hi)
}

func TestNewOmahaInteractor_NilGame(t *testing.T) {
	mp := new(presenter.MockOmahaPresenter)
	assert.Panics(t, func() {
		NewOmahaInteractor(nil, mp)
	})
}

func TestNewOmahaInteractor_NilPresenter(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	assert.Panics(t, func() {
		NewOmahaInteractor(mg, nil)
	})
}

func TestOmahaInteractor_Reset(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset output")

	result := hi.Reset()
	assert.Equal(t, "reset output", result)
	mg.AssertCalled(t, "Reset")
}

func TestOmahaInteractor_Reset_Error(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	err := errors.New("reset failed")
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := hi.Reset()
	assert.Equal(t, "error output", result)
}

func TestOmahaInteractor_ResetWithConfig_Error(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	cfg := domain.OmahaConfig{SmallBlind: 10, BigBlind: 20, InitChips: 2000, BlindLevelHands: 10}
	err := errors.New("reset failed")
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := hi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "error output", result)
}

func TestOmahaInteractor_ResetWithConfig(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	cfg := domain.OmahaConfig{SmallBlind: 10, BigBlind: 20, InitChips: 2000, BlindLevelHands: 10}
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset with config output")

	result := hi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "reset with config output", result)
	mg.AssertCalled(t, "SetConfig", cfg)
	mg.AssertCalled(t, "Reset")
}

func TestOmahaInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	mp.On("Output", mg, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")
	cfg := domain.OmahaConfig{SmallBlind: 0, BigBlind: 10, BlindLevelHands: 10}
	result := hi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "validation error", result)
	mg.AssertNotCalled(t, "SetConfig", mock.Anything)
}

func TestOmahaInteractor_Action(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	mg.On("PlayerAction", domain.OmahaActionCheck, 0, 0).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("action output")

	result := hi.Action(domain.OmahaActionCheck, 0, 0)
	assert.Equal(t, "action output", result)
	mg.AssertCalled(t, "PlayerAction", domain.OmahaActionCheck, 0, 0)
}

func TestOmahaInteractor_GetConfig(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	cfg := domain.DefaultOmahaConfig()
	mg.On("GetConfig").Return(cfg)

	result := hi.GetConfig()
	assert.Equal(t, cfg, result)
	mg.AssertCalled(t, "GetConfig")
}

func TestOmahaInteractor_ResetWithConfig_TableSizeChange(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	cfg := domain.DefaultOmahaConfig()
	cfg.TableSize = domain.HoldemTableSize6
	mg.On("GetPlayerCnt").Return(4)
	mg.On("Resize", mock.MatchedBy(func(players []*domain.OmahaPlayer) bool {
		return len(players) == 6 && players[0].GetIsHuman()
	})).Return()
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("resize output")

	result := hi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "resize output", result)
	mg.AssertCalled(t, "Resize", mock.Anything)
}

func TestOmahaInteractor_ResetWithConfig_SameTableSize(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	cfg := domain.DefaultOmahaConfig()
	cfg.TableSize = domain.HoldemTableSize4
	mg.On("GetPlayerCnt").Return(4)
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("no resize output")

	result := hi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "no resize output", result)
	mg.AssertNotCalled(t, "Resize", mock.Anything)
}

func TestOmahaInteractor_ResetWithConfig_TableSizeZero(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	cfg := domain.DefaultOmahaConfig()
	cfg.TableSize = 0 // not set, should skip resize
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("zero output")

	result := hi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "zero output", result)
	mg.AssertNotCalled(t, "Resize", mock.Anything)
}

func TestOmahaInteractor_ResetWithConfig_WithProfile(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	cfg := domain.OmahaConfig{SmallBlind: 10, BigBlind: 20, InitChips: 2000, BlindLevelHands: 10}
	profileData := []byte(`{"gamesPlayed":3}`)
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mg.On("ImportProfile", profileData).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("with profile output")

	result := hi.ResetWithConfig(cfg, profileData)
	assert.Equal(t, "with profile output", result)
	mg.AssertCalled(t, "ImportProfile", profileData)
}

func TestOmahaInteractor_ActionLog(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	mp.On("ActionLogOutput", mg).Return(`{"entries":[]}`)

	hi := NewOmahaInteractor(mg, mp)
	result := hi.ActionLog()
	assert.Equal(t, `{"entries":[]}`, result)
	mp.AssertExpectations(t)
}

func TestOmahaInteractor_Action_Error(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	err := errors.New("test error")
	mg.On("PlayerAction", domain.OmahaActionBet, 50, 0).Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := hi.Action(domain.OmahaActionBet, 50, 0)
	assert.Equal(t, "error output", result)
}

func TestOmahaInteractor_Rebuy(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	mg.On("Rebuy").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("rebuy output")

	result := hi.Rebuy()
	assert.Equal(t, "rebuy output", result)
	mg.AssertCalled(t, "Rebuy")
}

func TestOmahaInteractor_Rebuy_Error(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	err := errors.New("rebuy failed")
	mg.On("Rebuy").Return(err)
	mp.On("Output", mg, err).Return("rebuy error output")

	result := hi.Rebuy()
	assert.Equal(t, "rebuy error output", result)
}

func TestOmahaInteractor_SkipRebuy(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	mg.On("SkipRebuy").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("skip rebuy output")

	result := hi.SkipRebuy()
	assert.Equal(t, "skip rebuy output", result)
	mg.AssertCalled(t, "SkipRebuy")
}

func TestOmahaInteractor_SkipRebuy_Error(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	err := errors.New("skip rebuy failed")
	mg.On("SkipRebuy").Return(err)
	mp.On("Output", mg, err).Return("skip rebuy error output")

	result := hi.SkipRebuy()
	assert.Equal(t, "skip rebuy error output", result)
}

func TestOmahaInteractor_Addon(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	mg.On("Addon").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("addon output")

	result := hi.Addon()
	assert.Equal(t, "addon output", result)
	mg.AssertCalled(t, "Addon")
}

func TestOmahaInteractor_Addon_Error(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	err := errors.New("addon failed")
	mg.On("Addon").Return(err)
	mp.On("Output", mg, err).Return("addon error output")

	result := hi.Addon()
	assert.Equal(t, "addon error output", result)
}

func TestOmahaInteractor_SkipAddon(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	mg.On("SkipAddon").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("skip addon output")

	result := hi.SkipAddon()
	assert.Equal(t, "skip addon output", result)
	mg.AssertCalled(t, "SkipAddon")
}

func TestOmahaInteractor_SkipAddon_Error(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	err := errors.New("skip addon failed")
	mg.On("SkipAddon").Return(err)
	mp.On("Output", mg, err).Return("skip addon error output")

	result := hi.SkipAddon()
	assert.Equal(t, "skip addon error output", result)
}

func TestOmahaInteractor_Muck(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	mg.On("Muck").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("muck output")

	result := hi.Muck()
	assert.Equal(t, "muck output", result)
	mg.AssertCalled(t, "Muck")
}

func TestOmahaInteractor_Muck_Error(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	err := errors.New("muck failed")
	mg.On("Muck").Return(err)
	mp.On("Output", mg, err).Return("muck error output")

	result := hi.Muck()
	assert.Equal(t, "muck error output", result)
}

func TestOmahaInteractor_ShowHand(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	mg.On("ShowHand").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("show hand output")

	result := hi.ShowHand()
	assert.Equal(t, "show hand output", result)
	mg.AssertCalled(t, "ShowHand")
}

func TestOmahaInteractor_ShowHand_Error(t *testing.T) {
	mg := new(interfaces.MockOmahaGame)
	mp := new(presenter.MockOmahaPresenter)
	hi := NewOmahaInteractor(mg, mp)

	err := errors.New("show hand failed")
	mg.On("ShowHand").Return(err)
	mp.On("Output", mg, err).Return("show hand error output")

	result := hi.ShowHand()
	assert.Equal(t, "show hand error output", result)
}
