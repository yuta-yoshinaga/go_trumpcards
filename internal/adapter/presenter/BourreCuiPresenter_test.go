package presenter_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestBourreCuiPresenter_Output(t *testing.T) {
	bg := newBourreAllCpu()
	bg.Reset()
	p := new(presenter.BourreCuiPresenter)

	// Decide phase render
	out := p.Output(bg, nil)
	assert.Contains(t, out, "==========")

	// Drive to game end, rendering along the way
	for i := 0; i < 300000 && !bg.GetGameEndFlag(); i++ {
		if bg.GetPhase() == domain.BourrePhaseRoundEnd {
			bg.NextHand()
			continue
		}
		_ = p.Output(bg, nil)
		bg.CpuPlay()
	}
	endOut := p.Output(bg, nil)
	assert.Contains(t, endOut, "==========")
	assert.NotEmpty(t, strings.TrimSpace(endOut))
}

func TestBourreCuiPresenter_Error(t *testing.T) {
	bg := newBourreAllCpu()
	bg.Reset()
	p := new(presenter.BourreCuiPresenter)
	out := p.Output(bg, assertErr{})
	assert.Contains(t, out, "boom")
}

type assertErr struct{}

func (assertErr) Error() string { return "boom" }

func TestBourreCuiPresenter_ActionLog(t *testing.T) {
	bg := newBourreAllCpu()
	bg.Reset()
	p := new(presenter.BourreCuiPresenter)
	assert.NotNil(t, p.ActionLogOutput(bg))
}

// **CPU 名も i18n と太字を通す。**bourreName は人間だけ i18n を通し、CPU は
// `fmt.Sprintf("CPU %d")` の英語リテラルを返していた (#4719)。
func TestBourreCuiPresenter_CpuNamesGoThroughTheSharedHelper(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(false)
	defer color.SetNoColor(origNoColor)

	bg := newBourreAllCpu()
	bg.Reset()
	p := new(presenter.BourreCuiPresenter)
	// **リセット直後は bourreName を通らない。**トリックが場に出るまで進める。
	// 上限はリポジトリの慣習どおり 1000。トリックは数手で場に出るので、
	// これで届かないなら収束していないほうを疑う。
	var out string
	for i := 0; i < 1000; i++ {
		out = p.Output(bg, nil)
		if strings.Contains(out, i18n.T("cuiPlayerCpu")) || len(bg.GetCurrentTrick()) > 0 {
			break
		}
		if bg.GetPhase() == domain.BourrePhaseRoundEnd {
			bg.NextHand()
			continue
		}
		if bg.GetGameEndFlag() {
			break
		}
		bg.CpuPlay()
	}
	require.NotEmpty(t, bg.GetCurrentTrick(), "no trick was ever rendered")

	// **プレイヤー行は元から cuiPlayerName を使っている。**出力全体で探すと
	// 修正前でも当たるので、素の "CPU 1" が 1 つも残っていないことで見る。
	bold1 := color.Bold(i18n.Tf("cuiPlayerCpu", "idx", "1"))
	assert.Contains(t, out, bold1)
	assert.NotContains(t, strings.ReplaceAll(out, bold1, ""), "CPU 1")
	assert.NotContains(t, strings.ReplaceAll(out, color.Bold(i18n.Tf("cuiPlayerCpu", "idx", "2")), ""), "CPU 2")
}

// newBourreWithHuman はプレイヤー0を人間にしたブーレを返す。
func newBourreWithHuman() *domain.Bourre {
	players := make([]*domain.BourrePlayer, domain.BourrePlayerCnt)
	for i := range players {
		players[i] = domain.NewBourrePlayer(i == 0)
	}
	return domain.NewBourre(domain.NewTrumpCards(0), players, domain.DefaultBourreConfig())
}

// #5637: 参加して1トリックも取れないと「ブーレ」でポット全額を失う。Web は
// decideSummary でその額を出しているのに、CUI は「意思決定フェーズ」としか
// 言わず、初見のプレイヤーは何を賭けているのか分からないまま d を打っていた。
func TestBourreCuiPresenter_DecideShowsThePenaltyAtStake(t *testing.T) {
	bg := newBourreWithHuman()
	bg.Reset()
	require.Equal(t, domain.BourrePhaseDecide, bg.GetPhase())
	// 人間の手番になるまで CPU を進める (親の位置で順番が変わるため)。
	for i := 0; i < 10 && !bg.IsHumanTurn(); i++ {
		bg.CpuPlay()
	}
	require.True(t, bg.IsHumanTurn())
	require.Positive(t, bg.GetPot())

	out := new(presenter.BourreCuiPresenter).Output(bg, nil)

	assert.Contains(t, out, i18n.Tf("bourre.decidePenalty",
		"penalty", strconv.Itoa(bg.GetPot())))
}

