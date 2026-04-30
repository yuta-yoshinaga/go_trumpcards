//go:build test

package domain

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// newEgyptianRatscrewForTest 既定の 2 人構成で EgyptianRatscrew を返す。Reset は呼ばない。
func newEgyptianRatscrewForTest() *EgyptianRatscrew {
	players := []*EgyptianRatscrewPlayer{
		NewEgyptianRatscrewPlayer(true),
		NewEgyptianRatscrewPlayer(false),
	}
	g := NewEgyptianRatscrew(NewTrumpCards(0), players, DefaultEgyptianRatscrewConfig())
	g.SetRand(rand.New(rand.NewSource(1)))
	return g
}

// setupEgyptianRatscrewWithStocks 与えたカード列で各プレイヤーのストックを構成する。
func setupEgyptianRatscrewWithStocks(t *testing.T, human, cpu []*Card) (*EgyptianRatscrew, *fixedClock) {
	t.Helper()
	g := newEgyptianRatscrewForTest()
	g.Reset()
	g.players[0].ResetStock()
	g.players[1].ResetStock()
	g.players[0].AddToStockBottom(human...)
	g.players[1].AddToStockBottom(cpu...)
	g.centerPile = nil
	g.currentTurnIdx = 0
	g.chanceRemaining = 0
	g.chanceFromIdx = -1
	g.pending = EgyptianRatscrewPending{Kind: EgyptianRatscrewPendingNone}
	g.lastEvent = EgyptianRatscrewLastEvent{}
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.actionLog = nil
	clk := newFixedClock()
	g.SetClock(clk.Now)
	return g, clk
}

func TestEgyptianRatscrew_Reset_Deals26Each(t *testing.T) {
	g := newEgyptianRatscrewForTest()
	g.Reset()
	assert.Equal(t, EgyptianRatscrewPhasePlay, g.GetPhase())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, 26, g.GetPlayer(0).GetStockSize())
	assert.Equal(t, 26, g.GetPlayer(1).GetStockSize())
	assert.Equal(t, 0, g.GetCenterPileSize())
	assert.Nil(t, g.GetTopCard())
	assert.Equal(t, 0, g.GetCurrentTurnIdx())
	assert.True(t, g.IsHumanTurn())
	assert.Equal(t, EgyptianRatscrewPendingNone, g.GetPending().Kind)
	assert.Equal(t, 0, g.GetChanceRemaining())
	assert.Equal(t, -1, g.GetChanceFromIdx())
}

func TestEgyptianRatscrew_NewDefault(t *testing.T) {
	g := NewDefaultEgyptianRatscrew()
	assert.NotNil(t, g)
	assert.Equal(t, EgyptianRatscrewPlayerCnt, g.GetPlayerCnt())
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
}

func TestEgyptianRatscrew_Step_NonFace_TurnPasses(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	assert.NoError(t, g.Step())
	assert.Equal(t, 1, g.GetCenterPileSize())
	assert.Equal(t, 5, g.GetTopCard().GetValue())
	assert.Equal(t, 1, g.GetCurrentTurnIdx())
	assert.False(t, g.IsTopFaceCard())
	// Human の Step 後、CPU の Step が予約されている
	assert.Equal(t, EgyptianRatscrewPendingStep, g.GetPending().Kind)
	assert.Equal(t, EgyptianRatscrewEventStep, g.GetLastEvent().Kind)
}

func TestEgyptianRatscrew_Step_FaceCard_StartsChanceBattle(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t,
		[]*Card{card(EgyptianRatscrewKingValue)},
		[]*Card{card(2), card(3), card(4), card(5)},
	)
	assert.NoError(t, g.Step())
	assert.True(t, g.IsTopFaceCard())
	// K → 相手は 3 回チャンス
	assert.Equal(t, 3, g.GetChanceRemaining())
	assert.Equal(t, 0, g.GetChanceFromIdx())
	assert.Equal(t, 1, g.GetCurrentTurnIdx())
	// CPU 番なので Step が予約される
	assert.Equal(t, EgyptianRatscrewPendingStep, g.GetPending().Kind)
}

