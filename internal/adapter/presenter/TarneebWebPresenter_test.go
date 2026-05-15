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
}

func TestTarneebWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.TarneebWebPresenter)
	tn := newTarneebForWebTest()
	tn.SetPhase(domain.TarneebPhaseBid)
	tn.SetBidPlayerIdx(0)
	raw := p.HintOutput(tn)
	assert.Contains(t, raw, "hint")
}

func TestTarneebWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.TarneebWebPresenter)
	tn := newTarneebForWebTest()
	out := p.ActionLogOutput(tn)
	assert.NotEmpty(t, out)
}
