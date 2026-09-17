//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertJassDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func jassErrorJSON(t *testing.T, mutate func(map[string]any)) []byte {
	t.Helper()
	g := domain.NewDefaultJass()
	var state map[string]any
	raw, err := json.Marshal(g)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, &state))
	mutate(state)
	raw, err = json.Marshal(state)
	require.NoError(t, err)
	return raw
}

func unmarshalJassError(t *testing.T, mutate func(map[string]any)) error {
	t.Helper()
	g := domain.NewDefaultJass()
	return json.Unmarshal(jassErrorJSON(t, mutate), g)
}

func TestJassDomainErrorsHaveMessageCodes(t *testing.T) {
	g := newTestJass()
	g.SetPhase(domain.JassPhaseBidTrump)
	g.SetCurrentPlayerIdx(0)
	assertJassDomainError(t, g.PlayerChooseTrump(0), domain.ErrInvalidPlay, "jass.errInvalidSuit")

	g.SetPhase(domain.JassPhasePlay)
	assertJassDomainError(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "jass.errCardIndexOutOfRange")

	setJassHand(g, 0, []*domain.Card{jcard(domain.CardDesignSpade, 7), jcard(domain.CardDesignHeart, 6)})
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: jcard(domain.CardDesignSpade, 13)}})
	assertJassDomainError(t, g.PlayerPlay(1), domain.ErrInvalidPlay, "jass.errFollowLeadSuit")
}

func TestJassUnmarshalDomainErrors(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
		code   string
	}{
		{"invalid phase", func(s map[string]any) { s["ph"] = -1 }, "jass.errInvalidPhase"},
		{"invalid trump suit", func(s map[string]any) { s["ts"] = 99 }, "jass.errInvalidTrumpSuit"},
		{"invalid player count", func(s map[string]any) { s["pl"] = []any{} }, "jass.errInvalidPlayerCount"},
		{"player nil", func(s map[string]any) { s["pl"].([]any)[0] = nil }, "jass.errPlayerNil"},
		{"trick card nil", func(s map[string]any) { s["ct"] = []any{map[string]any{"pi": 0, "c": nil}} }, "jass.errTrickCardNil"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertJassDomainError(t, unmarshalJassError(t, tt.mutate), domain.ErrInvalidPlay, tt.code)
		})
	}
}
