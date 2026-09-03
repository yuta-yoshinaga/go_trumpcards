//go:build test

package presenter_test

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
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
		MinRaise         int    `json:"minRaise"`
		MaxBetAmount     int    `json:"maxBetAmount"`
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
	assert.Equal(t, g.GetMinRaise(), out.MinRaise)
	assert.Equal(t, g.GetMaxBetAmount(), out.MaxBetAmount)
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

// **最小レイズ幅を画面に出す (#6585)。** 種目ごとに変わるレイズ幅が分からないと、
// CUI プレイヤーは raise コマンドにいくら積めばよいか分からない。
func TestHorseCuiPresenter_OutputShowsMinRaise(t *testing.T) {
	for _, lang := range []string{"ja", "en"} {
		t.Run(lang, func(t *testing.T) {
			i18n.SetLang(lang)
			defer i18n.SetLang("ja")

			m := new(interfaces.MockHorseGame)
			m.On("GetVariant").Return(domain.HorseVariantHorse)
			m.On("GetConfig").Return(domain.HorseConfig{Seats: 2, HandsPerDiscipline: 5, InitialChips: 1000})
			m.On("GetDisciplineLetter").Return("H")
			m.On("GetDiscipline").Return(domain.HorseHoldem)
			m.On("GetHandInDiscipline").Return(1)
			m.On("GetHandNumber").Return(1)
			m.On("GetCommunityCards").Return([]*domain.Card(nil))
			m.On("GetSeatCount").Return(2)
			m.On("GetSeatName", 0).Return("Player 0")
			m.On("GetSeatLiveChips", 0).Return(1000)
			m.On("GetSeatCards", 0).Return([]*domain.Card(nil))
			m.On("GetSeatName", 1).Return("Player 1")
			m.On("GetSeatLiveChips", 1).Return(1000)
			m.On("GetSeatCards", 1).Return([]*domain.Card(nil))
			m.On("GetCurrentTurn").Return(0)
			m.On("GetPhase").Return(domain.HorsePhaseHand)
			m.On("GetGameEndFlag").Return(false)
			m.On("IsDrawPhase").Return(false)
			m.On("GetPot").Return(150)
			m.On("GetToCall").Return(30)

			minRaise := 60
			m.On("GetMinRaise").Return(minRaise)
			m.On("GetMaxBetAmount").Return(0)

			p := &presenter.HorseCuiPresenter{}
			out := p.Output(m, nil)

			// **期待値を i18n から組み立てず、モックの数値そのものを見る。**
			assert.Contains(t, out, strconv.Itoa(minRaise), "最小レイズ額が出力に含まれていない")
			// **プレースホルダが展開されず残っていないことを見る。**
			assert.NotContains(t, out, "{{", "未展開のプレースホルダが残っている")
			assert.NotContains(t, out, "}}", "未展開のプレースホルダが残っている")
			assert.NotContains(t, out, "horse.", "生キーが出ている")
			m.AssertExpectations(t)
		})
	}
}

