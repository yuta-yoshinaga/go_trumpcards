//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCalculation() *Calculation {
	c := NewCalculation(NewTrumpCards(0))
	c.Reset()
	return c
}

func TestCalculation_Reset_InitialState(t *testing.T) {
	c := newTestCalculation()

	assert.Equal(t, CalculationPhasePlaying, c.GetPhase())
	assert.Equal(t, 0, c.GetMoveCount())
	assert.Empty(t, c.GetActionLog())
	assert.False(t, c.IsStalemate())

	// 各ファンデーションは A,2,3,4 1枚ずつ
	f := c.GetFoundations()
	for i := range CalculationFoundationCnt {
		require.Len(t, f[i], 1, "foundation %d should start with 1 card", i)
		assert.Equal(t, i+1, f[i][0].GetValue(), "foundation %d base", i)
	}

	// ウェイストは空
	for _, w := range c.GetWastes() {
		assert.Empty(t, w)
	}

	// ストックは 52 - 4 = 48 枚
	assert.Equal(t, 48, c.GetStockCount())
}

func TestCalculation_Reset_ClearsPreviousState(t *testing.T) {
	c := newTestCalculation()
	_ = c.PlayStockToWaste(0)
	require.Greater(t, c.GetMoveCount(), 0)
	c.Reset()
	assert.Equal(t, 0, c.GetMoveCount())
	assert.Empty(t, c.GetActionLog())
	assert.Equal(t, CalculationPhasePlaying, c.GetPhase())
}

func TestCalculationNextValue(t *testing.T) {
	tests := []struct {
		name string
		v    int
		step int
		want int
	}{
		{"A + 1 = 2", 1, 1, 2},
		{"K + 1 = A (wrap)", 13, 1, 1},
		{"Q + 2 = A (wrap)", 12, 2, 1},
		{"J + 2 = K (no wrap)", 11, 2, 13},
		{"2 + 2 = 4", 2, 2, 4},
		{"3 + 3 = 6", 3, 3, 6},
		{"10 + 3 = K", 10, 3, 13},
		{"J + 3 = A (wrap)", 11, 3, 1},
		{"4 + 4 = 8", 4, 4, 8},
		{"10 + 4 = A (wrap)", 10, 4, 1},
		{"9 + 4 = K", 9, 4, 13},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, calculationNextValue(tt.v, tt.step))
		})
	}
}

// setStockWithTop sets stock so that `top` is the rightmost (LIFO top).
func setStockWithTop(c *Calculation, top *Card) {
	c.SetStock([]*Card{top})
}

func TestCalculation_PlayStockToFoundation_Valid(t *testing.T) {
	c := newTestCalculation()
	// F0 base = A (1), next expected = 2
	setStockWithTop(c, NewCard(CardDesignSpade, 2, false))

	err := c.PlayStockToFoundation(0)
	require.NoError(t, err)

	f := c.GetFoundations()
	require.Len(t, f[0], 2)
	assert.Equal(t, 2, f[0][1].GetValue())
	assert.Equal(t, 0, c.GetStockCount())
	assert.Equal(t, 1, c.GetMoveCount())
	require.Len(t, c.GetActionLog(), 1)
}

