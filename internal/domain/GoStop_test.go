//go:build test

package domain_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// gostopCard は Go-Stop テスト用のカード生成ショートカット (design=月, value=index)。
func gostopCard(month, index int) *domain.Card { return domain.NewCard(month, index, false) }

// gostopScoreOf は取り札集合とゴー回数に対する得点内訳と最終点を返す。
func gostopScoreOf(cards []*domain.Card, goCount int) (*domain.GoStopBreakdown, int) {
	human := domain.NewGoStopPlayer(true)
	human.AddCaptured(cards)
	for i := 0; i < goCount; i++ {
		human.IncGoCount()
	}
	cpu := domain.NewGoStopPlayer(false)
	g := domain.NewGoStop([]*domain.GoStopPlayer{human, cpu}, domain.DefaultGoStopConfig())
	return g.GetScore(0)
}

// 主要カード。
var (
	gsCrane   = gostopCard(1, 1)  // 松 光
	gsCurtain = gostopCard(3, 1)  // 桜 光
	gsMoon    = gostopCard(8, 1)  // 芒 光
	gsRain    = gostopCard(11, 1) // 柳 光 (雨)
	gsPhoenix = gostopCard(12, 1) // 桐 光
	gsWarbler = gostopCard(2, 1)  // 梅 열끗 (鶯)
	gsCuckoo  = gostopCard(4, 1)  // 藤 열끗 (不如帰)
	gsGeese   = gostopCard(8, 2)  // 芒 열끗 (雁)
	gsBridge  = gostopCard(5, 1)  // 菖蒲 열끗
	gsButter  = gostopCard(6, 1)  // 牡丹 열끗
	gsBoar    = gostopCard(7, 1)  // 萩 열끗
	gsSake    = gostopCard(9, 1)  // 菊 열끗 (盃)
	gsDeer    = gostopCard(10, 1) // 紅葉 열끗
	gsSwallow = gostopCard(11, 2) // 柳 열끗 (燕)
)

// gostopChaff は count 枚の피を返す (最大 24 枚)。
func gostopChaff(count int) []*domain.Card {
	pool := []*domain.Card{}
	for _, m := range []int{1, 2, 3, 4, 5} {
		pool = append(pool, gostopCard(m, 3), gostopCard(m, 4))
	}
	pool = append(pool, gostopCard(11, 4), gostopCard(12, 2), gostopCard(12, 3), gostopCard(12, 4))
	if count > len(pool) {
		count = len(pool)
	}
	return pool[:count]
}

// --- カテゴリ得点 ---

func TestGoStopScore_Gwang3NoRain(t *testing.T) {
	bd, total := gostopScoreOf([]*domain.Card{gsCrane, gsCurtain, gsPhoenix}, 0)
	assert.Equal(t, 3, bd.Gwang)
	assert.Equal(t, 3, bd.Base)
	assert.Equal(t, 3, total)
}

func TestGoStopScore_Gwang3WithRain(t *testing.T) {
	bd, _ := gostopScoreOf([]*domain.Card{gsCrane, gsRain, gsPhoenix}, 0)
	assert.Equal(t, 2, bd.Gwang)
	assert.Equal(t, 2, bd.Base)
}

func TestGoStopScore_Gwang4(t *testing.T) {
	bd, _ := gostopScoreOf([]*domain.Card{gsCrane, gsCurtain, gsMoon, gsPhoenix}, 0)
	assert.Equal(t, 4, bd.Gwang)
}

func TestGoStopScore_Gwang5(t *testing.T) {
	bd, _ := gostopScoreOf([]*domain.Card{gsCrane, gsCurtain, gsMoon, gsRain, gsPhoenix}, 0)
	assert.Equal(t, 15, bd.Gwang)
	assert.Equal(t, 5, bd.BrightCount)
}

func TestGoStopScore_Godori(t *testing.T) {
	bd, total := gostopScoreOf([]*domain.Card{gsWarbler, gsCuckoo, gsGeese}, 0)
	assert.Equal(t, 5, bd.Godori)
	assert.Equal(t, 0, bd.Yeol) // 3 匹では 열끗 未成立
	assert.Equal(t, 5, total)
}

func TestGoStopScore_GodoriIncompleteNoPoints(t *testing.T) {
	bd, _ := gostopScoreOf([]*domain.Card{gsWarbler, gsCuckoo}, 0)
	assert.Equal(t, 0, bd.Godori)
}

