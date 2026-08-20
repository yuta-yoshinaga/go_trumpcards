//go:build test

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

func setupEgyptianRatscrewTest() *domain.EgyptianRatscrew {
	g := domain.NewDefaultEgyptianRatscrew()
	g.Reset()
	return g
}

func TestEgyptianRatscrewWebPresenter_Output(t *testing.T) {
	p := new(presenter.EgyptianRatscrewWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		g := setupEgyptianRatscrewTest()
		result := p.Output(g, nil)
		assert.NotEmpty(t, result)

		var out controller.EgyptianRatscrewWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &out))
		assert.Len(t, out.Players, 2)
		assert.Equal(t, 26, out.Players[0].StockSize)
		assert.Equal(t, 26, out.Players[1].StockSize)
		assert.False(t, out.GameEndFlag)
		assert.Equal(t, -1, out.WinnerIdx)
		assert.True(t, out.IsHumanTurn)
		assert.Empty(t, out.MessageCode)
		assert.Equal(t, 0, out.ChanceRemaining)
		assert.Equal(t, -1, out.ChanceFromIdx)
		assert.False(t, out.IsSlappable)
		assert.False(t, out.IsTopFaceCard)
	})

	t.Run("error message", func(t *testing.T) {
		g := setupEgyptianRatscrewTest()
		result := p.Output(g, errors.New("bad"))
		var out controller.EgyptianRatscrewWebOutput
		_ = json.Unmarshal([]byte(result), &out)
		assert.Equal(t, "error", out.MessageCode)
		assert.Equal(t, "bad", out.Message)
	})

	t.Run("game end human win", func(t *testing.T) {
		g := setupEgyptianRatscrewTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(0)
		raw["ph"], _ = json.Marshal(domain.EgyptianRatscrewPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)

		result := p.Output(g, nil)
		var out controller.EgyptianRatscrewWebOutput
		_ = json.Unmarshal([]byte(result), &out)
		assert.Equal(t, "egyptianratscrew.result.humanWin", out.MessageCode)
		assert.True(t, out.GameEndFlag)
	})

	t.Run("game end CPU win", func(t *testing.T) {
		g := setupEgyptianRatscrewTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(1)
		raw["ph"], _ = json.Marshal(domain.EgyptianRatscrewPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)

		result := p.Output(g, nil)
		var out controller.EgyptianRatscrewWebOutput
		_ = json.Unmarshal([]byte(result), &out)
		assert.Equal(t, "egyptianratscrew.result.cpuWin", out.MessageCode)
	})
}

func TestEgyptianRatscrewWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.EgyptianRatscrewWebPresenter)
	g := setupEgyptianRatscrewTest()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// #5580: 規則の説明に使う回数はドメインの FaceCardChances から渡すこと。
// 直書きすると、回数を変えたとき説明だけが嘘になる。
func TestEgyptianRatscrewWebPresenter_ShipsTheFaceCardChances(t *testing.T) {
	g := domain.NewDefaultEgyptianRatscrew()
	g.Reset()

	var out controller.EgyptianRatscrewWebOutput
	require.NoError(t, json.Unmarshal([]byte(new(presenter.EgyptianRatscrewWebPresenter).Output(g, nil)), &out))
	require.NotNil(t, out.FaceChances)
	assert.Equal(t, domain.FaceCardChances(domain.EgyptianRatscrewJackValue), out.FaceChances.Jack)
	assert.Equal(t, domain.FaceCardChances(domain.EgyptianRatscrewQueenValue), out.FaceChances.Queen)
	assert.Equal(t, domain.FaceCardChances(domain.EgyptianRatscrewKingValue), out.FaceChances.King)
	assert.Equal(t, domain.FaceCardChances(domain.EgyptianRatscrewAceValue), out.FaceChances.Ace)
	// **4 つが同じ数字ではないこと。**全部 1 を返す実装でも上の検査は通りうる。
	assert.NotEqual(t, out.FaceChances.Jack, out.FaceChances.Ace)
}