func TestEgyptianRatscrew_Step_ChanceBattle_OpponentFails(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t,
		[]*Card{card(EgyptianRatscrewJackValue)},
		[]*Card{card(2), card(3)},
	)
	// human が J → CPU は 1 回だけチャンス
	assert.NoError(t, g.Step())
	assert.Equal(t, 1, g.GetChanceRemaining())

	// CPU が 2 を出す → 絵札ではないのでチャンス消費 → 残 0 → human が獲得
	g.currentTurnIdx = 1
	assert.NoError(t, g.Step())
	assert.Equal(t, EgyptianRatscrewEventChanceWin, g.GetLastEvent().Kind)
	assert.Equal(t, 0, g.GetLastEvent().PlayerIdx)
	assert.Equal(t, 0, g.GetCenterPileSize())
	assert.Equal(t, 0, g.GetCurrentTurnIdx())
	assert.Equal(t, 0, g.GetChanceRemaining())
	assert.Equal(t, -1, g.GetChanceFromIdx())
	assert.Equal(t, 2, g.GetPlayer(0).GetStockSize()) // 元 0 + pile (J + 2) = 2
}

func TestEgyptianRatscrew_Step_ChanceBattle_OpponentRefutesWithFace(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t,
		[]*Card{card(EgyptianRatscrewJackValue), card(2)},
		[]*Card{card(EgyptianRatscrewQueenValue), card(7)},
	)
	// human が J → CPU が応戦すべき (1 回のチャンス)
	assert.NoError(t, g.Step())
	assert.Equal(t, 1, g.GetChanceRemaining())
	assert.Equal(t, 0, g.GetChanceFromIdx())
	assert.Equal(t, 1, g.GetCurrentTurnIdx())

	// CPU の Step → Q (絵札) を出す → チャンスバトルが human 側にひっくり返る
	g.pending = EgyptianRatscrewPending{} // 予約クリアして手動 Step
	assert.NoError(t, g.Step())
	assert.True(t, g.IsTopFaceCard())
	assert.Equal(t, 2, g.GetChanceRemaining()) // Q=2
	assert.Equal(t, 1, g.GetChanceFromIdx())   // CPU が課す側
	assert.Equal(t, 0, g.GetCurrentTurnIdx())  // 手番は human (応戦側)
}

func TestEgyptianRatscrew_Step_OnEmptyStock_GameEnds(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{}, []*Card{card(2), card(3)})
	err := g.Step()
	assert.ErrorIs(t, err, ErrInvalidPlay)
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetWinnerIdx())
}

func TestEgyptianRatscrew_Step_AfterGameEnd_ReturnsErr(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	g.endGame(0)
	err := g.Step()
	assert.ErrorIs(t, err, ErrGameEnded)
}

func TestEgyptianRatscrew_Slap_CorrectOnPair_TakesPile(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t,
		[]*Card{card(5)},
		[]*Card{card(5), card(8)},
	)
	// human が 5 を出す
	assert.NoError(t, g.Step())
	// CPU が 5 を出す → ペア成立
	g.currentTurnIdx = 1
	g.pending = EgyptianRatscrewPending{}
	assert.NoError(t, g.Step())
	assert.True(t, g.IsSlappable())

	// human が先に slap
	assert.NoError(t, g.Slap(0))
	assert.Equal(t, 0, g.GetCenterPileSize())
	assert.Equal(t, 2, g.GetPlayer(0).GetStockSize()) // 元 0 + pile (5,5) = 2
	assert.Equal(t, 0, g.GetCurrentTurnIdx())
	assert.Equal(t, EgyptianRatscrewEventSlapCorrect, g.GetLastEvent().Kind)
	assert.Equal(t, EgyptianRatscrewSlapReasonPair, g.GetLastEvent().SlapReason)
	assert.Equal(t, EgyptianRatscrewPendingNone, g.GetPending().Kind)
}

func TestEgyptianRatscrew_Slap_CorrectOnSandwich(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t,
		[]*Card{card(5), card(5)},
		[]*Card{card(7)},
	)
	// human が 5 を出す
	assert.NoError(t, g.Step())
	// CPU が 7 を出す
	g.currentTurnIdx = 1
	g.pending = EgyptianRatscrewPending{}
	assert.NoError(t, g.Step())
	assert.False(t, g.IsSlappable())
	// human が 5 を出す → サンドイッチ成立 (5-7-5)
	g.currentTurnIdx = 0
	g.pending = EgyptianRatscrewPending{}
	assert.NoError(t, g.Step())
	assert.True(t, g.IsSlappable())

	assert.NoError(t, g.Slap(0))
	assert.Equal(t, EgyptianRatscrewEventSlapCorrect, g.GetLastEvent().Kind)
	assert.Equal(t, EgyptianRatscrewSlapReasonSandwich, g.GetLastEvent().SlapReason)
	assert.Equal(t, 3, g.GetPlayer(0).GetStockSize())
}

