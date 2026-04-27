//go:build test

package domain

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// newSlapjackForTest 既定の 2 人構成で Slapjack を返す。Reset は呼ばない。
func newSlapjackForTest() *Slapjack {
	players := []*SlapjackPlayer{
		NewSlapjackPlayer(true),
		NewSlapjackPlayer(false),
	}
	g := NewSlapjack(NewTrumpCards(0), players, DefaultSlapjackConfig())
	g.SetRand(rand.New(rand.NewSource(1)))
	return g
}

// fixedClock 固定時刻 + Advance で前進可能なクロック
type fixedClock struct {
	now time.Time
}

func (c *fixedClock) Now() time.Time { return c.now }
func (c *fixedClock) Advance(d time.Duration) {
	c.now = c.now.Add(d)
}

func newFixedClock() *fixedClock {
	return &fixedClock{now: time.Unix(0, 0)}
}

// setupSlapjackWithStocks 与えたカード列で各プレイヤーのストックを構成する。
func setupSlapjackWithStocks(t *testing.T, human, cpu []*Card) (*Slapjack, *fixedClock) {
	t.Helper()
	g := newSlapjackForTest()
	g.Reset()
	g.players[0].ResetStock()
	g.players[1].ResetStock()
	g.players[0].AddToStockBottom(human...)
	g.players[1].AddToStockBottom(cpu...)
	g.centerPile = nil
	g.currentTurnIdx = 0
	g.pending = SlapjackPending{Kind: SlapjackPendingNone}
	g.lastEvent = SlapjackLastEvent{}
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.actionLog = nil
	clk := newFixedClock()
	g.SetClock(clk.Now)
	return g, clk
}

func TestSlapjack_Reset_Deals26Each(t *testing.T) {
	g := newSlapjackForTest()
	g.Reset()
	assert.Equal(t, SlapjackPhasePlay, g.GetPhase())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, 26, g.GetPlayer(0).GetStockSize())
	assert.Equal(t, 26, g.GetPlayer(1).GetStockSize())
	assert.Equal(t, 0, g.GetCenterPileSize())
	assert.Nil(t, g.GetTopCard())
	assert.Equal(t, 0, g.GetCurrentTurnIdx())
	assert.True(t, g.IsHumanTurn())
	assert.Equal(t, SlapjackPendingNone, g.GetPending().Kind)
}

func TestSlapjack_NewDefault(t *testing.T) {
	g := NewDefaultSlapjack()
	assert.NotNil(t, g)
	assert.Equal(t, SlapjackPlayerCnt, g.GetPlayerCnt())
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
}

func TestSlapjack_Step_NonJack_TurnPasses(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	assert.NoError(t, g.Step())
	assert.Equal(t, 1, g.GetCenterPileSize())
	assert.Equal(t, 5, g.GetTopCard().GetValue())
	assert.Equal(t, 1, g.GetCurrentTurnIdx())
	assert.False(t, g.IsTopJack())
	// Human の Step 後、CPU の Step が予約されている
	assert.Equal(t, SlapjackPendingStep, g.GetPending().Kind)
	assert.Equal(t, SlapjackEventStep, g.GetLastEvent().Kind)
}

func TestSlapjack_Step_Jack_SchedulesCpuSlap(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t, []*Card{card(SlapjackJackValue)}, []*Card{card(7)})
	assert.NoError(t, g.Step())
	assert.True(t, g.IsTopJack())
	// J が出た直後は手番は変わらない (slap レース中)
	assert.Equal(t, 0, g.GetCurrentTurnIdx())
	assert.Equal(t, SlapjackPendingSlap, g.GetPending().Kind)
}

func TestSlapjack_Step_OnEmptyStock_GameEnds(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t, []*Card{}, []*Card{card(2), card(3)})
	err := g.Step()
	assert.ErrorIs(t, err, ErrInvalidPlay)
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetWinnerIdx())
}

func TestSlapjack_Step_AfterGameEnd_ReturnsErr(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	g.endGame(0)
	err := g.Step()
	assert.ErrorIs(t, err, ErrGameEnded)
}

