//go:build test

package presenter_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// newHorseForPresenter は開始直後の H.O.R.S.E. を返す。
func newHorseForPresenter(t *testing.T) *domain.Horse {
	t.Helper()
	g := domain.NewDefaultHorse()
	g.Reset()
	return g
}

// horseFoldToHandEnd は人間が降りてハンドを閉じる。
func horseFoldToHandEnd(t *testing.T, g *domain.Horse) {
	t.Helper()
	for range 40 {
		if g.GetGameEndFlag() || g.GetPhase() != domain.HorsePhaseHand {
			return
		}
		if err := g.PlayerAction(domain.HoldemActionFold, 0, 0); err != nil {
			return
		}
	}
	t.Fatal("ハンドが閉じなかった")
}

func TestHorseCuiPresenter_Output(t *testing.T) {
	i18n.SetLang("ja")
	p := &presenter.HorseCuiPresenter{}
	g := newHorseForPresenter(t)

	out := p.Output(g, nil)
	// **訳が引けていることまで見る。** キーの一致だけを見ると、ロケールが
	// 丸ごと欠けていても両辺が生キーで一致して通ってしまう。
	assert.Contains(t, out, "テキサスホールデム", "H の種目名が訳されていない")
	assert.NotContains(t, out, "horse.", "生キーが出ている")
	for i := range g.GetSeatCount() {
		assert.Contains(t, out, g.GetSeatName(i))
	}
	assert.Contains(t, out, "ポット")
}

func TestHorseCuiPresenter_OutputShowsTheError(t *testing.T) {
	i18n.SetLang("ja")
	p := &presenter.HorseCuiPresenter{}
	g := newHorseForPresenter(t)
	out := p.Output(g, assert.AnError)
	assert.Contains(t, out, assert.AnError.Error())
}

func TestHorseCuiPresenter_OutputAtHandEnd(t *testing.T) {
	i18n.SetLang("ja")
	p := &presenter.HorseCuiPresenter{}
	g := newHorseForPresenter(t)
	horseFoldToHandEnd(t, g)
	if g.GetGameEndFlag() {
		t.Skip("1 ハンドで決着した配り")
	}
	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.T("horse.promptHandEnd"))
	assert.NotContains(t, out, "horse.")
}

func TestHorseCuiPresenter_Hint(t *testing.T) {
	i18n.SetLang("ja")
	p := &presenter.HorseCuiPresenter{}
	g := newHorseForPresenter(t)

	// 打っている最中は種目を告げる。
	assert.Contains(t, p.HintOutput(g), "テキサスホールデム")

	horseFoldToHandEnd(t, g)
	hint := p.HintOutput(g)
	assert.NotContains(t, hint, "horse.")
	if g.GetGameEndFlag() {
		assert.Contains(t, hint, i18n.T("horse.hintNone"))
	} else {
		assert.Contains(t, hint, strings.TrimSuffix(i18n.T("horse.hintNextHand"), "\n"))
	}
}