func TestGoStopScore_Hongdan(t *testing.T) {
	bd, _ := gostopScoreOf([]*domain.Card{gostopCard(1, 2), gostopCard(2, 2), gostopCard(3, 2)}, 0)
	assert.Equal(t, 3, bd.Tti)
	assert.Equal(t, 3, bd.RibbonCount)
}

func TestGoStopScore_Chongdan(t *testing.T) {
	bd, _ := gostopScoreOf([]*domain.Card{gostopCard(6, 2), gostopCard(9, 2), gostopCard(10, 2)}, 0)
	assert.Equal(t, 3, bd.Tti)
}

func TestGoStopScore_Chodan(t *testing.T) {
	bd, _ := gostopScoreOf([]*domain.Card{gostopCard(4, 2), gostopCard(5, 2), gostopCard(7, 2)}, 0)
	assert.Equal(t, 3, bd.Tti)
}

func TestGoStopScore_TtiExtraNoTriple(t *testing.T) {
	// 6 枚の띠だが完成する三役は無い → 6-5 = 1 点。
	bd, _ := gostopScoreOf([]*domain.Card{
		gostopCard(1, 2), gostopCard(2, 2), // 홍단 2 枚
		gostopCard(6, 2), gostopCard(9, 2), // 청단 2 枚
		gostopCard(4, 2), gostopCard(5, 2), // 초단 2 枚
	}, 0)
	assert.Equal(t, 2, bd.Tti) // 三役無しの 6 枚띠 → 1 + (6-5) = 2
	assert.Equal(t, 6, bd.RibbonCount)
}

func TestGoStopScore_TtiTripleAndExtra(t *testing.T) {
	// 홍단完成 (3) + 4 枚目の띠 → 홍단 3、超過分は 4-5<0 なので 0。合計 3。
	bd, _ := gostopScoreOf([]*domain.Card{
		gostopCard(1, 2), gostopCard(2, 2), gostopCard(3, 2), gostopCard(11, 3),
	}, 0)
	assert.Equal(t, 3, bd.Tti)
}

func TestGoStopScore_Yeol5(t *testing.T) {
	bd, _ := gostopScoreOf([]*domain.Card{gsBridge, gsButter, gsBoar, gsSake, gsDeer}, 0)
	assert.Equal(t, 1, bd.Yeol)
	assert.Equal(t, 5, bd.AnimalCount)
}

func TestGoStopScore_YeolExtra(t *testing.T) {
	bd, _ := gostopScoreOf([]*domain.Card{gsBridge, gsButter, gsBoar, gsSake, gsDeer, gsSwallow}, 0)
	assert.Equal(t, 2, bd.Yeol)
}

func TestGoStopScore_Pi10(t *testing.T) {
	bd, _ := gostopScoreOf(gostopChaff(10), 0)
	assert.Equal(t, 1, bd.Pi)
	assert.Equal(t, 10, bd.PiCount)

	bd2, _ := gostopScoreOf(gostopChaff(11), 0)
	assert.Equal(t, 2, bd2.Pi)
}

func TestGoStopScore_PiBelowThreshold(t *testing.T) {
	bd, _ := gostopScoreOf(gostopChaff(9), 0)
	assert.Equal(t, 0, bd.Pi)
	assert.Equal(t, 9, bd.PiCount)
}

func TestGoStopScore_Empty(t *testing.T) {
	bd, total := gostopScoreOf(nil, 0)
	assert.Equal(t, 0, bd.Base)
	assert.Equal(t, 0, total)
}

func TestGoStopScore_Combined(t *testing.T) {
	// 5 光 + 3 鳥 (고도리) → 15 + 5 = 20。
	bd, total := gostopScoreOf([]*domain.Card{
		gsCrane, gsCurtain, gsMoon, gsRain, gsPhoenix, gsWarbler, gsCuckoo, gsGeese,
	}, 0)
	assert.Equal(t, 15, bd.Gwang)
	assert.Equal(t, 5, bd.Godori)
	assert.Equal(t, 20, bd.Base)
	assert.Equal(t, 20, total)
}

// --- ゴー掛け金 / 倍率 ---

