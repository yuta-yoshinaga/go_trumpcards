//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Config ---

func TestSpiteAndMaliceConfig_Default_Validate(t *testing.T) {
	cfg := DefaultSpiteAndMaliceConfig()
	assert.NoError(t, cfg.Validate())
	assert.Equal(t, SpiteAndMaliceGoalSizeDefault, cfg.GoalSize)
	assert.Equal(t, SpiteAndMaliceCpuDifficultyNormal, cfg.CpuDifficulty)
}

func TestSpiteAndMaliceConfig_Validate_Errors(t *testing.T) {
	tests := []SpiteAndMaliceConfig{
		{GoalSize: SpiteAndMaliceGoalSizeMin - 1, CpuDifficulty: SpiteAndMaliceCpuDifficultyNormal},
		{GoalSize: SpiteAndMaliceGoalSizeMax + 1, CpuDifficulty: SpiteAndMaliceCpuDifficultyNormal},
		{GoalSize: 15, CpuDifficulty: -1},
		{GoalSize: 15, CpuDifficulty: SpiteAndMaliceCpuDifficultyHard + 1},
	}
	for _, c := range tests {
		assert.Error(t, c.Validate())
	}
}

func TestSpiteAndMaliceConfig_JSONRoundTrip(t *testing.T) {
	cfg := SpiteAndMaliceConfig{GoalSize: 18, CpuDifficulty: SpiteAndMaliceCpuDifficultyHard}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	var decoded SpiteAndMaliceConfig
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, cfg, decoded)
}

// --- helpers ---

// Spite & Malice の試験用インスタンスを返す (Reset 済み)
func newTestSpiteAndMalice() *SpiteAndMalice {
	g := NewDefaultSpiteAndMalice()
	g.Reset()
	return g
}

// mkCard はテスト用に Card を生成するヘルパ。
func mkCard(design, value int) *Card {
	return NewCard(design, value, false)
}

func TestSpiteAndMalice_Reset_InitialState(t *testing.T) {
	g := newTestSpiteAndMalice()

	assert.Equal(t, SpiteAndMalicePhasePlaying, g.GetPhase())
	assert.Equal(t, SpiteAndMaliceHumanIdx, g.GetCurrent())
	assert.Equal(t, -1, g.GetWinner())
	assert.Empty(t, g.GetActionLog())
	for f := range SpiteAndMaliceFoundationCnt {
		assert.Empty(t, g.GetFoundations()[f])
	}

	// 各プレイヤー: goal = GoalSize, hand = HandMax
	for i := range SpiteAndMalicePlayerCnt {
		p := g.GetPlayer(i)
		require.NotNil(t, p)
		assert.Equal(t, SpiteAndMaliceGoalSizeDefault, p.GoalSize(), "player %d goal size", i)
		assert.Equal(t, SpiteAndMaliceHandMax, p.HandSize(), "player %d hand size", i)
		for s := range SpiteAndMaliceSideCnt {
			assert.Zero(t, p.SideSize(s))
		}
	}

	// ストックは 2 デッキ 104 - (goal * 2) - (hand * 2)
	expected := 2*CardCnt - 2*SpiteAndMaliceGoalSizeDefault - 2*SpiteAndMaliceHandMax
	assert.Equal(t, expected, g.GetStockSize())
}

func TestSpiteAndMalice_Reset_RecoversInvalidConfig(t *testing.T) {
	g := NewSpiteAndMalice(NewTrumpCardsWithDecks(2, 0), SpiteAndMaliceConfig{GoalSize: 999})
	g.Reset()
	assert.Equal(t, SpiteAndMaliceGoalSizeDefault, g.GetConfig().GoalSize)
}

// --- PlayFromHand ---

func TestSpiteAndMalice_PlayFromHand_StartsFoundationWithAce(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetPlayerHand(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignSpade, 1)})
	require.NoError(t, g.PlayFromHand(0, 0))

	f := g.GetFoundations()
	require.Len(t, f[0], 1)
	assert.Equal(t, 1, f[0][0].GetValue())
	assert.Equal(t, 1, g.GetMoveCount())
}

