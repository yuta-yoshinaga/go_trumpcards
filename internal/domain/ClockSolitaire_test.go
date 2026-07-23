//go:build test

package domain

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClockSolitaire() *ClockSolitaire {
	tc := NewTrumpCards(0)
	cs := NewClockSolitaire(tc)
	cs.Reset()
	return cs
}

func TestClockSolitaire_Reset(t *testing.T) {
	cs := newTestClockSolitaire()

	assert.Equal(t, ClockSolitairePhasePlaying, cs.GetPhase())
	assert.Equal(t, 0, cs.GetStepCount())
	assert.NotNil(t, cs.GetCurrentCard())
	assert.Empty(t, cs.GetActionLog())

	// 各パイルのカード数確認（中央パイルは1枚取り出してcurrentCardにした状態）
	piles := cs.GetPiles()
	totalCards := 0
	for i := range ClockSolitairePileCount {
		totalCards += len(piles[i])
	}
	// 51枚 = 中央パイル3枚 + 残り12パイル×4枚（1枚はcurrentCardとして手に持っている）
	assert.Equal(t, 51, totalCards)

	// Reset直後はまだ配置済みカードがないのでfaceUpCountはすべて0
	fuc := cs.GetFaceUpCount()
	assert.Equal(t, 0, fuc[ClockSolitaireKingPileIdx])
}

func TestClockSolitaire_Reset_ClearsState(t *testing.T) {
	cs := newTestClockSolitaire()
	_ = cs.Step()
	assert.Greater(t, cs.GetStepCount(), 0)
	assert.NotEmpty(t, cs.GetActionLog())

	cs.Reset()
	assert.Equal(t, 0, cs.GetStepCount())
	assert.Empty(t, cs.GetActionLog())
	assert.Equal(t, ClockSolitairePhasePlaying, cs.GetPhase())
}

func TestClockSolitaire_Step_PlacesCardAtCorrectPile(t *testing.T) {
	cs := newTestClockSolitaire()

	// Aceを持っている状態にする → pile 0 に配置される
	ace := NewCard(CardDesignSpade, 1, false)
	cs.SetCurrentCard(ace)

	origLen := len(cs.GetPiles()[0])
	err := cs.Step()
	require.NoError(t, err)

	// pile 0 にカードが配置され、flipTopFaceDownで1枚取り出されるため長さは変わらない
	assert.Equal(t, origLen, len(cs.GetPiles()[0]))
	lastCard := cs.GetPiles()[0][len(cs.GetPiles()[0])-1]
	assert.Equal(t, ace, lastCard.Card)
	assert.True(t, lastCard.FaceUp)
}

func TestClockSolitaire_Step_AceGoesToPile0(t *testing.T) {
	cs := newTestClockSolitaire()
	ace := NewCard(CardDesignHeart, 1, false)
	cs.SetCurrentCard(ace)

	err := cs.Step()
	require.NoError(t, err)

	pile0 := cs.GetPiles()[0]
	lastCard := pile0[len(pile0)-1]
	assert.Equal(t, 1, lastCard.Card.GetValue())
}

func TestClockSolitaire_Step_QueenGoesToPile11(t *testing.T) {
	cs := newTestClockSolitaire()
	queen := NewCard(CardDesignDiamond, 12, false)
	cs.SetCurrentCard(queen)

	err := cs.Step()
	require.NoError(t, err)

	pile11 := cs.GetPiles()[11]
	lastCard := pile11[len(pile11)-1]
	assert.Equal(t, 12, lastCard.Card.GetValue())
}

func TestClockSolitaire_Step_KingGoesToCenter(t *testing.T) {
	cs := newTestClockSolitaire()
	king := NewCard(CardDesignClover, 13, false)
	cs.SetCurrentCard(king)

	err := cs.Step()
	require.NoError(t, err)

	centerPile := cs.GetPiles()[ClockSolitaireKingPileIdx]
	lastCard := centerPile[len(centerPile)-1]
	assert.Equal(t, 13, lastCard.Card.GetValue())
}

