//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestBigBen() *BigBen {
	gc := NewDefaultBigBen()
	gc.Reset()
	return gc
}

// setBigBenBoard installs an exact position. Tests must never lean on the shuffle:
// asserting what is or is not playable against a real deal is a flake waiting
// to happen (see #4467).
func setBigBenBoard(gc *BigBen, foundation [BigBenFoundationCnt][]*Card, cols [][]*Card) {
	gc.foundation = foundation
	for i := range BigBenTableauCnt {
		gc.tableau[i] = nil
	}
	for i, c := range cols {
		pile := make([]*BigBenTableauCard, 0, len(c))
		for _, card := range c {
			pile = append(pile, &BigBenTableauCard{Card: card, FaceUp: true})
		}
		gc.tableau[i] = pile
	}
	gc.phase = BigBenPhasePlaying
	gc.isStalemate = false
	gc.history = nil
	gc.moveCount = 0
	gc.actionLog = nil
}

// bigBenStarterBoard seeds the twelve clock faces exactly as a real deal would.
func bigBenStarterBoard() [BigBenFoundationCnt][]*Card {
	var f [BigBenFoundationCnt][]*Card
	for i, s := range bigBenStarters {
		f[i] = []*Card{NewCard(s.design, s.value, true)}
	}
	return f
}

func TestBigBen_Reset_SeedsTheClockAndDealsEightByFive(t *testing.T) {
	gc := newTestBigBen()

	seeded := 0
	for i, pile := range gc.GetFoundation() {
		require.Len(t, pile, 1, "clock face %d holds exactly its starter", i)
		want := bigBenStarters[i]
		assert.Equal(t, want.design, pile[0].GetDesign(), "face %d suit", i)
		assert.Equal(t, want.value, pile[0].GetValue(), "face %d rank", i)
		seeded += len(pile)
	}
	assert.Equal(t, BigBenFoundationCnt, seeded)

	dealt := 0
	for col, pile := range gc.GetTableau() {
		assert.Len(t, pile, BigBenColumnLen, "column %d", col)
		for _, tc := range pile {
			assert.True(t, tc.FaceUp, "every card is face-up by rule")
		}
		dealt += len(pile)
	}
	assert.Equal(t, 40, dealt)
	// **配られるのは 52 枚だけ。**文字盤 12 + タブロー 40 で、2 組 104 枚の
	// ちょうど半分。残る 52 枚は山札になる（クローン元は 1 組なのでここで
	// 配り切っていた）。
	assert.Equal(t, CardCnt, seeded+dealt)
	assert.Equal(t, CardCnt, gc.GetStockCount(), "残り半分が山札")
	assert.True(t, gc.AllFaceUp())
	assert.False(t, gc.GetGameEndFlag())
}

// **文字盤は 9 時始まり。**クローン元は 1 時始まりで 1 時＝A、12 時＝Q だった。
// ここでは添字 0 が 9 時、添字 11 が 8 時。
func TestBigBen_TargetRankIsTheHour(t *testing.T) {
	assert.Equal(t, 9, BigBenTargetRank(0), "index 0 is 9 o'clock")
	assert.Equal(t, 12, BigBenTargetRank(3), "index 3 is 12 o'clock")
	assert.Equal(t, 1, BigBenTargetRank(4), "index 4 is 1 o'clock")
	assert.Equal(t, 8, BigBenTargetRank(11), "index 11 is 8 o'clock")
	// 範囲外は 0。これが 0 でないと、壊れた添字が黙って別の文字盤を指す。
	assert.Equal(t, 0, BigBenTargetRank(-1))
	assert.Equal(t, 0, BigBenTargetRank(BigBenFoundationCnt))
}

func TestBigBen_FoundationBuildsUpInSuit(t *testing.T) {
	gc := newTestBigBen()
	f := bigBenStarterBoard()
	// 添字 1 は 10 時で、♥3 から始まって ♥10 を目指す。次に要るのは ♥4。
	setBigBenBoard(gc, f, [][]*Card{
		{NewCard(CardDesignHeart, 4, true)},
		{NewCard(CardDesignSpade, 4, true)},
		{NewCard(CardDesignHeart, 6, true)},
	})

	assert.Error(t, gc.MoveTableauToFoundation(1, 1), "wrong suit")
	assert.Error(t, gc.MoveTableauToFoundation(2, 1), "not the next rank")
	require.NoError(t, gc.MoveTableauToFoundation(0, 1))
	assert.Len(t, gc.GetFoundation()[1], 2)
}

