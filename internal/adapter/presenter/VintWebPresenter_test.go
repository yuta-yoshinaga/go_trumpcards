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

func vtTestCard(suit, value int) *domain.Card { return domain.NewCard(suit, value, true) }

// makeVintPlayers builds four seats, the first human, with the given hands.
func makeVintPlayers(hands ...[]*domain.Card) []*domain.VintPlayer {
	out := make([]*domain.VintPlayer, 0, len(hands))
	for i, hand := range hands {
		p := domain.NewVintPlayer(i == 0)
		for _, c := range hand {
			p.AddCard(c)
		}
		out = append(out, p)
	}
	return out
}

// vintMockOpts tunes the parts of the stub that individual tests vary.
type vintMockOpts struct {
	phase    domain.VintPhase
	declarer int
	highBid  *domain.VintBid
	result   *domain.VintHandResult
	gameEnd  bool
	winner   int
}

func setupVintWebMock(o vintMockOpts) *interfaces.MockVintGame {
	m := new(interfaces.MockVintGame)
	players := makeVintPlayers(
		[]*domain.Card{vtTestCard(domain.CardDesignSpade, 1)},
		[]*domain.Card{vtTestCard(domain.CardDesignHeart, 2)},
		[]*domain.Card{vtTestCard(domain.CardDesignClover, 3)},
		[]*domain.Card{vtTestCard(domain.CardDesignDiamond, 4)},
	)
	m.On("GetPhase").Return(o.phase)
	m.On("GetHandNumber").Return(2)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(3)
	m.On("GetDeclarerIdx").Return(o.declarer)
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetTrick").Return([]*domain.Card{vtTestCard(domain.CardDesignSpade, 5), nil})
	m.On("GetTrickLeaderIdx").Return(0)
	m.On("GetTrickNumber").Return(3)
	m.On("GetGameEndFlag").Return(o.gameEnd)
	m.On("GetWinnerTeam").Return(o.winner)
	m.On("GetConfig").Return(domain.DefaultVintConfig())
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(true)
	m.On("VintValidPlays", 0).Return([]int{0})
	m.On("GetBids").Return([]*domain.VintBid{
		{Player: 1, Level: 3, Denom: domain.VintDenomHeart},
		{Player: 2, Level: 0},
		nil, // nil 混入でも落ちないこと
	})
	m.On("GetHighBid").Return(o.highBid)
	m.On("GetLastResult").Return(o.result)
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
		m.On("GetTricksWon", i).Return(i)
	}
	for team := range domain.VintTeamCnt {
		m.On("VintTeamTricks", team).Return(6 + team)
		m.On("GetBelow", team).Return(100 * (team + 1))
		m.On("GetAbove", team).Return(50 * (team + 1))
		m.On("GetGamesWon", team).Return(team)
	}
	return m
}

func defaultVintOpts() vintMockOpts {
	return vintMockOpts{
		phase:    domain.VintPhasePlay,
		declarer: 1,
		highBid:  &domain.VintBid{Player: 1, Level: 3, Denom: domain.VintDenomHeart},
		winner:   -1,
	}
}

func parseVintOutput(t *testing.T, s string) *controller.VintWebOutput {
	t.Helper()
	var out controller.VintWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return &out
}

// **ダミーが無いので、プレイ中は誰の手札も見えない。**味方の手札も伏せる。
func TestVintWebPresenter_HasNoDummy(t *testing.T) {
	m := setupVintWebMock(defaultVintOpts())
	out := parseVintOutput(t, new(presenter.VintWebPresenter).Output(m, nil))

	assert.Len(t, out.Players, 4)
	assert.Len(t, out.Players[0].Cards, 1, "the human sees its own hand")
	for i := 1; i < 4; i++ {
		assert.Empty(t, out.Players[i].Cards, "seat %d must stay hidden — there is no dummy in Vint", i)
		assert.Positive(t, out.Players[i].CardCount)
	}
	// 味方 (席 2) も伏せられている。
	assert.Empty(t, out.Players[2].Cards, "even the partner's hand stays concealed")
}

