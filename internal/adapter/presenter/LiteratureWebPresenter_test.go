//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func ltTestCard(suit, value int) *domain.Card { return domain.NewCard(suit, value, true) }

// makeLiteraturePlayers builds six seats, the first human, with the given hands.
func makeLiteraturePlayers(hands ...[]*domain.Card) []*domain.LiteraturePlayer {
	out := make([]*domain.LiteraturePlayer, 0, len(hands))
	for i, hand := range hands {
		p := domain.NewLiteraturePlayer(i == 0)
		for _, c := range hand {
			p.AddCard(c)
		}
		out = append(out, p)
	}
	return out
}

// literatureMockOpts tunes the parts of the stub that individual tests vary.
type literatureMockOpts struct {
	gameEnd   bool
	winner    int
	lastAsk   *domain.LiteratureAsk
	lastClaim *domain.LiteratureClaimResult
	states    [domain.LiteratureHalfSuitCnt]domain.LiteratureHalfSuitState
	team0     int
	team1     int
	cancelled int
	open      int
}

func setupLiteratureWebMock(o literatureMockOpts) *interfaces.MockLiteratureGame {
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
		nil, // nil 混入でも落ちないこと
	})
	m.On("GetClaims").Return([]*domain.LiteratureClaimResult{
		{Player: 0, HalfSuit: 1, Outcome: domain.LiteratureClaimCancelled, AwardedTeam: -1},
		nil,
	})
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

func defaultLiteratureOpts() literatureMockOpts {
	return literatureMockOpts{winner: -1, open: domain.LiteratureHalfSuitCnt}
}

func parseLiteratureOutput(t *testing.T, s string) *controller.LiteratureWebOutput {
	t.Helper()
	var out controller.LiteratureWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return &out
}

// **終局まで誰の手札も公開しない。**味方の手札が見えたら推理が成立しない。
func TestLiteratureWebPresenter_HidesEveryHandIncludingTeammates(t *testing.T) {
	out := parseLiteratureOutput(t, new(presenter.LiteratureWebPresenter).Output(setupLiteratureWebMock(defaultLiteratureOpts()), nil))
	assert.Len(t, out.Players, 6)
	assert.Len(t, out.Players[0].Cards, 1, "the human sees its own hand")
	for i := 1; i < 6; i++ {
		assert.Empty(t, out.Players[i].Cards, "seat %d must stay hidden", i)
		assert.Positive(t, out.Players[i].CardCount)
	}
	// **味方 (席 2・4) も伏せたまま。**
	assert.Empty(t, out.Players[2].Cards, "a teammate's hand is concealed too")
	assert.Empty(t, out.Players[4].Cards)

	o := defaultLiteratureOpts()
	o.gameEnd = true
	revealed := parseLiteratureOutput(t, new(presenter.LiteratureWebPresenter).Output(setupLiteratureWebMock(o), nil))
	for i := range revealed.Players {
		assert.NotEmpty(t, revealed.Players[i].Cards, "seat %d is revealed at the end", i)
	}
}

// **席は交互。**味方に要求できない規則の前提。
func TestLiteratureWebPresenter_SendsAlternatingTeams(t *testing.T) {
	out := parseLiteratureOutput(t, new(presenter.LiteratureWebPresenter).Output(setupLiteratureWebMock(defaultLiteratureOpts()), nil))
	for i := range out.Players {
		assert.Equal(t, i%2, out.Players[i].Team, "seat %d", i)
	}
	assert.NotEqual(t, out.Players[0].Team, out.Players[1].Team)
	assert.Equal(t, out.Players[0].Team, out.Players[2].Team)
}

// **要求の履歴は公開情報。**推理の材料なので全部送る。
func TestLiteratureWebPresenter_SendsTheAskHistory(t *testing.T) {
	out := parseLiteratureOutput(t, new(presenter.LiteratureWebPresenter).Output(setupLiteratureWebMock(defaultLiteratureOpts()), nil))
	// nil 混入でも落ちない。
	assert.Len(t, out.Asks, 1)
	assert.Equal(t, 0, out.Asks[0].From)
	assert.Equal(t, 1, out.Asks[0].To)
	assert.True(t, out.Asks[0].Success)
	assert.NotNil(t, out.Asks[0].Card)

	assert.Len(t, out.Claims, 1)
	assert.Equal(t, int(domain.LiteratureClaimCancelled), out.Claims[0].Outcome)
}