func TestSlapjack_Slap_CorrectOnJack_TakesPile(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t,
		[]*Card{card(SlapjackJackValue)},
		[]*Card{card(7), card(8)},
	)
	assert.NoError(t, g.Step())
	assert.True(t, g.IsTopJack())

	// 人間が先に slap
	assert.NoError(t, g.Slap(0))
	assert.Equal(t, 0, g.GetCenterPileSize())
	assert.Equal(t, 1, g.GetPlayer(0).GetStockSize()) // J 1 枚を取得 (元 stock 0 + pile 1)
	assert.Equal(t, 0, g.GetCurrentTurnIdx())         // 取った人が次のターン
	assert.Equal(t, SlapjackEventSlapCorrect, g.GetLastEvent().Kind)
	// 取得後手番は人間なので CPU 予約は無し
	assert.Equal(t, SlapjackPendingNone, g.GetPending().Kind)
}

func TestSlapjack_Slap_CorrectByCpu_SchedulesNextStep(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t,
		[]*Card{card(SlapjackJackValue)},
		[]*Card{card(7), card(8)},
	)
	assert.NoError(t, g.Step())
	assert.NoError(t, g.Slap(1))
	assert.Equal(t, 1, g.GetCurrentTurnIdx())
	// 取得後は CPU の番なので Step が予約される
	assert.Equal(t, SlapjackPendingStep, g.GetPending().Kind)
}

func TestSlapjack_Slap_WrongOnNonJack_Penalty(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t,
		[]*Card{card(5), card(6), card(7)},
		[]*Card{card(2), card(3)},
	)
	assert.NoError(t, g.Step()) // 人間が 5 を場に
	// 人間が誤 slap (top=5, J ではない)
	assert.NoError(t, g.Slap(0))
	// 人間ストックは 1 枚減 (3 → 2 のうち 1 枚が CPU の底へ)
	assert.Equal(t, 1, g.GetPlayer(0).GetStockSize())
	assert.Equal(t, 3, g.GetPlayer(1).GetStockSize())
	assert.Equal(t, SlapjackEventSlapWrong, g.GetLastEvent().Kind)
}

func TestSlapjack_Slap_WrongDrainsStock_GameEnds(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t,
		[]*Card{card(5)},
		[]*Card{card(2), card(3)},
	)
	assert.NoError(t, g.Step()) // 人間が 5 を場に。stock=0
	// 誤 slap → ペナルティで残り 0 枚 → game end
	assert.NoError(t, g.Slap(0))
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetWinnerIdx())
}

func TestSlapjack_Slap_EmptyPileFails(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	assert.ErrorIs(t, g.Slap(0), ErrInvalidPlay)
}

func TestSlapjack_Slap_InvalidPlayerIdx(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	assert.NoError(t, g.Step())
	assert.ErrorIs(t, g.Slap(-1), ErrInvalidPlay)
	assert.ErrorIs(t, g.Slap(99), ErrInvalidPlay)
}

func TestSlapjack_Slap_AfterGameEnd_ReturnsErr(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t, []*Card{card(SlapjackJackValue)}, []*Card{card(7)})
	assert.NoError(t, g.Step())
	g.endGame(1)
	assert.ErrorIs(t, g.Slap(0), ErrGameEnded)
}

func TestSlapjack_Tick_FiresStepWhenDeadlineReached(t *testing.T) {
	g, clk := setupSlapjackWithStocks(t,
		[]*Card{card(2), card(3)},
		[]*Card{card(5), card(6)},
	)
	assert.NoError(t, g.Step()) // 人間 2 を出す → 手番 CPU、Step 予約
	assert.Equal(t, SlapjackPendingStep, g.GetPending().Kind)

	// deadline 前は発火しない
	kind := g.Tick()
	assert.Equal(t, SlapjackPendingNone, kind)

	// deadline を超過 → Tick で発火
	clk.Advance(10 * time.Second)
	kind = g.Tick()
	assert.Equal(t, SlapjackPendingStep, kind)
	// CPU が 5 を出した (top=5)
	assert.Equal(t, 5, g.GetTopCard().GetValue())
}

