//go:build test

package domain

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// kmCard は domain.NewCard の薄いラッパ (テストヘルパ)。
func kmCard(d, v int) *Card { return NewCard(d, v, true) }

// kmSetHand はプレイヤーの手札を明示的に設定する (Reset 後の確定状態を作る)。
func kmSetHand(p *KempsPlayer, cards ...*Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// newTestKemps は決定的乱数源を持つ標準セットアップの Kemps を返す。
func newTestKemps() *Kemps {
	g := NewDefaultKemps()
	g.SetRand(rand.New(rand.NewSource(1)))
	return g
}

// 内部フィールドを直接操作するためのヘルパ (同一パッケージのテスト)。
func kmSetPhase(g *Kemps, ph KempsPhase)   { g.phase = ph }
func kmSetCurrentPlayer(g *Kemps, idx int) { g.currentPlayerIdx = idx }
func kmSetFourHolder(g *Kemps, idx int)    { g.fourHolderIdx = idx }
func kmSetField(g *Kemps, cards ...*Card)  { g.field = cards }
func kmSetGameEnd(g *Kemps)                { g.gameEndFlag = true; g.phase = KempsPhaseGameEnd }

func TestKemps_DefaultConstruction(t *testing.T) {
	g := NewDefaultKemps()
	assert.Equal(t, KempsPlayerCnt, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	for i := 1; i < KempsPlayerCnt; i++ {
		assert.False(t, g.GetPlayer(i).GetIsHuman())
	}
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(KempsPlayerCnt))
}

func TestKemps_TeamHelpers(t *testing.T) {
	assert.Equal(t, 0, KempsTeamOf(0))
	assert.Equal(t, 1, KempsTeamOf(1))
	assert.Equal(t, 0, KempsTeamOf(2))
	assert.Equal(t, 1, KempsTeamOf(3))
	assert.Equal(t, 2, KempsPartnerOf(0))
	assert.Equal(t, 3, KempsPartnerOf(1))
	assert.Equal(t, 0, KempsPartnerOf(2))
}

func TestKemps_ResetDealsHandsAndField(t *testing.T) {
	g := newTestKemps()
	g.Reset()
	assert.Equal(t, KempsPhaseExchange, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, KempsFieldSize, g.GetFieldSize())
	for i := 0; i < KempsPlayerCnt; i++ {
		assert.Equal(t, KempsHandSize, g.GetPlayer(i).GetCardsSize())
	}
	assert.Equal(t, 0, g.GetTeamScore(0))
	assert.Equal(t, 0, g.GetTeamScore(1))
	assert.Equal(t, -1, g.GetWinnerTeam())
	assert.False(t, g.GetGameEndFlag())
}

func TestKemps_HasFourOfAKind(t *testing.T) {
	p := NewKempsPlayer(true)
	kmSetHand(p, kmCard(1, 7), kmCard(2, 7), kmCard(3, 7), kmCard(4, 7))
	assert.True(t, p.HasFourOfAKind())

	kmSetHand(p, kmCard(1, 7), kmCard(2, 7), kmCard(3, 7), kmCard(4, 8))
	assert.False(t, p.HasFourOfAKind())

	kmSetHand(p, kmCard(1, 7), kmCard(2, 7), kmCard(3, 7))
	assert.False(t, p.HasFourOfAKind(), "fewer than 4 cards")
}

func TestKemps_SetSignalAndGetters(t *testing.T) {
	g := newTestKemps()
	g.Reset()
	g.PlayerSetSignal(int(SignalBlink))
	assert.Equal(t, SignalBlink, g.GetSignalType())
	// 範囲外はクランプ。
	g.PlayerSetSignal(99)
	assert.Equal(t, SignalSound, g.GetSignalType())
}

func TestKemps_PlayerSwap(t *testing.T) {
	g := newTestKemps()
	g.Reset()
	kmSetCurrentPlayer(g, 0)
	// 人間の手札とフィールドを確定状態にする。
	kmSetHand(g.GetPlayer(0), kmCard(1, 2), kmCard(2, 3), kmCard(3, 4), kmCard(4, 5))
	kmSetField(g, kmCard(1, 9), kmCard(2, 10), kmCard(3, 11), kmCard(4, 12))
	// **他の席も固定する。** `afterExchange` の `firstFourHolder()` は全席を
	// 走査するので、CPU に配られた手がフォーオブアカインドだと宣言ウィンドウが
	// 開いて手番が進まず、下の「手番が次に進む」が落ちる。`newTestKemps` の
	// `SetRand` は Kemps 自身の rng しか差し替えず、配りは `TrumpCards.Shuffle()`
	// —— つまりグローバルの乱数 —— を通るので、この前提は固定されていなかった。
	for i := 1; i < KempsPlayerCnt; i++ {
		kmSetHand(g.GetPlayer(i), kmCard(1, 3), kmCard(2, 5), kmCard(3, 7), kmCard(4, 9))
	}
	require.Equal(t, -1, g.firstFourHolder(), "誰もフォーオブアカインドを持っていないこと")

	require.NoError(t, g.PlayerSwap(0, 0))
	// 取った 9 が手札に、出した 2 がフィールドに。
	assert.Equal(t, 9, g.GetPlayer(0).GetCard(3).GetValue())
	assert.Equal(t, 2, g.GetFieldCard(0).GetValue())
	// 手番が次に進む。
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
}

func TestKemps_PlayerSwapErrors(t *testing.T) {
	g := newTestKemps()
	g.Reset()
	kmSetCurrentPlayer(g, 0)
	assert.ErrorIs(t, g.PlayerSwap(99, 0), ErrInvalidCard)
	assert.ErrorIs(t, g.PlayerSwap(0, 99), ErrInvalidCard)

	kmSetCurrentPlayer(g, 1)
	assert.ErrorIs(t, g.PlayerSwap(0, 0), ErrInvalidPlay)

	kmSetPhase(g, KempsPhaseDeclare)
	kmSetCurrentPlayer(g, 0)
	assert.ErrorIs(t, g.PlayerSwap(0, 0), ErrWrongPhase)

	kmSetGameEnd(g)
	assert.ErrorIs(t, g.PlayerSwap(0, 0), ErrGameEnded)
}

func TestKemps_PlayerPass(t *testing.T) {
	g := newTestKemps()
	g.Reset()
	kmSetCurrentPlayer(g, 0)
	require.NoError(t, g.PlayerPass())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())

	kmSetCurrentPlayer(g, 1)
	assert.ErrorIs(t, g.PlayerPass(), ErrInvalidPlay)
}

