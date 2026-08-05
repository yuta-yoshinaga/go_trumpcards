//go:build test

package presenter_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupBidEuchreCuiMock(o bidEuchreMockOpts) *interfaces.MockBidEuchreGame {
	m := new(interfaces.MockBidEuchreGame)
	players := makeBidEuchrePlayers(
		[]*domain.Card{beTestCard(domain.CardDesignSpade, 1)},
		[]*domain.Card{beTestCard(domain.CardDesignHeart, 11)},
		[]*domain.Card{beTestCard(domain.CardDesignClover, 11)},
		[]*domain.Card{beTestCard(domain.CardDesignDiamond, 9)},
	)
	m.On("GetPhase").Return(o.phase)
	m.On("GetHandNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(3)
	m.On("GetDeclarerIdx").Return(o.declarer)
	m.On("GetTrump").Return(o.trump)
	m.On("GetTrumpSuit").Return(domain.BidEuchreTrumpSuit(o.trump))
	m.On("IsTrumpChosen").Return(o.trumpChosen)
	m.On("GetTrick").Return([]*domain.Card{beTestCard(domain.CardDesignSpade, 13)})
	m.On("GetTrickNumber").Return(3)
	m.On("GetGameEndFlag").Return(o.gameEnd)
	m.On("GetWinnerTeam").Return(o.winner)
	m.On("GetHighBid").Return(o.highBid)
	m.On("GetLastResult").Return(o.result)
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(true)
	m.On("BidEuchreValidPlays", 0).Return([]int{0})
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
		m.On("GetTricksWon", i).Return(0)
	}
	for team := range domain.BidEuchreTeamCnt {
		m.On("GetScore", team).Return(0)
	}
	return m
}

// **キティが無く、誰の手札も公開されない。**
func TestBidEuchreCuiPresenter_HidesEveryOtherHand(t *testing.T) {
	out := new(presenter.BidEuchreCuiPresenter).Output(setupBidEuchreCuiMock(defaultBidEuchreOpts()), nil)
	assert.Contains(t, out, "[0]")
	assert.Contains(t, out, "非公開")
	assert.Contains(t, out, "[親]")
	assert.Contains(t, out, "[宣言]")
	if got := strings.Count(out, "非公開"); got != 3 {
		t.Errorf("%d hidden hands, want 3", got)
	}
}

// **親だけは同額で奪える。**入札画面で必ず読めること。
func TestBidEuchreCuiPresenter_SaysTheDealerMayEqualTheBid(t *testing.T) {
	o := defaultBidEuchreOpts()
	o.phase = domain.BidEuchrePhaseBid
	out := new(presenter.BidEuchreCuiPresenter).Output(setupBidEuchreCuiMock(o), nil)
	assert.Contains(t, out, "親だけは同額")
	// **最低 3。**
	assert.Contains(t, out, "最低3")
	assert.Contains(t, out, "b <3-6>")
}

// **ノートランプが 2 種類あり、ローは序列が逆転する。**
func TestBidEuchreCuiPresenter_ShowsBothNoTrumpForms(t *testing.T) {
	o := defaultBidEuchreOpts()
	o.phase = domain.BidEuchrePhaseChooseTrump
	out := new(presenter.BidEuchreCuiPresenter).Output(setupBidEuchreCuiMock(o), nil)

	assert.Contains(t, out, "切札の候補")
	assert.Contains(t, out, "0:♠")
	assert.Contains(t, out, "4:NTハイ")
	assert.Contains(t, out, "5:NTロー")
	// ハイとローが別項目として並ぶ。
	high := strings.Index(out, "4:NTハイ")
	low := strings.Index(out, "5:NTロー")
	assert.NotEqual(t, -1, high)
	assert.Less(t, high, low)
	assert.Contains(t, out, "NTローは序列が逆転し9が最強")
}

// **左ボワーは切札スートの札。**プレイ画面で読めること。
func TestBidEuchreCuiPresenter_SaysTheLeftBowerIsATrump(t *testing.T) {
	out := new(presenter.BidEuchreCuiPresenter).Output(setupBidEuchreCuiMock(defaultBidEuchreOpts()), nil)
	assert.Contains(t, out, "左ボワーは切札スートの札")
	assert.Contains(t, out, "出せる札: 0")
}

func TestBidEuchreCuiPresenter_ShowsTheContractOnlyOnceBid(t *testing.T) {
	withBid := new(presenter.BidEuchreCuiPresenter).Output(setupBidEuchreCuiMock(defaultBidEuchreOpts()), nil)
	assert.Contains(t, withBid, "契約:")
	assert.Contains(t, withBid, "4トリック")

	o := defaultBidEuchreOpts()
	o.highBid = nil
	assert.NotContains(t, new(presenter.BidEuchreCuiPresenter).Output(setupBidEuchreCuiMock(o), nil), "契約:")

	// **落札直後は切札が未定。**
	unchosen := defaultBidEuchreOpts()
	unchosen.trumpChosen = false
	assert.Contains(t, new(presenter.BidEuchreCuiPresenter).Output(setupBidEuchreCuiMock(unchosen), nil), "未定")
}

