//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func assertBauernschnapsenDomainError(t *testing.T, err error, sentinel error, code string) {
	t.Helper()
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	de, ok := err.(*domain.DomainError)
	require.True(t, ok, "expected DomainError, got %T", err)
	assert.Equal(t, code, de.MessageCode())
}

func bauernschnapsenErrorJSON(t *testing.T, mutate func(map[string]any)) []byte {
	t.Helper()
	g := domain.NewDefaultBauernschnapsen()
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

func unmarshalBauernschnapsenError(t *testing.T, mutate func(map[string]any)) error {
	t.Helper()
	g := domain.NewDefaultBauernschnapsen()
	return json.Unmarshal(bauernschnapsenErrorJSON(t, mutate), g)
}

func TestBauernschnapsenDomainErrorsHaveMessageCodes(t *testing.T) {
	g := domain.NewDefaultBauernschnapsen()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	err := g.DeclareContract(0, domain.BauernschnapsenContract(99), domain.CardDesignSpade)
	code, params := domain.ErrorMessageCode(err)
	require.Error(t, err)
	assert.Equal(t, "bauernschnapsen.errContractUnavailable", code)
	assert.Nil(t, params)

	g.SetPhase(domain.BauernschnapsenPhasePlay)
	err = g.DeclareContract(0, domain.BauernschnapsenContractNone, domain.CardDesignSpade)
	code, _ = domain.ErrorMessageCode(err)
	assert.Equal(t, "bauernschnapsen.errNotContractPhase", code)
}

func TestBauernschnapsenAdditionalDomainErrors(t *testing.T) {
	g := domain.NewDefaultBauernschnapsen()
	g.Reset()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).ResetRound()
	}
	g.SetCurrentPlayerIdx(1)
	assertBauernschnapsenDomainError(t, g.DeclareContract(0, domain.BauernschnapsenContractNone, domain.CardDesignSpade), domain.ErrInvalidPlay, "bauernschnapsen.errNotYourTurn")

	g.SetPhase(domain.BauernschnapsenPhasePlay)
	g.SetCurrentPlayerIdx(0)
	assertBauernschnapsenDomainError(t, g.PlayerPlay(-1), domain.ErrInvalidCard, "bauernschnapsen.errCardIndexOutOfRange")
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 7, false)}})
	assertBauernschnapsenDomainError(t, g.PlayerDeclareMarriage(0), domain.ErrInvalidPlay, "bauernschnapsen.errMarriageLeadOnly")
}

func TestBauernschnapsenUnmarshalDomainErrors(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
		code   string
	}{
		{"state slice", func(s map[string]any) { s["ct"] = make([]any, 5) }, "bauernschnapsen.errStateSliceInvalid"},
		{"player nil", func(s map[string]any) { p := s["pl"].([]any); p[0] = nil }, "bauernschnapsen.errPlayerNil"},
		{"trick card nil", func(s map[string]any) { s["ct"] = []any{map[string]any{"pi": 0, "c": nil}} }, "bauernschnapsen.errTrickCardNil"},
		{"action log entry nil", func(s map[string]any) { s["al"] = []any{nil} }, "bauernschnapsen.errActionLogEntryNil"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertBauernschnapsenDomainError(t, unmarshalBauernschnapsenError(t, tt.mutate), domain.ErrInvalidPlay, tt.code)
		})
	}
}
