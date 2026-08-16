//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- helpers ---

func newTestCruel() *domain.Cruel {
	tc := domain.NewTrumpCards(0)
	return domain.NewCruel(tc)
}

func setupPlayingCruel() *domain.Cruel {
	c := newTestCruel()
	c.Reset()
	return c
}

func clearCruelTableau(c *domain.Cruel) {
	var empty [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	c.SetTableau(empty)
}

func setCruelFoundationToAces(c *domain.Cruel) {
	var f [domain.CruelFoundationCnt][]*domain.Card
	for i := range domain.CruelFoundationCnt {
		f[i] = []*domain.Card{makeCard(i+1, 1)}
	}
	c.SetFoundation(f)
}

// --- Tests ---

func TestNewCruel(t *testing.T) {
	c := newTestCruel()
	assert.NotNil(t, c)
	assert.Equal(t, domain.CruelPhase(0), c.GetPhase())
}

func TestNewDefaultCruel(t *testing.T) {
	c := domain.NewDefaultCruel()
	assert.NotNil(t, c)
}

func TestCruel_Reset(t *testing.T) {
	c := setupPlayingCruel()

	assert.Equal(t, domain.CruelPhasePlaying, c.GetPhase())
	assert.Equal(t, 0, c.GetMoveCount())
	assert.False(t, c.CanUndo())
	assert.False(t, c.IsStalemate())

	// Foundations: every suit must hold exactly one Ace.
	foundation := c.GetFoundation()
	for i := range domain.CruelFoundationCnt {
		require.Equal(t, 1, len(foundation[i]), "foundation %d should hold one Ace", i)
		assert.Equal(t, 1, foundation[i][0].GetValue(), "foundation %d top must be Ace", i)
		assert.Equal(t, i+1, foundation[i][0].GetDesign(), "foundation %d must hold the matching suit", i)
	}

	// Tableau: 12 columns × 4 cards face-up = 48 cards, no Aces.
	tableau := c.GetTableau()
	total := 0
	for col := range domain.CruelTableauCnt {
		assert.Equal(t, domain.CruelInitialColSize, len(tableau[col]), "column %d should have 4 cards", col)
		for _, tc := range tableau[col] {
			assert.True(t, tc.FaceUp, "all tableau cards must be face-up")
			assert.NotEqual(t, 1, tc.Card.GetValue(), "Aces must not appear in the tableau")
		}
		total += len(tableau[col])
	}
	assert.Equal(t, 48, total)
}

func TestCruel_MoveTableauToTableau_Valid(t *testing.T) {
	c := setupPlayingCruel()
	clearCruelTableau(c)
	setCruelFoundationToAces(c)

	var tab [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	// fromCol=0 top: Spade 5
	tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 5, true)}
	// toCol=1 top: Spade 6 — same suit, one higher.
	tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 6, true)}
	c.SetTableau(tab)

	require.NoError(t, c.MoveTableauToTableau(0, 1))
	assert.Equal(t, 0, len(c.GetTableau()[0]))
	assert.Equal(t, 2, len(c.GetTableau()[1]))
	assert.Equal(t, 1, c.GetMoveCount())
	assert.True(t, c.CanUndo())
}