func TestHorseCuiPresenter_ActionLog(t *testing.T) {
	i18n.SetLang("ja")
	p := &presenter.HorseCuiPresenter{}
	g := newHorseForPresenter(t)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// **英語でも生キーが出ない。** 片方のロケールだけ整っていても気付けない。
func TestHorseCuiPresenter_English(t *testing.T) {
	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	out := (&presenter.HorseCuiPresenter{}).Output(newHorseForPresenter(t), nil)
	assert.Contains(t, out, "Texas Hold'em")
	assert.NotContains(t, out, "horse.")
}

func TestHorseWebPresenter_Output(t *testing.T) {
	p := &presenter.HorseWebPresenter{}
	g := newHorseForPresenter(t)

	var out struct {
		Seats []struct {
			ID      int    `json:"id"`
			Name    string `json:"name"`
			IsHuman bool   `json:"isHuman"`
			Chips   int    `json:"chips"`
		} `json:"seats"`
		Phase            int    `json:"phase"`
		Discipline       int    `json:"discipline"`
		DisciplineLetter string `json:"disciplineLetter"`
		DisciplineName   string `json:"disciplineName"`
		HandInDiscipline int    `json:"handInDiscipline"`
		HandNumber       int    `json:"handNumber"`
		CurrentTurn      int    `json:"currentTurn"`
		HumanSeat        int    `json:"humanSeat"`
		IsHumanTurn      bool   `json:"isHumanTurn"`
		Pot              int    `json:"pot"`
		TablePhase       int    `json:"tablePhase"`
		GameEndFlag      bool   `json:"gameEndFlag"`
		WinnerSeat       int    `json:"winnerSeat"`
		Message          string `json:"message"`
		MessageCode      string `json:"messageCode"`
		Config           struct {
			Seats              int `json:"seats"`
			InitialChips       int `json:"initialChips"`
			HandsPerDiscipline int `json:"handsPerDiscipline"`
		} `json:"config"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &out))

	assert.Len(t, out.Seats, g.GetSeatCount())
	for i, s := range out.Seats {
		assert.Equal(t, i, s.ID)
		assert.Equal(t, g.GetSeatName(i), s.Name)
		// **出るのは打っている最中の残高。** 正本はハンドが終わるまで動かない。
		assert.Equal(t, g.GetSeatLiveChips(i), s.Chips)
	}
	assert.Equal(t, "H", out.DisciplineLetter)
	assert.Equal(t, "holdem", out.DisciplineName)
	assert.Equal(t, 1, out.HandInDiscipline)
	assert.Equal(t, 1, out.HandNumber)
	assert.Equal(t, g.GetHumanSeat(), out.HumanSeat)
	assert.Equal(t, g.GetPot(), out.Pot)
	assert.False(t, out.GameEndFlag)
	// **決着していないうちは勝者を書かない。** 0 を出すと席 0 の勝ちに見える。
	assert.Equal(t, -1, out.WinnerSeat)
	assert.Equal(t, g.GetConfig().Seats, out.Config.Seats)
	assert.Equal(t, g.GetConfig().HandsPerDiscipline, out.Config.HandsPerDiscipline)
	assert.Empty(t, out.MessageCode)
}

func TestHorseWebPresenter_OutputError(t *testing.T) {
	p := &presenter.HorseWebPresenter{}
	var out struct {
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(
		[]byte(p.Output(newHorseForPresenter(t), assert.AnError)), &out))
	assert.Equal(t, assert.AnError.Error(), out.Message)
}

// **決着したら勝者を messageCode で返す。** 画面はここを訳して出す。
func TestHorseWebPresenter_OutputAtGameEnd(t *testing.T) {
	p := &presenter.HorseWebPresenter{}
	g := newHorseForPresenter(t)
	for range 400 {
		if g.GetGameEndFlag() {
			break
		}
		if g.GetPhase() == domain.HorsePhaseHandEnd {
			if err := g.NextHand(); err != nil {
				break
			}
			continue
		}
		if err := g.PlayerAction(domain.HoldemActionFold, 0, 0); err != nil {
			break
		}
	}
	if !g.GetGameEndFlag() {
		t.Skip("この配りでは決着まで届かなかった")
	}
	var out struct {
		WinnerSeat    int               `json:"winnerSeat"`
		MessageCode   string            `json:"messageCode"`
		MessageParams map[string]string `json:"messageParams"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &out))
	assert.Equal(t, "horse.result.winner", out.MessageCode)
	assert.GreaterOrEqual(t, out.WinnerSeat, 0)
	assert.Equal(t, g.GetSeatName(g.WinnerSeat()), out.MessageParams["name"])
}