// K -> A is the wraparound that lets a face started on a high card reach a low
// target; without it the 1..4 o'clock faces could never be finished.
func TestBigBen_FoundationWrapsFromKingToAce(t *testing.T) {
	gc := newTestBigBen()
	f := bigBenStarterBoard()
	// 添字 4 は 1 時で、♣6 から始まって A を目指す ── K を跨がないと届かない。
	// K の状態にしておき、次の A で完成することを見る。
	f[4] = []*Card{NewCard(CardDesignClover, CardValueMax, true)}
	setBigBenBoard(gc, f, [][]*Card{{NewCard(CardDesignClover, 1, true)}})

	require.NoError(t, gc.MoveTableauToFoundation(0, 4))
	assert.True(t, gc.IsFoundationComplete(4))
	assert.Equal(t, 1, bigBenNextRank(CardValueMax))
	assert.Equal(t, 5, bigBenNextRank(4))
}

// A completed face must stop accepting cards, or the wraparound would let it
// lap the whole suit.
func TestBigBen_CompletedFaceRejectsMoreCards(t *testing.T) {
	gc := newTestBigBen()
	f := bigBenStarterBoard()
	// 添字 1 は 10 時。目標の ♥10 に達している状態にする。
	f[1] = []*Card{NewCard(CardDesignHeart, 10, true)}
	setBigBenBoard(gc, f, [][]*Card{{NewCard(CardDesignHeart, 11, true)}})

	require.True(t, gc.IsFoundationComplete(1))
	assert.Error(t, gc.MoveTableauToFoundation(0, 1), "♥J would lap a finished face")
}

func TestBigBen_RejectsInvalidFoundationArguments(t *testing.T) {
	gc := newTestBigBen()
	setBigBenBoard(gc, bigBenStarterBoard(), [][]*Card{{NewCard(CardDesignHeart, 3, true)}})

	assert.Error(t, gc.MoveTableauToFoundation(0, -1), "index out of range")
	assert.Error(t, gc.MoveTableauToFoundation(0, BigBenFoundationCnt), "index out of range")
	assert.Error(t, gc.MoveTableauToFoundation(-1, 4), "column out of range")
	assert.Error(t, gc.MoveTableauToFoundation(1, 4), "column is empty")

	// An empty clock face can only happen in a restored snapshot; it must not
	// silently accept a card.
	var empty [BigBenFoundationCnt][]*Card
	setBigBenBoard(gc, empty, [][]*Card{{NewCard(CardDesignHeart, 3, true)}})
	assert.Error(t, gc.MoveTableauToFoundation(0, 4))
	assert.False(t, gc.IsFoundationComplete(4))
	assert.False(t, gc.IsFoundationComplete(-1), "out of range is never complete")
}

// **同スート降順。**元のサブテストは "suit is ignored" を主張していたが、
// それはクローン元グランドファーザーズ・クロックの規則。
func TestBigBen_TableauBuildsDownInSuitOnly(t *testing.T) {
	gc := newTestBigBen()
	setBigBenBoard(gc, bigBenStarterBoard(), [][]*Card{
		{NewCard(CardDesignSpade, 9, true)},
		{NewCard(CardDesignHeart, 8, true)},
		{NewCard(CardDesignSpade, 8, true)},
	})

	assert.Error(t, gc.MoveTableauToTableau(1, 0), "♥8 onto ♠9 — suit matters here")
	require.NoError(t, gc.MoveTableauToTableau(2, 0), "♠8 onto ♠9")
	assert.Len(t, gc.GetTableau()[0], 2)
	assert.Empty(t, gc.GetTableau()[2])
}

// **空き列は埋められない。**元のサブテストは "accepts any card" を主張していた。
func TestBigBen_EmptyColumnAcceptsNothing(t *testing.T) {
	gc := newTestBigBen()
	setBigBenBoard(gc, bigBenStarterBoard(), [][]*Card{
		{NewCard(CardDesignSpade, 9, true), NewCard(CardDesignDiamond, 2, true)},
	})

	assert.Error(t, gc.MoveTableauToTableau(0, 5))
	assert.Empty(t, gc.GetTableau()[5])
	assert.Len(t, gc.GetTableau()[0], 2, "拒まれた札は元の列に残る")
}

