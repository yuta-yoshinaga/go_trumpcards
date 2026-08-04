package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestMichiganCuiPresenter_OutputBetPhase(t *testing.T) {
	g := domain.NewDefaultMichigan()
	p := new(presenter.MichiganCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "ミシガン")
	assert.Contains(t, out, "ラウンド")
	assert.Contains(t, out, "ブードル")
}

func TestMichiganCuiPresenter_OutputBetHint(t *testing.T) {
	p := new(presenter.MichiganCuiPresenter)

	// Held: the human holds the card that claims boodle 0 → it is recommended.
	g := domain.NewDefaultMichigan()
	bc := g.GetBoodle(0).GetCard()
	g.GetPlayer(0).AddCard(domain.NewCard(bc.GetDesign(), bc.GetValue(), false))
	out := p.Output(g, nil)
	assert.Contains(t, out, "推奨")
	assert.Contains(t, out, "boodle0")

	// None: an empty human hand holds no boodle cards → the even-spread tip shows.
	g2 := domain.NewDefaultMichigan()
	g2.GetPlayer(0).ClearHand()
	out2 := p.Output(g2, nil)
	assert.Contains(t, out2, "均等")
}

// **確定済みブードルにも触れる。**Web は `betClaimedWarning` を出すのに CUI の
// ベットヒントは「持っている札」しか案内せず、賭けても回収できないブードルを
// 黙って見過ごしていた (#4926)。
func TestMichiganCuiPresenter_BetHintWarnsAboutClaimedBoodles(t *testing.T) {
	p := new(presenter.MichiganCuiPresenter)

	// 未確定だけなら警告は出ない。
	clean := domain.NewDefaultMichigan()
	assert.NotContains(t, p.Output(clean, nil), "回収できません")

	// boodle1 が確定済み。ヒントに警告が出る。
	g := domain.NewDefaultMichigan()
	g.GetBoodle(1).SetClaimedBy(2)
	out := p.Output(g, nil)
	assert.Contains(t, out, "回収できません")
	assert.Contains(t, out, "boodle1")
	assert.NotContains(t, out, "boodle0")

	// **確定済みは推奨から外れる。**札を持っていても回収できない以上、
	// 「厚めに賭けろ」と案内してはいけない。
	g2 := domain.NewDefaultMichigan()
	bc := g2.GetBoodle(0).GetCard()
	g2.GetPlayer(0).AddCard(domain.NewCard(bc.GetDesign(), bc.GetValue(), false))
	assert.Contains(t, p.Output(g2, nil), "厚めに")
	g2.GetBoodle(0).SetClaimedBy(3)
	out2 := p.Output(g2, nil)
	assert.Contains(t, out2, "回収できません")
	// 推奨行そのものが消え、「手札に無し」の案内に切り替わる。
	assert.NotContains(t, out2, "厚めに")
	assert.Contains(t, out2, "均等")
}

// ja / en 双方に訳がある。片方だけだと `--lang en` でキーが出る (#4926)。
func TestMichiganBetHintClaimed_TranslatedInBothLanguages(t *testing.T) {
	defer i18n.SetLang("ja")
	for _, lang := range []string{"ja", "en"} {
		i18n.SetLang(lang)
		got := i18n.T("michigan.betHintClaimed")
		assert.NotEqual(t, "michigan.betHintClaimed", got, "missing from %s", lang)
	}
	i18n.SetLang("ja")
	ja := i18n.T("michigan.betHintClaimed")
	i18n.SetLang("en")
	assert.NotEqual(t, ja, i18n.T("michigan.betHintClaimed"))
}

func TestMichiganCuiPresenter_OutputPlayPhase(t *testing.T) {
	g := domain.NewDefaultMichigan()
	require.NoError(t, g.PlaceHumanBet(michiganWebEvenBet(g.GetBetBudget())))
	p := new(presenter.MichiganCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "デッドハンド")
}

func TestMichiganCuiPresenter_OutputError(t *testing.T) {
	g := domain.NewDefaultMichigan()
	p := new(presenter.MichiganCuiPresenter)
	out := p.Output(g, errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestMichiganCuiPresenter_OutputResult(t *testing.T) {
	g := michiganResultGame(false)
	p := new(presenter.MichiganCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "ラウンド")
}

func TestMichiganCuiPresenter_OutputGameEnd(t *testing.T) {
	g := domain.NewDefaultMichigan()
	cfg := g.GetConfig()
	cfg.TargetRounds = 1
	g.SetConfig(cfg)
	g.Reset()
	require.NoError(t, g.PlaceHumanBet(michiganWebEvenBet(g.GetBetBudget())))
	for i := 0; i < 300 && g.GetPhase() == domain.MichiganPhasePlay && g.IsHumanTurn(); i++ {
		pi := g.GetPlayableIndices()
		require.NotEmpty(t, pi)
		require.NoError(t, g.PlayCard(pi[0]))
	}
	require.True(t, g.GetGameEndFlag())
	p := new(presenter.MichiganCuiPresenter)
	assert.NotEmpty(t, p.Output(g, nil))
}

func TestMichiganCuiPresenter_HintOutput(t *testing.T) {
	g := domain.NewDefaultMichigan()
	require.NoError(t, g.PlaceHumanBet(michiganWebEvenBet(g.GetBetBudget())))
	p := new(presenter.MichiganCuiPresenter)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestMichiganCuiPresenter_HintOutputNone(t *testing.T) {
	g := michiganResultGame(false) // result phase -> GetHint returns nil
	p := new(presenter.MichiganCuiPresenter)
	assert.NotEmpty(t, p.HintOutput(g))
}

func TestMichiganCuiPresenter_ActionLog(t *testing.T) {
	g := michiganResultGame(false)
	p := new(presenter.MichiganCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