func TestCalculation_PlayStockToFoundation_InvalidValue(t *testing.T) {
	c := newTestCalculation()
	// F0 needs 2, push a 5
	setStockWithTop(c, NewCard(CardDesignSpade, 5, false))

	err := c.PlayStockToFoundation(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot place card")
}

func TestCalculation_PlayStockToFoundation_InvalidIdx(t *testing.T) {
	c := newTestCalculation()
	setStockWithTop(c, NewCard(CardDesignSpade, 2, false))

	assert.Error(t, c.PlayStockToFoundation(-1))
	assert.Error(t, c.PlayStockToFoundation(CalculationFoundationCnt))
}

func TestCalculation_PlayStockToFoundation_EmptyStock(t *testing.T) {
	c := newTestCalculation()
	c.SetStock([]*Card{})

	err := c.PlayStockToFoundation(0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stock is empty")
}

func TestCalculation_PlayStockToFoundation_NotPlaying(t *testing.T) {
	c := newTestCalculation()
	c.SetPhase(CalculationPhaseGameOver)
	err := c.PlayStockToFoundation(0)
	assert.Error(t, err)
}

func TestCalculation_PlayStockToWaste_Valid(t *testing.T) {
	c := newTestCalculation()
	before := c.GetStockCount()
	err := c.PlayStockToWaste(1)
	require.NoError(t, err)
	assert.Equal(t, before-1, c.GetStockCount())
	assert.Len(t, c.GetWastes()[1], 1)
	assert.Equal(t, 1, c.GetMoveCount())
}

func TestCalculation_PlayStockToWaste_InvalidIdx(t *testing.T) {
	c := newTestCalculation()
	assert.Error(t, c.PlayStockToWaste(-1))
	assert.Error(t, c.PlayStockToWaste(CalculationWasteCnt))
}

func TestCalculation_PlayStockToWaste_EmptyStock(t *testing.T) {
	c := newTestCalculation()
	c.SetStock([]*Card{})
	err := c.PlayStockToWaste(0)
	assert.Error(t, err)
}

func TestCalculation_PlayStockToWaste_NotPlaying(t *testing.T) {
	c := newTestCalculation()
	c.SetPhase(CalculationPhaseGameOver)
	err := c.PlayStockToWaste(0)
	assert.Error(t, err)
}

func TestCalculation_PlayWasteToFoundation_Valid(t *testing.T) {
	c := newTestCalculation()
	// Put a 2 on waste 0; F0 expects 2
	w := [CalculationWasteCnt][]*Card{{NewCard(CardDesignSpade, 2, false)}, nil, nil, nil}
	c.SetWastes(w)

	err := c.PlayWasteToFoundation(0, 0)
	require.NoError(t, err)
	assert.Empty(t, c.GetWastes()[0])
	assert.Len(t, c.GetFoundations()[0], 2)
}

func TestCalculation_PlayWasteToFoundation_InvalidIndex(t *testing.T) {
	c := newTestCalculation()
	w := [CalculationWasteCnt][]*Card{{NewCard(CardDesignSpade, 2, false)}, nil, nil, nil}
	c.SetWastes(w)

	assert.Error(t, c.PlayWasteToFoundation(-1, 0))
	assert.Error(t, c.PlayWasteToFoundation(CalculationWasteCnt, 0))
	assert.Error(t, c.PlayWasteToFoundation(0, -1))
	assert.Error(t, c.PlayWasteToFoundation(0, CalculationFoundationCnt))
}

func TestCalculation_PlayWasteToFoundation_EmptyWaste(t *testing.T) {
	c := newTestCalculation()
	err := c.PlayWasteToFoundation(0, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "waste is empty")
}

func TestCalculation_PlayWasteToFoundation_InvalidValue(t *testing.T) {
	c := newTestCalculation()
	w := [CalculationWasteCnt][]*Card{{NewCard(CardDesignSpade, 5, false)}, nil, nil, nil}
	c.SetWastes(w)
	err := c.PlayWasteToFoundation(0, 0) // F0 expects 2
	assert.Error(t, err)
}

func TestCalculation_PlayWasteToFoundation_NotPlaying(t *testing.T) {
	c := newTestCalculation()
	c.SetPhase(CalculationPhaseGameOver)
	err := c.PlayWasteToFoundation(0, 0)
	assert.Error(t, err)
}

func TestCalculation_GameClear(t *testing.T) {
	c := newTestCalculation()

	// Foundation 0..3 に 12枚ずつ（A から J まで、等差数列で）埋めて、最後の K でクリア
	var f [CalculationFoundationCnt][]*Card
	for i := range CalculationFoundationCnt {
		step := i + 1
		v := step
		pile := []*Card{NewCard(CardDesignSpade, v, false)}
		for range 11 {
			v = calculationNextValue(v, step)
			pile = append(pile, NewCard(CardDesignSpade, v, false))
		}
		f[i] = pile
		// 12 枚になったので K が足りない（step=1 では V=12 = Q なので次は K）
	}
	c.SetFoundations(f)
	// 各ファンデーションの最後の値の確認
	// Now play stock=K onto F0 (step=1, top=Q → next=K).
	setStockWithTop(c, NewCard(CardDesignSpade, 13, false))
	err := c.PlayStockToFoundation(0)
	require.NoError(t, err)
	// F0 だけ完成だが他はまだ 12 枚なのでゲームクリアにはならない
	assert.Equal(t, CalculationPhasePlaying, c.GetPhase())

	// 他の3つも完成させる
	for i := 1; i < CalculationFoundationCnt; i++ {
		c.SetStock([]*Card{NewCard(CardDesignSpade, 13, false)})
		require.NoError(t, c.PlayStockToFoundation(i))
	}
	assert.Equal(t, CalculationPhaseGameClear, c.GetPhase())
}

func TestCalculation_GiveUp(t *testing.T) {
	c := newTestCalculation()
	c.GiveUp()
	assert.Equal(t, CalculationPhaseGameOver, c.GetPhase())
	require.NotEmpty(t, c.GetActionLog())
	assert.Equal(t, "giveup", c.GetActionLog()[len(c.GetActionLog())-1].ActionType)
}

func TestCalculation_GiveUp_WhenNotPlaying_NoOp(t *testing.T) {
	c := newTestCalculation()
	c.SetPhase(CalculationPhaseGameClear)
	c.GiveUp()
	// Still GameClear, no action log added
	assert.Equal(t, CalculationPhaseGameClear, c.GetPhase())
}

func TestCalculation_Undo_AfterPlay(t *testing.T) {
	c := newTestCalculation()
	topBefore := c.GetStockTop()
	require.NotNil(t, topBefore)

	// Move stock top to waste 0
	require.NoError(t, c.PlayStockToWaste(0))
	require.Len(t, c.GetWastes()[0], 1)

	require.True(t, c.CanUndo())
	require.NoError(t, c.Undo())
	assert.Empty(t, c.GetWastes()[0])
	assert.Equal(t, topBefore, c.GetStockTop())
}

func TestCalculation_Undo_Empty(t *testing.T) {
	c := newTestCalculation()
	assert.False(t, c.CanUndo())
	err := c.Undo()
	assert.Error(t, err)
}

func TestCalculation_Undo_NotPlaying(t *testing.T) {
	c := newTestCalculation()
	_ = c.PlayStockToWaste(0)
	c.SetPhase(CalculationPhaseGameOver)
	err := c.Undo()
	assert.Error(t, err)
}

func TestCalculation_UndoN(t *testing.T) {
	c := newTestCalculation()
	stockBefore := c.GetStockCount()
	require.NoError(t, c.PlayStockToWaste(0))
	require.NoError(t, c.PlayStockToWaste(1))
	require.NoError(t, c.UndoN(2))
	assert.Equal(t, stockBefore, c.GetStockCount())
}

func TestCalculation_UndoN_Error(t *testing.T) {
	c := newTestCalculation()
	// 0 action history -> UndoN(1) should fail wrapped
	err := c.UndoN(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undo step 1 failed")
}

func TestCalculation_UndoToEscape(t *testing.T) {
	c := newTestCalculation()
	assert.Equal(t, 0, c.UndoToEscape())

	// Simulate stalemate that started mid-history
	_ = c.PlayStockToWaste(0) // snapshot #1
	_ = c.PlayStockToWaste(1) // snapshot #2
	c.SetIsStalemate(true)

	// stalemate should require 0 undos if no history has non-stalemate
	// current history entries were recorded before isStalemate was set → false → escape with 1 undo
	n := c.UndoToEscape()
	assert.Equal(t, 1, n)
}

func TestCalculation_UndoToEscape_Unreachable(t *testing.T) {
	c := newTestCalculation()
	// Hand-craft history entries that all claim stalemate
	c.SetIsStalemate(true)
	c.history = []*calculationSnapshot{{isStalemate: true}, {isStalemate: true}}
	assert.Equal(t, -1, c.UndoToEscape())
}

func TestCalculation_GetHint_StockToFoundation(t *testing.T) {
	c := newTestCalculation()
	setStockWithTop(c, NewCard(CardDesignSpade, 2, false))

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "stock", h.FromZone)
	assert.Equal(t, 0, h.FoundationIdx)
}

func TestCalculation_GetHint_WasteToFoundation(t *testing.T) {
	c := newTestCalculation()
	// Empty stock, waste[2] top = 2 which fits foundation 0
	c.SetStock([]*Card{})
	w := [CalculationWasteCnt][]*Card{nil, nil, {NewCard(CardDesignSpade, 2, false)}, nil}
	c.SetWastes(w)

	h := c.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "waste", h.FromZone)
	assert.Equal(t, 2, h.WasteIdx)
	assert.Equal(t, 0, h.FoundationIdx)
}

func TestCalculation_GetHint_NonePlaying(t *testing.T) {
	c := newTestCalculation()
	c.SetPhase(CalculationPhaseGameOver)
	assert.Nil(t, c.GetHint())
}

func TestCalculation_GetHint_NoneAvailable(t *testing.T) {
	c := newTestCalculation()
	// Foundations full-ish but no playable card anywhere
	c.SetStock([]*Card{})
	// Set all waste tops to J (11); the four next-expected values are all specific ranks
	// that will not be 11 simultaneously — pick a card value that cannot fit any foundation.
	// Foundation 0..3 with base A,2,3,4 expect 2,4,6,8. So 11 fits nowhere.
	w := [CalculationWasteCnt][]*Card{
		{NewCard(CardDesignSpade, 11, false)},
		{NewCard(CardDesignHeart, 11, false)},
		{NewCard(CardDesignDiamond, 11, false)},
		{NewCard(CardDesignClover, 11, false)},
	}
	c.SetWastes(w)
	assert.Nil(t, c.GetHint())
}

func TestCalculation_Stalemate_DetectionEmptyStock(t *testing.T) {
	c := newTestCalculation()
	c.SetStock([]*Card{NewCard(CardDesignSpade, 11, false)}) // will go to waste
	// Stash into waste (the only legal thing to do since F expects 2, not J)
	require.NoError(t, c.PlayStockToWaste(0))
	// Now stock empty, waste top = J, no foundation accepts it → stalemate
	assert.True(t, c.IsStalemate())
}

func TestCalculation_Stalemate_NotStalemateWhenStockRemains(t *testing.T) {
	c := newTestCalculation()
	_ = c.PlayStockToWaste(0)
	assert.False(t, c.IsStalemate())
}

func TestCalculation_AutoComplete_StockNonEmpty(t *testing.T) {
	c := newTestCalculation()
	err := c.AutoComplete()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stock is not empty")
}

func TestCalculation_AutoComplete_NotPlaying(t *testing.T) {
	c := newTestCalculation()
	c.SetPhase(CalculationPhaseGameOver)
	err := c.AutoComplete()
	assert.Error(t, err)
}

func TestCalculation_AutoComplete_PlaysWasteTopsToFoundations(t *testing.T) {
	c := newTestCalculation()
	c.SetStock([]*Card{})

	// Set up waste tops that chain-fit foundation 0 (step 1): needs 2, then 3, then 4
	w := [CalculationWasteCnt][]*Card{
		{NewCard(CardDesignSpade, 4, false)},
		{NewCard(CardDesignSpade, 3, false)},
		{NewCard(CardDesignSpade, 2, false)},
		{},
	}
	c.SetWastes(w)

	require.NoError(t, c.AutoComplete())
	// Foundation 0 should now have A,2,3,4 (4 cards total).
	// But wait — after we play 2 onto F0, F0=A,2. Then we play 3 (F0 needs 3=2+1)? Yes. Then 4.
	// Note: foundation 1 (step 2) would accept 4 (base 2 + 2 = 4). So the 4 in waste[0] could
	// actually go to foundation 1 first. findFoundation picks the first match, so F0 (step 1)
	// at base 1 expects 2 — it won't take 4. F1 (step 2) base 2 expects 4 → match. So 4 goes
	// to F1 first. After that F1=[2,4]. Waste tops 3 and 2: F0 expects 2 → take 2 first.
	// After F0=[1,2], expects 3 → take 3. Foundation 1 now only has [2,4], expects 6.
	// Final: F0=[1,2,3], F1=[2,4], F2=[3], F3=[4], wastes 0..2 empty.
	f := c.GetFoundations()
	assert.Equal(t, 3, len(f[0]))
	assert.Equal(t, 2, len(f[1]))
}

func TestCalculation_JSON_RoundTrip(t *testing.T) {
	c := newTestCalculation()
	_ = c.PlayStockToWaste(0)

	data, err := json.Marshal(c)
	require.NoError(t, err)

	c2 := &Calculation{}
	require.NoError(t, json.Unmarshal(data, c2))

	assert.Equal(t, c.GetPhase(), c2.GetPhase())
	assert.Equal(t, c.GetMoveCount(), c2.GetMoveCount())
	assert.Equal(t, c.GetStockCount(), c2.GetStockCount())
	assert.Equal(t, len(c.GetActionLog()), len(c2.GetActionLog()))
}

func TestCalculation_UnmarshalJSON_InvalidJSON(t *testing.T) {
	c := &Calculation{}
	assert.Error(t, json.Unmarshal([]byte("not json"), c))
}

func TestCalculation_UnmarshalJSON_OversizedStock(t *testing.T) {
	bigStock := make([]*Card, calculationMaxSliceLen+1)
	for i := range bigStock {
		bigStock[i] = NewCard(CardDesignSpade, 1, false)
	}
	data, err := json.Marshal(calculationJSON{Stock: bigStock})
	require.NoError(t, err)

	c := &Calculation{}
	err = json.Unmarshal(data, c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

func TestCalculation_UnmarshalJSON_OversizedFoundation(t *testing.T) {
	var fd [CalculationFoundationCnt][]*Card
	big := make([]*Card, calculationMaxSliceLen+1)
	for i := range big {
		big[i] = NewCard(CardDesignSpade, 1, false)
	}
	fd[1] = big
	data, err := json.Marshal(calculationJSON{Foundations: fd})
	require.NoError(t, err)

	c := &Calculation{}
	err = json.Unmarshal(data, c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "foundation 1 exceeds")
}

func TestCalculation_UnmarshalJSON_OversizedWaste(t *testing.T) {
	var wa [CalculationWasteCnt][]*Card
	big := make([]*Card, calculationMaxSliceLen+1)
	for i := range big {
		big[i] = NewCard(CardDesignSpade, 1, false)
	}
	wa[2] = big
	data, err := json.Marshal(calculationJSON{Wastes: wa})
	require.NoError(t, err)

	c := &Calculation{}
	err = json.Unmarshal(data, c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "waste 2 exceeds")
}

func TestCalculation_UnmarshalJSON_NilSlices(t *testing.T) {
	data, err := json.Marshal(calculationJSON{})
	require.NoError(t, err)
	c := &Calculation{}
	require.NoError(t, json.Unmarshal(data, c))
	assert.NotNil(t, c.trumpCards)
	for i := range CalculationFoundationCnt {
		assert.NotNil(t, c.foundations[i])
	}
	for i := range CalculationWasteCnt {
		assert.NotNil(t, c.wastes[i])
	}
	assert.NotNil(t, c.stock)
	assert.NotNil(t, c.actionLog)
}

func TestCalculation_CanPlaceOnFoundation_FullPile(t *testing.T) {
	c := newTestCalculation()
	// Fill F0 with 13 cards of value A..K
	pile := make([]*Card, 13)
	for i := range pile {
		pile[i] = NewCard(CardDesignSpade, i+1, false)
	}
	f := c.GetFoundations()
	f[0] = pile
	c.SetFoundations(f)
	assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, 2, false), 0))
}

