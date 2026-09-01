//go:build test

package presenter_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupKilleCuiMock(phase domain.KillePhase, players []*domain.KillePlayer) *interfaces.MockKilleGame {
	m := new(interfaces.MockKilleGame)
	m.On("GetPhase").Return(phase)
	m.On("GetRoundNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(3)
	m.On("GetStockCount").Return(38)
	m.On("GetPot").Return(4)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLoserIdxs").Return([]int{1})
	m.On("GetPlayers").Return(players)
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
	}
	m.On("GetEvents").Return(([]*domain.KilleEvent)(nil)).Maybe()
	return m
}

// killeSeatLines は出力から席の行だけを取り出す (凡例やヘッダを除く)。
func killeSeatLines(out string) string {
	var seats []string
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, i18n.T("cuiPlayerYou")) ||
			strings.Contains(line, strings.SplitN(i18n.T("cuiPlayerCpu"), "{{", 2)[0]) {
			seats = append(seats, line)
		}
	}
	return strings.Join(seats, "\n")
}

func TestKilleCuiPresenter_HidesOtherHands(t *testing.T) {
	players := makeKillePlayers(domain.KilleNum5, domain.KillePig, domain.KilleCuckoo, domain.KilleNum9)
	m := setupKilleCuiMock(domain.KillePhaseExchange, players)
	out := new(presenter.KilleCuiPresenter).Output(m, nil)

	assert.Contains(t, out, "5", "the human's own card is shown")
	// **席の行だけを見る。**強さ序列の凡例 (#6522) は全種の名前を含むので、
	// 出力全体から名前を探すと「伏せているか」ではなく凡例を見てしまう。
	seats := killeSeatLines(out)
	assert.NotContains(t, seats, "Pig", "seat 1 must stay face down")
	assert.NotContains(t, seats, "Cuckoo", "seat 2 must stay face down")
	assert.Contains(t, out, "非公開")
	assert.Contains(t, out, "[親]")
	assert.Contains(t, out, "ポット: 4")
}

func TestKilleCuiPresenter_ShowdownNamesTheReason(t *testing.T) {
	players := makeKillePlayers(domain.KilleNum5, domain.KillePig, domain.KilleCuckoo, domain.KilleNum9)
	players[1].SetOut(domain.KilleKnockPig)
	players[2].SetOut(domain.KilleKnockHussar)
	players[3].SetOut(domain.KilleKnockLowest)
	m := setupKilleCuiMock(domain.KillePhaseShowdown, players)
	out := new(presenter.KilleCuiPresenter).Output(m, nil)

	// 公開されるので全員の札が出る。
	assert.Contains(t, out, "Pig")
	assert.Contains(t, out, "Cuckoo")
	// **なぜ落ちたかまで出す。**軽騎兵と豚は自分の手の強さと無関係に落ちる。
	assert.Contains(t, out, "豚に噛まれた")
	assert.Contains(t, out, "軽騎兵に返り討ち")
	assert.Contains(t, out, "[脱落]")
	assert.Contains(t, out, "1人が脱落")
}

func TestKilleCuiPresenter_DealerGetsItsOwnPrompt(t *testing.T) {
	players := makeKillePlayers(domain.KilleNum5, domain.KilleNum9, domain.KilleNum7, domain.KilleNum3)
	m := new(interfaces.MockKilleGame)
	m.On("GetPhase").Return(domain.KillePhaseExchange)
	m.On("GetRoundNumber").Return(0)
	m.On("GetCurrentPlayerIdx").Return(3)
	m.On("GetDealerIdx").Return(3)
	m.On("GetStockCount").Return(38)
	m.On("GetPot").Return(4)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetLoserIdxs").Return([]int{})
	m.On("GetPlayers").Return(players)
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
	}
	m.On("GetEvents").Return(([]*domain.KilleEvent)(nil))

	out := new(presenter.KilleCuiPresenter).Output(m, nil)
	assert.Contains(t, out, "山札から引いて交換")
}

func TestKilleCuiPresenter_SatisfiedAndEliminatedTags(t *testing.T) {
	players := makeKillePlayers(domain.KilleNum5, domain.KilleNum9, domain.KilleNum7, domain.KilleNum3)
	players[1].SetSatisfied(true)
	players[2].SetIsFinished(true)
	m := setupKilleCuiMock(domain.KillePhaseExchange, players)
	out := new(presenter.KilleCuiPresenter).Output(m, nil)

	assert.Contains(t, out, "[満足]")
	assert.Contains(t, out, "退場")
}

