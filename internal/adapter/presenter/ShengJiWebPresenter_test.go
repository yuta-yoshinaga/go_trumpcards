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

func sjiTestCard(suit, value int) *domain.Card { return domain.NewCard(suit, value, true) }

// makeShengJiPlayers builds four seats, the first human, with the given hands.
func makeShengJiPlayers(hands ...[]*domain.Card) []*domain.ShengJiPlayer {
	out := make([]*domain.ShengJiPlayer, 0, len(hands))
	for i, hand := range hands {
		p := domain.NewShengJiPlayer(i == 0)
		for _, c := range hand {
			p.AddCard(c)
		}
		out = append(out, p)
	}
	return out
}

// shengJiMockOpts tunes the parts of the stub that individual tests vary.
type shengJiMockOpts struct {
	phase       domain.ShengJiPhase
	gameEnd     bool
	winner      int
	trumpSuit   int
	declaration *domain.ShengJiDeclaration
	kitty       []*domain.Card
	trick       [][]*domain.Card
	leadCombo   *domain.ShengJiCombo
	lastResult  *domain.ShengJiHandResult
	declarable  map[int]int
	level       int
}

func setupShengJiMock(o shengJiMockOpts) *interfaces.MockShengJiGame {
	m := new(interfaces.MockShengJiGame)
	players := makeShengJiPlayers(
		[]*domain.Card{sjiTestCard(domain.CardDesignSpade, 2), sjiTestCard(domain.CardDesignHeart, 5)},
		[]*domain.Card{sjiTestCard(domain.CardDesignSpade, 3)},
		[]*domain.Card{sjiTestCard(domain.CardDesignSpade, 4)},
		[]*domain.Card{sjiTestCard(domain.CardDesignSpade, 6)},
	)
	m.On("GetPhase").Return(o.phase)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetGameEndFlag").Return(o.gameEnd)
	m.On("GetWinnerTeam").Return(o.winner)
	m.On("GetConfig").Return(domain.DefaultShengJiConfig())
	m.On("GetPlayers").Return(players)
	m.On("GetHandNumber").Return(3)
	m.On("GetLevel").Return(o.level)
	m.On("GetTeamLevel", 0).Return(o.level)
	m.On("GetTeamLevel", 1).Return(domain.ShengJiMinLevel)
	m.On("GetTeamPoints", 0).Return(0)
	m.On("GetTeamPoints", 1).Return(35)
	m.On("GetDeclarerTeam").Return(0)
	m.On("GetTrumpSuit").Return(o.trumpSuit)
	m.On("GetDeclaration").Return(o.declaration)
	m.On("GetKittySize").Return(len(o.kitty))
	m.On("GetKitty").Return(o.kitty)
	m.On("GetTrick").Return(o.trick)
	m.On("GetTrickLeader").Return(1)
	m.On("GetLeadCombo").Return(o.leadCombo)
	m.On("GetTrickCount").Return(4)
	m.On("GetLastTrickWinner").Return(2)
	m.On("GetLastResult").Return(o.lastResult)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
	}
	for suit := domain.CardDesignSpade; suit <= domain.CardDesignDiamond; suit++ {
		m.On("ShengJiDeclareStrength", 0, suit).Return(o.declarable[suit])
	}
	return m
}

func defaultShengJiOpts() shengJiMockOpts {
	return shengJiMockOpts{
		phase:     domain.ShengJiPhasePlay,
		winner:    -1,
		level:     5,
		trumpSuit: domain.CardDesignSpade,
		trick:     [][]*domain.Card{},
	}
}

func parseShengJiOutput(t *testing.T, s string) *controller.ShengJiWebOutput {
	t.Helper()
	out := new(controller.ShengJiWebOutput)
	assert.NoError(t, json.Unmarshal([]byte(s), out))
	return out
}