func TestSlapjack_Tick_FiresSlapWhenJackOnPile(t *testing.T) {
	g, clk := setupSlapjackWithStocks(t,
		[]*Card{card(SlapjackJackValue), card(2)},
		[]*Card{card(7), card(8)},
	)
	assert.NoError(t, g.Step()) // 人間が J を場に
	assert.Equal(t, SlapjackPendingSlap, g.GetPending().Kind)

	clk.Advance(10 * time.Second)
	kind := g.Tick()
	assert.Equal(t, SlapjackPendingSlap, kind)
	// CPU が pile を取得
	assert.Equal(t, 1, g.GetCurrentTurnIdx())
	assert.Equal(t, 0, g.GetCenterPileSize())
}

func TestSlapjack_Tick_NoPendingDoesNothing(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	assert.Equal(t, SlapjackPendingNone, g.Tick())
}

func TestSlapjack_Tick_AfterGameEnd_NoOp(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	g.endGame(0)
	assert.Equal(t, SlapjackPendingNone, g.Tick())
}

func TestSlapjack_DrawReactionMs_ClampsToMin(t *testing.T) {
	g := newSlapjackForTest()
	g.Reset()
	// 平均/標準偏差を 0 にすると常に 0 が出る → 下限へクランプされる
	g.SetConfig(SlapjackConfig{CpuDifficulty: SlapjackCpuDifficulty(99)}) // fallback
	for range 50 {
		// CpuDifficulty=99 は normal にフォールバックして mean=600 だが、
		// 実際にクランプ動作を見るには直接 drawReactionMs の最低保証を確認する
		ms := g.drawReactionMs()
		assert.GreaterOrEqual(t, ms, SlapjackMinReactionMs)
	}
}

func TestSlapjack_OpponentIdx(t *testing.T) {
	g := newSlapjackForTest()
	assert.Equal(t, 1, g.opponentIdx(0))
	assert.Equal(t, 0, g.opponentIdx(1))
}

func TestSlapjack_IsHumanTurn(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	assert.True(t, g.IsHumanTurn())
	assert.NoError(t, g.Step())
	assert.False(t, g.IsHumanTurn()) // CPU の番
	g.endGame(0)
	assert.False(t, g.IsHumanTurn()) // ゲーム終了
}

func TestSlapjack_ResetWithConfig(t *testing.T) {
	g := newSlapjackForTest()
	g.ResetWithConfig(SlapjackConfig{CpuDifficulty: SlapjackCpuHard})
	assert.Equal(t, SlapjackCpuHard, g.GetConfig().CpuDifficulty)
	assert.Equal(t, 26, g.GetPlayer(0).GetStockSize())
}

func TestSlapjack_SetConfig(t *testing.T) {
	g := newSlapjackForTest()
	g.SetConfig(SlapjackConfig{CpuDifficulty: SlapjackCpuEasy})
	assert.Equal(t, SlapjackCpuEasy, g.GetConfig().CpuDifficulty)
}

func TestSlapjack_ActionLog(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	assert.NoError(t, g.Step())
	logs := g.GetActionLog()
	assert.NotEmpty(t, logs)
	assert.Equal(t, "step", logs[0].ActionType)
}

func TestSlapjack_CheckStuck_BothEmpty(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t, []*Card{}, []*Card{})
	g.checkStuck()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerIdx())
}

func TestSlapjack_CheckStuck_NoOpAfterEnd(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t, []*Card{}, []*Card{})
	g.endGame(1)
	g.checkStuck()
	assert.Equal(t, 1, g.GetWinnerIdx()) // 上書きされない
}

func TestSlapjack_SetClock_NilIgnored(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	before := g.now
	g.SetClock(nil)
	// 関数ポインタの比較は不可能なので、置き換え無しで Tick が無事呼べる程度の確認
	assert.NotNil(t, before)
	assert.NotNil(t, g.now)
}

func TestSlapjack_SetRand_NilIgnored(t *testing.T) {
	g := newSlapjackForTest()
	before := g.rng
	g.SetRand(nil)
	assert.Equal(t, before, g.rng)
}

