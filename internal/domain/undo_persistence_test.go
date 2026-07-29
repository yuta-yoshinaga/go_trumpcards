//go:build test

package domain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The Cloudflare Workers are stateless per request: every call rebuilds the game
// from KV through UnmarshalJSON. A game whose undo stack does not round-trip
// therefore reports CanUndo() == false from the second request onwards, so the
// undo button and the stalemate-escape button silently do nothing in production
// while working perfectly in the CLI and the local server. See #4478.
//
// Asserting on the restored field counts is NOT enough: a snapshot struct with
// unexported fields marshals to `{}`, which preserves the undo *depth* while
// losing every position, so Undo would wipe the board instead of rewinding it.
// Each case below therefore performs a real move, round-trips, and then undoes.

type undoRoundTripCase struct {
	name string
	// setup returns a game that has made at least one undoable move, plus the
	// funcs needed to round-trip and check it.
	play func(t *testing.T) (game any, canUndo func() bool, undo func() error)
}

func TestUndoSurvivesAKVRoundTrip(t *testing.T) {
	cases := []undoRoundTripCase{
		{"SirTommy", func(t *testing.T) (any, func() bool, func() error) {
			g := NewDefaultSirTommy()
			g.Reset()
			require.NoError(t, g.PlayStockToWaste(0))
			return g, g.CanUndo, g.Undo
		}},
		{"Bisley", func(t *testing.T) (any, func() bool, func() error) {
			g := NewDefaultBisley()
			g.Reset()
			// Bisley opens with the four Aces already up, so a foundation move
			// always exists on a fresh deal.
			require.NoError(t, g.AutoComplete())
			return g, g.CanUndo, g.Undo
		}},
		{"NapoleonsSquare", func(t *testing.T) (any, func() bool, func() error) {
			g := NewDefaultNapoleonsSquare()
			g.Reset()
			require.NoError(t, g.Draw())
			return g, g.CanUndo, g.Undo
		}},
		{"GrandfathersClock", func(t *testing.T) (any, func() bool, func() error) {
			g := NewDefaultGrandfathersClock()
			g.Reset()
			h := g.GetHint()
			require.NotNil(t, h, "a fresh clock always has a move")
			require.Equal(t, "foundation", h.ToZone, "the first hint sends a card up")
			require.NoError(t, g.MoveTableauToFoundation(h.FromCol, h.ToIdx))
			return g, g.CanUndo, g.Undo
		}},
		{"MissMilligan", func(t *testing.T) (any, func() bool, func() error) {
			g := NewDefaultMissMilligan()
			g.Reset()
			require.NoError(t, g.Deal())
			return g, g.CanUndo, g.Undo
		}},
		{"Duchess", func(t *testing.T) (any, func() bool, func() error) {
			g := NewDefaultDuchess()
			g.Reset()
			require.NoError(t, g.ChooseBaseRank(0))
			return g, g.CanUndo, g.Undo
		}},
		{"Windmill", func(t *testing.T) (any, func() bool, func() error) {
			g := NewDefaultWindmill()
			g.Reset()
			require.NoError(t, g.Draw())
			return g, g.CanUndo, g.Undo
		}},
		{"AmericanToad", func(t *testing.T) (any, func() bool, func() error) {
			g := NewDefaultAmericanToad()
			g.Reset()
			require.NoError(t, g.Draw())
			return g, g.CanUndo, g.Undo
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			game, canUndo, _ := tc.play(t)
			require.True(t, canUndo(), "the test needs an undoable move to be meaningful")

			data, err := json.Marshal(game)
			require.NoError(t, err)

			// Round-trip into a fresh instance the way the Worker does.
			restored := newEmptyLike(t, tc.name)
			require.NoError(t, json.Unmarshal(data, restored))

			ru, rundo := undoFuncsFor(t, tc.name, restored)
			assert.True(t, ru(), "the undo stack must survive KV")
			assert.NoError(t, rundo(), "and undoing must actually work")
		})
	}
}

