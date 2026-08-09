//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// logReader is the one thing this test needs from a game.
type logReader interface {
	GetActionLog() []*ActionLogEntry
}

// These 37 games moved their action log from an own `actionLog` field to the
// embedded actionLogBase. Promotion keeps `g.actionLog` compiling in their
// MarshalJSON/UnmarshalJSON pairs, so nothing in them had to change -- which is
// exactly why a mistake here would be quiet. The log is persisted to KV for the
// Cloudflare Workers, so a codec that silently stopped carrying it would drop
// history per request rather than fail a build.
//
// Their existing round-trip tests do not assert on the log (ChineseTen's checks
// stock, layout, phase, scores and hands), so this covers the gap for all 19.
func TestActionLogBaseAdopters_LogSurvivesAKVRoundTrip(t *testing.T) {
	adopters := []struct {
		name  string
		build func() (any, any) // a game holding one entry, and an empty one to restore into
	}{
		{"BidEuchre", func() (any, any) {
			g := NewDefaultBidEuchre()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultBidEuchre()
		}},
		{"Boston", func() (any, any) {
			g := NewDefaultBoston()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultBoston()
		}},
		{"Bura", func() (any, any) {
			g := NewDefaultBura()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultBura()
		}},
		{"ChineseTen", func() (any, any) {
			g := NewDefaultChineseTen()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultChineseTen()
		}},
		{"Desmoche", func() (any, any) {
			g := NewDefaultDesmoche()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultDesmoche()
		}},
		{"Kaiser", func() (any, any) {
			g := NewDefaultKaiser()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultKaiser()
		}},
		{"Kille", func() (any, any) {
			g := NewDefaultKille()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultKille()
		}},
		{"Klaberjass", func() (any, any) {
			g := NewDefaultKlaberjass()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultKlaberjass()
		}},
		{"Loba", func() (any, any) {
			g := NewDefaultLoba()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultLoba()
		}},
		{"Mushi", func() (any, any) {
			g := NewDefaultMushi()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultMushi()
		}},
		{"NainJaune", func() (any, any) {
			g := NewDefaultNainJaune()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultNainJaune()
		}},
		{"Poch", func() (any, any) {
			g := NewDefaultPoch()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultPoch()
		}},
		{"PopeJoan", func() (any, any) {
			g := NewDefaultPopeJoan()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultPopeJoan()
		}},
		{"Sjavs", func() (any, any) {
			g := NewDefaultSjavs()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultSjavs()
		}},
		{"Skitgubbe", func() (any, any) {
			g := NewDefaultSkitgubbe()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultSkitgubbe()
		}},
		{"Toepen", func() (any, any) {
			g := NewDefaultToepen()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultToepen()
		}},
		{"Trex", func() (any, any) {
			g := NewDefaultTrex()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultTrex()
		}},
		{"Vint", func() (any, any) {
			g := NewDefaultVint()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultVint()
		}},
		{"Zwicker", func() (any, any) {
			g := NewDefaultZwicker()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultZwicker()
		}},
		{"Cego", func() (any, any) {
			g := NewDefaultCego()
			g.appendLog(1, "act", "detail", nil)
			return g, NewDefaultCego()
		}},
		{"FrenchTarot", func() (any, any) {
			g := NewDefaultFrenchTarot()
			g.appendLog(1, "act", "detail", nil)
			return g, NewDefaultFrenchTarot()
		}},
		{"Ganjifa", func() (any, any) {
			g := NewDefaultGanjifa()
			// Unlike the others, this game's UnmarshalJSON validates the state and
			// rejects a never-Reset game ("invalid state values in json"), so the
			// fixture has to be a dealt one.
			g.Reset()
			g.appendLog(1, "act", "detail", nil)
			return g, NewDefaultGanjifa()
		}},
		{"Koenigrufen", func() (any, any) {
			g := NewDefaultKoenigrufen()
			g.appendLog(1, "act", "detail", nil)
			return g, NewDefaultKoenigrufen()
		}},
		{"Scarto", func() (any, any) {
			g := NewDefaultScarto()
			g.appendLog(1, "act", "detail", nil)
			return g, NewDefaultScarto()
		}},
		{"Vira", func() (any, any) {
			g := NewDefaultVira()
			// Unlike the others, this game's UnmarshalJSON validates the state and
			// rejects a never-Reset game ("invalid state values in json"), so the
			// fixture has to be a dealt one.
			g.Reset()
			g.appendLog(1, "act", "detail", nil)
			return g, NewDefaultVira()
		}},
		{"Guandan", func() (any, any) {
			g := NewDefaultGuandan()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultGuandan()
		}},
		{"Karnoffel", func() (any, any) {
			g := NewDefaultKarnoffel()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultKarnoffel()
		}},
		{"Literature", func() (any, any) {
			g := NewDefaultLiterature()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultLiterature()
		}},
		{"ShengJi", func() (any, any) {
			g := NewDefaultShengJi()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultShengJi()
		}},
		{"SixBidSolo", func() (any, any) {
			g := NewDefaultSixBidSolo()
			g.addLog(1, "act", "detail", nil)
			return g, NewDefaultSixBidSolo()
		}},
		{"Aluette", func() (any, any) {
			g := NewDefaultAluette()
			g.Reset()
			g.appendLog(1, "act", "detail", nil)
			return g, NewDefaultAluette()
		}},
		{"Minchiate", func() (any, any) {
			g := NewDefaultMinchiate()
			g.Reset()
			g.appendLog(1, "act", "detail", nil)
			return g, NewDefaultMinchiate()
		}},
		{"Tarocchini", func() (any, any) {
			g := NewDefaultTarocchini()
			g.Reset()
			g.appendLog(1, "act", "detail", nil)
			return g, NewDefaultTarocchini()
		}},
		{"BlackHole", func() (any, any) {
			g := NewDefaultBlackHole()
			g.Reset()
			g.appendLog("act", "detail", nil)
			return g, NewDefaultBlackHole()
		}},
		{"DoubleKlondike", func() (any, any) {
			g := NewDefaultDoubleKlondike()
			g.Reset()
			g.appendLog("act", "detail", nil)
			return g, NewDefaultDoubleKlondike()
		}},
		{"LaBelleLucie", func() (any, any) {
			g := NewDefaultLaBelleLucie()
			g.Reset()
			g.appendLog("act", "detail", nil)
			return g, NewDefaultLaBelleLucie()
		}},
		{"SimpleSimon", func() (any, any) {
			g := NewDefaultSimpleSimon()
			g.Reset()
			g.appendLog("act", "detail", nil)
			return g, NewDefaultSimpleSimon()
		}},
	}

	assert.Len(t, adopters, 37, "every game embedding actionLogBase via addLog/appendLog must be listed")

	for _, a := range adopters {
		t.Run(a.name, func(t *testing.T) {
			saved, restored := a.build()

			before := saved.(logReader).GetActionLog()
			require.NotEmpty(t, before, "fixture must actually record an entry")

			data, err := json.Marshal(saved)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(data, restored))

			after := restored.(logReader).GetActionLog()
			require.Len(t, after, len(before), "the action log did not survive the round trip")
			// Compare the entry the fixture appended, which is the last one --
			// some games log during Reset, so it is not always the only one.
			b, a := before[len(before)-1], after[len(after)-1]
			assert.Equal(t, b.TurnNumber, a.TurnNumber)
			assert.Equal(t, b.PlayerIdx, a.PlayerIdx)
			assert.Equal(t, b.ActionType, a.ActionType)
			assert.Equal(t, b.Detail, a.Detail)
		})
	}
}
