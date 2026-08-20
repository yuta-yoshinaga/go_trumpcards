//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// chCard はテスト用のカード生成ヘルパー (design, value は plain int)。
func chCard(d, v int) *domain.Card {
	return domain.NewCard(d, v, false)
}

// chSetHand はプレイヤーの手札を明示的に設定する (Reset 後に AddCard)。
func chSetHand(p *domain.ChinchonPlayer, cards ...*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// newTestChinchon は全員人間の Chinchon を返す (CPU 自動進行を避けて決定的にテストするため)。
func newTestChinchon(n int) *domain.Chinchon {
	players := make([]*domain.ChinchonPlayer, 0, n)
	for i := 0; i < n; i++ {
		players = append(players, domain.NewChinchonPlayer(true))
	}
	cfg := domain.DefaultChinchonConfig()
	cfg.PlayerCount = n
	return domain.NewChinchon(players, cfg)
}

// chClearState は Reset が配ったカードを消し、明示的な状態構築に備える。
func chClearState(g *domain.Chinchon) {
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).Reset()
	}
	g.SetStock(nil)
	g.SetDiscardPile(nil)
}

func TestChinchonConfig_Validate(t *testing.T) {
	t.Run("default is valid", func(t *testing.T) {
		assert.NoError(t, domain.DefaultChinchonConfig().Validate())
	})
	t.Run("difficulty out of range", func(t *testing.T) {
		c := domain.DefaultChinchonConfig()
		c.CpuDifficulty = domain.ChinchonCpuDifficulty(99)
		assert.Error(t, c.Validate())
	})
	t.Run("player count below 2", func(t *testing.T) {
		c := domain.DefaultChinchonConfig()
		c.PlayerCount = 1
		assert.Error(t, c.Validate())
	})
	t.Run("player count above 4", func(t *testing.T) {
		c := domain.DefaultChinchonConfig()
		c.PlayerCount = 5
		assert.Error(t, c.Validate())
	})
	t.Run("knock threshold negative", func(t *testing.T) {
		c := domain.DefaultChinchonConfig()
		c.KnockThreshold = -1
		assert.Error(t, c.Validate())
	})
	t.Run("elimination limit below 1", func(t *testing.T) {
		c := domain.DefaultChinchonConfig()
		c.EliminationLimit = 0
		assert.Error(t, c.Validate())
	})
}

func TestChinchon_DeckIs40NoEights(t *testing.T) {
	g := domain.NewDefaultChinchon()
	g.Reset()
	total := g.GetDrawPileCount() + len(g.GetDiscardPile())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		total += g.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 40, total)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, 7, g.GetPlayer(i).GetCardsSize())
	}
	check := func(c *domain.Card) {
		assert.NotContains(t, []int{8, 9, 10}, c.GetValue())
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		for j := 0; j < p.GetCardsSize(); j++ {
			check(p.GetCard(j))
		}
	}
	for _, c := range g.GetDiscardPile() {
		check(c)
	}
}

func TestChinchon_ResetInitialState(t *testing.T) {
	g := domain.NewDefaultChinchon()
	g.Reset()
	assert.Equal(t, domain.ChinchonPhaseDraw, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, 0, g.GetCurrentPlayerIdx())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, -1, g.GetKnockerIdx())
	assert.Equal(t, 4, g.GetPlayerCnt())
	assert.True(t, g.IsHumanTurn())
}

func TestChinchon_PlayerCountConfigurable(t *testing.T) {
	for n := 2; n <= 4; n++ {
		cfg := domain.DefaultChinchonConfig()
		cfg.PlayerCount = n
		g := domain.NewChinchon([]*domain.ChinchonPlayer{}, cfg)
		g.Reset()
		assert.Equal(t, n, g.GetPlayerCnt())
	}
}

func TestChinchon_RankPositionAdjacency(t *testing.T) {
	// 7 (pos 7) と J (pos 8) は隣接し、有効なランを構成する。
	g := newTestChinchon(2)
	g.Reset()
	chClearState(g)
	// 手札: ♠5 ♠6 ♠7 ♠J(11) ♠Q(12) (5-6-7-J-Q は連続) + デッドウッド2枚
	p := g.GetPlayer(0)
	chSetHand(p,
		chCard(domain.CardDesignSpade, 5),
		chCard(domain.CardDesignSpade, 6),
		chCard(domain.CardDesignSpade, 7),
		chCard(domain.CardDesignSpade, 11),
		chCard(domain.CardDesignSpade, 12),
		chCard(domain.CardDesignHeart, 1),
		chCard(domain.CardDesignDiamond, 2),
	)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ChinchonPhaseDiscard)
	// ♦2 (idx 6) を捨ててノック: 残り6枚 = ♠5-6-7-J-Q ラン + ♥A デッドウッド(1点)
	require.NoError(t, g.PlayerKnock(6))
	// デッドウッド 1点 ≤ 閾値5 なので成功。
	assert.Equal(t, domain.ChinchonPhaseLayoff, g.GetPhase())
}

