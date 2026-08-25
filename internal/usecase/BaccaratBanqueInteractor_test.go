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

// banquePassThrough は「呼ばれたか」だけを見る素通しのプレゼンター。
type banquePassThrough struct{}

func (banquePassThrough) Output(_ interfaces.BaccaratBanqueGame, lastErr error) string {
	if lastErr != nil {
		return "err:" + lastErr.Error()
	}
	return "ok"
}
func (banquePassThrough) HintOutput(_ interfaces.BaccaratBanqueGame) string      { return "hint" }
func (banquePassThrough) ActionLogOutput(_ interfaces.BaccaratBanqueGame) string { return "log" }

func newBanqueReal() (*usecase.BaccaratBanqueInteractor, *domain.BaccaratBanque) {
	g := domain.NewDefaultBaccaratBanque()
	return usecase.NewBaccaratBanqueInteractor(g, banquePassThrough{}), g
}

// banqueDealtStack は 右・左・親 の順に 2 周配られる決め打ちのシューを返す。
func banqueDealtStack(vals ...int) []*domain.Card {
	out := make([]*domain.Card, 0, len(vals)+20)
	for _, v := range vals {
		out = append(out, domain.NewCard(domain.CardDesignSpade, v, true))
	}
	// 予備は多めに積む。**シューが尽きるとバンクが終わる**ので、足りないと
	// 「続けたのに進まない」形で試験が落ちる。
	for i := 0; i < 60; i++ {
		out = append(out, domain.NewCard(domain.CardDesignHeart, 10, true)) // 0 点
	}
	return out
}

// newBanqueAtDecision は「親がまだ引くか決めていない」局面を決め打ちで作る。
//
// **Reset のままでは足りない。** 親がナチュラルだとその場で決着まで進むので、
// 素の Reset に続けて Stand を打つ試験は配りしだいで「フェーズ違い」で落ちる。
func newBanqueAtDecision(t *testing.T) (*usecase.BaccaratBanqueInteractor, *domain.BaccaratBanque) {
	t.Helper()
	bi, g := newBanqueReal()
	bi.Reset()
	// 右 = 3 (必ず引く)、左 = 7 (必ず止まる)、親 = 6。
	g.SetShoeForTest(banqueDealtStack(1, 3, 3, 2, 4, 3))
	g.SetPhaseForTest(domain.BaccaratBanquePhaseResult)
	bi.NextCoup()
	require.Equal(t, domain.BaccaratBanquePhaseBanker, g.GetPhase(),
		"親の判断を待つ局面になっていない -- 前提が崩れている")
	return bi, g
}

func TestNewBaccaratBanqueInteractor_NilGuards(t *testing.T) {
	bp := new(presenter.MockBaccaratBanquePresenter)
	assert.PanicsWithValue(t, "BaccaratBanqueInteractor: g must not be nil", func() {
		usecase.NewBaccaratBanqueInteractor(nil, bp)
	})
	assert.PanicsWithValue(t, "BaccaratBanqueInteractor: bp must not be nil", func() {
		usecase.NewBaccaratBanqueInteractor(new(interfaces.MockBaccaratBanqueGame), nil)
	})
}

func TestBaccaratBanqueInteractor_ResetOpensOnTheBankerDecision(t *testing.T) {
	bi, g := newBanqueReal()
	assert.Equal(t, "ok", bi.Reset())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetCoupNumber())
	// **子は Reset のうちに決着している。** 人間が最初に触るのは自分の引き際 ──
	// ただし親がナチュラルなら引く判断そのものが無く、そのまま決着まで進む。
	banker := g.GetPlayer(domain.BaccaratBanqueBankerIdx)
	if domain.BaccaratBanqueIsNatural(banker.GetHand()) {
		assert.Equal(t, domain.BaccaratBanquePhaseResult, g.GetPhase())
		return
	}
	assert.Equal(t, domain.BaccaratBanquePhaseBanker, g.GetPhase())
}

// **引くと止まるを別々に叩けること。** 片方しか効いていないと、既定のまま
// 届いた要求が黙ってもう片方になる。
func TestBaccaratBanqueInteractor_DrawAndStandAreSeparateEntryPoints(t *testing.T) {
	t.Run("stand settles without taking a card", func(t *testing.T) {
		bi, g := newBanqueAtDecision(t)
		before := g.GetPlayer(domain.BaccaratBanqueBankerIdx).GetCardsSize()
		assert.Equal(t, "ok", bi.Stand())
		assert.Equal(t, before, g.GetPlayer(domain.BaccaratBanqueBankerIdx).GetCardsSize())
		assert.Equal(t, domain.BaccaratBanquePhaseResult, g.GetPhase())
	})

	t.Run("draw takes a card", func(t *testing.T) {
		bi, g := newBanqueAtDecision(t)
		banker := g.GetPlayer(domain.BaccaratBanqueBankerIdx)
		require.Equal(t, 2, banker.GetCardsSize())
		assert.Equal(t, "ok", bi.Draw())
		assert.Equal(t, 3, banker.GetCardsSize())
		assert.Equal(t, domain.BaccaratBanquePhaseResult, g.GetPhase())
	})
}