// newEmptyLike builds a fresh game of the named type, standing in for the
// Worker's RestoreXInteractor path.
func newEmptyLike(t *testing.T, name string) any {
	t.Helper()
	switch name {
	case "SirTommy":
		return NewDefaultSirTommy()
	case "Bisley":
		return NewDefaultBisley()
	case "NapoleonsSquare":
		return NewDefaultNapoleonsSquare()
	case "GrandfathersClock":
		return NewDefaultGrandfathersClock()
	case "MissMilligan":
		return NewDefaultMissMilligan()
	case "Duchess":
		return NewDefaultDuchess()
	case "Windmill":
		return NewDefaultWindmill()
	case "AmericanToad":
		return NewDefaultAmericanToad()
	}
	t.Fatalf("unknown game %q", name)
	return nil
}

func undoFuncsFor(t *testing.T, name string, g any) (func() bool, func() error) {
	t.Helper()
	type undoable interface {
		CanUndo() bool
		Undo() error
	}
	u, ok := g.(undoable)
	require.True(t, ok, "%s must expose CanUndo/Undo", name)
	return u.CanUndo, u.Undo
}

// historyFieldRe finds a domain type that keeps an undo stack.
var historyFieldRe = regexp.MustCompile(`(?m)^\s+history\s+\[\]\*(\w+)`)

// persistedHistoryRe reports whether a snapshot type reaches the wire. Struct
// fields are gofmt-aligned, so the gap before the tag is arbitrary whitespace --
// matching on a single space silently marks 35 correct games as broken.
func persistedHistoryRe(snapshot string) *regexp.Regexp {
	return regexp.MustCompile(`\[\]\*` + regexp.QuoteMeta(snapshot) + `\s+` + "`" + `json:`)
}

// knownUnpersistedHistories are games that predate #4478 and still lose their
// undo stack across a KV round trip. The list is deliberately explicit and
// finite: it exists so the guard below fails for anything NEW, and it must only
// ever shrink. Do not add to it -- fix the game instead.
var knownUnpersistedHistories = map[string]string{
	"BlackHole":      "#4478",
	"DoubleKlondike": "#4478",
	"LaBelleLucie":   "#4478",
	"Nertz":          "#4478",
	"RussianBank":    "#4478",
	"SimpleSimon":    "#4478",
}

// A new game that keeps an undo stack must persist it, or its undo button is
// dead in production. Nothing else catches that, because the CLI and the local
// server keep state in-process (#4478).
func TestEveryUndoStackIsPersisted(t *testing.T) {
	files, err := filepath.Glob("*.go")
	require.NoError(t, err)

	sources := map[string]string{}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		b, err := os.ReadFile(f)
		require.NoError(t, err)
		sources[f] = string(b)
	}

	var offenders []string
	for f, src := range sources {
		m := historyFieldRe.FindStringSubmatch(src)
		if m == nil {
			continue
		}
		game := strings.TrimSuffix(filepath.Base(f), ".go")
		snapshot := m[1]

		// The snapshot type must appear in some wire struct with a json tag.
		persisted := false
		want := persistedHistoryRe(snapshot)
		for _, other := range sources {
			if want.MatchString(other) {
				persisted = true
				break
			}
		}
		if persisted {
			assert.NotContains(t, knownUnpersistedHistories, game,
				"%s now persists its history -- remove it from knownUnpersistedHistories", game)
			continue
		}
		if _, known := knownUnpersistedHistories[game]; known {
			continue
		}
		offenders = append(offenders, game)
	}

	assert.Empty(t, offenders,
		"these games keep an undo stack that never reaches KV, so Undo/UndoN and the "+
			"stalemate-escape button do nothing on the Cloudflare Workers. Give the "+
			"snapshot type its own MarshalJSON/UnmarshalJSON and add a History field "+
			"with a json tag to the wire struct -- see AmericanToad or Canfield, and #4478")
}