func TestKemps_FourOfAKindOpensDeclareWindow(t *testing.T) {
	g := newTestKemps()
	g.Reset()
	kmSetCurrentPlayer(g, 0)
	// 人間が交換でフォーオブアカインドを完成させる。
	kmSetHand(g.GetPlayer(0), kmCard(1, 7), kmCard(2, 7), kmCard(3, 7), kmCard(4, 8))
	kmSetField(g, kmCard(1, 7), kmCard(2, 2), kmCard(3, 3), kmCard(4, 4))
	require.NoError(t, g.PlayerSwap(3, 0)) // 8 を出して 7 を取る
	assert.Equal(t, KempsPhaseDeclare, g.GetPhase())
	assert.Equal(t, 0, g.GetFourHolderIdx())
	assert.True(t, g.IsPartnerSignaling())
	assert.False(t, g.IsOpponentSignaling())
}

func TestKemps_DeclareKempsSuccess(t *testing.T) {
	g := newTestKemps()
	g.Reset()
	// 人間チーム (席 0) が保持。
	kmSetHand(g.GetPlayer(0), kmCard(1, 7), kmCard(2, 7), kmCard(3, 7), kmCard(4, 7))
	kmSetPhase(g, KempsPhaseDeclare)
	kmSetFourHolder(g, 0)
	require.NoError(t, g.PlayerDeclareKemps())
	assert.Equal(t, 1, g.GetTeamScore(KempsTeamOf(0)))
	assert.Equal(t, KempsResultKemps, g.GetRoundResult())
	assert.Equal(t, KempsPhaseRoundEnd, g.GetPhase())
}

func TestKemps_DeclareKempsMiss(t *testing.T) {
	g := newTestKemps()
	g.Reset()
	// 相手チーム (席 1) が保持しているのに人間が Kemps を宣言 → 空振り。
	kmSetPhase(g, KempsPhaseDeclare)
	kmSetFourHolder(g, 1)
	require.NoError(t, g.PlayerDeclareKemps())
	assert.Equal(t, 0, g.GetTeamScore(KempsTeamOf(0)))
	assert.Equal(t, KempsResultMiss, g.GetRoundResult())
}

func TestKemps_DeclareKempsWrongPhase(t *testing.T) {
	g := newTestKemps()
	g.Reset()
	assert.ErrorIs(t, g.PlayerDeclareKemps(), ErrWrongPhase)
	kmSetGameEnd(g)
	assert.ErrorIs(t, g.PlayerDeclareKemps(), ErrGameEnded)
}

