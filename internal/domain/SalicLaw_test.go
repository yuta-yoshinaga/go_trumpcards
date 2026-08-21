//go:build test

package domain

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSalicLaw() *SalicLaw {
	c := NewDefaultSalicLaw()
	c.Reset()
	return c
}

// clearSalicLawBoard wipes the dealt layout so a test can state exactly the
// position it cares about. Never assert on a freshly Reset board -- the deal is
// shuffled, so any such assertion is a hidden flake.
func clearSalicLawBoard(c *SalicLaw) {
	c.stock = nil
	c.isStalemate = false
	c.openPiles = 0
	for i := range SalicLawFoundationCnt {
		c.foundation[i] = nil
	}
	for i := range SalicLawTableauCnt {
		c.tableau[i] = nil
	}
}

// openAllSalicLawPiles seats a king at the base of every column, which is the
// shape the board always has once the deal is done. Tests that skip this are
// describing a position the game cannot reach.
func openAllSalicLawPiles(c *SalicLaw) {
	for i := range SalicLawTableauCnt {
		c.tableau[i] = []*Card{NewCard(CardDesignSpade, CardValueMax, true)}
	}
	c.openPiles = SalicLawTableauCnt
}

// salicLawPush puts a card on top of an already-open column.
func salicLawPush(c *SalicLaw, pile, design, value int) *Card {
	card := NewCard(design, value, true)
	c.tableau[pile] = append(c.tableau[pile], card)
	return card
}

func TestNewSalicLaw(t *testing.T) {
	assert.NotNil(t, NewSalicLaw(NewTrumpCardsWithDecks(2, 0)))
	assert.NotNil(t, NewDefaultSalicLaw())
}

func TestSalicLaw_Reset(t *testing.T) {
	c := newTestSalicLaw()

	assert.Equal(t, SalicLawPhasePlaying, c.GetPhase())
	assert.Equal(t, 0, c.GetMoveCount())
	assert.False(t, c.CanUndo())
	// 場に出るのは 96 枚。1 枚だけ K が据わっているので山札は 95 枚。
	assert.Equal(t, SalicLawTotalCards-1, c.GetStockCount())
}

func TestSalicLaw_ResetIsRepeatable(t *testing.T) {
	c := newTestSalicLaw()
	for c.GetStockCount() > 0 {
		require.NoError(t, c.Draw())
	}
	c.Reset()

	assert.Equal(t, SalicLawTotalCards-1, c.GetStockCount())
	assert.Equal(t, 1, c.GetOpenPiles())
	assert.Len(t, c.GetQueens(), SalicLawQueenCnt)
	assert.Equal(t, 0, c.GetMoveCount())
}

// 配り切っても 1 枚も失われない。Q 8 枚 + 場の 96 枚 = 104 枚。
func TestSalicLaw_ResetKeepsEveryCard(t *testing.T) {
	c := newTestSalicLaw()
	for c.GetStockCount() > 0 {
		require.NoError(t, c.Draw())
	}

	total := len(c.GetQueens())
	for _, pile := range c.GetTableau() {
		total += len(pile)
	}
	assert.Equal(t, CardCnt*2, total)
}

func TestSalicLaw_Draw(t *testing.T) {
	t.Run("a non-king lands on the current column", func(t *testing.T) {
		c := newTestSalicLaw()
		clearSalicLawBoard(c)
		c.tableau[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, true)}
		c.openPiles = 1
		c.stock = []*Card{NewCard(CardDesignHeart, 4, true)}

		require.NoError(t, c.Draw())
		assert.Len(t, c.GetTableau()[0], 2)
		assert.Equal(t, 1, c.GetOpenPiles(), "列は増えない")
	})

	t.Run("a king opens the next column", func(t *testing.T) {
		c := newTestSalicLaw()
		clearSalicLawBoard(c)
		c.tableau[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, true)}
		c.openPiles = 1
		c.stock = []*Card{NewCard(CardDesignHeart, CardValueMax, true)}

		require.NoError(t, c.Draw())
		assert.Len(t, c.GetTableau()[0], 1, "今の列には積まれない")
		require.Len(t, c.GetTableau()[1], 1)
		assert.Equal(t, 2, c.GetOpenPiles())
	})

	t.Run("the ninth king would have nowhere to go", func(t *testing.T) {
		// 実際の配りでは K はちょうど 8 枚なので起きないが、KV から壊れた盤が
		// 来ることはある。9 枚目は最後の列に普通に積む（パニックしない）。
		c := newTestSalicLaw()
		clearSalicLawBoard(c)
		openAllSalicLawPiles(c)
		c.stock = []*Card{NewCard(CardDesignHeart, CardValueMax, true)}

		require.NoError(t, c.Draw())
		assert.Equal(t, SalicLawTableauCnt, c.GetOpenPiles())
		assert.Len(t, c.GetTableau()[SalicLawTableauCnt-1], 2)
	})

	t.Run("an empty stock is refused", func(t *testing.T) {
		c := newTestSalicLaw()
		clearSalicLawBoard(c)
		openAllSalicLawPiles(c)
		assert.Error(t, c.Draw())
	})
}