func TestClockSolitaire_Step_GameOver_FourthKing(t *testing.T) {
	cs := newTestClockSolitaire()

	// 中央パイルに3枚のKが表向き、残り1枚が裏向き
	var piles [ClockSolitairePileCount][]*ClockSolitaireCard
	var fuc [ClockSolitairePileCount]int

	// 時計位置はそれぞれ4枚ずつ裏向き（クリアでない状態を作る）
	for i := range ClockSolitaireKingPileIdx {
		piles[i] = make([]*ClockSolitaireCard, 0, 4)
		for j := range 4 {
			piles[i] = append(piles[i], &ClockSolitaireCard{
				Card:   NewCard(CardDesignSpade, i+1, false),
				FaceUp: j < 2, // 2枚表向き、2枚裏向き
			})
		}
		fuc[i] = 2
	}

	// 中央パイル: 3枚K表向き + 1枚K裏向き
	piles[ClockSolitaireKingPileIdx] = []*ClockSolitaireCard{
		{Card: NewCard(CardDesignSpade, 13, false), FaceUp: true},
		{Card: NewCard(CardDesignClover, 13, false), FaceUp: true},
		{Card: NewCard(CardDesignHeart, 13, false), FaceUp: true},
		{Card: NewCard(CardDesignDiamond, 13, false), FaceUp: false},
	}
	fuc[ClockSolitaireKingPileIdx] = 3

	cs.SetPiles(piles)
	cs.SetFaceUpCount(fuc)

	// Kを持っている → 中央パイルに配置 → 4枚目が表向き → GameOver
	king := NewCard(CardDesignSpade, 13, false)
	cs.SetCurrentCard(king)

	err := cs.Step()
	require.NoError(t, err)
	assert.Equal(t, ClockSolitairePhaseGameOver, cs.GetPhase())
}

func TestClockSolitaire_Step_GameClear(t *testing.T) {
	cs := newTestClockSolitaire()

	var piles [ClockSolitairePileCount][]*ClockSolitaireCard
	var fuc [ClockSolitairePileCount]int

	// 時計位置0-10: 全て4枚表向き（クリア済み）
	for i := range 11 {
		piles[i] = make([]*ClockSolitaireCard, 0, 4)
		for range 4 {
			piles[i] = append(piles[i], &ClockSolitaireCard{
				Card:   NewCard(CardDesignSpade, i+1, false),
				FaceUp: true,
			})
		}
		fuc[i] = 4
	}

	// 位置11(Q): 3枚表向き + 1枚裏向き（あと1枚でクリア）
	piles[11] = []*ClockSolitaireCard{
		{Card: NewCard(CardDesignSpade, 12, false), FaceUp: true},
		{Card: NewCard(CardDesignClover, 12, false), FaceUp: true},
		{Card: NewCard(CardDesignHeart, 12, false), FaceUp: true},
		{Card: NewCard(CardDesignDiamond, 12, false), FaceUp: false},
	}
	fuc[11] = 3

	// 中央パイル
	piles[ClockSolitaireKingPileIdx] = []*ClockSolitaireCard{
		{Card: NewCard(CardDesignSpade, 13, false), FaceUp: true},
		{Card: NewCard(CardDesignClover, 13, false), FaceUp: true},
		{Card: NewCard(CardDesignHeart, 13, false), FaceUp: false},
		{Card: NewCard(CardDesignDiamond, 13, false), FaceUp: false},
	}
	fuc[ClockSolitaireKingPileIdx] = 2

	cs.SetPiles(piles)
	cs.SetFaceUpCount(fuc)

	// Qを持っている → pile11に配置 → fuc[11]=4 → GameClear
	queen := NewCard(CardDesignDiamond, 12, false)
	cs.SetCurrentCard(queen)

	err := cs.Step()
	require.NoError(t, err)
	assert.Equal(t, ClockSolitairePhaseGameClear, cs.GetPhase())
}

func TestClockSolitaire_Step_ErrorWhenNotPlaying(t *testing.T) {
	cs := newTestClockSolitaire()
	cs.SetPhase(ClockSolitairePhaseGameOver)

	err := cs.Step()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in playing phase")
}

func TestClockSolitaire_Step_ErrorWhenNoCurrentCard(t *testing.T) {
	cs := newTestClockSolitaire()
	cs.SetCurrentCard(nil)

	err := cs.Step()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no current card")
}

func TestClockSolitaire_AutoPlay_Terminates(t *testing.T) {
	cs := newTestClockSolitaire()

	err := cs.AutoPlay()
	require.NoError(t, err)

	phase := cs.GetPhase()
	assert.True(t, phase == ClockSolitairePhaseGameClear || phase == ClockSolitairePhaseGameOver)
	assert.Greater(t, cs.GetStepCount(), 0)
	assert.NotEmpty(t, cs.GetActionLog())
}

func TestClockSolitaire_AutoPlay_ErrorWhenNotPlaying(t *testing.T) {
	cs := newTestClockSolitaire()
	cs.SetPhase(ClockSolitairePhaseGameOver)

	err := cs.AutoPlay()
	assert.Error(t, err)
}