func TestChinchon_ChinchonInstantWinOnDiscard(t *testing.T) {
	// ドローで8枚にし、余分札を捨てて7枚同スート連続を残すとチンチョン (即時ゲーム勝利)。
	g := newTestChinchon(2)
	g.Reset()
	chClearState(g)
	p := g.GetPlayer(0)
	// 手札: ♥A-2-3-4-5-6 (6枚) + ♠K (余分)。山札から ♥7 を引く → 8枚。
	chSetHand(p,
		chCard(domain.CardDesignHeart, 1),
		chCard(domain.CardDesignHeart, 2),
		chCard(domain.CardDesignHeart, 3),
		chCard(domain.CardDesignHeart, 4),
		chCard(domain.CardDesignHeart, 5),
		chCard(domain.CardDesignHeart, 6),
		chCard(domain.CardDesignSpade, 13),
	)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ChinchonPhaseDraw)
	g.SetStock([]*domain.Card{chCard(domain.CardDesignHeart, 7)})
	require.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, domain.ChinchonPhaseDiscard, g.GetPhase())
	// ♠K を捨てると ♥A-2-3-4-5-6-7 の同スート7連続が残る → チンチョン。
	spadeKIdx := -1
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == domain.CardDesignSpade {
			spadeKIdx = i
		}
	}
	require.GreaterOrEqual(t, spadeKIdx, 0)
	require.NoError(t, g.PlayerDiscard(spadeKIdx))
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.ChinchonPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetWinnerIdx())
}

func TestChinchon_ChinchonSevenJAdjacency(t *testing.T) {
	// 7とJは隣接: ♦4-5-6-7-J-Q-K の7連続もチンチョン。
	g := newTestChinchon(2)
	g.Reset()
	chClearState(g)
	p := g.GetPlayer(0)
	chSetHand(p,
		chCard(domain.CardDesignDiamond, 4),
		chCard(domain.CardDesignDiamond, 5),
		chCard(domain.CardDesignDiamond, 6),
		chCard(domain.CardDesignDiamond, 7),
		chCard(domain.CardDesignDiamond, 11), // J
		chCard(domain.CardDesignDiamond, 12), // Q
		chCard(domain.CardDesignSpade, 1),    // 余分
	)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ChinchonPhaseDraw)
	g.SetStock([]*domain.Card{chCard(domain.CardDesignDiamond, 13)}) // ♦K
	require.NoError(t, g.PlayerDrawFromStock())
	spadeIdx := -1
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == domain.CardDesignSpade {
			spadeIdx = i
		}
	}
	require.GreaterOrEqual(t, spadeIdx, 0)
	require.NoError(t, g.PlayerDiscard(spadeIdx))
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerIdx())
}

func TestChinchon_KnockZeroDeadwood(t *testing.T) {
	g := newTestChinchon(2)
	g.Reset()
	chClearState(g)
	g.GetPlayer(1).AddCard(chCard(domain.CardDesignClover, 13)) // 相手にダミー
	chSetHand(g.GetPlayer(0),
		chCard(domain.CardDesignHeart, 1),
		chCard(domain.CardDesignHeart, 2),
		chCard(domain.CardDesignHeart, 3),
		chCard(domain.CardDesignHeart, 4),
		chCard(domain.CardDesignHeart, 5),
		chCard(domain.CardDesignHeart, 6),
		chCard(domain.CardDesignHeart, 7),
	)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ChinchonPhaseDiscard)
	require.NoError(t, g.PlayerKnock(6)) // ♥7 を捨てる → ♥A-2-3-4-5-6 完全ラン (デッドウッド0)
	assert.Equal(t, 0, domain.CalcDeadwoodValue(g.GetKnockerDeadwood()))
}