// **勝利には 5 組。**過半数であることをワイヤでも示す。
func TestLiteratureWebPresenter_SendsTheWinThreshold(t *testing.T) {
	out := parseLiteratureOutput(t, new(presenter.LiteratureWebPresenter).Output(setupLiteratureWebMock(defaultLiteratureOpts()), nil))
	assert.Equal(t, domain.LiteratureWinThreshold, out.WinThreshold)
	assert.Equal(t, 5, out.WinThreshold, "a majority of eight is five, not four")
	assert.Equal(t, domain.LiteratureHalfSuitCnt, out.HalfSuitCnt)
	// 各組の 6 枚も送る (選択肢を出すのに要る)。
	for half := range domain.LiteratureHalfSuitCnt {
		assert.Len(t, out.HalfSuitCards[half], domain.LiteratureHalfSuitSize, "half-suit %d", half)
	}
}

// **無効があるので合計が 8 になるとは限らない。**別枠で送る。
func TestLiteratureWebPresenter_SendsCancelledSeparately(t *testing.T) {
	o := defaultLiteratureOpts()
	o.team0, o.team1, o.cancelled, o.open = 3, 2, 1, 2
	o.states[0] = domain.LiteratureHalfCancelled
	out := parseLiteratureOutput(t, new(presenter.LiteratureWebPresenter).Output(setupLiteratureWebMock(o), nil))

	assert.Equal(t, [domain.LiteratureTeamCnt]int{3, 2}, out.TeamHalfSuits)
	assert.Equal(t, 1, out.CancelledCount)
	assert.Equal(t, 2, out.OpenCount)
	// 帰属そのものも送る。
	assert.Equal(t, int(domain.LiteratureHalfCancelled), out.HalfSuits[0])
	// 3 + 2 + 1 + 2 = 8。
	assert.Equal(t, domain.LiteratureHalfSuitCnt,
		out.TeamHalfSuits[0]+out.TeamHalfSuits[1]+out.CancelledCount+out.OpenCount)
}

// **宣言の結末は 3 通り。**「無効」は「相手に渡る」とは違う。
func TestLiteratureWebPresenter_DistinguishesTheThreeClaimOutcomes(t *testing.T) {
	for _, tc := range []struct {
		outcome domain.LiteratureClaimOutcome
		wantKey string
	}{
		{domain.LiteratureClaimWon, "literature.claimWon"},
		{domain.LiteratureClaimCancelled, "literature.claimCancelled"},
		{domain.LiteratureClaimLost, "literature.claimLost"},
	} {
		o := defaultLiteratureOpts()
		o.lastClaim = &domain.LiteratureClaimResult{Player: 0, HalfSuit: 0, Outcome: tc.outcome, AwardedTeam: -1}
		out := parseLiteratureOutput(t, new(presenter.LiteratureWebPresenter).Output(setupLiteratureWebMock(o), nil))
		assert.Equal(t, tc.wantKey, out.MessageCode)
	}
}

func TestLiteratureWebPresenter_AskMessages(t *testing.T) {
	for _, tc := range []struct {
		success bool
		wantKey string
	}{
		{true, "literature.askHit"},
		{false, "literature.askMiss"},
	} {
		o := defaultLiteratureOpts()
		o.lastAsk = &domain.LiteratureAsk{From: 0, To: 1, Card: ltTestCard(domain.CardDesignSpade, 3), Success: tc.success}
		out := parseLiteratureOutput(t, new(presenter.LiteratureWebPresenter).Output(setupLiteratureWebMock(o), nil))
		assert.Equal(t, tc.wantKey, out.MessageCode)
	}

	// 何も起きていなければ進行中。
	out := parseLiteratureOutput(t, new(presenter.LiteratureWebPresenter).Output(setupLiteratureWebMock(defaultLiteratureOpts()), nil))
	assert.Equal(t, "literature.playPhase", out.MessageCode)

	t.Run("an error wins over any phase message", func(t *testing.T) {
		out := parseLiteratureOutput(t, new(presenter.LiteratureWebPresenter).Output(setupLiteratureWebMock(defaultLiteratureOpts()), errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
		assert.Empty(t, out.MessageCode)
	})
}

// **同数で終わることがある。**無効が絡むため。
func TestLiteratureWebPresenter_GameEnd(t *testing.T) {
	for _, tc := range []struct {
		name    string
		team    int
		wantKey string
	}{
		{"the human's team wins", 0, "literature.result.humanWin"},
		{"the other team wins", 1, "literature.result.cpuWin"},
		{"nobody wins", -1, "literature.result.draw"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := defaultLiteratureOpts()
			o.gameEnd = true
			o.winner = tc.team
			out := parseLiteratureOutput(t, new(presenter.LiteratureWebPresenter).Output(setupLiteratureWebMock(o), nil))
			assert.Equal(t, tc.wantKey, out.MessageCode)
			assert.True(t, out.GameEndFlag)
			assert.Equal(t, tc.team, out.WinnerTeam)
		})
	}
}

func TestLiteratureWebPresenter_ActionLogOutput(t *testing.T) {
	m := setupLiteratureWebMock(defaultLiteratureOpts())
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotEmpty(t, new(presenter.LiteratureWebPresenter).ActionLogOutput(m))
}
