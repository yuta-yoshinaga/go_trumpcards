//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSirTommy() *SirTommy {
	s := NewSirTommy(NewTrumpCards(0))
	s.Reset()
	return s
}

// stackDeck replaces the stock with an exact sequence so a test can drive the
// game deterministically. The last element is the top of the stock (drawn first),
// matching how Reset builds it.
func stackDeck(s *SirTommy, cards []*Card) {
	s.stock = cards
	for i := range SirTommyFoundationCnt {
		s.foundations[i] = nil
	}
	for i := range SirTommyWasteCnt {
		s.wastes[i] = nil
	}
	s.phase = SirTommyPhasePlaying
	s.isStalemate = false
}

func TestSirTommy_Reset_InitialState(t *testing.T) {
	s := newTestSirTommy()

	assert.Equal(t, SirTommyPhasePlaying, s.GetPhase())
	assert.Equal(t, 0, s.GetMoveCount())
	assert.Empty(t, s.GetActionLog())
	assert.False(t, s.IsStalemate())
	assert.False(t, s.GetGameEndFlag())

	// Unlike Calculation, no foundation is seeded: every one starts empty and is
	// opened by an Ace drawn from the stock.
	for i, f := range s.GetFoundations() {
		assert.Empty(t, f, "foundation %d should start empty", i)
	}
	for i, w := range s.GetWastes() {
		assert.Empty(t, w, "waste %d should start empty", i)
	}
	assert.Equal(t, 52, s.GetStockCount(), "all 52 cards start in the stock")
	assert.NotNil(t, s.GetStockTop())
}

func TestSirTommy_AceOpensFoundation(t *testing.T) {
	s := newTestSirTommy()
	stackDeck(s, []*Card{NewCard(0, 1, true)}) // ace on top

	require.NoError(t, s.PlayStockToFoundation(0))
	require.Len(t, s.GetFoundations()[0], 1)
	assert.Equal(t, 1, s.GetFoundations()[0][0].GetValue())
	assert.Equal(t, 1, s.GetMoveCount())
	assert.Equal(t, 0, s.GetStockCount())
}

func TestSirTommy_NonAceCannotOpenFoundation(t *testing.T) {
	s := newTestSirTommy()
	stackDeck(s, []*Card{NewCard(0, 5, true)})

	require.Error(t, s.PlayStockToFoundation(0), "only an Ace may open an empty foundation")
	assert.Equal(t, 1, s.GetStockCount(), "a rejected move must not consume the card")
	assert.Equal(t, 0, s.GetMoveCount())
}

func TestSirTommy_FoundationBuildsUpIgnoringSuit(t *testing.T) {
	s := newTestSirTommy()
	// drawn in reverse order: ace first, then the 2 of a DIFFERENT suit
	stackDeck(s, []*Card{NewCard(3, 2, true), NewCard(0, 1, true)})

	require.NoError(t, s.PlayStockToFoundation(0))
	require.NoError(t, s.PlayStockToFoundation(0), "suit is ignored when building up")
	require.Len(t, s.GetFoundations()[0], 2)
	assert.Equal(t, 2, s.GetFoundations()[0][1].GetValue())
}

func TestSirTommy_FoundationRejectsWrongRank(t *testing.T) {
	s := newTestSirTommy()
	stackDeck(s, []*Card{NewCard(0, 4, true), NewCard(0, 1, true)})

	require.NoError(t, s.PlayStockToFoundation(0))
	require.Error(t, s.PlayStockToFoundation(0), "must be exactly one higher")
}

func TestSirTommy_FoundationDoesNotWrapPastKing(t *testing.T) {
	s := newTestSirTommy()
	s.foundations[0] = []*Card{NewCard(0, CardValueMax, true)}
	s.stock = []*Card{NewCard(1, 1, true)}

	require.Error(t, s.PlayStockToFoundation(0),
		"a completed foundation must not accept an Ace by wrapping around")
}

func TestSirTommy_StockToWasteAndWasteTopToFoundation(t *testing.T) {
	s := newTestSirTommy()
	stackDeck(s, []*Card{NewCard(1, 2, true), NewCard(0, 1, true)})

	require.NoError(t, s.PlayStockToFoundation(0)) // ace opens
	require.NoError(t, s.PlayStockToWaste(2))
	require.Len(t, s.GetWastes()[2], 1)

	require.NoError(t, s.PlayWasteToFoundation(2, 0))
	assert.Empty(t, s.GetWastes()[2])
	assert.Len(t, s.GetFoundations()[0], 2)
}