func TestChinchon_GetPlayerDeadwoodValue(t *testing.T) {
	g := newTestChinchon(2)
	g.Reset()
	chClearState(g)
	// A full 7-card heart run melds completely → deadwood 0.
	chSetHand(g.GetPlayer(0),
		chCard(domain.CardDesignHeart, 1),
		chCard(domain.CardDesignHeart, 2),
		chCard(domain.CardDesignHeart, 3),
		chCard(domain.CardDesignHeart, 4),
		chCard(domain.CardDesignHeart, 5),
		chCard(domain.CardDesignHeart, 6),
		chCard(domain.CardDesignHeart, 7),
	)
	assert.Equal(t, 0, g.GetPlayerDeadwoodValue(0))
	assert.Equal(t, g.GetConfig().KnockThreshold, g.GetKnockThreshold())
	// Out-of-range index returns 0 rather than panicking.
	assert.Equal(t, 0, g.GetPlayerDeadwoodValue(99))
}

func TestChinchon_DrawFromStock(t *testing.T) {
	g := newTestChinchon(2)
	g.Reset()
	before := g.GetPlayer(0).GetCardsSize()
	stockBefore := g.GetDrawPileCount()
	require.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, before+1, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, stockBefore-1, g.GetDrawPileCount())
	assert.Equal(t, domain.ChinchonPhaseDiscard, g.GetPhase())
}

func TestChinchon_DrawFromStock_WrongPhase(t *testing.T) {
	g := newTestChinchon(2)
	g.Reset()
	g.SetPhase(domain.ChinchonPhaseDiscard)
	assert.ErrorIs(t, g.PlayerDrawFromStock(), domain.ErrWrongPhase)
}

func TestChinchon_DrawFromStock_EmptyEndsRound(t *testing.T) {
	g := newTestChinchon(2)
	g.Reset()
	g.SetStock(nil)
	require.NoError(t, g.PlayerDrawFromStock())
	// 山札切れ → スコアリング → ラウンド終了またはゲーム終了。
	assert.Contains(t, []domain.ChinchonPhase{domain.ChinchonPhaseRoundEnd, domain.ChinchonPhaseGameEnd}, g.GetPhase())
}

func TestChinchon_DrawFromDiscard(t *testing.T) {
	g := newTestChinchon(2)
	g.Reset()
	chClearState(g)
	chSetHand(g.GetPlayer(0), chCard(domain.CardDesignSpade, 1))
	g.SetDiscardPile([]*domain.Card{chCard(domain.CardDesignHeart, 5)})
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ChinchonPhaseDraw)
	require.NoError(t, g.PlayerDrawFromDiscard())
	assert.Equal(t, 2, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, domain.ChinchonPhaseDiscard, g.GetPhase())
}

func TestChinchon_DrawFromDiscard_Empty(t *testing.T) {
	g := newTestChinchon(2)
	g.Reset()
	g.SetDiscardPile(nil)
	assert.Error(t, g.PlayerDrawFromDiscard())
}

func TestChinchon_Discard(t *testing.T) {
	g := newTestChinchon(2)
	g.Reset()
	chClearState(g)
	chSetHand(g.GetPlayer(0), chCard(domain.CardDesignSpade, 1), chCard(domain.CardDesignSpade, 2))
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ChinchonPhaseDiscard)
	require.NoError(t, g.PlayerDiscard(0))
	assert.Equal(t, 1, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	assert.Equal(t, domain.ChinchonPhaseDraw, g.GetPhase())
}

func TestChinchon_Discard_OutOfRange(t *testing.T) {
	g := newTestChinchon(2)
	g.Reset()
	g.SetPhase(domain.ChinchonPhaseDiscard)
	assert.Error(t, g.PlayerDiscard(999))
}

func TestChinchon_Knock_TooHighDeadwood(t *testing.T) {
	g := newTestChinchon(2)
	g.Reset()
	chClearState(g)
	// 全てバラバラの高デッドウッド手札 (7枚)。
	chSetHand(g.GetPlayer(0),
		chCard(domain.CardDesignSpade, 13),
		chCard(domain.CardDesignHeart, 12),
		chCard(domain.CardDesignDiamond, 11),
		chCard(domain.CardDesignClover, 7),
		chCard(domain.CardDesignSpade, 5),
		chCard(domain.CardDesignHeart, 3),
		chCard(domain.CardDesignDiamond, 1),
	)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ChinchonPhaseDiscard)
	assert.Error(t, g.PlayerKnock(0))
}