func TestHorseWebPresenter_HintAndLog(t *testing.T) {
	p := &presenter.HorseWebPresenter{}
	g := newHorseForPresenter(t)
	assert.JSONEq(t, p.Output(g, nil), p.HintOutput(g))
	// **棋譜は決着するまで空。** 進行中に見せると相手の手が読めてしまう。
	var log struct {
		Entries []map[string]any `json:"entries"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(g)), &log))
	assert.NotNil(t, log.Entries)
	assert.Empty(t, log.Entries)
}

// newEightGameAtDraw は 2-7 トリプルドローの引き直し待ちまで進めた卓を返す。
//
// **配りに依存させない。** ベットの手数は配りで変わるので、引き直しの番が
// 来るまで合法な手だけを打ち、来なければ配り直す。
func newEightGameAtDraw(t *testing.T) *domain.Horse {
	t.Helper()
	for range 30 {
		g := domain.NewDefaultEightGame()
		g.Reset()
		g.SetDisciplineForTest(domain.HorseTripleDraw)
		for range 200 {
			if g.IsDrawPhase() {
				return g
			}
			if g.GetPhase() != domain.HorsePhaseHand || !g.IsHumanTurn() {
				break
			}
			if err := g.PlayerAction(domain.HoldemActionCall, 0, 0); err != nil {
				if err2 := g.PlayerAction(domain.HoldemActionCheck, 0, 0); err2 != nil {
					break
				}
			}
		}
	}
	t.Fatal("引き直しの番が来なかった")
	return nil
}

// **画面の見出しはゲームの名前。** 2 つのゲームが 1 つの presenter を共有して
// いるので、固定にすると Eight-Game Mix の画面が「H.O.R.S.E.」を名乗る。
func TestHorseCuiPresenter_EightGameHasItsOwnTitle(t *testing.T) {
	i18n.SetLang("ja")
	p := &presenter.HorseCuiPresenter{}

	g := domain.NewDefaultEightGame()
	g.Reset()
	out := p.Output(g, nil)
	assert.Contains(t, out, i18n.T("eightgame.helpTitle"))
	assert.NotContains(t, out, i18n.T("horse.helpTitle"))

	assert.Contains(t, p.Output(newHorseForPresenter(t), nil), i18n.T("horse.helpTitle"))
}

// **引き直しの番はベットの案内では打てない。** ベットの行しか出さないと、
// 何を打てばよいのかが画面のどこにも無い。
func TestHorseCuiPresenter_DrawTurnPromptsForTheDraw(t *testing.T) {
	i18n.SetLang("ja")
	p := &presenter.HorseCuiPresenter{}
	g := newEightGameAtDraw(t)

	out := p.Output(g, nil)
	assert.Contains(t, out, strings.SplitN(i18n.T("horse.promptDraw"), "{{", 2)[0])
	assert.Contains(t, out, i18n.T("horse.promptDrawHelp"))
	assert.NotContains(t, out, i18n.T("horse.promptPlayHelp"), "ベットの案内が残っている")
	// **捨てる札を指す番号まで出す。** `d 0 2` と案内しておいて番号が画面に
	// 無いと、どれが 0 番なのかが分からない。
	assert.Contains(t, out, "[0]")
	assert.Contains(t, out, "[4]")
	assert.NotContains(t, out, "horse.", "生キーが出ている")
}

// ヒントも引き直しの番を知っている。
func TestHorseCuiPresenter_DrawTurnHint(t *testing.T) {
	i18n.SetLang("ja")
	p := &presenter.HorseCuiPresenter{}
	assert.Contains(t, p.HintOutput(newEightGameAtDraw(t)), i18n.T("horse.hintDraw"))
}

// **バリアントと種目の並びはサーバーが出す。** 画面がルート名から決め打つと、
// 5 種目の卓に 8 個の見出しが並ぶ。
func TestHorseWebPresenter_ReportsTheVariantAndRotation(t *testing.T) {
	p := &presenter.HorseWebPresenter{}
	var out struct {
		Variant     int   `json:"variant"`
		Rotation    []int `json:"rotation"`
		IsDrawPhase bool  `json:"isDrawPhase"`
		DrawIndex   int   `json:"drawIndex"`
	}

	g := domain.NewDefaultEightGame()
	g.Reset()
	require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &out))
	assert.Equal(t, int(domain.HorseVariantEightGame), out.Variant)
	assert.Len(t, out.Rotation, 8)
	assert.Equal(t, int(domain.HorseTripleDraw), out.Rotation[7])
	assert.False(t, out.IsDrawPhase, "ホールデムの手が引き直しを名乗っている")
	assert.Zero(t, out.DrawIndex)

	require.NoError(t, json.Unmarshal([]byte(p.Output(newHorseForPresenter(t), nil)), &out))
	assert.Equal(t, int(domain.HorseVariantHorse), out.Variant)
	assert.Len(t, out.Rotation, 5)
}

// 引き直しの番であることと、何回目かが Web にも届く。
func TestHorseWebPresenter_ReportsTheDrawTurn(t *testing.T) {
	p := &presenter.HorseWebPresenter{}
	var out struct {
		IsDrawPhase bool `json:"isDrawPhase"`
		DrawIndex   int  `json:"drawIndex"`
		IsHumanTurn bool `json:"isHumanTurn"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.Output(newEightGameAtDraw(t), nil)), &out))
	assert.True(t, out.IsDrawPhase)
	assert.Contains(t, []int{1, 2, 3}, out.DrawIndex)
	assert.True(t, out.IsHumanTurn, "手番でなければ画面は札を選ばせない")
}
