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

// anacondaWebSetKept は player の手札を任意枚数に差し替える。
func anacondaWebSetKept(p *domain.AnacondaPlayer, cs ...*domain.Card) {
	p.ClearHand()
	for _, c := range cs {
		p.AddCard(c)
	}
}

func anacondaWebCard(d, v int) *domain.Card { return domain.NewCard(d, v, false) }

// anacondaWebWeak は弱いハイカード 5 枚。
func anacondaWebWeak() []*domain.Card {
	return []*domain.Card{
		anacondaWebCard(domain.CardDesignSpade, 2), anacondaWebCard(domain.CardDesignHeart, 3),
		anacondaWebCard(domain.CardDesignClover, 5), anacondaWebCard(domain.CardDesignDiamond, 7),
		anacondaWebCard(domain.CardDesignSpade, 9),
	}
}

func TestAnacondaWebPresenter_OutputPassPhase(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	p := new(presenter.AnacondaWebPresenter)
	out := p.Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.AnacondaPhasePass), decoded["phase"])
	assert.Contains(t, decoded, "players")
	assert.Contains(t, decoded, "pot")
	assert.Contains(t, decoded, "config")
	assert.Equal(t, "anaconda.passPhase", decoded["messageCode"])
}

func TestAnacondaWebPresenter_Error(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	p := new(presenter.AnacondaWebPresenter)
	out := p.Output(g, errors.New("boom"))

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "boom", decoded["message"])
}

func TestAnacondaWebPresenter_SetAndRollPhase(t *testing.T) {
	p := new(presenter.AnacondaWebPresenter)

	g := domain.NewDefaultAnaconda()
	g.SetPhase(domain.AnacondaPhaseSet)
	var setOut map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &setOut))
	assert.Equal(t, "anaconda.setPhase", setOut["messageCode"])

	g.SetPhase(domain.AnacondaPhaseRoll)
	g.SetRollIndex(2)
	var rollOut map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &rollOut))
	assert.Equal(t, "anaconda.rollPhase", rollOut["messageCode"])
}

func TestAnacondaWebPresenter_ResultHumanWin(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	// seat0 four of a kind, others weak.
	anacondaWebSetKept(g.GetPlayer(0),
		anacondaWebCard(domain.CardDesignSpade, 8), anacondaWebCard(domain.CardDesignHeart, 8),
		anacondaWebCard(domain.CardDesignClover, 8), anacondaWebCard(domain.CardDesignDiamond, 8),
		anacondaWebCard(domain.CardDesignSpade, 2))
	for i := 1; i < g.GetPlayerCnt(); i++ {
		anacondaWebSetKept(g.GetPlayer(i), anacondaWebWeak()...)
	}
	g.SetPhase(domain.AnacondaPhaseRoll)
	g.SetPot(100)
	g.ResolveShowdownForTest()

	p := new(presenter.AnacondaWebPresenter)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &decoded))
	assert.Equal(t, float64(domain.AnacondaPhaseResult), decoded["phase"])
	assert.Equal(t, "anaconda.roundEndHumanWin", decoded["messageCode"])
	assert.Equal(t, float64(0), decoded["winnerIdx"])

	players, ok := decoded["players"].([]any)
	require.True(t, ok)
	human := players[0].(map[string]any)
	assert.Equal(t, "fourkind", human["handName"])
	assert.Equal(t, true, human["isWinner"])
}

func TestAnacondaWebPresenter_ResultCpuWin(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	// seat1 four of a kind, human weak → human loses.
	for i := 0; i < g.GetPlayerCnt(); i++ {
		anacondaWebSetKept(g.GetPlayer(i), anacondaWebWeak()...)
	}
	anacondaWebSetKept(g.GetPlayer(1),
		anacondaWebCard(domain.CardDesignSpade, 8), anacondaWebCard(domain.CardDesignHeart, 8),
		anacondaWebCard(domain.CardDesignClover, 8), anacondaWebCard(domain.CardDesignDiamond, 8),
		anacondaWebCard(domain.CardDesignSpade, 2))
	g.SetPhase(domain.AnacondaPhaseRoll)
	g.SetPot(80)
	g.ResolveShowdownForTest()

	p := new(presenter.AnacondaWebPresenter)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &decoded))
	assert.Equal(t, "anaconda.roundEndHumanLose", decoded["messageCode"])
	assert.Equal(t, float64(1), decoded["winnerIdx"])
}

func TestAnacondaWebPresenter_GameEnd(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		anacondaWebSetKept(g.GetPlayer(i), anacondaWebWeak()...)
	}
	g.SetPhase(domain.AnacondaPhaseRoll)
	g.ResolveShowdownForTest()
	require.True(t, g.GetGameEndFlag())

	p := new(presenter.AnacondaWebPresenter)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &decoded))
	assert.Equal(t, true, decoded["gameEndFlag"])
	assert.Contains(t, decoded["messageCode"], "anaconda.result.")
}

func TestAnacondaWebPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	p := new(presenter.AnacondaWebPresenter)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.HintOutput(g)), &decoded))
	assert.Contains(t, decoded, "hint")
}

func TestAnacondaWebPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	require.NoError(t, g.Pass([]int{0, 1, 2}))
	p := new(presenter.AnacondaWebPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestAnacondaWebPresenterOutputCarriesTheHint(t *testing.T) {
	// Reset 直後にヒントが出る。300 回試して nil 0 件で確認済み。
	g := domain.NewDefaultAnaconda()
	g.Reset()
	require.NotNil(t, g.GetHint(), "fixture must actually produce a hint")

	out := new(presenter.AnacondaWebPresenter).Output(g, nil)
	assert.Contains(t, out, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, out, "anaconda.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestAnacondaWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	g.Reset()
	assert.Contains(t, new(presenter.AnacondaWebPresenter).HintOutput(g), "anaconda.hintRequested")
}

// **ヒントが無いときの分岐も見る。**Output() の受動ヒントは nil のとき
// `hint` キーごと落ちる。HintOutput() は noHint を返す。codecov が
// PR #4593 でこの 2 本を未到達として報告した。
func TestAnacondaWebPresenterWithoutAHint(t *testing.T) {
	g := domain.NewDefaultAnaconda()
	g.GetPlayer(0).SetOut(true) // 降りた人に助言することがない
	require.Nil(t, g.GetHint(), "fixture must actually produce no hint")

	p := new(presenter.AnacondaWebPresenter)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &decoded))
	assert.NotContains(t, decoded, "hint")

	assert.Contains(t, p.HintOutput(g), "anaconda.noHint")
}