func TestKilleCuiPresenter_ErrorAndGameEnd(t *testing.T) {
	players := makeKillePlayers(domain.KilleNum5, domain.KilleNum9, domain.KilleNum7, domain.KilleNum3)

	m := setupKilleCuiMock(domain.KillePhaseExchange, players)
	assert.Contains(t, new(presenter.KilleCuiPresenter).Output(m, errors.New("boom")), "boom")

	end := new(interfaces.MockKilleGame)
	end.On("GetPhase").Return(domain.KillePhaseGameEnd)
	end.On("GetRoundNumber").Return(9)
	end.On("GetCurrentPlayerIdx").Return(0)
	end.On("GetDealerIdx").Return(3)
	end.On("GetStockCount").Return(38)
	end.On("GetPot").Return(0)
	end.On("GetGameEndFlag").Return(true)
	end.On("GetWinnerIdx").Return(0)
	end.On("GetLoserIdxs").Return([]int{1, 2, 3})
	end.On("GetEvents").Return(([]*domain.KilleEvent)(nil))
	end.On("GetPlayers").Return(players)
	for i := range players {
		end.On("GetPlayer", i).Return(players[i])
	}
	assert.Contains(t, new(presenter.KilleCuiPresenter).Output(end, nil), "ゲーム終了")
}

func TestKilleCuiPresenter_ActionLogOutput(t *testing.T) {
	players := makeKillePlayers(domain.KilleNum5)
	m := setupKilleCuiMock(domain.KillePhaseExchange, players)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotNil(t, new(presenter.KilleCuiPresenter).ActionLogOutput(m))
}

// #5725: Web は kille-events で交換の実況を、reentriesUsed で買い戻し回数を出して
// いるのに、CUI は GetEvents() も GetReentries() も呼んでおらず、誰が誰と交換し、
// 誰が何回買い戻したのかを画面から読み取れなかった。
func TestKilleCuiPresenter_ShowsExchangesAndReentries(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true) // 名前は Bold で装飾されるので、素の文字列で照合する
	defer color.SetNoColor(orig)
	players := makeKillePlayers(domain.KilleNum5, domain.KillePig, domain.KilleCuckoo, domain.KilleNum9)
	// 席 1 が 2 回買い戻している。
	players[1].AddReentry()
	players[1].AddReentry()
	m := setupKilleCuiMock(domain.KillePhaseExchange, players)
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetEvents")
	m.On("GetEvents").Return([]*domain.KilleEvent{
		{Kind: "swap", Actor: 0, Target: 1},
		{Kind: "stock", Actor: 3, Target: -1},
	})

	out := new(presenter.KilleCuiPresenter).Output(m, nil)

	assert.Contains(t, out, i18n.T("kille.eventsTitle"))
	assert.Contains(t, out, i18n.Tf("kille.event.swap",
		"actor", i18n.T("cuiPlayerYou"), "target", i18n.Tf("cuiPlayerCpu", "idx", "1")))
	assert.Contains(t, out, i18n.Tf("kille.event.stock",
		"actor", i18n.Tf("cuiPlayerCpu", "idx", "3")))
	// 買い戻し回数は上限つきで出す (残り回数が判断材料になる)。
	assert.Contains(t, out, i18n.Tf("kille.reentriesUsed",
		"used", "2", "max", strconv.Itoa(domain.KilleMaxReentries)))
}

// 交換が起きていないラウンドでは一覧ごと出さない。
func TestKilleCuiPresenter_NoEventBlockWhenNothingHappened(t *testing.T) {
	players := makeKillePlayers(domain.KilleNum5, domain.KillePig, domain.KilleCuckoo, domain.KilleNum9)
	m := setupKilleCuiMock(domain.KillePhaseExchange, players)

	out := new(presenter.KilleCuiPresenter).Output(m, nil)

	assert.NotContains(t, out, i18n.T("kille.eventsTitle"))
}

// **42 枚は独自の序列を持つ。**Harlequin が最強で Mask が最弱という並びを
// 覚えていないと交換も満足も判断できないのに、CUI には手掛かりが無かった (#6522)。
func TestKilleCuiPresenter_ShowsTheStrengthLadder(t *testing.T) {
	g := domain.NewDefaultKille()
	g.Reset()
	out := new(presenter.KilleCuiPresenter).Output(g, nil)

	// 並びは Web の KILLE_LADDER と同じ。数札は 1 つの帯にまとめる。
	want := strings.Join([]string{
		"Harlequin", "Cuckoo", "Hussar", "Pig", "Cavalier", "Inn",
		"12 … 1", "Wreath", "Flowerpot", "Mask",
	}, " > ")
	assert.Contains(t, out, i18n.Tf("kille.ladder", "ladder", want))
	assert.NotContains(t, out, "{{")
}

// **並びはドメインの種の値そのもの。**表を手で並べ替えても、値の降順から
// 外れたらここで落ちる ── 覚え書きが実際の強さと食い違うのが最悪の形。
func TestKilleLadder_IsStrictlyDescendingByRank(t *testing.T) {
	ordered := []domain.KilleRank{
		domain.KilleHarlequin, domain.KilleCuckoo, domain.KilleHussar,
		domain.KillePig, domain.KilleCavalier, domain.KilleInn,
		domain.KilleNum12, domain.KilleNum1,
		domain.KilleWreath, domain.KilleFlowerpot, domain.KilleMask,
	}
	for i := 1; i < len(ordered); i++ {
		assert.Greater(t, int(ordered[i-1]), int(ordered[i]),
			"%s は %s より強くなければならない",
			domain.KilleRankName(ordered[i-1]), domain.KilleRankName(ordered[i]))
	}
}
