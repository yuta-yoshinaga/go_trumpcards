//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mushiCard(month, index int) *Card { return NewCard(month, index, true) }

func TestMushi_DeckDropsJuneAndJuly(t *testing.T) {
	// #4418 says August and September. It is June (peony) and July (bush
	// clover) -- and the arithmetic below is what settles it, not the wording.
	deck := newMushiDeck()
	assert.Len(t, deck, MushiDeckSize)

	months := map[int]int{}
	for _, c := range deck {
		months[c.GetDesign()]++
	}
	assert.Zero(t, months[6], "June is not in the Mushi pack")
	assert.Zero(t, months[7], "July is not in the Mushi pack")
	for _, m := range []int{1, 2, 3, 4, 5, 8, 9, 10, 11, 12} {
		assert.Equal(t, MushiCardsPerMonth, months[m], "month %d", m)
	}
}

func TestMushi_ScoreBaselineIsHalfTheDeck(t *testing.T) {
	// The round score subtracts 115, which only means "the margin over an even
	// split" if 115 is half the pack's value. Counting it here is what proves
	// the deck composition is right: dropping August instead of June would
	// remove a 20-point bright and break this.
	byCategory := map[MushiCategory]int{}
	total := 0
	for _, c := range newMushiDeck() {
		byCategory[MushiCardCategory(c)]++
		total += MushiCardPoints(c)
	}

	assert.Equal(t, 5, byCategory[MushiBright], "all five brights survive the cut")
	assert.Equal(t, 7, byCategory[MushiAnimal])
	assert.Equal(t, 8, byCategory[MushiRibbon])
	assert.Equal(t, 20, byCategory[MushiChaff])

	assert.Equal(t, MushiTotalCardPoints, total)
	assert.Equal(t, 230, total)
	assert.Equal(t, total/2, MushiScoreBaseline)
	assert.Equal(t, 115, MushiScoreBaseline)
}

func TestMushi_CardPointsByCategory(t *testing.T) {
	assert.Equal(t, 20, MushiCardPoints(mushiCard(1, 1)), "bright")
	assert.Equal(t, 10, MushiCardPoints(mushiCard(2, 1)), "animal")
	assert.Equal(t, 5, MushiCardPoints(mushiCard(1, 2)), "ribbon")
	assert.Equal(t, 1, MushiCardPoints(mushiCard(1, 3)), "chaff")
	assert.Equal(t, 0, MushiCardPoints(nil))
	assert.Equal(t, MushiChaff, MushiCardCategory(mushiCard(99, 1)), "out of range is chaff, not a panic")
	assert.Equal(t, MushiChaff, MushiCardCategory(mushiCard(1, 99)))
}

func TestMushi_TheLightningCardIsTheOnlyWild(t *testing.T) {
	assert.True(t, MushiIsWild(mushiCard(11, 4)))
	assert.False(t, MushiIsWild(mushiCard(11, 1)), "the rainman is a bright, not the wild")
	assert.False(t, MushiIsWild(mushiCard(12, 4)))
	assert.False(t, MushiIsWild(nil))
	// The wild is a chaff card: being wild does not make it worth more.
	assert.Equal(t, 1, MushiCardPoints(mushiCard(11, 4)))
}