func TestCruel_MoveTableauToTableau_Errors(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*domain.Cruel)
		from    int
		to      int
		wantErr string
	}{
		{
			name:  "phase not playing",
			setup: func(c *domain.Cruel) { c.SetPhase(domain.CruelPhaseGameOver) },
			from:  0, to: 1,
			wantErr: "not in playing phase",
		},
		{
			name:  "invalid from column negative",
			setup: func(c *domain.Cruel) {},
			from:  -1, to: 1,
			wantErr: "invalid from column",
		},
		{
			name:  "invalid from column high",
			setup: func(c *domain.Cruel) {},
			from:  domain.CruelTableauCnt, to: 1,
			wantErr: "invalid from column",
		},
		{
			name:  "invalid to column",
			setup: func(c *domain.Cruel) {},
			from:  0, to: domain.CruelTableauCnt,
			wantErr: "invalid to column",
		},
		{
			name:  "same column",
			setup: func(c *domain.Cruel) {},
			from:  0, to: 0,
			wantErr: "from and to columns are the same",
		},
		{
			name: "empty source",
			setup: func(c *domain.Cruel) {
				clearCruelTableau(c)
			},
			from: 0, to: 1,
			wantErr: "tableau column is empty",
		},
		{
			name: "destination empty (cannot place on empty col)",
			setup: func(c *domain.Cruel) {
				clearCruelTableau(c)
				var tab [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
				tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 13, true)}
				c.SetTableau(tab)
			},
			from: 0, to: 1,
			wantErr: "cannot place card on tableau",
		},
		{
			name: "mismatched suit",
			setup: func(c *domain.Cruel) {
				clearCruelTableau(c)
				var tab [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
				tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 5, true)}
				tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 6, true)}
				c.SetTableau(tab)
			},
			from: 0, to: 1,
			wantErr: "cannot place card on tableau",
		},
		{
			name: "mismatched rank",
			setup: func(c *domain.Cruel) {
				clearCruelTableau(c)
				var tab [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
				tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 4, true)}
				tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 6, true)}
				c.SetTableau(tab)
			},
			from: 0, to: 1,
			wantErr: "cannot place card on tableau",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := setupPlayingCruel()
			tt.setup(c)
			err := c.MoveTableauToTableau(tt.from, tt.to)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestCruel_MoveTableauToFoundation_Valid(t *testing.T) {
	c := setupPlayingCruel()
	clearCruelTableau(c)
	setCruelFoundationToAces(c)

	var tab [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 2, true)}
	c.SetTableau(tab)

	require.NoError(t, c.MoveTableauToFoundation(0))
	assert.Equal(t, 0, len(c.GetTableau()[0]))
	assert.Equal(t, 2, len(c.GetFoundation()[domain.CardDesignSpade-1]))
	assert.Equal(t, 1, c.GetMoveCount())
}