// **未達で失うのは宣言額、守備側は取ったトリックを得点する。**
func TestBidEuchreCuiPresenter_ExplainsTheSettlement(t *testing.T) {
	o := defaultBidEuchreOpts()
	o.phase = domain.BidEuchrePhaseHandEnd
	o.result = &domain.BidEuchreHandResult{
		Points: [domain.BidEuchreTeamCnt]int{-5, 4},
		Tricks: [domain.BidEuchreTeamCnt]int{2, 4},
		Made:   false,
		Bid:    5,
	}
	out := new(presenter.BidEuchreCuiPresenter).Output(setupBidEuchreCuiMock(o), nil)
	assert.Contains(t, out, "宣言失敗")
	assert.Contains(t, out, "取った数ではなく宣言額")
	assert.Contains(t, out, "-5")
	assert.Contains(t, out, "守備側は未達でも取ったトリックを得点")

	made := defaultBidEuchreOpts()
	made.phase = domain.BidEuchrePhaseHandEnd
	made.result = &domain.BidEuchreHandResult{
		Points: [domain.BidEuchreTeamCnt]int{4, 2},
		Tricks: [domain.BidEuchreTeamCnt]int{4, 2},
		Made:   true,
		Bid:    3,
	}
	assert.Contains(t, new(presenter.BidEuchreCuiPresenter).Output(setupBidEuchreCuiMock(made), nil), "宣言達成")
}

// **落札前でも精算行が落ちない。**declarerIdx は -1 のまま。
func TestBidEuchreCuiPresenter_SettlementSurvivesNoDeclarer(t *testing.T) {
	o := defaultBidEuchreOpts()
	o.phase = domain.BidEuchrePhaseHandEnd
	o.declarer = -1
	o.result = &domain.BidEuchreHandResult{
		Points: [domain.BidEuchreTeamCnt]int{4, 2},
		Tricks: [domain.BidEuchreTeamCnt]int{4, 2},
		Made:   true,
		Bid:    3,
	}
	assert.NotPanics(t, func() {
		new(presenter.BidEuchreCuiPresenter).Output(setupBidEuchreCuiMock(o), nil)
	})
}

func TestBidEuchreCuiPresenter_PhasePrompts(t *testing.T) {
	for _, tc := range []struct {
		phase domain.BidEuchrePhase
		want  string
	}{
		{domain.BidEuchrePhaseBid, "親だけは同額"},
		{domain.BidEuchrePhaseChooseTrump, "切札の候補"},
		{domain.BidEuchrePhasePlay, "追随は強制"},
		{domain.BidEuchrePhaseHandEnd, "次の局へ"},
	} {
		o := defaultBidEuchreOpts()
		o.phase = tc.phase
		assert.Contains(t, new(presenter.BidEuchreCuiPresenter).Output(setupBidEuchreCuiMock(o), nil), tc.want)
	}
}

// 勝利点 32 とチーム得点が読めること。
func TestBidEuchreCuiPresenter_ShowsTheScoreSheet(t *testing.T) {
	out := new(presenter.BidEuchreCuiPresenter).Output(setupBidEuchreCuiMock(defaultBidEuchreOpts()), nil)
	assert.Contains(t, out, "勝利点: 32点")
	assert.Contains(t, out, "チーム0:")
	assert.Contains(t, out, "チーム1:")
}

func TestBidEuchreCuiPresenter_ErrorAndGameEnd(t *testing.T) {
	out := new(presenter.BidEuchreCuiPresenter).Output(setupBidEuchreCuiMock(defaultBidEuchreOpts()), errors.New("boom"))
	assert.Contains(t, out, "boom")

	o := defaultBidEuchreOpts()
	o.phase = domain.BidEuchrePhaseGameEnd
	o.gameEnd = true
	o.winner = 0
	assert.Contains(t, new(presenter.BidEuchreCuiPresenter).Output(setupBidEuchreCuiMock(o), nil), "ゲーム終了")
}

func TestBidEuchreCuiPresenter_ActionLogOutput(t *testing.T) {
	m := setupBidEuchreCuiMock(defaultBidEuchreOpts())
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotNil(t, new(presenter.BidEuchreCuiPresenter).ActionLogOutput(m))
}

// **「立っている宣言＋1、親なら同額」を毎回暗算させない** (#4899)。
func TestBidEuchreCuiPresenter_ShowsTheLegalBidRange(t *testing.T) {
	p := new(presenter.BidEuchreCuiPresenter)

	build := func(high *domain.BidEuchreBid, bidder int) string {
		o := defaultBidEuchreOpts()
		o.phase = domain.BidEuchrePhaseBid
		o.highBid = high
		m := setupBidEuchreCuiMock(o)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetBidPlayerIdx")
		m.On("GetBidPlayerIdx").Return(bidder)
		return p.Output(m, nil)
	}

	// 宣言が無ければ最低値から (受け入れ条件1)。
	assert.Contains(t, build(nil, 1),
		"有効な入札: "+strconv.Itoa(domain.BidEuchreMinBid)+"〜"+strconv.Itoa(domain.BidEuchreMaxBid))

	// 非ディーラーは +1 から (受け入れ条件2)。
	assert.Contains(t, build(&domain.BidEuchreBid{Player: 0, Value: 4}, 1), "有効な入札: 5〜")

	// **親だけは同額で奪える** (受け入れ条件3)。ディーラーは席 3。
	assert.Contains(t, build(&domain.BidEuchreBid{Player: 0, Value: 4}, 3), "有効な入札: 4〜")

	// 上限まで宣言されたら、非ディーラーには選べる値が無い。
	top := build(&domain.BidEuchreBid{Player: 0, Value: domain.BidEuchreMaxBid}, 1)
	assert.Contains(t, top, "これ以上の入札はできません")
	assert.NotContains(t, top, "有効な入札:")

	// 同じ状況でも親は同額で奪えるので、範囲が出る。
	assert.Contains(t, build(&domain.BidEuchreBid{Player: 0, Value: domain.BidEuchreMaxBid}, 3), "有効な入札:")
}