func TestGoStopScore_GoBonus(t *testing.T) {
	base := []*domain.Card{gsCrane, gsCurtain, gsPhoenix} // gwang=3, base=3
	bd0, t0 := gostopScoreOf(base, 0)
	assert.Equal(t, 3, t0)
	assert.Equal(t, 1, bd0.GoMult)

	bd1, t1 := gostopScoreOf(base, 1)
	assert.Equal(t, 4, t1) // 3+1
	assert.Equal(t, 1, bd1.GoMult)

	_, t2 := gostopScoreOf(base, 2)
	assert.Equal(t, 5, t2) // 3+2

	bd3, t3 := gostopScoreOf(base, 3)
	assert.Equal(t, 2, bd3.GoMult)
	assert.Equal(t, 12, t3) // (3+3)*2

	bd4, t4 := gostopScoreOf(base, 4)
	assert.Equal(t, 4, bd4.GoMult)
	assert.Equal(t, 28, t4) // (3+4)*4
}

// --- ゲーム進行 ---

func newTestGoStop(t *testing.T, diff domain.GoStopCpuDifficulty) *domain.GoStop {
	t.Helper()
	g := domain.NewDefaultGoStop()
	cfg := domain.DefaultGoStopConfig()
	cfg.CpuDifficulty = diff
	g.SetConfig(cfg)
	g.Reset()
	return g
}

func TestGoStopReset_Deal(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	assert.Equal(t, domain.GoStopPhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, domain.GoStopFieldSize, len(g.GetFieldCards()))
	total := len(g.GetFieldCards()) + g.GetRemainingDeck()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, domain.GoStopHandSize, g.GetPlayer(i).GetCardsSize())
		total += g.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 48, total)
	assert.Equal(t, 20, g.GetRemainingDeck()) // 48 - 10*2 - 8
}

func TestGoStopPlayerPlay_Errors(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	assert.Error(t, g.PlayerPlay(99, -1))
	g.SetPhase(domain.GoStopPhaseRoundEnd)
	assert.Error(t, g.PlayerPlay(0, -1))
	g.SetPhase(domain.GoStopPhasePlay)
	g.SetCurrentTurn(1)
	assert.ErrorIs(t, g.PlayerPlay(0, -1), domain.ErrNotHumanTurn)
}

func TestGoStopPlayerPlay_InvalidFieldChoice(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	human := g.GetPlayer(0)
	human.Reset()
	human.AddCard(gostopCard(1, 1))
	g.SetFieldCards([]*domain.Card{gostopCard(5, 1)})
	assert.Error(t, g.PlayerPlay(0, 0))
}

func TestGoStopCapture(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	human := g.GetPlayer(0)
	human.Reset()
	human.AddCard(gostopCard(3, 3)) // 桜 피
	g.SetFieldCards([]*domain.Card{gostopCard(3, 4), gostopCard(7, 3)})
	require.NoError(t, g.PlayerPlay(0, -1))
	assert.GreaterOrEqual(t, human.CapturedCount(), 2)
}

func TestGoStopDecide_WrongPhase(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	assert.Error(t, g.PlayerDecide(true))
}

func TestGoStopNextRound_GuardedToRoundEnd(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	before := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, before, g.GetRoundNumber())
}

func TestGoStopHint(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	h := g.GetHint()
	require.NotNil(t, h)
	assert.GreaterOrEqual(t, h.CardIndex, 0)
	assert.Equal(t, -1, h.Go)

	g.SetPhase(domain.GoStopPhaseGoDecision)
	h2 := g.GetHint()
	require.NotNil(t, h2)
	assert.NotEqual(t, -1, h2.Go)
}

// gostopGwang3 は雨無しの 3 光 (base=3)。
func gostopGwang3() []*domain.Card {
	return []*domain.Card{gostopCard(1, 1), gostopCard(3, 1), gostopCard(12, 1)}
}

// gostopGodori は 3 鳥 (base=5, gwang=0)。
func gostopGodori() []*domain.Card { return []*domain.Card{gsWarbler, gsCuckoo, gsGeese} }

func TestGoStopDecide_GoContinues(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetPhase(domain.GoStopPhaseGoDecision)
	require.NoError(t, g.PlayerDecide(true))
	assert.Equal(t, 1, g.GetPlayer(0).GetGoCount())
	assert.True(t, g.GetPlayer(0).GetCalledGo())
	assert.Equal(t, domain.GoStopPhasePlay, g.GetPhase())
}