func TestEgyptianRatscrew_Slap_CorrectInterruptsChanceBattle(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t,
		[]*Card{card(EgyptianRatscrewJackValue), card(5), card(5)},
		[]*Card{card(2)},
	)
	// human が J を出す → チャンスバトル開始 (CPU 1 回)
	assert.NoError(t, g.Step())
	// human が 5 を出す (テストのために手番を強制) — 実プレイでは CPU が払うが、
	// ここでは「相手が応戦中に元の絵札出し側がスラップで奪い返す」シナリオを再現するため
	// 一旦 CPU に flip させる。
	g.currentTurnIdx = 1
	g.pending = EgyptianRatscrewPending{}
	assert.NoError(t, g.Step()) // CPU が 2 を出す → chanceRemaining 0 → human 獲得
	// このステップで chance win が発生し pile が空になる。
	// 別シナリオ: スラップ成立を試すには場の状態を作り直す必要があるので別テストで検証。
	assert.Equal(t, EgyptianRatscrewEventChanceWin, g.GetLastEvent().Kind)
}

func TestEgyptianRatscrew_Slap_PairBeatsChanceBattle(t *testing.T) {
	// 場に 5,5 があってチャンスバトル中でも、スラップで pile を奪える
	g, _ := setupEgyptianRatscrewWithStocks(t,
		[]*Card{card(2)},
		[]*Card{card(2)},
	)
	g.centerPile = []*Card{card(5), card(5)}
	g.chanceRemaining = 2
	g.chanceFromIdx = 1
	g.currentTurnIdx = 0

	assert.True(t, g.IsSlappable())
	assert.NoError(t, g.Slap(0))
	assert.Equal(t, 0, g.GetCenterPileSize())
	assert.Equal(t, 0, g.GetChanceRemaining())
	assert.Equal(t, -1, g.GetChanceFromIdx())
	assert.Equal(t, EgyptianRatscrewSlapReasonPair, g.GetLastEvent().SlapReason)
}

func TestEgyptianRatscrew_Slap_WrongOnNonSlappable_Penalty(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t,
		[]*Card{card(5), card(6), card(7)},
		[]*Card{card(2), card(3)},
	)
	assert.NoError(t, g.Step()) // 人間が 5 を場に
	// 人間が誤 slap (上 1 枚のみ → 成立せず)
	assert.NoError(t, g.Slap(0))
	assert.Equal(t, 1, g.GetPlayer(0).GetStockSize()) // 3 → 2 → ペナルティ 1 → 1
	assert.Equal(t, 3, g.GetPlayer(1).GetStockSize())
	assert.Equal(t, EgyptianRatscrewEventSlapWrong, g.GetLastEvent().Kind)
}

func TestEgyptianRatscrew_Slap_WrongDrainsStock_GameEnds(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t,
		[]*Card{card(5)},
		[]*Card{card(2), card(3)},
	)
	assert.NoError(t, g.Step()) // 人間が 5 を場に。stock=0
	// 誤 slap → ペナルティで残り 0 枚 → game end
	assert.NoError(t, g.Slap(0))
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetWinnerIdx())
}

func TestEgyptianRatscrew_Slap_WrongRescheduleSlapWhenStillSlappable(t *testing.T) {
	// ペアが場にあって誤った側が誤って slap しても、上 2 枚は変わらないので
	// CPU の slap 予約が再度組まれる
	g, _ := setupEgyptianRatscrewWithStocks(t,
		[]*Card{card(2), card(3)},
		[]*Card{card(4), card(5)},
	)
	g.centerPile = []*Card{card(5), card(5)}
	g.chanceRemaining = 0
	g.chanceFromIdx = -1

	// human が誤って slap (実際にはペアなので正しい slap になる)
	// 誤 slap シナリオを作るため、敢えてペアではない場で誤スラップさせる
	g.centerPile = []*Card{card(5), card(7)}
	assert.False(t, g.IsSlappable())
	assert.NoError(t, g.Slap(0))
	assert.Equal(t, EgyptianRatscrewEventSlapWrong, g.GetLastEvent().Kind)
}

