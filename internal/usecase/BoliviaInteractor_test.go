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

func TestNewBoliviaInteractor_NilGuards(t *testing.T) {
	pMock := new(presenter.MockBoliviaPresenter)

	t.Run("panics when g is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "BoliviaInteractor: g must not be nil", func() {
			usecase.NewBoliviaInteractor(nil, pMock)
		})
	})

	t.Run("panics when gp is nil", func(t *testing.T) {
		gameMock := new(interfaces.MockBoliviaGame)
		assert.PanicsWithValue(t, "BoliviaInteractor: gp must not be nil", func() {
			usecase.NewBoliviaInteractor(gameMock, nil)
		})
	})
}

func TestBoliviaInteractor_Reset(t *testing.T) {
	mockOutput := `{"phase":0}`
	pMock := new(presenter.MockBoliviaPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBoliviaGame)
	gameMock.On("Reset").Return()
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("GetPhase").Return(domain.BoliviaPhaseDraw)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewBoliviaInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.Reset())
	gameMock.AssertCalled(t, "Reset")
}

func TestBoliviaInteractor_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("valid config sets config then resets", func(t *testing.T) {
		pMock := new(presenter.MockBoliviaPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBoliviaGame)
		cfg := domain.BoliviaConfig{CpuDifficulty: domain.BoliviaCpuDifficultyHard, PointLimit: 5000}
		gameMock.On("SetConfig", cfg).Return()
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BoliviaPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBoliviaInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.ResetWithConfig(cfg))
		gameMock.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config returns error without resetting", func(t *testing.T) {
		pMock := new(presenter.MockBoliviaPresenter)
		gameMock := new(interfaces.MockBoliviaGame)
		pMock.On("Output", gameMock, mock.MatchedBy(func(err error) bool { return err != nil })).Return("validation error")

		ci := usecase.NewBoliviaInteractor(gameMock, pMock)
		cfg := domain.BoliviaConfig{CpuDifficulty: domain.BoliviaCpuDifficulty(-1), PointLimit: 100}
		assert.Equal(t, "validation error", ci.ResetWithConfig(cfg))
		gameMock.AssertNotCalled(t, "SetConfig", mock.Anything)
	})
}

func TestBoliviaInteractor_DrawFromStock(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output without drawing", func(t *testing.T) {
		pMock := new(presenter.MockBoliviaPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBoliviaGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewBoliviaInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromStock())
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})

	t.Run("not human turn returns output without drawing", func(t *testing.T) {
		pMock := new(presenter.MockBoliviaPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBoliviaGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewBoliviaInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromStock())
		gameMock.AssertNotCalled(t, "PlayerDrawFromStock")
	})

	t.Run("draw error is returned", func(t *testing.T) {
		drawErr := errors.New("draw error")
		pMock := new(presenter.MockBoliviaPresenter)
		pMock.On("Output", mock.Anything, drawErr).Return(mockOutput)
		gameMock := new(interfaces.MockBoliviaGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromStock").Return(drawErr)

		ci := usecase.NewBoliviaInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromStock())
	})

	t.Run("valid draw runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockBoliviaPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBoliviaGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerDrawFromStock").Return(nil)
		gameMock.On("GetPhase").Return(domain.BoliviaPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBoliviaInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromStock())
		gameMock.AssertCalled(t, "PlayerDrawFromStock")
	})
}

func TestBoliviaInteractor_DrawFromDiscard(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("error is returned", func(t *testing.T) {
		drawErr := errors.New("draw error")
		pMock := new(presenter.MockBoliviaPresenter)
		pMock.On("Output", mock.Anything, drawErr).Return(mockOutput)
		gameMock := new(interfaces.MockBoliviaGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerDrawFromDiscard", []int{0, 1}).Return(drawErr)

		ci := usecase.NewBoliviaInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromDiscard([]int{0, 1}))
	})

	t.Run("valid draw runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockBoliviaPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBoliviaGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerDrawFromDiscard", []int{0, 1}).Return(nil)
		gameMock.On("GetPhase").Return(domain.BoliviaPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBoliviaInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.DrawFromDiscard([]int{0, 1}))
		gameMock.AssertCalled(t, "PlayerDrawFromDiscard", []int{0, 1})
	})
}