func TestKemps_CounterKempsSuccess(t *testing.T) {
	g := newTestKemps()
	g.Reset()
	// 相手 (席 1) がフォーオブアカインドを保持 → カウンター成功でチーム A に +1。
	kmSetHand(g.GetPlayer(1), kmCard(1, 5), kmCard(2, 5), kmCard(3, 5), kmCard(4, 5))
	kmSetPhase(g, KempsPhaseDeclare)
	kmSetFourHolder(g, 1)
	require.NoError(t, g.PlayerDeclareCounterKemps(1))
	assert.Equal(t, 1, g.GetTeamScore(KempsTeamOf(0)))
	assert.Equal(t, KempsResultCounter, g.GetRoundResult())
}

func TestKemps_CounterKempsFail(t *testing.T) {
	g := newTestKemps()
	g.Reset()
	// 席 0 (自チーム/A) に +1 してから、的外れなカウンターでペナルティ -1 を確認。
	g.SetTeamScore(KempsTeamOf(0), 2)
	kmSetHand(g.GetPlayer(1), kmCard(1, 5), kmCard(2, 6), kmCard(3, 7), kmCard(4, 8))
	kmSetPhase(g, KempsPhaseDeclare)
	kmSetFourHolder(g, 0)
	require.NoError(t, g.PlayerDeclareCounterKemps(1)) // 席 1 は四枚なし → 失敗
	assert.Equal(t, 1, g.GetTeamScore(KempsTeamOf(0)))
	assert.Equal(t, KempsResultCounterFail, g.GetRoundResult())
}

func TestKemps_CounterKempsInvalidSeat(t *testing.T) {
	g := newTestKemps()
	g.Reset()
	kmSetPhase(g, KempsPhaseDeclare)
	kmSetFourHolder(g, 1)
	assert.ErrorIs(t, g.PlayerDeclareCounterKemps(99), ErrInvalidPlay)
}

func TestKemps_PassDuringDeclareResolves(t *testing.T) {
	g := newTestKemps()
	g.Reset()
	g.SetRand(rand.New(rand.NewSource(1)))
	kmSetHand(g.GetPlayer(1), kmCard(1, 5), kmCard(2, 5), kmCard(3, 5), kmCard(4, 5))
	kmSetPhase(g, KempsPhaseDeclare)
	kmSetFourHolder(g, 1)
	require.NoError(t, g.PlayerPass()) // 見送り → 自動解決
	assert.Equal(t, KempsPhaseRoundEnd, g.GetPhase())
}

func TestKemps_WinAtTargetScore(t *testing.T) {
	g := newTestKemps()
	g.Reset()
	g.SetTeamScore(KempsTeamOf(0), KempsTargetScore-1)
	kmSetHand(g.GetPlayer(0), kmCard(1, 7), kmCard(2, 7), kmCard(3, 7), kmCard(4, 7))
	kmSetPhase(g, KempsPhaseDeclare)
	kmSetFourHolder(g, 0)
	require.NoError(t, g.PlayerDeclareKemps())
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, KempsTeamOf(0), g.GetWinnerTeam())
	assert.Equal(t, KempsPhaseGameEnd, g.GetPhase())
	// 勝利チームのプレイヤーが finished。
	assert.True(t, g.GetPlayer(0).GetIsFinished())
	assert.True(t, g.GetPlayer(2).GetIsFinished())
	assert.False(t, g.GetPlayer(1).GetIsFinished())
}

func TestKemps_NextRoundAndGameEndGuards(t *testing.T) {
	g := newTestKemps()
	g.Reset()
	r1 := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, r1+1, g.GetRoundNumber())

	kmSetGameEnd(g)
	before := g.GetRoundNumber()
	g.NextRound() // ゲーム終了済みなら何もしない
	assert.Equal(t, before, g.GetRoundNumber())
}

// TestKemps_FullCpuGameTerminates は全員 CPU でゲームが必ず終了することを検証する。
func TestKemps_FullCpuGameTerminates(t *testing.T) {
	for _, diff := range []KempsCpuDifficulty{KempsCpuEasy, KempsCpuNormal, KempsCpuHard} {
		players := []*KempsPlayer{
			NewKempsPlayer(false), NewKempsPlayer(false),
			NewKempsPlayer(false), NewKempsPlayer(false),
		}
		g := NewKemps(NewTrumpCards(0), players, KempsConfig{CpuDifficulty: diff, TargetScore: 3})
		g.SetRand(rand.New(rand.NewSource(7)))
		g.Reset()
		for step := 0; step < 5_000_000 && !g.GetGameEndFlag(); step++ {
			switch g.GetPhase() {
			case KempsPhaseExchange, KempsPhaseDeclare:
				g.CpuPlay()
			case KempsPhaseRoundEnd:
				g.NextRound()
			default:
				g.CpuPlay()
			}
		}
		assert.True(t, g.GetGameEndFlag(), "diff=%d should terminate", diff)
		assert.GreaterOrEqual(t, g.GetWinnerTeam(), 0)
	}
}