func TestSpiteAndMalice_PlayFromHand_Sequential(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetFoundation(0, []*Card{mkCard(CardDesignSpade, 1), mkCard(CardDesignHeart, 2)})
	g.SetPlayerHand(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignDiamond, 3)})
	require.NoError(t, g.PlayFromHand(0, 0))
	assert.Equal(t, 3, g.GetFoundationTopValue(0))
}

func TestSpiteAndMalice_PlayFromHand_WildOnFoundation(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetFoundation(0, []*Card{mkCard(CardDesignSpade, 1), mkCard(CardDesignHeart, 2)})
	g.SetPlayerHand(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignClover, 13)})
	require.NoError(t, g.PlayFromHand(0, 0))
	// K は 3 として振る舞う
	assert.Equal(t, 3, g.GetFoundationTopValue(0))
	// 次に 4 を置けるはず
	g.SetPlayerHand(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignHeart, 4)})
	require.NoError(t, g.PlayFromHand(0, 0))
}

func TestSpiteAndMalice_PlayFromHand_InvalidValue(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetFoundation(0, []*Card{mkCard(CardDesignSpade, 1)})
	g.SetPlayerHand(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignDiamond, 5)})
	err := g.PlayFromHand(0, 0)
	assert.Error(t, err)
}

func TestSpiteAndMalice_PlayFromHand_InvalidIndices(t *testing.T) {
	g := newTestSpiteAndMalice()
	assert.Error(t, g.PlayFromHand(-1, 0))
	assert.Error(t, g.PlayFromHand(0, -1))
	assert.Error(t, g.PlayFromHand(0, SpiteAndMaliceFoundationCnt))
}

func TestSpiteAndMalice_PlayFromHand_GameOver(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetPhase(SpiteAndMalicePhaseGameOver)
	assert.Error(t, g.PlayFromHand(0, 0))
}

// --- PlayFromGoal ---

func TestSpiteAndMalice_PlayFromGoal_Valid(t *testing.T) {
	g := newTestSpiteAndMalice()
	// top of goal = ace (末尾がトップ)
	g.SetPlayerGoal(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignDiamond, 5), mkCard(CardDesignSpade, 1)})
	require.NoError(t, g.PlayFromGoal(0))
	assert.Equal(t, 1, g.GetFoundationTopValue(0))
	// 次のゴールカードが見える
	assert.Equal(t, 5, g.GetPlayer(SpiteAndMaliceHumanIdx).GoalTop().GetValue())
}

func TestSpiteAndMalice_PlayFromGoal_WinsGame(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetPlayerGoal(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignSpade, 1)})
	require.NoError(t, g.PlayFromGoal(0))
	assert.Equal(t, SpiteAndMalicePhaseGameOver, g.GetPhase())
	assert.Equal(t, SpiteAndMaliceHumanIdx, g.GetWinner())
}

func TestSpiteAndMalice_PlayFromGoal_Empty(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetPlayerGoal(SpiteAndMaliceHumanIdx, nil)
	assert.Error(t, g.PlayFromGoal(0))
}

func TestSpiteAndMalice_PlayFromGoal_Invalid(t *testing.T) {
	g := newTestSpiteAndMalice()
	assert.Error(t, g.PlayFromGoal(-1))
	assert.Error(t, g.PlayFromGoal(SpiteAndMaliceFoundationCnt))
	g.SetPhase(SpiteAndMalicePhaseGameOver)
	assert.Error(t, g.PlayFromGoal(0))
}

// --- PlayFromSide ---

func TestSpiteAndMalice_PlayFromSide_Valid(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetPlayerSide(SpiteAndMaliceHumanIdx, 0, []*Card{mkCard(CardDesignSpade, 1)})
	require.NoError(t, g.PlayFromSide(0, 0))
	assert.Equal(t, 1, g.GetFoundationTopValue(0))
	assert.Equal(t, 0, g.GetPlayer(SpiteAndMaliceHumanIdx).SideSize(0))
}

func TestSpiteAndMalice_PlayFromSide_Empty(t *testing.T) {
	g := newTestSpiteAndMalice()
	assert.Error(t, g.PlayFromSide(0, 0))
}

