//go:build test

package presenter_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// dehlaPakadGame は 5 枚を配り終えた卓を返す (席 0 が切り札を決める)。
func dehlaPakadGame() *domain.DehlaPakad {
	d := domain.NewDefaultDehlaPakad()
	d.Reset()
	return d
}

// dehlaPakadStartPlay は切り札を決めて人間の手番まで進める。
func dehlaPakadStartPlay(t *testing.T, d *domain.DehlaPakad) {
	t.Helper()
	require.NoError(t, d.SelectTrump(domain.CardDesignHeart))
	for i := 0; i < 64; i++ {
		if d.GetPhase() != domain.DehlaPakadPhasePlay || d.IsHumanTurn() {
			return
		}
		d.CpuPlay()
	}
}

func TestDehlaPakadCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.DehlaPakadCuiPresenter)

	t.Run("shows the hand, the tens and the kots", func(t *testing.T) {
		out := p.Output(dehlaPakadGame(), nil)
		// **訳が引けていることまで見る。** キー一致だけを見ると、ロケールが
		// 丸ごと欠けていても両辺が生キーで一致して通ってしまう。
		assert.Contains(t, out, "デーラ・パカド")
		assert.NotContains(t, out, "dehlapakad.", "生キーが出ている")
		assert.Contains(t, out, strings.SplitN(i18n.T("dehlapakad.hand"), "{{", 2)[0])
		assert.Contains(t, out, strings.SplitN(i18n.T("dehlapakad.teamLine"), "{{", 2)[0])
	})

	// **切り札を決める前に見えるのは 5 枚だけ。** 13 枚見せると別のゲームになる。
	t.Run("shows only five cards before the trump is called", func(t *testing.T) {
		out := p.Output(dehlaPakadGame(), nil)
		assert.Contains(t, out, "[4]")
		assert.NotContains(t, out, "[5]", "宣言前に 6 枚目が見えている")
		assert.Contains(t, out, strings.SplitN(i18n.T("dehlapakad.promptTrump"), "{{", 2)[0])
		assert.NotContains(t, out, strings.SplitN(i18n.T("dehlapakad.trump"), "{{", 2)[0])
	})

	// **絵札は A/J/Q/K で出す。** 10 が的のゲームで、A が「1」と出ると
	// 数札の並びに紛れる。
	t.Run("prints court cards by their face label", func(t *testing.T) {
		out := p.Output(dehlaPakadGame(), nil)
		assert.NotRegexp(t, `[♠♣♥♦](11|12|13)\b`, out, "絵札が数値で出ている")
		assert.NotRegexp(t, `[♠♣♥♦]1\b`, out, "エースが数値で出ている")
	})

	t.Run("play phase lists the playable cards", func(t *testing.T) {
		d := dehlaPakadGame()
		dehlaPakadStartPlay(t, d)
		if d.GetPhase() != domain.DehlaPakadPhasePlay || !d.IsHumanTurn() {
			t.Skip("配りによっては人間の手番の前にトリックが揃う")
		}
		out := p.Output(d, nil)
		assert.Contains(t, out, strings.SplitN(i18n.T("dehlapakad.playableList"), "{{", 2)[0])
		assert.Contains(t, out, i18n.T("dehlapakad.promptPlayHelp"))
	})

	// **これがこのゲームの心臓部。** 山と「次も取れば誰が持っていくか」を
	// 出さないと、いま何を賭けているのかが読めない。
	t.Run("shows the centre pile and who would collect it", func(t *testing.T) {
		d := dehlaPakadGame()
		dehlaPakadStartPlay(t, d)
		dehlaPakadPlayOneTrick(t, d)
		require.NotEmpty(t, d.GetCentrePile(), "1 トリック目で山が空になっている")
		out := p.Output(d, nil)
		assert.Contains(t, out, strings.SplitN(i18n.T("dehlapakad.centrePile"), "{{", 2)[0])
		assert.Contains(t, out, strings.SplitN(i18n.T("dehlapakad.pileGoesTo"), "{{", 2)[0])
	})

	t.Run("says nothing about the pile while it is empty", func(t *testing.T) {
		out := p.Output(dehlaPakadGame(), nil)
		assert.NotContains(t, out, strings.SplitN(i18n.T("dehlapakad.centrePile"), "{{", 2)[0])
	})

	t.Run("errors are shown", func(t *testing.T) {
		assert.Contains(t, p.Output(dehlaPakadGame(), assert.AnError), assert.AnError.Error())
	})

	// **英語も訳が引ける。** 反対の言語が漏れていれば生キーが出る。
	t.Run("renders in english too", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		out := p.Output(dehlaPakadGame(), nil)
		assert.NotContains(t, out, "dehlapakad.")
		assert.NotContains(t, out, "デーラ・パカド", "日本語が漏れている")
	})
}

func TestDehlaPakadCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.DehlaPakadCuiPresenter)

	// 宣言フェーズでは切り札を勧める。
	out := p.HintOutput(dehlaPakadGame())
	assert.NotContains(t, out, "dehlapakad.", "生キーが出ている")
	assert.Contains(t, out, strings.SplitN(i18n.T("dehlapakad.hintTrump"), "{{", 2)[0])

	d := dehlaPakadGame()
	dehlaPakadStartPlay(t, d)
	if d.GetPhase() == domain.DehlaPakadPhasePlay && d.IsHumanTurn() {
		out = p.HintOutput(d)
		assert.NotContains(t, out, "dehlapakad.")
		assert.Contains(t, out, "[")
	}
}