// **ポットリミット上限を画面に出す (#6622)。** PLO ラウンドなどでポット超過額が
// 分からないと、CUI プレイヤーは超過額を打ってサーバに弾かれるまで分からない。
// 一方で固定リミット (HORSE) では上限を出さず、従来の表示を保つ。
func TestHorseCuiPresenter_OutputShowsMaxBetAmount(t *testing.T) {
	for _, lang := range []string{"ja", "en"} {
		t.Run(lang, func(t *testing.T) {
			i18n.SetLang(lang)
			defer i18n.SetLang("ja")

			setupMock := func(maxBet int) *interfaces.MockHorseGame {
				m := new(interfaces.MockHorseGame)
				m.On("GetVariant").Return(domain.HorseVariantEightGame)
				m.On("GetConfig").Return(domain.HorseConfig{Seats: 2, HandsPerDiscipline: 5, InitialChips: 1000, Variant: domain.HorseVariantEightGame})
				m.On("GetDisciplineLetter").Return("PLO")
				m.On("GetDiscipline").Return(domain.HorsePLOmaha)
				m.On("GetHandInDiscipline").Return(1)
				m.On("GetHandNumber").Return(6)
				m.On("GetCommunityCards").Return([]*domain.Card(nil))
				m.On("GetSeatCount").Return(2)
				m.On("GetSeatName", 0).Return("Player 0")
				m.On("GetSeatLiveChips", 0).Return(1000)
				m.On("GetSeatCards", 0).Return([]*domain.Card(nil))
				m.On("GetSeatName", 1).Return("Player 1")
				m.On("GetSeatLiveChips", 1).Return(1000)
				m.On("GetSeatCards", 1).Return([]*domain.Card(nil))
				m.On("GetCurrentTurn").Return(0)
				m.On("GetPhase").Return(domain.HorsePhaseHand)
				m.On("GetGameEndFlag").Return(false)
				m.On("IsDrawPhase").Return(false)
				m.On("GetPot").Return(100)
				m.On("GetToCall").Return(20)
				m.On("GetMinRaise").Return(20)
				m.On("GetMaxBetAmount").Return(maxBet)
				return m
			}

			p := &presenter.HorseCuiPresenter{}

			// 1. maxBet > 0 のとき (PLOラウンド): 上限が表示される
			m1 := setupMock(120)
			out1 := p.Output(m1, nil)
			assert.Contains(t, out1, "120", "上限額が出力に含まれていない")
			if lang == "ja" {
				assert.Contains(t, out1, "上限: 120")
			} else {
				assert.Contains(t, out1, "max: 120")
			}
			assert.NotContains(t, out1, "{{", "未展開のプレースホルダが残っている")
			assert.NotContains(t, out1, "}}", "未展開のプレースホルダが残っている")
			assert.NotContains(t, out1, "horse.", "生キーが出ている")
			m1.AssertExpectations(t)

			// 2. 負のコントロール: 値を変えると表示も変わる
			m2 := setupMock(250)
			out2 := p.Output(m2, nil)
			assert.Contains(t, out2, "250", "変更後の上限額が出力に含まれていない")
			if lang == "ja" {
				assert.Contains(t, out2, "上限: 250")
			} else {
				assert.Contains(t, out2, "max: 250")
			}
			m2.AssertExpectations(t)

			// 3. 負のコントロール: maxBet == 0 (固定リミット/HORSE) では上限が出ない
			m3 := setupMock(0)
			out3 := p.Output(m3, nil)
			if lang == "ja" {
				assert.NotContains(t, out3, "上限", "固定リミットで上限が表示されている")
			} else {
				assert.NotContains(t, out3, "max:", "固定リミットで上限が表示されている")
			}
			m3.AssertExpectations(t)
		})
	}
}

func TestHorseWebPresenter_OutputMaxBetAmount(t *testing.T) {
	setupMock := func(maxBet int) *interfaces.MockHorseGame {
		m := new(interfaces.MockHorseGame)
		m.On("GetConfig").Return(domain.HorseConfig{Seats: 2, HandsPerDiscipline: 5, InitialChips: 1000})
		m.On("GetSeatCount").Return(2)
		m.On("GetSeatName", 0).Return("Player 0")
		m.On("GetSeatLiveChips", 0).Return(1000)
		m.On("GetSeatCards", 0).Return([]*domain.Card(nil))
		m.On("GetSeatIsHuman", 0).Return(true)
		m.On("GetSeatName", 1).Return("Player 1")
		m.On("GetSeatLiveChips", 1).Return(1000)
		m.On("GetSeatCards", 1).Return([]*domain.Card(nil))
		m.On("GetSeatIsHuman", 1).Return(false)
		m.On("GetPhase").Return(domain.HorsePhaseHand)
		m.On("GetDiscipline").Return(domain.HorsePLOmaha)
		m.On("GetDisciplineLetter").Return("PLO")
		m.On("GetHandInDiscipline").Return(1)
		m.On("GetHandNumber").Return(6)
		m.On("GetCurrentTurn").Return(0)
		m.On("GetHumanSeat").Return(0)
		m.On("IsHumanTurn").Return(true)
		m.On("GetCommunityCards").Return([]*domain.Card(nil))
		m.On("GetPot").Return(100)
		m.On("GetToCall").Return(20)
		m.On("GetMinRaise").Return(20)
		m.On("GetMaxBetAmount").Return(maxBet)
		m.On("GetTablePhase").Return(0)
		m.On("GetVariant").Return(domain.HorseVariantEightGame)
		m.On("GetRotation").Return([]domain.HorseDiscipline{domain.HorsePLOmaha})
		m.On("IsDrawPhase").Return(false)
		m.On("GetDrawIndex").Return(0)
		m.On("GetGameEndFlag").Return(false)
		return m
	}

	p := &presenter.HorseWebPresenter{}

	// maxBet == 150
	m1 := setupMock(150)
	var out1 struct {
		MaxBetAmount int `json:"maxBetAmount"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.Output(m1, nil)), &out1))
	assert.Equal(t, 150, out1.MaxBetAmount)

	// 負のコントロール: maxBet == 0
	m2 := setupMock(0)
	var out2 struct {
		MaxBetAmount int `json:"maxBetAmount"`
	}
	require.NoError(t, json.Unmarshal([]byte(p.Output(m2, nil)), &out2))
	assert.Equal(t, 0, out2.MaxBetAmount)
}