func TestSirTommy_OnlyWasteTopIsMovable(t *testing.T) {
	s := newTestSirTommy()
	s.foundations[0] = []*Card{NewCard(0, 1, true)}
	// buried 2 under a King: the 2 would fit the foundation but is not on top
	s.wastes[0] = []*Card{NewCard(1, 2, true), NewCard(2, CardValueMax, true)}

	require.Error(t, s.PlayWasteToFoundation(0, 0), "only the top card of a waste may move")
	assert.Len(t, s.GetWastes()[0], 2)
}

func TestSirTommy_InvalidIndexesAndEmptyPiles(t *testing.T) {
	s := newTestSirTommy()

	assert.Error(t, s.PlayStockToFoundation(-1))
	assert.Error(t, s.PlayStockToFoundation(SirTommyFoundationCnt))
	assert.Error(t, s.PlayStockToWaste(-1))
	assert.Error(t, s.PlayStockToWaste(SirTommyWasteCnt))
	assert.Error(t, s.PlayWasteToFoundation(-1, 0))
	assert.Error(t, s.PlayWasteToFoundation(0, -1))
	assert.Error(t, s.PlayWasteToFoundation(0, 0), "empty waste")

	stackDeck(s, nil)
	assert.Error(t, s.PlayStockToWaste(0), "empty stock")
	assert.Error(t, s.PlayStockToFoundation(0), "empty stock")
}

func TestSirTommy_GameClearWhenAllFiftyTwoAreUp(t *testing.T) {
	s := newTestSirTommy()
	for i := range SirTommyFoundationCnt {
		pile := make([]*Card, 0, CardValueMax)
		for v := 1; v <= CardValueMax; v++ {
			pile = append(pile, NewCard(i, v, true))
		}
		s.foundations[i] = pile
	}
	s.checkGameClear()

	assert.Equal(t, SirTommyPhaseGameClear, s.GetPhase())
	assert.True(t, s.GetGameEndFlag())
}

func TestSirTommy_StalemateOnlyWhenStockEmptyAndNothingPlayable(t *testing.T) {
	s := newTestSirTommy()

	// stock still has cards -> never a stalemate, they can always go to a waste
	stackDeck(s, []*Card{NewCard(0, 9, true)})
	s.checkStalemate()
	assert.False(t, s.IsStalemate())

	// stock empty and no waste top fits any foundation
	stackDeck(s, nil)
	s.foundations[0] = []*Card{NewCard(0, 1, true)}
	s.wastes[0] = []*Card{NewCard(1, 9, true)}
	s.checkStalemate()
	assert.True(t, s.IsStalemate())

	// same, but now a waste top does fit
	s.wastes[1] = []*Card{NewCard(2, 2, true)}
	s.checkStalemate()
	assert.False(t, s.IsStalemate())
}

func TestSirTommy_Hint(t *testing.T) {
	s := newTestSirTommy()

	// Deal a stock whose top is not an Ace, with every foundation empty, so
	// nothing can go to a foundation. Asserting this on a freshly shuffled deck
	// would be a 4-in-52 flake: a real Reset leaves whatever the shuffle put on
	// top, and an Ace there is a legitimate hint.
	//
	// #5552 以降、この局面は **nil ではなく置き場所の助言**を返す。ウェイストに
	// どう置くかがこのゲーム唯一の戦略的判断で、そこを黙るヒントは最頻出の
	// 局面で役に立っていなかった。
	stackDeck(s, []*Card{NewCard(0, 9, true)})
	placement := s.GetHint()
	require.NotNil(t, placement, "置き場所の助言を返す")
	assert.Equal(t, "waste", placement.ToZone)
	assert.GreaterOrEqual(t, placement.WasteIdx, 0)
	assert.Less(t, placement.WasteIdx, SirTommyWasteCnt)

	// stock top is an Ace -> hint points at the stock
	stackDeck(s, []*Card{NewCard(0, 1, true)})
	h := s.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "stock", h.FromZone)
	assert.Equal(t, "foundation", h.ToZone)
	assert.Equal(t, -1, h.WasteIdx)

	// waste top fits -> hint points at that waste
	stackDeck(s, nil)
	s.foundations[0] = []*Card{NewCard(0, 1, true)}
	s.wastes[3] = []*Card{NewCard(1, 2, true)}
	h = s.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "waste", h.FromZone)
	assert.Equal(t, 3, h.WasteIdx)
	assert.Equal(t, 0, h.FoundationIdx)

	s.GiveUp()
	assert.Nil(t, s.GetHint(), "no hint once the game is over")
}

