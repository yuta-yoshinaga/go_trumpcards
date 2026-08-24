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

// quodlibetGame は配り終えた卓を返す (席 0 が第 1 ディールの親)。
func quodlibetGame() *domain.Quodlibet {
	q := domain.NewDefaultQuodlibet()
	q.Reset()
	return q
}

// quodlibetStartDeal はコントラクトを決めて人間の手番まで進める。
//
// **その輪に入るまでは選べない。** 第 3 の輪の種目を第 1 ディールで指定しても
// 弾かれるので、必要なら先に輪を消化する。
func quodlibetStartDeal(t *testing.T, q *domain.Quodlibet, contract int) {
	t.Helper()
	quodlibetReachWheel(t, q, domain.QuodlibetRoundOf(contract))
	require.NoError(t, q.SelectContract(contract))
	for i := 0; i < 64; i++ {
		if q.GetPhase() != domain.QuodlibetPhasePlay || q.IsHumanTurn() {
			return
		}
		q.CpuPlay()
	}
}

func TestQuodlibetCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.QuodlibetCuiPresenter)

	t.Run("shows the deal, the wheel and the penalty column", func(t *testing.T) {
		out := p.Output(quodlibetGame(), nil)
		// **訳が引けていることまで見る。** キー一致だけを見ると、ロケールが
		// 丸ごと欠けていても両辺が生キーで一致して通ってしまう。
		assert.Contains(t, out, "クオドリベット")
		assert.NotContains(t, out, "quodlibet.", "生キーが出ている")
		assert.Contains(t, out, strings.SplitN(i18n.T("quodlibet.deal"), "{{", 2)[0])
		assert.Contains(t, out, strings.SplitN(i18n.T("quodlibet.playerLine"), "{{", 2)[0])
	})

	// **選べるのはこの輪の残りだけ。** 全 12 種目を並べると、押せない選択肢を
	// 勧めることになる。
	t.Run("lists only the contracts left in this wheel", func(t *testing.T) {
		out := p.Output(quodlibetGame(), nil)
		assert.Contains(t, out, i18n.T("quodlibet.contractName.plus"))
		assert.Contains(t, out, i18n.T("quodlibet.contractName.alarich"))
		assert.NotContains(t, out, i18n.T("quodlibet.contractName.noReds"), "第 2 の輪が出ている")
		assert.NotContains(t, out, i18n.T("quodlibet.contractName.snack"), "第 3 の輪が出ている")
	})

	// **絵札は A/J/Q/K で出す。** 数値のままだと 7〜10 と地続きに見え、
	// Q と J にだけ点が付く種目で何を避ければよいのか読めない。
	t.Run("prints court cards by their face label", func(t *testing.T) {
		q := quodlibetGame()
		out := p.Output(q, nil)
		assert.NotRegexp(t, `[♠♣♥♦](11|12|13)\b`, out, "絵札が数値で出ている")
		assert.NotRegexp(t, `[♠♣♥♦]1\b`, out, "エースが数値で出ている")
	})

	t.Run("play phase lists the playable cards", func(t *testing.T) {
		q := quodlibetGame()
		quodlibetStartDeal(t, q, domain.QuodlibetMinus)
		if q.GetPhase() != domain.QuodlibetPhasePlay || !q.IsHumanTurn() {
			t.Skip("配りによっては人間の手番の前にトリックが揃う")
		}
		out := p.Output(q, nil)
		assert.Contains(t, out, strings.SplitN(i18n.T("quodlibet.playableList"), "{{", 2)[0])
		assert.Contains(t, out, i18n.T("quodlibet.promptPlayHelp"))
	})

	// **四分と小食いは場が違う。** 重ねや並びを出さないと、なぜその札しか
	// 出せないのかが読めない。
	t.Run("shows the shed layout for a shedding contract", func(t *testing.T) {
		q := quodlibetGame()
		quodlibetStartDeal(t, q, domain.QuodlibetSnack)
		out := p.Output(q, nil)
		assert.Contains(t, out, strings.SplitN(i18n.T("quodlibet.tableRow"), "{{", 2)[0])
	})

	t.Run("does not show a shed layout for a trick contract", func(t *testing.T) {
		q := quodlibetGame()
		quodlibetStartDeal(t, q, domain.QuodlibetPlus)
		out := p.Output(q, nil)
		assert.NotContains(t, out, i18n.T("quodlibet.stackEmpty"), "トリック種目に重ねを出している")
	})

	t.Run("errors are shown", func(t *testing.T) {
		out := p.Output(quodlibetGame(), assert.AnError)
		assert.Contains(t, out, assert.AnError.Error())
	})

	// **勝つのは罰点が最少の人。** 名前だけ出すと勝負の向きが読めない。
	t.Run("announces the seats on the fewest penalty points", func(t *testing.T) {
		q := quodlibetGame()
		quodlibetPlayMatch(t, q)
		out := p.Output(q, nil)
		assert.Contains(t, out, strings.SplitN(i18n.T("quodlibet.gameEnd"), "{{", 2)[0])
	})

	// **英語も訳が引ける。** 反対の言語が漏れていれば生キーが出る。
	t.Run("renders in english too", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		out := p.Output(quodlibetGame(), nil)
		assert.NotContains(t, out, "quodlibet.")
		assert.NotContains(t, out, "クオドリベット", "日本語が漏れている")
	})
}

func TestQuodlibetCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("ja")
	p := new(presenter.QuodlibetCuiPresenter)

	// 選択フェーズでは種目を勧める。
	out := p.HintOutput(quodlibetGame())
	assert.NotContains(t, out, "quodlibet.", "生キーが出ている")
	assert.Contains(t, out, strings.SplitN(i18n.T("quodlibet.hintContract"), "{{", 2)[0])

	q := quodlibetGame()
	quodlibetStartDeal(t, q, domain.QuodlibetMinus)
	if q.GetPhase() == domain.QuodlibetPhasePlay && q.IsHumanTurn() {
		out = p.HintOutput(q)
		assert.NotContains(t, out, "quodlibet.")
		assert.Contains(t, out, "[")
	}
}

func TestQuodlibetWebPresenter_Output(t *testing.T) {
	i18n.SetLang("ja")
	p := new(presenter.QuodlibetWebPresenter)
	q := quodlibetGame()

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(q, nil)), &res))

	assert.Equal(t, domain.QuodlibetPhaseSelectContract, res["phase"])
	assert.Equal(t, float64(domain.QuodlibetTotalDeals), res["totalDeals"])
	assert.Equal(t, float64(1), res["roundNumber"])
	assert.Len(t, res["players"], domain.QuodlibetPlayerCnt)
	assert.Equal(t, true, res["isContractPhase"])
	// **選べるのはこの輪の 4 種目だけ。**
	assert.Len(t, res["availableContracts"], domain.QuodlibetContractsPerRound)
	assert.Len(t, res["availableContractNames"], domain.QuodlibetContractsPerRound)
	// null ではなく空配列で運ぶ。
	assert.NotNil(t, res["playableIndices"])
	assert.NotNil(t, res["stack"])
	assert.NotNil(t, res["tablePlaced"])
	assert.NotNil(t, res["dealHistory"])
}

// **第 3 の輪では手札の見え方そのものが規則。** 出し分けないと、開いたズボンも
// 狩猟もただのマイナスになる。
func TestQuodlibetWebPresenter_ThirdWheelChangesVisibility(t *testing.T) {
	p := new(presenter.QuodlibetWebPresenter)

	read := func(contract int) (int, int) {
		q := quodlibetGame()
		quodlibetStartDeal(t, q, contract)
		var res map[string]any
		require.NoError(t, json.Unmarshal([]byte(p.Output(q, nil)), &res))
		players := res["players"].([]any)
		human := len(players[0].(map[string]any)["cards"].([]any))
		other := len(players[1].(map[string]any)["cards"].([]any))
		return human, other
	}

	// 通常の種目: 自分だけが見える。
	human, other := read(domain.QuodlibetPlus)
	assert.Positive(t, human)
	assert.Zero(t, other)

	// 開いたズボン: 自分「だけ」が見えない。
	human, other = read(domain.QuodlibetOpen)
	assert.Zero(t, human, "自分の手札が見えてしまっている")
	assert.Positive(t, other, "他人の手札が伏せられている")

	// 狩猟: 全員が見える。
	human, other = read(domain.QuodlibetHunt)
	assert.Positive(t, human)
	assert.Positive(t, other)
}

// **パスできるのは出せる札が 1 枚も無いときだけ。**
func TestQuodlibetWebPresenter_CanPassOnlyWhenStuck(t *testing.T) {
	p := new(presenter.QuodlibetWebPresenter)

	q := quodlibetGame()
	quodlibetStartDeal(t, q, domain.QuodlibetMinus)
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(q, nil)), &res))
	assert.Equal(t, false, res["canPass"], "トリック種目でパスを勧めている")

	q = quodlibetGame()
	quodlibetStartDeal(t, q, domain.QuodlibetSnack)
	require.NoError(t, json.Unmarshal([]byte(p.Output(q, nil)), &res))
	idx, _ := res["playableIndices"].([]any)
	if len(idx) > 0 {
		assert.Equal(t, false, res["canPass"], "出せる札があるのにパスを勧めている")
	}
}

