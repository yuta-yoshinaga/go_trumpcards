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

func knTestCard(suit, value int) *domain.Card { return domain.NewCard(suit, value, true) }

// makeKarnoffelPlayers builds four seats, the first human, with the given hands.
func makeKarnoffelPlayers(hands ...[]*domain.Card) []*domain.KarnoffelPlayer {
	out := make([]*domain.KarnoffelPlayer, 0, len(hands))
	for i, hand := range hands {
		p := domain.NewKarnoffelPlayer(i == 0)
		for _, c := range hand {
			p.AddCard(c)
		}
		out = append(out, p)
	}
	return out
}

// karnoffelMockOpts tunes the parts of the stub that individual tests vary.
type karnoffelMockOpts struct {
	phase   domain.KarnoffelPhase
	chosen  int
	result  *domain.KarnoffelHandResult
	gameEnd bool
	winner  int
}

func setupKarnoffelWebMock(o karnoffelMockOpts) *interfaces.MockKarnoffelGame {
	m := new(interfaces.MockKarnoffelGame)
	players := makeKarnoffelPlayers(
		[]*domain.Card{knTestCard(domain.CardDesignSpade, 13)},
		[]*domain.Card{knTestCard(domain.CardDesignHeart, 11)},
		[]*domain.Card{knTestCard(domain.CardDesignClover, 6)},
		[]*domain.Card{knTestCard(domain.CardDesignDiamond, 2)},
	)
	m.On("GetPhase").Return(o.phase)
	m.On("GetHandNumber").Return(2)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(3)
	m.On("GetChosenSuit").Return(o.chosen)
	m.On("GetTrick").Return([]*domain.Card{knTestCard(domain.CardDesignSpade, 9), nil})
	m.On("GetTrickLeaderIdx").Return(0)
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
		m.On("GetTricksWon", i).Return(i)
		// **表向きの札は席ごとに 1 枚。**切札の根拠がここにある。
		m.On("GetUpCard", i).Return(knTestCard(domain.CardDesignHeart, 3+i))
	}
	for team := range domain.KarnoffelTeamCnt {
		m.On("KarnoffelTeamTricks", team).Return(1 + team)
		m.On("GetHandsWon", team).Return(team)
	}
	return m
}

func defaultKarnoffelOpts() karnoffelMockOpts {
	return karnoffelMockOpts{
		phase:  domain.KarnoffelPhasePlay,
		chosen: domain.CardDesignHeart,
		winner: -1,
	}
}

func parseKarnoffelOutput(t *testing.T, s string) *controller.KarnoffelWebOutput {
	t.Helper()
	var out controller.KarnoffelWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return &out
}

// **表向きの札は全員ぶん公開される。**切札の根拠が見えないと盤面が読めない。
func TestKarnoffelWebPresenter_SendsEveryFaceUpCard(t *testing.T) {
	out := parseKarnoffelOutput(t, new(presenter.KarnoffelWebPresenter).Output(setupKarnoffelWebMock(defaultKarnoffelOpts()), nil))
	assert.Len(t, out.Players, 4)
	for i := range out.Players {
		assert.NotNil(t, out.Players[i].UpCard, "seat %d's face-up card must be visible to everyone", i)
	}
	assert.Equal(t, domain.CardDesignHeart, out.ChosenSuit)
}

// 手札そのものはプレイ中は見えない。
func TestKarnoffelWebPresenter_HidesTheOtherHands(t *testing.T) {
	out := parseKarnoffelOutput(t, new(presenter.KarnoffelWebPresenter).Output(setupKarnoffelWebMock(defaultKarnoffelOpts()), nil))
	assert.Len(t, out.Players[0].Cards, 1, "the human sees its own hand")
	for i := 1; i < 4; i++ {
		assert.Empty(t, out.Players[i].Cards, "seat %d must stay hidden", i)
		assert.Positive(t, out.Players[i].CardCount)
	}

	o := defaultKarnoffelOpts()
	o.phase = domain.KarnoffelPhaseHandEnd
	revealed := parseKarnoffelOutput(t, new(presenter.KarnoffelWebPresenter).Output(setupKarnoffelWebMock(o), nil))
	for i := range revealed.Players {
		assert.NotEmpty(t, revealed.Players[i].Cards, "seat %d must be revealed", i)
	}
}

// **5 枚配りで 3 トリック先取。**issue の 12 枚とは違う。
func TestKarnoffelWebPresenter_SendsTheHandShape(t *testing.T) {
	out := parseKarnoffelOutput(t, new(presenter.KarnoffelWebPresenter).Output(setupKarnoffelWebMock(defaultKarnoffelOpts()), nil))
	assert.Equal(t, domain.KarnoffelHandSize, out.HandSize)
	assert.Equal(t, 5, out.HandSize, "five each, not twelve")
	assert.Equal(t, domain.KarnoffelTricksToWin, out.TricksToWin)
	assert.Equal(t, 3, out.TricksToWin)
	assert.Equal(t, domain.KarnoffelDefaultTarget, out.TargetHands)
}

// **出せる札はサーバーが決める。**第 1 トリックのリードに悪魔は使えない。
func TestKarnoffelWebPresenter_SendsValidPlaysOnlyOnTheHumansPlayTurn(t *testing.T) {
	out := parseKarnoffelOutput(t, new(presenter.KarnoffelWebPresenter).Output(setupKarnoffelWebMock(defaultKarnoffelOpts()), nil))
	assert.Equal(t, []int{0}, out.ValidPlays)

	o := defaultKarnoffelOpts()
	o.phase = domain.KarnoffelPhaseHandEnd
	settled := setupKarnoffelWebMock(o)
	settledOut := parseKarnoffelOutput(t, new(presenter.KarnoffelWebPresenter).Output(settled, nil))
	assert.Empty(t, settledOut.ValidPlays)
	settled.AssertNotCalled(t, "KarnoffelValidPlays", 0)
}

