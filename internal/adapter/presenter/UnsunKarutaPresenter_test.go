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

// unsunKarutaGame は配り終えた 8 人卓を返す。
func unsunKarutaGame() *domain.UnsunKaruta {
	g := domain.NewDefaultUnsunKaruta()
	g.Reset()
	return g
}

// unsunKarutaAdvanceToHuman は人間の手番かトリック終了まで CPU を回す。
func unsunKarutaAdvanceToHuman(g *domain.UnsunKaruta) {
	for i := 0; i < 64; i++ {
		if g.GetPhase() != domain.UnsunKarutaPhasePlay || g.IsHumanTurn() {
			return
		}
		g.CpuPlay()
	}
}

func TestUnsunKarutaCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.UnsunKarutaCuiPresenter)

	t.Run("shows the deal, the trump and both teams", func(t *testing.T) {
		g := unsunKarutaGame()
		out := p.Output(g, nil)
		// **訳が引けていることまで見る。** キーの一致だけを見ると、ロケールが
		// 丸ごと欠けていても両辺が生キーで一致して通ってしまう。
		assert.Contains(t, out, "うんすんカルタ")
		assert.NotContains(t, out, "unsunkaruta.", "生キーが出ている")
		// **切り札は返した 1 枚で決まる。** 数札の強弱がスートで逆になるので、
		// どのスートが切り札かが出ていないと盤面が読めない。
		assert.Contains(t, out, strings.SplitN(i18n.T("unsunkaruta.trump"), "{{", 2)[0])
		assert.Contains(t, out, i18n.T("unsunkaruta.suit."+domain.UnsunKarutaSuitName(g.GetTrumpSuit())))
		assert.Contains(t, out, strings.SplitN(i18n.T("unsunkaruta.teamLine"), "{{", 2)[0])
	})

	t.Run("lists the human hand with indices and the playable cards", func(t *testing.T) {
		g := unsunKarutaGame()
		unsunKarutaAdvanceToHuman(g)
		if g.GetPhase() != domain.UnsunKarutaPhasePlay || !g.IsHumanTurn() {
			t.Skip("配りによっては人間の手番の前にトリックが揃う")
		}
		out := p.Output(g, nil)
		assert.Contains(t, out, "[0]")
		// 出せる札を出す。Web は playableIndices でボタンを絞るのに、CUI に
		// それが無いと総当たりで探すことになる。
		assert.Contains(t, out, strings.SplitN(i18n.T("unsunkaruta.playableList"), "{{", 2)[0])
		assert.Contains(t, out, i18n.T("unsunkaruta.promptPlayHelp"))
	})

	// **フォロー義務は宣言で生まれる。** 出さないと、なぜ札が絞られるのかが
	// 端末から読めない。
	t.Run("announces the follow obligation only after a declaration", func(t *testing.T) {
		g := unsunKarutaGame()
		unsunKarutaAdvanceToHuman(g)
		if !g.IsHumanTurn() || g.GetLeadPlayerIdx() != g.GetCurrentPlayerIdx() {
			t.Skip("人間のリードでないディール")
		}
		require.False(t, g.IsMustFollow())
		assert.NotContains(t, p.Output(g, nil), i18n.T("unsunkaruta.mustFollow"))

		valid := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
		require.NotEmpty(t, valid)
		require.NoError(t, g.PlayerPlay(valid[0], true))
		require.True(t, g.IsMustFollow())
		assert.Contains(t, p.Output(g, nil), i18n.T("unsunkaruta.mustFollow"))
	})

	// 宣言できるのはリードのときだけ。案内も同じ条件で出る。
	t.Run("offers the declaration only on a lead", func(t *testing.T) {
		g := unsunKarutaGame()
		unsunKarutaAdvanceToHuman(g)
		if g.GetPhase() != domain.UnsunKarutaPhasePlay || !g.IsHumanTurn() {
			t.Skip("配りによっては人間の手番の前にトリックが揃う")
		}
		out := p.Output(g, nil)
		if g.CanDeclare() {
			assert.Contains(t, out, i18n.T("unsunkaruta.promptDeclare"))
		} else {
			assert.NotContains(t, out, i18n.T("unsunkaruta.promptDeclare"))
		}
	})

	// **誰が取ったのかを名指しする。** 8 枚並んだ盤面から勝者を読むには
	// 「切り札か、無ければ台札の最強」を毎回自分で解くことになる。
	t.Run("names the seat that took the trick", func(t *testing.T) {
		g := unsunKarutaGame()
		unsunKarutaPlayTrick(t, g)
		require.Equal(t, domain.UnsunKarutaPhaseTrickEnd, g.GetPhase())
		winner := g.GetLastTrickWinner()
		require.GreaterOrEqual(t, winner, 0)
		out := p.Output(g, nil)
		assert.Contains(t, out, i18n.Tf("unsunkaruta.promptTrickEndWinner",
			"name", unsunKarutaSeatName(winner)))
		assert.Contains(t, out, i18n.T("unsunkaruta.promptTrickEndHelp"))
	})

	t.Run("errors are shown", func(t *testing.T) {
		g := unsunKarutaGame()
		assert.Contains(t, p.Output(g, assert.AnError), assert.AnError.Error())
	})

	t.Run("announces the winning team, and a draw", func(t *testing.T) {
		g := unsunKarutaGame()
		unsunKarutaPlayMatch(t, g)
		out := p.Output(g, nil)
		if g.GetWinnerTeam() < 0 {
			assert.Contains(t, out, i18n.T("unsunkaruta.gameEndDraw"))
			return
		}
		assert.Contains(t, out, i18n.Tf("unsunkaruta.gameEnd", "team", itoa(g.GetWinnerTeam())))
	})

	// **英語も訳が引ける。** 反対の言語が漏れていれば生キーが出る。
	t.Run("renders in english too", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		out := p.Output(unsunKarutaGame(), nil)
		assert.NotContains(t, out, "unsunkaruta.")
		assert.NotContains(t, out, "うんすんカルタ", "日本語が漏れている")
	})
}

func TestUnsunKarutaCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.UnsunKarutaCuiPresenter)

	g := unsunKarutaGame()
	unsunKarutaAdvanceToHuman(g)
	out := p.HintOutput(g)
	assert.NotContains(t, out, "unsunkaruta.", "生キーが出ている")
	if hint := g.GetHint(); hint != nil && len(hint.CardIndices) > 0 {
		assert.Contains(t, out, "[")
	}
}

func TestUnsunKarutaWebPresenter_Output(t *testing.T) {
	i18n.SetLang("ja")
	p := new(presenter.UnsunKarutaWebPresenter)
	g := unsunKarutaGame()
	unsunKarutaAdvanceToHuman(g)

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &res))

	assert.Equal(t, float64(1), res["roundNumber"])
	assert.Equal(t, float64(domain.UnsunKarutaTrickCount), res["trickCount"])
	assert.Len(t, res["players"], domain.UnsunKarutaPlayerCnt)
	assert.Len(t, res["teamTricks"], domain.UnsunKarutaTeamCnt)
	assert.Len(t, res["teamScores"], domain.UnsunKarutaTeamCnt)
	// **切り札のスート名は文字列で運ぶ。** 番号だけだと i18n が引けない。
	assert.NotEmpty(t, res["trumpSuitName"])

	// **75 枚は 52 枚デッキの記号で書けない。** 手続き描画の自己記述子
	// (deck / glyph / label) が付いていないと、画面が白札になる。
	players, ok := res["players"].([]any)
	require.True(t, ok)
	human, ok := players[0].(map[string]any)
	require.True(t, ok)
	cards, ok := human["cards"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, cards)
	first, ok := cards[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "unsun", first["deck"])
	assert.NotEmpty(t, first["glyph"])
	assert.NotEmpty(t, first["label"])

	// **味方は席番号では分からない。** humanTeam が無いと画面が組を出せない。
	assert.Equal(t, float64(domain.UnsunKarutaTeamOf(0)), res["humanTeam"])
	// CPU の手札は伏せたまま。枚数だけは見える。
	cpu, ok := players[1].(map[string]any)
	require.True(t, ok)
	assert.Empty(t, cpu["cards"])
	assert.Positive(t, cpu["cardCount"])

	// 配った直後は全員 9 枚。
	var fresh map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(unsunKarutaGame(), nil)), &fresh))
	freshCpu := fresh["players"].([]any)[1].(map[string]any)
	assert.Equal(t, float64(domain.UnsunKarutaHandSize), freshCpu["cardCount"])
}

// 丸物は赤、長物は黒。数札の強さが逆になる 2 系統を色で見分ける。
func TestUnsunKarutaWebPresenter_RoundSuitsAreRed(t *testing.T) {
	p := new(presenter.UnsunKarutaWebPresenter)
	g := unsunKarutaGame()
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &res))
	players := res["players"].([]any)
	cards := players[0].(map[string]any)["cards"].([]any)
	for _, c := range cards {
		card := c.(map[string]any)
		design, _ := card["design"].(string)
		want := "black"
		// design は 52 枚デッキの名前に写る: パオ=SPADE, イス=CLOVER が長物。
		if design != "SPADE" && design != "CLOVER" {
			want = "red"
		}
		assert.Equal(t, want, card["color"], "design=%v", design)
	}
}

