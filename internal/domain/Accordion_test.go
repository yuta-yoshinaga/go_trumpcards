//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- helpers ---

func newTestAccordion() *domain.Accordion {
	tc := domain.NewTrumpCards(0)
	return domain.NewAccordion(tc)
}

func setupPlayingAccordion() *domain.Accordion {
	a := newTestAccordion()
	a.Reset()
	return a
}

// setAccordionPiles は指定の (design, value) ペアから 1 枚ずつの山を作って設定する
func setAccordionPiles(a *domain.Accordion, specs ...[2]int) {
	piles := make([][]*domain.Card, 0, len(specs))
	for _, s := range specs {
		piles = append(piles, []*domain.Card{domain.NewCard(s[0], s[1], false)})
	}
	a.SetPiles(piles)
}

// --- Tests ---

func TestNewAccordion(t *testing.T) {
	a := newTestAccordion()
	assert.NotNil(t, a)
	assert.Equal(t, domain.AccordionPhase(0), a.GetPhase())
}

func TestAccordion_Reset(t *testing.T) {
	a := setupPlayingAccordion()
	assert.Equal(t, domain.AccordionPhasePlaying, a.GetPhase())
	assert.Equal(t, 0, a.GetMoveCount())
	assert.Equal(t, domain.AccordionPileCnt, a.GetPileCount())
	// 各パイルは 1 枚のカードを持つ
	totalCards := 0
	for _, pile := range a.GetPiles() {
		assert.Equal(t, 1, len(pile))
		totalCards += len(pile)
	}
	assert.Equal(t, domain.CardCnt, totalCards)
}

func TestAccordion_Move_Offset1_SameSuit(t *testing.T) {
	a := setupPlayingAccordion()
	setAccordionPiles(a,
		[2]int{domain.CardDesignSpade, 5},
		[2]int{domain.CardDesignSpade, 10},
	)
	err := a.Move(1, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, a.GetPileCount())
	assert.Equal(t, 1, a.GetMoveCount())
	// 重ねられた山の一番上は 移動元の top
	top := a.GetPiles()[0][len(a.GetPiles()[0])-1]
	assert.Equal(t, domain.CardDesignSpade, top.GetDesign())
	assert.Equal(t, 10, top.GetValue())
	// 1つの山に 2 枚が入っている
	assert.Equal(t, 2, len(a.GetPiles()[0]))
	// ゲームクリア状態
	assert.Equal(t, domain.AccordionPhaseGameClear, a.GetPhase())
}

func TestAccordion_Move_Offset1_SameRank(t *testing.T) {
	a := setupPlayingAccordion()
	setAccordionPiles(a,
		[2]int{domain.CardDesignHeart, 7},
		[2]int{domain.CardDesignClover, 7},
		[2]int{domain.CardDesignSpade, 2}, // 終了回避のダミー
	)
	err := a.Move(1, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, a.GetPileCount())
}

func TestAccordion_Move_Offset3(t *testing.T) {
	a := setupPlayingAccordion()
	setAccordionPiles(a,
		[2]int{domain.CardDesignSpade, 5},
		[2]int{domain.CardDesignHeart, 9},
		[2]int{domain.CardDesignDiamond, 2},
		[2]int{domain.CardDesignSpade, 12}, // 3つ左 = index 0 と同スート
	)
	err := a.Move(3, 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, a.GetPileCount())
	// 先頭は新しく降ってきた Spade 12
	assert.Equal(t, 12, a.GetPiles()[0][len(a.GetPiles()[0])-1].GetValue())
}

func TestAccordion_Move_InvalidOffset(t *testing.T) {
	a := setupPlayingAccordion()
	setAccordionPiles(a,
		[2]int{domain.CardDesignSpade, 5},
		[2]int{domain.CardDesignSpade, 6},
		[2]int{domain.CardDesignSpade, 7},
	)
	// offset=2 は無効
	err := a.Move(2, 0)
	assert.Error(t, err)
}

func TestAccordion_Move_NoMatch(t *testing.T) {
	a := setupPlayingAccordion()
	setAccordionPiles(a,
		[2]int{domain.CardDesignSpade, 5},
		[2]int{domain.CardDesignHeart, 10},
	)
	err := a.Move(1, 0)
	assert.Error(t, err)
}

func TestAccordion_Move_OutOfRange(t *testing.T) {
	a := setupPlayingAccordion()
	setAccordionPiles(a,
		[2]int{domain.CardDesignSpade, 5},
		[2]int{domain.CardDesignSpade, 6},
	)
	assert.Error(t, a.Move(-1, 0))
	assert.Error(t, a.Move(0, -1))
	assert.Error(t, a.Move(5, 0))
	assert.Error(t, a.Move(0, 5))
}

