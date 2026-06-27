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

func newTestBarbu() *domain.Barbu { return domain.NewDefaultBarbu() }

func TestNewBarbuInteractor_NilGuards(t *testing.T) {
	bpMock := new(presenter.MockBarbuPresenter)
	t.Run("panics when bg is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BarbuInteractor: bg must not be nil", func() {
			usecase.NewBarbuInteractor(nil, bpMock)
		})
	})
	t.Run("panics when bp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BarbuInteractor: bp must not be nil", func() {
			usecase.NewBarbuInteractor(newTestBarbu(), nil)
		})
	})
}

func TestBarbuInteractor_Methods(t *testing.T) {
	mockOutput := `{"players":[]}`
	bpMock := new(presenter.MockBarbuPresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	bi := usecase.NewBarbuInteractor(newTestBarbu(), bpMock)

	assert.Equal(t, mockOutput, bi.Reset())
	assert.Equal(t, mockOutput, bi.SelectContract(domain.BarbuContractNoTricks, -1))
	assert.Equal(t, mockOutput, bi.Play(0, nil))

	cfg := domain.DefaultBarbuConfig()
	assert.Equal(t, mockOutput, bi.ResetWithConfig(cfg))
	assert.Equal(t, domain.BarbuDifficultyNormal, bi.GetConfig().CpuDifficulty)
}

func TestBarbuInteractor_ResetWithInvalidConfig(t *testing.T) {
	mockOutput := `{"err":1}`
	bpMock := new(presenter.MockBarbuPresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	bi := usecase.NewBarbuInteractor(newTestBarbu(), bpMock)
	// invalid difficulty fails validation → Output called with error, no reset
	assert.Equal(t, mockOutput, bi.ResetWithConfig(domain.BarbuConfig{CpuDifficulty: 99}))
}

func TestBarbuInteractor_Snapshot(t *testing.T) {
	bpMock := new(presenter.MockBarbuPresenter)
	bi := usecase.NewBarbuInteractor(newTestBarbu(), bpMock)
	data, err := bi.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestBarbuInteractor_ActionLog(t *testing.T) {
	bpMock := new(presenter.MockBarbuPresenter)
	gameMock := new(interfaces.MockBarbuGame)
	bpMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)
	bi := usecase.NewBarbuInteractor(gameMock, bpMock)
	assert.Equal(t, `{"entries":[]}`, bi.ActionLog())
	bpMock.AssertExpectations(t)
}

func TestBarbuInteractor_Hint(t *testing.T) {
	bpMock := new(presenter.MockBarbuPresenter)
	gameMock := new(interfaces.MockBarbuGame)
	bpMock.On("HintOutput", gameMock).Return("hint output")
	bi := usecase.NewBarbuInteractor(gameMock, bpMock)
	assert.Equal(t, "hint output", bi.Hint())
	bpMock.AssertExpectations(t)
}

func TestBarbuInteractor_GuardsWhenGameEnded(t *testing.T) {
	mockOutput := `{"end":1}`
	bpMock := new(presenter.MockBarbuPresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBarbuGame)
	gameMock.On("GetGameEndFlag").Return(true)
	gameMock.On("IsHumanTurn").Return(false)
	bi := usecase.NewBarbuInteractor(gameMock, bpMock)

	// NextDeal short-circuits on game end (no NextDeal call expected)
	assert.Equal(t, mockOutput, bi.NextDeal())
	// SelectContract guarded by game end
	assert.Equal(t, mockOutput, bi.SelectContract(0, -1))
	// Play guarded (not playable)
	assert.Equal(t, mockOutput, bi.Play(0, nil))
	gameMock.AssertNotCalled(t, "NextDeal")
	gameMock.AssertNotCalled(t, "SelectContract", mock.Anything, mock.Anything)
	gameMock.AssertNotCalled(t, "PlayerPlay", mock.Anything, mock.Anything)
}

func TestBarbuInteractor_RunCpuTurnsStopsAtDealEnd(t *testing.T) {
	mockOutput := `{"ok":1}`
	bpMock := new(presenter.MockBarbuPresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBarbuGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)
	gameMock.On("GetPhase").Return(domain.BarbuPhaseDealEnd)
	bi := usecase.NewBarbuInteractor(gameMock, bpMock)
	// Reset → runCpuTurns: not human, not ended, phase=DealEnd → returns without CpuPlay
	assert.Equal(t, mockOutput, bi.Reset())
	gameMock.AssertNotCalled(t, "CpuPlay")
}

func TestBarbuInteractor_RunCpuTurnsDrivesCpu(t *testing.T) {
	mockOutput := `{"ok":1}`
	bpMock := new(presenter.MockBarbuPresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBarbuGame)
	gameMock.On("Reset").Return()
	// first IsHumanTurn → false (drives one CpuPlay), then true (exits loop)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false).Once()
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetPhase").Return(domain.BarbuPhasePlay)
	gameMock.On("CpuPlay").Return()
	bi := usecase.NewBarbuInteractor(gameMock, bpMock)
	assert.Equal(t, mockOutput, bi.Reset())
	gameMock.AssertCalled(t, "CpuPlay")
}

func TestRestoreBarbuInteractor(t *testing.T) {
	bpMock := new(presenter.MockBarbuPresenter)
	src := newTestBarbu()
	src.Reset()
	bi := usecase.NewBarbuInteractor(src, bpMock)
	data, err := bi.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreBarbuInteractor(data, bpMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)

	_, err = usecase.RestoreBarbuInteractor([]byte("not json"), bpMock)
	assert.Error(t, err)
}