func TestSirTommy_GiveUp(t *testing.T) {
	s := newTestSirTommy()
	s.GiveUp()
	assert.Equal(t, SirTommyPhaseGameOver, s.GetPhase())
	assert.True(t, s.GetGameEndFlag())
	require.NotEmpty(t, s.GetActionLog())

	before := len(s.GetActionLog())
	s.GiveUp()
	assert.Len(t, s.GetActionLog(), before, "giving up twice logs once")
}

func TestSirTommy_UndoRestoresPreviousState(t *testing.T) {
	s := newTestSirTommy()
	stackDeck(s, []*Card{NewCard(1, 2, true), NewCard(0, 1, true)})

	assert.False(t, s.CanUndo())
	require.NoError(t, s.PlayStockToFoundation(0))
	assert.True(t, s.CanUndo())

	require.NoError(t, s.Undo())
	assert.Empty(t, s.GetFoundations()[0])
	assert.Equal(t, 2, s.GetStockCount())
	assert.Equal(t, 0, s.GetMoveCount())
	assert.Error(t, s.Undo(), "nothing left to undo")
}

func TestSirTommy_AutoCompleteFinishesAReachableGame(t *testing.T) {
	s := newTestSirTommy()
	stackDeck(s, nil)
	// four foundations open on Aces, every remaining card sitting on a waste top
	for i := range SirTommyFoundationCnt {
		s.foundations[i] = []*Card{NewCard(i, 1, true)}
		pile := make([]*Card, 0, CardValueMax-1)
		for v := CardValueMax; v >= 2; v-- {
			pile = append(pile, NewCard(i, v, true))
		}
		s.wastes[i] = pile
	}

	require.NoError(t, s.AutoComplete())
	assert.Equal(t, SirTommyPhaseGameClear, s.GetPhase())
}

func TestSirTommy_JSONRoundTrip(t *testing.T) {
	s := newTestSirTommy()
	require.NoError(t, s.PlayStockToWaste(0))

	data, err := json.Marshal(s)
	require.NoError(t, err)

	restored := NewSirTommy(NewTrumpCards(0))
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, s.GetPhase(), restored.GetPhase())
	assert.Equal(t, s.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, s.GetStockCount(), restored.GetStockCount())
	assert.Len(t, restored.GetWastes()[0], 1)
}

func TestSirTommy_UnmarshalRejectsOutOfRangeState(t *testing.T) {
	s := NewSirTommy(NewTrumpCards(0))
	assert.Error(t, json.Unmarshal([]byte(`{"ps":99}`), s), "phase out of range")
	assert.Error(t, json.Unmarshal([]byte(`{"mc":-1}`), s), "negative move count")
	assert.Error(t, json.Unmarshal([]byte(`not json`), s))
}

func TestSirTommy_NewDefaultUsesAStandardDeck(t *testing.T) {
	s := NewDefaultSirTommy()
	s.Reset()
	assert.Equal(t, 52, s.GetStockCount())
}

func TestSirTommy_UndoN(t *testing.T) {
	s := newTestSirTommy()
	stackDeck(s, []*Card{NewCard(0, 7, true), NewCard(1, 8, true), NewCard(2, 9, true)})

	assert.Error(t, s.UndoN(0), "n must be positive")
	assert.Error(t, s.UndoN(1), "no history yet")

	require.NoError(t, s.PlayStockToWaste(0))
	require.NoError(t, s.PlayStockToWaste(1))
	require.NoError(t, s.PlayStockToWaste(2))
	assert.Error(t, s.UndoN(4), "more than the history holds")

	require.NoError(t, s.UndoN(2))
	assert.Equal(t, 1, s.GetMoveCount())
	assert.Equal(t, 2, s.GetStockCount())
}

