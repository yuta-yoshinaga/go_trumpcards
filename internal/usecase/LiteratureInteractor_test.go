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

const literatureMockOutput = `{"phase":0}`

func TestNewLiteratureInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockLiteraturePresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "LiteratureInteractor: g must not be nil", func() {
			usecase.NewLiteratureInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockLiteratureGame)
		assert.PanicsWithValue(t, "LiteratureInteractor: gp must not be nil", func() {
			usecase.NewLiteratureInteractor(gameMock, nil)
		})
	})
}

func TestLiteratureInteractor_Reset(t *testing.T) {
	pMock := new(presenter.MockLiteraturePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(literatureMockOutput)
	gameMock := new(interfaces.MockLiteratureGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)

	li := usecase.NewLiteratureInteractor(gameMock, pMock)
	assert.Equal(t, literatureMockOutput, li.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestLiteratureInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockLiteraturePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(literatureMockOutput)
		gameMock := new(interfaces.MockLiteratureGame)
		cfg := domain.DefaultLiteratureConfig()
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)

		li := usecase.NewLiteratureInteractor(gameMock, pMock)
		assert.Equal(t, literatureMockOutput, li.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config never reaches the game", func(t *testing.T) {
		pMock := new(presenter.MockLiteraturePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(literatureMockOutput)
		gameMock := new(interfaces.MockLiteratureGame)

		li := usecase.NewLiteratureInteractor(gameMock, pMock)
		assert.Equal(t, literatureMockOutput, li.ResetWithConfig(domain.LiteratureConfig{CpuDifficulty: 9}))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
		gameMock.AssertNotCalled(t, "Reset")
	})
}

// **要求も宣言も currentIdx から出る。**
func TestLiteratureInteractor_UsesTheCurrentSeat(t *testing.T) {
	t.Run("ask", func(t *testing.T) {
		pMock := new(presenter.MockLiteraturePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(literatureMockOutput)
		gameMock := new(interfaces.MockLiteratureGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("GetCurrentPlayerIdx").Return(2)
		gameMock.On("Ask", 2, 3, mock.MatchedBy(func(c *domain.Card) bool {
			return c != nil && c.GetDesign() == domain.CardDesignHeart && c.GetValue() == 5
		})).Return(nil)

		li := usecase.NewLiteratureInteractor(gameMock, pMock)
		assert.Equal(t, literatureMockOutput, li.Ask(3, domain.CardDesignHeart, 5))
		gameMock.AssertNumberOfCalls(t, "Ask", 1)
	})

	t.Run("claim", func(t *testing.T) {
		pMock := new(presenter.MockLiteraturePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(literatureMockOutput)
		gameMock := new(interfaces.MockLiteratureGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("GetCurrentPlayerIdx").Return(0)
		holders := []int{0, 0, 2, 2, 4, 4}
		gameMock.On("Claim", 0, 3, holders).Return(nil)

		li := usecase.NewLiteratureInteractor(gameMock, pMock)
		assert.Equal(t, literatureMockOutput, li.Claim(3, holders))
		gameMock.AssertCalled(t, "Claim", 0, 3, holders)
	})
}

func TestLiteratureInteractor_BlockedWhenNotTheHumansTurn(t *testing.T) {
	pMock := new(presenter.MockLiteraturePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(literatureMockOutput)
	gameMock := new(interfaces.MockLiteratureGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(false)

	li := usecase.NewLiteratureInteractor(gameMock, pMock)
	assert.Equal(t, literatureMockOutput, li.Ask(1, domain.CardDesignSpade, 2))
	assert.Equal(t, literatureMockOutput, li.Claim(0, []int{0, 0, 0, 0, 0, 0}))
	gameMock.AssertNotCalled(t, "Ask", mock.Anything, mock.Anything, mock.Anything)
	gameMock.AssertNotCalled(t, "Claim", mock.Anything, mock.Anything, mock.Anything)
}

func TestLiteratureInteractor_DomainErrorIsPresented(t *testing.T) {
	wantErr := errors.New("you may only ask an opponent")
	pMock := new(presenter.MockLiteraturePresenter)
	pMock.On("Output", mock.Anything, wantErr).Return(literatureMockOutput)
	gameMock := new(interfaces.MockLiteratureGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetCurrentPlayerIdx").Return(0)
	gameMock.On("Ask", 0, 2, mock.Anything).Return(wantErr)

	li := usecase.NewLiteratureInteractor(gameMock, pMock)
	assert.Equal(t, literatureMockOutput, li.Ask(2, domain.CardDesignSpade, 2))
	pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
	gameMock.AssertNotCalled(t, "CpuPlay")
}

// **CPU ループは人間の手番とゲーム終了で止まる。**
func TestLiteratureInteractor_RunCpuTurnsStops(t *testing.T) {
	t.Run("at the human's turn", func(t *testing.T) {
		pMock := new(presenter.MockLiteraturePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(literatureMockOutput)
		gameMock := new(interfaces.MockLiteratureGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false).Times(4)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("CpuPlay").Return()

		li := usecase.NewLiteratureInteractor(gameMock, pMock)
		li.Reset()
		gameMock.AssertNumberOfCalls(t, "CpuPlay", 4)
	})

	t.Run("at game end", func(t *testing.T) {
		pMock := new(presenter.MockLiteraturePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(literatureMockOutput)
		gameMock := new(interfaces.MockLiteratureGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(true)

		li := usecase.NewLiteratureInteractor(gameMock, pMock)
		li.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})
}

func TestLiteratureInteractor_GetConfigAndActionLog(t *testing.T) {
	cfg := domain.DefaultLiteratureConfig()
	pMock := new(presenter.MockLiteraturePresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return(`[]`)
	gameMock := new(interfaces.MockLiteratureGame)
	gameMock.On("GetConfig").Return(cfg)

	li := usecase.NewLiteratureInteractor(gameMock, pMock)
	assert.Equal(t, cfg, li.GetConfig())
	assert.Equal(t, `[]`, li.ActionLog())
}

func TestLiteratureInteractor_SnapshotAndRestore(t *testing.T) {
	pMock := new(presenter.MockLiteraturePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(literatureMockOutput)

	g := domain.NewDefaultLiterature()
	g.Reset()
	li := usecase.NewLiteratureInteractor(g, pMock)
	data, err := li.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreLiteratureInteractor(data, pMock)
	assert.NoError(t, err)
	assert.Equal(t, g.GetConfig(), restored.GetConfig())

	_, err = usecase.RestoreLiteratureInteractor([]byte(`{`), pMock)
	assert.Error(t, err)
}