func TestCalculation_GetStockTop_Empty(t *testing.T) {
	c := newTestCalculation()
	c.SetStock([]*Card{})
	assert.Nil(t, c.GetStockTop())
}

func TestCalculation_FindFoundation_NotFound(t *testing.T) {
	c := newTestCalculation()
	// J (11) fits nowhere at initial state (foundations need 2,4,6,8)
	assert.Equal(t, -1, c.findFoundation(NewCard(CardDesignSpade, 11, false)))
}

// **各列が +1/+2/+3/+4 ずつ 13 を法として進む (#4794)。**Web は次に置ける
// ランクをバッジで常時出しているのに、CUI は毎手この暗算を強いていた。
func TestCalculation_GetNextFoundationRank(t *testing.T) {
	c := newTestCalculation()

	// 初期状態は各ファンデーションに A,2,3,4 が1枚ずつ。
	// 列0 (+1) は A の次で 2、列1 (+2) は 2 の次で 4、列2 (+3) は 3 の次で 6、
	// 列3 (+4) は 4 の次で 8。
	t.Run("applies each pile's own step", func(t *testing.T) {
		assert.Equal(t, 2, c.GetNextFoundationRank(0))
		assert.Equal(t, 4, c.GetNextFoundationRank(1))
		assert.Equal(t, 6, c.GetNextFoundationRank(2))
		assert.Equal(t, 8, c.GetNextFoundationRank(3))
	})

	// **13 を超えたら折り返す。**単なる足し算だと 14 以上を案内してしまう。
	t.Run("wraps past the king", func(t *testing.T) {
		assert.Equal(t, calculationNextValue(13, 1), 1, "K の次は A (+1 列)")
		assert.Equal(t, calculationNextValue(12, 4), 3, "Q の +4 は 3")
	})

	// **13枚そろった山に「次」は無い。**出すと、置けない札を案内することになる。
	t.Run("reports nothing once a pile is complete", func(t *testing.T) {
		full := newTestCalculation()
		pile := make([]*Card, CardValueMax)
		for i := range pile {
			pile[i] = NewCard(CardDesignSpade, i+1, false)
		}
		full.foundations[0] = pile
		assert.Equal(t, 0, full.GetNextFoundationRank(0))
	})

	t.Run("reports nothing for an out-of-range pile", func(t *testing.T) {
		assert.Equal(t, 0, c.GetNextFoundationRank(-1))
		assert.Equal(t, 0, c.GetNextFoundationRank(CalculationFoundationCnt))
	})

	// **案内したランクは実際に置ける。**別実装だと置けない札を案内する。
	t.Run("the rank it names is the one canPlaceOnFoundation accepts", func(t *testing.T) {
		for i := range CalculationFoundationCnt {
			next := c.GetNextFoundationRank(i)
			require.NotEqual(t, 0, next, "foundation %d", i)
			assert.True(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, next, false), i),
				"foundation %d に %d が置けない", i, next)
			// ひとつ違うランクは弾かれる。
			wrong := next%CardValueMax + 1
			assert.False(t, c.canPlaceOnFoundation(NewCard(CardDesignSpade, wrong, false), i),
				"foundation %d に %d が置けてしまう", i, wrong)
		}
	})
}