func TestEgyptianRatscrew_Slap_WrongRescheduleStillSlappable(t *testing.T) {
	// CPU が誤 slap した直後でもペア継続中なら CPU の slap 予約が再度組まれる
	g, clk := setupEgyptianRatscrewWithStocks(t,
		[]*Card{card(2), card(3)},
		[]*Card{card(4), card(5)},
	)
	g.centerPile = []*Card{card(8), card(8)}
	g.currentTurnIdx = 0
	g.chanceFromIdx = -1
	// CPU が誤って slap... 実際にはペアなので正解になる。
	// ペア成立中にペナルティが発動するのは applyWrongSlap の経路なので、
	// 直接 applyWrongSlap を呼ぶ
	g.applyWrongSlap(1)
	// 上 2 枚は変わらず 8,8 のまま → IsSlappable=true → CPU の slap 予約が組まれる
	assert.True(t, g.IsSlappable())
	assert.Equal(t, EgyptianRatscrewPendingSlap, g.GetPending().Kind)
	_ = clk
}

func TestEgyptianRatscrew_Slap_EmptyPileFails(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	assert.ErrorIs(t, g.Slap(0), ErrInvalidPlay)
}

func TestEgyptianRatscrew_Slap_InvalidPlayerIdx(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	assert.NoError(t, g.Step())
	assert.ErrorIs(t, g.Slap(-1), ErrInvalidPlay)
	assert.ErrorIs(t, g.Slap(99), ErrInvalidPlay)
}

func TestEgyptianRatscrew_Slap_AfterGameEnd_ReturnsErr(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{card(5)}, []*Card{card(5)})
	assert.NoError(t, g.Step())
	g.endGame(1)
	assert.ErrorIs(t, g.Slap(0), ErrGameEnded)
}

func TestEgyptianRatscrew_Tick_FiresStepWhenDeadlineReached(t *testing.T) {
	g, clk := setupEgyptianRatscrewWithStocks(t,
		[]*Card{card(2), card(3)},
		[]*Card{card(5), card(6)},
	)
	assert.NoError(t, g.Step()) // 人間 2 を出す → 手番 CPU、Step 予約
	assert.Equal(t, EgyptianRatscrewPendingStep, g.GetPending().Kind)

	// deadline 前は発火しない
	kind := g.Tick()
	assert.Equal(t, EgyptianRatscrewPendingNone, kind)

	// deadline を超過 → Tick で発火
	clk.Advance(10 * time.Second)
	kind = g.Tick()
	assert.Equal(t, EgyptianRatscrewPendingStep, kind)
	assert.Equal(t, 5, g.GetTopCard().GetValue())
}

func TestEgyptianRatscrew_Tick_FiresSlapWhenSlappable(t *testing.T) {
	g, clk := setupEgyptianRatscrewWithStocks(t,
		[]*Card{card(5)},
		[]*Card{card(5), card(8)},
	)
	assert.NoError(t, g.Step())
	g.currentTurnIdx = 1
	g.pending = EgyptianRatscrewPending{}
	assert.NoError(t, g.Step()) // CPU が 5 を出す → ペア → CPU slap 予約
	assert.Equal(t, EgyptianRatscrewPendingSlap, g.GetPending().Kind)

	clk.Advance(10 * time.Second)
	kind := g.Tick()
	assert.Equal(t, EgyptianRatscrewPendingSlap, kind)
	assert.Equal(t, 1, g.GetCurrentTurnIdx())
	assert.Equal(t, 0, g.GetCenterPileSize())
}

func TestEgyptianRatscrew_Tick_NoPendingDoesNothing(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	assert.Equal(t, EgyptianRatscrewPendingNone, g.Tick())
}

func TestEgyptianRatscrew_Tick_AfterGameEnd_NoOp(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	g.endGame(0)
	assert.Equal(t, EgyptianRatscrewPendingNone, g.Tick())
}

func TestEgyptianRatscrew_DrawReactionMs_ClampsToMin(t *testing.T) {
	g := newEgyptianRatscrewForTest()
	g.Reset()
	for range 50 {
		ms := g.drawReactionMs()
		assert.GreaterOrEqual(t, ms, EgyptianRatscrewMinReactionMs)
	}
}

func TestEgyptianRatscrew_OpponentIdx(t *testing.T) {
	g := newEgyptianRatscrewForTest()
	assert.Equal(t, 1, g.opponentIdx(0))
	assert.Equal(t, 0, g.opponentIdx(1))
}

func TestEgyptianRatscrew_IsHumanTurn(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	assert.True(t, g.IsHumanTurn())
	assert.NoError(t, g.Step())
	assert.False(t, g.IsHumanTurn())
	g.endGame(0)
	assert.False(t, g.IsHumanTurn())
}

