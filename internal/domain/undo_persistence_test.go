//go:build test

package domain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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
		// The six below predate #4478 and were fixed last. None of them has a
		// move that is legal on every deal, so each one deals until its own hint
		// appears and then plays exactly what the hint says.
		{"BlackHole", func(t *testing.T) (any, func() bool, func() error) {
			g := NewDefaultBlackHole()
			var h *BlackHoleHint
			for range dealAttempts {
				g.Reset()
				if h = g.GetHint(); h != nil {
					break
				}
			}
			require.NotNil(t, h, "no deal in %d had a legal move", dealAttempts)
			require.NoError(t, g.MoveFanToBlackHole(h.Fan))
			return g, g.CanUndo, g.Undo
		}},
		{"SimpleSimon", func(t *testing.T) (any, func() bool, func() error) {
			g := NewDefaultSimpleSimon()
			var h *SimpleSimonHint
			for range dealAttempts {
				g.Reset()
				if h = g.GetHint(); h != nil {
					break
				}
			}
			require.NotNil(t, h, "no deal in %d had a legal move", dealAttempts)
			require.NoError(t, g.MoveSequence(h.FromCol, h.CardIndex, h.ToCol))
			return g, g.CanUndo, g.Undo
		}},
		{"LaBelleLucie", func(t *testing.T) (any, func() bool, func() error) {
			g := NewDefaultLaBelleLucie()
			var h *LaBelleLucieHint
			for range dealAttempts {
				g.Reset()
				if h = g.GetHint(); h != nil {
					break
				}
			}
			require.NotNil(t, h, "no deal in %d had a legal move", dealAttempts)
			if h.ToFoundation {
				require.NoError(t, g.MoveFanToFoundation(h.FromFan))
			} else {
				require.NoError(t, g.MoveFanToFan(h.FromFan, h.ToFan))
			}
			return g, g.CanUndo, g.Undo
		}},
		{"DoubleKlondike", func(t *testing.T) (any, func() bool, func() error) {
			g := NewDefaultDoubleKlondike()
			g.Reset()
			require.NoError(t, g.Draw())
			return g, g.CanUndo, g.Undo
		}},
		{"Nertz", func(t *testing.T) (any, func() bool, func() error) {
			g := NewDefaultNertz()
			g.Reset()
			require.NoError(t, g.DrawStock(0))
			return g, g.CanUndo, g.Undo
		}},
		{"RussianBank", func(t *testing.T) (any, func() bool, func() error) {
			g := NewDefaultRussianBank()
			var h *RussianBankHint
			for range dealAttempts {
				g.Reset()
				if h = g.GetHint(); h != nil {
					break
				}
			}
			require.NotNil(t, h, "no deal in %d had a legal move", dealAttempts)
			src := RussianBankSource{Zone: h.Zone, FromOpponent: h.FromOpponent, Col: h.Col}
			if h.ToFoundation {
				require.NoError(t, g.MoveToFoundation(src))
			} else {
				require.NoError(t, g.MoveToTableau(src, h.ToCol))
			}
			return g, g.CanUndo, g.Undo
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			game, canUndo, undo := tc.play(t)
			require.True(t, canUndo(), "the test needs an undoable move to be meaningful")

			data, err := json.Marshal(game)
			require.NoError(t, err)

			// The reference board is the in-process game undone directly: its
			// history never went near JSON, so it is what a correct KV round trip
			// has to reproduce.
			require.NoError(t, undo())
			want := boardFingerprint(t, game)

			// Round-trip into a fresh instance the way the Worker does.
			restored := newEmptyLike(t, tc.name)
			require.NoError(t, json.Unmarshal(data, restored))

			ru, rundo := undoFuncsFor(t, tc.name, restored)
			require.True(t, ru(), "the undo stack must survive KV")
			require.NoError(t, rundo(), "and undoing must actually work")

			// Depth alone is not enough. A snapshot type with no codec of its own
			// serialises as `{}`, which keeps CanUndo true and lets Undo return
			// nil while restoring a BLANK board -- the undo wipes the game instead
			// of rewinding it. Only comparing the board catches that.
			assert.Equal(t, want, boardFingerprint(t, restored),
				"undo after a KV round trip must restore the pre-move board, not blank it")
		})
	}
}

