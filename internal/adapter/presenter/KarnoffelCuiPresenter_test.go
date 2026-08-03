//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupKarnoffelCuiMock(o karnoffelMockOpts) *interfaces.MockKarnoffelGame {
	m := new(interfaces.MockKarnoffelGame)
	players := makeKarnoffelPlayers(
		[]*domain.Card{knTestCard(domain.CardDesignSpade, 13)},
		[]*domain.Card{knTestCard(domain.CardDesignHeart, 11)},
		[]*domain.Card{knTestCard(domain.CardDesignClover, 6)},
		[]*domain.Card{knTestCard(domain.CardDesignDiamond, 2)},
	)
	m.On("GetPhase").Return(o.phase)
	m.On("GetHandNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(3)
	m.On("GetChosenSuit").Return(o.chosen)
	m.On("GetTrick").Return([]*domain.Card{knTestCard(domain.CardDesignSpade, 9)})
	m.On("GetTrickNumber").Return(2)
	m.On("GetGameEndFlag").Return(o.gameEnd)
	m.On("GetWinnerTeam").Return(o.winner)
	m.On("GetConfig").Return(domain.DefaultKarnoffelConfig())
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(true)
	m.On("KarnoffelValidPlays", 0).Return([]int{0})
	m.On("GetLastResult").Return(o.result)
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
		m.On("GetTricksWon", i).Return(0)
		m.On("GetUpCard", i).Return(knTestCard(domain.CardDesignHeart, 3+i))
	}
	for team := range domain.KarnoffelTeamCnt {
		m.On("KarnoffelTeamTricks", team).Return(0)
		m.On("GetHandsWon", team).Return(0)
	}
	return m
}

// **切札は表向きの 4 枚のうち最も低い札が決める。**その説明を常時出す。
func TestKarnoffelCuiPresenter_ExplainsHowTheSuitWasChosen(t *testing.T) {
	out := new(presenter.KarnoffelCuiPresenter).Output(setupKarnoffelCuiMock(defaultKarnoffelOpts()), nil)
	assert.Contains(t, out, "選ばれたスート")
	assert.Contains(t, out, "表向きに配られた4枚のうち最も低い札")
	// 各席の表向きの札が読める。
	assert.Contains(t, out, "表 ")
}

// 手札は自分のぶんだけ。
func TestKarnoffelCuiPresenter_HidesTheOtherHands(t *testing.T) {
	out := new(presenter.KarnoffelCuiPresenter).Output(setupKarnoffelCuiMock(defaultKarnoffelOpts()), nil)
	assert.Contains(t, out, "[0]")
	assert.Contains(t, out, "[親]")
	if got := strings.Count(out, "非公開"); got != 3 {
		t.Errorf("%d hidden hands, want 3", got)
	}

	o := defaultKarnoffelOpts()
	o.phase = domain.KarnoffelPhaseHandEnd
	o.result = &domain.KarnoffelHandResult{WinnerTeam: 0, Tricks: [domain.KarnoffelTeamCnt]int{3, 1}}
	revealed := new(presenter.KarnoffelCuiPresenter).Output(setupKarnoffelCuiMock(o), nil)
	assert.NotContains(t, revealed, "非公開")
}

// **序列は表で見せる。**悪魔の特殊性は文章にしないと伝わらない。
func TestKarnoffelCuiPresenter_ShowsTheRankingLadder(t *testing.T) {
	out := new(presenter.KarnoffelCuiPresenter).Output(setupKarnoffelCuiMock(defaultKarnoffelOpts()), nil)
	assert.Contains(t, out, "J(カルニッフェル)")
	assert.Contains(t, out, "7(悪魔・リード時のみ)")
	assert.Contains(t, out, "6(法王)")
	assert.Contains(t, out, "2(皇帝)")
	// **部分切札の負け方も出す。**
	assert.Contains(t, out, "3はKに、4はK/Qに、5は絵札すべてに負けます")
	assert.Contains(t, out, "追随して出した7はあらゆる札に負けます")
	// カルニッフェルが悪魔より前に来る。
	assert.Less(t, strings.Index(out, "J(カルニッフェル)"), strings.Index(out, "7(悪魔"))
}

// **追随の義務は無い。**プレイ画面で読めること。
func TestKarnoffelCuiPresenter_SaysThereIsNoFollowSuitRule(t *testing.T) {
	out := new(presenter.KarnoffelCuiPresenter).Output(setupKarnoffelCuiMock(defaultKarnoffelOpts()), nil)
	assert.Contains(t, out, "追随の義務はありません")
	assert.Contains(t, out, "第1トリックのリードに悪魔は使えません")
	assert.Contains(t, out, "出せる札: 0")
}

func TestKarnoffelCuiPresenter_ShowsTheScoreSheet(t *testing.T) {
	out := new(presenter.KarnoffelCuiPresenter).Output(setupKarnoffelCuiMock(defaultKarnoffelOpts()), nil)
	assert.Contains(t, out, "3局先取")
	assert.Contains(t, out, "1局は3トリック先取")
	assert.Contains(t, out, "チーム0:")
	assert.Contains(t, out, "チーム1:")
}

// **3 トリックに届かなければ勝者なし。**その区別が読めること。
func TestKarnoffelCuiPresenter_ReportsTheHandResult(t *testing.T) {
	o := defaultKarnoffelOpts()
	o.phase = domain.KarnoffelPhaseHandEnd
	o.result = &domain.KarnoffelHandResult{WinnerTeam: 1, Tricks: [domain.KarnoffelTeamCnt]int{2, 3}}
	won := new(presenter.KarnoffelCuiPresenter).Output(setupKarnoffelCuiMock(o), nil)
	assert.Contains(t, won, "チーム1が局を取りました")
	assert.Contains(t, won, "2-3")

	drawn := defaultKarnoffelOpts()
	drawn.phase = domain.KarnoffelPhaseHandEnd
	drawn.result = &domain.KarnoffelHandResult{WinnerTeam: -1, Tricks: [domain.KarnoffelTeamCnt]int{2, 2}}
	out := new(presenter.KarnoffelCuiPresenter).Output(setupKarnoffelCuiMock(drawn), nil)
	assert.Contains(t, out, "どちらも3トリックに届きませんでした")
	assert.Contains(t, out, "次の局へ")
}

func TestKarnoffelCuiPresenter_ErrorAndGameEnd(t *testing.T) {
	out := new(presenter.KarnoffelCuiPresenter).Output(setupKarnoffelCuiMock(defaultKarnoffelOpts()), errors.New("boom"))
	assert.Contains(t, out, "boom")

	o := defaultKarnoffelOpts()
	o.phase = domain.KarnoffelPhaseGameEnd
	o.gameEnd = true
	o.winner = 0
	assert.Contains(t, new(presenter.KarnoffelCuiPresenter).Output(setupKarnoffelCuiMock(o), nil), "ゲーム終了")
}

// 切札が未決定でも落ちない。
func TestKarnoffelCuiPresenter_SurvivesNoChosenSuit(t *testing.T) {
	o := defaultKarnoffelOpts()
	o.chosen = 0
	assert.NotPanics(t, func() {
		new(presenter.KarnoffelCuiPresenter).Output(setupKarnoffelCuiMock(o), nil)
	})
}

func TestKarnoffelCuiPresenter_ActionLogOutput(t *testing.T) {
	m := setupKarnoffelCuiMock(defaultKarnoffelOpts())
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotNil(t, new(presenter.KarnoffelCuiPresenter).ActionLogOutput(m))
}