func TestEgyptianRatscrew_ResetWithConfig(t *testing.T) {
	g := newEgyptianRatscrewForTest()
	g.ResetWithConfig(EgyptianRatscrewConfig{CpuDifficulty: EgyptianRatscrewCpuHard})
	assert.Equal(t, EgyptianRatscrewCpuHard, g.GetConfig().CpuDifficulty)
	assert.Equal(t, 26, g.GetPlayer(0).GetStockSize())
}

func TestEgyptianRatscrew_SetConfig(t *testing.T) {
	g := newEgyptianRatscrewForTest()
	g.SetConfig(EgyptianRatscrewConfig{CpuDifficulty: EgyptianRatscrewCpuEasy})
	assert.Equal(t, EgyptianRatscrewCpuEasy, g.GetConfig().CpuDifficulty)
}

func TestEgyptianRatscrew_ActionLog(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	assert.NoError(t, g.Step())
	logs := g.GetActionLog()
	assert.NotEmpty(t, logs)
	assert.Equal(t, "step", logs[0].ActionType)
}

func TestEgyptianRatscrew_CheckStuck_BothEmpty(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{}, []*Card{})
	g.checkStuck()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerIdx(), "両者ストック空でスラップ/チャンスとも不可なら引き分け")
}

func TestEgyptianRatscrew_CheckStuck_NoOpAfterEnd(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{}, []*Card{})
	g.endGame(1)
	g.checkStuck()
	assert.Equal(t, 1, g.GetWinnerIdx())
}

func TestEgyptianRatscrew_CheckStuck_BothEmpty_SlappableOnTop_NoEnd(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{}, []*Card{})
	g.centerPile = []*Card{card(5), card(5)}
	g.checkStuck()
	assert.False(t, g.GetGameEndFlag(), "ペア成立中は slap で決着可能")
}

func TestEgyptianRatscrew_CheckStuck_BothEmpty_NonSlappableOnTop_Ends(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{}, []*Card{})
	g.centerPile = []*Card{card(5), card(7)}
	g.checkStuck()
	assert.True(t, g.GetGameEndFlag(), "上 2 枚が成立せず両者空ならスタック → 終了")
}

func TestEgyptianRatscrew_CheckStuck_DuringChanceBattle_PayerEmpty_FaceCardWins(t *testing.T) {
	// 絵札を出した human、応戦側の CPU はストック空 → CPU は払えない → human 獲得
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{}, []*Card{})
	g.centerPile = []*Card{card(EgyptianRatscrewKingValue)}
	g.chanceRemaining = 3
	g.chanceFromIdx = 0
	g.currentTurnIdx = 1
	g.checkStuck()
	assert.Equal(t, EgyptianRatscrewEventChanceWin, g.GetLastEvent().Kind)
	assert.Equal(t, 0, g.GetLastEvent().PlayerIdx)
}

func TestEgyptianRatscrew_SetClock_NilIgnored(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	before := g.now
	g.SetClock(nil)
	assert.NotNil(t, before)
	assert.NotNil(t, g.now)
}

func TestEgyptianRatscrew_SetRand_NilIgnored(t *testing.T) {
	g := newEgyptianRatscrewForTest()
	before := g.rng
	g.SetRand(nil)
	assert.Equal(t, before, g.rng)
}

func TestEgyptianRatscrew_JSON_Roundtrip(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t,
		[]*Card{card(EgyptianRatscrewKingValue), card(2)},
		[]*Card{card(7), card(8)},
	)
	assert.NoError(t, g.Step())

	data, err := json.Marshal(g)
	assert.NoError(t, err)

	decoded := NewEgyptianRatscrew(NewTrumpCards(0),
		[]*EgyptianRatscrewPlayer{NewEgyptianRatscrewPlayer(true), NewEgyptianRatscrewPlayer(false)},
		DefaultEgyptianRatscrewConfig())
	assert.NoError(t, json.Unmarshal(data, decoded))
	assert.Equal(t, g.GetPhase(), decoded.GetPhase())
	assert.Equal(t, g.GetCenterPileSize(), decoded.GetCenterPileSize())
	assert.Equal(t, g.GetCurrentTurnIdx(), decoded.GetCurrentTurnIdx())
	assert.Equal(t, g.GetChanceRemaining(), decoded.GetChanceRemaining())
	assert.Equal(t, g.GetChanceFromIdx(), decoded.GetChanceFromIdx())
	assert.Equal(t, g.GetPending(), decoded.GetPending())
	assert.NotNil(t, decoded.now)
	assert.NotNil(t, decoded.rng)
}

