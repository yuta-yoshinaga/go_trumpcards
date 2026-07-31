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

const killeMockOutput = `{"phase":0}`

// killeSeats builds a seat list whose first entry is the human.
func killeSeats() []*domain.KillePlayer {
	return []*domain.KillePlayer{
		domain.NewKillePlayer(true),
		domain.NewKillePlayer(false),
	}
}

func TestNewKilleInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockKillePresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "KilleInteractor: g must not be nil", func() {
			usecase.NewKilleInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockKilleGame)
		assert.PanicsWithValue(t, "KilleInteractor: gp must not be nil", func() {
			usecase.NewKilleInteractor(gameMock, nil)
		})
	})
}

func TestKilleInteractor_Reset(t *testing.T) {
	pMock := new(presenter.MockKillePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(killeMockOutput)
	gameMock := new(interfaces.MockKilleGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KillePhaseExchange)
	gameMock.On("IsHumanTurn").Return(true)

	ki := usecase.NewKilleInteractor(gameMock, pMock)
	assert.Equal(t, killeMockOutput, ki.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestKilleInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		pMock := new(presenter.MockKillePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(killeMockOutput)
		gameMock := new(interfaces.MockKilleGame)
		cfg := domain.KilleConfig{CpuDifficulty: domain.KilleCpuDifficultyNormal, Stake: 5}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.KillePhaseExchange)
		gameMock.On("IsHumanTurn").Return(true)

		ki := usecase.NewKilleInteractor(gameMock, pMock)
		assert.Equal(t, killeMockOutput, ki.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config never reaches the game", func(t *testing.T) {
		pMock := new(presenter.MockKillePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(killeMockOutput)
		gameMock := new(interfaces.MockKilleGame)

		ki := usecase.NewKilleInteractor(gameMock, pMock)
		assert.Equal(t, killeMockOutput, ki.ResetWithConfig(domain.KilleConfig{Stake: 0}))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
		gameMock.AssertNotCalled(t, "Reset")
	})
}

func TestKilleInteractor_Exchange(t *testing.T) {
	t.Run("passes the seat whose turn it is", func(t *testing.T) {
		pMock := new(presenter.MockKillePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(killeMockOutput)
		gameMock := new(interfaces.MockKilleGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("GetCurrentPlayerIdx").Return(2)
		gameMock.On("Exchange", 2).Return(nil)
		gameMock.On("GetPhase").Return(domain.KillePhaseShowdown)

		ki := usecase.NewKilleInteractor(gameMock, pMock)
		assert.Equal(t, killeMockOutput, ki.Exchange())
		gameMock.AssertCalled(t, "Exchange", 2)
	})

	t.Run("blocked when it is not the human's turn", func(t *testing.T) {
		pMock := new(presenter.MockKillePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(killeMockOutput)
		gameMock := new(interfaces.MockKilleGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ki := usecase.NewKilleInteractor(gameMock, pMock)
		assert.Equal(t, killeMockOutput, ki.Exchange())
		gameMock.AssertNotCalled(t, "Exchange", mock.Anything)
	})

	t.Run("a domain error is presented", func(t *testing.T) {
		wantErr := errors.New("boom")
		pMock := new(presenter.MockKillePresenter)
		pMock.On("Output", mock.Anything, wantErr).Return(killeMockOutput)
		gameMock := new(interfaces.MockKilleGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("GetCurrentPlayerIdx").Return(0)
		gameMock.On("Exchange", 0).Return(wantErr)

		ki := usecase.NewKilleInteractor(gameMock, pMock)
		assert.Equal(t, killeMockOutput, ki.Exchange())
		pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
	})
}

func TestKilleInteractor_Satisfied(t *testing.T) {
	pMock := new(presenter.MockKillePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(killeMockOutput)
	gameMock := new(interfaces.MockKilleGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("GetCurrentPlayerIdx").Return(1)
	gameMock.On("Satisfied", 1).Return(nil)
	gameMock.On("GetPhase").Return(domain.KillePhaseShowdown)

	ki := usecase.NewKilleInteractor(gameMock, pMock)
	assert.Equal(t, killeMockOutput, ki.Satisfied())
	gameMock.AssertCalled(t, "Satisfied", 1)
}

// **CPU の買い戻しは NextRound より先に解決しなければならない。**
// NextRound が買い戻さなかった脱落者を退場させてしまうので、順序が逆だと
// CPU は二度と戻ってこられない。
func TestKilleInteractor_NextRoundResolvesCpuBuyBacksFirst(t *testing.T) {
	pMock := new(presenter.MockKillePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(killeMockOutput)
	seats := killeSeats()
	gameMock := new(interfaces.MockKilleGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPlayers").Return(seats)
	gameMock.On("GetPlayer", 0).Return(seats[0])
	gameMock.On("GetPlayer", 1).Return(seats[1])
	gameMock.On("KilleCpuReenterDecide", 1).Return(true)
	gameMock.On("Reenter", 1).Return(nil)
	gameMock.On("NextRound").Return(nil)
	gameMock.On("GetPhase").Return(domain.KillePhaseExchange)
	gameMock.On("IsHumanTurn").Return(true)

	ki := usecase.NewKilleInteractor(gameMock, pMock)
	assert.Equal(t, killeMockOutput, ki.NextRound())

	// 人間の席には勝手に払わせない。
	gameMock.AssertNotCalled(t, "Reenter", 0)
	gameMock.AssertCalled(t, "Reenter", 1)
	order := []string{}
	for _, c := range gameMock.Calls {
		if c.Method == "Reenter" || c.Method == "NextRound" {
			order = append(order, c.Method)
		}
	}
	assert.Equal(t, []string{"Reenter", "NextRound"}, order)
}

func TestKilleInteractor_NextRoundBlockedAfterGameEnd(t *testing.T) {
	pMock := new(presenter.MockKillePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(killeMockOutput)
	gameMock := new(interfaces.MockKilleGame)
	gameMock.On("GetGameEndFlag").Return(true)

	ki := usecase.NewKilleInteractor(gameMock, pMock)
	assert.Equal(t, killeMockOutput, ki.NextRound())
	gameMock.AssertNotCalled(t, "NextRound")
}

func TestKilleInteractor_Reenter(t *testing.T) {
	t.Run("buys the human back in and deals again", func(t *testing.T) {
		pMock := new(presenter.MockKillePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(killeMockOutput)
		seats := killeSeats()
		gameMock := new(interfaces.MockKilleGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPlayers").Return(seats)
		gameMock.On("GetPlayer", 0).Return(seats[0])
		gameMock.On("GetPlayer", 1).Return(seats[1])
		gameMock.On("Reenter", 0).Return(nil)
		gameMock.On("KilleCpuReenterDecide", 1).Return(false)
		gameMock.On("NextRound").Return(nil)
		gameMock.On("GetPhase").Return(domain.KillePhaseExchange)
		gameMock.On("IsHumanTurn").Return(true)

		ki := usecase.NewKilleInteractor(gameMock, pMock)
		assert.Equal(t, killeMockOutput, ki.Reenter())
		gameMock.AssertCalled(t, "Reenter", 0)
		gameMock.AssertCalled(t, "NextRound")
	})

	t.Run("a failed buy-back does not deal the next round", func(t *testing.T) {
		wantErr := errors.New("already bought back")
		pMock := new(presenter.MockKillePresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(killeMockOutput)
		seats := killeSeats()
		gameMock := new(interfaces.MockKilleGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPlayers").Return(seats)
		gameMock.On("GetPlayer", 0).Return(seats[0])
		gameMock.On("GetPlayer", 1).Return(seats[1])
		gameMock.On("Reenter", 0).Return(wantErr)

		ki := usecase.NewKilleInteractor(gameMock, pMock)
		assert.Equal(t, killeMockOutput, ki.Reenter())
		gameMock.AssertNotCalled(t, "NextRound")
	})
}

// NextRound がエラーを返したらそれを見せる。
func TestKilleInteractor_NextRoundError(t *testing.T) {
	wantErr := errors.New("the round is still in progress")
	pMock := new(presenter.MockKillePresenter)
	pMock.On("Output", mock.Anything, wantErr).Return(killeMockOutput)
	seats := killeSeats()
	gameMock := new(interfaces.MockKilleGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPlayers").Return(seats)
	gameMock.On("GetPlayer", 0).Return(seats[0])
	gameMock.On("GetPlayer", 1).Return(seats[1])
	gameMock.On("KilleCpuReenterDecide", 1).Return(false)
	gameMock.On("NextRound").Return(wantErr)

	ki := usecase.NewKilleInteractor(gameMock, pMock)
	assert.Equal(t, killeMockOutput, ki.NextRound())
	pMock.AssertCalled(t, "Output", mock.Anything, wantErr)
}

func TestKilleInteractor_GetConfigAndActionLog(t *testing.T) {
	cfg := domain.DefaultKilleConfig()
	pMock := new(presenter.MockKillePresenter)
	pMock.On("ActionLogOutput", mock.Anything).Return(`[]`)
	gameMock := new(interfaces.MockKilleGame)
	gameMock.On("GetConfig").Return(cfg)

	ki := usecase.NewKilleInteractor(gameMock, pMock)
	assert.Equal(t, cfg, ki.GetConfig())
	assert.Equal(t, `[]`, ki.ActionLog())
}

// **CPU ループは人間の手番で必ず止まる。**止まらないと人間が操作できない。
func TestKilleInteractor_RunCpuTurnsStopsAtHuman(t *testing.T) {
	pMock := new(presenter.MockKillePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(killeMockOutput)
	gameMock := new(interfaces.MockKilleGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.KillePhaseExchange)
	gameMock.On("IsHumanTurn").Return(false).Times(3)
	gameMock.On("IsHumanTurn").Return(true)
	gameMock.On("CpuPlay").Return()

	ki := usecase.NewKilleInteractor(gameMock, pMock)
	ki.Reset()
	gameMock.AssertNumberOfCalls(t, "CpuPlay", 3)
}

func TestKilleInteractor_SnapshotAndRestore(t *testing.T) {
	pMock := new(presenter.MockKillePresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(killeMockOutput)

	g := domain.NewDefaultKille()
	g.Reset()
	ki := usecase.NewKilleInteractor(g, pMock)
	data, err := ki.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreKilleInteractor(data, pMock)
	assert.NoError(t, err)
	assert.Equal(t, g.GetConfig(), restored.GetConfig())

	_, err = usecase.RestoreKilleInteractor([]byte(`{`), pMock)
	assert.Error(t, err)
}
