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

func newBadugiMocks() (*interfaces.MockBadugiGame, *presenter.MockBadugiPresenter) {
	return new(interfaces.MockBadugiGame), new(presenter.MockBadugiPresenter)
}

func TestNewBadugiInteractor(t *testing.T) {
	mg, mp := newBadugiMocks()
	bi := usecase.NewBadugiInteractor(mg, mp)
	assert.NotNil(t, bi)
}

func TestNewBadugiInteractor_NilGame(t *testing.T) {
	_, mp := newBadugiMocks()
	assert.PanicsWithValue(t, "BadugiInteractor: g must not be nil", func() {
		usecase.NewBadugiInteractor(nil, mp)
	})
}

func TestNewBadugiInteractor_NilPresenter(t *testing.T) {
	mg, _ := newBadugiMocks()
	assert.PanicsWithValue(t, "BadugiInteractor: pp must not be nil", func() {
		usecase.NewBadugiInteractor(mg, nil)
	})
}

func TestBadugiInteractor_Reset(t *testing.T) {
	mg, mp := newBadugiMocks()
	bi := usecase.NewBadugiInteractor(mg, mp)

	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset output")

	assert.Equal(t, "reset output", bi.Reset())
	mg.AssertCalled(t, "Reset")
}

func TestBadugiInteractor_Reset_Error(t *testing.T) {
	mg, mp := newBadugiMocks()
	bi := usecase.NewBadugiInteractor(mg, mp)

	err := errors.New("reset failed")
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	assert.Equal(t, "error output", bi.Reset())
}

func TestBadugiInteractor_GetConfig(t *testing.T) {
	mg, mp := newBadugiMocks()
	bi := usecase.NewBadugiInteractor(mg, mp)

	want := domain.BadugiConfig{InitChips: 1000, Ante: 10, MinBet: 10, CpuCount: 2}
	mg.On("GetConfig").Return(want)

	assert.Equal(t, want, bi.GetConfig())
}

func TestBadugiInteractor_ResetWithConfig(t *testing.T) {
	mg, mp := newBadugiMocks()
	bi := usecase.NewBadugiInteractor(mg, mp)

	cfg := domain.BadugiConfig{InitChips: 2000, Ante: 20, MinBet: 20, CpuCount: 2}
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mp.On("Output", mg, mock.Anything).Return("reset with config output")

	assert.Equal(t, "reset with config output", bi.ResetWithConfig(cfg, nil))
	mg.AssertCalled(t, "SetConfig", cfg)
	mg.AssertCalled(t, "Reset")
}

func TestBadugiInteractor_ResetWithConfig_InvalidConfig(t *testing.T) {
	mg, mp := newBadugiMocks()
	bi := usecase.NewBadugiInteractor(mg, mp)

	// CpuCount 0 fails validation.
	cfg := domain.BadugiConfig{InitChips: 1000, Ante: 10, MinBet: 10, CpuCount: 0}
	mp.On("Output", mg, mock.Anything).Return("invalid config output")

	assert.Equal(t, "invalid config output", bi.ResetWithConfig(cfg, nil))
	mg.AssertNotCalled(t, "SetConfig")
	mg.AssertNotCalled(t, "Reset")
}

func TestBadugiInteractor_ResetWithConfig_ResetError(t *testing.T) {
	mg, mp := newBadugiMocks()
	bi := usecase.NewBadugiInteractor(mg, mp)

	cfg := domain.BadugiConfig{InitChips: 1000, Ante: 10, MinBet: 10, CpuCount: 3}
	err := errors.New("reset failed")
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(err)
	mp.On("Output", mg, err).Return("error output")

	assert.Equal(t, "error output", bi.ResetWithConfig(cfg, nil))
}

func TestBadugiInteractor_ResetWithConfig_ImportsProfile(t *testing.T) {
	mg, mp := newBadugiMocks()
	bi := usecase.NewBadugiInteractor(mg, mp)

	cfg := domain.BadugiConfig{InitChips: 2000, Ante: 20, MinBet: 20, CpuCount: 2}
	profileData := []byte(`{"gamesPlayed":3}`)
	mg.On("SetConfig", cfg).Return()
	mg.On("Reset").Return(nil)
	mg.On("ImportProfile", profileData).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("with profile output")

	assert.Equal(t, "with profile output", bi.ResetWithConfig(cfg, profileData))
	mg.AssertCalled(t, "ImportProfile", profileData)
}

func TestBadugiInteractor_Action(t *testing.T) {
	mg, mp := newBadugiMocks()
	bi := usecase.NewBadugiInteractor(mg, mp)

	mg.On("PlayerAction", domain.BadugiActionCheck, 0, 0).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("action output")

	assert.Equal(t, "action output", bi.Action(domain.BadugiActionCheck, 0, 0))
	mg.AssertCalled(t, "PlayerAction", domain.BadugiActionCheck, 0, 0)
}

func TestBadugiInteractor_Action_Error(t *testing.T) {
	mg, mp := newBadugiMocks()
	bi := usecase.NewBadugiInteractor(mg, mp)

	err := errors.New("action failed")
	mg.On("PlayerAction", domain.BadugiActionBet, 50, 0).Return(err)
	mp.On("Output", mg, err).Return("error output")

	assert.Equal(t, "error output", bi.Action(domain.BadugiActionBet, 50, 0))
}

func TestBadugiInteractor_Exchange(t *testing.T) {
	mg, mp := newBadugiMocks()
	bi := usecase.NewBadugiInteractor(mg, mp)

	indices := []int{0, 2}
	mg.On("PlayerExchange", indices, 4200).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("exchange output")

	assert.Equal(t, "exchange output", bi.Exchange(indices, 4200))
	mg.AssertCalled(t, "PlayerExchange", indices, 4200)
}

func TestBadugiInteractor_Stand(t *testing.T) {
	mg, mp := newBadugiMocks()
	bi := usecase.NewBadugiInteractor(mg, mp)

	mg.On("PlayerStand", 4200).Return(nil)
	mp.On("Output", mg, mock.Anything).Return("stand output")

	assert.Equal(t, "stand output", bi.Stand(4200))
	mg.AssertCalled(t, "PlayerStand", 4200)
}

func TestBadugiInteractor_ActionLog(t *testing.T) {
	mg, mp := newBadugiMocks()
	mp.On("ActionLogOutput", mg).Return(`{"entries":[]}`)
	bi := usecase.NewBadugiInteractor(mg, mp)

	assert.Equal(t, `{"entries":[]}`, bi.ActionLog())
	mp.AssertExpectations(t)
}

func TestBadugiInteractor_Hint(t *testing.T) {
	mg, mp := newBadugiMocks()
	mp.On("HintOutput", mg).Return("hint output")
	bi := usecase.NewBadugiInteractor(mg, mp)

	assert.Equal(t, "hint output", bi.Hint())
	mp.AssertExpectations(t)
}

func TestBadugiInteractor_Snapshot(t *testing.T) {
	g := domain.NewDefaultBadugi()
	bi := usecase.NewBadugiInteractor(g, new(presenter.MockBadugiPresenter))
	data, err := bi.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreBadugiInteractor(t *testing.T) {
	g := domain.NewDefaultBadugi()
	assert.NoError(t, g.Reset())
	bi := usecase.NewBadugiInteractor(g, new(presenter.MockBadugiPresenter))
	data, err := bi.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreBadugiInteractor(data, new(presenter.MockBadugiPresenter))
	assert.NoError(t, err)
	assert.NotNil(t, restored)
	assert.Equal(t, g.GetConfig(), restored.GetConfig())
}
