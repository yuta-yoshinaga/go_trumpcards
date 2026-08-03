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

func setupLiteratureCuiMock(o literatureMockOpts) *interfaces.MockLiteratureGame {
	m := new(interfaces.MockLiteratureGame)
	players := makeLiteraturePlayers(
		[]*domain.Card{ltTestCard(domain.CardDesignSpade, 2)},
		[]*domain.Card{ltTestCard(domain.CardDesignSpade, 3)},
		[]*domain.Card{ltTestCard(domain.CardDesignSpade, 4)},
		[]*domain.Card{ltTestCard(domain.CardDesignSpade, 5)},
		[]*domain.Card{ltTestCard(domain.CardDesignSpade, 6)},
		[]*domain.Card{ltTestCard(domain.CardDesignSpade, 7)},
	)
	m.On("GetPhase").Return(domain.LiteraturePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetGameEndFlag").Return(o.gameEnd)
	m.On("GetWinnerTeam").Return(o.winner)
	m.On("GetConfig").Return(domain.DefaultLiteratureConfig())
	m.On("GetPlayers").Return(players)
	m.On("GetLastAsk").Return(o.lastAsk)
	m.On("GetLastClaim").Return(o.lastClaim)
	m.On("GetAsks").Return([]*domain.LiteratureAsk{
		{From: 0, To: 1, Card: ltTestCard(domain.CardDesignSpade, 3), Success: true},
		{From: 1, To: 0, Card: ltTestCard(domain.CardDesignHeart, 9), Success: false},
	})
	m.On("GetClaims").Return([]*domain.LiteratureClaimResult{})
	m.On("LiteratureTeamHalfSuits", 0).Return(o.team0)
	m.On("LiteratureTeamHalfSuits", 1).Return(o.team1)
	m.On("LiteratureCancelledCount").Return(o.cancelled)
	m.On("LiteratureOpenCount").Return(o.open)
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
	}
	for half := range domain.LiteratureHalfSuitCnt {
		m.On("GetHalfSuitState", half).Return(o.states[half])
	}
	return m
}

// **勝利には 5 組。**4 組では決着しないことを画面に書く。
func TestLiteratureCuiPresenter_StatesTheRealThreshold(t *testing.T) {
	out := new(presenter.LiteratureCuiPresenter).Output(setupLiteratureCuiMock(defaultLiteratureOpts()), nil)
	assert.Contains(t, out, "勝利には5組")
	assert.Contains(t, out, "8組の過半数=5組")
	assert.Contains(t, out, "4組では相手も4組になり得るので決着しません")
}

// **終局まで誰の手札も見えない。**味方も含めて。
func TestLiteratureCuiPresenter_HidesEveryHandIncludingTeammates(t *testing.T) {
	out := new(presenter.LiteratureCuiPresenter).Output(setupLiteratureCuiMock(defaultLiteratureOpts()), nil)
	if got := strings.Count(out, "非公開"); got != 5 {
		t.Errorf("%d hidden hands, want 5 — every seat but the human", got)
	}

	o := defaultLiteratureOpts()
	o.gameEnd = true
	assert.NotContains(t, new(presenter.LiteratureCuiPresenter).Output(setupLiteratureCuiMock(o), nil), "非公開")
}

// **要求の履歴は公開情報。**的中と空振りを区別して出す。
func TestLiteratureCuiPresenter_ShowsTheAskHistory(t *testing.T) {
	out := new(presenter.LiteratureCuiPresenter).Output(setupLiteratureCuiMock(defaultLiteratureOpts()), nil)
	assert.Contains(t, out, "全員に公開されています")
	assert.Contains(t, out, "的中")
	assert.Contains(t, out, "空振り")
}

// **要求の 4 条件を画面に書く。**味方に訊けないのが最も間違えやすい。
func TestLiteratureCuiPresenter_StatesTheAskConditions(t *testing.T) {
	out := new(presenter.LiteratureCuiPresenter).Output(setupLiteratureCuiMock(defaultLiteratureOpts()), nil)
	assert.Contains(t, out, "相手チームにのみ")
	assert.Contains(t, out, "自分がその組を1枚以上持つ")
	assert.Contains(t, out, "自分が持っていない札")
	assert.Contains(t, out, "相手に手札が残っている")
}

// **宣言の結末は 3 通り。**無効が「相手に渡る」と読めてはいけない。
func TestLiteratureCuiPresenter_DistinguishesTheThreeClaimOutcomes(t *testing.T) {
	won := defaultLiteratureOpts()
	won.lastClaim = &domain.LiteratureClaimResult{HalfSuit: 0, Outcome: domain.LiteratureClaimWon, AwardedTeam: 0}
	assert.Contains(t, new(presenter.LiteratureCuiPresenter).Output(setupLiteratureCuiMock(won), nil), "チーム0が獲得しました")

	cancelled := defaultLiteratureOpts()
	cancelled.lastClaim = &domain.LiteratureClaimResult{HalfSuit: 0, Outcome: domain.LiteratureClaimCancelled, AwardedTeam: -1}
	out := new(presenter.LiteratureCuiPresenter).Output(setupLiteratureCuiMock(cancelled), nil)
	assert.Contains(t, out, "【無効】になりました")
	assert.Contains(t, out, "相手には渡りません")

	lost := defaultLiteratureOpts()
	lost.lastClaim = &domain.LiteratureClaimResult{HalfSuit: 0, Outcome: domain.LiteratureClaimLost, AwardedTeam: 1}
	assert.Contains(t, new(presenter.LiteratureCuiPresenter).Output(setupLiteratureCuiMock(lost), nil), "相手が1枚以上持っていました")
}

// **無効は帰属の一覧でも区別される。**
func TestLiteratureCuiPresenter_ShowsHalfSuitOwnership(t *testing.T) {
	o := defaultLiteratureOpts()
	o.states[0] = domain.LiteratureHalfTeam0
	o.states[1] = domain.LiteratureHalfTeam1
	o.states[2] = domain.LiteratureHalfCancelled
	out := new(presenter.LiteratureCuiPresenter).Output(setupLiteratureCuiMock(o), nil)

	assert.Contains(t, out, "[0] ♠低位(2-7): チーム0")
	assert.Contains(t, out, "[1] ♠高位(9-A): チーム1")
	assert.Contains(t, out, "[2] ♣低位(2-7): 【無効】")
	assert.Contains(t, out, "[3] ♣高位(9-A): 未決")
}

// **同数で終わることがある。**無効が絡むため。
func TestLiteratureCuiPresenter_ErrorAndGameEnd(t *testing.T) {
	out := new(presenter.LiteratureCuiPresenter).Output(setupLiteratureCuiMock(defaultLiteratureOpts()), errors.New("boom"))
	assert.Contains(t, out, "boom")

	won := defaultLiteratureOpts()
	won.gameEnd = true
	won.winner = 0
	assert.Contains(t, new(presenter.LiteratureCuiPresenter).Output(setupLiteratureCuiMock(won), nil), "チーム0の勝ち")

	drawn := defaultLiteratureOpts()
	drawn.gameEnd = true
	drawn.winner = -1
	assert.Contains(t, new(presenter.LiteratureCuiPresenter).Output(setupLiteratureCuiMock(drawn), nil), "勝者なし")
}

func TestLiteratureCuiPresenter_ActionLogOutput(t *testing.T) {
	m := setupLiteratureCuiMock(defaultLiteratureOpts())
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotNil(t, new(presenter.LiteratureCuiPresenter).ActionLogOutput(m))
}
