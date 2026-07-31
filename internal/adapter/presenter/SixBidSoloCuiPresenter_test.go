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

func setupSixBidSoloCuiMock(o sixBidSoloMockOpts) *interfaces.MockSixBidSoloGame {
	m := new(interfaces.MockSixBidSoloGame)
	players := makeSixBidSoloPlayers(
		[]*domain.Card{sbsTestCard(domain.CardDesignSpade, 1)},
		[]*domain.Card{sbsTestCard(domain.CardDesignHeart, 10)},
		[]*domain.Card{sbsTestCard(domain.CardDesignClover, 13)},
	)
	m.On("GetPhase").Return(o.phase)
	m.On("GetHandNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(2)
	m.On("GetDeclarerIdx").Return(o.declarer)
	m.On("GetTrumpSuit").Return(o.trumpSuit)
	m.On("IsDeclared").Return(o.declared)
	m.On("IsSpreadOpen").Return(o.spreadOpen)
	m.On("GetCalledCard").Return(o.called)
	m.On("GetWidow").Return([]*domain.Card{
		sbsTestCard(domain.CardDesignDiamond, 1),
		sbsTestCard(domain.CardDesignDiamond, 10),
		sbsTestCard(domain.CardDesignDiamond, 6),
	})
	m.On("SixBidSoloWidowPoints").Return(21)
	m.On("GetTrick").Return([]*domain.Card{sbsTestCard(domain.CardDesignSpade, 13)})
	m.On("GetTrickNumber").Return(3)
	m.On("GetGameEndFlag").Return(o.gameEnd)
	m.On("GetWinnerIdx").Return(o.winner)
	m.On("GetConfig").Return(domain.DefaultSixBidSoloConfig())
	m.On("GetHighBid").Return(o.highBid)
	m.On("GetLastResult").Return(o.result)
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(true)
	m.On("SixBidSoloValidPlays", 0).Return([]int{0})
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
		m.On("GetPoints", i).Return(0)
		m.On("GetTricksWon", i).Return(0)
		m.On("GetScore", i).Return(0)
	}
	return m
}

// **ウィドウは精算まで伏せたまま。**枚数だけ出す。
func TestSixBidSoloCuiPresenter_KeepsTheWidowFaceDown(t *testing.T) {
	out := new(presenter.SixBidSoloCuiPresenter).Output(setupSixBidSoloCuiMock(defaultSixBidSoloOpts()), nil)
	assert.Contains(t, out, "ウィドウ")
	assert.Contains(t, out, "伏せ 3枚")

	o := defaultSixBidSoloOpts()
	o.phase = domain.SixBidSoloPhaseHandEnd
	o.result = &domain.SixBidSoloHandResult{Kind: domain.SixBidSoloBidSolo, Made: true, DeclarerPoints: 65, Target: 61, WidowPoints: 21}
	revealed := new(presenter.SixBidSoloCuiPresenter).Output(setupSixBidSoloCuiMock(o), nil)
	assert.Contains(t, revealed, "21pt", "the widow's value is shown once it is revealed")
	assert.NotContains(t, revealed, "伏せ 3枚")
}

// 他家の手札はプレイ中は見えない。
func TestSixBidSoloCuiPresenter_HidesTheOtherHands(t *testing.T) {
	out := new(presenter.SixBidSoloCuiPresenter).Output(setupSixBidSoloCuiMock(defaultSixBidSoloOpts()), nil)
	assert.Contains(t, out, "[0]")
	assert.Contains(t, out, "[親]")
	assert.Contains(t, out, "[宣言]")
	if got := strings.Count(out, "非公開"); got != 2 {
		t.Errorf("%d hidden hands, want 2", got)
	}
}

// **スプレッド・ミゼールでは宣言者の手札が公開される。**
func TestSixBidSoloCuiPresenter_ExposesTheSpreadMisereHand(t *testing.T) {
	o := defaultSixBidSoloOpts()
	o.declarer = 1
	o.spreadOpen = true
	out := new(presenter.SixBidSoloCuiPresenter).Output(setupSixBidSoloCuiMock(o), nil)
	if got := strings.Count(out, "非公開"); got != 1 {
		t.Errorf("%d hidden hands, want 1 — the declarer's is laid down", got)
	}
}

// **通常ビッドは 61 点以上。**入札画面で序列と一緒に読めること。
func TestSixBidSoloCuiPresenter_ShowsTheLadderWhileBidding(t *testing.T) {
	o := defaultSixBidSoloOpts()
	o.phase = domain.SixBidSoloPhaseBid
	o.highBid = nil
	out := new(presenter.SixBidSoloCuiPresenter).Output(setupSixBidSoloCuiMock(o), nil)

	assert.Contains(t, out, "ビッドの序列")
	for _, want := range []string{"1:ソロ", "2:ハート・ソロ", "3:ミゼール", "4:ギャランティー・ソロ", "5:スプレッド・ミゼール", "6:コール・ソロ"} {
		assert.Contains(t, out, want)
	}
	// 並びは低い順。
	assert.Less(t, strings.Index(out, "1:ソロ"), strings.Index(out, "6:コール・ソロ"))
	assert.Contains(t, out, "61点以上")
}