func TestDehlaPakadWebPresenter_Output(t *testing.T) {
	i18n.SetLang("ja")
	p := new(presenter.DehlaPakadWebPresenter)
	d := dehlaPakadGame()

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(d, nil)), &res))

	assert.Equal(t, domain.DehlaPakadPhaseSelectTrump, res["phase"])
	assert.Equal(t, true, res["isTrumpPhase"])
	assert.Equal(t, float64(-1), res["trumpSuit"], "宣言前に切り札が入っている")
	assert.Equal(t, float64(domain.DehlaPakadTrickCount), res["trickCount"])
	assert.Len(t, res["players"], domain.DehlaPakadPlayerCnt)
	assert.Len(t, res["teamTens"], domain.DehlaPakadTeamCnt)
	assert.Len(t, res["teamKots"], domain.DehlaPakadTeamCnt)
	// **決めるのは親の右隣。** 画面がボタンを誰に出すかがこれで決まる。
	assert.Equal(t, float64(domain.DehlaPakadNextSeat(d.GetDealerIdx())), res["trumpChooserIdx"])
	// null ではなく空配列で運ぶ。
	assert.NotNil(t, res["playableIndices"])
	assert.NotNil(t, res["handHistory"])

	players := res["players"].([]any)
	human := players[0].(map[string]any)
	assert.Len(t, human["cards"], domain.DehlaPakadFirstBatch, "宣言前は 5 枚だけ")
	assert.Equal(t, true, human["isTrumpChooser"])
	cpu := players[1].(map[string]any)
	assert.Empty(t, cpu["cards"], "CPU の手札が見えている")
}

// **山と「次も取れば誰が持っていくか」を運ぶ。** これが無いと、画面が
// このゲームの肝を出せない。
func TestDehlaPakadWebPresenter_CarriesTheCentrePile(t *testing.T) {
	p := new(presenter.DehlaPakadWebPresenter)
	d := dehlaPakadGame()
	dehlaPakadStartPlay(t, d)
	dehlaPakadPlayOneTrick(t, d)

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(d, nil)), &res))
	assert.Equal(t, float64(domain.DehlaPakadPlayerCnt), res["centrePileCount"], "1 トリック分が山に無い")
	assert.Equal(t, float64(d.GetCentrePileTens()), res["centrePileTens"])
	assert.Equal(t, float64(d.GetPrevTrickWinner()), res["prevTrickWinner"])
	assert.GreaterOrEqual(t, res["prevTrickWinner"], float64(0))
}

func TestDehlaPakadWebPresenter_MessageCodes(t *testing.T) {
	p := new(presenter.DehlaPakadWebPresenter)
	d := dehlaPakadGame()

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(d, nil)), &res))
	assert.Equal(t, "dehlapakad.selectTrump", res["messageCode"])

	dehlaPakadStartPlay(t, d)
	require.NoError(t, json.Unmarshal([]byte(p.Output(d, nil)), &res))
	assert.Contains(t, []any{"dehlapakad.playPhase", "dehlapakad.playPhase.tensAtStake", "dehlapakad.handEnd"},
		res["messageCode"])

	// **ヒントは頼まれたときだけ名乗る。**
	require.NoError(t, json.Unmarshal([]byte(p.HintOutput(d)), &res))
	assert.Contains(t, []any{"dehlapakad.hintRequested", "dehlapakad.noHint"}, res["messageCode"])
}

func TestDehlaPakadWebPresenter_ErrorMessage(t *testing.T) {
	p := new(presenter.DehlaPakadWebPresenter)
	d := dehlaPakadGame()
	err := d.SelectTrump(99)
	require.Error(t, err)
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(d, err)), &res))
	assert.Equal(t, "dehlapakad.errTrumpSuit", res["messageCode"])
	assert.NotEmpty(t, res["message"])
}

func TestDehlaPakadPresenters_ActionLogOutput(t *testing.T) {
	i18n.SetLang("ja")
	d := dehlaPakadGame()
	dehlaPakadStartPlay(t, d)

	cui := new(presenter.DehlaPakadCuiPresenter)
	assert.NotEmpty(t, cui.ActionLogOutput(d))

	web := new(presenter.DehlaPakadWebPresenter)
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(web.ActionLogOutput(d)), &res))
	assert.Contains(t, res, "entries")
}

// dehlaPakadPlayOneTrick は 1 トリックぶん打ち切る。
func dehlaPakadPlayOneTrick(t *testing.T, d *domain.DehlaPakad) {
	t.Helper()
	start := d.GetTrickNumber()
	for i := 0; i < 32 && d.GetTrickNumber() == start && d.GetPhase() == domain.DehlaPakadPhasePlay; i++ {
		if d.IsHumanTurn() {
			valid := d.GetPlayableIndices(d.GetCurrentPlayerIdx())
			require.NotEmpty(t, valid)
			require.NoError(t, d.PlayerPlay(valid[0]))
			continue
		}
		d.CpuPlay()
	}
}