// boardFingerprint is a game's serialised state with the two fields that are
// expected to differ stripped: the action log (Undo rewinds the board, not the
// transcript) and the history itself.
func boardFingerprint(t *testing.T, game any) string {
	t.Helper()
	data, err := json.Marshal(game)
	require.NoError(t, err)
	var m map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &m))
	delete(m, "al")
	delete(m, "hi")
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte(':')
		b.Write(m[k])
		b.WriteByte('\n')
	}
	return b.String()
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
	case "BlackHole":
		return NewDefaultBlackHole()
	case "SimpleSimon":
		return NewDefaultSimpleSimon()
	case "LaBelleLucie":
		return NewDefaultLaBelleLucie()
	case "DoubleKlondike":
		return NewDefaultDoubleKlondike()
	case "Nertz":
		return NewDefaultNertz()
	case "RussianBank":
		return NewDefaultRussianBank()
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

// knownUnpersistedHistories is the exemption list the guard consults. It is
// EMPTY, and #4478 is closed -- every domain that keeps an undo stack now sends
// it to KV. Adding an entry here re-opens that bug for one game, so don't:
// give the snapshot type a MarshalJSON/UnmarshalJSON pair instead. The map is
// kept rather than deleted so the guard keeps reporting a *named* regression
// ("X now persists...") if someone does add one.
var knownUnpersistedHistories = map[string]string{}

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

