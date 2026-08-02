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

// bouillotteWebSetHand は player の手札を 3 枚に差し替える (プレゼンターテスト用ヘルパー)。
func bouillotteWebSetHand(p *domain.BouillottePlayer, cards ...*domain.Card) {
	p.ClearHand()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// bouillotteResultGame は決定的な結果フェーズのゲームを組み立てる。
// humanWins=false のとき seat0 (人間) はハイカード (負け)、seat1 (CPU) はブルラン (勝ち)。
func bouillotteResultGame(humanWins bool) *domain.Bouillotte {
	g := domain.NewDefaultBouillotte()
	g.SetPhase(domain.BouillottePhaseBetting)
	g.SetRetourne(domain.NewCard(domain.CardDesignDiamond, 8, false))
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetFolded(false)
		g.GetPlayer(i).SetOut(false)
	}
	weak := []*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false), domain.NewCard(domain.CardDesignClover, 9, false), domain.NewCard(domain.CardDesignHeart, 8, false)}
	brelan := []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false), domain.NewCard(domain.CardDesignClover, 1, false), domain.NewCard(domain.CardDesignHeart, 1, false)}
	if humanWins {
		bouillotteWebSetHand(g.GetPlayer(0), brelan...)
		bouillotteWebSetHand(g.GetPlayer(1), weak...)
	} else {
		bouillotteWebSetHand(g.GetPlayer(0), weak...)
		bouillotteWebSetHand(g.GetPlayer(1), brelan...)
	}
	for i := 2; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetFolded(true)
	}
	g.SetPot(100)
	g.ResolveForTest()
	return g
}

func TestBouillotteWebPresenter_OutputBettingPhase(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	p := new(presenter.BouillotteWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, decoded, "players")
	assert.Contains(t, decoded, "pot")
	assert.Contains(t, decoded, "config")
	assert.Contains(t, decoded, "retourne")
	assert.Contains(t, decoded, "currentPlayerIdx")
	if decoded["phase"] == float64(domain.BouillottePhaseBetting) && !decoded["gameEndFlag"].(bool) {
		assert.Equal(t, "bouillotte.bettingPhase", decoded["messageCode"])
	}
}

func TestBouillotteWebPresenter_Error(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	p := new(presenter.BouillotteWebPresenter)
	out := p.Output(g, errors.New("boom"))

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "boom", decoded["message"])
}

func TestBouillotteWebPresenter_ResultHumanLose(t *testing.T) {
	g := bouillotteResultGame(false)
	p := new(presenter.BouillotteWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.BouillottePhaseResult), decoded["phase"])
	assert.Equal(t, "bouillotte.roundEndHumanLose", decoded["messageCode"])
	assert.Equal(t, float64(1), decoded["winnerIdx"])

	players, ok := decoded["players"].([]any)
	require.True(t, ok)
	cpu := players[1].(map[string]any)
	assert.Equal(t, "brelan", cpu["handName"])
	assert.Equal(t, true, cpu["isWinner"])
	// The human's own hand name is populated too (revealed cards).
	human := players[0].(map[string]any)
	assert.Equal(t, "highcard", human["handName"])
}

func TestBouillotteWebPresenter_ResultHumanWin(t *testing.T) {
	g := bouillotteResultGame(true)
	p := new(presenter.BouillotteWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "bouillotte.roundEndHumanWin", decoded["messageCode"])
	assert.Equal(t, float64(0), decoded["winnerIdx"])
}

func TestBouillotteWebPresenter_GameEnd(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	for i := 0; i < 100 && g.GetPhase() == domain.BouillottePhaseBetting && g.IsHumanTurn(); i++ {
		require.NoError(t, g.PlayerCall())
	}
	require.True(t, g.GetGameEndFlag())

	p := new(presenter.BouillotteWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, true, decoded["gameEndFlag"])
	assert.Contains(t, decoded["messageCode"], "bouillotte.result.")
}

func TestBouillotteWebPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	p := new(presenter.BouillotteWebPresenter)
	out := p.HintOutput(g)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, decoded, "players")
}

func TestBouillotteWebPresenter_ActionLog(t *testing.T) {
	g := bouillotteResultGame(false)
	p := new(presenter.BouillotteWebPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestBouillotteWebPresenterOutputCarriesTheHint(t *testing.T) {
	// Reset 直後にヒントが出る。300 回試して nil 0 件で確認済み。
	g := domain.NewDefaultBouillotte()
	g.Reset()
	require.NotNil(t, g.GetHint(), "fixture must actually produce a hint")

	out := new(presenter.BouillotteWebPresenter).Output(g, nil)
	assert.Contains(t, out, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, out, "bouillotte.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestBouillotteWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g := domain.NewDefaultBouillotte()
	g.Reset()
	assert.Contains(t, new(presenter.BouillotteWebPresenter).HintOutput(g), "bouillotte.hintRequested")
}