// 罰金は手持ちチップが上限 (domain の min(potValue, chips))。ポットのほうが
// 大きいときにポット額を出すと、払えない額を警告することになる。
func TestBourreCuiPresenter_DecidePenaltyIsCappedByTheChipsHeld(t *testing.T) {
	bg := newBourreWithHuman()
	bg.Reset()
	for i := 0; i < 10 && !bg.IsHumanTurn(); i++ {
		bg.CpuPlay()
	}
	require.True(t, bg.IsHumanTurn())

	human := bg.GetPlayer(bg.GetCurrentPlayerIdx())
	human.SubtractChips(human.GetChips() - 1)
	require.Less(t, human.GetChips(), bg.GetPot())

	out := new(presenter.BourreCuiPresenter).Output(bg, nil)

	assert.Contains(t, out, i18n.Tf("bourre.decidePenalty", "penalty", "1"))
}

// 決着フェーズ以外や CPU の手番では出さない (Web も本人手番だけ出している)。
func TestBourreCuiPresenter_DecidePenaltyOnlyOnTheHumanDecideTurn(t *testing.T) {
	bg := newBourreWithHuman()
	bg.Reset()
	prefix, _, ok := strings.Cut(i18n.Tf("bourre.decidePenalty", "penalty", "\x00"), "\x00")
	require.True(t, ok)

	bg.SetPhase(domain.BourrePhasePlay)
	assert.NotContains(t, new(presenter.BourreCuiPresenter).Output(bg, nil), prefix,
		"プレイフェーズでは出さない")

	bg.SetPhase(domain.BourrePhaseDecide)
	cpu := newBourreAllCpu()
	cpu.Reset()
	require.False(t, cpu.IsHumanTurn())
	assert.NotContains(t, new(presenter.BourreCuiPresenter).Output(cpu, nil), prefix,
		"CPU の手番では出さない")
}

// 繰越ポットは配り直しの時点で pot に畳み込まれる (Bourre.nextHand) ので、
// 表示額にも自動的に乗る。**pot と carryPot を足すと二重に数える**ことになる
// ため、この経路を実際にポットが繰り越されたハンドで固定しておく。
func TestBourreCuiPresenter_DecidePenaltyIncludesTheCarriedPot(t *testing.T) {
	bg := newBourreWithHuman()
	// 全員が降りた配りだけがポットを繰り越す。CPU は手札次第で参加するので、
	// **繰り越しが起きるまで配り直す** (配り依存で skip すると、9/10 の実行で
	// 何も確かめないテストになる)。
	var firstPot, carried int
	for attempt := 0; attempt < 1000 && carried == 0; attempt++ {
		bg.Reset()
		firstPot = bg.GetPot()
		for i := 0; i < 100 && bg.GetPhase() == domain.BourrePhaseDecide; i++ {
			if bg.IsHumanTurn() {
				require.NoError(t, bg.PlayerDecide(false))
				continue
			}
			bg.CpuPlay()
		}
		carried = bg.GetCarryPot()
	}
	require.Positive(t, carried, "1000 回配っても全員フォールドの手が出なかった")

	bg.NextHand()
	require.Equal(t, domain.BourrePhaseDecide, bg.GetPhase())
	require.Zero(t, bg.GetCarryPot(), "前提: 繰越は pot に畳み込まれて 0 になる")
	require.Greater(t, bg.GetPot(), firstPot)
	for i := 0; i < 10 && !bg.IsHumanTurn(); i++ {
		bg.CpuPlay()
	}
	require.True(t, bg.IsHumanTurn())

	out := new(presenter.BourreCuiPresenter).Output(bg, nil)

	assert.Contains(t, out, i18n.Tf("bourre.decidePenalty",
		"penalty", strconv.Itoa(min(bg.GetPot(), bg.GetPlayer(bg.GetCurrentPlayerIdx()).GetChips()))))
	assert.Greater(t, bg.GetPot(), carried, "繰越ぶんが表示額に乗っている")
}