func TestSpiteAndMalice_PlayFromSide_Invalid(t *testing.T) {
	g := newTestSpiteAndMalice()
	assert.Error(t, g.PlayFromSide(-1, 0))
	assert.Error(t, g.PlayFromSide(0, -1))
	assert.Error(t, g.PlayFromSide(SpiteAndMaliceSideCnt, 0))
	g.SetPhase(SpiteAndMalicePhaseGameOver)
	assert.Error(t, g.PlayFromSide(0, 0))
}

// --- Foundation completion ---

func TestSpiteAndMalice_Foundation_CompletesOnQueen(t *testing.T) {
	g := newTestSpiteAndMalice()
	cards := []*Card{
		mkCard(CardDesignSpade, 1), mkCard(CardDesignSpade, 2), mkCard(CardDesignSpade, 3),
		mkCard(CardDesignSpade, 4), mkCard(CardDesignSpade, 5), mkCard(CardDesignSpade, 6),
		mkCard(CardDesignSpade, 7), mkCard(CardDesignSpade, 8), mkCard(CardDesignSpade, 9),
		mkCard(CardDesignSpade, 10), mkCard(CardDesignSpade, 11),
	}
	g.SetFoundation(0, cards)
	g.SetPlayerHand(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignHeart, 12)})
	require.NoError(t, g.PlayFromHand(0, 0))
	// 完成済みに回収される
	assert.Equal(t, 0, len(g.GetFoundations()[0]))
	assert.Equal(t, SpiteAndMaliceFoundationMax, g.GetCompletedSize())
}

// --- Discard / EndTurn ---

func TestSpiteAndMalice_Discard_EndsTurnAndRefills(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetPlayerHand(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignSpade, 5)})
	require.NoError(t, g.Discard(0, 0))
	// 次プレイヤーへ
	assert.Equal(t, SpiteAndMaliceCpuIdx, g.GetCurrent())
	// サイド 0 に捨て札
	assert.Equal(t, 1, g.GetPlayer(SpiteAndMaliceHumanIdx).SideSize(0))
	// 次プレイヤーは既に 5 枚手札を持っていたのでそのまま
	assert.Equal(t, SpiteAndMaliceHandMax, g.GetPlayer(SpiteAndMaliceCpuIdx).HandSize())
}

func TestSpiteAndMalice_Discard_Invalid(t *testing.T) {
	g := newTestSpiteAndMalice()
	assert.Error(t, g.Discard(-1, 0))
	assert.Error(t, g.Discard(0, -1))
	assert.Error(t, g.Discard(0, SpiteAndMaliceSideCnt))
	g.SetPhase(SpiteAndMalicePhaseGameOver)
	assert.Error(t, g.Discard(0, 0))
}

// --- CPU ---

func TestSpiteAndMalice_CpuStep_PlaysGoalFirst(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetCurrent(SpiteAndMaliceCpuIdx)
	g.SetPlayerGoal(SpiteAndMaliceCpuIdx, []*Card{mkCard(CardDesignSpade, 1)})
	// ファウンデーションは空
	require.NoError(t, g.CpuStep())
	// ゴールから A を出したので winner
	assert.Equal(t, SpiteAndMaliceCpuIdx, g.GetWinner())
}

func TestSpiteAndMalice_CpuStep_DiscardsWhenNoPlay(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetCurrent(SpiteAndMaliceCpuIdx)
	// 手札は全て同じカード (4 枚以上必要なし)、ファウンデーション・ゴールも非プレイ可能に設定
	g.SetPlayerHand(SpiteAndMaliceCpuIdx, []*Card{mkCard(CardDesignSpade, 5)})
	g.SetPlayerGoal(SpiteAndMaliceCpuIdx, []*Card{mkCard(CardDesignHeart, 9)})
	// ファウンデーション全て空 (= 1 しか置けない)
	require.NoError(t, g.CpuStep())
	// ターンが人間に戻っている
	assert.Equal(t, SpiteAndMaliceHumanIdx, g.GetCurrent())
}