func TestVintWebPresenter_RevealsAtTheSettlement(t *testing.T) {
	o := defaultVintOpts()
	o.phase = domain.VintPhaseHandEnd
	out := parseVintOutput(t, new(presenter.VintWebPresenter).Output(setupVintWebMock(o), nil))
	for i := range out.Players {
		assert.NotEmpty(t, out.Players[i].Cards, "seat %d must be revealed", i)
	}
}

// **単価はスートとレベルの両方で決まる。**基準表を送る。
func TestVintWebPresenter_SendsTheTrickValueTable(t *testing.T) {
	out := parseVintOutput(t, new(presenter.VintWebPresenter).Output(setupVintWebMock(defaultVintOpts()), nil))

	assert.Equal(t, 4, out.TrickValues[domain.VintDenomSpade])
	assert.Equal(t, 6, out.TrickValues[domain.VintDenomClub])
	assert.Equal(t, 8, out.TrickValues[domain.VintDenomDiamond])
	assert.Equal(t, 10, out.TrickValues[domain.VintDenomHeart])
	assert.Equal(t, 12, out.TrickValues[domain.VintDenomNoTrump])
	// **♠ が最弱。**ブリッジと逆であることが単価にも表れている。
	assert.Less(t, out.TrickValues[domain.VintDenomSpade], out.TrickValues[domain.VintDenomClub])

	// 立っている宣言の単価はレベルぶん上がっている (3♥ = 10 + 20 = 30)。
	assert.NotNil(t, out.HighBid)
	assert.Equal(t, 30, out.HighBid.TrickValue)
}

// **両チームが線下に得点する。**精算の内訳をそのまま送る。
func TestVintWebPresenter_SendsBothSidesTrickPoints(t *testing.T) {
	o := defaultVintOpts()
	o.phase = domain.VintPhaseHandEnd
	o.result = &domain.VintHandResult{
		TrickPoints:    [domain.VintTeamCnt]int{210, 180},
		HonourPoints:   [domain.VintTeamCnt]int{600, 0},
		AcePoints:      [domain.VintTeamCnt]int{1200, 0},
		Penalty:        [domain.VintTeamCnt]int{0, 1500},
		Made:           false,
		DeclarerTricks: 7,
		TrickValue:     30,
	}
	out := parseVintOutput(t, new(presenter.VintWebPresenter).Output(setupVintWebMock(o), nil))

	assert.NotNil(t, out.LastResult)
	// **守備側も線下に点が入る。**issue の「宣言側だけ」は誤り。
	assert.Equal(t, 210, out.LastResult.TrickPoints[0])
	assert.Equal(t, 180, out.LastResult.TrickPoints[1])
	assert.Equal(t, 600, out.LastResult.HonourPoints[0])
	assert.Equal(t, 1200, out.LastResult.AcePoints[0])
	assert.Equal(t, 1500, out.LastResult.Penalty[1])
	assert.False(t, out.LastResult.Made)
	assert.Equal(t, 7, out.LastResult.DeclarerTricks)
	assert.Equal(t, "vint.handSet", out.MessageCode)
}

// **出せる札はサーバーが決める。**追随が強制。
func TestVintWebPresenter_SendsValidPlaysOnlyOnTheHumansPlayTurn(t *testing.T) {
	out := parseVintOutput(t, new(presenter.VintWebPresenter).Output(setupVintWebMock(defaultVintOpts()), nil))
	assert.Equal(t, []int{0}, out.ValidPlays)

	o := defaultVintOpts()
	o.phase = domain.VintPhaseBid
	bidding := setupVintWebMock(o)
	bidOut := parseVintOutput(t, new(presenter.VintWebPresenter).Output(bidding, nil))
	assert.Empty(t, bidOut.ValidPlays)
	bidding.AssertNotCalled(t, "VintValidPlays", 0)
}