func TestChinchon_KnockAndScore(t *testing.T) {
	g := newTestChinchon(2)
	g.Reset()
	chClearState(g)
	// プレイヤー0: ♠A-2-3 ラン + ♥5-6-7 ラン + 捨てる1枚 → デッドウッド0でノック。
	chSetHand(g.GetPlayer(0),
		chCard(domain.CardDesignSpade, 1),
		chCard(domain.CardDesignSpade, 2),
		chCard(domain.CardDesignSpade, 3),
		chCard(domain.CardDesignHeart, 5),
		chCard(domain.CardDesignHeart, 6),
		chCard(domain.CardDesignHeart, 7),
		chCard(domain.CardDesignDiamond, 13), // 捨てる
	)
	// プレイヤー1: 高デッドウッド。
	chSetHand(g.GetPlayer(1),
		chCard(domain.CardDesignSpade, 13),
		chCard(domain.CardDesignHeart, 13),
		chCard(domain.CardDesignDiamond, 12),
		chCard(domain.CardDesignClover, 11),
		chCard(domain.CardDesignSpade, 7),
		chCard(domain.CardDesignHeart, 4),
		chCard(domain.CardDesignDiamond, 2),
	)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ChinchonPhaseDiscard)
	require.NoError(t, g.PlayerKnock(6))
	assert.Equal(t, 0, g.GetPlayer(0).GetRoundScore())
	assert.Equal(t, domain.ChinchonPhaseLayoff, g.GetPhase())
	// プレイヤー1がレイオフをスキップしてスコアリング。
	g.SetCurrentPlayerIdx(1)
	require.NoError(t, g.PlayerLayoff(nil))
	assert.Greater(t, g.GetPlayer(1).GetCumulativeScore(), 0)
	assert.Equal(t, 0, g.GetPlayer(0).GetCumulativeScore())
}

func TestChinchon_Layoff(t *testing.T) {
	g := newTestChinchon(2)
	g.Reset()
	chClearState(g)
	// ノッカーのメルド ♠A-2-3 を設定し、レイオフ可能な ♠4 を相手に持たせる。
	g.SetKnockerIdx(0)
	g.SetKnockerMelds([][]*domain.Card{{
		chCard(domain.CardDesignSpade, 1),
		chCard(domain.CardDesignSpade, 2),
		chCard(domain.CardDesignSpade, 3),
	}})
	chSetHand(g.GetPlayer(1), chCard(domain.CardDesignSpade, 4), chCard(domain.CardDesignHeart, 13))
	g.SetCurrentPlayerIdx(1)
	g.SetPhase(domain.ChinchonPhaseLayoff)
	require.NoError(t, g.PlayerLayoff([]int{0}))
	// ♠4 がメルドに付いて手札から消える → 残り ♥K のみ。
	assert.Equal(t, 1, g.GetPlayer(1).GetCardsSize())
}

func TestChinchon_Layoff_Invalid(t *testing.T) {
	g := newTestChinchon(2)
	g.Reset()
	chClearState(g)
	g.SetKnockerIdx(0)
	g.SetKnockerMelds([][]*domain.Card{{
		chCard(domain.CardDesignSpade, 1),
		chCard(domain.CardDesignSpade, 2),
		chCard(domain.CardDesignSpade, 3),
	}})
	chSetHand(g.GetPlayer(1), chCard(domain.CardDesignHeart, 13))
	g.SetCurrentPlayerIdx(1)
	g.SetPhase(domain.ChinchonPhaseLayoff)
	assert.Error(t, g.PlayerLayoff([]int{0})) // ♥K はレイオフ不可
	assert.Error(t, g.PlayerLayoff([]int{5})) // 範囲外
	assert.Error(t, g.PlayerLayoff([]int{0, 0}))
}