// #5551: +1/+2/+3/+4 の歩幅を何手も辿るのは暗算負荷が高い。Web は最大6手先まで
// バッジで出しているのに、CUI は 1 手先しか出していなかった。
func TestCalculation_GetUpcomingFoundationRanks(t *testing.T) {
	c := NewDefaultCalculation()
	c.Reset()

	set := func(fIdx int, values ...int) {
		f := c.GetFoundations()
		pile := make([]*Card, 0, len(values))
		for _, v := range values {
			pile = append(pile, NewCard(CardDesignSpade, v, false))
		}
		f[fIdx] = pile
		c.SetFoundations(f)
	}

	// +3 の列 (index 2) が 3 で止まっている: 6, 9, 12, 2, 5, 8。
	set(2, 3)
	assert.Equal(t, []int{6, 9, 12, 2, 5, 8}, c.GetUpcomingFoundationRanks(2, 6))

	// +1 の列 (index 0) が 1 のとき: 2, 3, 4。
	set(0, 1)
	assert.Equal(t, []int{2, 3, 4}, c.GetUpcomingFoundationRanks(0, 3))

	// **13枚に達したら打ち切る。**残り2枚しか置けない山で6手先は返さない。
	full := make([]int, 0, 11)
	for v := 1; v <= 11; v++ {
		full = append(full, v)
	}
	set(0, full...)
	assert.Len(t, c.GetUpcomingFoundationRanks(0, 6), 2)

	// 完成した山は空。
	full = append(full, 12, 13)
	set(0, full...)
	assert.Empty(t, c.GetUpcomingFoundationRanks(0, 6))

	// 範囲外は空。
	assert.Empty(t, c.GetUpcomingFoundationRanks(-1, 6))
	assert.Empty(t, c.GetUpcomingFoundationRanks(CalculationFoundationCnt, 6))
}

// **1手先は既存の GetNextFoundationRank と一致すること。**別々に計算していると、
// 同じ画面が 2 つの「次のランク」を出すことになる。
func TestCalculation_UpcomingRanksAgreeWithNextRank(t *testing.T) {
	c := NewDefaultCalculation()
	c.Reset()
	for i := range CalculationFoundationCnt {
		upcoming := c.GetUpcomingFoundationRanks(i, 6)
		if next := c.GetNextFoundationRank(i); next > 0 {
			require.NotEmpty(t, upcoming, "foundation %d", i)
			assert.Equal(t, next, upcoming[0], "foundation %d", i)
		}
	}
}