func TestSalicLaw_MoveTableauToFoundation(t *testing.T) {
	t.Run("an ace opens a foundation", func(t *testing.T) {
		c := newTestSalicLaw()
		clearSalicLawBoard(c)
		openAllSalicLawPiles(c)
		salicLawPush(c, 3, CardDesignHeart, 1)

		require.NoError(t, c.MoveTableauToFoundation(3))
		assert.Len(t, c.GetFoundation()[0], 1)
		assert.Len(t, c.GetTableau()[3], 1)
		assert.Equal(t, 1, c.GetMoveCount())
	})

	t.Run("the base king cannot be sent", func(t *testing.T) {
		c := newTestSalicLaw()
		clearSalicLawBoard(c)
		openAllSalicLawPiles(c)

		err := c.MoveTableauToFoundation(0)
		require.Error(t, err)
		code, _ := ErrorMessageCode(err)
		assert.Equal(t, "saliclaw.errKingIsTheBase", code)
	})

	t.Run("a card no foundation wants is refused", func(t *testing.T) {
		c := newTestSalicLaw()
		clearSalicLawBoard(c)
		openAllSalicLawPiles(c)
		salicLawPush(c, 0, CardDesignHeart, 9)

		assert.Error(t, c.MoveTableauToFoundation(0))
		assert.Len(t, c.GetTableau()[0], 2, "拒まれた札は残る")
	})
}

func TestSalicLaw_GetHint(t *testing.T) {
	t.Run("prefers a foundation move", func(t *testing.T) {
		c := newTestSalicLaw()
		clearSalicLawBoard(c)
		openAllSalicLawPiles(c)
		salicLawPush(c, 2, CardDesignHeart, 1)

		h := c.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "tableau", h.FromZone)
		assert.Equal(t, 2, h.FromIdx)
		assert.Equal(t, "foundation", h.ToZone)
	})

	t.Run("then a move onto a bare king", func(t *testing.T) {
		c := newTestSalicLaw()
		clearSalicLawBoard(c)
		openAllSalicLawPiles(c)
		salicLawPush(c, 2, CardDesignHeart, 9)

		h := c.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "tableau", h.FromZone)
		assert.Equal(t, 2, h.FromIdx)
		assert.Equal(t, "tableau", h.ToZone)
	})

	t.Run("then dealing, and it names no destination column", func(t *testing.T) {
		c := newTestSalicLaw()
		clearSalicLawBoard(c)
		// 1 列だけ開いていて、置き先になる「K だけの列」が他に無い盤。
		c.tableau[0] = []*Card{
			NewCard(CardDesignSpade, CardValueMax, true),
			NewCard(CardDesignHeart, 9, true),
		}
		c.openPiles = 1
		c.stock = []*Card{NewCard(CardDesignClover, 4, true)}

		h := c.GetHint()
		require.NotNil(t, h)
		assert.Equal(t, "stock", h.FromZone)
		// 捨て札は無いので、行き先も stock のまま。waste が漏れると CUI が
		// 存在しないゾーンを読み上げる。
		assert.Equal(t, "stock", h.ToZone)
		assert.Equal(t, -1, h.ToIdx)
	})

	t.Run("nothing left", func(t *testing.T) {
		c := newTestSalicLaw()
		clearSalicLawBoard(c)
		c.tableau[0] = []*Card{
			NewCard(CardDesignSpade, CardValueMax, true),
			NewCard(CardDesignHeart, 9, true),
		}
		c.openPiles = 1

		assert.Nil(t, c.GetHint())
	})

	t.Run("no hint once the game has ended", func(t *testing.T) {
		c := newTestSalicLaw()
		c.phase = SalicLawPhaseGameOver
		assert.Nil(t, c.GetHint())
		assert.Nil(t, c.foundationHint())
		assert.Nil(t, c.tableauHint())
	})
}

