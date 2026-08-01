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

func gdTestCard(suit, value int) *domain.Card { return domain.NewCard(suit, value, true) }

// makeGuandanPlayers builds four seats, the first human, with the given hands.
func makeGuandanPlayers(hands ...[]*domain.Card) []*domain.GuandanPlayer {
	out := make([]*domain.GuandanPlayer, 0, len(hands))
	for i, hand := range hands {
		p := domain.NewGuandanPlayer(i == 0)
		for _, c := range hand {
			p.AddCard(c)
		}
		out = append(out, p)
	}
	return out
}

// guandanMockOpts tunes the parts of the stub that individual tests vary.
type guandanMockOpts struct {
	phase      domain.GuandanPhase
	gameEnd    bool
	winner     int
	level      int
	combo      *domain.GuandanCombo
	finished   []int
	tributes   []*domain.GuandanTribute
	cancelled  bool
	lastResult *domain.GuandanHandResult
}

func setupGuandanMock(o guandanMockOpts) *interfaces.MockGuandanGame {
	m := new(interfaces.MockGuandanGame)
	players := makeGuandanPlayers(
		[]*domain.Card{gdTestCard(domain.CardDesignSpade, 2), gdTestCard(domain.CardDesignHeart, 1)},
		[]*domain.Card{gdTestCard(domain.CardDesignSpade, 3)},
		[]*domain.Card{gdTestCard(domain.CardDesignSpade, 4)},
		[]*domain.Card{gdTestCard(domain.CardDesignSpade, 5)},
	)
	m.On("GetPhase").Return(o.phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetGameEndFlag").Return(o.gameEnd)
	m.On("GetWinnerTeam").Return(o.winner)
	m.On("GetConfig").Return(domain.DefaultGuandanConfig())
	m.On("GetPlayers").Return(players)
	m.On("GetHandNumber").Return(3)
	m.On("GetLevel").Return(o.level)
	m.On("GetTeamLevel", 0).Return(o.level)
	m.On("GetTeamLevel", 1).Return(domain.GuandanMinLevel)
	m.On("GetDeclarerTeam").Return(0)
	m.On("GetLastCombo").Return(o.combo)
	m.On("GetLastPlayerIdx").Return(1)
	m.On("GetFinished").Return(o.finished)
	m.On("GetTributes").Return(o.tributes)
	m.On("IsTributeCancelled").Return(o.cancelled)
	m.On("GetLastResult").Return(o.lastResult)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m
}

func defaultGuandanOpts() guandanMockOpts {
	return guandanMockOpts{
		phase:    domain.GuandanPhasePlay,
		winner:   -1,
		level:    domain.GuandanMinLevel,
		finished: []int{},
		tributes: []*domain.GuandanTribute{},
	}
}

func parseGuandanOutput(t *testing.T, s string) *controller.GuandanWebOutput {
	t.Helper()
	out := new(controller.GuandanWebOutput)
	assert.NoError(t, json.Unmarshal([]byte(s), out))
	return out
}

func TestGuandanWebPresenter_Output(t *testing.T) {
	p := new(presenter.GuandanWebPresenter)

	// **他人の手札は伏せたまま。**枚数だけ出す。
	t.Run("only the human's hand is revealed while the hand is in play", func(t *testing.T) {
		out := parseGuandanOutput(t, p.Output(setupGuandanMock(defaultGuandanOpts()), nil))
		assert.Len(t, out.Players, domain.GuandanPlayerCnt)
		assert.Len(t, out.Players[0].Cards, 2)
		assert.Equal(t, 2, out.Players[0].CardCount)
		assert.Empty(t, out.Players[1].Cards)
		assert.Equal(t, 1, out.Players[1].CardCount)
		assert.True(t, out.Players[0].IsCurrentTurn)
		assert.False(t, out.Players[1].IsCurrentTurn)
	})

	// **パートナーは向かい合わせ。**0/2 と 1/3。
	t.Run("teams pair the seats across the table", func(t *testing.T) {
		out := parseGuandanOutput(t, p.Output(setupGuandanMock(defaultGuandanOpts()), nil))
		assert.Equal(t, []int{0, 1, 0, 1},
			[]int{out.Players[0].Team, out.Players[1].Team, out.Players[2].Team, out.Players[3].Team})
	})

	t.Run("every hand is revealed once the hand is settled", func(t *testing.T) {
		o := defaultGuandanOpts()
		o.phase = domain.GuandanPhaseHandEnd
		out := parseGuandanOutput(t, p.Output(setupGuandanMock(o), nil))
		assert.Len(t, out.Players[1].Cards, 1)
	})

	// **着順は次局の貢と昇級量を決める。**
	t.Run("the finishing order is reported per seat", func(t *testing.T) {
		o := defaultGuandanOpts()
		o.finished = []int{2, 0}
		out := parseGuandanOutput(t, p.Output(setupGuandanMock(o), nil))
		assert.Equal(t, 2, out.Players[0].FinishedRank)
		assert.Equal(t, 1, out.Players[2].FinishedRank)
		assert.Equal(t, 0, out.Players[1].FinishedRank)
		assert.Equal(t, []int{2, 0}, out.Finished)
	})

	t.Run("the table combination is carried over", func(t *testing.T) {
		o := defaultGuandanOpts()
		o.combo = &domain.GuandanCombo{Kind: domain.GuandanComboBomb, Rank: 9, Size: 4}
		out := parseGuandanOutput(t, p.Output(setupGuandanMock(o), nil))
		assert.NotNil(t, out.LastCombo)
		assert.Equal(t, int(domain.GuandanComboBomb), out.LastCombo.Kind)
		assert.Equal(t, 4, out.LastCombo.Size)
		assert.Equal(t, 1, out.LastPlayerIdx)
	})

	t.Run("an empty table has no combination", func(t *testing.T) {
		out := parseGuandanOutput(t, p.Output(setupGuandanMock(defaultGuandanOpts()), nil))
		assert.Nil(t, out.LastCombo)
	})

	// **貢は次局の手札を動かす。**返した札まで見えないと不公平に見える。
	t.Run("tributes carry both the paid and the returned card", func(t *testing.T) {
		o := defaultGuandanOpts()
		o.phase = domain.GuandanPhaseTribute
		o.tributes = []*domain.GuandanTribute{
			{From: 3, To: 0, Card: gdTestCard(domain.CardDesignSpade, 1), Returned: gdTestCard(domain.CardDesignClover, 2)},
			{From: 2, To: 1, Card: gdTestCard(domain.CardDesignHeart, 13)},
			nil, // nil 混入でも落ちないこと
		}
		out := parseGuandanOutput(t, p.Output(setupGuandanMock(o), nil))
		assert.Len(t, out.Tributes, 2)
		assert.Equal(t, 3, out.Tributes[0].From)
		assert.NotNil(t, out.Tributes[0].Returned)
		assert.Nil(t, out.Tributes[1].Returned)
		assert.Equal(t, "guandan.tributePhase", out.MessageCode)
	})

	t.Run("a cancelled tribute says so", func(t *testing.T) {
		o := defaultGuandanOpts()
		o.phase = domain.GuandanPhaseTribute
		o.cancelled = true
		out := parseGuandanOutput(t, p.Output(setupGuandanMock(o), nil))
		assert.True(t, out.TributeCancelled)
		assert.Equal(t, "guandan.tributeCancelled", out.MessageCode)
	})

	// **上昇量は 1 / 2 / 4。**表を送らないとフロントが説明できない。
	t.Run("the advance table is published", func(t *testing.T) {
		out := parseGuandanOutput(t, p.Output(setupGuandanMock(defaultGuandanOpts()), nil))
		assert.Equal(t, domain.GuandanAdvanceFirstSecond, out.AdvanceFirstSecond)
		assert.Equal(t, domain.GuandanAdvanceFirstThird, out.AdvanceFirstThird)
		assert.Equal(t, domain.GuandanAdvanceFirstFourth, out.AdvanceFirstFourth)
		assert.Equal(t, domain.GuandanMinLevel, out.MinLevel)
		assert.Equal(t, domain.GuandanMaxLevel, out.MaxLevel)
	})

	t.Run("the settled hand carries its result", func(t *testing.T) {
		o := defaultGuandanOpts()
		o.phase = domain.GuandanPhaseHandEnd
		o.lastResult = &domain.GuandanHandResult{
			Order: [domain.GuandanPlayerCnt]int{0, 2, 1, 3}, WinnerTeam: 0,
			Advance: domain.GuandanAdvanceFirstSecond, FirstSecond: true,
		}
		out := parseGuandanOutput(t, p.Output(setupGuandanMock(o), nil))
		assert.NotNil(t, out.LastResult)
		assert.Equal(t, domain.GuandanAdvanceFirstSecond, out.LastResult.Advance)
		assert.True(t, out.LastResult.FirstSecond)
		assert.Equal(t, "guandan.handFirstSecond", out.MessageCode)
	})

	t.Run("an ordinary hand end uses the plain code", func(t *testing.T) {
		o := defaultGuandanOpts()
		o.phase = domain.GuandanPhaseHandEnd
		o.lastResult = &domain.GuandanHandResult{WinnerTeam: 1, Advance: domain.GuandanAdvanceFirstFourth}
		out := parseGuandanOutput(t, p.Output(setupGuandanMock(o), nil))
		assert.Equal(t, "guandan.handEnd", out.MessageCode)
	})

	// **勝敗は席ではなくチームで見る。**人間は席 0 = チーム 0。
	t.Run("the winning team decides the closing message", func(t *testing.T) {
		o := defaultGuandanOpts()
		o.gameEnd = true
		o.winner = 0
		out := parseGuandanOutput(t, p.Output(setupGuandanMock(o), nil))
		assert.Equal(t, "guandan.result.humanWin", out.MessageCode)
		assert.Equal(t, 0, out.WinnerTeam)

		o.winner = 1
		out = parseGuandanOutput(t, p.Output(setupGuandanMock(o), nil))
		assert.Equal(t, "guandan.result.cpuWin", out.MessageCode)
	})

	t.Run("an error is surfaced verbatim", func(t *testing.T) {
		out := parseGuandanOutput(t, p.Output(setupGuandanMock(defaultGuandanOpts()), errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
		assert.Empty(t, out.MessageCode)
	})
}

func TestGuandanWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.GuandanWebPresenter)
	assert.JSONEq(t, `{"entries":[]}`, p.ActionLogOutput(setupGuandanMock(defaultGuandanOpts())))
}
