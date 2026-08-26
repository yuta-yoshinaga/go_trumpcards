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

func TestNewCoincheInteractor_NilGuards(t *testing.T) {
	bpMock := new(presenter.MockCoinchePresenter)

	t.Run("panics when b is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "CoincheInteractor: b must not be nil", func() {
			usecase.NewCoincheInteractor(nil, bpMock)
		})
	})

	t.Run("panics when bp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockCoincheGame)
		assert.PanicsWithValue(t, "CoincheInteractor: bp must not be nil", func() {
			usecase.NewCoincheInteractor(gameMock, nil)
		})
	})
}

func setupCoincheInteractorMocks(phase domain.CoinchePhase) (*interfaces.MockCoincheGame, *presenter.MockCoinchePresenter) {
	bpMock := new(presenter.MockCoinchePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"phase":0}`)
	gameMock := new(interfaces.MockCoincheGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(phase)
	gameMock.On("IsHumanBidTurn").Return(true)
	gameMock.On("IsHumanTurn").Return(true)
	return gameMock, bpMock
}

func TestCoincheInteractor_Reset(t *testing.T) {
	gameMock, bpMock := setupCoincheInteractorMocks(domain.CoinchePhaseBid)
	gameMock.On("Reset").Return()

	bi := usecase.NewCoincheInteractor(gameMock, bpMock)
	result := bi.Reset()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "Reset")
}

func TestCoincheInteractor_ResetWithConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		gameMock, bpMock := setupCoincheInteractorMocks(domain.CoinchePhaseBid)
		cfg := domain.CoincheConfig{
			CpuDifficulty:        domain.CoincheCpuDifficultyHard,
			TargetScore:          500,
			DixDeDer:             10,
			EnableBeloteRebelote: true,
		}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()

		bi := usecase.NewCoincheInteractor(gameMock, bpMock)
		result := bi.ResetWithConfig(cfg)
		assert.Contains(t, result, "phase")
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns presenter output", func(t *testing.T) {
		gameMock, bpMock := setupCoincheInteractorMocks(domain.CoinchePhaseBid)
		invalidCfg := domain.CoincheConfig{
			CpuDifficulty: 99,
			TargetScore:   1000,
			DixDeDer:      10,
		}

		bi := usecase.NewCoincheInteractor(gameMock, bpMock)
		result := bi.ResetWithConfig(invalidCfg)
		assert.Contains(t, result, "phase")
	})
}

func TestCoincheInteractor_Bid(t *testing.T) {
	t.Run("forwards both the points and the suit", func(t *testing.T) {
		gameMock, bpMock := setupCoincheInteractorMocks(domain.CoinchePhaseBid)
		gameMock.On("PlayerBid", 110, domain.CardDesignSpade).Return(nil)

		bi := usecase.NewCoincheInteractor(gameMock, bpMock)
		result := bi.Bid(110, domain.CardDesignSpade)
		assert.Contains(t, result, "phase")
		// **点とスートは 2 つで 1 つの宣言。** 片方でも落ちると、宣言した
		// 契約と盤面の切り札が食い違う。
		gameMock.AssertCalled(t, "PlayerBid", 110, domain.CardDesignSpade)
	})

	t.Run("surfaces the domain error", func(t *testing.T) {
		gameMock, bpMock := setupCoincheInteractorMocks(domain.CoinchePhaseBid)
		gameMock.On("PlayerBid", 85, domain.CardDesignSpade).Return(errors.New("not a contract value"))

		bi := usecase.NewCoincheInteractor(gameMock, bpMock)
		assert.Contains(t, bi.Bid(85, domain.CardDesignSpade), "phase")
	})

	t.Run("blocked once the game has ended", func(t *testing.T) {
		bpMock := new(presenter.MockCoinchePresenter)
		bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
		gameMock := new(interfaces.MockCoincheGame)
		gameMock.On("GetGameEndFlag").Return(true)

		bi := usecase.NewCoincheInteractor(gameMock, bpMock)
		assert.Contains(t, bi.Bid(110, domain.CardDesignSpade), "gameEnd")
		gameMock.AssertNotCalled(t, "PlayerBid", mock.Anything, mock.Anything)
	})
}

func TestCoincheInteractor_Pass(t *testing.T) {
	gameMock, bpMock := setupCoincheInteractorMocks(domain.CoinchePhaseBid)
	gameMock.On("PlayerPassBid").Return(nil)

	bi := usecase.NewCoincheInteractor(gameMock, bpMock)
	assert.Contains(t, bi.Pass(), "phase")
	gameMock.AssertCalled(t, "PlayerPassBid")
}

// 倍化の 3 操作はどれも同じ後始末 (CPU を回して出力) を要る。
func TestCoincheInteractor_DoublingActions(t *testing.T) {
	for _, tc := range []struct {
		name   string
		method string
		call   func(usecase.CoincheInteractorIF) string
	}{
		{"coinche", "PlayerCoinche", func(bi usecase.CoincheInteractorIF) string { return bi.Coinche() }},
		{"surcoinche", "PlayerSurcoinche", func(bi usecase.CoincheInteractorIF) string { return bi.Surcoinche() }},
		{"decline", "PlayerDeclineDouble", func(bi usecase.CoincheInteractorIF) string { return bi.DeclineDouble() }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gameMock, bpMock := setupCoincheInteractorMocks(domain.CoinchePhaseDouble)
			gameMock.On(tc.method).Return(nil)

			bi := usecase.NewCoincheInteractor(gameMock, bpMock)
			assert.Contains(t, tc.call(bi), "phase")
			gameMock.AssertCalled(t, tc.method)
		})
	}
}

func TestCoincheInteractor_Play(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		gameMock, bpMock := setupCoincheInteractorMocks(domain.CoinchePhasePlay)
		gameMock.On("PlayerPlay", 0).Return(nil)

		bi := usecase.NewCoincheInteractor(gameMock, bpMock)
		result := bi.Play(0)
		assert.Contains(t, result, "phase")
	})

	t.Run("error", func(t *testing.T) {
		gameMock, bpMock := setupCoincheInteractorMocks(domain.CoinchePhasePlay)
		gameMock.On("PlayerPlay", 0).Return(errors.New("must follow"))

		bi := usecase.NewCoincheInteractor(gameMock, bpMock)
		result := bi.Play(0)
		assert.Contains(t, result, "phase")
	})

	t.Run("not human turn", func(t *testing.T) {
		bpMock := new(presenter.MockCoinchePresenter)
		bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"notHuman":true}`)
		gameMock := new(interfaces.MockCoincheGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		bi := usecase.NewCoincheInteractor(gameMock, bpMock)
		result := bi.Play(0)
		assert.Contains(t, result, "notHuman")
	})
}