func TestGoStopDecide_StopWinsRound(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetPhase(domain.GoStopPhaseGoDecision)
	g.GetPlayer(0).AddCaptured(gostopGwang3())
	// 敗者に 5 枚以上の피を与え、피박を回避。光札は無い → 光박成立 (×2)。
	g.GetPlayer(1).AddCaptured(gostopChaff(6))
	require.NoError(t, g.PlayerDecide(false))
	assert.Equal(t, domain.GoStopPhaseRoundEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetRoundWinner())
	res := g.GetLastRoundResult()
	require.NotNil(t, res)
	assert.Equal(t, 3, res.BasePoints)
	assert.True(t, res.GwangBak)
	assert.False(t, res.PiBak)
	assert.Equal(t, 2, res.BakMult)
	assert.Equal(t, 6, res.Total)
	assert.Equal(t, 6, g.GetPlayer(0).GetScore())
}

func TestGoStopBak_PiBak(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetPhase(domain.GoStopPhaseGoDecision)
	g.GetPlayer(0).AddCaptured(append(gostopGodori(), gostopChaff(10)...)) // Godori(5) + Pi(1) = 6 pts
	g.GetPlayer(1).AddCaptured(gostopChaff(3))                             // 피<5 → 피박
	require.NoError(t, g.PlayerDecide(false))
	res := g.GetLastRoundResult()
	require.NotNil(t, res)
	assert.False(t, res.GwangBak)
	assert.True(t, res.PiBak)
	assert.False(t, res.GoBak)
	assert.Equal(t, 2, res.BakMult)
	assert.Equal(t, 12, res.Total) // 6 * 2
}

func TestGoStopBak_GoBak(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetPhase(domain.GoStopPhaseGoDecision)
	g.GetPlayer(0).AddCaptured(gostopGodori()) // gwang=0
	// 敗者は 5 枚以上の피 (피박回避) + ゴー宣言済み (고박)。
	g.GetPlayer(1).AddCaptured(gostopChaff(6))
	g.GetPlayer(1).IncGoCount()
	require.NoError(t, g.PlayerDecide(false))
	res := g.GetLastRoundResult()
	require.NotNil(t, res)
	assert.False(t, res.GwangBak)
	assert.False(t, res.PiBak)
	assert.True(t, res.GoBak)
	assert.Equal(t, 2, res.BakMult)
	assert.Equal(t, 10, res.Total)
}

func TestGoStopBak_Stacked(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetPhase(domain.GoStopPhaseGoDecision)
	g.GetPlayer(0).AddCaptured(append(gostopGwang3(), gostopChaff(10)...)) // Gwang(3)+Pi(1) = 4 pts
	g.GetPlayer(1).AddCaptured(gostopChaff(2))                             // 光札0 + 피<5 → 光박×피박
	g.GetPlayer(1).IncGoCount()                                            // 고박
	require.NoError(t, g.PlayerDecide(false))
	res := g.GetLastRoundResult()
	require.NotNil(t, res)
	assert.Equal(t, 8, res.BakMult) // 2*2*2
	assert.Equal(t, 32, res.Total)  // 4 * 8
}

func TestGoStopEndRound_GoScoreMultiplier(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	// 3 回ゴー → ×2 倍率。決断を 3 回続けてからストップ。
	for i := 0; i < 3; i++ {
		g.SetCurrentTurn(0)
		g.SetPhase(domain.GoStopPhaseGoDecision)
		require.NoError(t, g.PlayerDecide(true))
	}
	g.SetCurrentTurn(0)
	g.SetPhase(domain.GoStopPhaseGoDecision)
	g.GetPlayer(0).AddCaptured(gostopGwang3())
	g.GetPlayer(1).AddCaptured(gostopChaff(6)) // 피박回避、光札0 → 光박のみ ×2
	require.NoError(t, g.PlayerDecide(false))
	res := g.GetLastRoundResult()
	require.NotNil(t, res)
	assert.Equal(t, 3, res.GoCount)
	assert.Equal(t, 12, res.GoScore) // (3+3)*2
	assert.Equal(t, 24, res.Total)   // 12 * 2 (光박)
}

