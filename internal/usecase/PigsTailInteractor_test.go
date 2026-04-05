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

func newTestPigsTail() *domain.PigsTail {
	players := []*domain.PigsTailPlayer{
		domain.NewPigsTailPlayer(true),
		domain.NewPigsTailPlayer(false),
		domain.NewPigsTailPlayer(false),
		domain.NewPigsTailPlayer(false),
	}
	return domain.NewPigsTail(domain.NewTrumpCards(0), players)
}

func TestNewPigsTailInteractor_NilGuards(t *testing.T) {
	ptpMock := new(presenter.MockPigsTailPresenter)
	t.Run("panics when pt is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "PigsTailInteractor: pt must not be nil", func() {
			usecase.NewPigsTailInteractor(nil, ptpMock)
		})
	})
	t.Run("panics when ptp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "PigsTailInteractor: ptp must not be nil", func() {
			usecase.NewPigsTailInteractor(newTestPigsTail(), nil)
		})
	})
}

func TestPigsTailInteractor_Reset(t *testing.T) {
	mockOutput := `{"players":[]}`
	ptpMock := new(presenter.MockPigsTailPresenter)
	ptpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	tpi := usecase.NewPigsTailInteractor(newTestPigsTail(), ptpMock)

	result := tpi.Reset(domain.DefaultPigsTailConfig())
	assert.Equal(t, mockOutput, result)
}

func TestPigsTailInteractor_Action(t *testing.T) {
	mockOutput := `{"players":[]}`
	ptpMock := new(presenter.MockPigsTailPresenter)
	ptpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	tpi := usecase.NewPigsTailInteractor(newTestPigsTail(), ptpMock)

	// Reset first to initialize game
	tpi.Reset(domain.DefaultPigsTailConfig())
	result := tpi.Action(0)
	assert.Equal(t, mockOutput, result)
}

func TestPigsTailInteractor_MockGame(t *testing.T) {
	mockOutput := `{"players":[]}`
	ptpMock := new(presenter.MockPigsTailPresenter)
	ptpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockPigsTailGame)
	gameMock.On("Reset").Return()
	gameMock.On("SetConfig", mock.Anything).Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuAction").Return(nil)
	gameMock.On("PlayerAction", mock.Anything).Return(nil)

	pi := usecase.NewPigsTailInteractor(gameMock, ptpMock)

	t.Run("Reset calls SetConfig and game.Reset", func(t *testing.T) {
		result := pi.Reset(domain.DefaultPigsTailConfig())
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "SetConfig", mock.Anything)
		gameMock.AssertCalled(t, "Reset")
	})
	t.Run("Action calls game.PlayerAction when human turn", func(t *testing.T) {
		result := pi.Action(0)
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerAction", 0)
	})
}

func TestPigsTailInteractor_ActionLog(t *testing.T) {
	mockLogOutput := `[{"t":1}]`
	ptpMock := new(presenter.MockPigsTailPresenter)
	ptpMock.On("ActionLogOutput", mock.Anything).Return(mockLogOutput)
	tpi := usecase.NewPigsTailInteractor(newTestPigsTail(), ptpMock)

	result := tpi.ActionLog()
	assert.Equal(t, mockLogOutput, result)
}

func TestPigsTailInteractor_GetConfig(t *testing.T) {
	ptpMock := new(presenter.MockPigsTailPresenter)
	ptpMock.On("Output", mock.Anything, mock.Anything).Return("")
	tpi := usecase.NewPigsTailInteractor(newTestPigsTail(), ptpMock)

	cfg := tpi.GetConfig()
	assert.Equal(t, domain.DefaultPigsTailConfig(), cfg)
}

func TestPigsTailInteractor_SnapshotRestore(t *testing.T) {
	ptpMock := new(presenter.MockPigsTailPresenter)
	ptpMock.On("Output", mock.Anything, mock.Anything).Return("")
	tpi := usecase.NewPigsTailInteractor(newTestPigsTail(), ptpMock)
	tpi.Reset(domain.DefaultPigsTailConfig())

	data, err := tpi.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	restored, err := usecase.RestorePigsTailInteractor(data, ptpMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestorePigsTailInteractor_InvalidJSON(t *testing.T) {
	ptpMock := new(presenter.MockPigsTailPresenter)
	_, err := usecase.RestorePigsTailInteractor([]byte(`{invalid`), ptpMock)
	assert.Error(t, err)
}