func TestChinchon_Elimination(t *testing.T) {
	g := newTestChinchon(2)
	cfg := g.GetConfig()
	cfg.EliminationLimit = 5
	g.SetConfig(cfg)
	g.Reset()
	chClearState(g)
	// プレイヤー1に高デッドウッドを持たせ、ノックで脱落上限を超えさせる。
	g.GetPlayer(1).SetCumulativeScore(0)
	chSetHand(g.GetPlayer(0),
		chCard(domain.CardDesignSpade, 1),
		chCard(domain.CardDesignSpade, 2),
		chCard(domain.CardDesignSpade, 3),
		chCard(domain.CardDesignHeart, 5),
		chCard(domain.CardDesignHeart, 6),
		chCard(domain.CardDesignHeart, 7),
		chCard(domain.CardDesignDiamond, 13),
	)
	chSetHand(g.GetPlayer(1),
		chCard(domain.CardDesignSpade, 13),
		chCard(domain.CardDesignHeart, 13),
		chCard(domain.CardDesignDiamond, 12),
		chCard(domain.CardDesignClover, 11),
		chCard(domain.CardDesignSpade, 7),
		chCard(domain.CardDesignHeart, 4),
		chCard(domain.CardDesignDiamond, 2),
	)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.ChinchonPhaseDiscard)
	require.NoError(t, g.PlayerKnock(6))
	g.SetCurrentPlayerIdx(1)
	require.NoError(t, g.PlayerLayoff(nil))
	// プレイヤー1の累積点 > 5 → 脱落 → 残り1人 (プレイヤー0) がマッチ勝者。
	assert.True(t, g.GetPlayer(1).GetEliminated())
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerIdx())
}

func TestChinchon_NextRound(t *testing.T) {
	g := newTestChinchon(2)
	g.Reset()
	g.SetPhase(domain.ChinchonPhaseRoundEnd)
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, domain.ChinchonPhaseDraw, g.GetPhase())
}

func TestChinchon_NextRound_WrongPhase(t *testing.T) {
	g := newTestChinchon(2)
	g.Reset()
	g.NextRound() // Draw フェーズなので no-op
	assert.Equal(t, 1, g.GetRoundNumber())
}

func TestChinchon_FullCpuGameTerminates(t *testing.T) {
	cfg := domain.DefaultChinchonConfig()
	cfg.EliminationLimit = 30 // 早めに脱落させてマッチを短縮。
	cfg.PlayerCount = 4
	players := make([]*domain.ChinchonPlayer, 0, cfg.PlayerCount)
	for i := 0; i < cfg.PlayerCount; i++ {
		players = append(players, domain.NewChinchonPlayer(false))
	}
	g := domain.NewChinchon(players, cfg)
	g.Reset()

	guard := 0
	for !g.GetGameEndFlag() && guard < 200000 {
		if g.GetPhase() == domain.ChinchonPhaseRoundEnd {
			g.NextRound()
		} else {
			g.CpuPlay()
		}
		guard++
	}
	assert.True(t, g.GetGameEndFlag(), "full-CPU game must terminate")
	assert.Less(t, guard, 200000)
}

func TestChinchon_JSONRoundTrip(t *testing.T) {
	g := domain.NewDefaultChinchon()
	g.Reset()
	data, err := g.MarshalJSON()
	require.NoError(t, err)

	var restored domain.Chinchon
	require.NoError(t, restored.UnmarshalJSON(data))
	assert.Equal(t, g.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetDrawPileCount(), restored.GetDrawPileCount())
}

func TestChinchon_UnmarshalRejectsInvalid(t *testing.T) {
	t.Run("malformed json", func(t *testing.T) {
		var g domain.Chinchon
		assert.Error(t, g.UnmarshalJSON([]byte("not json")))
	})
	t.Run("invalid config", func(t *testing.T) {
		var g domain.Chinchon
		assert.Error(t, g.UnmarshalJSON([]byte(`{"cf":{"cd":99,"pc":4,"kt":5,"el":100},"pl":[]}`)))
	})
	t.Run("player count mismatch", func(t *testing.T) {
		var g domain.Chinchon
		assert.Error(t, g.UnmarshalJSON([]byte(`{"cf":{"cd":1,"pc":4,"kt":5,"el":100},"pl":[]}`)))
	})
	t.Run("nil player element", func(t *testing.T) {
		var g domain.Chinchon
		assert.Error(t, g.UnmarshalJSON([]byte(`{"cf":{"cd":1,"pc":2,"kt":5,"el":100},"pl":[null,null]}`)))
	})
	t.Run("phase out of range", func(t *testing.T) {
		g := domain.NewDefaultChinchon()
		g.Reset()
		data, _ := g.MarshalJSON()
		var j map[string]any
		_ = json.Unmarshal(data, &j)
		j["ps"] = 99
		bad, _ := json.Marshal(j)
		var g2 domain.Chinchon
		assert.Error(t, g2.UnmarshalJSON(bad))
	})
	t.Run("current index out of range", func(t *testing.T) {
		g := domain.NewDefaultChinchon()
		g.Reset()
		data, _ := g.MarshalJSON()
		var j map[string]any
		_ = json.Unmarshal(data, &j)
		j["ci"] = 99
		bad, _ := json.Marshal(j)
		var g2 domain.Chinchon
		assert.Error(t, g2.UnmarshalJSON(bad))
	})
}