func TestAccordion_Move_NotPlaying(t *testing.T) {
	a := setupPlayingAccordion()
	a.SetPhase(domain.AccordionPhaseGameOver)
	err := a.Move(1, 0)
	assert.Error(t, err)
}

func TestAccordion_GiveUp(t *testing.T) {
	a := setupPlayingAccordion()
	a.GiveUp()
	assert.Equal(t, domain.AccordionPhaseGameOver, a.GetPhase())
	// 二重呼び出しでも平気
	a.GiveUp()
	assert.Equal(t, domain.AccordionPhaseGameOver, a.GetPhase())
}

func TestAccordion_Hint_Offset3Preferred(t *testing.T) {
	a := setupPlayingAccordion()
	setAccordionPiles(a,
		[2]int{domain.CardDesignSpade, 5},
		[2]int{domain.CardDesignSpade, 6},
		[2]int{domain.CardDesignDiamond, 2},
		[2]int{domain.CardDesignSpade, 12}, // offset=3 一致
	)
	h := a.GetHint()
	assert.NotNil(t, h)
	// offset=3 優先で選ばれる
	assert.Equal(t, 3, h.FromIdx)
	assert.Equal(t, 0, h.ToIdx)
}

func TestAccordion_Hint_NoMove(t *testing.T) {
	a := setupPlayingAccordion()
	setAccordionPiles(a,
		[2]int{domain.CardDesignSpade, 2},
		[2]int{domain.CardDesignHeart, 5},
	)
	h := a.GetHint()
	assert.Nil(t, h)
}

func TestAccordion_IsStalemate_Getter(t *testing.T) {
	a := setupPlayingAccordion()
	assert.False(t, a.IsStalemate())
	a.SetIsStalemate(true)
	assert.True(t, a.IsStalemate())
}

func TestAccordion_Hint_NotPlaying(t *testing.T) {
	a := setupPlayingAccordion()
	a.SetPhase(domain.AccordionPhaseGameOver)
	assert.Nil(t, a.GetHint())
}

func TestAccordion_Undo(t *testing.T) {
	a := setupPlayingAccordion()
	setAccordionPiles(a,
		[2]int{domain.CardDesignSpade, 5},
		[2]int{domain.CardDesignSpade, 10},
		[2]int{domain.CardDesignDiamond, 3},
	)
	assert.False(t, a.CanUndo())
	err := a.Move(1, 0)
	assert.NoError(t, err)
	assert.True(t, a.CanUndo())
	assert.Equal(t, 2, a.GetPileCount())
	err = a.Undo()
	assert.NoError(t, err)
	assert.Equal(t, 3, a.GetPileCount())
	assert.Equal(t, 0, a.GetMoveCount())
	// 履歴が空なら undo 不可
	assert.False(t, a.CanUndo())
	err = a.Undo()
	assert.Error(t, err)
}

func TestAccordion_UndoN(t *testing.T) {
	a := setupPlayingAccordion()
	setAccordionPiles(a,
		[2]int{domain.CardDesignSpade, 5},
		[2]int{domain.CardDesignSpade, 10},
		[2]int{domain.CardDesignSpade, 7},
		[2]int{domain.CardDesignDiamond, 4}, // 4 枚にしてクリアに達しないようにする
	)
	assert.NoError(t, a.Move(1, 0))
	assert.NoError(t, a.Move(1, 0))
	assert.Equal(t, 2, a.GetMoveCount())
	assert.NoError(t, a.UndoN(2))
	assert.Equal(t, 4, a.GetPileCount())
	// 0 回 undo は no-op
	assert.NoError(t, a.UndoN(0))
	// 過剰 undo はエラー
	assert.Error(t, a.UndoN(1))
}

func TestAccordion_Undo_NotPlaying(t *testing.T) {
	a := setupPlayingAccordion()
	a.SetPhase(domain.AccordionPhaseGameOver)
	assert.Error(t, a.Undo())
	assert.False(t, a.CanUndo())
}

func TestAccordion_UndoToEscape(t *testing.T) {
	a := setupPlayingAccordion()
	setAccordionPiles(a,
		[2]int{domain.CardDesignSpade, 5},
		[2]int{domain.CardDesignSpade, 10},
	)
	// 勝利手を実行 → GameClear になり stalemate=false
	assert.NoError(t, a.Move(1, 0))
	assert.Equal(t, 0, a.UndoToEscape())
	// stalemate ならヒストリを遡る
	a.SetIsStalemate(true)
	// 1つだけ履歴があり、そのスナップショットは stalemate=false だったため
	assert.Equal(t, 1, a.UndoToEscape())
	// 履歴がない場合は -1
	a2 := setupPlayingAccordion()
	a2.SetIsStalemate(true)
	assert.Equal(t, -1, a2.UndoToEscape())
}