func TestBoliviaInteractor_Meld(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("error is returned", func(t *testing.T) {
		meldErr := errors.New("meld error")
		pMock := new(presenter.MockBoliviaPresenter)
		pMock.On("Output", mock.Anything, meldErr).Return(mockOutput)
		gameMock := new(interfaces.MockBoliviaGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true)
		gameMock.On("PlayerMeld", [][]int{{0, 1, 2}}).Return(meldErr)

		ci := usecase.NewBoliviaInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Meld([][]int{{0, 1, 2}}))
	})

	t.Run("valid meld runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockBoliviaPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBoliviaGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerMeld", [][]int{{0, 1, 2}}).Return(nil)
		gameMock.On("GetPhase").Return(domain.BoliviaPhaseMeld)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBoliviaInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Meld([][]int{{0, 1, 2}}))
		gameMock.AssertCalled(t, "PlayerMeld", [][]int{{0, 1, 2}})
	})
}

func TestBoliviaInteractor_SkipMeld(t *testing.T) {
	mockOutput := `{"phase":0}`
	pMock := new(presenter.MockBoliviaPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBoliviaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true).Once()
	gameMock.On("PlayerSkipMeld").Return(nil)
	gameMock.On("GetPhase").Return(domain.BoliviaPhaseMeld)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewBoliviaInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.SkipMeld())
	gameMock.AssertCalled(t, "PlayerSkipMeld")
}

func TestBoliviaInteractor_Discard(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("not human turn returns output", func(t *testing.T) {
		pMock := new(presenter.MockBoliviaPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBoliviaGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(false)

		ci := usecase.NewBoliviaInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Discard(0))
	})

	t.Run("valid discard runs CPU turns", func(t *testing.T) {
		pMock := new(presenter.MockBoliviaPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBoliviaGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("IsHumanTurn").Return(true).Once()
		gameMock.On("PlayerDiscard", 3).Return(nil)
		gameMock.On("GetPhase").Return(domain.BoliviaPhaseDiscard)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBoliviaInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.Discard(3))
		gameMock.AssertCalled(t, "PlayerDiscard", 3)
	})
}

func TestBoliviaInteractor_GoOut(t *testing.T) {
	mockOutput := `{"phase":0}`
	pMock := new(presenter.MockBoliviaPresenter)
	pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
	gameMock := new(interfaces.MockBoliviaGame)
	gameMock.On("GetGameEndFlag").Return(false)
	gameMock.On("IsHumanTurn").Return(true).Once()
	gameMock.On("PlayerGoOut").Return(nil)
	gameMock.On("GetPhase").Return(domain.BoliviaPhaseDiscard)
	gameMock.On("IsHumanTurn").Return(true)

	ci := usecase.NewBoliviaInteractor(gameMock, pMock)
	assert.Equal(t, mockOutput, ci.GoOut())
	gameMock.AssertCalled(t, "PlayerGoOut")
}

func TestBoliviaInteractor_NextRound(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("game ended returns output", func(t *testing.T) {
		pMock := new(presenter.MockBoliviaPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBoliviaGame)
		gameMock.On("GetGameEndFlag").Return(true)

		ci := usecase.NewBoliviaInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertNotCalled(t, "NextRound")
	})

	t.Run("valid next round", func(t *testing.T) {
		pMock := new(presenter.MockBoliviaPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBoliviaGame)
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("NextRound").Return()
		gameMock.On("GetPhase").Return(domain.BoliviaPhaseDraw)
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBoliviaInteractor(gameMock, pMock)
		assert.Equal(t, mockOutput, ci.NextRound())
		gameMock.AssertCalled(t, "NextRound")
	})
}

func TestBoliviaInteractor_GetConfigAndActionLog(t *testing.T) {
	pMock := new(presenter.MockBoliviaPresenter)
	gameMock := new(interfaces.MockBoliviaGame)
	expected := domain.BoliviaConfig{CpuDifficulty: domain.BoliviaCpuDifficultyHard, PointLimit: 5000}
	gameMock.On("GetConfig").Return(expected)
	pMock.On("ActionLogOutput", gameMock).Return(`{"entries":[]}`)

	ci := usecase.NewBoliviaInteractor(gameMock, pMock)
	assert.Equal(t, expected, ci.GetConfig())
	assert.Equal(t, `{"entries":[]}`, ci.ActionLog())
}