func TestKemps_IsHumanTurn(t *testing.T) {
	g := newTestKemps()
	g.Reset()
	kmSetCurrentPlayer(g, 0)
	assert.True(t, g.IsHumanTurn())
	kmSetCurrentPlayer(g, 1)
	assert.False(t, g.IsHumanTurn())

	kmSetPhase(g, KempsPhaseDeclare)
	assert.True(t, g.IsHumanTurn())

	// 全員 CPU のとき宣言フェーズでも人間手番ではない。
	g2 := NewKemps(NewTrumpCards(0), []*KempsPlayer{
		NewKempsPlayer(false), NewKempsPlayer(false),
		NewKempsPlayer(false), NewKempsPlayer(false),
	}, DefaultKempsConfig())
	g2.Reset()
	kmSetPhase(g2, KempsPhaseDeclare)
	assert.False(t, g2.IsHumanTurn())

	kmSetGameEnd(g2)
	assert.False(t, g2.IsHumanTurn())
}

func TestKemps_ConfigValidate(t *testing.T) {
	assert.NoError(t, DefaultKempsConfig().Validate())
	assert.Error(t, KempsConfig{CpuDifficulty: 99, TargetScore: 5}.Validate())
	assert.Error(t, KempsConfig{CpuDifficulty: KempsCpuNormal, TargetScore: 0}.Validate())

	assert.InDelta(t, 0.10, KempsConfig{CpuDifficulty: KempsCpuEasy}.CounterChance(), 0.001)
	assert.InDelta(t, 0.45, KempsConfig{CpuDifficulty: KempsCpuHard}.CounterChance(), 0.001)
	assert.InDelta(t, 0.25, KempsConfig{CpuDifficulty: KempsCpuNormal}.CounterChance(), 0.001)
}

func TestKemps_JSONRoundTrip(t *testing.T) {
	g := newTestKemps()
	g.Reset()
	g.PlayerSetSignal(int(SignalBlink))
	g.SetTeamScore(0, 2)

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored Kemps
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetSignalType(), restored.GetSignalType())
	assert.Equal(t, 2, restored.GetTeamScore(0))
	assert.Equal(t, g.GetFieldSize(), restored.GetFieldSize())
}

func TestKemps_UnmarshalRejectsBadState(t *testing.T) {
	good := newTestKemps()
	good.Reset()
	base, _ := json.Marshal(good)

	mutate := func(fields map[string]any) []byte {
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(base, &raw)
		for k, v := range fields {
			raw[k], _ = json.Marshal(v)
		}
		out, _ := json.Marshal(raw)
		return out
	}

	cases := map[string]map[string]any{
		"bad phase":        {"ph": 99},
		"bad signal":       {"sg": 99},
		"bad current idx":  {"ci": 99},
		"bad winner":       {"wt": 99},
		"bad four holder":  {"fh": 99},
		"bad team score":   {"sc": []int{-1, 0}},
		"bad config":       {"cf": map[string]int{"cd": 99, "ts": 5}},
		"oversized field":  {"fd": make([]*Card, KempsFieldSize+1)},
		"bad round winner": {"rw": 99},
	}
	for name, fields := range cases {
		t.Run(name, func(t *testing.T) {
			var g Kemps
			assert.Error(t, json.Unmarshal(mutate(fields), &g))
		})
	}

	// nil プレイヤーは拒否。
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(base, &raw)
	raw["ps"], _ = json.Marshal([]*KempsPlayer{nil, nil, nil, nil})
	out, _ := json.Marshal(raw)
	var g Kemps
	assert.Error(t, json.Unmarshal(out, &g))

	// 非 JSON は失敗。
	assert.Error(t, json.Unmarshal([]byte("not-json"), &g))
}

func TestKemps_PlayerJSONRoundTrip(t *testing.T) {
	p := NewKempsPlayer(true)
	kmSetHand(p, kmCard(1, 7), kmCard(2, 7))
	data, err := json.Marshal(p)
	require.NoError(t, err)
	var restored KempsPlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, 2, restored.GetCardsSize())

	// nil GamePlayer のフォールバック。
	var empty KempsPlayer
	require.NoError(t, json.Unmarshal([]byte(`{}`), &empty))
	assert.NotNil(t, empty.GamePlayer)
}

// テスト用セッターが実際にフラグを立てること。presenter 側からしか呼ばれないと
// ドメインのカバレッジに乗らない。
func TestKemps_SetGameEndFlagForTest(t *testing.T) {
	g := NewDefaultKemps()
	g.Reset()
	if g.GetGameEndFlag() {
		t.Fatal("a fresh game must not be over")
	}
	g.SetGameEndFlagForTest(true)
	if !g.GetGameEndFlag() {
		t.Fatal("the setter did not take effect")
	}
}