func TestSirTommy_UndoToEscapeAndAllFaceUp(t *testing.T) {
	s := newTestSirTommy()
	assert.False(t, s.AllFaceUp(), "the stock still holds cards after Reset")
	assert.Equal(t, 0, s.UndoToEscape(), "not stalemated -> nothing to undo")

	// Two cards, neither playable: after both are parked the stock empties and
	// nothing fits any foundation, which is the stalemate.
	stackDeck(s, []*Card{NewCard(0, 9, true), NewCard(1, 7, true)})
	require.NoError(t, s.PlayStockToWaste(0))
	require.NoError(t, s.PlayStockToWaste(1))
	assert.True(t, s.AllFaceUp())
	require.True(t, s.IsStalemate())
	assert.Equal(t, 1, s.UndoToEscape(),
		"one undo puts a card back in the stock, which is by definition not stalemate")

	// Stalemate with no history at all cannot be escaped.
	fresh := newTestSirTommy()
	stackDeck(fresh, nil)
	fresh.checkStalemate()
	require.True(t, fresh.IsStalemate())
	assert.Equal(t, -1, fresh.UndoToEscape())
}

func TestSirTommy_AutoCompleteRefusesWhileStockRemains(t *testing.T) {
	s := newTestSirTommy()
	// A playable Ace on top, but the stock is not exhausted: auto-complete is an
	// endgame convenience, not an autoplayer, and the frontend button agrees.
	stackDeck(s, []*Card{NewCard(0, 9, true), NewCard(0, 1, true)})

	require.Error(t, s.AutoComplete(), "stock is not empty")
	assert.Equal(t, 2, s.GetStockCount(), "a refused auto-complete changes nothing")
	assert.Equal(t, 0, s.GetMoveCount())
}

// #5552: ドキュメントが「置き場所を選ぶ判断がゲーム性の中心」と書いている当の
// 局面 — ストックの札をどのウェイストに置くか — でヒントが常に nil だった。
func TestSirTommy_GetHint_WastePlacement(t *testing.T) {
	newGame := func(stockTop int, wastes [][]int) *SirTommy {
		s := newTestSirTommy()
		s.Reset()
		s.stock = []*Card{NewCard(CardDesignSpade, stockTop, false)}
		for i := range SirTommyWasteCnt {
			pile := make([]*Card, 0, len(wastes[i]))
			for _, v := range wastes[i] {
				pile = append(pile, NewCard(CardDesignHeart, v, false))
			}
			s.wastes[i] = pile
		}
		for i := range SirTommyFoundationCnt {
			s.foundations[i] = nil
		}
		return s
	}

	t.Run("keeps a waste descending by choosing the tightest higher top", func(t *testing.T) {
		// 7 を置く。上が 13 / 9 / 3 / 空 → 9 が最も近い上位。
		s := newGame(7, [][]int{{13}, {9}, {3}, {}})
		h := s.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "stock", h.FromZone)
		assert.Equal(t, "waste", h.ToZone)
		assert.Equal(t, 1, h.WasteIdx)
		assert.Equal(t, -1, h.FoundationIdx)
	})

	t.Run("uses an empty pile when nothing higher is on top", func(t *testing.T) {
		s := newGame(9, [][]int{{5}, {3}, {}, {2}})
		h := s.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, 2, h.WasteIdx)
	})

	// **どれかを埋めるしかないなら、いちばん高い札の上に置く。**低い札を
	// 埋めると、それが必要になったときに掘り出せない。
	t.Run("buries the highest top when every pile is occupied and lower", func(t *testing.T) {
		s := newGame(9, [][]int{{5}, {8}, {2}, {6}})
		h := s.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, 1, h.WasteIdx)
	})

	// **ファンデーションに置ける手が優先。**置き場所の助言はその次。
	t.Run("prefers a foundation move when one exists", func(t *testing.T) {
		s := newGame(1, [][]int{{5}, {8}, {2}, {6}}) // A はファンデーションへ
		h := s.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "foundation", h.ToZone)
		assert.Equal(t, 0, h.FoundationIdx)
	})

	t.Run("no hint once the stock is empty and nothing plays", func(t *testing.T) {
		s := newGame(9, [][]int{{5}, {8}, {2}, {6}})
		s.stock = nil
		assert.Nil(t, s.GetHint())
	})
}
