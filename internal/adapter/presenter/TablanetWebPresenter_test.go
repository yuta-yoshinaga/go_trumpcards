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

func TestTablanetWebPresenter_Output(t *testing.T) {
	g := domain.NewDefaultTablanet()
	g.Reset()
	g.SetCurrentTurn(0)

	p := new(presenter.TablanetWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.TablanetPhasePlay), decoded["phase"])
	players, ok := decoded["players"].([]any)
	require.True(t, ok)
	assert.Len(t, players, domain.TablanetPlayerCnt)
	assert.Contains(t, decoded, "tableCards")
	assert.Contains(t, decoded, "playableIndices")
	assert.Contains(t, decoded, "captureOptions")
}

func TestTablanetWebPresenter_Error(t *testing.T) {
	g := domain.NewDefaultTablanet()
	g.Reset()
	p := new(presenter.TablanetWebPresenter)
	out := p.Output(g, errors.New("boom"))
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "boom", decoded["message"])
}

func TestTablanetWebPresenter_GameEnd(t *testing.T) {
	g := domain.NewDefaultTablanet()
	g.Reset()
	g.SetPhase(domain.TablanetPhaseGameEnd)

	p := new(presenter.TablanetWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "tablanet.result.scores", decoded["messageCode"])
	assert.Contains(t, decoded, "lastDealDetail")
}

func TestTablanetWebPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultTablanet()
	g.Reset()
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))

	p := new(presenter.TablanetWebPresenter)
	out := p.HintOutput(g)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, decoded, "hint")
}

func TestTablanetWebPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultTablanet()
	g.Reset()
	p := new(presenter.TablanetWebPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestTablanetWebPresenterOutputCarriesTheHint(t *testing.T) {
	// 既存の HintOutput テストと同じ組み立て。場と手札を固定するので配りに依存しない。
	g := domain.NewDefaultTablanet()
	g.Reset()
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	require.NotNil(t, g.GetHint(), "fixture must actually produce a hint")

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(new(presenter.TablanetWebPresenter).Output(g, nil)), &decoded))
	assert.Contains(t, decoded, "hint", "Output must carry the hint -- the frontend reads state.hint")
	assert.NotEqual(t, "tablanet.hintRequested", decoded["messageCode"])
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestTablanetWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g := domain.NewDefaultTablanet()
	g.Reset()
	g.SetCurrentTurn(0)
	g.SetTableCards([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
	g.GetPlayer(0).Reset()
	g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))

	assert.Contains(t, new(presenter.TablanetWebPresenter).HintOutput(g), "tablanet.hintRequested")
}