func TestAccordion_ActionLog(t *testing.T) {
	a := setupPlayingAccordion()
	setAccordionPiles(a,
		[2]int{domain.CardDesignSpade, 5},
		[2]int{domain.CardDesignSpade, 10},
		[2]int{domain.CardDesignDiamond, 3},
	)
	assert.Empty(t, a.GetActionLog())
	assert.NoError(t, a.Move(1, 0))
	entries := a.GetActionLog()
	assert.Len(t, entries, 1)
	assert.Equal(t, "move", entries[0].ActionType)
	a.GiveUp()
	entries = a.GetActionLog()
	assert.Len(t, entries, 2)
	assert.Equal(t, "giveup", entries[1].ActionType)
}

func TestAccordion_JSONRoundTrip(t *testing.T) {
	a := setupPlayingAccordion()
	setAccordionPiles(a,
		[2]int{domain.CardDesignSpade, 5},
		[2]int{domain.CardDesignSpade, 10},
		[2]int{domain.CardDesignDiamond, 3},
	)
	assert.NoError(t, a.Move(1, 0))
	data, err := json.Marshal(a)
	assert.NoError(t, err)

	b := domain.NewDefaultAccordion()
	assert.NoError(t, json.Unmarshal(data, b))
	assert.Equal(t, a.GetPileCount(), b.GetPileCount())
	assert.Equal(t, a.GetMoveCount(), b.GetMoveCount())
	assert.Equal(t, a.GetPhase(), b.GetPhase())
}

func TestAccordion_UnmarshalJSON_SizeLimit(t *testing.T) {
	bigLog := `{"al": [`
	for i := 0; i < 1001; i++ {
		if i > 0 {
			bigLog += ","
		}
		bigLog += `{"t":0,"p":0,"a":"x","d":"y","c":[]}`
	}
	bigLog += `]}`
	b := domain.NewDefaultAccordion()
	err := json.Unmarshal([]byte(bigLog), b)
	assert.Error(t, err)
}

func TestNewDefaultAccordion(t *testing.T) {
	a := domain.NewDefaultAccordion()
	assert.NotNil(t, a)
}

// **Web は1クリックで最後まで自動化できるのに、ネイティブ CUI には
// オートコンプリートが存在せず、同じ手を延々と打たされていた (#4793)。**
func TestAccordion_AutoComplete(t *testing.T) {
	card := func(design, value int) *domain.Card { return domain.NewCard(design, value, false) }
	board := func(cards ...*domain.Card) *domain.Accordion {
		a := newTestAccordion()
		piles := make([][]*domain.Card, len(cards))
		for i, c := range cards {
			piles[i] = []*domain.Card{c}
		}
		a.SetPiles(piles)
		return a
	}

	t.Run("merges until no move is left", func(t *testing.T) {
		// 同スート4枚。offset=3 と offset=1 の手が続き、最後は1山になる。
		a := board(
			card(domain.CardDesignSpade, 1), card(domain.CardDesignSpade, 2),
			card(domain.CardDesignSpade, 3), card(domain.CardDesignSpade, 4),
		)
		require.NoError(t, a.AutoComplete())
		assert.Equal(t, 1, len(a.GetPiles()), "全部まとまるはず")
		assert.Nil(t, a.GetHint(), "打てる手が残っていない")
	})

	// **1手も動かせなければエラー。**押しても何も起きないより、押せない理由が
	// 返るほうがよい。
	t.Run("reports an error when nothing can be merged", func(t *testing.T) {
		a := board(
			card(domain.CardDesignSpade, 1), card(domain.CardDesignHeart, 5),
			card(domain.CardDesignClover, 9), card(domain.CardDesignDiamond, 12),
		)
		require.Nil(t, a.GetHint(), "前提: 打てる手が無い")
		assert.Error(t, a.AutoComplete())
	})

	// **理由まで正しく返す。**GetHint もフェーズを見るのでエラーにはなるが、
	// 「打てる手が無い」ではなく「フェーズが違う」と伝わるほうが親切。
	t.Run("rejected outside the playing phase, with that reason", func(t *testing.T) {
		a := board(card(domain.CardDesignSpade, 1), card(domain.CardDesignSpade, 2))
		a.GiveUp()
		err := a.AutoComplete()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "playing phase")
	})

	// **打つ手はヒントと同じ規則で選ぶ。**別実装だと、ヒントが勧める手と自動で
	// 打たれる手が食い違う。
	t.Run("takes the move the hint points at", func(t *testing.T) {
		a := board(
			card(domain.CardDesignSpade, 1), card(domain.CardDesignHeart, 5),
			card(domain.CardDesignClover, 9), card(domain.CardDesignDiamond, 1),
		)
		hint := a.GetHint()
		require.NotNil(t, hint)
		before := len(a.GetPiles())
		require.NoError(t, a.AutoComplete())
		assert.Less(t, len(a.GetPiles()), before)
	})
}