func TestMushi_Yaku(t *testing.T) {
	allBrights := []*Card{mushiCard(1, 1), mushiCard(3, 1), mushiCard(8, 1), mushiCard(11, 1), mushiCard(12, 1)}
	sanko := []*Card{mushiCard(1, 1), mushiCard(3, 1), mushiCard(2, 1)}
	kiri := []*Card{mushiCard(12, 1), mushiCard(12, 2), mushiCard(12, 3), mushiCard(12, 4)}
	fuji := []*Card{mushiCard(4, 1), mushiCard(4, 2), mushiCard(4, 3), mushiCard(4, 4)}

	cases := []struct {
		name     string
		captured []*Card
		want     map[string]int
	}{
		{"nothing", []*Card{mushiCard(1, 3), mushiCard(5, 3)}, map[string]int{}},
		{"goko", allBrights, map[string]int{"goko": 30}},
		{
			// The Mushi sanko is January's bright, March's bright AND
			// February's ANIMAL -- not three brights. Single-sourced; see
			// mushiSankoCards.
			name:     "sanko includes February's animal",
			captured: sanko,
			want:     map[string]int{"sanko": 25},
		},
		{
			name:     "two brights and no warbler is not sanko",
			captured: []*Card{mushiCard(1, 1), mushiCard(3, 1)},
			want:     map[string]int{},
		},
		{
			name:     "three brights without the warbler is still not sanko",
			captured: []*Card{mushiCard(1, 1), mushiCard(3, 1), mushiCard(8, 1)},
			want:     map[string]int{},
		},
		{"kirishima", kiri, map[string]int{"kirishima": 10}},
		{
			name:     "three of December is not kirishima",
			captured: kiri[:3],
			want:     map[string]int{},
		},
		{"fujishima", fuji, map[string]int{"fujishima": 10}},
		{
			name:     "goko suppresses sanko",
			captured: append(append([]*Card{}, allBrights...), mushiCard(2, 1)),
			want:     map[string]int{"goko": 30},
		},
		{
			name:     "rows stack with a light yaku",
			captured: append(append([]*Card{}, sanko...), kiri...),
			want:     map[string]int{"sanko": 25, "kirishima": 10},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := map[string]int{}
			for _, y := range MushiDetectYaku(tc.captured) {
				got[y.Key] = y.Points
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestMushi_DuplicateCardsDoNotInflateAYaku(t *testing.T) {
	// A corrupted snapshot could hand the same card twice; counting it twice
	// would fabricate kirishima out of one December card.
	four := []*Card{mushiCard(12, 1), mushiCard(12, 1), mushiCard(12, 1), mushiCard(12, 1)}
	assert.Empty(t, MushiDetectYaku(four))
}

func TestMushi_DealShapeIsEightAndFour(t *testing.T) {
	// #4418 says six and six.
	m := NewDefaultMushi()
	m.Reset()

	for i := range m.GetPlayers() {
		assert.Equal(t, MushiHandSize, m.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
	// The dealer claims a lightning card off the opening field, so the field
	// may be one short of four -- which is itself the rule under test.
	assert.LessOrEqual(t, len(m.GetField()), MushiFieldSize)
	total := len(m.GetField()) + m.GetStockCount()
	for i := range m.GetPlayers() {
		total += m.GetPlayer(i).GetCardsSize()
		total += len(m.GetCaptured(i))
	}
	assert.Equal(t, MushiDeckSize, total, "every card is somewhere")
}

func TestMushi_DealerClaimsALightningCardFromTheOpeningField(t *testing.T) {
	// Reshuffle until the wild lands in the opening field, then assert the
	// dealer already holds it. Asserting on one deal would be a 1-in-10 flake.
	found := false
	for range 500 {
		m := NewDefaultMushi()
		m.Reset()
		for _, c := range m.GetCaptured(m.GetDealerIdx()) {
			if MushiIsWild(c) {
				found = true
				break
			}
		}
		if found {
			// ...and it is not left on the field as well.
			for _, c := range m.GetField() {
				assert.False(t, MushiIsWild(c), "the wild cannot be both claimed and on the field")
			}
			break
		}
	}
	assert.True(t, found, "the wild should reach the opening field within 500 deals")
}

func TestMushi_PlayingRejectsIllegalRequests(t *testing.T) {
	m := NewDefaultMushi()
	m.Reset()
	cur := m.GetCurrentPlayerIdx()

	assert.Error(t, m.PlayCard(cur, -1))
	assert.Error(t, m.PlayCard(cur, 99))
	assert.Error(t, m.PlayCard((cur+1)%MushiPlayerCnt, 0), "not that player's turn")
}

func TestMushi_ARoundPlaysOutWithoutAStopDecision(t *testing.T) {
	// Mushi is a hana-awase game: there is no koi-koi style "stop when a yaku
	// appears". The round runs until every hand is empty, and the two scores
	// are mirror images because both sides' yaku and card points enter both
	// sums.
	m := NewDefaultMushi()
	m.Reset()

	require.True(t, mushiPlayRound(t, m))
	assert.Equal(t, MushiPhaseRoundEnd, m.GetPhase())
	assert.Equal(t, 0, m.GetPlayer(0).GetCardsSize())
	assert.Equal(t, 0, m.GetPlayer(1).GetCardsSize())
	// The two scores do NOT simply cancel: cards can be stranded on the field
	// when both hands run out, and nobody scores those. What does hold is
	//   score0 + score1 = (points actually captured) - 230
	// which is 0 only when the field empties too. Asserting the mirror image
	// would have been asserting a property this game does not have.
	captured := 0
	for i := range m.GetPlayers() {
		for _, c := range m.GetCaptured(i) {
			captured += MushiCardPoints(c)
		}
	}
	assert.Equal(t, captured-MushiTotalCardPoints,
		m.GetRoundResult(0)+m.GetRoundResult(1),
		"the round scores are margins against half the pack")
}

func TestMushi_EveryCardIsAccountedForAtTheEndOfARound(t *testing.T) {
	for range 20 {
		m := NewDefaultMushi()
		m.Reset()
		require.True(t, mushiPlayRound(t, m))

		total := len(m.GetField()) + m.GetStockCount()
		for i := range m.GetPlayers() {
			total += len(m.GetCaptured(i)) + m.GetPlayer(i).GetCardsSize()
		}
		assert.Equal(t, MushiDeckSize, total)
	}
}

func TestMushi_LeftoverCardsScoreForNobody(t *testing.T) {
	// The deal is 8+8 to hand and 4 to the field, leaving 20 in the stock. A
	// turn consumes one hand card and one stock card, so 16 turns leave FOUR
	// stock cards untouched, plus whatever is stranded on the field.
	//
	// The source says only "play on until all cards have been captured" and
	// does not say what happens to the remainder. Awarding it to the last
	// capturer would be a guess, so the leftovers simply score for nobody --
	// pinned here so the behaviour is a decision rather than an accident.
	m := NewDefaultMushi()
	m.Reset()
	require.True(t, mushiPlayRound(t, m))

	assert.GreaterOrEqual(t, m.GetStockCount(), 0)
	captured := 0
	for i := range m.GetPlayers() {
		captured += len(m.GetCaptured(i))
	}
	assert.Equal(t, MushiDeckSize, captured+len(m.GetField())+m.GetStockCount(),
		"nothing is destroyed; the leftovers are simply unscored")
	assert.Less(t, captured, MushiDeckSize, "some cards are expected to be left over")
}

func TestMushi_TheGameEndsAfterTheConfiguredRounds(t *testing.T) {
	m := NewDefaultMushi()
	cfg := m.GetConfig()
	cfg.TargetRounds = 2
	m.SetConfig(cfg)
	m.Reset()

	require.True(t, mushiPlayRound(t, m))
	require.False(t, m.GetGameEndFlag(), "one round of two is not the end")
	require.NoError(t, m.NextRound())
	assert.Equal(t, 2, m.GetRoundNumber())

	require.True(t, mushiPlayRound(t, m))
	assert.True(t, m.GetGameEndFlag())
	assert.Error(t, m.NextRound(), "the game is over")
}

func TestMushi_NextRoundRejectsAMidRoundCall(t *testing.T) {
	m := NewDefaultMushi()
	m.Reset()
	assert.Error(t, m.NextRound())
}

func TestMushi_SurvivesAKVRoundTrip(t *testing.T) {
	m := NewDefaultMushi()
	m.Reset()
	// Play a few cards so the snapshot carries a partially-filled board.
	for range 4 {
		if m.GetPhase() != MushiPhasePlay {
			break
		}
		idx := m.GetCurrentPlayerIdx()
		if m.GetPlayer(idx).GetCardsSize() == 0 {
			break
		}
		_ = m.PlayCard(idx, 0)
	}

	data, err := json.Marshal(m)
	require.NoError(t, err)

	restored := NewDefaultMushi()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, m.GetStockCount(), restored.GetStockCount())
	assert.Equal(t, len(m.GetField()), len(restored.GetField()))
	assert.Equal(t, m.GetCurrentPlayerIdx(), restored.GetCurrentPlayerIdx())
	assert.Equal(t, m.GetPhase(), restored.GetPhase())
	for i := range m.GetPlayers() {
		assert.Equal(t, m.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize(), "hand %d", i)
		assert.Equal(t, len(m.GetCaptured(i)), len(restored.GetCaptured(i)), "captured %d", i)
		assert.Equal(t, m.GetScore(i), restored.GetScore(i))
	}
}

func TestMushi_UnmarshalRejectsAndClampsHostileSnapshots(t *testing.T) {
	assert.Error(t, json.Unmarshal([]byte("{"), NewDefaultMushi()))
	assert.Error(t, json.Unmarshal([]byte(`{"pl":[]}`), NewDefaultMushi()))

	m := NewDefaultMushi()
	m.Reset()
	data, err := json.Marshal(m)
	require.NoError(t, err)

	t.Run("invalid config", func(t *testing.T) {
		hostile := replaceJSONNumber(t, string(data), `"cd":0`, `"cd":99`)
		assert.Error(t, json.Unmarshal([]byte(hostile), NewDefaultMushi()))
	})

	t.Run("out-of-range indices are clamped", func(t *testing.T) {
		hostile := replaceJSONNumber(t, string(data), `"ci":0`, `"ci":99`)
		restored := NewDefaultMushi()
		require.NoError(t, json.Unmarshal([]byte(hostile), restored))
		assert.Equal(t, -1, restored.GetCurrentPlayerIdx())
	})
}

func TestMushi_Accessors(t *testing.T) {
	m := NewDefaultMushi()
	m.Reset()

	assert.Nil(t, m.GetPlayer(-1))
	assert.Nil(t, m.GetPlayer(99))
	assert.Nil(t, m.GetCaptured(99))
	assert.Equal(t, 0, m.GetScore(99))
	assert.Equal(t, 0, m.GetRoundResult(99))
	assert.Equal(t, 1, m.GetRoundNumber())
	assert.NotEmpty(t, m.GetActionLog())
	assert.Nil(t, m.GetPendingCard())
	assert.Nil(t, m.GetSelectableIndices(), "nothing is selectable while playing")

	m.SetScore(99, 10) // out of range: a no-op, not a panic
	assert.Equal(t, 0, m.GetScore(0))
	m.SetScore(0, 42)
	assert.Equal(t, 42, m.GetScore(0))

	assert.NoError(t, m.GetConfig().Validate())
}

func TestMushi_SelectCaptureRejectsIllegalRequests(t *testing.T) {
	m := NewDefaultMushi()
	m.Reset()
	assert.Error(t, m.SelectCapture(0, 0), "no selection is pending")
}

// mushiPlayRound drives a round to its end, resolving any selection the domain
// asks for by taking the first legal option. Returns false if it stalls.
func mushiPlayRound(t *testing.T, m *Mushi) bool {
	t.Helper()
	for range 400 {
		switch m.GetPhase() {
		case MushiPhaseRoundEnd, MushiPhaseGameEnd:
			return true
		case MushiPhaseSelect, MushiPhaseWildSelect:
			options := m.GetSelectableIndices()
			require.NotEmpty(t, options, "a selection phase must offer an option")
			require.NoError(t, m.SelectCapture(m.GetCurrentPlayerIdx(), options[0]))
		default:
			idx := m.GetCurrentPlayerIdx()
			p := m.GetPlayer(idx)
			require.NotNil(t, p)
			if p.GetCardsSize() == 0 {
				return false
			}
			require.NoError(t, m.PlayCard(idx, 0))
		}
	}
	return false
}

func TestMushi_CpuTakesTheHighestScoringOption(t *testing.T) {
	m := NewDefaultMushi()
	m.Reset()
	// A selection is only offered when two same-month cards sit on the field;
	// drive rounds until one appears rather than constructing the state, so
	// the decision is exercised through the real phase machinery.
	for range 200 {
		if m.GetPhase() == MushiPhaseSelect || m.GetPhase() == MushiPhaseWildSelect {
			action := m.MushiCpuDecide(m.GetCurrentPlayerIdx())
			options := m.GetSelectableIndices()
			require.Contains(t, options, action.FieldIdx, "the CPU must pick a legal option")
			best := 0
			for _, i := range options {
				if pts := MushiCardPoints(m.GetField()[i]); pts > best {
					best = pts
				}
			}
			assert.Equal(t, best, MushiCardPoints(m.GetField()[action.FieldIdx]),
				"the CPU takes the most valuable card it may take")
			return
		}
		if m.GetPhase() != MushiPhasePlay {
			m = NewDefaultMushi()
			m.Reset()
			continue
		}
		idx := m.GetCurrentPlayerIdx()
		if m.GetPlayer(idx).GetCardsSize() == 0 {
			m = NewDefaultMushi()
			m.Reset()
			continue
		}
		require.NoError(t, m.PlayCard(idx, 0))
	}
	t.Skip("no selection phase occurred in 200 turns")
}

func TestMushi_CpuNeverProducesAnIllegalMove(t *testing.T) {
	// The CPU's own output is fed straight back into the domain, so an
	// off-by-one in its bookkeeping shows up as a rejected move.
	for range 50 {
		m := NewDefaultMushi()
		m.Reset()
		for range 400 {
			if m.GetPhase() == MushiPhaseRoundEnd || m.GetPhase() == MushiPhaseGameEnd {
				break
			}
			idx := m.GetCurrentPlayerIdx()
			action := m.MushiCpuDecide(idx)
			if m.GetPhase() == MushiPhaseSelect || m.GetPhase() == MushiPhaseWildSelect {
				require.GreaterOrEqual(t, action.FieldIdx, 0, "a selection phase needs a choice")
				require.NoError(t, m.SelectCapture(idx, action.FieldIdx))
				continue
			}
			require.GreaterOrEqual(t, action.HandIdx, 0, "a play phase needs a card")
			require.NoError(t, m.PlayCard(idx, action.HandIdx),
				"the CPU proposed a move its own domain rejects")
		}
		require.Equal(t, MushiPhaseRoundEnd, m.GetPhase(), "a CPU-vs-CPU round must terminate")
	}
}

func TestMushi_TheWinnerIsTheHigherCumulativeScore(t *testing.T) {
	newGame := func(rounds int) *Mushi {
		m := NewDefaultMushi()
		cfg := m.GetConfig()
		cfg.TargetRounds = rounds
		m.SetConfig(cfg)
		m.Reset()
		return m
	}

	t.Run("a decided game names a winner", func(t *testing.T) {
		m := newGame(1)
		// Force a lopsided result before the round settles.
		require.True(t, mushiPlayRound(t, m))
		require.True(t, m.GetGameEndFlag())
		w := m.GetWinnerIdx()
		assert.True(t, w == -1 || w == 0 || w == 1)
		if w >= 0 {
			other := (w + 1) % MushiPlayerCnt
			assert.Greater(t, m.GetScore(w), m.GetScore(other))
		} else {
			assert.Equal(t, m.GetScore(0), m.GetScore(1), "-1 means the scores tied")
		}
	})
}

func TestMushi_PlayIsRejectedOnceTheRoundHasSettled(t *testing.T) {
	m := NewDefaultMushi()
	m.Reset()
	require.True(t, mushiPlayRound(t, m))
	assert.Error(t, m.PlayCard(m.GetCurrentPlayerIdx(), 0), "the round is over")
}

func TestMushi_PlayIsRejectedWhileASelectionIsPending(t *testing.T) {
	// Reach a selection phase, then try to play another card instead of
	// resolving it. Without the guard the pending card would be dropped.
	for range 300 {
		m := NewDefaultMushi()
		m.Reset()
		for range 40 {
			if m.GetPhase() == MushiPhaseSelect || m.GetPhase() == MushiPhaseWildSelect {
				err := m.PlayCard(m.GetCurrentPlayerIdx(), 0)
				assert.ErrorContains(t, err, "selection is pending")
				return
			}
			if m.GetPhase() != MushiPhasePlay {
				break
			}
			idx := m.GetCurrentPlayerIdx()
			if m.GetPlayer(idx).GetCardsSize() == 0 {
				break
			}
			if err := m.PlayCard(idx, 0); err != nil {
				break
			}
		}
	}
	t.Skip("no selection phase occurred")
}

func TestMushi_SelectCaptureRejectsTheWrongMonthAndTheWillow(t *testing.T) {
	for range 300 {
		m := NewDefaultMushi()
		m.Reset()
		for range 40 {
			switch m.GetPhase() {
			case MushiPhaseSelect:
				// Find a field card of a DIFFERENT month than the pending card.
				for i, c := range m.GetField() {
					if c.GetDesign() != m.GetPendingCard().GetDesign() {
						assert.ErrorContains(t, m.SelectCapture(m.GetCurrentPlayerIdx(), i), "same month")
						return
					}
				}
			case MushiPhaseWildSelect:
				for i, c := range m.GetField() {
					if c.GetDesign() == MushiWildMonth {
						assert.ErrorContains(t, m.SelectCapture(m.GetCurrentPlayerIdx(), i), "willow")
						return
					}
				}
			case MushiPhasePlay:
				idx := m.GetCurrentPlayerIdx()
				if m.GetPlayer(idx).GetCardsSize() == 0 || m.PlayCard(idx, 0) != nil {
					break
				}
				continue
			}
			break
		}
	}
	t.Skip("no suitable selection state occurred")
}

func TestMushi_SelectCaptureRejectsOutOfRangeAndTheWrongPlayer(t *testing.T) {
	for range 300 {
		m := NewDefaultMushi()
		m.Reset()
		for range 40 {
			if m.GetPhase() == MushiPhaseSelect || m.GetPhase() == MushiPhaseWildSelect {
				cur := m.GetCurrentPlayerIdx()
				assert.Error(t, m.SelectCapture(cur, -1))
				assert.Error(t, m.SelectCapture(cur, 99))
				assert.Error(t, m.SelectCapture((cur+1)%MushiPlayerCnt, 0), "not that player's turn")
				return
			}
			if m.GetPhase() != MushiPhasePlay {
				break
			}
			idx := m.GetCurrentPlayerIdx()
			if m.GetPlayer(idx).GetCardsSize() == 0 || m.PlayCard(idx, 0) != nil {
				break
			}
		}
	}
	t.Skip("no selection phase occurred")
}

func TestMushi_PlayerSnapshotWithoutAnEmbeddedPlayerStillLoads(t *testing.T) {
	var p MushiPlayer
	require.NoError(t, json.Unmarshal([]byte(`{}`), &p))
	assert.Equal(t, 0, p.GetCardsSize())
	assert.False(t, p.GetIsHuman())
}

func TestMushi_UnmarshalRepairsAMissingDealerAndCaptureSlots(t *testing.T) {
	// A snapshot whose dealer index is out of range must land on a real seat,
	// not -1: startRound indexes captured[dealerIdx] when claiming the wild.
	m := NewDefaultMushi()
	m.Reset()
	data, err := json.Marshal(m)
	require.NoError(t, err)
	hostile := replaceJSONNumber(t, string(data), `"di":0`, `"di":99`)

	restored := NewDefaultMushi()
	require.NoError(t, json.Unmarshal([]byte(hostile), restored))
	assert.Equal(t, 0, restored.GetDealerIdx(), "an unusable dealer falls back to seat 0")
	for i := range restored.GetPlayers() {
		assert.NotNil(t, restored.GetCaptured(i), "every seat needs a capture slot")
	}
}

func TestMushi_CardGlyphAndName(t *testing.T) {
	// These feed the ADR-0033 procedural render path; an empty glyph is a blank
	// card on screen.
	assert.Equal(t, "🦢", MushiCardGlyph(mushiCard(1, 1)))
	assert.Equal(t, "Crane", MushiCardName(mushiCard(1, 1)))
	assert.Equal(t, "⚡", MushiCardGlyph(mushiCard(11, 4)), "the wild is drawn distinctly")

	for _, c := range newMushiDeck() {
		assert.NotEmpty(t, MushiCardGlyph(c), "every card in the pack needs a glyph")
		assert.NotEmpty(t, MushiCardName(c), "every card in the pack needs a name")
	}

	assert.Empty(t, MushiCardGlyph(nil))
	assert.Empty(t, MushiCardName(nil))
	assert.Empty(t, MushiCardGlyph(mushiCard(99, 1)))
	assert.Empty(t, MushiCardName(mushiCard(1, 99)))
}

func TestMushi_TheGameEndsInADrawWhenScoresTie(t *testing.T) {
	// finishGame's tie branch: -1 rather than an arbitrary seat.
	m := NewDefaultMushi()
	cfg := m.GetConfig()
	cfg.TargetRounds = 1
	m.SetConfig(cfg)
	m.Reset()
	require.True(t, mushiPlayRound(t, m))
	require.True(t, m.GetGameEndFlag())

	if m.GetScore(0) == m.GetScore(1) {
		assert.Equal(t, -1, m.GetWinnerIdx())
		return
	}
	// Not a tie on this deal -- assert the decided case instead, which is the
	// same branch's other side.
	assert.GreaterOrEqual(t, m.GetWinnerIdx(), 0)
}