func TestCruel_MoveTableauToFoundation_Errors(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*domain.Cruel)
		col     int
		wantErr string
	}{
		{
			name:    "phase not playing",
			setup:   func(c *domain.Cruel) { c.SetPhase(domain.CruelPhaseGameClear) },
			col:     0,
			wantErr: "not in playing phase",
		},
		{
			name:    "invalid column negative",
			setup:   func(c *domain.Cruel) {},
			col:     -1,
			wantErr: "invalid column",
		},
		{
			name:    "invalid column high",
			setup:   func(c *domain.Cruel) {},
			col:     domain.CruelTableauCnt,
			wantErr: "invalid column",
		},
		{
			name: "empty column",
			setup: func(c *domain.Cruel) {
				clearCruelTableau(c)
			},
			col:     0,
			wantErr: "tableau column is empty",
		},
		{
			name: "rank mismatch (3 onto A)",
			setup: func(c *domain.Cruel) {
				clearCruelTableau(c)
				setCruelFoundationToAces(c)
				var tab [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
				tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 3, true)}
				c.SetTableau(tab)
			},
			col:     0,
			wantErr: "cannot place card on foundation",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := setupPlayingCruel()
			tt.setup(c)
			err := c.MoveTableauToFoundation(tt.col)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestCruel_Shift_PreservesOrderAndRefills(t *testing.T) {
	c := setupPlayingCruel()
	clearCruelTableau(c)

	// Build a deterministic 6-card layout split across 3 columns.
	var tab [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	tab[0] = []*domain.KlondikeTableauCard{
		makeTableauCard(domain.CardDesignSpade, 2, true),
		makeTableauCard(domain.CardDesignSpade, 3, true),
	}
	tab[1] = []*domain.KlondikeTableauCard{
		makeTableauCard(domain.CardDesignHeart, 4, true),
		makeTableauCard(domain.CardDesignHeart, 5, true),
		makeTableauCard(domain.CardDesignHeart, 6, true),
	}
	tab[2] = []*domain.KlondikeTableauCard{
		makeTableauCard(domain.CardDesignClover, 7, true),
	}
	c.SetTableau(tab)

	require.NoError(t, c.Shift())

	// With only 6 cards, columns 0 and 1 fill (4 + 2), the rest stay empty.
	got := c.GetTableau()
	require.Equal(t, 4, len(got[0]))
	require.Equal(t, 2, len(got[1]))
	for i := 2; i < domain.CruelTableauCnt; i++ {
		assert.Equal(t, 0, len(got[i]), "column %d should be empty after shift", i)
	}

	// Relative order is preserved (col-major flatten then 4-per-col redeal).
	expected := []struct {
		design, value int
	}{
		{domain.CardDesignSpade, 2},
		{domain.CardDesignSpade, 3},
		{domain.CardDesignHeart, 4},
		{domain.CardDesignHeart, 5},
		// → column boundary
		{domain.CardDesignHeart, 6},
		{domain.CardDesignClover, 7},
	}
	flat := append([]*domain.KlondikeTableauCard{}, got[0]...)
	flat = append(flat, got[1]...)
	for i, want := range expected {
		assert.Equal(t, want.design, flat[i].Card.GetDesign(), "card %d design", i)
		assert.Equal(t, want.value, flat[i].Card.GetValue(), "card %d value", i)
	}
	assert.True(t, c.CanUndo())
}

func TestCruel_Shift_Errors(t *testing.T) {
	t.Run("phase not playing", func(t *testing.T) {
		c := setupPlayingCruel()
		c.SetPhase(domain.CruelPhaseGameClear)
		err := c.Shift()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not in playing phase")
	})
	t.Run("no tableau cards", func(t *testing.T) {
		c := setupPlayingCruel()
		clearCruelTableau(c)
		err := c.Shift()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no tableau cards")
	})
}

func TestCruel_Shift_UndoRestoresLayout(t *testing.T) {
	c := setupPlayingCruel()
	originalTableau := deepCopyCruelTableau(c.GetTableau())

	require.NoError(t, c.Shift())
	require.NoError(t, c.Undo())

	restored := c.GetTableau()
	for i := range domain.CruelTableauCnt {
		require.Equal(t, len(originalTableau[i]), len(restored[i]), "col %d length", i)
		for j, tc := range originalTableau[i] {
			assert.Equal(t, tc.Card.GetDesign(), restored[i][j].Card.GetDesign(),
				"col %d card %d design after undo", i, j)
			assert.Equal(t, tc.Card.GetValue(), restored[i][j].Card.GetValue(),
				"col %d card %d value after undo", i, j)
		}
	}
}

func deepCopyCruelTableau(t [domain.CruelTableauCnt][]*domain.KlondikeTableauCard) [domain.CruelTableauCnt][]*domain.KlondikeTableauCard {
	var out [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	for i := range domain.CruelTableauCnt {
		out[i] = make([]*domain.KlondikeTableauCard, len(t[i]))
		for j, tc := range t[i] {
			out[i][j] = &domain.KlondikeTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	return out
}

func TestCruel_GiveUp(t *testing.T) {
	c := setupPlayingCruel()
	c.GiveUp()
	assert.Equal(t, domain.CruelPhaseGameOver, c.GetPhase())
	// Re-calling on a non-playing game must be a no-op.
	c.GiveUp()
	assert.Equal(t, domain.CruelPhaseGameOver, c.GetPhase())
}

func TestCruel_GetHint_FoundationPriority(t *testing.T) {
	c := setupPlayingCruel()
	clearCruelTableau(c)
	setCruelFoundationToAces(c)

	var tab [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	// col 0 has a Spade 2 (foundation-ready); col 1/2 hold tableau-only moves.
	tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 2, true)}
	tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 5, true)}
	tab[2] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignHeart, 6, true)}
	c.SetTableau(tab)

	hint := c.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "foundation", hint.ToZone)
	assert.Equal(t, 0, hint.FromCol)
}

func TestCruel_GetHint_TableauFallback(t *testing.T) {
	c := setupPlayingCruel()
	clearCruelTableau(c)
	setCruelFoundationToAces(c)

	var tab [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 5, true)}
	tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 6, true)}
	c.SetTableau(tab)

	hint := c.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "tableau", hint.ToZone)
	assert.Equal(t, 0, hint.FromCol)
	assert.Equal(t, 1, hint.ToCol)
}

func TestCruel_GetHint_NoneWhenStuck(t *testing.T) {
	c := setupPlayingCruel()
	clearCruelTableau(c)
	setCruelFoundationToAces(c)
	// Single card with no destination.
	var tab [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 13, true)}
	c.SetTableau(tab)
	assert.Nil(t, c.GetHint())
}

