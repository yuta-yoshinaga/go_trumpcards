//go:build test

package usecase_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

const trenteEtQuaranteMockOutput = `{"phase":1}`

func newTrenteEtQuaranteMocks() (*interfaces.MockTrenteEtQuaranteGame, *presenter.MockTrenteEtQuarantePresenter) {
	return new(interfaces.MockTrenteEtQuaranteGame), new(presenter.MockTrenteEtQuarantePresenter)
}

func TestNewTrenteEtQuaranteInteractor_NilGuards(t *testing.T) {
	cpMock := new(presenter.MockTrenteEtQuarantePresenter)
	assert.PanicsWithValue(t, "TrenteEtQuaranteInteractor: bg must not be nil", func() {
		usecase.NewTrenteEtQuaranteInteractor(nil, cpMock)
	})
	gameMock := new(interfaces.MockTrenteEtQuaranteGame)
	assert.PanicsWithValue(t, "TrenteEtQuaranteInteractor: cp must not be nil", func() {
		usecase.NewTrenteEtQuaranteInteractor(gameMock, nil)
	})
}

func TestTrenteEtQuaranteInteractor_Bet(t *testing.T) {
	gm, cp := newTrenteEtQuaranteMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(trenteEtQuaranteMockOutput)
	gm.On("PlaceBet", domain.TrenteEtQuaranteBetRouge, 100).Return(nil)

	bi := usecase.NewTrenteEtQuaranteInteractor(gm, cp)
	assert.Equal(t, trenteEtQuaranteMockOutput, bi.Bet(domain.TrenteEtQuaranteBetRouge, 100))
	gm.AssertCalled(t, "PlaceBet", domain.TrenteEtQuaranteBetRouge, 100)
}

func TestTrenteEtQuaranteInteractor_Bet_Error(t *testing.T) {
	gm, cp := newTrenteEtQuaranteMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(trenteEtQuaranteMockOutput)
	gm.On("PlaceBet", domain.TrenteEtQuaranteBetNoir, 5).Return(assert.AnError)

	bi := usecase.NewTrenteEtQuaranteInteractor(gm, cp)
	assert.Equal(t, trenteEtQuaranteMockOutput, bi.Bet(domain.TrenteEtQuaranteBetNoir, 5))
}

func TestTrenteEtQuaranteInteractor_ResetAndNextRound(t *testing.T) {
	gm, cp := newTrenteEtQuaranteMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(trenteEtQuaranteMockOutput)
	gm.On("Reset").Return()
	gm.On("NextRound").Return()

	bi := usecase.NewTrenteEtQuaranteInteractor(gm, cp)
	assert.Equal(t, trenteEtQuaranteMockOutput, bi.Reset())
	assert.Equal(t, trenteEtQuaranteMockOutput, bi.NextRound())
	gm.AssertCalled(t, "Reset")
	gm.AssertCalled(t, "NextRound")
}

func TestTrenteEtQuaranteInteractor_ResetWithConfigInvalid(t *testing.T) {
	gm, cp := newTrenteEtQuaranteMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(trenteEtQuaranteMockOutput)
	bi := usecase.NewTrenteEtQuaranteInteractor(gm, cp)
	out := bi.ResetWithConfig(domain.TrenteEtQuaranteConfig{DefaultBet: 99})
	assert.Equal(t, trenteEtQuaranteMockOutput, out)
	gm.AssertNotCalled(t, "Reset")
}

func TestTrenteEtQuaranteInteractor_ResetWithConfigValid(t *testing.T) {
	gm, cp := newTrenteEtQuaranteMocks()
	cp.On("Output", mock.Anything, mock.Anything).Return(trenteEtQuaranteMockOutput)
	gm.On("SetConfig", mock.Anything).Return()
	gm.On("Reset").Return()
	bi := usecase.NewTrenteEtQuaranteInteractor(gm, cp)
	out := bi.ResetWithConfig(domain.DefaultTrenteEtQuaranteConfig())
	assert.Equal(t, trenteEtQuaranteMockOutput, out)
	gm.AssertCalled(t, "Reset")
}

func TestTrenteEtQuaranteInteractor_HintAndLog(t *testing.T) {
	gm, cp := newTrenteEtQuaranteMocks()
	cp.On("HintOutput", mock.Anything).Return("hint")
	cp.On("ActionLogOutput", mock.Anything).Return("log")

	bi := usecase.NewTrenteEtQuaranteInteractor(gm, cp)
	assert.Equal(t, "hint", bi.Hint())
	assert.Equal(t, "log", bi.ActionLog())
}

func TestTrenteEtQuaranteInteractor_GetConfig(t *testing.T) {
	gm, cp := newTrenteEtQuaranteMocks()
	cfg := domain.TrenteEtQuaranteConfig{DefaultBet: domain.TrenteEtQuaranteBetCouleur}
	gm.On("GetConfig").Return(cfg)
	bi := usecase.NewTrenteEtQuaranteInteractor(gm, cp)
	assert.Equal(t, cfg, bi.GetConfig())
}

func TestTrenteEtQuaranteInteractor_RealFlow(t *testing.T) {
	cp := new(presenter.MockTrenteEtQuarantePresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(trenteEtQuaranteMockOutput)

	g := domain.NewDefaultTrenteEtQuarante()
	bi := usecase.NewTrenteEtQuaranteInteractor(g, cp)

	bi.Reset()
	assert.Equal(t, domain.TrenteEtQuarantePhaseBet, g.GetPhase())
	bi.Bet(domain.TrenteEtQuaranteBetNoir, 100)
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.TrenteEtQuarantePhaseResult, g.GetPhase())

	bi.NextRound()
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, domain.TrenteEtQuarantePhaseBet, g.GetPhase())
}

func TestTrenteEtQuaranteInteractor_SnapshotAndRestore(t *testing.T) {
	cp := new(presenter.MockTrenteEtQuarantePresenter)
	cp.On("Output", mock.Anything, mock.Anything).Return(trenteEtQuaranteMockOutput)

	real := usecase.NewTrenteEtQuaranteInteractor(domain.NewDefaultTrenteEtQuarante(), cp)
	real.Reset()
	data, err := real.Snapshot()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestoreTrenteEtQuaranteInteractor(data, cp)
	require.NoError(t, err)
	assert.NotNil(t, restored)

	_, err = usecase.RestoreTrenteEtQuaranteInteractor([]byte("not json"), cp)
	assert.Error(t, err)

	var g domain.TrenteEtQuarante
	require.NoError(t, json.Unmarshal(data, &g))
}
