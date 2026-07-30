//go:build test

package domain

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ctCard(design, value int) *Card { return NewCard(design, value, true) }

func TestChineseTen_TieScoreIsHalfTheRedPoints(t *testing.T) {
	// The tie score of 105 only means "an even split" if the red cards are
	// worth 210 between them. Counting the deck is what verifies the scoring
	// table -- get any rank's value wrong and these two stop matching.
	total := 0
	reds, blacks := 0, 0
	for _, c := range newChineseTenDeck() {
		pts := ChineseTenCardPoints(c)
		total += pts
		if chineseTenIsRed(c) {
			reds++
		} else {
			blacks++
			assert.Zero(t, pts, "black cards score nothing in the two-player game")
		}
	}
	assert.Equal(t, 26, reds)
	assert.Equal(t, 26, blacks)
	assert.Equal(t, ChineseTenTotalRedPoints, total)
	assert.Equal(t, 210, total)
	assert.Equal(t, total/2, ChineseTenTieScore)
	assert.Equal(t, 105, ChineseTenTieScore)
}

func TestChineseTen_CardPoints(t *testing.T) {
	for v := 2; v <= 8; v++ {
		assert.Equal(t, v, ChineseTenCardPoints(ctCard(CardDesignHeart, v)), "red %d is face value", v)
	}
	for _, v := range []int{9, 10, 11, 12, 13} {
		assert.Equal(t, 10, ChineseTenCardPoints(ctCard(CardDesignDiamond, v)), "red %d is ten", v)
	}
	assert.Equal(t, 20, ChineseTenCardPoints(ctCard(CardDesignHeart, 1)), "the red ace is the biggest single card")

	// Only red counts. A black ace is worth as much as a black two: nothing.
	assert.Zero(t, ChineseTenCardPoints(ctCard(CardDesignSpade, 1)))
	assert.Zero(t, ChineseTenCardPoints(ctCard(CardDesignClover, 13)))
	assert.Zero(t, ChineseTenCardPoints(nil))
}

func TestChineseTen_Captures(t *testing.T) {
	// A-9 pair to ten; 10/J/Q/K pair by rank. They are SEPARATE rules -- a ten
	// does not capture by summing, and a nine does not capture another nine.
	cases := []struct {
		name           string
		played, target *Card
		want           bool
	}{
		{"ace takes nine", ctCard(CardDesignSpade, 1), ctCard(CardDesignHeart, 9), true},
		{"nine takes ace", ctCard(CardDesignSpade, 9), ctCard(CardDesignHeart, 1), true},
		{"three takes seven", ctCard(CardDesignSpade, 3), ctCard(CardDesignHeart, 7), true},
		{"five takes five", ctCard(CardDesignSpade, 5), ctCard(CardDesignHeart, 5), true},
		{"nine does not take nine", ctCard(CardDesignSpade, 9), ctCard(CardDesignHeart, 9), false},
		{"two does not take seven", ctCard(CardDesignSpade, 2), ctCard(CardDesignHeart, 7), false},
		{"ten takes ten", ctCard(CardDesignSpade, 10), ctCard(CardDesignHeart, 10), true},
		{"king takes king", ctCard(CardDesignSpade, 13), ctCard(CardDesignHeart, 13), true},
		{"queen does not take king", ctCard(CardDesignSpade, 12), ctCard(CardDesignHeart, 13), false},
		{"ten does not take a numeral", ctCard(CardDesignSpade, 10), ctCard(CardDesignHeart, 1), false},
		{"nil either side", nil, ctCard(CardDesignHeart, 1), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, ChineseTenCaptures(tc.played, tc.target))
		})
	}
	assert.False(t, ChineseTenCaptures(ctCard(CardDesignSpade, 1), nil))
}