func TestCruel_GetHint_NilWhenNotPlaying(t *testing.T) {
	c := setupPlayingCruel()
	c.SetPhase(domain.CruelPhaseGameOver)
	assert.Nil(t, c.GetHint())
}

func TestCruel_Stalemate(t *testing.T) {
	c := setupPlayingCruel()
	clearCruelTableau(c)
	setCruelFoundationToAces(c)

	// Move a card so checkStalemate runs after a real action.
	var tab [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 5, true)}
	tab[1] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 6, true)}
	c.SetTableau(tab)

	require.NoError(t, c.MoveTableauToTableau(0, 1))
	// After moving 5♠ onto 6♠, no playable card remains for tableau→tableau,
	// and 7♠ is not exposed for foundation. Stalemate flag must be set.
	assert.True(t, c.IsStalemate())
}

func TestCruel_AutoComplete_Success(t *testing.T) {
	c := setupPlayingCruel()
	clearCruelTableau(c)
	setCruelFoundationToAces(c)

	var tab [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	// Each column ends with the next card of its suit — auto-complete sweeps them.
	for suit := 1; suit <= domain.CruelFoundationCnt; suit++ {
		tab[suit-1] = []*domain.KlondikeTableauCard{
			makeTableauCard(suit, 2, true),
		}
	}
	c.SetTableau(tab)

	require.NoError(t, c.AutoComplete())
	for i := range domain.CruelFoundationCnt {
		assert.Equal(t, 2, len(c.GetFoundation()[i]))
	}
}

func TestCruel_AutoComplete_GameClear(t *testing.T) {
	c := setupPlayingCruel()
	clearCruelTableau(c)

	// All Aces already placed.
	var f [domain.CruelFoundationCnt][]*domain.Card
	for s := 1; s <= domain.CruelFoundationCnt; s++ {
		pile := make([]*domain.Card, 0, domain.CardValueMax-1)
		for v := 1; v <= domain.CardValueMax-1; v++ {
			pile = append(pile, makeCard(s, v))
		}
		f[s-1] = pile
	}
	c.SetFoundation(f)

	// Single K in tableau per suit — sweeps to foundation and clears the game.
	var tab [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	for s := 1; s <= domain.CruelFoundationCnt; s++ {
		tab[s-1] = []*domain.KlondikeTableauCard{makeTableauCard(s, domain.CardValueMax, true)}
	}
	c.SetTableau(tab)

	require.NoError(t, c.AutoComplete())
	assert.Equal(t, domain.CruelPhaseGameClear, c.GetPhase())
	assert.True(t, c.GetGameEndFlag())
}

func TestCruel_AutoComplete_RefreshesStalemate(t *testing.T) {
	c := setupPlayingCruel()
	clearCruelTableau(c)
	setCruelFoundationToAces(c)

	// Partial-clear scenario: the 2s sweep to foundation, but the remaining
	// cards have no playable move. Stalemate must be re-evaluated.
	var tab [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	for s := 1; s <= domain.CruelFoundationCnt; s++ {
		tab[s-1] = []*domain.KlondikeTableauCard{makeTableauCard(s, 2, true)}
	}
	// Add an isolated K with no matching neighbour to lock the board after the sweep.
	tab[4] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, domain.CardValueMax, true)}
	c.SetTableau(tab)
	c.SetIsStalemate(false)

	require.NoError(t, c.AutoComplete())
	// After AC, the lone K can't move and can't go to foundation (next needed
	// after 2 is 3); the flag must now read true so the UI surfaces Shift.
	assert.True(t, c.IsStalemate(), "AutoComplete should refresh stalemate flag")
}

func TestCruel_AutoComplete_NotPlaying(t *testing.T) {
	c := setupPlayingCruel()
	c.SetPhase(domain.CruelPhaseGameOver)
	err := c.AutoComplete()
	require.Error(t, err)
}

func TestCruel_Undo_NoHistory(t *testing.T) {
	c := setupPlayingCruel()
	err := c.Undo()
	require.Error(t, err)
}

func TestCruel_Undo_NotPlaying(t *testing.T) {
	c := setupPlayingCruel()
	c.SetPhase(domain.CruelPhaseGameOver)
	err := c.Undo()
	require.Error(t, err)
}

func TestCruel_UndoN(t *testing.T) {
	c := setupPlayingCruel()
	clearCruelTableau(c)
	setCruelFoundationToAces(c)

	var tab [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	tab[0] = []*domain.KlondikeTableauCard{
		makeTableauCard(domain.CardDesignSpade, 4, true),
		makeTableauCard(domain.CardDesignSpade, 3, true),
		makeTableauCard(domain.CardDesignSpade, 2, true),
	}
	c.SetTableau(tab)

	require.NoError(t, c.MoveTableauToFoundation(0))
	require.NoError(t, c.MoveTableauToFoundation(0))

	require.NoError(t, c.UndoN(2))
	assert.Equal(t, 3, len(c.GetTableau()[0]))
}

func TestCruel_UndoN_PartialFailurePropagates(t *testing.T) {
	c := setupPlayingCruel()
	// No history → first undo fails immediately.
	err := c.UndoN(1)
	require.Error(t, err)
}

func TestCruel_UndoToEscape(t *testing.T) {
	c := setupPlayingCruel()
	assert.Equal(t, 0, c.UndoToEscape(), "not in stalemate → 0")

	c.SetIsStalemate(true)
	// No history captured → unrecoverable.
	assert.Equal(t, -1, c.UndoToEscape())
}

func TestCruel_GetActionLog(t *testing.T) {
	c := setupPlayingCruel()
	log := c.GetActionLog()
	assert.Empty(t, log)

	clearCruelTableau(c)
	setCruelFoundationToAces(c)
	var tab [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	tab[0] = []*domain.KlondikeTableauCard{makeTableauCard(domain.CardDesignSpade, 2, true)}
	c.SetTableau(tab)
	require.NoError(t, c.MoveTableauToFoundation(0))

	log = c.GetActionLog()
	require.Len(t, log, 1)
	assert.Equal(t, "move", log[0].ActionType)
}

// #5496: AutoComplete は1枚も動かなくても「オートコンプリートを実行しました」を
// 行動ログに残し、undo 用のスナップショットまで積んでいた。**何も起きていないのに
// 記録だけが残る。**
func TestCruel_AutoCompleteRecordsNothingWhenNothingMoves(t *testing.T) {
	c := setupPlayingCruel()
	clearCruelTableau(c)
	// Reset は各組札にエースを置く。タブローの一番上が5なら、次に要る2ではない
	// のでどこにも置けない。
	var tab [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	tab[0] = []*domain.KlondikeTableauCard{{Card: makeCard(domain.CardDesignSpade, 5), FaceUp: true}}
	c.SetTableau(tab)

	logsBefore := len(c.GetActionLog())
	movesBefore := c.GetMoveCount()
	canUndoBefore := c.CanUndo()

	require.NoError(t, c.AutoComplete())

	assert.Equal(t, movesBefore, c.GetMoveCount(), "1枚も動いていない")
	assert.Len(t, c.GetActionLog(), logsBefore, "動いていないのに実行記録が残っている")
	assert.Equal(t, canUndoBefore, c.CanUndo(), "何もしていないのに undo 先が積まれている")
}

// 正常系は今までどおり記録する。上のテストが「常にログを残さない」実装でも
// 通ってしまわないための負のコントロール。
func TestCruel_AutoCompleteStillRecordsWhenCardsMove(t *testing.T) {
	c := setupPlayingCruel()
	clearCruelTableau(c)
	var tab [domain.CruelTableauCnt][]*domain.KlondikeTableauCard
	// 組札の♠はエースまで出ているので、次に置けるのは2。
	tab[0] = []*domain.KlondikeTableauCard{{Card: makeCard(domain.CardDesignSpade, 2), FaceUp: true}}
	c.SetTableau(tab)

	logsBefore := len(c.GetActionLog())
	require.NoError(t, c.AutoComplete())

	assert.Positive(t, c.GetMoveCount())
	assert.Greater(t, len(c.GetActionLog()), logsBefore)
	assert.True(t, c.CanUndo())
}
