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

func dilotiGame() *domain.Diloti {
	d := domain.NewDefaultDiloti()
	d.Reset()
	return d
}

func dilotiCard(design, value int) *domain.Card { return domain.NewCard(design, value, true) }

func TestDilotiCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.DilotiCuiPresenter)

	t.Run("shows the round, the stock and the table", func(t *testing.T) {
		out := p.Output(dilotiGame(), nil)
		// **訳が引けていることまで見る。**
		assert.Contains(t, out, "ディロティ")
		assert.NotContains(t, out, "diloti.", "生キーが出ている")
		assert.Contains(t, out, strings.SplitN(i18n.T("diloti.round"), "{{", 2)[0])
		assert.Contains(t, out, strings.SplitN(i18n.T("diloti.deck"), "{{", 2)[0])
		assert.Contains(t, out, strings.SplitN(i18n.T("diloti.table"), "{{", 2)[0])
	})

	// **場札には番号が要る。** 取る対象はこの番号で指す。
	t.Run("numbers the table cards", func(t *testing.T) {
		out := p.Output(dilotiGame(), nil)
		for _, n := range []string{"[0]", "[1]", "[2]", "[3]"} {
			assert.Contains(t, out, n)
		}
	})

	// **絵札は A/J/Q/K で出す。** 絵札は合計に入らず同ランクでしか取れないので、
	// 11/12/13 と数字で出ると「合計に使えそう」に見える。
	t.Run("prints court cards by their face label", func(t *testing.T) {
		d := dilotiGame()
		d.SetTableForTest([]*domain.Card{
			dilotiCard(domain.CardDesignSpade, 11),
			dilotiCard(domain.CardDesignClover, 12),
			dilotiCard(domain.CardDesignHeart, 13),
			dilotiCard(domain.CardDesignDiamond, 1),
		}, nil)
		out := p.Output(d, nil)
		assert.NotRegexp(t, `[♠♣♥♦](11|12|13)\b`, out, "絵札が数値で出ている")
		for _, want := range []string{"♠J", "♣Q", "♥K", "♦A"} {
			assert.Contains(t, out, want)
		}
	})

	// **宣言も番号付きで見える。** 見えないと `d0` で取る対象を指せない。
	t.Run("shows the declarations with their index and kind", func(t *testing.T) {
		d := dilotiGame()
		decl := domain.NewDilotiDeclaration(1, 8,
			[]*domain.Card{dilotiCard(domain.CardDesignSpade, 3), dilotiCard(domain.CardDesignHeart, 5)})
		d.SetTableForTest([]*domain.Card{dilotiCard(domain.CardDesignClover, 2)},
			[]*domain.DilotiDeclaration{decl})
		out := p.Output(d, nil)
		assert.Contains(t, out, strings.SplitN(i18n.T("diloti.declaration"), "{{", 2)[0])
		assert.Contains(t, out, i18n.T("diloti.declPlain"))
		assert.NotContains(t, out, i18n.T("diloti.declGroup"), "単一宣言がグループ扱いになっている")

		decl.AddGroup([]*domain.Card{dilotiCard(domain.CardDesignClover, 8)})
		out2 := p.Output(d, nil)
		assert.Contains(t, out2, i18n.T("diloti.declGroup"), "グループ宣言だと分からない")
	})

	// **何が取れて何が宣言できるかを出す。**
	t.Run("lists what each card can take and declare", func(t *testing.T) {
		d := dilotiGame()
		d.SetTableForTest([]*domain.Card{
			dilotiCard(domain.CardDesignHeart, 3),
			dilotiCard(domain.CardDesignClover, 4),
		}, nil)
		d.GetPlayer(0).Reset()
		d.GetPlayer(0).AddCard(dilotiCard(domain.CardDesignSpade, 3)) // 同ランク取り
		d.GetPlayer(0).AddCard(dilotiCard(domain.CardDesignDiamond, 7))
		require.NotEmpty(t, d.GetTakeOptions(0, 0), "固定した盤面で取れない")
		require.NotEmpty(t, d.GetDeclareOptions(0, 0), "固定した盤面で宣言できない")

		out := p.Output(d, nil)
		assert.Contains(t, out, strings.SplitN(i18n.T("diloti.takeOption"), "{{", 2)[0])
		assert.Contains(t, out, strings.SplitN(i18n.T("diloti.declareOption"), "{{", 2)[0])
	})

	t.Run("errors are shown", func(t *testing.T) {
		assert.Contains(t, p.Output(dilotiGame(), assert.AnError), assert.AnError.Error())
	})

	// **英語も訳が引ける。**
	t.Run("renders in english too", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		out := p.Output(dilotiGame(), nil)
		assert.NotContains(t, out, "diloti.")
		assert.NotContains(t, out, "ディロティ", "日本語が漏れている")
	})
}