func TestSalicLaw_Stalemate(t *testing.T) {
	c := newTestSalicLaw()
	clearSalicLawBoard(c)
	c.tableau[0] = []*Card{
		NewCard(CardDesignSpade, CardValueMax, true),
		NewCard(CardDesignHeart, 9, true),
	}
	c.openPiles = 1
	c.checkStalemate()

	assert.True(t, c.IsStalemate())

	// 負のコントロール: 山札が 1 枚でもあれば配れるので手詰まりではない。
	c.stock = []*Card{NewCard(CardDesignClover, 4, true)}
	c.checkStalemate()
	assert.False(t, c.IsStalemate())
}

func TestSalicLaw_AutoComplete(t *testing.T) {
	t.Run("sends every card it can", func(t *testing.T) {
		c := newTestSalicLaw()
		clearSalicLawBoard(c)
		openAllSalicLawPiles(c)
		salicLawPush(c, 0, CardDesignHeart, 1)
		salicLawPush(c, 1, CardDesignClover, 1)

		require.NoError(t, c.AutoComplete())
		assert.Len(t, c.GetFoundation()[0], 1)
		assert.Len(t, c.GetFoundation()[1], 1)
		assert.Len(t, c.GetTableau()[0], 1)
	})

	t.Run("refuses when nothing can move", func(t *testing.T) {
		c := newTestSalicLaw()
		clearSalicLawBoard(c)
		openAllSalicLawPiles(c)
		assert.Error(t, c.AutoComplete())
	})
}

func TestSalicLaw_GameClear(t *testing.T) {
	c := newTestSalicLaw()
	clearSalicLawBoard(c)
	openAllSalicLawPiles(c)
	// 組札 0 だけ 1 枚足りない状態を作る。7 つ埋めただけでクリアになるなら、
	// checkGameClear が全部を見ていない。
	for i := range SalicLawFoundationCnt {
		last := SalicLawFoundationTarget
		if i == 0 {
			last = SalicLawFoundationTarget - 1
		}
		for v := 1; v <= last; v++ {
			c.foundation[i] = append(c.foundation[i], NewCard(CardDesignSpade, v, true))
		}
	}
	require.NotEqual(t, SalicLawPhaseGameClear, c.GetPhase(), "1 枚足りないうちはクリアでない")

	// 最後の 1 枚を通常の手で送る。checkGameClear を直接叩かず、手からクリアに
	// 入ることを見る。
	salicLawPush(c, 0, CardDesignSpade, SalicLawFoundationTarget)
	require.NoError(t, c.MoveTableauToFoundation(0))

	assert.Equal(t, SalicLawPhaseGameClear, c.GetPhase())
	assert.True(t, c.GetGameEndFlag())
}

func TestSalicLaw_GiveUp(t *testing.T) {
	c := newTestSalicLaw()
	c.GiveUp()
	assert.Equal(t, SalicLawPhaseGameOver, c.GetPhase())
	assert.True(t, c.GetGameEndFlag())

	// 二度目は何も起きない。
	c.GiveUp()
	assert.Equal(t, SalicLawPhaseGameOver, c.GetPhase())
}

func TestSalicLaw_Undo(t *testing.T) {
	c := newTestSalicLaw()
	clearSalicLawBoard(c)
	openAllSalicLawPiles(c)
	salicLawPush(c, 0, CardDesignHeart, 1)

	require.NoError(t, c.MoveTableauToFoundation(0))
	require.True(t, c.CanUndo())
	require.NoError(t, c.Undo())

	assert.Len(t, c.GetTableau()[0], 2)
	assert.Empty(t, c.GetFoundation()[0])
	assert.False(t, c.CanUndo())
	assert.Error(t, c.Undo())
}