func TestSpiteAndMalice_CpuStep_WrongTurn(t *testing.T) {
	g := newTestSpiteAndMalice()
	assert.Error(t, g.CpuStep()) // current is human
}

func TestSpiteAndMalice_CpuStep_Hard_AvoidsDiscardingOpponentTarget(t *testing.T) {
	g := NewSpiteAndMalice(NewTrumpCardsWithDecks(2, 0), SpiteAndMaliceConfig{
		GoalSize:      SpiteAndMaliceGoalSizeDefault,
		CpuDifficulty: SpiteAndMaliceCpuDifficultyHard,
	})
	g.Reset()
	g.SetCurrent(SpiteAndMaliceCpuIdx)
	// 相手のゴール top が 5 (= 次に必要なのは 6? 違う、次に出すのは自分のゴールの 5 なので、
	// 相手が使える = 5 を捨てると相手が即ゴールに使えてしまう可能性あり -> CPU は 5 を温存)
	g.SetPlayerGoal(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignSpade, 5)})
	g.SetPlayerGoal(SpiteAndMaliceCpuIdx, []*Card{mkCard(CardDesignHeart, 9)})
	g.SetPlayerHand(SpiteAndMaliceCpuIdx, []*Card{mkCard(CardDesignSpade, 5), mkCard(CardDesignHeart, 8)})
	// ファウンデーションは空 (= 1 しか置けない)
	require.NoError(t, g.CpuStep())
	// 8 を捨てて 5 は温存
	assert.Equal(t, 1, g.GetPlayer(SpiteAndMaliceCpuIdx).HandSize(), "should have kept one card")
}

// --- IsCpuTurn ---

func TestSpiteAndMalice_IsCpuTurn(t *testing.T) {
	g := newTestSpiteAndMalice()
	assert.False(t, g.IsCpuTurn())
	g.SetCurrent(SpiteAndMaliceCpuIdx)
	assert.True(t, g.IsCpuTurn())
	g.SetPhase(SpiteAndMalicePhaseGameOver)
	assert.False(t, g.IsCpuTurn())
}

// --- Hint ---

func TestSpiteAndMalice_GetHint_PrefersGoal(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetPlayerGoal(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignSpade, 1)})
	g.SetPlayerHand(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignDiamond, 1)})
	h := g.GetHint()
	require.NotNil(t, h)
	assert.False(t, h.Discard)
	assert.Equal(t, SpiteAndMaliceSourceGoal, h.Source)
}

func TestSpiteAndMalice_GetHint_FallsBackToDiscard(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetPlayerGoal(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignHeart, 9)})
	g.SetPlayerHand(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignClover, 7)})
	h := g.GetHint()
	require.NotNil(t, h)
	assert.True(t, h.Discard)
}

func TestSpiteAndMalice_GetHint_NilOnGameOver(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetPhase(SpiteAndMalicePhaseGameOver)
	assert.Nil(t, g.GetHint())
}

// --- Mid-turn refill ---

func TestSpiteAndMalice_AutoRefillsHandOnEmpty(t *testing.T) {
	g := newTestSpiteAndMalice()
	// foundation 空 -> A プレイ
	g.SetPlayerHand(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignSpade, 1)})
	require.NoError(t, g.PlayFromHand(0, 0))
	// 手札 0 -> 自動で 5 枚補充
	assert.Equal(t, SpiteAndMaliceHandMax, g.GetPlayer(SpiteAndMaliceHumanIdx).HandSize())
}

// --- Stock refill from completed ---

func TestSpiteAndMalice_RefillStockFromCompleted(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetStock(nil)
	g.SetCompleted([]*Card{mkCard(CardDesignSpade, 1), mkCard(CardDesignHeart, 2)})
	g.SetPlayerHand(SpiteAndMaliceHumanIdx, nil) // 空にしておく
	// ディスカードでターン終了 -> 相手ターンへ -> drawToHand が呼ばれる
	g.SetPlayerHand(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignClover, 7)})
	g.SetPlayerHand(SpiteAndMaliceCpuIdx, nil)
	require.NoError(t, g.Discard(0, 0))
	// 相手 (CPU) 側にカードが補充された
	assert.Greater(t, g.GetPlayer(SpiteAndMaliceCpuIdx).HandSize(), 0)
	// completed が空になる
	assert.Equal(t, 0, g.GetCompletedSize())
}