func TestGoStopEndRound_FinishesGameOnTarget(t *testing.T) {
	g := domain.NewDefaultGoStop()
	cfg := domain.DefaultGoStopConfig()
	cfg.TargetScore = 1
	g.SetConfig(cfg)
	g.Reset()
	g.SetCurrentTurn(0)
	g.SetPhase(domain.GoStopPhaseGoDecision)
	g.GetPlayer(0).AddCaptured(gostopGwang3())
	require.NoError(t, g.PlayerDecide(false))
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.GoStopPhaseGameEnd, g.GetPhase())
	assert.Equal(t, 0, g.GetWinner())
	assert.Equal(t, domain.GoStopResultWin, g.GetResult())
	assert.False(t, g.IsHumanTurn())
}

func TestGoStopNextRound_Advances(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.SetPhase(domain.GoStopPhaseGoDecision)
	g.GetPlayer(0).AddCaptured(gostopGwang3())
	require.NoError(t, g.PlayerDecide(false))
	require.Equal(t, domain.GoStopPhaseRoundEnd, g.GetPhase())
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, domain.GoStopPhasePlay, g.GetPhase())
}

func TestGoStopCpuPlay_TwoMatchFieldChoice(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	g.SetCurrentTurn(1)
	cpu := g.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(gostopCard(3, 1))
	g.SetFieldCards([]*domain.Card{gostopCard(3, 2), gostopCard(3, 3), gostopCard(7, 4)})
	g.CpuPlay()
	assert.GreaterOrEqual(t, cpu.CapturedCount(), 2)
}

func TestGoStopCpuPlay_ThreeMatchSweep(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyHard)
	g.SetCurrentTurn(1)
	cpu := g.GetPlayer(1)
	cpu.Reset()
	cpu.AddCard(gostopCard(5, 1))
	g.SetFieldCards([]*domain.Card{gostopCard(5, 2), gostopCard(5, 3), gostopCard(5, 4)})
	g.CpuPlay()
	assert.GreaterOrEqual(t, cpu.CapturedCount(), 4)
}

func TestGoStopCpuGuards(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	g.CpuPlay()
	g.SetPhase(domain.GoStopPhaseGoDecision)
	g.CpuDecide()
	g.SetPhase(domain.GoStopPhaseRoundEnd)
	g.CpuPlay()
	assert.Equal(t, domain.GoStopPhaseRoundEnd, g.GetPhase())
}

func TestGoStopCaptureOptions(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	human := g.GetPlayer(0)
	human.Reset()
	human.AddCard(gostopCard(3, 1))
	human.AddCard(gostopCard(12, 1))
	g.SetFieldCards([]*domain.Card{gostopCard(3, 2), gostopCard(3, 3)})
	opts := g.GetCaptureOptions(0)
	assert.Contains(t, opts, 0)
	assert.NotContains(t, opts, 1)
	g.SetPhase(domain.GoStopPhaseRoundEnd)
	assert.Empty(t, g.GetCaptureOptions(0))
	g.SetPhase(domain.GoStopPhasePlay)
	assert.Empty(t, g.GetCaptureOptions(99))
}

func TestGoStopPlayableIndices(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	assert.Len(t, g.GetPlayableIndices(0), domain.GoStopHandSize)
	assert.Nil(t, g.GetPlayableIndices(99))
	g.SetPhase(domain.GoStopPhaseRoundEnd)
	assert.Nil(t, g.GetPlayableIndices(0))
}

func TestGoStopAccessors(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
	bd, pts := g.GetScore(99)
	assert.Nil(t, bd)
	assert.Equal(t, 0, pts)
	assert.Equal(t, domain.GoStopCpuDifficultyNormal, g.GetConfig().CpuDifficulty)
	assert.NotNil(t, g.GetActionLog())
	assert.Equal(t, -1, g.GetWinner())
	assert.Equal(t, -1, g.GetRoundWinner())
	assert.Equal(t, domain.GoStopResultNone, g.GetResult())
	assert.Nil(t, g.GetPendingBreakdown())
}

func TestGoStopCardAccessors(t *testing.T) {
	assert.NotEmpty(t, domain.GoStopCardGlyph(gsCrane))
	assert.Equal(t, domain.GoStopGwang, domain.GoStopCardCategory(gsCrane))
	assert.Equal(t, domain.GoStopRibbonBlue, domain.GoStopCardRibbonColor(gostopCard(6, 2)))
	assert.Equal(t, domain.GoStopRibbonRedPoetry, domain.GoStopCardRibbonColor(gostopCard(1, 2)))
	assert.NotEmpty(t, domain.GoStopCardLabel(gsCrane))
	assert.Equal(t, domain.GoStopPi, domain.GoStopCardCategory(nil))
	assert.Equal(t, "??", domain.GoStopCardLabel(nil))
}