func TestCoincheInteractor_NextTrick(t *testing.T) {
	gameMock, bpMock := setupCoincheInteractorMocks(domain.CoinchePhasePlay)
	gameMock.On("ResolveTrick").Return()
	gameMock.On("NextTrick").Return()

	bi := usecase.NewCoincheInteractor(gameMock, bpMock)
	result := bi.NextTrick()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "ResolveTrick")
	gameMock.AssertCalled(t, "NextTrick")
}

func TestCoincheInteractor_NextTrick_GameEnd(t *testing.T) {
	bpMock := new(presenter.MockCoinchePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
	gameMock := new(interfaces.MockCoincheGame)
	gameMock.On("ResolveTrick").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	bi := usecase.NewCoincheInteractor(gameMock, bpMock)
	result := bi.NextTrick()
	assert.Contains(t, result, "gameEnd")
}

func TestCoincheInteractor_NextRound(t *testing.T) {
	gameMock, bpMock := setupCoincheInteractorMocks(domain.CoinchePhaseBid)
	gameMock.On("ScoreRound").Return()
	gameMock.On("NextRound").Return()

	bi := usecase.NewCoincheInteractor(gameMock, bpMock)
	result := bi.NextRound()
	assert.Contains(t, result, "phase")
	gameMock.AssertCalled(t, "ScoreRound")
	gameMock.AssertCalled(t, "NextRound")
}

func TestCoincheInteractor_NextRound_GameEnd(t *testing.T) {
	bpMock := new(presenter.MockCoinchePresenter)
	bpMock.On("Output", mock.Anything, mock.Anything).Return(`{"gameEnd":true}`)
	gameMock := new(interfaces.MockCoincheGame)
	gameMock.On("ScoreRound").Return()
	gameMock.On("GetGameEndFlag").Return(true)

	bi := usecase.NewCoincheInteractor(gameMock, bpMock)
	result := bi.NextRound()
	assert.Contains(t, result, "gameEnd")
}

func TestCoincheInteractor_GetConfig(t *testing.T) {
	gameMock, bpMock := setupCoincheInteractorMocks(domain.CoinchePhaseBid)
	cfg := domain.DefaultCoincheConfig()
	gameMock.On("GetConfig").Return(cfg)

	bi := usecase.NewCoincheInteractor(gameMock, bpMock)
	assert.Equal(t, cfg, bi.GetConfig())
}

func TestCoincheInteractor_Hint(t *testing.T) {
	gameMock, bpMock := setupCoincheInteractorMocks(domain.CoinchePhaseBid)
	bpMock.On("HintOutput", gameMock).Return(`{"hint":"x"}`)

	bi := usecase.NewCoincheInteractor(gameMock, bpMock)
	assert.Contains(t, bi.Hint(), "hint")
}

func TestCoincheInteractor_ActionLog(t *testing.T) {
	gameMock, bpMock := setupCoincheInteractorMocks(domain.CoinchePhaseBid)
	bpMock.On("ActionLogOutput", gameMock).Return(`[{"a":"b"}]`)

	bi := usecase.NewCoincheInteractor(gameMock, bpMock)
	assert.Contains(t, bi.ActionLog(), "a")
}

func TestRestoreCoincheInteractor(t *testing.T) {
	players := []*domain.CoinchePlayer{
		domain.NewCoinchePlayer(true, 0),
		domain.NewCoinchePlayer(false, 1),
		domain.NewCoinchePlayer(false, 0),
		domain.NewCoinchePlayer(false, 1),
	}
	b := domain.NewCoinche(domain.NewTrumpCards32(), players, domain.DefaultCoincheConfig())
	bpMock := new(presenter.MockCoinchePresenter)
	bi := usecase.NewCoincheInteractor(b, bpMock)

	data, err := bi.Snapshot()
	assert.NoError(t, err)

	restored, err := usecase.RestoreCoincheInteractor(data, bpMock)
	assert.NoError(t, err)
	assert.NotNil(t, restored)
}

func TestRestoreCoincheInteractor_InvalidJSON(t *testing.T) {
	bpMock := new(presenter.MockCoinchePresenter)
	_, err := usecase.RestoreCoincheInteractor([]byte("invalid"), bpMock)
	assert.Error(t, err)
}
