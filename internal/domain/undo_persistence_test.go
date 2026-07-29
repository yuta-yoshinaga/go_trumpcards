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

// dealAttempts caps the reshuffles a case may take while looking for a deal it
// can play. A game that needs more than this is not "unlucky" -- it is broken.
const dealAttempts = 100

func TestUndoSurvivesAKVRoundTrip(t *testing.T) {
	cases := []undoRoundTripCase{
		{"SirTommy", func(t *testing.T) (any, func() bool, func() error) {
			g := NewDefaultSirTommy()
			g.Reset()
			require.NoError(t, g.PlayStockToWaste(0))
			return g, g.CanUndo, g.Undo
		}},
		{"Bisley", func(t *testing.T) (any, func() bool, func() error) {
			// Bisley opens with the four Aces up, but the card that continues a
			// foundation still has to be at the head of a column, so a fresh deal
			// is NOT guaranteed to have a move -- an earlier revision of this case
			// called AutoComplete and reddened develop on roughly 1 deal in 40.
			// Deal until there is something to play, then play what the game says.
			g := NewDefaultBisley()
			var hint *BisleyHint
			for range dealAttempts {
				g.Reset()
				if hint = g.GetHint(); hint != nil {
					break
				}
			}
			require.NotNil(t, hint, "no deal in %d had a legal move", dealAttempts)
			switch hint.ToZone {
			case "ace":
				require.NoError(t, g.MoveTableauToAceFoundation(hint.FromCol))
			case "king":
				require.NoError(t, g.MoveTableauToKingFoundation(hint.FromCol))
			default:
				require.NoError(t, g.MoveTableauToTableau(hint.FromCol, hint.ToIdx))
			}
			return g, g.CanUndo, g.Undo
		}},
		{"NapoleonsSquare", func(t *testing.T) (any, func() bool, func() error) {
			g := NewDefaultNapoleonsSquare()
			g.Reset()
			require.NoError(t, g.Draw())
			return g, g.CanUndo, g.Undo
		}},
		{"GrandfathersClock", func(t *testing.T) (any, func() bool, func() error) {
			// Same shuffle hazard as Bisley: the first hint is only a foundation
			// move on most deals, not all of them.
			g := NewDefaultGrandfathersClock()
			var h *GrandfathersClockHint
			for range dealAttempts {
				g.Reset()
				if h = g.GetHint(); h != nil && h.ToZone == "foundation" {
					break
				}
				h = nil
			}
			require.NotNil(t, h, "no deal in %d opened with a foundation move", dealAttempts)
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

// bigCards returns a slice long enough to trip a MaxSliceLen guard.
func bigCards(n int) []*Card {
	out := make([]*Card, n)
	for i := range out {
		out[i] = NewCard(CardDesignSpade, 1, true)
	}
	return out
}

// bigLog returns an action log long enough to trip a MaxSliceLen guard.
func bigLog(n int) []*ActionLogEntry {
	out := make([]*ActionLogEntry, n)
	for i := range out {
		out[i] = &ActionLogEntry{}
	}
	return out
}

// KV holds whatever an earlier version -- or a tampering client -- wrote, and a
// snapshot is just as reachable as the top-level state. Both layers need bounds,
// and the guards added with the History field are only real if something trips
// them (#4478).
func TestUndoPersistenceRespectsMaxSliceLen(t *testing.T) {
	const over = 1001

	cases := []struct {
		name string
		// tooManySnapshots marshals a wire payload whose History is oversized.
		tooManySnapshots func() ([]byte, error)
		// tooLongLog marshals a wire payload whose ActionLog is oversized.
		tooLongLog func() ([]byte, error)
		// bloatedSnapshot marshals a single snapshot with an oversized pile.
		bloatedSnapshot func() ([]byte, error)
		// restore feeds bytes to the game's UnmarshalJSON.
		restore func([]byte) error
		// restoreSnapshot feeds bytes to the snapshot's UnmarshalJSON.
		restoreSnapshot func([]byte) error
	}{
		{
			name: "SirTommy",
			tooManySnapshots: func() ([]byte, error) {
				return json.Marshal(&sirTommyJSON{History: make([]*sirTommySnapshot, over)})
			},
			tooLongLog: func() ([]byte, error) {
				return json.Marshal(&sirTommyJSON{ActionLog: bigLog(over)})
			},
			bloatedSnapshot: func() ([]byte, error) {
				return json.Marshal(&sirTommySnapshotJSON{Stock: bigCards(over)})
			},
			restore:         func(b []byte) error { return NewDefaultSirTommy().UnmarshalJSON(b) },
			restoreSnapshot: func(b []byte) error { return new(sirTommySnapshot).UnmarshalJSON(b) },
		},
		{
			name: "Bisley",
			tooManySnapshots: func() ([]byte, error) {
				return json.Marshal(&bisleyJSON{History: make([]*bisleySnapshot, over)})
			},
			tooLongLog: func() ([]byte, error) {
				return json.Marshal(&bisleyJSON{ActionLog: bigLog(over)})
			},
			bloatedSnapshot: func() ([]byte, error) {
				return json.Marshal(&bisleySnapshotJSON{
					AceFoundations: [BisleyFoundationCnt][]*Card{bigCards(over)},
				})
			},
			restore:         func(b []byte) error { return NewDefaultBisley().UnmarshalJSON(b) },
			restoreSnapshot: func(b []byte) error { return new(bisleySnapshot).UnmarshalJSON(b) },
		},
		{
			name: "NapoleonsSquare",
			tooManySnapshots: func() ([]byte, error) {
				return json.Marshal(&napoleonsSquareJSON{History: make([]*napoleonsSquareSnapshot, over)})
			},
			tooLongLog: func() ([]byte, error) {
				return json.Marshal(&napoleonsSquareJSON{ActionLog: bigLog(over)})
			},
			bloatedSnapshot: func() ([]byte, error) {
				return json.Marshal(&napoleonsSquareSnapshotJSON{Stock: bigCards(over)})
			},
			restore:         func(b []byte) error { return NewDefaultNapoleonsSquare().UnmarshalJSON(b) },
			restoreSnapshot: func(b []byte) error { return new(napoleonsSquareSnapshot).UnmarshalJSON(b) },
		},
		{
			name: "GrandfathersClock",
			tooManySnapshots: func() ([]byte, error) {
				return json.Marshal(&grandfathersClockJSON{History: make([]*grandfathersClockSnapshot, over)})
			},
			tooLongLog: func() ([]byte, error) {
				return json.Marshal(&grandfathersClockJSON{ActionLog: bigLog(over)})
			},
			bloatedSnapshot: func() ([]byte, error) {
				return json.Marshal(&grandfathersClockSnapshotJSON{
					Foundation: [GrandfathersClockFoundationCnt][]*Card{bigCards(over)},
				})
			},
			restore:         func(b []byte) error { return NewDefaultGrandfathersClock().UnmarshalJSON(b) },
			restoreSnapshot: func(b []byte) error { return new(grandfathersClockSnapshot).UnmarshalJSON(b) },
		},
		{
			name: "MissMilligan",
			tooManySnapshots: func() ([]byte, error) {
				return json.Marshal(&missMilliganJSON{History: make([]*missMilliganSnapshot, over)})
			},
			tooLongLog: func() ([]byte, error) {
				return json.Marshal(&missMilliganJSON{ActionLog: bigLog(over)})
			},
			bloatedSnapshot: func() ([]byte, error) {
				return json.Marshal(&missMilliganSnapshotJSON{Stock: bigCards(over)})
			},
			restore:         func(b []byte) error { return NewDefaultMissMilligan().UnmarshalJSON(b) },
			restoreSnapshot: func(b []byte) error { return new(missMilliganSnapshot).UnmarshalJSON(b) },
		},
		{
			name: "Duchess",
			tooManySnapshots: func() ([]byte, error) {
				return json.Marshal(&duchessJSON{History: make([]*duchessSnapshot, over)})
			},
			tooLongLog: func() ([]byte, error) {
				return json.Marshal(&duchessJSON{ActionLog: bigLog(over)})
			},
			bloatedSnapshot: func() ([]byte, error) {
				return json.Marshal(&duchessSnapshotJSON{Stock: bigCards(over)})
			},
			restore:         func(b []byte) error { return NewDefaultDuchess().UnmarshalJSON(b) },
			restoreSnapshot: func(b []byte) error { return new(duchessSnapshot).UnmarshalJSON(b) },
		},
		{
			name: "Windmill",
			tooManySnapshots: func() ([]byte, error) {
				return json.Marshal(&windmillJSON{History: make([]*windmillSnapshot, over)})
			},
			tooLongLog: func() ([]byte, error) {
				return json.Marshal(&windmillJSON{ActionLog: bigLog(over)})
			},
			bloatedSnapshot: func() ([]byte, error) {
				return json.Marshal(&windmillSnapshotJSON{Stock: bigCards(over)})
			},
			restore:         func(b []byte) error { return NewDefaultWindmill().UnmarshalJSON(b) },
			restoreSnapshot: func(b []byte) error { return new(windmillSnapshot).UnmarshalJSON(b) },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for name, build := range map[string]func() ([]byte, error){
				"too many snapshots":     tc.tooManySnapshots,
				"too long an action log": tc.tooLongLog,
			} {
				t.Run(name, func(t *testing.T) {
					data, err := build()
					require.NoError(t, err)
					assert.Error(t, tc.restore(data))
				})
			}

			t.Run("an oversized pile inside a snapshot", func(t *testing.T) {
				data, err := tc.bloatedSnapshot()
				require.NoError(t, err)
				assert.Error(t, tc.restoreSnapshot(data))
			})

			t.Run("malformed snapshot json", func(t *testing.T) {
				assert.Error(t, tc.restoreSnapshot([]byte("not json")))
			})
		})
	}
}