// --- canPlaceOnFoundation ---

func TestSpiteAndMalice_CanPlaceOnFoundation_EmptyRequiresAce(t *testing.T) {
	g := newTestSpiteAndMalice()
	assert.True(t, g.canPlaceOnFoundation(mkCard(CardDesignSpade, 1), 0))
	assert.False(t, g.canPlaceOnFoundation(mkCard(CardDesignSpade, 2), 0))
	// ワイルドは空 foundation にも置ける
	assert.True(t, g.canPlaceOnFoundation(mkCard(CardDesignSpade, SpiteAndMaliceWildValue), 0))
}

func TestSpiteAndMalice_CanPlaceOnFoundation_NilCard(t *testing.T) {
	g := newTestSpiteAndMalice()
	assert.False(t, g.canPlaceOnFoundation(nil, 0))
}

// --- JSON ---

func TestSpiteAndMalice_JSONRoundTrip(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetPlayerHand(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignSpade, 1)})
	require.NoError(t, g.PlayFromHand(0, 0))

	data, err := json.Marshal(g)
	require.NoError(t, err)

	g2 := NewDefaultSpiteAndMalice()
	require.NoError(t, json.Unmarshal(data, g2))
	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetCurrent(), g2.GetCurrent())
	assert.Equal(t, g.GetMoveCount(), g2.GetMoveCount())
	assert.Equal(t, g.GetFoundationTopValue(0), g2.GetFoundationTopValue(0))
	assert.Equal(t, g.GetPlayer(SpiteAndMaliceHumanIdx).HandSize(), g2.GetPlayer(SpiteAndMaliceHumanIdx).HandSize())
}

func TestSpiteAndMalice_Unmarshal_ResetsWinnerIfNotGameOver(t *testing.T) {
	// 不正ペイロード: phase=Playing だが winner=1 -> リセットされる
	payload := []byte(`{"ph":0,"wn":1,"cu":0,"mc":0,"cf":{"gs":20,"cd":1}}`)
	g := NewDefaultSpiteAndMalice()
	require.NoError(t, json.Unmarshal(payload, g))
	assert.Equal(t, -1, g.GetWinner())
}

// --- Player ---

func TestSpiteAndMalicePlayer_HandOps(t *testing.T) {
	p := NewSpiteAndMalicePlayer(false)
	p.AddToHand(mkCard(CardDesignSpade, 1))
	p.AddToHand(mkCard(CardDesignHeart, 2))
	assert.Equal(t, 2, p.HandSize())
	removed := p.RemoveFromHand(0)
	assert.Equal(t, 1, removed.GetValue())
	assert.Equal(t, 1, p.HandSize())
	// 範囲外
	assert.Nil(t, p.RemoveFromHand(-1))
	assert.Nil(t, p.RemoveFromHand(42))
}

func TestSpiteAndMalicePlayer_GoalOps(t *testing.T) {
	p := NewSpiteAndMalicePlayer(false)
	p.AddToGoal(mkCard(CardDesignSpade, 1)) // 底
	p.AddToGoal(mkCard(CardDesignHeart, 2)) // 1 枚目 (底がさらに下になり、最後に追加した 2 が一番下 -> つまり 1 がトップ)
	// 末尾がトップなので goal[len-1] = Spade 1
	require.NotNil(t, p.GoalTop())
	assert.Equal(t, 1, p.GoalTop().GetValue())
	assert.Equal(t, 2, p.GoalSize())
	p.PopGoal()
	assert.Equal(t, 2, p.GoalTop().GetValue())
	p.PopGoal()
	assert.Nil(t, p.GoalTop())
	assert.Nil(t, p.PopGoal())
}

