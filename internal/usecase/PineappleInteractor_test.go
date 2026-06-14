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

func TestNewPineappleInteractor(t *testing.T) {
	mg := new(interfaces.MockPineappleGame)
	mp := new(presenter.MockPineapplePresenter)
	pi := NewPineappleInteractor(mg, mp)
	assert.NotNil(t, pi)
}

func TestNewPineappleInteractor_NilGame(t *testing.T) {
	mp := new(presenter.MockPineapplePresenter)
	assert.Panics(t, func() {
		NewPineappleInteractor(nil, mp)
	})
}

func TestNewPineappleInteractor_NilPresenter(t *testing.T) {
	mg := new(interfaces.MockPineappleGame)
	assert.Panics(t, func() {
		NewPineappleInteractor(mg, nil)
	})
}

func TestPineappleInteractor_Reset(t *testing.T) {
	mg := new(interfaces.MockPineappleGame)
	mp := new(presenter.MockPineapplePresenter)
	pi := NewPineappleInteractor(mg, mp)

	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset output")

	result := pi.Reset()
	assert.Equal(t, "reset output", result)
	mg.AssertCalled(t, "Reset")
}

func TestPineappleInteractor_Reset_Error(t *testing.T) {
	mg := new(interfaces.MockPineappleGame)
	mp := new(presenter.MockPineapplePresenter)
	pi := NewPineappleInteractor(mg, mp)

	err := errors.New("reset failed")
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := pi.Reset()
	assert.Equal(t, "error output", result)
}

func TestPineappleInteractor_ResetWithConfig(t *testing.T) {
	mg := new(interfaces.MockPineappleGame)
	mp := new(presenter.MockPineapplePresenter)
	pi := NewPineappleInteractor(mg, mp)

	cfg := domain.PineappleConfig{SmallBlind: 10, BigBlind: 20, InitChips: 2000, BlindLevelHands: 10}
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("ok")

	result := pi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "ok", result)
}

func TestPineappleInteractor_ResetWithConfig_ValidationError(t *testing.T) {
	mg := new(interfaces.MockPineappleGame)
	mp := new(presenter.MockPineapplePresenter)
	pi := NewPineappleInteractor(mg, mp)

	cfg := domain.PineappleConfig{SmallBlind: 0, BigBlind: 0} // invalid
	mp.On("Output", mg, mock.Anything).Return("validation error")

	result := pi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "validation error", result)
}

func TestPineappleInteractor_ResetWithConfig_Resize(t *testing.T) {
	mg := new(interfaces.MockPineappleGame)
	mp := new(presenter.MockPineapplePresenter)
	pi := NewPineappleInteractor(mg, mp)

	cfg := domain.PineappleConfig{SmallBlind: 5, BigBlind: 10, InitChips: 1000, BlindLevelHands: 10, TableSize: 6}
	mg.On("GetPlayerCnt").Return(4)
	mg.On("Resize", mock.Anything).Return()
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("ok")

	result := pi.ResetWithConfig(cfg, nil)
	assert.Equal(t, "ok", result)
	mg.AssertCalled(t, "Resize", mock.Anything)
}

func TestPineappleInteractor_Action(t *testing.T) {
	mg := new(interfaces.MockPineappleGame)
	mp := new(presenter.MockPineapplePresenter)
	pi := NewPineappleInteractor(mg, mp)

	mg.On("PlayerAction", 1, 100, 500).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("action output")

	result := pi.Action(1, 100, 500)
	assert.Equal(t, "action output", result)
}

func TestPineappleInteractor_Discard(t *testing.T) {
	mg := new(interfaces.MockPineappleGame)
	mp := new(presenter.MockPineapplePresenter)
	pi := NewPineappleInteractor(mg, mp)

	mg.On("DiscardCard", 2).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("discard output")

	result := pi.Discard(2)
	assert.Equal(t, "discard output", result)
	mg.AssertCalled(t, "DiscardCard", 2)
}

func TestPineappleInteractor_DiscardMany(t *testing.T) {
	mg := new(interfaces.MockPineappleGame)
	mp := new(presenter.MockPineapplePresenter)
	pi := NewPineappleInteractor(mg, mp)

	mg.On("DiscardCards", []int{1, 3}).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("discard many output")

	result := pi.DiscardMany([]int{1, 3})
	assert.Equal(t, "discard many output", result)
	mg.AssertCalled(t, "DiscardCards", []int{1, 3})
}