// **配りもアンドゥできる。**K で開いた列は、戻したら閉じていなければならない。
// openPiles をスナップショットに載せ忘れると、盤は戻るのに列数だけ進んだままになり、
// 次の Draw が既に札のある列を上書きする。
func TestSalicLaw_UndoClosesAPileThatAKingOpened(t *testing.T) {
	c := newTestSalicLaw()
	clearSalicLawBoard(c)
	c.tableau[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, true)}
	c.openPiles = 1
	c.stock = []*Card{NewCard(CardDesignHeart, CardValueMax, true)}

	require.NoError(t, c.Draw())
	require.Equal(t, 2, c.GetOpenPiles())

	require.NoError(t, c.Undo())
	assert.Equal(t, 1, c.GetOpenPiles(), "戻したら列は閉じる")
	assert.Empty(t, c.GetTableau()[1])
	assert.Equal(t, 1, c.GetStockCount())
}

func TestSalicLaw_UndoN(t *testing.T) {
	c := newTestSalicLaw()
	clearSalicLawBoard(c)
	openAllSalicLawPiles(c)
	salicLawPush(c, 0, CardDesignHeart, 1)
	salicLawPush(c, 1, CardDesignClover, 1)

	require.NoError(t, c.MoveTableauToFoundation(0))
	require.NoError(t, c.MoveTableauToFoundation(1))
	require.NoError(t, c.UndoN(2))

	assert.Len(t, c.GetTableau()[0], 2)
	assert.Len(t, c.GetTableau()[1], 2)
	assert.Error(t, c.UndoN(1))
}

func TestSalicLaw_UndoToEscape(t *testing.T) {
	c := newTestSalicLaw()
	clearSalicLawBoard(c)
	// 列0 に 2 枚積んでおく。1 枚だけだと、動かした瞬間に列0 自身が
	// 「K だけの列」になって手が戻せてしまい、手詰まりにならない。
	c.tableau[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, true)}
	c.tableau[1] = []*Card{NewCard(CardDesignClover, CardValueMax, true)}
	c.openPiles = 2
	salicLawPush(c, 0, CardDesignHeart, 9)
	salicLawPush(c, 0, CardDesignDiamond, 8)
	c.checkStalemate()
	require.False(t, c.IsStalemate(), "♦8 は列1へ動かせる")

	require.NoError(t, c.MoveTableauToTableau(0, 1))
	c.checkStalemate()
	require.True(t, c.IsStalemate(), "K だけの列がもう無い")

	assert.Equal(t, 1, c.UndoToEscape())
}

func TestSalicLaw_ActionLog(t *testing.T) {
	c := newTestSalicLaw()
	clearSalicLawBoard(c)
	openAllSalicLawPiles(c)
	salicLawPush(c, 0, CardDesignHeart, 1)
	require.NoError(t, c.MoveTableauToFoundation(0))

	log := c.GetActionLog()
	require.NotEmpty(t, log)
	assert.Equal(t, "move", log[len(log)-1].ActionType)
}

func TestSalicLaw_JSONRoundTrip(t *testing.T) {
	c := newTestSalicLaw()
	require.NoError(t, c.Draw())

	data, err := json.Marshal(c)
	require.NoError(t, err)

	restored := NewDefaultSalicLaw()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, c.GetStockCount(), restored.GetStockCount())
	assert.Equal(t, c.GetOpenPiles(), restored.GetOpenPiles())
	assert.Equal(t, c.GetMoveCount(), restored.GetMoveCount())
	assert.Equal(t, c.GetTableau(), restored.GetTableau())
}

func TestSalicLaw_UndoSurvivesAKVRoundTrip(t *testing.T) {
	c := newTestSalicLaw()
	clearSalicLawBoard(c)
	openAllSalicLawPiles(c)
	salicLawPush(c, 0, CardDesignHeart, 1)
	require.NoError(t, c.MoveTableauToFoundation(0))

	data, err := json.Marshal(c)
	require.NoError(t, err)
	restored := NewDefaultSalicLaw()
	require.NoError(t, json.Unmarshal(data, restored))

	require.True(t, restored.CanUndo(), "アンドゥ履歴が KV を越えて残る")
	require.NoError(t, restored.Undo())
	assert.Len(t, restored.GetTableau()[0], 2)
	assert.Empty(t, restored.GetFoundation()[0])
}