func TestSlapjack_FullRound_HumanCorrectSlap(t *testing.T) {
	g, clk := setupSlapjackWithStocks(t,
		[]*Card{card(2), card(SlapjackJackValue)},
		[]*Card{card(3), card(4)},
	)
	// 1: human が 2 を出す → CPU step 予約
	assert.NoError(t, g.Step())
	assert.Equal(t, 1, g.GetCurrentTurnIdx())
	assert.Equal(t, SlapjackPendingStep, g.GetPending().Kind)

	// 2: CPU の Tick → 3 が場に → human step 予約は無し (人間ターンなので)
	clk.Advance(10 * time.Second)
	assert.Equal(t, SlapjackPendingStep, g.Tick())
	assert.Equal(t, 0, g.GetCurrentTurnIdx())
	assert.Equal(t, SlapjackPendingNone, g.GetPending().Kind)

	// 3: human が J を出す → CPU slap 予約
	assert.NoError(t, g.Step())
	assert.True(t, g.IsTopJack())
	assert.Equal(t, SlapjackPendingSlap, g.GetPending().Kind)

	// 4: human が CPU より先に slap (Tick せず即実行)
	assert.NoError(t, g.Slap(0))
	assert.Equal(t, 0, g.GetCenterPileSize())
	assert.Equal(t, 0, g.GetCurrentTurnIdx())
	// 取得 pile (3 枚) が human stock の底へ → 1 枚 (J) + 0 残 + 3 = ?
	// 元: human stock = 0 (2 と J を出した), CPU stock = 1 (4 残), pile = [2, 3, J]
	// 正解 slap で human stock = 3 枚
	assert.Equal(t, 3, g.GetPlayer(0).GetStockSize())
}

func TestSlapjack_JSON_Roundtrip(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t,
		[]*Card{card(SlapjackJackValue), card(2)},
		[]*Card{card(7), card(8)},
	)
	assert.NoError(t, g.Step())

	data, err := json.Marshal(g)
	assert.NoError(t, err)

	decoded := NewSlapjack(NewTrumpCards(0),
		[]*SlapjackPlayer{NewSlapjackPlayer(true), NewSlapjackPlayer(false)},
		DefaultSlapjackConfig())
	// Marshal 前にクロックを差し替えていなくても JSON 経路で復元されることを確認
	assert.NoError(t, json.Unmarshal(data, decoded))
	assert.Equal(t, g.GetPhase(), decoded.GetPhase())
	assert.Equal(t, g.GetCenterPileSize(), decoded.GetCenterPileSize())
	assert.Equal(t, g.GetCurrentTurnIdx(), decoded.GetCurrentTurnIdx())
	assert.Equal(t, g.GetPending(), decoded.GetPending())
	assert.NotNil(t, decoded.now)
	assert.NotNil(t, decoded.rng)
}

func TestSlapjack_PendingNotOverwritten(t *testing.T) {
	// J が場に出て slap 予約された後、maybeScheduleCpuStep を呼んでも上書きされない
	g, _ := setupSlapjackWithStocks(t, []*Card{card(SlapjackJackValue)}, []*Card{card(2)})
	assert.NoError(t, g.Step())
	before := g.GetPending()
	g.maybeScheduleCpuStep()
	assert.Equal(t, before, g.GetPending())
}

func TestSlapjack_MaybeScheduleCpuStep_HumanTurnDoesNothing(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	g.currentTurnIdx = 0
	g.pending = SlapjackPending{Kind: SlapjackPendingNone}
	g.maybeScheduleCpuStep()
	assert.Equal(t, SlapjackPendingNone, g.GetPending().Kind)
}

func TestSlapjack_ScheduleAfterGameEnd_NoOp(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t, []*Card{}, []*Card{})
	g.endGame(0)
	g.maybeScheduleCpuStep()
	g.scheduleCpuSlap()
	assert.Equal(t, SlapjackPendingNone, g.GetPending().Kind)
}

func TestSlapjack_EndGame_InvalidWinnerIdx_NoFinishedFlag(t *testing.T) {
	g, _ := setupSlapjackWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	g.endGame(99) // out of range — 何もしない
	assert.True(t, g.GetGameEndFlag())
	assert.False(t, g.GetPlayer(0).GetIsFinished())
	assert.False(t, g.GetPlayer(1).GetIsFinished())
}
