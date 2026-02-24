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

func newTestDaifugo() *domain.Daifugo {
	config := domain.DefaultDaifugoConfig()
	players := []*domain.DaifugoPlayer{
		domain.NewDaifugoPlayer(true),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
		domain.NewDaifugoPlayer(false),
	}
	return domain.NewDaifugo(domain.NewTrumpCards(config.JokerCount), players, config)
}

func TestNewDaifugoInteractor_NilGuards(t *testing.T) {
	dgpMock := new(presenter.MockDaifugoPresenter)
	t.Run("panics when dg is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "DaifugoInteractor: dg must not be nil", func() {
			usecase.NewDaifugoInteractor(nil, dgpMock)
		})
	})
	t.Run("panics when dgp is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "DaifugoInteractor: dgp must not be nil", func() {
			usecase.NewDaifugoInteractor(newTestDaifugo(), nil)
		})
	})
}

func TestDaifugoInteractor_Method(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableCards":[],"lastPlayPlayerIdx":-1,"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
	dgpMock := new(presenter.MockDaifugoPresenter)
	dgpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	tdi := usecase.NewDaifugoInteractor(newTestDaifugo(), dgpMock)

	t.Run("success Reset", func(t *testing.T) {
		assert.Equal(t, mockOutput, tdi.Reset())
	})

	t.Run("success Play with pass (empty indices)", func(t *testing.T) {
		assert.Equal(t, mockOutput, tdi.Play([]int{}))
	})

	t.Run("success Play with indices", func(t *testing.T) {
		assert.Equal(t, mockOutput, tdi.Play([]int{0}))
	})
}

func TestDaifugoInteractor_MockGame(t *testing.T) {
	mockOutput := `{"players":[]}`
	dgpMock := new(presenter.MockDaifugoPresenter)
	dgpMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockDaifugoGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()
	gameMock.On("PlayerPlay", mock.Anything).Return(nil)

	di := usecase.NewDaifugoInteractor(gameMock, dgpMock)

	t.Run("Reset calls game.Reset", func(t *testing.T) {
		result := di.Reset()
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "Reset")
	})
	t.Run("Play calls game.PlayerPlay when human turn", func(t *testing.T) {
		result := di.Play([]int{0})
		assert.Equal(t, mockOutput, result)
		gameMock.AssertCalled(t, "PlayerPlay", []int{0})
	})
}
