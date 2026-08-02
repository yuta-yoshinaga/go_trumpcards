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

// gutsWebSetHand は player の手札を 2 枚に差し替える (プレゼンターテスト用ヘルパー)。
func gutsWebSetHand(p *domain.GutsPlayer, d1, v1, d2, v2 int) {
	p.ClearHand()
	p.AddCard(domain.NewCard(d1, v1, false))
	p.AddCard(domain.NewCard(d2, v2, false))
}

// gutsWebResultGame は決定的な結果フェーズのゲームを組み立てる。
// seat0 (人間) はノーペア (負け)、seat1 (CPU) はエースのペア (勝ち)。
// 他の座席はアウト宣言にして乱数配札の影響を排除する。
func gutsWebResultGame() *domain.Guts {
	g := domain.NewDefaultGuts()
	gutsWebSetHand(g.GetPlayer(0), 0, 5, 1, 8) // 8-high, no pair
	gutsWebSetHand(g.GetPlayer(1), 0, 1, 1, 1) // pair of aces
	g.GetPlayer(0).SetIn(true)
	g.GetPlayer(1).SetIn(true)
	for i := 2; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetIn(false)
	}
	g.SettleForTest()
	return g
}

func TestGutsWebPresenter_OutputDeclarePhase(t *testing.T) {
	g := domain.NewDefaultGuts()
	p := new(presenter.GutsWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.GutsPhaseDeclare), decoded["phase"])
	assert.Contains(t, decoded, "players")
	assert.Contains(t, decoded, "pot")
	assert.Contains(t, decoded, "config")
	assert.Equal(t, "guts.declarePhase", decoded["messageCode"])
}

func TestGutsWebPresenter_Error(t *testing.T) {
	g := domain.NewDefaultGuts()
	p := new(presenter.GutsWebPresenter)
	out := p.Output(g, errors.New("boom"))

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "boom", decoded["message"])
}

func TestGutsWebPresenter_ResultHumanLose(t *testing.T) {
	g := gutsWebResultGame()
	p := new(presenter.GutsWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.GutsPhaseResult), decoded["phase"])
	assert.Equal(t, "guts.roundEndHumanLose", decoded["messageCode"])
	assert.Equal(t, float64(1), decoded["winnerIdx"])
	// The human (seat 0) stayed in and lost, so it is a matcher.
	matchers, ok := decoded["matchers"].([]any)
	require.True(t, ok)
	assert.Contains(t, matchers, float64(0))

	// The winning CPU's hand is revealed with a hand-name key.
	players, ok := decoded["players"].([]any)
	require.True(t, ok)
	cpu := players[1].(map[string]any)
	assert.Equal(t, "pair", cpu["handName"])
	assert.Equal(t, true, cpu["isWinner"])
	// The human's own hand name is populated too (revealed cards).
	human := players[0].(map[string]any)
	assert.Equal(t, "highcard", human["handName"])
}

func TestGutsWebPresenter_ResultHumanWin(t *testing.T) {
	g := domain.NewDefaultGuts()
	gutsWebSetHand(g.GetPlayer(0), 0, 1, 1, 1) // pair of aces (win)
	gutsWebSetHand(g.GetPlayer(1), 0, 5, 1, 8) // 8-high, no pair
	g.GetPlayer(0).SetIn(true)
	g.GetPlayer(1).SetIn(true)
	for i := 2; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetIn(false)
	}
	g.SettleForTest()

	p := new(presenter.GutsWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "guts.roundEndHumanWin", decoded["messageCode"])
	assert.Equal(t, float64(0), decoded["winnerIdx"])
}

func TestGutsWebPresenter_ResultCarry(t *testing.T) {
	g := domain.NewDefaultGuts()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		g.GetPlayer(i).SetIn(false)
	}
	g.SettleForTest()

	p := new(presenter.GutsWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "guts.roundEndCarry", decoded["messageCode"])
	assert.Equal(t, float64(-1), decoded["winnerIdx"])
}

func TestGutsWebPresenter_GameEnd(t *testing.T) {
	g := domain.NewDefaultGuts()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	require.NoError(t, g.Declare(true))
	require.True(t, g.GetGameEndFlag())

	p := new(presenter.GutsWebPresenter)
	out := p.Output(g, nil)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, true, decoded["gameEndFlag"])
	assert.Contains(t, decoded["messageCode"], "guts.result.")
}

func TestGutsWebPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultGuts()
	p := new(presenter.GutsWebPresenter)
	out := p.HintOutput(g)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Contains(t, decoded, "hint")
}

func TestGutsWebPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultGuts()
	require.NoError(t, g.Declare(true))
	p := new(presenter.GutsWebPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestGutsWebPresenterOutputCarriesTheHint(t *testing.T) {
	// 既存の HintOutput テストと同じ状態。300 回試して nil 0 件で確認済み。
	g := domain.NewDefaultGuts()
	g.Reset()
	require.NotNil(t, g.GetHint(), "fixture must actually produce a hint")

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(new(presenter.GutsWebPresenter).Output(g, nil)), &decoded))
	assert.Contains(t, decoded, "hint", "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotEqual(t, "guts.hintRequested", decoded["messageCode"])
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestGutsWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g := domain.NewDefaultGuts()
	g.Reset()
	assert.Contains(t, new(presenter.GutsWebPresenter).HintOutput(g), "guts.hintRequested")
}

// **ヒントが無いときの分岐も見る。**Output() の受動ヒントは nil のとき
// `hint` キーごと落ちる。HintOutput() は noHint を返す。codecov が
// PR #4591 でこの 2 本を未到達として報告した。
func TestGutsWebPresenterWithoutAHint(t *testing.T) {
	g := gutsWebResultGame() // 結果フェーズ = 宣言フェーズではないので GetHint は nil
	require.Nil(t, g.GetHint(), "fixture must actually produce no hint")

	p := new(presenter.GutsWebPresenter)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &decoded))
	assert.NotContains(t, decoded, "hint")

	assert.Contains(t, p.HintOutput(g), "guts.noHint")
}