func TestGoStopIsHumanTurn(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	g.SetCurrentTurn(0)
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentTurn(1)
	assert.False(t, g.IsHumanTurn())
	g.SetCurrentTurn(0)
	g.SetPhase(domain.GoStopPhaseRoundEnd)
	assert.False(t, g.IsHumanTurn())
}

// gostopDrive はドメインだけで CPU/人間を駆動し終局まで進める。
func gostopDrive(t *testing.T, g *domain.GoStop) {
	t.Helper()
	for step := 0; step < 20000 && !g.GetGameEndFlag(); step++ {
		switch g.GetPhase() {
		case domain.GoStopPhasePlay:
			if g.IsHumanTurn() {
				require.NoError(t, g.PlayerPlay(0, -1))
			} else {
				g.CpuPlay()
			}
		case domain.GoStopPhaseGoDecision:
			if g.IsHumanTurn() {
				require.NoError(t, g.PlayerDecide(false)) // 人間は常にストップ
			} else {
				g.CpuDecide()
			}
		case domain.GoStopPhaseRoundEnd:
			g.NextRound()
		case domain.GoStopPhaseGameEnd:
			return
		}
	}
}

func TestGoStopFullGame_Normal(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	gostopDrive(t, g)
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.GoStopPhaseGameEnd, g.GetPhase())
}

func TestGoStopFullGame_Hard(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyHard)
	gostopDrive(t, g)
	assert.True(t, g.GetGameEndFlag())
	assert.GreaterOrEqual(t, g.GetPlayer(0).GetScore(), 0)
	assert.GreaterOrEqual(t, g.GetPlayer(1).GetScore(), 0)
}

func TestGoStopFullGame_Easy(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyEasy)
	gostopDrive(t, g)
	assert.True(t, g.GetGameEndFlag())
}

// --- Config ---

func TestGoStopConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultGoStopConfig().Validate())
	assert.Error(t, domain.GoStopConfig{CpuDifficulty: 99, TargetScore: 7}.Validate())
	assert.Error(t, domain.GoStopConfig{CpuDifficulty: 0, TargetScore: 0}.Validate())
	assert.Error(t, domain.GoStopConfig{CpuDifficulty: 0, TargetScore: 99999}.Validate())
}

// --- JSON ---

func TestGoStopJSON_RoundTrip(t *testing.T) {
	g := newTestGoStop(t, domain.GoStopCpuDifficultyNormal)
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored domain.GoStop
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, len(g.GetFieldCards()), len(restored.GetFieldCards()))
	assert.Equal(t, g.GetRemainingDeck(), restored.GetRemainingDeck())
}

func TestGoStopJSON_Invalid(t *testing.T) {
	var g domain.GoStop
	assert.Error(t, json.Unmarshal([]byte("not json"), &g))
	assert.Error(t, json.Unmarshal([]byte(`{"pl":[],"cf":{"cd":0,"ts":7},"ph":0,"ct":0,"rw":-1,"wn":-1}`), &g))
	assert.Error(t, json.Unmarshal([]byte(`{"pl":[{"gp":{}},{"gp":{}}],"cf":{"cd":0,"ts":7},"ph":9,"ct":0,"rw":-1,"wn":-1}`), &g))
	assert.Error(t, json.Unmarshal([]byte(`{"pl":[{"gp":{}},{"gp":{}}],"cf":{"cd":0,"ts":7},"ph":0,"ct":0,"rw":-1,"wn":-1,"fd":[{"d":99,"v":1}]}`), &g))
}

