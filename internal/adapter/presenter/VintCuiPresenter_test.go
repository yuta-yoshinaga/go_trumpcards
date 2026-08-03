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

func setupVintCuiMock(o vintMockOpts) *interfaces.MockVintGame {
	m := new(interfaces.MockVintGame)
	players := makeVintPlayers(
		[]*domain.Card{vtTestCard(domain.CardDesignSpade, 1)},
		[]*domain.Card{vtTestCard(domain.CardDesignHeart, 2)},
		[]*domain.Card{vtTestCard(domain.CardDesignClover, 3)},
		[]*domain.Card{vtTestCard(domain.CardDesignDiamond, 4)},
	)
	m.On("GetPhase").Return(o.phase)
	m.On("GetHandNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(3)
	m.On("GetDeclarerIdx").Return(o.declarer)
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetTrick").Return([]*domain.Card{vtTestCard(domain.CardDesignSpade, 5)})
	m.On("GetTrickNumber").Return(3)
	m.On("GetGameEndFlag").Return(o.gameEnd)
	m.On("GetWinnerTeam").Return(o.winner)
	m.On("GetHighBid").Return(o.highBid)
	m.On("GetLastResult").Return(o.result)
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(true)
	m.On("VintValidPlays", 0).Return([]int{0})
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
		m.On("GetTricksWon", i).Return(0)
	}
	for team := range domain.VintTeamCnt {
		m.On("GetBelow", team).Return(0)
		m.On("GetAbove", team).Return(0)
		m.On("GetGamesWon", team).Return(0)
	}
	return m
}

// **ダミーが無いので、プレイ中は誰の手札も見えない。**
func TestVintCuiPresenter_HasNoDummy(t *testing.T) {
	out := new(presenter.VintCuiPresenter).Output(setupVintCuiMock(defaultVintOpts()), nil)
	assert.Contains(t, out, "[0]")
	assert.Contains(t, out, "非公開")
	assert.Contains(t, out, "[親]")
	assert.Contains(t, out, "[宣言]")
	// 味方の手札も伏せられる (「非公開」が 3 席ぶん出る)。
	if got := strings.Count(out, "非公開"); got != 3 {
		t.Errorf("%d hidden hands, want 3 — there is no dummy in Vint", got)
	}
}

// **♠ が最弱で NT が最強。**ブリッジと逆なので必ず見せる。
func TestVintCuiPresenter_ShowsTheDenominationLadderWhileBidding(t *testing.T) {
	o := defaultVintOpts()
	o.phase = domain.VintPhaseBid
	o.highBid = nil
	out := new(presenter.VintCuiPresenter).Output(setupVintCuiMock(o), nil)

	assert.Contains(t, out, "序列")
	// 並びは 0:♠ < 1:♣ < 2:♦ < 3:♥ < 4:NT。
	spade := strings.Index(out, "0:♠")
	club := strings.Index(out, "1:♣")
	nt := strings.Index(out, "4:")
	assert.NotEqual(t, -1, spade)
	assert.Less(t, spade, club, "spades come first — they are the LOWEST")
	assert.Less(t, club, nt)
	// レベル 1 の単価も出す。
	assert.Contains(t, out, "(4)")
	assert.Contains(t, out, "(12)")
	assert.Contains(t, out, "ブリッジとは逆")
}

// **出せる札を出さないと操作できない。**追随が強制。
func TestVintCuiPresenter_ListsThePlayableIndexes(t *testing.T) {
	out := new(presenter.VintCuiPresenter).Output(setupVintCuiMock(defaultVintOpts()), nil)
	assert.Contains(t, out, "出せる札: 0")
}

func TestVintCuiPresenter_ShowsTheContractOnlyOnceBid(t *testing.T) {
	withBid := new(presenter.VintCuiPresenter).Output(setupVintCuiMock(defaultVintOpts()), nil)
	assert.Contains(t, withBid, "契約:")
	// 3♥ の単価は 10 + 20 = 30。
	assert.Contains(t, withBid, "30")

	o := defaultVintOpts()
	o.highBid = nil
	assert.NotContains(t, new(presenter.VintCuiPresenter).Output(setupVintCuiMock(o), nil), "契約:")
}

// **両チームの線下加点を出す。**守備側も得点することが読めないと意味が通らない。
func TestVintCuiPresenter_ShowsBothSidesTrickPoints(t *testing.T) {
	o := defaultVintOpts()
	o.phase = domain.VintPhaseHandEnd
	o.result = &domain.VintHandResult{
		TrickPoints:    [domain.VintTeamCnt]int{210, 180},
		Made:           true,
		DeclarerTricks: 9,
		TrickValue:     30,
	}
	out := new(presenter.VintCuiPresenter).Output(setupVintCuiMock(o), nil)
	assert.Contains(t, out, "宣言達成")
	assert.Contains(t, out, "210")
	assert.Contains(t, out, "180")
	assert.Contains(t, out, "両チームとも")
}

func TestVintCuiPresenter_TellsAFailedContractApart(t *testing.T) {
	o := defaultVintOpts()
	o.phase = domain.VintPhaseHandEnd
	o.result = &domain.VintHandResult{Made: false, DeclarerTricks: 6, TrickValue: 30}
	out := new(presenter.VintCuiPresenter).Output(setupVintCuiMock(o), nil)
	assert.Contains(t, out, "宣言失敗")
	assert.Contains(t, out, "不足数 × レベル × 500")
}

func TestVintCuiPresenter_PhasePrompts(t *testing.T) {
	for _, tc := range []struct {
		phase domain.VintPhase
		want  string
	}{
		{domain.VintPhaseBid, "序列"},
		{domain.VintPhasePlay, "追随は強制"},
		{domain.VintPhaseHandEnd, "次の局へ"},
	} {
		o := defaultVintOpts()
		o.phase = tc.phase
		assert.Contains(t, new(presenter.VintCuiPresenter).Output(setupVintCuiMock(o), nil), tc.want)
	}
}

// 線下・線上・ゲーム数がすべて読めること。
func TestVintCuiPresenter_ShowsTheScoreSheet(t *testing.T) {
	out := new(presenter.VintCuiPresenter).Output(setupVintCuiMock(defaultVintOpts()), nil)
	assert.Contains(t, out, "線下")
	assert.Contains(t, out, "線上")
	assert.Contains(t, out, "ゲーム")
}

func TestVintCuiPresenter_ErrorAndGameEnd(t *testing.T) {
	out := new(presenter.VintCuiPresenter).Output(setupVintCuiMock(defaultVintOpts()), errors.New("boom"))
	assert.Contains(t, out, "boom")

	o := defaultVintOpts()
	o.phase = domain.VintPhaseGameEnd
	o.gameEnd = true
	o.winner = 0
	assert.Contains(t, new(presenter.VintCuiPresenter).Output(setupVintCuiMock(o), nil), "ラバー終了")
}

func TestVintCuiPresenter_ActionLogOutput(t *testing.T) {
	m := setupVintCuiMock(defaultVintOpts())
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotNil(t, new(presenter.VintCuiPresenter).ActionLogOutput(m))
}