func TestShengJiWebPresenter_Output(t *testing.T) {
	p := new(presenter.ShengJiWebPresenter)

	t.Run("only the human's hand is revealed while the hand is in play", func(t *testing.T) {
		out := parseShengJiOutput(t, p.Output(setupShengJiMock(defaultShengJiOpts()), nil))
		assert.Len(t, out.Players, domain.ShengJiPlayerCnt)
		assert.Len(t, out.Players[0].Cards, 2)
		assert.Empty(t, out.Players[1].Cards)
		assert.Equal(t, 1, out.Players[1].CardCount)
		assert.True(t, out.Players[0].IsCurrentTurn)
	})

	// **どちら側かが見えないと打ちようがない。**宣言側と守備側で目的が逆。
	t.Run("each seat carries its side", func(t *testing.T) {
		out := parseShengJiOutput(t, p.Output(setupShengJiMock(defaultShengJiOpts()), nil))
		assert.True(t, out.Players[0].IsDeclarer)
		assert.False(t, out.Players[1].IsDeclarer)
		assert.True(t, out.Players[2].IsDeclarer)
		assert.Equal(t, []int{0, 1, 0, 1},
			[]int{out.Players[0].Team, out.Players[1].Team, out.Players[2].Team, out.Players[3].Team})
	})

	t.Run("every hand is revealed once the hand is settled", func(t *testing.T) {
		o := defaultShengJiOpts()
		o.phase = domain.ShengJiPhaseHandEnd
		out := parseShengJiOutput(t, p.Output(setupShengJiMock(o), nil))
		assert.Len(t, out.Players[1].Cards, 1)
	})

	// **底牌は終局まで送らない。**送ると宣言側の埋め方が筒抜けになる。
	t.Run("the kitty is a count until the hand is settled", func(t *testing.T) {
		o := defaultShengJiOpts()
		o.kitty = nil
		out := parseShengJiOutput(t, p.Output(setupShengJiMock(o), nil))
		assert.Empty(t, out.Kitty)

		o.phase = domain.ShengJiPhaseHandEnd
		o.kitty = []*domain.Card{sjiTestCard(domain.CardDesignDiamond, 13)}
		out = parseShengJiOutput(t, p.Output(setupShengJiMock(o), nil))
		assert.Len(t, out.Kitty, 1)
		assert.Equal(t, 1, out.KittySize)
	})

	// **80 点は 200 点の 4 割。**この 2 つが読めないと守備側の目標が伝わらない。
	t.Run("the point targets are published", func(t *testing.T) {
		out := parseShengJiOutput(t, p.Output(setupShengJiMock(defaultShengJiOpts()), nil))
		assert.Equal(t, domain.ShengJiTotalPoints, out.TotalPoints)
		assert.Equal(t, domain.ShengJiDefenderTarget, out.DefenderTarget)
		assert.Equal(t, domain.ShengJiAdvanceStep, out.AdvanceStep)
		assert.Equal(t, domain.ShengJiMinLevel, out.MinLevel)
		assert.Equal(t, domain.ShengJiMaxLevel, out.MaxLevel)
		assert.Equal(t, domain.ShengJiKittySize, out.KittySizeMax)
		// 守備側が集めた点も届くこと。
		assert.Equal(t, 35, out.TeamPoints[1])
	})

	t.Run("the declaration and the trump suit carry over", func(t *testing.T) {
		o := defaultShengJiOpts()
		o.declaration = &domain.ShengJiDeclaration{Seat: 2, Suit: domain.CardDesignHeart, Strength: 2}
		o.trumpSuit = domain.CardDesignHeart
		out := parseShengJiOutput(t, p.Output(setupShengJiMock(o), nil))
		assert.NotNil(t, out.Declaration)
		assert.Equal(t, 2, out.Declaration.Seat)
		assert.Equal(t, domain.CardDesignHeart, out.Declaration.Suit)
		assert.Equal(t, 2, out.Declaration.Strength)
		assert.Equal(t, domain.CardDesignHeart, out.TrumpSuit)
	})

	// **無主も正当な状態。**
	t.Run("no trump is reported as zero", func(t *testing.T) {
		o := defaultShengJiOpts()
		o.trumpSuit = domain.ShengJiNoTrump
		out := parseShengJiOutput(t, p.Output(setupShengJiMock(o), nil))
		assert.Equal(t, domain.ShengJiNoTrump, out.TrumpSuit)
		assert.Nil(t, out.Declaration)
	})

	// **持っていないスートは宣言できない。**選択肢を出さないと操作できない。
	t.Run("declarable suits are offered only while declaring", func(t *testing.T) {
		o := defaultShengJiOpts()
		o.phase = domain.ShengJiPhaseDeclare
		o.declarable = map[int]int{domain.CardDesignHeart: 2, domain.CardDesignClover: 1}
		out := parseShengJiOutput(t, p.Output(setupShengJiMock(o), nil))
		assert.Equal(t, 2, out.DeclarableSuits["3"])
		assert.Equal(t, 1, out.DeclarableSuits["2"])
		assert.NotContains(t, out.DeclarableSuits, "1")

		o.phase = domain.ShengJiPhasePlay
		out = parseShengJiOutput(t, p.Output(setupShengJiMock(o), nil))
		assert.Empty(t, out.DeclarableSuits)
	})

	// トリックはリード順なので、席は先頭からの距離で決まる。
	t.Run("the trick carries the seat that played each hand", func(t *testing.T) {
		o := defaultShengJiOpts()
		o.trick = [][]*domain.Card{
			{sjiTestCard(domain.CardDesignHeart, 7)},
			{sjiTestCard(domain.CardDesignHeart, 9)},
		}
		o.leadCombo = &domain.ShengJiCombo{Kind: domain.ShengJiComboSingle, Rank: 7, Size: 1}
		out := parseShengJiOutput(t, p.Output(setupShengJiMock(o), nil))
		assert.Len(t, out.Trick, 2)
		assert.Equal(t, 1, out.Trick[0].Seat)
		assert.Equal(t, 2, out.Trick[1].Seat)
		assert.NotNil(t, out.LeadCombo)
		assert.Equal(t, int(domain.ShengJiComboSingle), out.LeadCombo.Kind)
	})

	// **80 点で宣言側が交代する。**守りきった局とは別のメッセージにする。
	t.Run("the settled hand distinguishes held from taken", func(t *testing.T) {
		o := defaultShengJiOpts()
		o.phase = domain.ShengJiPhaseHandEnd
		o.lastResult = &domain.ShengJiHandResult{
			DeclarerTeam: 0, DefenderPoints: 35, DeclarerHeld: true, Advance: 2, AdvancingTeam: 0,
		}
		out := parseShengJiOutput(t, p.Output(setupShengJiMock(o), nil))
		assert.NotNil(t, out.LastResult)
		assert.True(t, out.LastResult.DeclarerHeld)
		assert.Equal(t, "shengji.handHeld", out.MessageCode)

		o.lastResult = &domain.ShengJiHandResult{
			DeclarerTeam: 0, DefenderPoints: 120, KittyPoints: 40, KittyMultiplier: 4,
			DeclarerHeld: false, Advance: 1, AdvancingTeam: 1,
		}
		out = parseShengJiOutput(t, p.Output(setupShengJiMock(o), nil))
		assert.False(t, out.LastResult.DeclarerHeld)
		assert.Equal(t, 40, out.LastResult.KittyPoints)
		assert.Equal(t, 4, out.LastResult.KittyMultiplier)
		assert.Equal(t, "shengji.handTaken", out.MessageCode)
	})

	t.Run("each phase has its own code", func(t *testing.T) {
		for phase, want := range map[domain.ShengJiPhase]string{
			domain.ShengJiPhaseDeclare: "shengji.declarePhase",
			domain.ShengJiPhaseKitty:   "shengji.kittyPhase",
			domain.ShengJiPhasePlay:    "shengji.playPhase",
		} {
			o := defaultShengJiOpts()
			o.phase = phase
			out := parseShengJiOutput(t, p.Output(setupShengJiMock(o), nil))
			assert.Equal(t, want, out.MessageCode)
		}
	})

	// **勝敗は席ではなくチームで見る。**人間は席 0 = チーム 0。
	t.Run("the winning team decides the closing message", func(t *testing.T) {
		o := defaultShengJiOpts()
		o.gameEnd = true
		o.winner = 0
		out := parseShengJiOutput(t, p.Output(setupShengJiMock(o), nil))
		assert.Equal(t, "shengji.result.humanWin", out.MessageCode)

		o.winner = 1
		out = parseShengJiOutput(t, p.Output(setupShengJiMock(o), nil))
		assert.Equal(t, "shengji.result.cpuWin", out.MessageCode)
	})

	t.Run("an error is surfaced verbatim", func(t *testing.T) {
		out := parseShengJiOutput(t, p.Output(setupShengJiMock(defaultShengJiOpts()), errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
		assert.Empty(t, out.MessageCode)
	})
}

func TestShengJiWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ShengJiWebPresenter)
	assert.JSONEq(t, `{"entries":[]}`, p.ActionLogOutput(setupShengJiMock(defaultShengJiOpts())))
}