// bigDocument returns a JSON array literal longer than n bytes, for the two
// games whose snapshot holds a document and is bounded by size rather than by
// element count.
func bigDocument(n int) json.RawMessage {
	body := strings.Repeat("0,", n/2+2)
	return json.RawMessage("[" + body + "0]")
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
	// bigOver clears the 10000-element caps the older games use.
	const bigOver = 10001

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
		// The six below cap at 10000 rather than 1000, so they need their own
		// oversize figure. Nertz and RussianBank bound their snapshots by BYTES,
		// not by element count, because the snapshot holds a JSON document.
		{
			name: "BlackHole",
			tooManySnapshots: func() ([]byte, error) {
				return json.Marshal(&blackHoleJSON{History: make([]*blackHoleSnapshot, bigOver)})
			},
			tooLongLog: func() ([]byte, error) {
				return json.Marshal(&blackHoleJSON{ActionLog: bigLog(bigOver)})
			},
			bloatedSnapshot: func() ([]byte, error) {
				return json.Marshal(&blackHoleSnapshotJSON{BlackHole: bigCards(bigOver)})
			},
			restore:         func(b []byte) error { return NewDefaultBlackHole().UnmarshalJSON(b) },
			restoreSnapshot: func(b []byte) error { return new(blackHoleSnapshot).UnmarshalJSON(b) },
		},
		{
			name: "SimpleSimon",
			tooManySnapshots: func() ([]byte, error) {
				return json.Marshal(&simpleSimonJSON{History: make([]*simpleSimonSnapshot, bigOver)})
			},
			tooLongLog: func() ([]byte, error) {
				return json.Marshal(&simpleSimonJSON{ActionLog: bigLog(bigOver)})
			},
			bloatedSnapshot: func() ([]byte, error) {
				return json.Marshal(&simpleSimonSnapshotJSON{
					Columns: [SimpleSimonColCnt][]*Card{bigCards(bigOver)},
				})
			},
			restore:         func(b []byte) error { return NewDefaultSimpleSimon().UnmarshalJSON(b) },
			restoreSnapshot: func(b []byte) error { return new(simpleSimonSnapshot).UnmarshalJSON(b) },
		},
		{
			name: "LaBelleLucie",
			tooManySnapshots: func() ([]byte, error) {
				return json.Marshal(&laBelleLucieJSON{History: make([]*laBelleLucieSnapshot, bigOver)})
			},
			tooLongLog: func() ([]byte, error) {
				return json.Marshal(&laBelleLucieJSON{ActionLog: bigLog(bigOver)})
			},
			bloatedSnapshot: func() ([]byte, error) {
				return json.Marshal(&laBelleLucieSnapshotJSON{Fans: [][]*Card{bigCards(bigOver)}})
			},
			restore:         func(b []byte) error { return NewDefaultLaBelleLucie().UnmarshalJSON(b) },
			restoreSnapshot: func(b []byte) error { return new(laBelleLucieSnapshot).UnmarshalJSON(b) },
		},
		{
			name: "DoubleKlondike",
			tooManySnapshots: func() ([]byte, error) {
				return json.Marshal(&doubleKlondikeJSON{History: make([]*doubleKlondikeSnapshot, bigOver)})
			},
			tooLongLog: func() ([]byte, error) {
				return json.Marshal(&doubleKlondikeJSON{ActionLog: bigLog(bigOver)})
			},
			bloatedSnapshot: func() ([]byte, error) {
				return json.Marshal(&doubleKlondikeSnapshotJSON{Stock: bigCards(bigOver)})
			},
			restore:         func(b []byte) error { return NewDefaultDoubleKlondike().UnmarshalJSON(b) },
			restoreSnapshot: func(b []byte) error { return new(doubleKlondikeSnapshot).UnmarshalJSON(b) },
		},
		{
			name: "Nertz",
			tooManySnapshots: func() ([]byte, error) {
				return json.Marshal(&nertzJSON{History: make([]*nertzSnapshot, bigOver)})
			},
			tooLongLog: func() ([]byte, error) {
				return json.Marshal(&nertzJSON{ActionLog: bigLog(bigOver)})
			},
			bloatedSnapshot: func() ([]byte, error) {
				return json.Marshal(&nertzSnapshotJSON{Players: bigDocument(nertzSnapshotMaxBytes)})
			},
			restore:         func(b []byte) error { return NewDefaultNertz().UnmarshalJSON(b) },
			restoreSnapshot: func(b []byte) error { return new(nertzSnapshot).UnmarshalJSON(b) },
		},
		{
			name: "RussianBank",
			tooManySnapshots: func() ([]byte, error) {
				return json.Marshal(&russianBankJSON{History: make([]*russianBankSnapshot, bigOver)})
			},
			tooLongLog: func() ([]byte, error) {
				return json.Marshal(&russianBankJSON{ActionLog: bigLog(bigOver)})
			},
			bloatedSnapshot: func() ([]byte, error) {
				return json.Marshal(&russianBankSnapshotJSON{State: bigDocument(russianBankSnapshotMaxBytes)})
			},
			restore:         func(b []byte) error { return NewDefaultRussianBank().UnmarshalJSON(b) },
			restoreSnapshot: func(b []byte) error { return new(russianBankSnapshot).UnmarshalJSON(b) },
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

// TestUndoHistoryGrowsLinearly guards the shape of the undo stack, not just its
// presence.
//
// RussianBank is the only game that snapshots by marshalling a whole wire
// struct rather than copying named fields. When History joined that wire struct
// (#4478), each snapshot started embedding every earlier snapshot, so the
// payload DOUBLED per move: 5.6 KB after move 1, 1.45 MB after move 9. Nothing
// else would have caught it -- the round-trip test plays one move, and the size
// caps only fire once a session is already unrestorable.
//
// A per-move delta is the right assertion because the absolute size depends on
// the deal: exponential growth blows the bound within a handful of moves, while
// linear growth keeps every delta near one board.
func TestUndoHistoryGrowsLinearly(t *testing.T) {
	const maxDeltaBytes = 8 * 1024

	// Deals differ in how long the hint keeps finding a move -- some run dry at
	// move 6 -- so keep dealing until one lasts long enough to be evidence.
	const wantMoves = 5
	g := NewDefaultRussianBank()
	var moves int
	for range dealAttempts {
		g.Reset()
		moves = 0
		prev := len(mustMarshal(t, g))
		for moves < 12 {
			h := g.GetHint()
			if h == nil {
				break
			}
			src := RussianBankSource{Zone: h.Zone, FromOpponent: h.FromOpponent, Col: h.Col}
			var err error
			if h.ToFoundation {
				err = g.MoveToFoundation(src)
			} else {
				err = g.MoveToTableau(src, h.ToCol)
			}
			if err != nil {
				break
			}
			moves++
			size := len(mustMarshal(t, g))
			require.LessOrEqual(t, size-prev, maxDeltaBytes,
				"move %d grew the payload by %d bytes; a snapshot is embedding the history",
				moves, size-prev)
			prev = size
		}
		if moves >= wantMoves {
			return
		}
	}
	t.Fatalf("no deal in %d lasted %d moves; the probe never got to measure",
		dealAttempts, wantMoves)
}

// mustMarshal marshals or fails the test.
func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}