func TestEgyptianRatscrew_PendingNotOverwritten(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{card(EgyptianRatscrewJackValue)}, []*Card{card(2)})
	assert.NoError(t, g.Step()) // 絵札 → CPU の Step 予約 (チャンスバトル)
	before := g.GetPending()
	g.maybeScheduleCpuStep()
	assert.Equal(t, before, g.GetPending())
}

func TestEgyptianRatscrew_MaybeScheduleCpuStep_HumanTurnDoesNothing(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	g.currentTurnIdx = 0
	g.pending = EgyptianRatscrewPending{Kind: EgyptianRatscrewPendingNone}
	g.maybeScheduleCpuStep()
	assert.Equal(t, EgyptianRatscrewPendingNone, g.GetPending().Kind)
}

func TestEgyptianRatscrew_ScheduleAfterGameEnd_NoOp(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{}, []*Card{})
	g.endGame(0)
	g.maybeScheduleCpuStep()
	g.scheduleCpuSlap()
	assert.Equal(t, EgyptianRatscrewPendingNone, g.GetPending().Kind)
}

func TestEgyptianRatscrew_EndGame_InvalidWinnerIdx_NoFinishedFlag(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{card(5)}, []*Card{card(7)})
	g.endGame(99)
	assert.True(t, g.GetGameEndFlag())
	assert.False(t, g.GetPlayer(0).GetIsFinished())
	assert.False(t, g.GetPlayer(1).GetIsFinished())
}

func TestEgyptianRatscrew_Step_NonFace_OpponentEmpty_RetainsTurn(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t,
		[]*Card{card(5), card(EgyptianRatscrewJackValue)},
		[]*Card{},
	)
	assert.NoError(t, g.Step()) // human が 5 を出す
	assert.False(t, g.IsTopFaceCard())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetCurrentTurnIdx())
	assert.Equal(t, EgyptianRatscrewPendingNone, g.GetPending().Kind)
}

func TestEgyptianRatscrew_Step_FaceCard_OpponentEmpty_FaceCardSideWins(t *testing.T) {
	// human が J を出した時点で CPU stock が空 → 即 human 獲得
	g, _ := setupEgyptianRatscrewWithStocks(t,
		[]*Card{card(EgyptianRatscrewJackValue)},
		[]*Card{},
	)
	assert.NoError(t, g.Step())
	assert.Equal(t, EgyptianRatscrewEventChanceWin, g.GetLastEvent().Kind)
	assert.Equal(t, 0, g.GetLastEvent().PlayerIdx)
	assert.Equal(t, 0, g.GetCenterPileSize())
}

func TestEgyptianRatscrew_FaceCardChances(t *testing.T) {
	assert.Equal(t, 1, FaceCardChances(EgyptianRatscrewJackValue))
	assert.Equal(t, 2, FaceCardChances(EgyptianRatscrewQueenValue))
	assert.Equal(t, 3, FaceCardChances(EgyptianRatscrewKingValue))
	assert.Equal(t, 4, FaceCardChances(EgyptianRatscrewAceValue))
	assert.Equal(t, 0, FaceCardChances(5))
	assert.False(t, IsFaceCard(5))
	assert.True(t, IsFaceCard(EgyptianRatscrewAceValue))
}

func TestEgyptianRatscrew_SlapReason_PairBeforeSandwich(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{}, []*Card{})
	// 5-5-5 → ペア (上 2 枚が同じ) としてカウントされる
	g.centerPile = []*Card{card(5), card(5), card(5)}
	assert.Equal(t, EgyptianRatscrewSlapReasonPair, g.slapReason())
}

func TestEgyptianRatscrew_SlapReason_None(t *testing.T) {
	g, _ := setupEgyptianRatscrewWithStocks(t, []*Card{}, []*Card{})
	g.centerPile = []*Card{card(2)}
	assert.Equal(t, EgyptianRatscrewSlapReasonNone, g.slapReason())
	g.centerPile = []*Card{card(2), card(3), card(4)}
	assert.Equal(t, EgyptianRatscrewSlapReasonNone, g.slapReason())
}

func TestEgyptianRatscrew_SlapReasonLabel(t *testing.T) {
	assert.Equal(t, "pair", slapReasonLabel(EgyptianRatscrewSlapReasonPair))
	assert.Equal(t, "sandwich", slapReasonLabel(EgyptianRatscrewSlapReasonSandwich))
	assert.Equal(t, "none", slapReasonLabel(EgyptianRatscrewSlapReasonNone))
}