// **コール・ソロの指名札を見せる。**交換が起きたことが読めないと意味が通らない。
func TestSixBidSoloCuiPresenter_ShowsTheCalledCard(t *testing.T) {
	o := defaultSixBidSoloOpts()
	o.called = sbsTestCard(domain.CardDesignHeart, 1)
	out := new(presenter.SixBidSoloCuiPresenter).Output(setupSixBidSoloCuiMock(o), nil)
	assert.Contains(t, out, "指名札")
	assert.Contains(t, out, "交換に応じます")

	// 指名が無ければ出さない。
	assert.NotContains(t, new(presenter.SixBidSoloCuiPresenter).Output(setupSixBidSoloCuiMock(defaultSixBidSoloOpts()), nil), "指名札")
}

func TestSixBidSoloCuiPresenter_ShowsTheContractOnlyOnceBid(t *testing.T) {
	withBid := new(presenter.SixBidSoloCuiPresenter).Output(setupSixBidSoloCuiMock(defaultSixBidSoloOpts()), nil)
	assert.Contains(t, withBid, "契約:")
	// **♠ のギャランティーは 80 点。**
	assert.Contains(t, withBid, "80点")

	o := defaultSixBidSoloOpts()
	o.highBid = nil
	assert.NotContains(t, new(presenter.SixBidSoloCuiPresenter).Output(setupSixBidSoloCuiMock(o), nil), "契約:")

	// 落札直後は切札が未定。
	unchosen := defaultSixBidSoloOpts()
	unchosen.declared = false
	assert.Contains(t, new(presenter.SixBidSoloCuiPresenter).Output(setupSixBidSoloCuiMock(unchosen), nil), "未定")

	// ミゼール系は切札なし。
	misere := defaultSixBidSoloOpts()
	misere.highBid = &domain.SixBidSoloBid{Player: 1, Kind: domain.SixBidSoloBidMisere}
	misere.trumpSuit = 0
	assert.Contains(t, new(presenter.SixBidSoloCuiPresenter).Output(setupSixBidSoloCuiMock(misere), nil), "なし")
}

// **ウィドウ加算はミゼール系では 0。**精算でその区別が読めること。
func TestSixBidSoloCuiPresenter_ExplainsTheSettlement(t *testing.T) {
	o := defaultSixBidSoloOpts()
	o.phase = domain.SixBidSoloPhaseHandEnd
	o.result = &domain.SixBidSoloHandResult{
		Kind:           domain.SixBidSoloBidSolo,
		DeclarerPoints: 40,
		WidowPoints:    0,
		Target:         61,
		Made:           false,
		Value:          40,
	}
	out := new(presenter.SixBidSoloCuiPresenter).Output(setupSixBidSoloCuiMock(o), nil)
	assert.Contains(t, out, "失敗")
	assert.Contains(t, out, "40点")
	assert.Contains(t, out, "目標 61点")
	assert.Contains(t, out, "ウィドウ加算: 0点")
	assert.Contains(t, out, "ミゼール系では加算されません")
	assert.Contains(t, out, "受け払い")

	made := defaultSixBidSoloOpts()
	made.phase = domain.SixBidSoloPhaseHandEnd
	made.result = &domain.SixBidSoloHandResult{Kind: domain.SixBidSoloBidSolo, DeclarerPoints: 70, Target: 61, Made: true, Value: 20}
	assert.Contains(t, new(presenter.SixBidSoloCuiPresenter).Output(setupSixBidSoloCuiMock(made), nil), "達成")
}

func TestSixBidSoloCuiPresenter_PhasePrompts(t *testing.T) {
	for _, tc := range []struct {
		phase domain.SixBidSoloPhase
		want  string
	}{
		{domain.SixBidSoloPhaseBid, "ビッドの序列"},
		{domain.SixBidSoloPhaseDeclare, "コール・ソロは d"},
		{domain.SixBidSoloPhasePlay, "追随は強制"},
		{domain.SixBidSoloPhaseHandEnd, "次の局へ"},
	} {
		o := defaultSixBidSoloOpts()
		o.phase = tc.phase
		assert.Contains(t, new(presenter.SixBidSoloCuiPresenter).Output(setupSixBidSoloCuiMock(o), nil), tc.want)
	}
}

func TestSixBidSoloCuiPresenter_ListsThePlayableIndexes(t *testing.T) {
	out := new(presenter.SixBidSoloCuiPresenter).Output(setupSixBidSoloCuiMock(defaultSixBidSoloOpts()), nil)
	assert.Contains(t, out, "出せる札: 0")
	assert.Contains(t, out, "場のカード点: 120")
}

func TestSixBidSoloCuiPresenter_ErrorAndGameEnd(t *testing.T) {
	out := new(presenter.SixBidSoloCuiPresenter).Output(setupSixBidSoloCuiMock(defaultSixBidSoloOpts()), errors.New("boom"))
	assert.Contains(t, out, "boom")

	o := defaultSixBidSoloOpts()
	o.phase = domain.SixBidSoloPhaseGameEnd
	o.gameEnd = true
	o.winner = 0
	assert.Contains(t, new(presenter.SixBidSoloCuiPresenter).Output(setupSixBidSoloCuiMock(o), nil), "ゲーム終了")
}

func TestSixBidSoloCuiPresenter_ActionLogOutput(t *testing.T) {
	m := setupSixBidSoloCuiMock(defaultSixBidSoloOpts())
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotNil(t, new(presenter.SixBidSoloCuiPresenter).ActionLogOutput(m))
}
