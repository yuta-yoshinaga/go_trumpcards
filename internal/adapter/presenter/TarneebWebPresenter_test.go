//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTarneebForWebTest() *domain.Tarneeb {
	tn := domain.NewDefaultTarneeb()
	tn.Reset()
	return tn
}

type webOutPartial struct {
	Phase       int               `json:"phase"`
	TrumpSuit   int               `json:"trumpSuit"`
	TeamScores  []int             `json:"teamScores"`
	Message     string            `json:"message,omitempty"`
	MessageCode string            `json:"messageCode,omitempty"`
	Players     []json.RawMessage `json:"players"`
	GameEndFlag bool              `json:"gameEndFlag"`
	WinnerTeam  int               `json:"winnerTeam"`
}

func TestTarneebWebPresenter_Output(t *testing.T) {
	p := new(presenter.TarneebWebPresenter)

	t.Run("default bid phase", func(t *testing.T) {
		tn := newTarneebForWebTest()
		raw := p.Output(tn, nil)
		var got webOutPartial
		require.NoError(t, json.Unmarshal([]byte(raw), &got))
		assert.Equal(t, int(domain.TarneebPhaseBid), got.Phase)
		assert.Equal(t, "tarneeb.bidPhase", got.MessageCode)
		assert.Equal(t, []int{0, 0}, got.TeamScores)
		assert.False(t, got.GameEndFlag)
	})

	t.Run("trump phase", func(t *testing.T) {
		tn := newTarneebForWebTest()
		tn.SetPhase(domain.TarneebPhaseTrumpDeclaration)
		var got webOutPartial
		require.NoError(t, json.Unmarshal([]byte(p.Output(tn, nil)), &got))
		assert.Equal(t, "tarneeb.trumpPhase", got.MessageCode)
	})

	t.Run("play phase lead vs follow", func(t *testing.T) {
		tn := newTarneebForWebTest()
		tn.SetPhase(domain.TarneebPhasePlay)
		var got webOutPartial
		require.NoError(t, json.Unmarshal([]byte(p.Output(tn, nil)), &got))
		assert.Equal(t, "tarneeb.playPhase.lead", got.MessageCode)
	})

	t.Run("play phase follow when trick has cards", func(t *testing.T) {
		tn := newTarneebForWebTest()
		tn.SetPhase(domain.TarneebPhasePlay)
		tn.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 7, false)},
		})
		var got webOutPartial
		require.NoError(t, json.Unmarshal([]byte(p.Output(tn, nil)), &got))
		assert.Equal(t, "tarneeb.playPhase.follow", got.MessageCode)
		// 4 players always in a default Tarneeb game.
		assert.Equal(t, 4, len(got.Players))
	})

	t.Run("error returned as message", func(t *testing.T) {
		tn := newTarneebForWebTest()
		raw := p.Output(tn, errors.New("boom"))
		var got webOutPartial
		require.NoError(t, json.Unmarshal([]byte(raw), &got))
		assert.Equal(t, "boom", got.Message)
	})

	t.Run("trick end + round end", func(t *testing.T) {
		tn := newTarneebForWebTest()
		tn.SetPhase(domain.TarneebPhaseTrickEnd)
		var got webOutPartial
		require.NoError(t, json.Unmarshal([]byte(p.Output(tn, nil)), &got))
		assert.Equal(t, "tarneeb.trickEnd", got.MessageCode)

		tn.SetPhase(domain.TarneebPhaseRoundEnd)
		require.NoError(t, json.Unmarshal([]byte(p.Output(tn, nil)), &got))
		assert.Equal(t, "tarneeb.roundEnd", got.MessageCode)
	})

	t.Run("game end human-team win", func(t *testing.T) {
		tn := newTarneebForWebTest()
		tn.SetPhase(domain.TarneebPhaseGameEnd)
		tn.SetTeamScore(0, 31)
		tn.SetBidWinnerIdx(0)
		tn.SetHighestBid(7)
		// human is at idx 0 (team 0). Force GameEnd to be picked up by
		// the presenter's buildMessage by stamping the flag.
		// gameEndFlag is internal — set via ScoreRound triggering checkGameEnd.
		tn.SetPhase(domain.TarneebPhaseRoundEnd)
		// Add some tricks for the human team so the bidder hits its bid.
		for i := 0; i < 4; i++ {
			tn.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
		}
		for i := 0; i < 4; i++ {
			tn.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, false)})
		}
		for i := 0; i < 3; i++ {
			tn.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
		}
		for i := 0; i < 2; i++ {
			tn.GetPlayer(3).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 4, false)})
		}
		tn.ScoreRound()
		require.True(t, tn.GetGameEndFlag())
		var got webOutPartial
		require.NoError(t, json.Unmarshal([]byte(p.Output(tn, nil)), &got))
		assert.Equal(t, "tarneeb.gameEndHumanWin", got.MessageCode)
		assert.True(t, got.GameEndFlag)
		assert.Equal(t, 0, got.WinnerTeam)
	})

	t.Run("game end CPU-team win", func(t *testing.T) {
		tn := newTarneebForWebTest()
		tn.SetPhase(domain.TarneebPhaseRoundEnd)
		tn.SetTeamScore(1, 24)
		tn.SetBidWinnerIdx(1)
		tn.SetHighestBid(7)
		// CPU team (1) takes 8 tricks, hits the bid, reaches 32.
		for i := 0; i < 4; i++ {
			tn.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
		}
		for i := 0; i < 4; i++ {
			tn.GetPlayer(3).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, false)})
		}
		for i := 0; i < 3; i++ {
			tn.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
		}
		for i := 0; i < 2; i++ {
			tn.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 4, false)})
		}
		tn.ScoreRound()
		require.True(t, tn.GetGameEndFlag())
		var got webOutPartial
		require.NoError(t, json.Unmarshal([]byte(p.Output(tn, nil)), &got))
		assert.Equal(t, "tarneeb.gameEndCpuWin", got.MessageCode)
		assert.Equal(t, 1, got.WinnerTeam)
	})
}

func TestTarneebWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.TarneebWebPresenter)
	tn := newTarneebForWebTest()
	tn.SetPhase(domain.TarneebPhaseBid)
	tn.SetBidPlayerIdx(0)
	raw := p.HintOutput(tn)
	assert.Contains(t, raw, "hint")
}

func TestTarneebWebPresenter_HintOutput_Empty(t *testing.T) {
	p := new(presenter.TarneebWebPresenter)
	tn := newTarneebForWebTest()
	tn.SetPhase(domain.TarneebPhasePlay)
	tn.SetCurrentPlayerIdx(1) // not the human's turn
	raw := p.HintOutput(tn)
	// no hint should be emitted but the structure should still be valid JSON
	var got map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(raw), &got))
	_, hasHint := got["hint"]
	assert.False(t, hasHint)
}

func TestTarneebWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TarneebWebPresenter)
	tn := newTarneebForWebTest()
	out := p.ActionLogOutput(tn)
	assert.NotEmpty(t, out)
}