func TestKarnoffelWebPresenter_TopLevelFields(t *testing.T) {
	out := parseKarnoffelOutput(t, new(presenter.KarnoffelWebPresenter).Output(setupKarnoffelWebMock(defaultKarnoffelOpts()), nil))

	assert.Equal(t, int(domain.KarnoffelPhasePlay), out.Phase)
	assert.Equal(t, 2, out.HandNumber)
	// nil をまぜたトリックでも落ちない。
	assert.Len(t, out.Trick, 1)
	assert.Equal(t, 2, out.TrickNumber)
	assert.Equal(t, [domain.KarnoffelTeamCnt]int{1, 2}, out.TeamTricks)
	assert.Equal(t, [domain.KarnoffelTeamCnt]int{0, 1}, out.HandsWon)
	// パートナーは向かい合わせ。
	assert.Equal(t, out.Players[0].Team, out.Players[2].Team)
	assert.NotEqual(t, out.Players[0].Team, out.Players[1].Team)
	assert.True(t, out.Players[3].IsDealer)
	assert.True(t, out.Players[0].IsCurrentTurn)
}

// **3 トリックに届かなければ勝者なし。**その区別が伝わること。
func TestKarnoffelWebPresenter_SendsTheHandResult(t *testing.T) {
	o := defaultKarnoffelOpts()
	o.phase = domain.KarnoffelPhaseHandEnd
	o.result = &domain.KarnoffelHandResult{
		WinnerTeam: 0,
		Tricks:     [domain.KarnoffelTeamCnt]int{3, 2},
		ChosenSuit: domain.CardDesignHeart,
	}
	out := parseKarnoffelOutput(t, new(presenter.KarnoffelWebPresenter).Output(setupKarnoffelWebMock(o), nil))
	assert.NotNil(t, out.LastResult)
	assert.Equal(t, 0, out.LastResult.WinnerTeam)
	assert.Equal(t, [domain.KarnoffelTeamCnt]int{3, 2}, out.LastResult.Tricks)
	assert.Equal(t, domain.CardDesignHeart, out.LastResult.ChosenSuit)
	assert.Equal(t, "karnoffel.handEnd", out.MessageCode)

	drawn := defaultKarnoffelOpts()
	drawn.phase = domain.KarnoffelPhaseHandEnd
	drawn.result = &domain.KarnoffelHandResult{WinnerTeam: -1, Tricks: [domain.KarnoffelTeamCnt]int{2, 2}}
	drawnOut := parseKarnoffelOutput(t, new(presenter.KarnoffelWebPresenter).Output(setupKarnoffelWebMock(drawn), nil))
	assert.Equal(t, -1, drawnOut.LastResult.WinnerTeam)
	assert.Equal(t, "karnoffel.handDrawn", drawnOut.MessageCode)
}

func TestKarnoffelWebPresenter_Messages(t *testing.T) {
	for _, tc := range []struct {
		phase   domain.KarnoffelPhase
		result  *domain.KarnoffelHandResult
		wantKey string
	}{
		{domain.KarnoffelPhasePlay, nil, "karnoffel.playPhase"},
		{domain.KarnoffelPhaseHandEnd, &domain.KarnoffelHandResult{WinnerTeam: 1}, "karnoffel.handEnd"},
		{domain.KarnoffelPhaseHandEnd, &domain.KarnoffelHandResult{WinnerTeam: -1}, "karnoffel.handDrawn"},
		// 結果が無い局面でも落ちない。
		{domain.KarnoffelPhaseHandEnd, nil, "karnoffel.handEnd"},
	} {
		o := defaultKarnoffelOpts()
		o.phase = tc.phase
		o.result = tc.result
		out := parseKarnoffelOutput(t, new(presenter.KarnoffelWebPresenter).Output(setupKarnoffelWebMock(o), nil))
		assert.Equal(t, tc.wantKey, out.MessageCode)
	}

	t.Run("an error wins over any phase message", func(t *testing.T) {
		out := parseKarnoffelOutput(t, new(presenter.KarnoffelWebPresenter).Output(setupKarnoffelWebMock(defaultKarnoffelOpts()), errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
		assert.Empty(t, out.MessageCode)
	})
}

// **勝敗はチームで決まる。**人間は席 0 = チーム 0。
func TestKarnoffelWebPresenter_GameEndIsByTeam(t *testing.T) {
	for _, tc := range []struct {
		name    string
		team    int
		wantKey string
	}{
		{"the human's team wins", 0, "karnoffel.result.humanWin"},
		{"the other team wins", 1, "karnoffel.result.cpuWin"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := defaultKarnoffelOpts()
			o.phase = domain.KarnoffelPhaseGameEnd
			o.gameEnd = true
			o.winner = tc.team
			out := parseKarnoffelOutput(t, new(presenter.KarnoffelWebPresenter).Output(setupKarnoffelWebMock(o), nil))
			assert.Equal(t, tc.wantKey, out.MessageCode)
			assert.True(t, out.GameEndFlag)
			assert.Equal(t, tc.team, out.WinnerTeam)
		})
	}
}

func TestKarnoffelWebPresenter_ActionLogOutput(t *testing.T) {
	m := setupKarnoffelWebMock(defaultKarnoffelOpts())
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotEmpty(t, new(presenter.KarnoffelWebPresenter).ActionLogOutput(m))
}
