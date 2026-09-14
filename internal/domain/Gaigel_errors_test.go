//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertGaigelDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func gaigelErrorJSON(t *testing.T, mutate func(map[string]any)) []byte {
	t.Helper()
	g := domain.NewDefaultGaigel()
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

func unmarshalGaigelError(t *testing.T, mutate func(map[string]any)) error {
	t.Helper()
	g := domain.NewDefaultGaigel()
	return json.Unmarshal(gaigelErrorJSON(t, mutate), g)
}

func TestGaigelDomainErrorsHaveMessageCodes(t *testing.T) {
	g := domain.NewDefaultGaigel()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
	err := g.PlayerPlay(-1)
	code, params := domain.ErrorMessageCode(err)
	require.Error(t, err)
	assert.Equal(t, "gaigel.errCardIndexOutOfRange", code)
	assert.Nil(t, params)

	err = g.PlayerDeclareMarriage(0)
	code, _ = domain.ErrorMessageCode(err)
	assert.Equal(t, "gaigel.errMarriageUnavailable", code)
}

func TestGaigelAdditionalDomainErrors(t *testing.T) {
	g := domain.NewDefaultGaigel()
	g.Reset()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).ResetRound()
	}
	g.SetPhase(domain.GaigelPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 7, false)}})
	assertGaigelDomainError(t, g.PlayerDeclareMarriage(0), domain.ErrInvalidPlay, "gaigel.errMarriageLeadOnly")
}

func TestGaigelUnmarshalDomainErrors(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
		code   string
	}{
		{"state slice", func(s map[string]any) { s["ct"] = make([]any, 5) }, "gaigel.errStateSliceInvalid"},
		{"player nil", func(s map[string]any) { p := s["pl"].([]any); p[0] = nil }, "gaigel.errPlayerNil"},
		{"trick card nil", func(s map[string]any) { s["ct"] = []any{map[string]any{"pi": 0, "c": nil}} }, "gaigel.errTrickCardNil"},
		{"action log entry nil", func(s map[string]any) { s["al"] = []any{nil} }, "gaigel.errActionLogEntryNil"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGaigelDomainError(t, unmarshalGaigelError(t, tt.mutate), domain.ErrInvalidPlay, tt.code)
		})
	}
}