func TestBigBen_RejectsInvalidTableauArguments(t *testing.T) {
	gc := newTestBigBen()
	setBigBenBoard(gc, bigBenStarterBoard(), [][]*Card{{NewCard(CardDesignSpade, 9, true)}})

	assert.Error(t, gc.MoveTableauToTableau(-1, 1), "from out of range")
	assert.Error(t, gc.MoveTableauToTableau(0, BigBenTableauCnt), "to out of range")
	assert.Error(t, gc.MoveTableauToTableau(0, 0), "same column")
	assert.Error(t, gc.MoveTableauToTableau(1, 0), "column is empty")
}

func TestBigBen_GameClearWhenEveryFaceReachesItsHour(t *testing.T) {
	gc := newTestBigBen()
	var f [BigBenFoundationCnt][]*Card
	for i, s := range bigBenStarters {
		f[i] = []*Card{NewCard(s.design, BigBenTargetRank(i), true)}
	}
	// 添字 11 は 8 時で ♦8 を目指す。1 歩手前（♦7）に戻しておく。
	last := bigBenStarters[11]
	f[11] = []*Card{NewCard(last.design, 7, true)}
	setBigBenBoard(gc, f, [][]*Card{{NewCard(last.design, 8, true)}})

	require.NoError(t, gc.MoveTableauToFoundation(0, 11))
	assert.Equal(t, BigBenPhaseGameClear, gc.GetPhase())
	assert.True(t, gc.GetGameEndFlag())
}

func TestBigBen_GiveUpEndsTheGameOnce(t *testing.T) {
	gc := newTestBigBen()
	gc.GiveUp()
	assert.Equal(t, BigBenPhaseGameOver, gc.GetPhase())
	logLen := len(gc.GetActionLog())

	gc.GiveUp()
	assert.Len(t, gc.GetActionLog(), logLen, "a second give-up is a no-op")

	assert.Error(t, gc.MoveTableauToFoundation(0, 0))
	assert.Error(t, gc.MoveTableauToTableau(0, 1))
	assert.Error(t, gc.AutoComplete())
	assert.Nil(t, gc.GetHint())
}

func TestBigBen_HintPrefersTheClockFace(t *testing.T) {
	gc := newTestBigBen()
	// **文字盤 12 枚がそのまま「今すぐ置ける札」の一覧になる。**次に要るのは
	// ♣3 ♥4 ♠5 ♦6 ♣7 ♥8 ♠9 ♦10 ♣J ♥Q ♠K ♦A の 12 枚だけ。それ以外の札を
	// 置かないと、意図しない文字盤への手が先に見つかる。
	setBigBenBoard(gc, bigBenStarterBoard(), [][]*Card{
		{NewCard(CardDesignSpade, 2, true)},
		{NewCard(CardDesignSpade, 1, true)}, // a legal tableau move onto ♠2
		{NewCard(CardDesignHeart, 4, true)}, // but ♥4 goes onto face 1 (♥3)
	})

	h := gc.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "foundation", h.ToZone)
	assert.Equal(t, 2, h.FromCol)
	assert.Equal(t, 1, h.ToIdx)
}

func TestBigBen_HintFallsBackToTheTableau(t *testing.T) {
	gc := newTestBigBen()
	// どちらも文字盤へは行けない札。♠A は ♠2 の上に置ける。
	setBigBenBoard(gc, bigBenStarterBoard(), [][]*Card{
		{NewCard(CardDesignSpade, 2, true)},
		{NewCard(CardDesignSpade, 1, true)},
	})
	gc.stock = nil

	h := gc.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "tableau", h.ToZone)
	assert.Equal(t, 1, h.FromCol)
	assert.Equal(t, 0, h.ToIdx)
	// The hint must be a move the domain actually accepts.
	require.NoError(t, gc.MoveTableauToTableau(h.FromCol, h.ToIdx))
}

// **山札があるうちは手詰まりではない。**補充がこのゲームの逃げ道なので、
// ヒントがそれを言わないと「配れば動くのに詰み」と表示される。
// (クローン元は山札を持たないので、この分岐そのものが無かった。)
func TestBigBen_HintSuggestsDealingWhenNothingElseMoves(t *testing.T) {
	gc := newTestBigBen()
	setBigBenBoard(gc, bigBenStarterBoard(), [][]*Card{
		{NewCard(CardDesignSpade, 2, true)},
	})
	gc.stock = []*Card{NewCard(CardDesignClover, 4, true)}

	h := gc.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "stock", h.FromZone)
	assert.Equal(t, "stock", h.ToZone, "補充は移動ではないので列を指さない")
	assert.Equal(t, -1, h.ToIdx)
	gc.checkStalemate()
	assert.False(t, gc.IsStalemate(), "配れるうちは詰みではない")

	// 負のコントロール: 山札が尽きて初めて手詰まり。
	gc.stock = nil
	assert.Nil(t, gc.GetHint())
	gc.checkStalemate()
	assert.True(t, gc.IsStalemate())
	assert.Equal(t, -1, gc.UndoToEscape(), "no history to rewind into")
}