func TestBoliviaInteractor_RunCpuTurns(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("stops when phase is RoundEnd", func(t *testing.T) {
		pMock := new(presenter.MockBoliviaPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBoliviaGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BoliviaPhaseRoundEnd)

		ci := usecase.NewBoliviaInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertNotCalled(t, "CpuPlay")
	})

	t.Run("CPU plays then stops at human turn", func(t *testing.T) {
		pMock := new(presenter.MockBoliviaPresenter)
		pMock.On("Output", mock.Anything, mock.Anything).Return(mockOutput)
		gameMock := new(interfaces.MockBoliviaGame)
		gameMock.On("Reset").Return()
		gameMock.On("GetGameEndFlag").Return(false)
		gameMock.On("GetPhase").Return(domain.BoliviaPhaseDraw)
		gameMock.On("IsHumanTurn").Return(false).Once()
		gameMock.On("CpuPlay").Return()
		gameMock.On("IsHumanTurn").Return(true)

		ci := usecase.NewBoliviaInteractor(gameMock, pMock)
		ci.Reset()
		gameMock.AssertCalled(t, "CpuPlay")
	})
}

// boliviaPassThrough は「呼ばれたか」だけを見る素通しのプレゼンター。
type boliviaPassThrough struct{}

func (boliviaPassThrough) Output(_ interfaces.BoliviaGame, lastErr error) string {
	if lastErr != nil {
		return "err:" + lastErr.Error()
	}
	return "ok"
}
func (boliviaPassThrough) HintOutput(_ interfaces.BoliviaGame) string      { return "hint" }
func (boliviaPassThrough) ActionLogOutput(_ interfaces.BoliviaGame) string { return "log" }

// **Worker が毎リクエスト通る道を実際に通す。**
//
// ここが 0% のあいだに、復元でワイルドメルドがセットに化ける不具合を出荷
// しかけた (レビュー指摘)。ドメインの往復テストだけでは、KV から戻す
// `RestoreBoliviaInteractor` 経由で同じことが起きるかを誰も見ていない。
func TestRestoreBoliviaInteractor_KeepsTheBoliviaAndKeepsPlaying(t *testing.T) {
	g := domain.NewDefaultBolivia()
	g.Reset()

	// 人間の席にボリビアとエスカレラを置く。
	seven := func(d, from int) []*domain.Card {
		out := make([]*domain.Card, 0, 7)
		for i := 0; i < 7; i++ {
			out = append(out, domain.NewCard(d, from+i, true))
		}
		return out
	}
	wild := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, true),
		domain.NewCard(domain.CardDesignHeart, 2, true),
		domain.NewCard(domain.CardDesignClover, 2, true),
		domain.NewCard(domain.CardDesignDiamond, 2, true),
		domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, true),
		domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, true),
		domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, true),
	}
	g.GetPlayer(0).SetMelds([]*domain.BoliviaMeld{
		{Kind: domain.BoliviaMeldWild, Cards: wild},
		{Kind: domain.BoliviaMeldEscalera, IsNatural: true, Cards: seven(domain.CardDesignHeart, 4)},
	})

	bi := usecase.NewBoliviaInteractor(g, boliviaPassThrough{})
	data, err := bi.Snapshot()
	assert.NoError(t, err)
	assert.Greater(t, len(data), 2, "snapshot が `{}` -- MarshalJSON が無い")

	restored, err := usecase.RestoreBoliviaInteractor(data, boliviaPassThrough{})
	assert.NoError(t, err)
	rp := restored.Game.GetPlayer(0)

	melds := rp.GetMelds()
	assert.Len(t, melds, 2)
	assert.Equal(t, domain.BoliviaMeldWild, melds[0].Kind, "復元でボリビアがセットに化けている")
	assert.True(t, melds[0].IsBoliviaCanasta())
	assert.True(t, rp.HasBolivia(), "復元後に HasBolivia が落ちている")
	assert.True(t, rp.HasEscalera())

	// 盤の続きも保たれていること。
	assert.Equal(t, g.GetRoundNumber(), restored.Game.GetRoundNumber())
	assert.Equal(t, g.GetPhase(), restored.Game.GetPhase())
	// 復元した盤で指し続けられる。
	assert.NotEmpty(t, restored.ActionLog())
	// 復元した盤は元の盤と同じフェーズ規則で動く ── ドローは通り、
	// フェーズ違いのメルドスキップは断られる。
	assert.Equal(t, domain.BoliviaPhaseDraw, restored.Game.GetPhase())
	assert.Contains(t, restored.SkipMeld(), "err:")
	assert.Equal(t, "ok", restored.DrawFromStock())
	assert.Equal(t, domain.BoliviaPhaseMeld, restored.Game.GetPhase())
}

// 壊れた JSON は素直に失敗すること。
func TestRestoreBoliviaInteractor_RejectsGarbage(t *testing.T) {
	_, err := usecase.RestoreBoliviaInteractor([]byte("not json"), boliviaPassThrough{})
	assert.Error(t, err)
}