func TestBaccaratBanqueInteractor_NextCoupAndRetire(t *testing.T) {
	t.Run("nextcoup deals the next coup and the bank survives", func(t *testing.T) {
		bi, g := newBanqueAtDecision(t)
		coup, held := g.GetCoupNumber(), g.GetBankHeld()
		require.Equal(t, "ok", bi.Stand())
		require.Equal(t, domain.BaccaratBanquePhaseResult, g.GetPhase())
		assert.Equal(t, "ok", bi.NextCoup())
		// **負けても席は動かない。** ここが chemin de fer との唯一の違い。
		assert.Equal(t, coup+1, g.GetCoupNumber())
		assert.Equal(t, held+1, g.GetBankHeld())
		assert.False(t, g.GetGameEndFlag())
	})

	t.Run("retire ends the bank", func(t *testing.T) {
		bi, g := newBanqueAtDecision(t)
		require.Equal(t, "ok", bi.Stand())
		assert.Equal(t, "ok", bi.Retire())
		assert.True(t, g.GetGameEndFlag())
		assert.True(t, g.IsRetired())
	})

	// **終わった盤は動かない。** 例外を投げるのではなく盤を出し直すのが
	// このリポジトリの約束 (guardGameEnd) なので、見るのは文言ではなく状態。
	t.Run("commands after the bank ended change nothing", func(t *testing.T) {
		bi, g := newBanqueAtDecision(t)
		require.Equal(t, "ok", bi.Stand())
		require.Equal(t, "ok", bi.Retire())
		require.True(t, g.GetGameEndFlag())
		coup, held, shoe := g.GetCoupNumber(), g.GetBankHeld(), g.GetShoeRemaining()
		hand := g.GetPlayer(domain.BaccaratBanqueBankerIdx).GetCardsSize()
		chips := g.GetPlayer(domain.BaccaratBanqueBankerIdx).GetChips()
		for _, out := range []string{bi.Draw(), bi.Stand(), bi.Retire(), bi.NextCoup()} {
			assert.NotContains(t, out, "panic")
		}
		assert.Equal(t, coup, g.GetCoupNumber())
		assert.Equal(t, held, g.GetBankHeld())
		assert.Equal(t, shoe, g.GetShoeRemaining(), "終局後に札が減っている")
		assert.Equal(t, hand, g.GetPlayer(domain.BaccaratBanqueBankerIdx).GetCardsSize())
		assert.Equal(t, chips, g.GetPlayer(domain.BaccaratBanqueBankerIdx).GetChips())
		assert.Equal(t, domain.BaccaratBanquePhaseGameEnd, g.GetPhase())
	})
}

func TestBaccaratBanqueInteractor_ResetWithConfig(t *testing.T) {
	bi, g := newBanqueReal()

	t.Run("a valid config is applied", func(t *testing.T) {
		cfg := domain.DefaultBaccaratBanqueConfig()
		cfg.StartChips = 5000
		cfg.BetAmount = 100
		cfg.CpuDifficulty = domain.BaccaratBanqueCpuDifficultyEasy
		assert.Equal(t, "ok", bi.ResetWithConfig(cfg))
		assert.Equal(t, cfg, bi.GetConfig())
		// **親の残高そのものは見ない。** Reset は 1 クー目を配るので、親が
		// ナチュラルだとその場で決着し、読んだときには賭け金ぶん動いている。
		// 動かないのは席をまたいだ合計 (精算はゼロサム) のほう。
		total := 0
		for i := 0; i < g.GetPlayerCnt(); i++ {
			total += g.GetPlayer(i).GetChips()
		}
		assert.Equal(t, 5000*g.GetPlayerCnt(), total)
		assert.Equal(t, 100, g.GetPlayer(domain.BaccaratBanqueRightIdx).GetBet())
	})

	// 負のコントロール: 無効な設定は当たらず、直前の設定が残る。
	t.Run("an invalid config is refused and leaves the old one", func(t *testing.T) {
		kept := bi.GetConfig()
		bad := kept
		bad.StartChips = 1
		out := bi.ResetWithConfig(bad)
		assert.Contains(t, out, "err:")
		assert.Equal(t, kept, bi.GetConfig())
	})
}

func TestBaccaratBanqueInteractor_HintAndLog(t *testing.T) {
	bi, _ := newBanqueReal()
	bi.Reset()
	assert.Equal(t, "hint", bi.Hint())
	assert.Equal(t, "log", bi.ActionLog())
}

// **保存して読み戻した盤で指し続けられること。** 非公開フィールドしか無い型は
// MarshalJSON が無いと `{}` になり、復元した盤が最初の 1 手で破綻する。
func TestBaccaratBanqueInteractor_SnapshotRoundTrip(t *testing.T) {
	bi, g := newBanqueAtDecision(t)
	require.Equal(t, "ok", bi.Stand())
	require.Equal(t, "ok", bi.NextCoup())
	coup := g.GetCoupNumber()

	data, err := bi.Snapshot()
	require.NoError(t, err)
	assert.Greater(t, len(data), 2, "snapshot が `{}` -- MarshalJSON が無い")

	restored, err := usecase.RestoreBaccaratBanqueInteractor(data, banquePassThrough{})
	require.NoError(t, err)
	rg := restored.Game
	assert.Equal(t, g.GetCoupNumber(), rg.GetCoupNumber())
	assert.Equal(t, g.GetBankHeld(), rg.GetBankHeld())
	assert.Equal(t, g.GetShoeRemaining(), rg.GetShoeRemaining())
	assert.Equal(t, g.GetPhase(), rg.GetPhase())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, g.GetPlayer(i).GetChips(), rg.GetPlayer(i).GetChips())
		assert.Equal(t, g.GetPlayer(i).GetTotal(), rg.GetPlayer(i).GetTotal())
	}
	// 復元した盤で指し続けられる (退化していない)。
	require.False(t, rg.GetGameEndFlag(), "復元した盤が既に終わっている")
	require.Equal(t, "ok", restored.Stand())
	assert.Equal(t, domain.BaccaratBanquePhaseResult, rg.GetPhase())
	restored.NextCoup()
	assert.Equal(t, coup+1, rg.GetCoupNumber())
}