func TestBigBen_AutoCompleteOnlyFeedsTheClock(t *testing.T) {
	gc := newTestBigBen()
	// 添字 1 (10 時) は ♥3 から始まる。次は ♥4、その次は ♥5。
	setBigBenBoard(gc, bigBenStarterBoard(), [][]*Card{
		{NewCard(CardDesignHeart, 5, true), NewCard(CardDesignHeart, 4, true)},
	})

	require.NoError(t, gc.AutoComplete())
	assert.Len(t, gc.GetFoundation()[1], 3, "♥3 then ♥4 then ♥5")
	assert.Empty(t, gc.GetTableau()[0])
}

// AutoComplete must never make a tableau move, or it would shuffle the board
// instead of clearing it — it drives off foundationHint, not GetHint.
func TestBigBen_AutoCompleteIgnoresTableauMoves(t *testing.T) {
	gc := newTestBigBen()
	setBigBenBoard(gc, bigBenStarterBoard(), [][]*Card{
		{NewCard(CardDesignSpade, 2, true)},
		{NewCard(CardDesignSpade, 1, true)},
	})

	gc.stock = nil
	require.NotNil(t, gc.GetHint(), "a tableau move exists")
	assert.Error(t, gc.AutoComplete(), "but nothing reaches a clock face")
	assert.Equal(t, 0, gc.GetMoveCount())
	assert.Len(t, gc.GetTableau()[0], 1, "the board is untouched")
	assert.Len(t, gc.GetTableau()[1], 1)
}

func TestBigBen_UndoRestoresBothZones(t *testing.T) {
	gc := newTestBigBen()
	// 列1 は同スート降順に積める形（♠3 の上に ♠2）。空き列には置けないので、
	// 移動元の列が空にならない盤を選ぶ。
	setBigBenBoard(gc, bigBenStarterBoard(), [][]*Card{
		{NewCard(CardDesignSpade, 3, true)},
		{NewCard(CardDesignHeart, 4, true)},
		{NewCard(CardDesignSpade, 2, true)},
	})

	assert.False(t, gc.CanUndo())
	assert.Error(t, gc.Undo(), "nothing to undo")
	assert.Error(t, gc.UndoN(0), "n must be positive")
	assert.Error(t, gc.UndoN(1), "no history yet")

	require.NoError(t, gc.MoveTableauToFoundation(1, 1))
	require.NoError(t, gc.MoveTableauToTableau(2, 0))
	assert.True(t, gc.CanUndo())
	assert.Equal(t, 2, gc.GetMoveCount())

	assert.Error(t, gc.UndoN(5), "more than the history holds")
	require.NoError(t, gc.UndoN(2))
	assert.Equal(t, 0, gc.GetMoveCount())
	assert.Len(t, gc.GetFoundation()[1], 1)
	assert.Len(t, gc.GetTableau()[0], 1)
	assert.Len(t, gc.GetTableau()[1], 1)
	assert.Len(t, gc.GetTableau()[2], 1)
}

func TestBigBen_UndoToEscapeCountsBackToAPlayablePosition(t *testing.T) {
	gc := newTestBigBen()

	// 行き止まりを作るには手数が要る。**♠2 を ♠3 に載せる 1 手だけ**が合法で、
	// それを打つと ♦2 が現れるが、そこからは何も動かない盤にする。
	//
	// 塞ぎ札は 2 条件で選ぶ: (1) どの文字盤の「次の 1 枚」でもないこと ──
	// 今欲しいのは ♣3 ♥4 ♠5 ♦6 ♣7 ♥8 ♠9 ♦10 ♣J ♥Q ♠K ♦A の 12 枚だけ、
	// (2) 同スートで 1 つ違いの組を作らないこと。クローン元はスート不問だった
	// ので、そちらの塞ぎ札をそのまま使うと合法手が残る。
	cols := [][]*Card{
		{NewCard(CardDesignSpade, 3, true)},
		{NewCard(CardDesignDiamond, 2, true), NewCard(CardDesignSpade, 2, true)},
		{NewCard(CardDesignHeart, 2, true)},
		{NewCard(CardDesignDiamond, 4, true)},
		{NewCard(CardDesignClover, 5, true)},
		{NewCard(CardDesignHeart, 11, true)},
		{NewCard(CardDesignDiamond, 12, true)},
		{NewCard(CardDesignClover, CardValueMax, true)},
	}
	require.Len(t, cols, BigBenTableauCnt, "no column may be left empty")
	setBigBenBoard(gc, bigBenStarterBoard(), cols)
	// 補充できるうちは詰みにならないので、山札も切る。
	gc.stock = nil

	assert.False(t, gc.IsStalemate())
	assert.Equal(t, 0, gc.UndoToEscape(), "not stalemated yet")

	require.NoError(t, gc.MoveTableauToTableau(1, 0), "♠2 onto ♠3 is the only move")
	assert.True(t, gc.IsStalemate(), "and it strands the board")
	assert.Equal(t, 1, gc.UndoToEscape())

	require.NoError(t, gc.UndoN(gc.UndoToEscape()))
	assert.False(t, gc.IsStalemate())
	assert.Len(t, gc.GetTableau()[1], 2, "♠2 is back on top of ♦2")
}