func TestSpiteAndMalicePlayer_SideOps(t *testing.T) {
	p := NewSpiteAndMalicePlayer(false)
	p.PushSide(0, mkCard(CardDesignSpade, 1))
	p.PushSide(0, mkCard(CardDesignHeart, 2))
	assert.Equal(t, 2, p.SideSize(0))
	assert.Equal(t, 2, p.SideTop(0).GetValue())
	p.PopSide(0)
	assert.Equal(t, 1, p.SideTop(0).GetValue())
	// 範囲外
	p.PushSide(-1, mkCard(CardDesignSpade, 1))
	p.PushSide(SpiteAndMaliceSideCnt, mkCard(CardDesignSpade, 1))
	assert.Nil(t, p.SideTop(-1))
	assert.Nil(t, p.SideTop(SpiteAndMaliceSideCnt))
	assert.Equal(t, 0, p.SideSize(-1))
	assert.Equal(t, 0, p.SideSize(SpiteAndMaliceSideCnt))
	assert.Nil(t, p.PopSide(-1))
	assert.Nil(t, p.PopSide(SpiteAndMaliceSideCnt))
	assert.Nil(t, p.GetSide(-1))
}

func TestSpiteAndMalice_EffectiveValue_ChainedWilds(t *testing.T) {
	g := newTestSpiteAndMalice()
	// 1, K (=2), K (=3) と積む
	g.SetFoundation(0, []*Card{
		mkCard(CardDesignSpade, 1),
		mkCard(CardDesignHeart, SpiteAndMaliceWildValue),
		mkCard(CardDesignClover, SpiteAndMaliceWildValue),
	})
	assert.Equal(t, 3, g.GetFoundationTopValue(0))
	// 次は 4 が置けるはず
	g.SetPlayerHand(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignDiamond, 4)})
	require.NoError(t, g.PlayFromHand(0, 0))
}

func TestSpiteAndMalice_GetFoundationTopValue_Invalid(t *testing.T) {
	g := newTestSpiteAndMalice()
	assert.Equal(t, 0, g.GetFoundationTopValue(-1))
	assert.Equal(t, 0, g.GetFoundationTopValue(SpiteAndMaliceFoundationCnt))
	// 空ファウンデーション
	assert.Equal(t, 0, g.GetFoundationTopValue(0))
}

func TestSpiteAndMalice_GetPlayer_Invalid(t *testing.T) {
	g := newTestSpiteAndMalice()
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(SpiteAndMalicePlayerCnt))
}

func TestSpiteAndMalice_SetConfig(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetConfig(SpiteAndMaliceConfig{GoalSize: 10, CpuDifficulty: SpiteAndMaliceCpuDifficultyEasy})
	assert.Equal(t, 10, g.GetConfig().GoalSize)
}

// pickDiscardSide の全枝: (1) 空サイドあり、(2) 全サイド非空で最短差分
func TestSpiteAndMalice_CpuStep_DiscardSideSelection(t *testing.T) {
	g := NewSpiteAndMalice(NewTrumpCardsWithDecks(2, 0), SpiteAndMaliceConfig{
		GoalSize:      SpiteAndMaliceGoalSizeDefault,
		CpuDifficulty: SpiteAndMaliceCpuDifficultyNormal,
	})
	g.Reset()
	g.SetCurrent(SpiteAndMaliceCpuIdx)
	// 全サイド非空
	g.SetPlayerSide(SpiteAndMaliceCpuIdx, 0, []*Card{mkCard(CardDesignSpade, 2)})
	g.SetPlayerSide(SpiteAndMaliceCpuIdx, 1, []*Card{mkCard(CardDesignSpade, 8)})
	g.SetPlayerSide(SpiteAndMaliceCpuIdx, 2, []*Card{mkCard(CardDesignSpade, 11)})
	g.SetPlayerSide(SpiteAndMaliceCpuIdx, 3, []*Card{mkCard(CardDesignSpade, 6)})
	g.SetPlayerHand(SpiteAndMaliceCpuIdx, []*Card{mkCard(CardDesignHeart, 7)})
	g.SetPlayerGoal(SpiteAndMaliceCpuIdx, []*Card{mkCard(CardDesignHeart, 9)})
	// Play 不可能なので Discard 発動。7 に最も近い 8 (side 1) か 6 (side 3) の
	// どちらかが選ばれる。index=3 は先にループで見つかるが、8 の方が近い。
	// 実装はまず |8-7|=1 を見つけ、|6-7|=1 も同じなので先勝ち (side 1)。
	require.NoError(t, g.CpuStep())
	// side 1 が 2 枚になっている
	assert.Equal(t, 2, g.GetPlayer(SpiteAndMaliceCpuIdx).SideSize(1))
}

