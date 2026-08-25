//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// contPassThrough は「呼ばれたか」だけを見る素通しのプレゼンター。
type contPassThrough struct{}

func (contPassThrough) Output(_ interfaces.ContinentalRummyGame, lastErr error) string {
	if lastErr != nil {
		return "err:" + lastErr.Error()
	}
	return "ok"
}
func (contPassThrough) HintOutput(_ interfaces.ContinentalRummyGame) string      { return "hint" }
func (contPassThrough) ActionLogOutput(_ interfaces.ContinentalRummyGame) string { return "log" }

func newContReal() (*usecase.ContinentalRummyInteractor, *domain.ContinentalRummy) {
	g := domain.NewDefaultContinentalRummy()
	return usecase.NewContinentalRummyInteractor(g, contPassThrough{}), g
}

// contAtHumanDraw は「人間が引く番」まで進んだ盤を返す。
func contAtHumanDraw(t *testing.T) (*usecase.ContinentalRummyInteractor, *domain.ContinentalRummy) {
	t.Helper()
	for try := 0; try < 50; try++ {
		ci, g := newContReal()
		ci.Reset()
		if g.GetPhase() == domain.ContinentalRummyPhaseDraw && g.IsHumanTurn() {
			return ci, g
		}
	}
	t.Fatal("50 回配っても人間の引く番にならなかった -- 前提が崩れている")
	return nil, nil
}

func TestNewContinentalRummyInteractor_NilGuards(t *testing.T) {
	cp := new(presenter.MockContinentalRummyPresenter)
	assert.PanicsWithValue(t, "ContinentalRummyInteractor: g must not be nil", func() {
		usecase.NewContinentalRummyInteractor(nil, cp)
	})
	assert.PanicsWithValue(t, "ContinentalRummyInteractor: cp must not be nil", func() {
		usecase.NewContinentalRummyInteractor(new(interfaces.MockContinentalRummyGame), nil)
	})
}

// **Reset は人間の番まで進める。** 進めないと最初の手番で固まる。
func TestContinentalRummyInteractor_ResetReachesTheHumanTurn(t *testing.T) {
	ci, g := newContReal()
	assert.Equal(t, "ok", ci.Reset())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetRoundNumber())
	if g.GetPhase() == domain.ContinentalRummyPhaseRoundEnd {
		return // CPU が Reset のうちに上がった配り
	}
	assert.True(t, g.IsHumanTurn(), "人間の番に到達していない")
	assert.Equal(t, ContinentalRummyHandSizeForTest, g.GetPlayer(domain.ContinentalRummyHumanIdx).GetCardsSize())
}

// ContinentalRummyHandSizeForTest は配布枚数 (テストからの参照用)。
const ContinentalRummyHandSizeForTest = 15

// **山と捨て札は別の入口。** 片方しか効いていないと、既定のまま届いた要求が
// 黙ってもう片方になる。
func TestContinentalRummyInteractor_DrawStockAndDrawDiscardAreSeparate(t *testing.T) {
	t.Run("stock adds a card and moves to the discard phase", func(t *testing.T) {
		ci, g := contAtHumanDraw(t)
		before := g.GetPlayer(domain.ContinentalRummyHumanIdx).GetCardsSize()
		stock := g.GetStockCount()
		assert.Equal(t, "ok", ci.DrawStock())
		assert.Equal(t, before+1, g.GetPlayer(domain.ContinentalRummyHumanIdx).GetCardsSize())
		assert.Equal(t, stock-1, g.GetStockCount())
		assert.Equal(t, domain.ContinentalRummyPhaseDiscard, g.GetPhase())
	})

	t.Run("take lifts the discard top and leaves the stock alone", func(t *testing.T) {
		ci, g := contAtHumanDraw(t)
		top := g.GetDiscardTop()
		require.NotNil(t, top)
		stock := g.GetStockCount()
		assert.Equal(t, "ok", ci.DrawDiscard())
		assert.Equal(t, stock, g.GetStockCount(), "捨て札を取ったのに山が減っている")
		assert.NotEqual(t, top, g.GetDiscardTop())
		assert.Equal(t, domain.ContinentalRummyPhaseDiscard, g.GetPhase())
	})
}

