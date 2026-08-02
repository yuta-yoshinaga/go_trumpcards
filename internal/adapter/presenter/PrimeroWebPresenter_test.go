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

// primeroWebSetHand は player の手札を 4 枚に差し替える (プレゼンターテスト用ヘルパー)。
func primeroWebSetHand(p *domain.PrimeroPlayer, cards ...*domain.Card) {
	p.ClearHand()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// primeroResultGame は決定的な結果フェーズのゲームを組み立てる。
// humanWins=false のとき seat0 (人間) はヌメルス (負け)、seat1 (CPU) はフルクサス (勝ち)。
func primeroResultGame(humanWins bool) *domain.Primero {
	g := domain.NewDefaultPrimero()
	g.SetPhase(domain.PrimeroPhaseBetting)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetFolded(false)
		g.GetPlayer(i).SetOut(false)
	}
	// Numerus (weak): two spades + two hearts, faces only.
	weak := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 11, false), domain.NewCard(domain.CardDesignSpade, 12, false),
		domain.NewCard(domain.CardDesignHeart, 13, false), domain.NewCard(domain.CardDesignHeart, 11, false),
	}
	// Fluxus (strong): four hearts.
	fluxus := []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 7, false), domain.NewCard(domain.CardDesignHeart, 6, false),
		domain.NewCard(domain.CardDesignHeart, 1, false), domain.NewCard(domain.CardDesignHeart, 5, false),
	}
	if humanWins {
		primeroWebSetHand(g.GetPlayer(0), fluxus...)
		primeroWebSetHand(g.GetPlayer(1), weak...)
	} else {
		primeroWebSetHand(g.GetPlayer(0), weak...)
		primeroWebSetHand(g.GetPlayer(1), fluxus...)
	}
	for i := 2; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetFolded(true)
	}
	g.SetPot(100)
	g.ResolveForTest()
	return g
}

func TestPrimeroWebPresenter_OutputBettingPhase(t *testing.T) {
	g := domain.NewDefaultPrimero()
	p := new(presenter.PrimeroWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, decoded, "players")
	assert.Contains(t, decoded, "pot")
	assert.Contains(t, decoded, "config")
	assert.Contains(t, decoded, "currentPlayerIdx")
	// There is no retourne field for Primero.
	assert.NotContains(t, decoded, "retourne")
	if decoded["phase"] == float64(domain.PrimeroPhaseBetting) && !decoded["gameEndFlag"].(bool) {
		assert.Equal(t, "primero.bettingPhase", decoded["messageCode"])
	}
}

func TestPrimeroWebPresenter_Error(t *testing.T) {
	g := domain.NewDefaultPrimero()
	p := new(presenter.PrimeroWebPresenter)
	out := p.Output(g, errors.New("boom"))

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "boom", decoded["message"])
}

func TestPrimeroWebPresenter_ResultHumanLose(t *testing.T) {
	g := primeroResultGame(false)
	p := new(presenter.PrimeroWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.PrimeroPhaseResult), decoded["phase"])
	assert.Equal(t, "primero.roundEndHumanLose", decoded["messageCode"])
	assert.Equal(t, float64(1), decoded["winnerIdx"])

	players, ok := decoded["players"].([]any)
	require.True(t, ok)
	cpu := players[1].(map[string]any)
	assert.Equal(t, "fluxus", cpu["handName"])
	assert.Equal(t, true, cpu["isWinner"])
	// The human's own hand name is populated too (revealed cards).
	human := players[0].(map[string]any)
	assert.Equal(t, "numerus", human["handName"])
}

func TestPrimeroWebPresenter_ResultHumanWin(t *testing.T) {
	g := primeroResultGame(true)
	p := new(presenter.PrimeroWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "primero.roundEndHumanWin", decoded["messageCode"])
	assert.Equal(t, float64(0), decoded["winnerIdx"])
}

func TestPrimeroWebPresenter_GameEnd(t *testing.T) {
	g := domain.NewDefaultPrimero()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	for i := 0; i < 100 && g.GetPhase() == domain.PrimeroPhaseBetting && g.IsHumanTurn(); i++ {
		require.NoError(t, g.PlayerCall())
	}
	require.True(t, g.GetGameEndFlag())

	p := new(presenter.PrimeroWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, true, decoded["gameEndFlag"])
	assert.Contains(t, decoded["messageCode"], "primero.result.")
}

func TestPrimeroWebPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultPrimero()
	p := new(presenter.PrimeroWebPresenter)
	out := p.HintOutput(g)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, decoded, "players")
}

func TestPrimeroWebPresenter_ActionLog(t *testing.T) {
	g := primeroResultGame(false)
	p := new(presenter.PrimeroWebPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestPrimeroWebPresenterOutputCarriesTheHint(t *testing.T) {
	// 既存の HintOutput テストと同じ状態。
	g := domain.NewDefaultPrimero()
	g.Reset()
	require.NotNil(t, g.GetHint(), "fixture must actually produce a hint")

	out := new(presenter.PrimeroWebPresenter).Output(g, nil)
	assert.Contains(t, out, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, out, "primero.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestPrimeroWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g := domain.NewDefaultPrimero()
	g.Reset()
	assert.Contains(t, new(presenter.PrimeroWebPresenter).HintOutput(g), "primero.hintRequested")
}