func TestChinchonPlayer_JSONRoundTrip(t *testing.T) {
	p := domain.NewChinchonPlayer(true)
	p.AddCard(chCard(domain.CardDesignSpade, 1))
	p.SetCumulativeScore(42)
	p.SetEliminated(true)
	data, err := p.MarshalJSON()
	require.NoError(t, err)
	var restored domain.ChinchonPlayer
	require.NoError(t, restored.UnmarshalJSON(data))
	assert.Equal(t, 42, restored.GetCumulativeScore())
	assert.True(t, restored.GetEliminated())
	assert.Equal(t, 1, restored.GetCardsSize())
}

func TestChinchon_GameEndedGuards(t *testing.T) {
	g := newTestChinchon(2)
	g.Reset()
	g.SetPhase(domain.ChinchonPhaseGameEnd)
	// gameEndFlag は SetPhase では立たないので、JSON 経由で立てる代わりに各アクションのフェーズガードを検証。
	g.SetPhase(domain.ChinchonPhaseDraw)
	g.SetCurrentPlayerIdx(0)
	// Layoff を Draw フェーズで呼ぶとエラー。
	assert.ErrorIs(t, g.PlayerLayoff(nil), domain.ErrWrongPhase)
	assert.ErrorIs(t, g.PlayerDiscard(0), domain.ErrWrongPhase)
	assert.ErrorIs(t, g.PlayerKnock(0), domain.ErrWrongPhase)
}

// **合計は既存の deadwoodLine と一致する。**分割と合計が別経路になると、
// 内訳の和が表示値とずれる (#4838)。
func TestChinchon_GetPlayerMeldSplit(t *testing.T) {
	g := domain.NewDefaultChinchon()
	g.Reset()
	p := g.GetPlayer(0)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	// ♠5-6-7 のラン + 端数 ♥9, ♣2。
	for _, c := range []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 5, false),
		domain.NewCard(domain.CardDesignSpade, 6, false),
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignHeart, 9, false),
		domain.NewCard(domain.CardDesignClover, 2, false),
	} {
		p.AddCard(c)
	}

	melds, dead := g.GetPlayerMeldSplit(0)
	assert.Len(t, melds, 1)
	assert.Len(t, melds[0], 3)
	assert.Len(t, dead, 2)
	// 内訳の合計が GetPlayerDeadwoodValue と一致する。
	assert.Equal(t, g.GetPlayerDeadwoodValue(0), domain.CalcDeadwoodValue(dead))
	assert.Equal(t, 11, domain.CalcDeadwoodValue(dead))

	// 範囲外は空。
	m2, d2 := g.GetPlayerMeldSplit(99)
	assert.Nil(t, m2)
	assert.Nil(t, d2)
}

// **8/9/10 を抜いた 40 枚デッキなので、7 の次は J。** 公開版が内部の定義から
// ずれると、画面のラン判定だけが古い並びで動く (#5665)。
func TestChinchonRankPositionIsContiguousOverTheDeck(t *testing.T) {
	// A..7 → 1..7、J/Q/K → 8..10 と隙間なく並ぶ。
	want := map[int]int{1: 1, 2: 2, 3: 3, 4: 4, 5: 5, 6: 6, 7: 7, 11: 8, 12: 9, 13: 10}
	for value, pos := range want {
		if got := domain.ChinchonRankPosition(value); got != pos {
			t.Errorf("domain.ChinchonRankPosition(%d) = %d, want %d", value, got, pos)
		}
	}
	// **7 と J は隣り合う。** ここが 4 空くと、♠7-♠J のランが組めなくなる。
	if domain.ChinchonRankPosition(11)-domain.ChinchonRankPosition(7) != 1 {
		t.Errorf("7 と J が隣接していない: %d → %d", domain.ChinchonRankPosition(7), domain.ChinchonRankPosition(11))
	}
	// デッキに無いランクは並びの外。
	for _, value := range []int{0, 8, 9, 10, 14} {
		if got := domain.ChinchonRankPosition(value); got > 0 && got <= 10 {
			t.Errorf("domain.ChinchonRankPosition(%d) = %d, デッキに無いランクが並びに入っている", value, got)
		}
	}
}
