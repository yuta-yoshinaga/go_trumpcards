//go:build test

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

func newTestDurak() *domain.Durak {
	players := []*domain.DurakPlayer{
		domain.NewDurakPlayer(true),
		domain.NewDurakPlayer(false),
		domain.NewDurakPlayer(false),
		domain.NewDurakPlayer(false),
	}
	return domain.NewDurak(domain.NewTrumpCardsShortDeck(), players)
}

func TestNewDurakInteractor_NilGuards(t *testing.T) {
	dpMock := new(presenter.MockDurakPresenter)
	t.Run("panics when d is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "DurakInteractor: d must not be nil", func() {
			usecase.NewDurakInteractor(nil, dpMock)
		})
	})
	t.Run("panics when dp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "DurakInteractor: dp must not be nil", func() {
			usecase.NewDurakInteractor(newTestDurak(), nil)
		})
	})
}

func TestDurakInteractor_Methods(t *testing.T) {
	mockOutput := `{"test":"ok"}`
	dpMock := new(presenter.MockDurakPresenter)
	dpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	di := usecase.NewDurakInteractor(newTestDurak(), dpMock)

	t.Run("Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, di.Reset())
	})

	t.Run("Attack", func(t *testing.T) {
		assert.Equal(t, mockOutput, di.Attack(0))
	})

	t.Run("Defend", func(t *testing.T) {
		assert.Equal(t, mockOutput, di.Defend(0, 0))
	})

	t.Run("Pass", func(t *testing.T) {
		assert.Equal(t, mockOutput, di.Pass())
	})

	t.Run("TakeCards", func(t *testing.T) {
		assert.Equal(t, mockOutput, di.TakeCards())
	})

	t.Run("Transfer", func(t *testing.T) {
		assert.Equal(t, mockOutput, di.Transfer(0))
	})

	t.Run("ResetWithConfig", func(t *testing.T) {
		config := domain.DefaultDurakConfig()
		assert.Equal(t, mockOutput, di.ResetWithConfig(config))
	})

	t.Run("Sort", func(t *testing.T) {
		assert.Equal(t, mockOutput, di.Sort(domain.DurakSortBySuit))
	})
}

// **CUI に hint コマンドすら無かった (#4740)。**interactor が presenter へ
// 委譲する経路を踏む。
func TestDurakInteractor_Hint(t *testing.T) {
	dpMock := new(presenter.MockDurakPresenter)
	gameMock := new(interfaces.MockDurakGame)
	dpMock.On("HintOutput", gameMock).Return(`{"hint":{"reason":"attack_weakest"}}`)

	di := usecase.NewDurakInteractor(gameMock, dpMock)
	assert.Equal(t, `{"hint":{"reason":"attack_weakest"}}`, di.Hint())
	dpMock.AssertExpectations(t)
}

func TestDurakInteractor_ActionLog(t *testing.T) {
	dpMock := new(presenter.MockDurakPresenter)
	gameMock := new(interfaces.MockDurakGame)
	dpMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)

	di := usecase.NewDurakInteractor(gameMock, dpMock)
	result := di.ActionLog()
	assert.Equal(t, `{"entries":[]}`, result)
	dpMock.AssertExpectations(t)
}

func TestDurakInteractor_MockGame(t *testing.T) {
	mockOutput := `{"test":"ok"}`
	dpMock := new(presenter.MockDurakPresenter)
	dpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockDurakGame)

	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(true)
	gameMock.On("IsHumanTurn").Return(true)

	di := usecase.NewDurakInteractor(gameMock, dpMock)

	t.Run("Reset with mock", func(t *testing.T) {
		assert.Equal(t, mockOutput, di.Reset())
		gameMock.AssertCalled(t, "Reset")
	})

	t.Run("Attack with mock", func(t *testing.T) {
		gameMock.On("PlayerAttack", 0).Return(nil)
		assert.Equal(t, mockOutput, di.Attack(0))
	})

	t.Run("Defend with mock", func(t *testing.T) {
		gameMock.On("PlayerDefend", 0, 0).Return(nil)
		assert.Equal(t, mockOutput, di.Defend(0, 0))
	})

	t.Run("Pass with mock", func(t *testing.T) {
		gameMock.On("PlayerPass").Return(nil)
		assert.Equal(t, mockOutput, di.Pass())
	})

	t.Run("TakeCards with mock", func(t *testing.T) {
		gameMock.On("PlayerTakeCards").Return(nil)
		assert.Equal(t, mockOutput, di.TakeCards())
	})

	t.Run("Transfer with mock", func(t *testing.T) {
		// この盤は GetGameEndFlag が true なので guardNotPlayable が止める。
		// **ドメインまで届かないこと**が確かめたいこと —— 終わった局に
		// 転送が通ると、他の行動と挙動が食い違う。
		gameMock.On("PlayerTransfer", mock.Anything).Return(nil)
		assert.Equal(t, mockOutput, di.Transfer(2))
		gameMock.AssertNotCalled(t, "PlayerTransfer", mock.Anything)
	})

	t.Run("GetConfig with mock", func(t *testing.T) {
		gameMock.On("GetConfig").Return(domain.DefaultDurakConfig())
		cfg := di.GetConfig()
		assert.Equal(t, domain.DurakPlayerCntDefault, cfg.PlayerCount)
	})
}

func TestDurakInteractor_Snapshot(t *testing.T) {
	dpMock := new(presenter.MockDurakPresenter)
	dpMock.On("Output", mock.Anything, mock.Anything).Return(`{}`)
	game := newTestDurak()
	di := usecase.NewDurakInteractor(game, dpMock)
	di.Reset()

	data, err := di.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestRestoreDurakInteractor(t *testing.T) {
	dpMock := new(presenter.MockDurakPresenter)
	dpMock.On("Output", mock.Anything, mock.Anything).Return(`{}`)
	game := newTestDurak()
	di := usecase.NewDurakInteractor(game, dpMock)
	di.Reset()

	data, err := di.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreDurakInteractor(data, dpMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}
