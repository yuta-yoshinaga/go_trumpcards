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

// These 19 games moved their action log from an own `actionLog` field to the
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
	}

	assert.Len(t, adopters, 19, "every game delegating addLog to appendLog must be listed")

	for _, a := range adopters {
		t.Run(a.name, func(t *testing.T) {
			saved, restored := a.build()

			before := saved.(logReader).GetActionLog()
			require.Len(t, before, 1, "fixture must actually record an entry")

			data, err := json.Marshal(saved)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(data, restored))

			after := restored.(logReader).GetActionLog()
			require.Len(t, after, 1, "the action log did not survive the round trip")
			assert.Equal(t, before[0].TurnNumber, after[0].TurnNumber)
			assert.Equal(t, before[0].PlayerIdx, after[0].PlayerIdx)
			assert.Equal(t, before[0].ActionType, after[0].ActionType)
			assert.Equal(t, before[0].Detail, after[0].Detail)
		})
	}
}