func TestBigBen_ActionLogUsesZeroBasedIndices(t *testing.T) {
	gc := newTestBigBen()
	setBigBenBoard(gc, bigBenStarterBoard(), [][]*Card{
		{NewCard(CardDesignSpade, 3, true)},
		{NewCard(CardDesignHeart, 4, true)},
		{NewCard(CardDesignSpade, 2, true)},
	})

	require.NoError(t, gc.MoveTableauToFoundation(1, 1))
	require.NoError(t, gc.MoveTableauToTableau(2, 0))

	log := gc.GetActionLog()
	require.Len(t, log, 2)
	assert.Equal(t, "タブロー列1→文字盤1", log[0].Detail)
	assert.Equal(t, "タブロー列2→タブロー列0", log[1].Detail)
}

func TestBigBen_JSONRoundTrip(t *testing.T) {
	gc := newTestBigBen()
	data, err := json.Marshal(gc)
	require.NoError(t, err)

	restored := NewBigBen(NewTrumpCards(0))
	require.NoError(t, json.Unmarshal(data, restored))
	assert.Equal(t, gc.GetPhase(), restored.GetPhase())
	assert.Equal(t, gc.GetMoveCount(), restored.GetMoveCount())
	// **山札も往復すること。**載せ忘れると Worker で補充が二度と効かない。
	assert.Equal(t, gc.GetStockCount(), restored.GetStockCount())
	assert.Len(t, restored.GetFoundation()[0], 1)
	assert.Len(t, restored.GetTableau()[0], BigBenColumnLen)
}

func TestBigBen_UnmarshalRejectsOutOfRangeState(t *testing.T) {
	gc := NewDefaultBigBen()
	assert.Error(t, json.Unmarshal([]byte(`nope`), gc))
	assert.Error(t, json.Unmarshal([]byte(`{"ps":99}`), gc))
	assert.Error(t, json.Unmarshal([]byte(`{"mc":-1}`), gc))

	card := `{"d":0,"v":1,"u":true}`
	oversizeFoundation := `[` + card
	for range CardValueMax {
		oversizeFoundation += `,` + card
	}
	oversizeFoundation += `]`
	assert.Error(t, json.Unmarshal([]byte(`{"fd":[`+oversizeFoundation+`]}`), gc), "clock face too large")

	tc := `{"c":` + card + `,"f":true}`
	oversizeTableau := `[` + tc
	for range CardCnt {
		oversizeTableau += `,` + tc
	}
	oversizeTableau += `]`
	assert.Error(t, json.Unmarshal([]byte(`{"tb":[`+oversizeTableau+`]}`), gc), "tableau too large")
}

func TestBigBen_NewDefaultUsesAStandardDeck(t *testing.T) {
	gc := NewDefaultBigBen()
	gc.Reset()
	total := 0
	for _, f := range gc.GetFoundation() {
		total += len(f)
	}
	for _, col := range gc.GetTableau() {
		total += len(col)
	}
	assert.Equal(t, CardCnt, total)
}

func TestBigBen_StarterIndexIgnoresNonStarters(t *testing.T) {
	assert.Equal(t, -1, bigBenStarterIndex(nil))
	assert.Equal(t, -1, bigBenStarterIndex(NewCard(CardDesignSpade, 2, true)),
		"♠2 is not one of the twelve; ♣2 is")
	assert.Equal(t, 0, bigBenStarterIndex(NewCard(CardDesignClover, 2, true)))
}