func TestPineappleInteractor_Discard_Error(t *testing.T) {
	mg := new(interfaces.MockPineappleGame)
	mp := new(presenter.MockPineapplePresenter)
	pi := NewPineappleInteractor(mg, mp)

	err := errors.New("wrong phase")
	mg.On("DiscardCard", 0).Return(err)
	mp.On("Output", mg, err).Return("error output")

	result := pi.Discard(0)
	assert.Equal(t, "error output", result)
}

func TestPineappleInteractor_GetConfig(t *testing.T) {
	mg := new(interfaces.MockPineappleGame)
	mp := new(presenter.MockPineapplePresenter)
	pi := NewPineappleInteractor(mg, mp)

	expected := domain.DefaultPineappleConfig()
	mg.On("GetConfig").Return(expected)

	cfg := pi.GetConfig()
	assert.Equal(t, expected, cfg)
}

func TestPineappleInteractor_Rebuy(t *testing.T) {
	mg := new(interfaces.MockPineappleGame)
	mp := new(presenter.MockPineapplePresenter)
	pi := NewPineappleInteractor(mg, mp)

	mg.On("Rebuy").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("rebuy output")
	assert.Equal(t, "rebuy output", pi.Rebuy())

	mg2 := new(interfaces.MockPineappleGame)
	mp2 := new(presenter.MockPineapplePresenter)
	pi2 := NewPineappleInteractor(mg2, mp2)
	mg2.On("SkipRebuy").Return(nil)
	mp2.On("Output", mg2, mock.Anything).Return("skip")
	assert.Equal(t, "skip", pi2.SkipRebuy())
}

func TestPineappleInteractor_Addon(t *testing.T) {
	mg := new(interfaces.MockPineappleGame)
	mp := new(presenter.MockPineapplePresenter)
	pi := NewPineappleInteractor(mg, mp)

	mg.On("Addon").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("addon output")
	assert.Equal(t, "addon output", pi.Addon())

	mg2 := new(interfaces.MockPineappleGame)
	mp2 := new(presenter.MockPineapplePresenter)
	pi2 := NewPineappleInteractor(mg2, mp2)
	mg2.On("SkipAddon").Return(nil)
	mp2.On("Output", mg2, mock.Anything).Return("skip")
	assert.Equal(t, "skip", pi2.SkipAddon())
}

func TestPineappleInteractor_MuckShow(t *testing.T) {
	mg := new(interfaces.MockPineappleGame)
	mp := new(presenter.MockPineapplePresenter)
	pi := NewPineappleInteractor(mg, mp)

	mg.On("Muck").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("muck output")
	assert.Equal(t, "muck output", pi.Muck())

	mg2 := new(interfaces.MockPineappleGame)
	mp2 := new(presenter.MockPineapplePresenter)
	pi2 := NewPineappleInteractor(mg2, mp2)
	mg2.On("ShowHand").Return(nil)
	mp2.On("Output", mg2, mock.Anything).Return("show output")
	assert.Equal(t, "show output", pi2.ShowHand())
}

func TestPineappleInteractor_ActionLog(t *testing.T) {
	mg := new(interfaces.MockPineappleGame)
	mp := new(presenter.MockPineapplePresenter)
	pi := NewPineappleInteractor(mg, mp)

	mp.On("ActionLogOutput", mg).Return("log output")
	assert.Equal(t, "log output", pi.ActionLog())
}

func TestPineappleInteractor_Snapshot(t *testing.T) {
	cfg := domain.DefaultPineappleConfig()
	players := domain.NewPineapplePlayersForTable(cfg.TableSize)
	game := domain.NewPineapple(domain.NewTrumpCards(0), players, cfg)
	pi := NewPineappleInteractor(game, &presenter.MockPineapplePresenter{})

	data, err := pi.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestorePineappleInteractor(t *testing.T) {
	cfg := domain.DefaultPineappleConfig()
	players := domain.NewPineapplePlayersForTable(cfg.TableSize)
	game := domain.NewPineapple(domain.NewTrumpCards(0), players, cfg)
	original := NewPineappleInteractor(game, &presenter.MockPineapplePresenter{})

	data, err := original.Snapshot()
	assert.NoError(t, err)

	mp := new(presenter.MockPineapplePresenter)
	restored, err := RestorePineappleInteractor(data, mp)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestorePineappleInteractor_InvalidJSON(t *testing.T) {
	mp := new(presenter.MockPineapplePresenter)
	_, err := RestorePineappleInteractor([]byte("invalid"), mp)
	assert.Error(t, err)
}