func TestUnsunKarutaWebPresenter_PlayableIndicesAreEmptyOffTurn(t *testing.T) {
	p := new(presenter.UnsunKarutaWebPresenter)
	g := unsunKarutaGame()
	// トリック終了フェーズでは 1 枚も出せない。
	unsunKarutaPlayTrick(t, g)
	require.Equal(t, domain.UnsunKarutaPhaseTrickEnd, g.GetPhase())
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &res))
	assert.Empty(t, res["playableIndices"])
	assert.NotNil(t, res["playableIndices"], "null ではなく空配列で運ぶ")
}

func TestUnsunKarutaWebPresenter_MessageCodes(t *testing.T) {
	p := new(presenter.UnsunKarutaWebPresenter)
	g := unsunKarutaGame()

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &res))
	assert.Contains(t, res["messageCode"], "unsunkaruta.playPhase.")

	unsunKarutaPlayTrick(t, g)
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &res))
	assert.Equal(t, "unsunkaruta.trickEnd", res["messageCode"])

	// **ヒントは頼まれたときだけ名乗る。** Output と HintOutput で
	// messageCode が違うから、画面がバナーを出し分けられる。
	require.NoError(t, json.Unmarshal([]byte(p.HintOutput(g)), &res))
	assert.Contains(t, []any{"unsunkaruta.hintRequested", "unsunkaruta.noHint"}, res["messageCode"])
}

func TestUnsunKarutaWebPresenter_ErrorMessage(t *testing.T) {
	p := new(presenter.UnsunKarutaWebPresenter)
	g := unsunKarutaGame()
	unsunKarutaAdvanceToHuman(g)
	require.True(t, g.IsHumanTurn(), "人間の手番まで進んでいない")
	err := g.PlayerPlay(999, false)
	require.Error(t, err)
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, err)), &res))
	assert.Equal(t, "unsunkaruta.errCardRange", res["messageCode"])
	assert.NotEmpty(t, res["message"])
}

func TestUnsunKarutaPresenters_ActionLogOutput(t *testing.T) {
	i18n.SetLang("ja")
	g := unsunKarutaGame()
	unsunKarutaPlayTrick(t, g)

	cui := new(presenter.UnsunKarutaCuiPresenter)
	assert.NotEmpty(t, cui.ActionLogOutput(g))

	web := new(presenter.UnsunKarutaWebPresenter)
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(web.ActionLogOutput(g)), &res))
	assert.Contains(t, res, "entries")
}

// unsunKarutaPlayTrick は 1 トリックぶん、合法手の先頭を出し続けて解決する。
//
// **解決まで含める。** インタラクターは 8 枚目が出た時点で ResolveTrick を
// 呼ぶので、プレゼンターが未解決のトリック終了を見ることは本番では起きない。
func unsunKarutaPlayTrick(t *testing.T, g *domain.UnsunKaruta) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if g.GetPhase() != domain.UnsunKarutaPhasePlay {
			g.ResolveTrick()
			return
		}
		if g.IsHumanTurn() {
			valid := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
			require.NotEmpty(t, valid, "出せる札が 1 枚も無い")
			require.NoError(t, g.PlayerPlay(valid[0], false))
			continue
		}
		g.CpuPlay()
	}
	t.Fatal("トリックが終わらない")
}

// unsunKarutaPlayMatch はマッチを終局まで打ち切る。
func unsunKarutaPlayMatch(t *testing.T, g *domain.UnsunKaruta) {
	t.Helper()
	for step := 0; step < 4000 && !g.GetGameEndFlag(); step++ {
		switch g.GetPhase() {
		case domain.UnsunKarutaPhasePlay:
			if g.IsHumanTurn() {
				valid := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, valid, "出せる札が 1 枚も無い")
				require.NoError(t, g.PlayerPlay(valid[0], false))
				continue
			}
			g.CpuPlay()
		case domain.UnsunKarutaPhaseTrickEnd:
			g.ResolveTrick()
			g.NextTrick()
		case domain.UnsunKarutaPhaseRoundEnd:
			g.ScoreRound()
			if g.GetGameEndFlag() {
				return
			}
			g.NextRound()
		case domain.UnsunKarutaPhaseGameEnd:
			return
		}
	}
	require.True(t, g.GetGameEndFlag(), "マッチが終わらない")
}

// itoa は strconv.Itoa の薄い別名 (テスト内の可読性のため)。
func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf []byte
	for v > 0 {
		buf = append([]byte{byte('0' + v%10)}, buf...)
		v /= 10
	}
	if neg {
		return "-" + string(buf)
	}
	return string(buf)
}

// unsunKarutaSeatName は CUI の席名 ("あなた" / "CPU n") を返す。
func unsunKarutaSeatName(idx int) string {
	if idx == 0 {
		return i18n.T("cuiPlayerYou")
	}
	return i18n.Tf("cuiPlayerCpu", "idx", itoa(idx))
}