func TestSpiteAndMalice_CpuStep_EasyDifficulty(t *testing.T) {
	g := NewSpiteAndMalice(NewTrumpCardsWithDecks(2, 0), SpiteAndMaliceConfig{
		GoalSize:      SpiteAndMaliceGoalSizeDefault,
		CpuDifficulty: SpiteAndMaliceCpuDifficultyEasy,
	})
	g.Reset()
	g.SetCurrent(SpiteAndMaliceCpuIdx)
	g.SetPlayerHand(SpiteAndMaliceCpuIdx, []*Card{mkCard(CardDesignSpade, 5), mkCard(CardDesignHeart, 8)})
	g.SetPlayerGoal(SpiteAndMaliceCpuIdx, []*Card{mkCard(CardDesignHeart, 9)})
	// ファウンデーション空 -> プレイ不可 -> Easy は先頭 (5) を捨てる
	require.NoError(t, g.CpuStep())
	// 5 がどこかのサイドに行った
	cpu := g.GetPlayer(SpiteAndMaliceCpuIdx)
	total := 0
	for s := range SpiteAndMaliceSideCnt {
		total += cpu.SideSize(s)
	}
	assert.Equal(t, 1, total)
}

// --- AutoComplete ---

func TestSpiteAndMalice_AutoComplete_PlaysGoalThenHandThenSide(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetCurrent(SpiteAndMaliceHumanIdx)
	// Foundation 0 currently holds 1; legal next play is 2.
	g.SetFoundation(0, []*Card{mkCard(CardDesignSpade, 1)})
	// Hand has the 2 (playable on foundation 0) and a 9 (no anchor).
	g.SetPlayerHand(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignHeart, 2), mkCard(CardDesignHeart, 9)})
	// Goal top is 3 — playable after the 2.
	g.SetPlayerGoal(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignClover, 3)})

	require.NoError(t, g.AutoComplete())

	// Foundation 0 should now have 1, 2, 3 stacked.
	assert.Equal(t, 3, g.GetFoundationTopValue(0))
	// Goal got drained, hand drained the 2; the 9 stays.
	human := g.GetPlayer(SpiteAndMaliceHumanIdx)
	assert.Equal(t, 0, human.GoalSize())
	assert.Equal(t, 1, human.HandSize())
	// Auto-complete should not rotate the turn.
	assert.Equal(t, SpiteAndMaliceHumanIdx, g.GetCurrent())
}

func TestSpiteAndMalice_AutoComplete_NoOpWhenNothingPlayable(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetCurrent(SpiteAndMaliceHumanIdx)
	// Foundation empty; hand has no Ace and no wild — nothing legal.
	g.SetPlayerHand(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignSpade, 5), mkCard(CardDesignHeart, 8)})
	// Reset() leaves the goal pile shuffle-dependent; pin its top to a
	// non-playable mid-rank so AutoComplete never plays from goal either.
	g.SetPlayerGoal(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignClover, 7)})
	for s := range SpiteAndMaliceSideCnt {
		g.SetPlayerSide(SpiteAndMaliceHumanIdx, s, nil)
	}
	moveCountBefore := g.GetMoveCount()
	require.NoError(t, g.AutoComplete())
	assert.Equal(t, moveCountBefore, g.GetMoveCount(), "no moves should have been made")
}

func TestSpiteAndMalice_AutoComplete_RefusesOnCpuTurn(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetCurrent(SpiteAndMaliceCpuIdx)
	err := g.AutoComplete()
	require.Error(t, err)
}