func TestVintWebPresenter_TopLevelFields(t *testing.T) {
	out := parseVintOutput(t, new(presenter.VintWebPresenter).Output(setupVintWebMock(defaultVintOpts()), nil))

	assert.Equal(t, int(domain.VintPhasePlay), out.Phase)
	assert.Equal(t, 2, out.HandNumber)
	assert.Equal(t, 1, out.DeclarerIdx)
	assert.Equal(t, domain.CardDesignHeart, out.TrumpSuit)
	// nil をまぜたトリックでも落ちない。
	assert.Len(t, out.Trick, 1)
	assert.Equal(t, 3, out.TrickNumber)
	// nil をまぜた宣言履歴でも落ちない。
	assert.Len(t, out.Bids, 2)
	assert.Equal(t, 3, out.Bids[0].Level)
	// **パスも 0 として送る。**
	assert.Equal(t, 0, out.Bids[1].Level)
	assert.Equal(t, [domain.VintTeamCnt]int{6, 7}, out.TeamTricks)
	assert.Equal(t, [domain.VintTeamCnt]int{100, 200}, out.Below)
	assert.Equal(t, [domain.VintTeamCnt]int{50, 100}, out.Above)
	assert.Equal(t, [domain.VintTeamCnt]int{0, 1}, out.GamesWon)
	assert.Equal(t, domain.VintGameTarget, out.GameTarget)
	assert.Equal(t, domain.VintMinLevel, out.MinLevel)
	assert.Equal(t, domain.VintMaxLevel, out.MaxLevel)
	// パートナーは向かい合わせ。
	assert.Equal(t, out.Players[0].Team, out.Players[2].Team)
	assert.NotEqual(t, out.Players[0].Team, out.Players[1].Team)
	assert.True(t, out.Players[3].IsDealer)
	assert.True(t, out.Players[1].IsDeclarer)
	assert.True(t, out.Players[0].IsCurrentTurn)
}

func TestVintWebPresenter_Messages(t *testing.T) {
	for _, tc := range []struct {
		phase   domain.VintPhase
		result  *domain.VintHandResult
		wantKey string
	}{
		{domain.VintPhaseBid, nil, "vint.bidPhase"},
		{domain.VintPhasePlay, nil, "vint.playPhase"},
		{domain.VintPhaseHandEnd, &domain.VintHandResult{Made: true}, "vint.handMade"},
		{domain.VintPhaseHandEnd, &domain.VintHandResult{Made: false}, "vint.handSet"},
		// 結果が無い局面でも落ちない。
		{domain.VintPhaseHandEnd, nil, "vint.handMade"},
	} {
		o := defaultVintOpts()
		o.phase = tc.phase
		o.result = tc.result
		out := parseVintOutput(t, new(presenter.VintWebPresenter).Output(setupVintWebMock(o), nil))
		assert.Equal(t, tc.wantKey, out.MessageCode)
	}

	t.Run("an error wins over any phase message", func(t *testing.T) {
		out := parseVintOutput(t, new(presenter.VintWebPresenter).Output(setupVintWebMock(defaultVintOpts()), errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
		assert.Empty(t, out.MessageCode)
	})
}

// **勝敗はチームで決まる。**人間は席 0 = チーム 0。
func TestVintWebPresenter_GameEndIsByTeam(t *testing.T) {
	for _, tc := range []struct {
		name    string
		team    int
		wantKey string
	}{
		{"the human's team takes the rubber", 0, "vint.result.humanWin"},
		{"the other team takes the rubber", 1, "vint.result.cpuWin"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := defaultVintOpts()
			o.phase = domain.VintPhaseGameEnd
			o.gameEnd = true
			o.winner = tc.team
			out := parseVintOutput(t, new(presenter.VintWebPresenter).Output(setupVintWebMock(o), nil))
			assert.Equal(t, tc.wantKey, out.MessageCode)
			assert.True(t, out.GameEndFlag)
			assert.Equal(t, tc.team, out.WinnerTeam)
		})
	}
}

func TestVintWebPresenter_NoContractYet(t *testing.T) {
	o := defaultVintOpts()
	o.phase = domain.VintPhaseBid
	o.declarer = -1
	o.highBid = nil
	out := parseVintOutput(t, new(presenter.VintWebPresenter).Output(setupVintWebMock(o), nil))
	assert.Nil(t, out.HighBid)
	assert.Equal(t, -1, out.DeclarerIdx)
	assert.Nil(t, out.LastResult)
	// 単価表は宣言前から送る。選択肢を出すのに要る。
	assert.Equal(t, 12, out.TrickValues[domain.VintDenomNoTrump])
}

func TestVintWebPresenter_ActionLogOutput(t *testing.T) {
	m := setupVintWebMock(defaultVintOpts())
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotEmpty(t, new(presenter.VintWebPresenter).ActionLogOutput(m))
}