func TestContinentalRummyInteractor_DiscardPassesTheTurn(t *testing.T) {
	ci, g := contAtHumanDraw(t)
	require.Equal(t, "ok", ci.DrawStock())
	before := g.GetPlayer(domain.ContinentalRummyHumanIdx).GetCardsSize()
	assert.Equal(t, "ok", ci.Discard(0))
	assert.Equal(t, before-1, g.GetPlayer(domain.ContinentalRummyHumanIdx).GetCardsSize())

	// 出す番でないのに捨てられない。
	assert.Contains(t, ci.Discard(0), "err:")
}

// **形にならない 15 枚では上がれない。**
func TestContinentalRummyInteractor_GoOutRefusesAHandThatIsNotALayout(t *testing.T) {
	ci, g := contAtHumanDraw(t)
	require.Equal(t, "ok", ci.DrawStock())
	if _, ok := g.CanGoOut(); ok {
		t.Skip("たまたま上がれる手が配られた")
	}
	assert.Contains(t, ci.GoOut(0), "err:")
	assert.Equal(t, domain.ContinentalRummyPhaseDiscard, g.GetPhase(), "断ったのにフェーズが進んでいる")
}

func TestContinentalRummyInteractor_ResetWithConfig(t *testing.T) {
	ci, g := newContReal()

	t.Run("a valid config is applied", func(t *testing.T) {
		cfg := domain.DefaultContinentalRummyConfig()
		cfg.TotalRounds = domain.ContinentalRummyMaxRounds
		cfg.CpuDifficulty = domain.ContinentalRummyCpuDifficultyEasy
		assert.Equal(t, "ok", ci.ResetWithConfig(cfg))
		assert.Equal(t, cfg, ci.GetConfig())
		assert.Equal(t, cfg.TotalRounds, g.GetConfig().TotalRounds)
	})

	// 負のコントロール: 無効な設定は当たらず、直前の設定が残る。
	t.Run("an invalid config is refused and leaves the old one", func(t *testing.T) {
		kept := ci.GetConfig()
		bad := kept
		bad.TotalRounds = 0
		assert.Contains(t, ci.ResetWithConfig(bad), "err:")
		assert.Equal(t, kept, ci.GetConfig())
	})
}

func TestContinentalRummyInteractor_HintAndLog(t *testing.T) {
	ci, _ := newContReal()
	ci.Reset()
	assert.Equal(t, "hint", ci.Hint())
	assert.Equal(t, "log", ci.ActionLog())
}

// **保存して読み戻した盤で指し続けられること。**
func TestContinentalRummyInteractor_SnapshotRoundTrip(t *testing.T) {
	ci, g := contAtHumanDraw(t)
	require.Equal(t, "ok", ci.DrawStock())

	data, err := ci.Snapshot()
	require.NoError(t, err)
	assert.Greater(t, len(data), 2, "snapshot が `{}` -- MarshalJSON が無い")

	restored, err := usecase.RestoreContinentalRummyInteractor(data, contPassThrough{})
	require.NoError(t, err)
	rg := restored.Game
	assert.Equal(t, g.GetPhase(), rg.GetPhase())
	assert.Equal(t, g.GetRoundNumber(), rg.GetRoundNumber())
	assert.Equal(t, g.GetStockCount(), rg.GetStockCount())
	assert.Equal(t, g.GetCurrentPlayerIdx(), rg.GetCurrentPlayerIdx())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, g.GetPlayer(i).GetCardsSize(), rg.GetPlayer(i).GetCardsSize())
		assert.Equal(t, g.GetPlayer(i).GetScore(), rg.GetPlayer(i).GetScore())
	}
	// 復元した盤で指し続けられる (退化していない)。
	require.Equal(t, "ok", restored.Discard(0))
	assert.False(t, rg.GetGameEndFlag())
}