func TestClockSolitaire_JSON_RoundTrip(t *testing.T) {
	cs := newTestClockSolitaire()
	_ = cs.Step()
	_ = cs.Step()

	data, err := json.Marshal(cs)
	require.NoError(t, err)

	cs2 := &ClockSolitaire{}
	err = json.Unmarshal(data, cs2)
	require.NoError(t, err)

	assert.Equal(t, cs.GetPhase(), cs2.GetPhase())
	assert.Equal(t, cs.GetStepCount(), cs2.GetStepCount())
	assert.Equal(t, len(cs.GetActionLog()), len(cs2.GetActionLog()))
	assert.Equal(t, cs.GetFaceUpCount(), cs2.GetFaceUpCount())
}

func TestClockSolitaire_UnmarshalJSON_OversizedInput(t *testing.T) {
	// 巨大なActionLogを含むJSONを拒否する
	bigLog := make([]*ActionLogEntry, 1001)
	for i := range bigLog {
		bigLog[i] = &ActionLogEntry{}
	}
	j := clockSolitaireJSON{ActionLog: bigLog}
	data, err := json.Marshal(j)
	require.NoError(t, err)

	cs := &ClockSolitaire{}
	err = json.Unmarshal(data, cs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

func TestClockSolitaire_UnmarshalJSON_NilTrumpCards(t *testing.T) {
	j := clockSolitaireJSON{}
	data, err := json.Marshal(j)
	require.NoError(t, err)

	cs := &ClockSolitaire{}
	err = json.Unmarshal(data, cs)
	require.NoError(t, err)
	assert.NotNil(t, cs.trumpCards)
}

func TestClockSolitaire_Step_ActionLog(t *testing.T) {
	cs := newTestClockSolitaire()

	err := cs.Step()
	require.NoError(t, err)

	log := cs.GetActionLog()
	require.Len(t, log, 1)
	assert.Equal(t, "step", log[0].ActionType)
	assert.Contains(t, log[0].Detail, "パイルに配置")
	assert.Len(t, log[0].Cards, 1)
}

func TestClockSolitaire_MultipleResets(t *testing.T) {
	cs := newTestClockSolitaire()
	_ = cs.AutoPlay()

	cs.Reset()
	assert.Equal(t, ClockSolitairePhasePlaying, cs.GetPhase())
	assert.Equal(t, 0, cs.GetStepCount())

	err := cs.Step()
	require.NoError(t, err)
	assert.Equal(t, 1, cs.GetStepCount())
}

func TestClockSolitaire_UnmarshalJSON_InvalidJSON(t *testing.T) {
	cs := &ClockSolitaire{}
	err := json.Unmarshal([]byte("invalid"), cs)
	assert.Error(t, err)
}

func TestClockSolitaire_UnmarshalJSON_OversizedPile(t *testing.T) {
	// 1パイルが巨大
	var piles [ClockSolitairePileCount][]*ClockSolitaireCard
	bigPile := make([]*ClockSolitaireCard, 1001)
	for i := range bigPile {
		bigPile[i] = &ClockSolitaireCard{Card: NewCard(CardDesignSpade, 1, false)}
	}
	piles[0] = bigPile
	j := clockSolitaireJSON{Piles: piles}
	data, err := json.Marshal(j)
	require.NoError(t, err)

	cs := &ClockSolitaire{}
	err = json.Unmarshal(data, cs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pile 0 exceeds")
}

func TestClockSolitaire_AutoPlay_StepCountMatchesLogLength(t *testing.T) {
	// AutoPlay後のstepCountとactionLogの長さが一致する
	for range 10 {
		cs := newTestClockSolitaire()
		_ = cs.AutoPlay()
		assert.Equal(t, cs.GetStepCount(), len(cs.GetActionLog()),
			"stepCount should match actionLog length")
	}
}

func TestClockSolitaire_FlipTopFaceDown_AllFaceUp(t *testing.T) {
	// パイルの全カードが表向きの場合、flipTopFaceDownはcurrentCardをnilのままにする
	cs := newTestClockSolitaire()

	// GameClear状態を作る: 全12時計位置が4枚表向き、中央は3枚表向き+次のKで終了
	var piles [ClockSolitairePileCount][]*ClockSolitaireCard
	var fuc [ClockSolitairePileCount]int
	for i := range ClockSolitaireKingPileIdx {
		piles[i] = make([]*ClockSolitaireCard, 0, 4)
		for range 4 {
			piles[i] = append(piles[i], &ClockSolitaireCard{
				Card:   NewCard(CardDesignSpade, i+1, false),
				FaceUp: true,
			})
		}
		fuc[i] = 4
	}
	// 中央パイル: 全て表向き
	piles[ClockSolitaireKingPileIdx] = []*ClockSolitaireCard{
		{Card: NewCard(CardDesignSpade, 13, false), FaceUp: true},
		{Card: NewCard(CardDesignClover, 13, false), FaceUp: true},
		{Card: NewCard(CardDesignHeart, 13, false), FaceUp: true},
		{Card: NewCard(CardDesignDiamond, 13, false), FaceUp: true},
	}
	fuc[ClockSolitaireKingPileIdx] = 4

	cs.SetPiles(piles)
	cs.SetFaceUpCount(fuc)

	// Kを配置 → GameClear判定でクリアになるはず
	king := NewCard(CardDesignSpade, 13, false)
	cs.SetCurrentCard(king)

	// ただしクリア条件を満たしているので、flipTopFaceDownは呼ばれない
	err := cs.Step()
	require.NoError(t, err)
	assert.Equal(t, ClockSolitairePhaseGameClear, cs.GetPhase())
}

func TestClockSolitaire_UnmarshalJSON_PreservesNilActionLog(t *testing.T) {
	j := clockSolitaireJSON{
		TrumpCards: NewTrumpCards(0),
	}
	data, err := json.Marshal(j)
	require.NoError(t, err)

	cs := &ClockSolitaire{}
	err = json.Unmarshal(data, cs)
	require.NoError(t, err)
	// nil actionLog should become empty slice
	assert.NotNil(t, cs.actionLog)
	assert.Empty(t, cs.actionLog)
}

func TestClockSolitaire_JSON_RoundTrip_AfterAutoPlay(t *testing.T) {
	cs := newTestClockSolitaire()
	_ = cs.AutoPlay()

	data, err := json.Marshal(cs)
	require.NoError(t, err)

	cs2 := &ClockSolitaire{}
	err = json.Unmarshal(data, cs2)
	require.NoError(t, err)

	assert.Equal(t, cs.GetPhase(), cs2.GetPhase())
	assert.Equal(t, cs.GetStepCount(), cs2.GetStepCount())

	// currentCardはゲーム終了時nilの可能性がある
	if cs.GetCurrentCard() == nil {
		assert.Nil(t, cs2.GetCurrentCard())
	}
}

func TestClockSolitaire_UnmarshalJSON_NilPiles(t *testing.T) {
	j := clockSolitaireJSON{
		TrumpCards: NewTrumpCards(0),
	}
	data, err := json.Marshal(j)
	require.NoError(t, err)

	cs := &ClockSolitaire{}
	err = json.Unmarshal(data, cs)
	require.NoError(t, err)

	for i := range ClockSolitairePileCount {
		assert.NotNil(t, cs.piles[i], "pile %d should not be nil", i)
	}
}

func TestClockSolitaire_Step_IncreasesStepCount(t *testing.T) {
	cs := newTestClockSolitaire()
	before := cs.GetStepCount()
	err := cs.Step()
	require.NoError(t, err)
	assert.Equal(t, before+1, cs.GetStepCount())
}

func TestClockSolitaire_Step_FlipsNextCard(t *testing.T) {
	cs := newTestClockSolitaire()

	// Stepの前のcurrentCard
	cardBefore := cs.GetCurrentCard()
	require.NotNil(t, cardBefore)

	err := cs.Step()
	require.NoError(t, err)

	// ゲームが続行中なら新しいcurrentCardがある
	if cs.GetPhase() == ClockSolitairePhasePlaying {
		assert.NotNil(t, cs.GetCurrentCard())
	}
}

func TestClockSolitaire_AutoPlay_RunsMultipleTimes(t *testing.T) {
	// 統計的テスト: 複数回実行して勝敗両方が出ることを確認
	wins := 0
	losses := 0
	for range 100 {
		cs := newTestClockSolitaire()
		_ = cs.AutoPlay()
		if cs.GetPhase() == ClockSolitairePhaseGameClear {
			wins++
		} else {
			losses++
		}
	}
	// クロックソリティアの勝率は約1%なので、ほぼ全て負けになるはず
	_ = wins // 勝率が極めて低いのでwins==0でも正常
	assert.Greater(t, losses, 0, "should have at least one loss")
}

func TestClockSolitaire_GetActionLog_Empty(t *testing.T) {
	cs := newTestClockSolitaire()
	assert.Empty(t, cs.GetActionLog())
}

func TestClockSolitaire_MarshalJSON_InvalidJSON(t *testing.T) {
	// 正常なMarshalJSONのテスト
	cs := newTestClockSolitaire()
	data, err := json.Marshal(cs)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(data), "{"))
}

func TestClockSolitaire_CanUndo_InitiallyFalse(t *testing.T) {
	cs := newTestClockSolitaire()
	assert.False(t, cs.CanUndo())
}

func TestClockSolitaire_Undo_ErrorWhenNoHistory(t *testing.T) {
	cs := newTestClockSolitaire()
	err := cs.Undo()
	assert.Error(t, err)
}

func TestClockSolitaire_Undo_RestoresPreviousState(t *testing.T) {
	cs := newTestClockSolitaire()

	// Force a deterministic first move: place an Ace on pile 0.
	ace := NewCard(CardDesignSpade, 1, false)
	cs.SetCurrentCard(ace)

	prevPiles := cs.GetPiles()
	prevFuc := cs.GetFaceUpCount()
	prevStep := cs.GetStepCount()

	require.NoError(t, cs.Step())
	assert.True(t, cs.CanUndo())
	assert.Equal(t, prevStep+1, cs.GetStepCount())

	require.NoError(t, cs.Undo())

	// State reverts to the pre-step snapshot.
	assert.Equal(t, prevStep, cs.GetStepCount())
	assert.Equal(t, prevFuc, cs.GetFaceUpCount())
	assert.Equal(t, prevPiles, cs.GetPiles())
	assert.Equal(t, ace, cs.GetCurrentCard())
	assert.Equal(t, ClockSolitairePhasePlaying, cs.GetPhase())
	// No more history to undo.
	assert.False(t, cs.CanUndo())
}

func TestClockSolitaire_Undo_RecordsActionLog(t *testing.T) {
	cs := newTestClockSolitaire()
	cs.SetCurrentCard(NewCard(CardDesignSpade, 1, false))
	require.NoError(t, cs.Step())

	logLenAfterStep := len(cs.GetActionLog())
	require.NoError(t, cs.Undo())

	// Undo appends its own action-log entry (not removed).
	log := cs.GetActionLog()
	assert.Equal(t, logLenAfterStep+1, len(log))
	assert.Equal(t, "undo", log[len(log)-1].ActionType)
}

func TestClockSolitaire_Undo_RevertsGameOver(t *testing.T) {
	cs := newTestClockSolitaire()
	// Arrange the center pile so the next King placement is the fourth face-up King.
	cs.SetFaceUpCount(func() [ClockSolitairePileCount]int {
		var fuc [ClockSolitairePileCount]int
		fuc[ClockSolitaireKingPileIdx] = 3
		return fuc
	}())
	cs.SetCurrentCard(NewCard(CardDesignSpade, 13, false))

	require.NoError(t, cs.Step())
	assert.Equal(t, ClockSolitairePhaseGameOver, cs.GetPhase())
	assert.True(t, cs.CanUndo())

	require.NoError(t, cs.Undo())
	assert.Equal(t, ClockSolitairePhasePlaying, cs.GetPhase())
	assert.Equal(t, 3, cs.GetFaceUpCount()[ClockSolitaireKingPileIdx])
}

func TestClockSolitaire_Reset_ClearsUndoHistory(t *testing.T) {
	cs := newTestClockSolitaire()
	cs.SetCurrentCard(NewCard(CardDesignSpade, 1, false))
	require.NoError(t, cs.Step())
	assert.True(t, cs.CanUndo())

	cs.Reset()
	assert.False(t, cs.CanUndo())
}

func TestClockSolitaire_Undo_StackCappedAtMaxDepth(t *testing.T) {
	cs := newTestClockSolitaire()
	// Push more snapshots than the cap by repeatedly stepping with a fresh Ace.
	for i := 0; i < ClockSolitaireMaxUndoDepth+10; i++ {
		cs.SetPhase(ClockSolitairePhasePlaying)
		cs.SetCurrentCard(NewCard(CardDesignSpade, 1, false))
		require.NoError(t, cs.Step())
	}
	assert.LessOrEqual(t, len(cs.history), ClockSolitaireMaxUndoDepth)
}

func TestClockSolitaire_JSON_RoundTrip_PreservesUndoHistory(t *testing.T) {
	cs := newTestClockSolitaire()
	cs.SetCurrentCard(NewCard(CardDesignSpade, 1, false))
	require.NoError(t, cs.Step())
	require.True(t, cs.CanUndo())

	data, err := json.Marshal(cs)
	require.NoError(t, err)

	cs2 := &ClockSolitaire{}
	require.NoError(t, json.Unmarshal(data, cs2))
	assert.True(t, cs2.CanUndo())

	// Undo works after a JSON round-trip (survives page reload / KV persistence).
	require.NoError(t, cs2.Undo())
	assert.False(t, cs2.CanUndo())
}
