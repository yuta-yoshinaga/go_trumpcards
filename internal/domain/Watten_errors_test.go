//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertWattenDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func wattenErrorJSON(t *testing.T, mutate func(map[string]any)) []byte {
	t.Helper()
	g := domain.NewDefaultWatten()
	g.Reset()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		p.ResetRound()
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 7+i, false))
	}
	var state map[string]any
	raw, err := json.Marshal(g)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, &state))
	mutate(state)
	raw, err = json.Marshal(state)
	require.NoError(t, err)
	return raw
}

func unmarshalWattenError(t *testing.T, mutate func(map[string]any)) error {
	t.Helper()
	g := domain.NewDefaultWatten()
	return json.Unmarshal(wattenErrorJSON(t, mutate), g)
}

func TestWattenDomainErrorsHaveMessageCodes(t *testing.T) {
	g := domain.NewDefaultWatten()
	g.Reset()
	err := g.PlayerDeclare(5, domain.CardDesignSpade)
	code, params := domain.ErrorMessageCode(err)
	require.Error(t, err)
	assert.Equal(t, "watten.errInvalidSchlagRank", code)
	assert.Nil(t, params)

	g.SetPhase(domain.WattenPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetRaiseCountForTest(g.GetConfig().MaxRaises)
	err = g.PlayerRaise()
	code, _ = domain.ErrorMessageCode(err)
	assert.Equal(t, "watten.errCannotRaiseNow", code)
}

func TestWattenAdditionalDomainErrors(t *testing.T) {
	g := domain.NewDefaultWatten()
	g.Reset()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).ResetRound()
	}
	g.SetPhase(domain.WattenPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertWattenDomainError(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "watten.errCardIndexOutOfRange")
}

func TestWattenUnmarshalDomainErrors(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
		code   string
	}{
		{"player nil", func(s map[string]any) { p := s["pl"].([]any); p[0] = nil }, "watten.errPlayerNil"},
		{"player team out of range", func(s map[string]any) { p := s["pl"].([]any); p[0].(map[string]any)["tm"] = 99 }, "watten.errPlayerTeamOutOfRange"},
		{"too many trick cards", func(s map[string]any) { s["ct"] = make([]any, 5) }, "watten.errTooManyTrickCards"},
		{"trick card nil", func(s map[string]any) { s["ct"] = []any{map[string]any{"pi": 0, "c": nil}} }, "watten.errTrickCardNil"},
		{"trick player index out of range", func(s map[string]any) {
			s["ct"] = []any{map[string]any{"pi": 99, "c": map[string]any{"d": 1, "v": 7, "w": false}}}
		}, "watten.errTrickPlayerIndexOutOfRange"},
		{"action log too large", func(s map[string]any) { s["al"] = make([]any, 2001) }, "watten.errActionLogTooLarge"},
		{"current player index out of range", func(s map[string]any) { s["cp"] = 99 }, "watten.errCurrentPlayerIndexOutOfRange"},
		{"player index out of range", func(s map[string]any) { s["li"] = 99 }, "watten.errPlayerIndexOutOfRange"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertWattenDomainError(t, unmarshalWattenError(t, tt.mutate), domain.ErrInvalidPlay, tt.code)
		})
	}
}