func TestChineseTen_DealIsTwelveEachAndFourOnTheTable(t *testing.T) {
	// #4378 says "6-8 cards each" and "a few" on the table.
	c := NewDefaultChineseTen()
	c.Reset()

	for i := range c.GetPlayers() {
		assert.Equal(t, ChineseTenHandSize, c.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
	assert.Len(t, c.GetLayout(), ChineseTenLayoutSize)
	assert.Equal(t, ChineseTenDeckSize-2*ChineseTenHandSize-ChineseTenLayoutSize, c.GetStockCount())

	total := len(c.GetLayout()) + c.GetStockCount()
	for i := range c.GetPlayers() {
		total += c.GetPlayer(i).GetCardsSize() + len(c.GetCaptured(i))
	}
	assert.Equal(t, ChineseTenDeckSize, total, "every card is somewhere")
}

func TestChineseTen_ATurnAlsoTurnsAStockCard(t *testing.T) {
	// The second step is the game, and #4378 omits it entirely. One hand card
	// out means one stock card consumed, which is why the two run out together
	// and there is no redeal.
	c := NewDefaultChineseTen()
	c.Reset()
	stockBefore := c.GetStockCount()

	idx := c.GetCurrentPlayerIdx()
	require.NoError(t, c.PlayCard(idx, 0))
	for c.GetPhase() == ChineseTenPhaseSelect {
		require.NoError(t, c.SelectCapture(c.GetCurrentPlayerIdx(), c.GetSelectableIndices()[0]))
	}

	assert.Equal(t, stockBefore-1, c.GetStockCount(),
		"playing one card must also turn exactly one from the stock")
}

func TestChineseTen_HandsAndStockRunOutTogether(t *testing.T) {
	c := NewDefaultChineseTen()
	c.Reset()
	require.True(t, ctPlayOut(t, c))

	assert.Zero(t, c.GetStockCount(), "24 hand cards consume all 24 stock cards")
	for i := range c.GetPlayers() {
		assert.Zero(t, c.GetPlayer(i).GetCardsSize(), "hand %d", i)
	}
}

func TestChineseTen_EveryCardIsAccountedForAtTheEnd(t *testing.T) {
	for range 20 {
		c := NewDefaultChineseTen()
		c.Reset()
		require.True(t, ctPlayOut(t, c))

		total := len(c.GetLayout()) + c.GetStockCount()
		for i := range c.GetPlayers() {
			total += len(c.GetCaptured(i)) + c.GetPlayer(i).GetCardsSize()
		}
		assert.Equal(t, ChineseTenDeckSize, total)
	}
}

func TestChineseTen_ScoresAreTheRedPointsActuallyCaptured(t *testing.T) {
	for range 20 {
		c := NewDefaultChineseTen()
		c.Reset()
		require.True(t, ctPlayOut(t, c))

		for i := range c.GetPlayers() {
			want := 0
			for _, card := range c.GetCaptured(i) {
				want += ChineseTenCardPoints(card)
			}
			assert.Equal(t, want, c.GetScore(i), "seat %d's score must equal its captured reds", i)
		}
		// Cards stranded in the layout score for nobody, so the two scores sum
		// to at most the deck's 210 -- never more.
		assert.LessOrEqual(t, c.GetScore(0)+c.GetScore(1), ChineseTenTotalRedPoints)
	}
}

func TestChineseTen_OnePlayTakesAtMostOneLayoutCard(t *testing.T) {
	// Two fives on the table and a five in hand: only one of them may go.
	c := NewDefaultChineseTen()
	c.Reset()
	c.SetLayoutForTest([]*Card{
		ctCard(CardDesignHeart, 5), ctCard(CardDesignDiamond, 5), ctCard(CardDesignSpade, 13),
	})
	// Pin the stock: the turn's second step turns a card that captures too, so
	// a shuffled stock makes the captured count depend on the deal. Same trap
	// as TestChineseTen_ANonCapturingCardJoinsTheLayout -- I made it twice.
	c.SetStockForTest([]*Card{ctCard(CardDesignClover, 2)}) // needs an eight; there is none
	p := c.GetPlayer(0)
	p.Reset()
	p.AddCard(ctCard(CardDesignClover, 5))
	c.SetCurrentPlayerForTest(0)

	require.NoError(t, c.PlayCard(0, 0))
	require.Equal(t, ChineseTenPhaseSelect, c.GetPhase(), "two matches means a choice")
	assert.Len(t, c.GetSelectableIndices(), 2)

	require.NoError(t, c.SelectCapture(0, c.GetSelectableIndices()[0]))
	assert.Len(t, c.GetCaptured(0), 2, "the played card plus exactly one from the table")
	assert.Len(t, c.GetLayout(), 3, "the other five stays, and the turned two joins it")
}

func TestChineseTen_ANonCapturingCardJoinsTheLayout(t *testing.T) {
	c := NewDefaultChineseTen()
	c.Reset()
	c.SetLayoutForTest([]*Card{ctCard(CardDesignSpade, 13)})
	// The stock must be pinned too. The turn's SECOND step turns a card that
	// can capture on its own -- including capturing the card just placed -- so
	// leaving the stock shuffled makes this assertion a coin flip. It was one:
	// -count=100 caught it before this ever reached CI.
	c.SetStockForTest([]*Card{ctCard(CardDesignClover, 3)}) // needs a seven; there is none
	p := c.GetPlayer(0)
	p.Reset()
	p.AddCard(ctCard(CardDesignClover, 2)) // needs an eight; there is none
	c.SetCurrentPlayerForTest(0)

	require.NoError(t, c.PlayCard(0, 0))
	assert.Empty(t, c.GetCaptured(0), "neither the played card nor the turned one could capture")
	assert.Len(t, c.GetLayout(), 3, "both cards joined the layout")
}

func TestChineseTen_TheTurnedCardCanCaptureOnItsOwn(t *testing.T) {
	// The other half of the same rule: the hand card whiffs, but the card
	// turned from the stock takes something. This is what makes the second
	// step worth having, and what made the test above flaky.
	c := NewDefaultChineseTen()
	c.Reset()
	c.SetLayoutForTest([]*Card{ctCard(CardDesignHeart, 9)})
	c.SetStockForTest([]*Card{ctCard(CardDesignClover, 1)}) // an ace takes the nine
	p := c.GetPlayer(0)
	p.Reset()
	p.AddCard(ctCard(CardDesignClover, 13)) // a king takes nothing here
	c.SetCurrentPlayerForTest(0)

	require.NoError(t, c.PlayCard(0, 0))
	assert.Len(t, c.GetCaptured(0), 2, "the turned ace took the nine")
	assert.Equal(t, ChineseTenCardPoints(ctCard(CardDesignHeart, 9)), c.GetScore(0),
		"the red nine is worth ten and the black ace nothing")
}

func TestChineseTen_RejectsIllegalRequests(t *testing.T) {
	c := NewDefaultChineseTen()
	c.Reset()
	cur := c.GetCurrentPlayerIdx()

	assert.Error(t, c.PlayCard(cur, -1))
	assert.Error(t, c.PlayCard(cur, 99))
	assert.Error(t, c.PlayCard((cur+1)%ChineseTenPlayerCnt, 0), "not that player's turn")
	assert.Error(t, c.SelectCapture(cur, 0), "no selection is pending")
}

func TestChineseTen_SelectCaptureRejectsAnUnmatchableCard(t *testing.T) {
	c := NewDefaultChineseTen()
	c.Reset()
	c.SetLayoutForTest([]*Card{
		ctCard(CardDesignHeart, 5), ctCard(CardDesignDiamond, 5), ctCard(CardDesignSpade, 13),
	})
	p := c.GetPlayer(0)
	p.Reset()
	p.AddCard(ctCard(CardDesignClover, 5))
	c.SetCurrentPlayerForTest(0)
	require.NoError(t, c.PlayCard(0, 0))
	require.Equal(t, ChineseTenPhaseSelect, c.GetPhase())

	// Index 2 is the king -- a five cannot take it.
	assert.ErrorContains(t, c.SelectCapture(0, 2), "cannot be captured")
	assert.Error(t, c.SelectCapture(0, 99))
	assert.Error(t, c.SelectCapture(1, 0), "not that player's turn")
}

func TestChineseTen_TheHigherRedTotalWinsAndEqualIsADraw(t *testing.T) {
	c := NewDefaultChineseTen()
	c.Reset()
	require.True(t, ctPlayOut(t, c))
	require.True(t, c.GetGameEndFlag())

	switch w := c.GetWinnerIdx(); w {
	case -1:
		assert.Equal(t, c.GetScore(0), c.GetScore(1), "-1 means the totals matched")
	default:
		assert.Greater(t, c.GetScore(w), c.GetScore((w+1)%ChineseTenPlayerCnt))
	}
}

func TestChineseTen_SurvivesAKVRoundTrip(t *testing.T) {
	c := NewDefaultChineseTen()
	c.Reset()
	for range 3 {
		if c.GetPhase() != ChineseTenPhasePlay {
			break
		}
		_ = c.PlayCard(c.GetCurrentPlayerIdx(), 0)
	}

	data, err := json.Marshal(c)
	require.NoError(t, err)

	restored := NewDefaultChineseTen()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, c.GetStockCount(), restored.GetStockCount())
	assert.Equal(t, len(c.GetLayout()), len(restored.GetLayout()))
	assert.Equal(t, c.GetPhase(), restored.GetPhase())
	assert.Equal(t, c.GetCurrentPlayerIdx(), restored.GetCurrentPlayerIdx())
	for i := range c.GetPlayers() {
		assert.Equal(t, c.GetScore(i), restored.GetScore(i), "score %d", i)
		assert.Equal(t, len(c.GetCaptured(i)), len(restored.GetCaptured(i)), "captured %d", i)
		assert.Equal(t, c.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize(), "hand %d", i)
	}
}

func TestChineseTen_UnmarshalRejectsAndClampsHostileSnapshots(t *testing.T) {
	assert.Error(t, json.Unmarshal([]byte("{"), NewDefaultChineseTen()))
	assert.Error(t, json.Unmarshal([]byte(`{"pl":[]}`), NewDefaultChineseTen()))

	c := NewDefaultChineseTen()
	c.Reset()
	data, err := json.Marshal(c)
	require.NoError(t, err)

	t.Run("invalid config", func(t *testing.T) {
		hostile := replaceJSONNumber(t, string(data), `"cd":0`, `"cd":99`)
		assert.Error(t, json.Unmarshal([]byte(hostile), NewDefaultChineseTen()))
	})

	t.Run("out-of-range seats are clamped", func(t *testing.T) {
		cur := fmt.Sprintf(`"ci":%d`, c.GetCurrentPlayerIdx())
		hostile := replaceJSONNumber(t, string(data), cur, `"ci":99`)
		restored := NewDefaultChineseTen()
		require.NoError(t, json.Unmarshal([]byte(hostile), restored))
		assert.Equal(t, -1, restored.GetCurrentPlayerIdx())
	})
}

func TestChineseTen_Accessors(t *testing.T) {
	c := NewDefaultChineseTen()
	c.Reset()

	assert.Nil(t, c.GetPlayer(-1))
	assert.Nil(t, c.GetPlayer(99))
	assert.Nil(t, c.GetCaptured(99))
	assert.Equal(t, 0, c.GetScore(99))
	assert.Nil(t, c.GetPendingCard())
	assert.Nil(t, c.GetSelectableIndices(), "nothing is selectable while playing")
	assert.NotEmpty(t, c.GetActionLog())

	c.SetScore(99, 10) // out of range: a no-op, not a panic
	assert.Equal(t, 0, c.GetScore(0))
	c.SetScore(0, 42)
	assert.Equal(t, 42, c.GetScore(0))

	cfg := c.GetConfig()
	assert.NoError(t, cfg.Validate())
	c.SetConfig(cfg)
	assert.Equal(t, cfg, c.GetConfig())
}

func TestChineseTen_PlayerSnapshotWithoutAnEmbeddedPlayerStillLoads(t *testing.T) {
	var p ChineseTenPlayer
	require.NoError(t, json.Unmarshal([]byte(`{}`), &p))
	assert.Equal(t, 0, p.GetCardsSize())
	assert.False(t, p.GetIsHuman())
}

func TestChineseTen_CpuNeverProducesAnIllegalMove(t *testing.T) {
	// The CPU's output is fed straight back into the domain, so an off-by-one
	// in its bookkeeping surfaces as a rejected move.
	for range 50 {
		c := NewDefaultChineseTen()
		c.Reset()
		for range 200 {
			if c.GetGameEndFlag() {
				break
			}
			idx := c.GetCurrentPlayerIdx()
			action := c.ChineseTenCpuDecide(idx)
			if c.GetPhase() == ChineseTenPhaseSelect {
				require.GreaterOrEqual(t, action.LayoutIdx, 0, "a selection needs a choice")
				require.NoError(t, c.SelectCapture(idx, action.LayoutIdx))
				continue
			}
			require.GreaterOrEqual(t, action.HandIdx, 0, "a play phase needs a card")
			require.NoError(t, c.PlayCard(idx, action.HandIdx),
				"the CPU proposed a move its own domain rejects")
		}
		require.True(t, c.GetGameEndFlag(), "a CPU-vs-CPU game must terminate")
	}
}

func TestChineseTen_CpuPrefersTheHighestScoringCapture(t *testing.T) {
	// A red king (10 pts) and a black king (0 pts) both takeable: the red one
	// is the whole reason to play the card.
	c := NewDefaultChineseTen()
	c.Reset()
	c.SetLayoutForTest([]*Card{ctCard(CardDesignSpade, 13), ctCard(CardDesignHeart, 13)})
	p := c.GetPlayer(0)
	p.Reset()
	p.AddCard(ctCard(CardDesignClover, 13))
	c.SetCurrentPlayerForTest(0)

	require.NoError(t, c.PlayCard(0, 0))
	require.Equal(t, ChineseTenPhaseSelect, c.GetPhase())

	chosen := c.ChineseTenCpuDecide(0).LayoutIdx
	assert.Equal(t, 10, ChineseTenCardPoints(c.GetLayout()[chosen]),
		"the CPU must take the red king, not the black one")
}

func TestChineseTen_CpuDiscardsItsCheapestCardWhenNothingCaptures(t *testing.T) {
	c := NewDefaultChineseTen()
	c.Reset()
	c.SetLayoutForTest([]*Card{ctCard(CardDesignSpade, 13)})
	p := c.GetPlayer(0)
	p.Reset()
	// Neither captures a king; the black two is worth nothing, the red ace 20.
	p.AddCard(ctCard(CardDesignHeart, 1))
	p.AddCard(ctCard(CardDesignSpade, 2))
	c.SetCurrentPlayerForTest(0)

	chosen := c.ChineseTenCpuDecide(0).HandIdx
	assert.Equal(t, 0, ChineseTenCardPoints(p.GetCard(chosen)),
		"a card it must give away should be the one worth nothing")
}

// ctPlayOut drives a game to its end, always playing the first card and
// resolving any choice with the first legal option. Returns false if it stalls.
func ctPlayOut(t *testing.T, c *ChineseTen) bool {
	t.Helper()
	for range 400 {
		if c.GetGameEndFlag() {
			return true
		}
		if c.GetPhase() == ChineseTenPhaseSelect {
			options := c.GetSelectableIndices()
			require.NotEmpty(t, options, "a selection phase must offer an option")
			require.NoError(t, c.SelectCapture(c.GetCurrentPlayerIdx(), options[0]))
			continue
		}
		idx := c.GetCurrentPlayerIdx()
		if idx < 0 || c.GetPlayer(idx).GetCardsSize() == 0 {
			return false
		}
		require.NoError(t, c.PlayCard(idx, 0))
	}
	return false
}
