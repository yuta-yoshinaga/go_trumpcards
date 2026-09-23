package presenter_test

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestPresidentWebPresenter_Output(t *testing.T) {
	p := new(presenter.PresidentWebPresenter)

	t.Run("initial state serialises", func(t *testing.T) {
		players := makePresidentPlayersForPresenter()
		pg := domain.NewPresident(domain.NewTrumpCards(0), players, domain.DefaultPresidentConfig())
		for i := 0; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignHeart, 5+i, false))
		}
		raw := p.Output(pg, nil)
		var out controller.PresidentWebOutput
		require.NoError(t, json.Unmarshal([]byte(raw), &out))
		assert.Len(t, out.Players, 4)
		assert.Equal(t, 0, out.CurrentTurn)
		assert.Equal(t, -1, out.LastPlayPlayerIdx)
		assert.True(t, out.Config.RevolutionEnabled)
	})

	t.Run("error message is propagated", func(t *testing.T) {
		players := makePresidentPlayersForPresenter()
		pg := domain.NewPresident(domain.NewTrumpCards(0), players, domain.DefaultPresidentConfig())
		raw := p.Output(pg, errors.New("test-err"))
		var out controller.PresidentWebOutput
		require.NoError(t, json.Unmarshal([]byte(raw), &out))
		assert.Equal(t, "test-err", out.Message)
	})
}

func TestPresidentWebPresenter_ActionLog(t *testing.T) {
	p := new(presenter.PresidentWebPresenter)
	players := makePresidentPlayersForPresenter()
	pg := domain.NewPresident(domain.NewTrumpCards(0), players, domain.DefaultPresidentConfig())
	out := p.ActionLogOutput(pg)
	assert.NotEmpty(t, out)
}

func TestPresidentWebPresenter_HintOutput(t *testing.T) {
	// Web hints are client-side, so HintOutput mirrors Output.
	p := new(presenter.PresidentWebPresenter)
	players := makePresidentPlayersForPresenter()
	pg := domain.NewPresident(domain.NewTrumpCards(0), players, domain.DefaultPresidentConfig())
	assert.Equal(t, p.Output(pg, nil), p.HintOutput(pg))
}

func TestPresidentWebPresenter_ClubThreeStarterMessage(t *testing.T) {
	p := new(presenter.PresidentWebPresenter)

	for _, tt := range []struct {
		name       string
		starterIdx int
		message    string
		code       string
	}{
		// **文言はロケール側。** Message は空で、messageCode だけが席を分ける。
		{"human starter", 0, "", "president.start.clubThreeHuman"},
		{"CPU starter", 2, "", "president.start.clubThreeCPU"},
		{"no club-3 starter", -1, "", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			pg := presidentForStarterMessage(t, tt.starterIdx)
			var out controller.PresidentWebOutput
			require.NoError(t, json.Unmarshal([]byte(p.Output(pg, nil)), &out))
			assert.Equal(t, tt.message, out.Message)
			assert.Equal(t, tt.code, out.MessageCode)
			// 席が params に載っていること。載せ忘れると 2 席が同じ文言になる。
			if tt.code == "president.start.clubThreeCPU" {
				assert.Equal(t, strconv.Itoa(tt.starterIdx), out.MessageParams["idx"])
			}
		})
	}
}