func TestSpiteAndMalice_CanAutoComplete_TrueWhenMoveAvailable(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetCurrent(SpiteAndMaliceHumanIdx)
	g.SetFoundation(0, []*Card{mkCard(CardDesignSpade, 1)})
	g.SetPlayerHand(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignHeart, 2)})
	assert.True(t, g.CanAutoComplete())
}

func TestSpiteAndMalice_CanAutoComplete_FalseOnCpuTurn(t *testing.T) {
	g := newTestSpiteAndMalice()
	g.SetCurrent(SpiteAndMaliceCpuIdx)
	g.SetFoundation(0, []*Card{mkCard(CardDesignSpade, 1)})
	g.SetPlayerHand(SpiteAndMaliceCpuIdx, []*Card{mkCard(CardDesignHeart, 2)})
	assert.False(t, g.CanAutoComplete())
}

func TestSpiteAndMalicePlayer_JSONRoundTrip(t *testing.T) {
	p := NewSpiteAndMalicePlayer(true)
	p.AddToHand(mkCard(CardDesignSpade, 1))
	p.AddToGoal(mkCard(CardDesignHeart, 2))
	p.PushSide(0, mkCard(CardDesignClover, 3))
	data, err := json.Marshal(p)
	require.NoError(t, err)
	q := &SpiteAndMalicePlayer{}
	require.NoError(t, json.Unmarshal(data, q))
	assert.True(t, q.GetIsCpu())
	assert.Equal(t, 1, q.HandSize())
	assert.Equal(t, 1, q.GoalSize())
	assert.Equal(t, 1, q.SideSize(0))
}

// **ゴール札が出せるかは勝敗条件そのもの。**Web の isGoalTopPlayableToFoundation と
// 同じ規則 (空の基礎札には A、K はワイルド、完成した山には置けない) であること。
func TestSpiteAndMalice_IsGoalTopPlayable(t *testing.T) {
	t.Run("out of range and empty goal", func(t *testing.T) {
		g := newTestSpiteAndMalice()
		assert.False(t, g.IsGoalTopPlayable(-1))
		assert.False(t, g.IsGoalTopPlayable(99))
		g.SetPlayerGoal(SpiteAndMaliceHumanIdx, nil)
		assert.False(t, g.IsGoalTopPlayable(SpiteAndMaliceHumanIdx))
	})

	t.Run("an ace opens an empty foundation", func(t *testing.T) {
		g := newTestSpiteAndMalice()
		g.SetPlayerGoal(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignSpade, 1)})
		assert.True(t, g.IsGoalTopPlayable(SpiteAndMaliceHumanIdx))
	})

	t.Run("only the next rank fits a started foundation", func(t *testing.T) {
		g := newTestSpiteAndMalice()
		for i := range SpiteAndMaliceFoundationCnt {
			g.SetFoundation(i, []*Card{mkCard(CardDesignSpade, 1), mkCard(CardDesignHeart, 2)})
		}
		g.SetPlayerGoal(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignDiamond, 5)})
		assert.False(t, g.IsGoalTopPlayable(SpiteAndMaliceHumanIdx))
		g.SetPlayerGoal(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignDiamond, 3)})
		assert.True(t, g.IsGoalTopPlayable(SpiteAndMaliceHumanIdx))
	})

	t.Run("a king is wild but not on a completed pile", func(t *testing.T) {
		g := newTestSpiteAndMalice()
		full := make([]*Card, 0, SpiteAndMaliceFoundationMax)
		for v := 1; v <= SpiteAndMaliceFoundationMax; v++ {
			full = append(full, mkCard(CardDesignSpade, v))
		}
		for i := range SpiteAndMaliceFoundationCnt {
			g.SetFoundation(i, full)
		}
		g.SetPlayerGoal(SpiteAndMaliceHumanIdx, []*Card{mkCard(CardDesignClover, SpiteAndMaliceWildValue)})
		assert.False(t, g.IsGoalTopPlayable(SpiteAndMaliceHumanIdx))

		g.SetFoundation(0, []*Card{mkCard(CardDesignSpade, 1)})
		assert.True(t, g.IsGoalTopPlayable(SpiteAndMaliceHumanIdx))
	})
}