func TestDilotiCuiPresenter_RoundResultAndHint(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.DilotiCuiPresenter)

	// **局の終わりが終局を兼ねることがある。** 1 局で 61 点に届く配りだと
	// 集計ではなく勝者の行が出るので、集計だけを固定して待つ ── 素直に
	// 1 局打って assert すると、その配りを引いた回だけ落ちる。
	d := dilotiRoundEndedButNotOver(t)
	out := p.Output(d, nil)
	assert.Contains(t, out, i18n.T("diloti.resultTitle"))
	assert.NotContains(t, out, "diloti.", "生キーが出ている")
	for _, key := range []string{"cards", "aces", "tenOfDiamonds", "twoOfClubs", "xeri"} {
		assert.Contains(t, out, i18n.T("diloti.score."+key), "得点項目 %s が出ていない", key)
	}

	hintOut := p.HintOutput(dilotiGame())
	assert.NotContains(t, hintOut, "diloti.", "生キーが出ている")
	assert.Contains(t, hintOut, strings.SplitN(i18n.T("diloti.hintCard"), "{{", 2)[0])
}

func TestDilotiWebPresenter_Output(t *testing.T) {
	i18n.SetLang("ja")
	p := new(presenter.DilotiWebPresenter)
	d := dilotiGame()

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(d, nil)), &res))

	assert.Equal(t, domain.DilotiPhasePlay, res["phase"])
	assert.Equal(t, float64(1), res["roundNumber"])
	assert.Len(t, res["players"], domain.DilotiPlayerCnt)
	assert.Len(t, res["table"], domain.DilotiTableSize)
	assert.Equal(t, float64(d.GetDeckRemaining()), res["deckRemaining"])
	assert.NotNil(t, res["takeOptions"])
	assert.NotNil(t, res["declareOptions"])
	assert.NotNil(t, res["canTrail"])
	assert.Len(t, res["canTrail"], domain.DilotiHandSize)

	players := res["players"].([]any)
	human := players[0].(map[string]any)
	assert.Len(t, human["cards"], domain.DilotiHandSize)
	// **相手の手札は伏せる。**
	cpu := players[1].(map[string]any)
	assert.Empty(t, cpu["cards"], "CPU の手札が見えている")
	assert.Equal(t, float64(domain.DilotiHandSize), cpu["cardCount"])
}

// **打てる手はサーバが数えて渡す。** 同ランク・合計一致・宣言が絡むので、
// クライアントに解かせると必ずずれる。
func TestDilotiWebPresenter_CarriesTheMoveOptions(t *testing.T) {
	p := new(presenter.DilotiWebPresenter)
	d := dilotiGame()
	decl := domain.NewDilotiDeclaration(1, 5,
		[]*domain.Card{dilotiCard(domain.CardDesignSpade, 5)})
	d.SetTableForTest([]*domain.Card{
		dilotiCard(domain.CardDesignHeart, 2),
		dilotiCard(domain.CardDesignClover, 3),
	}, []*domain.DilotiDeclaration{decl})
	d.GetPlayer(0).Reset()
	d.GetPlayer(0).AddCard(dilotiCard(domain.CardDesignDiamond, 5))

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(d, nil)), &res))
	takes := res["takeOptions"].([]any)
	require.Len(t, takes, 1)
	opts := takes[0].([]any)
	require.NotEmpty(t, opts, "2+3=5 と宣言 5 のどちらも挙がっていない")

	sawDecl, sawSweep := false, false
	for _, o := range opts {
		m := o.(map[string]any)
		if len(m["declIdxs"].([]any)) > 0 {
			sawDecl = true
		}
		if len(m["tableIdxs"].([]any)) == 2 && len(m["declIdxs"].([]any)) == 1 {
			sawSweep = true
		}
	}
	assert.True(t, sawDecl, "宣言を取る手が出ていない")
	assert.True(t, sawSweep, "場と宣言をまとめて払う手 (クセリ) が出ていない")

	// 宣言も渡す。
	decls := res["declarations"].([]any)
	require.Len(t, decls, 1)
	assert.Equal(t, float64(5), decls[0].(map[string]any)["value"])
	assert.Equal(t, false, decls[0].(map[string]any)["isGroup"])
}

func TestDilotiWebPresenter_NoOptionsOffTurn(t *testing.T) {
	p := new(presenter.DilotiWebPresenter)
	d := dilotiGame()
	dilotiRunToRoundEnd(t, d)

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(d, nil)), &res))
	assert.Empty(t, res["takeOptions"], "手番でないのに候補が出ている")
	require.NotNil(t, res["lastResult"])
	last := res["lastResult"].(map[string]any)
	assert.Len(t, last["lines"], 5)
	assert.Len(t, last["totals"], domain.DilotiPlayerCnt)
	assert.Len(t, last["xeris"], domain.DilotiPlayerCnt)
}

