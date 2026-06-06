package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newBourreAllCpu() *domain.Bourre {
	players := make([]*domain.BourrePlayer, domain.BourrePlayerCnt)
	for i := range players {
		players[i] = domain.NewBourrePlayer(false)
	}
	return domain.NewBourre(domain.NewTrumpCards(0), players, domain.DefaultBourreConfig())
}

func TestBourreWebPresenter_DecidePhase(t *testing.T) {
	bg := newBourreAllCpu()
	bg.Reset()

	p := new(presenter.BourreWebPresenter)
	var resp controller.BourreWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(bg, nil)), &resp))
	assert.Equal(t, "decide", resp.Phase)
	assert.Len(t, resp.Players, domain.BourrePlayerCnt)
	assert.False(t, resp.GameEndFlag)
	assert.Equal(t, domain.BourrePlayerCnt*domain.BourreAnte, resp.Pot)
}

func TestBourreWebPresenter_PlayAndEnd(t *testing.T) {
	bg := newBourreAllCpu()
	bg.Reset()
	p := new(presenter.BourreWebPresenter)

	sawPlay := false
	for i := 0; i < 300000 && !bg.GetGameEndFlag(); i++ {
		if bg.GetPhase() == domain.BourrePhaseRoundEnd {
			bg.NextHand()
			continue
		}
		if bg.GetPhase() == domain.BourrePhasePlay {
			sawPlay = true
		}
		bg.CpuPlay()
	}

	var end controller.BourreWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(bg, nil)), &end))
	assert.Equal(t, "gameEnd", end.Phase)
	assert.True(t, end.GameEndFlag)
	assert.NotEmpty(t, end.Message)
	assert.GreaterOrEqual(t, end.WinnerIdx, 0)
	assert.True(t, sawPlay, "expected at least one play phase over a full game")
}

func TestBourreWebPresenter_Error(t *testing.T) {
	bg := newBourreAllCpu()
	bg.Reset()
	p := new(presenter.BourreWebPresenter)
	var resp controller.BourreWebOutput
	require.NoError(t, json.Unmarshal([]byte(p.Output(bg, errors.New("boom"))), &resp))
	assert.Equal(t, "boom", resp.Message)
}

func TestBourreWebPresenter_ActionLog(t *testing.T) {
	bg := newBourreAllCpu()
	bg.Reset()
	p := new(presenter.BourreWebPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(bg))
}