func TestSalicLaw_UnmarshalJSONValidation(t *testing.T) {
	huge := make([]*Card, SalicLawTotalCards+1)
	for i := range huge {
		huge[i] = NewCard(CardDesignSpade, 1, true)
	}
	overFoundation := make([]*Card, SalicLawFoundationTarget+1)
	for i := range overFoundation {
		overFoundation[i] = NewCard(CardDesignSpade, 1, true)
	}
	overGuard := make([]*Card, salicLawMaxSliceLen+1)
	for i := range overGuard {
		overGuard[i] = NewCard(CardDesignSpade, 1, true)
	}
	bigLogEntries := make([]*ActionLogEntry, salicLawMaxSliceLen+1)
	for i := range bigLogEntries {
		bigLogEntries[i] = &ActionLogEntry{}
	}

	for _, tc := range []struct {
		name string
		j    salicLawJSON
	}{
		{"phase below range", salicLawJSON{Phase: -1}},
		{"phase above range", salicLawJSON{Phase: SalicLawPhaseGameOver + 1}},
		{"negative move count", salicLawJSON{MoveCount: -1}},
		{"stock overflows", salicLawJSON{Stock: huge}},
		{"open piles below range", salicLawJSON{OpenPiles: -1}},
		{"open piles above range", salicLawJSON{OpenPiles: SalicLawTableauCnt + 1}},
		{"foundation overflows", salicLawJSON{Foundation: [SalicLawFoundationCnt][]*Card{overFoundation}}},
		{"tableau overflows", salicLawJSON{Tableau: [SalicLawTableauCnt][]*Card{huge}}},
		{"action log overflows", salicLawJSON{ActionLog: bigLogEntries}},
		{"history overflows", salicLawJSON{History: make([]*salicLawSnapshot, salicLawMaxSliceLen+1)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(&tc.j)
			require.NoError(t, err)
			assert.Error(t, NewDefaultSalicLaw().UnmarshalJSON(data))
		})
	}

	t.Run("malformed json", func(t *testing.T) {
		assert.Error(t, NewDefaultSalicLaw().UnmarshalJSON([]byte("not json")))
	})

	t.Run("an oversized pile inside a snapshot", func(t *testing.T) {
		data, err := json.Marshal(&salicLawSnapshotJSON{Stock: overGuard})
		require.NoError(t, err)
		assert.Error(t, new(salicLawSnapshot).UnmarshalJSON(data))
	})

	t.Run("malformed snapshot json", func(t *testing.T) {
		assert.Error(t, new(salicLawSnapshot).UnmarshalJSON([]byte("not json")))
	})

	t.Run("a valid snapshot is accepted", func(t *testing.T) {
		data, err := json.Marshal(&salicLawJSON{Phase: SalicLawPhasePlaying, MoveCount: 3})
		require.NoError(t, err)
		c := NewDefaultSalicLaw()
		require.NoError(t, c.UnmarshalJSON(data))
		assert.Equal(t, 3, c.GetMoveCount())
	})
}

// Drive a full game to make sure no path panics and the invariants hold
// throughout. The deal is random, so this is a fuzz-ish smoke test rather than
// an assertion about any particular position.
func TestSalicLaw_FullGameDrive(t *testing.T) {
	for range 20 {
		c := newTestSalicLaw()
		for range 600 {
			if c.GetGameEndFlag() {
				break
			}
			h := c.GetHint()
			if h == nil {
				break
			}
			var err error
			switch {
			case h.FromZone == "stock":
				err = c.Draw()
			case h.ToZone == "foundation":
				err = c.MoveTableauToFoundation(h.FromIdx)
			default:
				err = c.MoveTableauToTableau(h.FromIdx, h.ToIdx)
			}
			require.NoError(t, err)

			for i := range SalicLawFoundationCnt {
				assert.LessOrEqual(t, len(c.GetFoundation()[i]), SalicLawFoundationTarget)
			}
			// K は必ず列の底に据わったまま。剥がれたら勘定が崩れる。
			for i := range c.GetOpenPiles() {
				require.NotEmpty(t, c.GetTableau()[i])
				assert.Equal(t, CardValueMax, c.GetTableau()[i][0].GetValue())
			}
		}
	}
}