func TestDilotiWebPresenter_MessageCodes(t *testing.T) {
	p := new(presenter.DilotiWebPresenter)
	d := dilotiGame()

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(d, nil)), &res))
	assert.Contains(t, []any{"diloti.playPhase", "diloti.playPhase.canXeri"}, res["messageCode"])

	// **1 枚で場を払える手があるなら、そう伝える。** クセリ 1 回で 10 点あり、
	// 固定点の合計 11 点に匹敵する。
	d2 := dilotiGame()
	d2.SetTableForTest([]*domain.Card{dilotiCard(domain.CardDesignHeart, 6)}, nil)
	d2.GetPlayer(0).Reset()
	d2.GetPlayer(0).AddCard(dilotiCard(domain.CardDesignSpade, 6))
	require.NoError(t, json.Unmarshal([]byte(p.Output(d2, nil)), &res))
	assert.Equal(t, "diloti.playPhase.canXeri", res["messageCode"])

	// 払えない盤面では出ない (負のコントロール)。
	d3 := dilotiGame()
	d3.SetTableForTest([]*domain.Card{
		dilotiCard(domain.CardDesignHeart, 6), dilotiCard(domain.CardDesignClover, 9),
	}, nil)
	d3.GetPlayer(0).Reset()
	d3.GetPlayer(0).AddCard(dilotiCard(domain.CardDesignSpade, 6))
	require.NoError(t, json.Unmarshal([]byte(p.Output(d3, nil)), &res))
	assert.Equal(t, "diloti.playPhase", res["messageCode"])

	dilotiRunToRoundEnd(t, d)
	require.NoError(t, json.Unmarshal([]byte(p.Output(d, nil)), &res))
	assert.Contains(t, []any{"diloti.roundEnd", "diloti.result.humanWin", "diloti.result.cpuWin"},
		res["messageCode"])

	require.NoError(t, json.Unmarshal([]byte(p.HintOutput(dilotiGame())), &res))
	assert.Contains(t, []any{"diloti.hintRequested", "diloti.noHint"}, res["messageCode"])
}

func TestDilotiWebPresenter_ErrorMessage(t *testing.T) {
	p := new(presenter.DilotiWebPresenter)
	d := dilotiGame()
	err := d.PlayerPlay(99, domain.DilotiActionTrail, nil, nil, 0)
	require.Error(t, err)
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(d, err)), &res))
	assert.Equal(t, "diloti.errCardRange", res["messageCode"])
	assert.NotEmpty(t, res["message"])
}

func TestDilotiPresenters_ActionLogOutput(t *testing.T) {
	i18n.SetLang("ja")
	d := dilotiGame()
	dilotiRunToRoundEnd(t, d)

	cui := new(presenter.DilotiCuiPresenter)
	assert.NotEmpty(t, cui.ActionLogOutput(d))

	web := new(presenter.DilotiWebPresenter)
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(web.ActionLogOutput(d)), &res))
	assert.Contains(t, res, "entries")
}

// dilotiRoundEndedButNotOver は「局は終わったが終局していない」盤を返す。
//
// 1 局で目標点に届く配りもあるので、そうでない配りを引くまで配り直す。
func dilotiRoundEndedButNotOver(t *testing.T) *domain.Diloti {
	t.Helper()
	for try := 0; try < 50; try++ {
		d := dilotiGame()
		dilotiRunToRoundEnd(t, d)
		if d.GetPhase() == domain.DilotiPhaseRoundEnd && !d.GetGameEndFlag() {
			return d
		}
	}
	t.Fatal("50 局打っても『終局しない局の終わり』にならない — 前提が崩れている")
	return nil
}

// dilotiRunToRoundEnd は局を最後まで打つ。
func dilotiRunToRoundEnd(t *testing.T, d *domain.Diloti) {
	t.Helper()
	for step := 0; step < 500 && d.GetPhase() == domain.DilotiPhasePlay; step++ {
		if d.IsHumanTurn() {
			h := d.GetHint()
			require.GreaterOrEqual(t, h.Move.HandIdx, 0)
			require.NoError(t, d.PlayerPlay(h.Move.HandIdx, h.Move.Action,
				h.Move.TableIdxs, h.Move.DeclIdxs, h.Move.Value))
			continue
		}
		d.CpuPlay()
	}
	require.NotEqual(t, domain.DilotiPhasePlay, d.GetPhase(), "局が終わらない")
}
