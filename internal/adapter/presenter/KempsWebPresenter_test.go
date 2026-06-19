//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func setupKempsTest() *domain.Kemps {
	g := domain.NewDefaultKemps()
	g.Reset()
	return g
}

// kempsSetField mutates JSON-encoded fields on a Kemps game.
func kempsSetField(g *domain.Kemps, fields map[string]any) {
	data, _ := json.Marshal(g)
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(data, &raw)
	for k, v := range fields {
		raw[k], _ = json.Marshal(v)
	}
	newData, _ := json.Marshal(raw)
	_ = json.Unmarshal(newData, g)
}

func TestKempsWebPresenter_Output(t *testing.T) {
	p := new(presenter.KempsWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		g := setupKempsTest()
		result := p.Output(g, nil)
		assert.NotEmpty(t, result)

		var out controller.KempsWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &out))
		assert.Len(t, out.Players, domain.KempsPlayerCnt)
		assert.Len(t, out.Field, domain.KempsFieldSize)
		assert.Len(t, out.TeamScores, domain.KempsTeamCnt)
		assert.Equal(t, -1, out.WinnerTeam)
		assert.False(t, out.GameEndFlag)
		// 人間 (idx 0) のみ手札公開。
		assert.NotEmpty(t, out.Players[0].Hand)
		assert.Empty(t, out.Players[1].Hand)
		assert.True(t, out.Players[0].IsHuman)
		assert.Equal(t, 0, out.Players[0].Team)
		assert.Equal(t, 1, out.Players[1].Team)
	})

	t.Run("error message", func(t *testing.T) {
		g := setupKempsTest()
		var out controller.KempsWebOutput
		_ = json.Unmarshal([]byte(p.Output(g, errors.New("bad"))), &out)
		assert.Equal(t, "error", out.MessageCode)
		assert.Equal(t, "bad", out.Message)
	})

	t.Run("game end human win", func(t *testing.T) {
		g := setupKempsTest()
		kempsSetField(g, map[string]any{"ge": true, "wt": 0, "ph": domain.KempsPhaseGameEnd})
		var out controller.KempsWebOutput
		_ = json.Unmarshal([]byte(p.Output(g, nil)), &out)
		assert.Equal(t, "kemps.result.humanWin", out.MessageCode)
		assert.True(t, out.GameEndFlag)
	})

	t.Run("game end cpu win", func(t *testing.T) {
		g := setupKempsTest()
		kempsSetField(g, map[string]any{"ge": true, "wt": 1, "ph": domain.KempsPhaseGameEnd})
		var out controller.KempsWebOutput
		_ = json.Unmarshal([]byte(p.Output(g, nil)), &out)
		assert.Equal(t, "kemps.result.cpuWin", out.MessageCode)
	})

	t.Run("partner signaling exposed in declare phase", func(t *testing.T) {
		g := setupKempsTest()
		kempsSetField(g, map[string]any{"ph": domain.KempsPhaseDeclare, "fh": 0})
		var out controller.KempsWebOutput
		_ = json.Unmarshal([]byte(p.Output(g, nil)), &out)
		assert.True(t, out.PartnerSignaling)
		assert.False(t, out.OpponentSignaling)
	})

	t.Run("opponent signaling cue in declare phase", func(t *testing.T) {
		g := setupKempsTest()
		kempsSetField(g, map[string]any{"ph": domain.KempsPhaseDeclare, "fh": 1})
		var out controller.KempsWebOutput
		_ = json.Unmarshal([]byte(p.Output(g, nil)), &out)
		assert.False(t, out.PartnerSignaling)
		assert.True(t, out.OpponentSignaling)
	})
}

func TestKempsWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.KempsWebPresenter)
	g := setupKempsTest()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