// プレイヤーが踏める拒否はすべてメッセージコードを名乗ること。生の英文を返すと
// 日本語ロケールの CUI が英語のまま表示する。
func TestSalicLaw_PlayerFacingErrorsCarryAMessageCode(t *testing.T) {
	setup := func(fn func(c *SalicLaw)) *SalicLaw {
		c := newTestSalicLaw()
		clearSalicLawBoard(c)
		openAllSalicLawPiles(c)
		fn(c)
		return c
	}

	cases := []struct {
		name string
		code string
		act  func() error
	}{
		{"stock exhausted", "saliclaw.errStockEmptyNoRedeal", func() error {
			return setup(func(c *SalicLaw) {}).Draw()
		}},
		{"column empty", "saliclaw.errPileEmpty", func() error {
			c := setup(func(c *SalicLaw) { c.tableau[0] = nil })
			return c.MoveTableauToFoundation(0)
		}},
		{"the base king", "saliclaw.errKingIsTheBase", func() error {
			return setup(func(c *SalicLaw) {}).MoveTableauToFoundation(0)
		}},
		{"no foundation for that card", "saliclaw.errNoFoundationForCard", func() error {
			c := setup(func(c *SalicLaw) { salicLawPush(c, 0, CardDesignSpade, 9) })
			return c.MoveTableauToFoundation(0)
		}},
		{"same column", "saliclaw.errSamePile", func() error {
			return setup(func(c *SalicLaw) {}).MoveTableauToTableau(1, 1)
		}},
		{"source column empty", "saliclaw.errPileEmpty", func() error {
			c := setup(func(c *SalicLaw) { c.tableau[0] = nil })
			return c.MoveTableauToTableau(0, 1)
		}},
		{"the base king cannot be moved sideways", "saliclaw.errKingIsTheBase", func() error {
			return setup(func(c *SalicLaw) {}).MoveTableauToTableau(0, 1)
		}},
		{"destination is not a bare king", "saliclaw.errCannotPlaceOnPile", func() error {
			c := setup(func(c *SalicLaw) {
				salicLawPush(c, 0, CardDesignSpade, 2)
				salicLawPush(c, 1, CardDesignHeart, 7)
			})
			return c.MoveTableauToTableau(0, 1)
		}},
		{"nothing to auto-complete", "saliclaw.errNothingToAutoComplete", func() error {
			return setup(func(c *SalicLaw) {}).AutoComplete()
		}},
		{"nothing to undo", "saliclaw.errNothingToUndo", func() error {
			c := setup(func(c *SalicLaw) {})
			c.history = nil
			return c.Undo()
		}},
		{"not playing", "saliclaw.errNotPlaying", func() error {
			c := setup(func(c *SalicLaw) { c.phase = SalicLawPhaseGameOver })
			return c.Draw()
		}},
		{"invalid column", "saliclaw.errInvalidPile", func() error {
			return setup(func(c *SalicLaw) {}).MoveTableauToFoundation(99)
		}},
		{"invalid column sideways", "saliclaw.errInvalidPile", func() error {
			return setup(func(c *SalicLaw) {}).MoveTableauToTableau(0, 99)
		}},
		{"not playing sideways", "saliclaw.errNotPlaying", func() error {
			c := setup(func(c *SalicLaw) { c.phase = SalicLawPhaseGameOver })
			return c.MoveTableauToTableau(0, 1)
		}},
		{"not playing to a foundation", "saliclaw.errNotPlaying", func() error {
			c := setup(func(c *SalicLaw) { c.phase = SalicLawPhaseGameOver })
			return c.MoveTableauToFoundation(0)
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.act()
			require.Error(t, err)
			code, _ := ErrorMessageCode(err)
			assert.Equal(t, tc.code, code)
			// センチネルは errors.Is で見分けられること。コードだけ足して
			// 種別を失うと、呼び出し側の分岐が黙って死ぬ。
			assert.True(t, errors.Is(err, ErrInvalidPlay) || errors.Is(err, ErrWrongPhase) ||
				errors.Is(err, ErrDeckExhausted), "unexpected sentinel: %v", err)
		})
	}
}

// AutoComplete がプレイ中以外で拒む経路も同じコードを名乗ること
// (requirePlaying を通らない独自チェックなので、上の表とは別に見る)。
func TestSalicLaw_AutoCompleteOutsidePlayingCarriesACode(t *testing.T) {
	c := newTestSalicLaw()
	c.phase = SalicLawPhaseGameOver
	err := c.AutoComplete()
	require.Error(t, err)
	code, _ := ErrorMessageCode(err)
	assert.Equal(t, "saliclaw.errNotPlaying", code)
}