func TestQuodlibetWebPresenter_MessageCodes(t *testing.T) {
	p := new(presenter.QuodlibetWebPresenter)
	q := quodlibetGame()

	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(q, nil)), &res))
	assert.Equal(t, "quodlibet.selectContract", res["messageCode"])

	quodlibetStartDeal(t, q, domain.QuodlibetMinus)
	require.NoError(t, json.Unmarshal([]byte(p.Output(q, nil)), &res))
	assert.Contains(t, []any{"quodlibet.playPhase", "quodlibet.dealEnd"}, res["messageCode"])

	// **ヒントは頼まれたときだけ名乗る。**
	require.NoError(t, json.Unmarshal([]byte(p.HintOutput(q)), &res))
	assert.Contains(t, []any{"quodlibet.hintRequested", "quodlibet.noHint"}, res["messageCode"])
}

func TestQuodlibetWebPresenter_ErrorMessage(t *testing.T) {
	p := new(presenter.QuodlibetWebPresenter)
	q := quodlibetGame()
	// 別の輪の種目は選べない。
	err := q.SelectContract(domain.QuodlibetSnack)
	require.Error(t, err)
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(p.Output(q, err)), &res))
	assert.Equal(t, "quodlibet.errContractUnavailable", res["messageCode"])
	assert.NotEmpty(t, res["message"])
}

func TestQuodlibetPresenters_ActionLogOutput(t *testing.T) {
	i18n.SetLang("ja")
	q := quodlibetGame()
	quodlibetStartDeal(t, q, domain.QuodlibetMinus)

	cui := new(presenter.QuodlibetCuiPresenter)
	assert.NotEmpty(t, cui.ActionLogOutput(q))

	web := new(presenter.QuodlibetWebPresenter)
	var res map[string]any
	require.NoError(t, json.Unmarshal([]byte(web.ActionLogOutput(q)), &res))
	assert.Contains(t, res, "entries")
}

// quodlibetPlayMatch はマッチを終局まで打ち切る。
func quodlibetPlayMatch(t *testing.T, q *domain.Quodlibet) {
	t.Helper()
	for step := 0; step < 6000 && !q.GetGameEndFlag(); step++ {
		switch q.GetPhase() {
		case domain.QuodlibetPhaseSelectContract:
			if q.IsHumanTurn() {
				avail := q.GetAvailableContracts()
				require.NotEmpty(t, avail)
				require.NoError(t, q.SelectContract(avail[0]))
				continue
			}
			q.CpuSelectContract()
		case domain.QuodlibetPhasePlay:
			if q.IsHumanTurn() {
				idx := -1
				if valid := q.GetPlayableIndices(q.GetCurrentTurn()); len(valid) > 0 {
					idx = valid[0]
				}
				require.NoError(t, q.PlayerPlay(idx))
				continue
			}
			q.CpuPlay()
		case domain.QuodlibetPhaseDealEnd:
			q.NextDeal()
		default:
			return
		}
	}
	require.True(t, q.GetGameEndFlag(), "マッチが終わらない")
}

// quodlibetReachWheel は指定の輪 (0-2) に入るまでディールを消化する。
func quodlibetReachWheel(t *testing.T, q *domain.Quodlibet, wheel int) {
	t.Helper()
	for step := 0; step < 6000; step++ {
		if q.GetDealNumber()/domain.QuodlibetContractsPerRound == wheel {
			return
		}
		switch q.GetPhase() {
		case domain.QuodlibetPhaseSelectContract:
			if q.IsHumanTurn() {
				avail := q.GetAvailableContracts()
				require.NotEmpty(t, avail)
				require.NoError(t, q.SelectContract(avail[0]))
				continue
			}
			q.CpuSelectContract()
		case domain.QuodlibetPhasePlay:
			if q.IsHumanTurn() {
				idx := -1
				if valid := q.GetPlayableIndices(q.GetCurrentTurn()); len(valid) > 0 {
					idx = valid[0]
				}
				require.NoError(t, q.PlayerPlay(idx))
				continue
			}
			q.CpuPlay()
		case domain.QuodlibetPhaseDealEnd:
			q.NextDeal()
		default:
			t.Fatalf("輪 %d に届く前に終局した", wheel)
		}
	}
	t.Fatalf("輪 %d に届かない", wheel)
}