func TestGoStopJSON_MoreInvalid(t *testing.T) {
	valid2 := `"pl":[{"gp":{}},{"gp":{}}]`
	base := func(tail string) string {
		return `{` + valid2 + `,"cf":{"cd":0,"ts":7},"ph":0,"ct":0,"rw":-1,"wn":-1` + tail + `}`
	}
	cases := []string{
		`{` + valid2 + `,"cf":{"cd":99,"ts":7},"ph":0,"ct":0,"rw":-1,"wn":-1}`,
		`{"pl":[null,null],"cf":{"cd":0,"ts":7},"ph":0,"ct":0,"rw":-1,"wn":-1}`,
		`{` + valid2 + `,"cf":{"cd":0,"ts":7},"ph":0,"ct":9,"rw":-1,"wn":-1}`,
		`{` + valid2 + `,"cf":{"cd":0,"ts":7},"ph":0,"ct":0,"rw":9,"wn":-1}`,
		`{` + valid2 + `,"cf":{"cd":0,"ts":7},"ph":0,"ct":0,"rw":-1,"wn":9}`,
		base(`,"dp":[{"d":0,"v":0}]`),
		base(`,"fd":[null]`),
	}
	for i, c := range cases {
		var g domain.GoStop
		assert.Errorf(t, json.Unmarshal([]byte(c), &g), "case %d should error", i)
	}

	// oversized slice (actionLog > gostopMaxSliceLen)
	var sb strings.Builder
	sb.WriteString(`{` + valid2 + `,"cf":{"cd":0,"ts":7},"ph":0,"ct":0,"rw":-1,"wn":-1,"al":[`)
	for i := 0; i < 1001; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(`{}`)
	}
	sb.WriteString(`]}`)
	var g domain.GoStop
	assert.Error(t, json.Unmarshal([]byte(sb.String()), &g))
}

func TestGoStopPlayerJSON(t *testing.T) {
	p := domain.NewGoStopPlayer(true)
	p.AddCaptured([]*domain.Card{gsCrane})
	p.AddScore(7)
	p.IncGoCount()
	p.SetCalledGo(true)
	p.SetLastScorePoints(3)
	data, err := json.Marshal(p)
	require.NoError(t, err)
	var restored domain.GoStopPlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, 7, restored.GetScore())
	assert.Equal(t, 1, restored.GetGoCount())
	assert.True(t, restored.GetCalledGo())
	assert.Equal(t, 3, restored.GetLastScorePoints())
	assert.Equal(t, 1, restored.CapturedCount())
}

// gostopGolden mirrors frontend/src/utils/__fixtures__/gostopNearYaku.golden.json.
type gostopGolden struct {
	Cases []struct {
		Name   string `json:"name"`
		Counts struct {
			Bright, Ribbon, Animal, Pi int
		} `json:"counts"`
		Near []struct {
			Category  string `json:"category"`
			Target    string `json:"target"`
			Current   int    `json:"current"`
			Remaining int    `json:"remaining"`
		} `json:"near"`
	} `json:"cases"`
}

// #5710: 「あと何枚でどの役か」は Go/Stop の判断材料そのもので、Web は
// computeNearYaku で出している。CUI 側にも同じ判定が要るので、両者を同じ
// golden vector で縛る。
func TestGoStopComputeNearYaku_GoldenVectors(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(
		"..", "..", "frontend", "src", "utils", "__fixtures__", "gostopNearYaku.golden.json"))
	require.NoError(t, err)
	var golden gostopGolden
	require.NoError(t, json.Unmarshal(raw, &golden))
	require.NotEmpty(t, golden.Cases)

	sawEmpty, sawMultiple := false, false
	for _, c := range golden.Cases {
		t.Run(c.Name, func(t *testing.T) {
			got := domain.GoStopComputeNearYaku(&domain.GoStopBreakdown{
				BrightCount: c.Counts.Bright,
				RibbonCount: c.Counts.Ribbon,
				AnimalCount: c.Counts.Animal,
				PiCount:     c.Counts.Pi,
			})

			require.Len(t, got, len(c.Near))
			for i, want := range c.Near {
				assert.Equal(t, want.Category, got[i].Category)
				assert.Equal(t, want.Target, got[i].Target)
				assert.Equal(t, want.Current, got[i].Current)
				assert.Equal(t, want.Remaining, got[i].Remaining)
			}
		})
		if len(c.Near) == 0 {
			sawEmpty = true
		}
		if len(c.Near) > 1 {
			sawMultiple = true
		}
	}
	// 負のコントロール: 何も近くない case と複数同時の case が要る。
	assert.True(t, sawEmpty, "the golden vectors must include a hand with nothing near")
	assert.True(t, sawMultiple, "the golden vectors must include several near yaku at once")
}

// 内訳が無い局面 (まだ得点していない) では何も返さない。
func TestGoStopComputeNearYaku_NilBreakdown(t *testing.T) {
	assert.Empty(t, domain.GoStopComputeNearYaku(nil))
}
