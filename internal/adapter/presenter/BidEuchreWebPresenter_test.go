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

func beTestCard(suit, value int) *domain.Card { return domain.NewCard(suit, value, true) }

// makeBidEuchrePlayers builds four seats, the first human, with the given hands.
func makeBidEuchrePlayers(hands ...[]*domain.Card) []*domain.BidEuchrePlayer {
	out := make([]*domain.BidEuchrePlayer, 0, len(hands))
	for i, hand := range hands {
		p := domain.NewBidEuchrePlayer(i == 0)
		for _, c := range hand {
			p.AddCard(c)
		}
		out = append(out, p)
	}
	return out
}

// bidEuchreMockOpts tunes the parts of the stub that individual tests vary.
type bidEuchreMockOpts struct {
	phase       domain.BidEuchrePhase
	declarer    int
	trump       domain.BidEuchreTrump
	trumpChosen bool
	highBid     *domain.BidEuchreBid
	result      *domain.BidEuchreHandResult
	gameEnd     bool
	winner      int
}

func setupBidEuchreWebMock(o bidEuchreMockOpts) *interfaces.MockBidEuchreGame {
	m := new(interfaces.MockBidEuchreGame)
	players := makeBidEuchrePlayers(
		[]*domain.Card{beTestCard(domain.CardDesignSpade, 1)},
		[]*domain.Card{beTestCard(domain.CardDesignHeart, 11)},
		[]*domain.Card{beTestCard(domain.CardDesignClover, 11)},
		[]*domain.Card{beTestCard(domain.CardDesignDiamond, 9)},
	)
	m.On("GetPhase").Return(o.phase)
	m.On("GetHandNumber").Return(2)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(3)
	m.On("GetDeclarerIdx").Return(o.declarer)
	m.On("GetTrump").Return(o.trump)
	m.On("GetTrumpSuit").Return(domain.BidEuchreTrumpSuit(o.trump))
	m.On("IsTrumpChosen").Return(o.trumpChosen)
	m.On("GetTrick").Return([]*domain.Card{beTestCard(domain.CardDesignSpade, 13), nil})
	m.On("GetTrickLeaderIdx").Return(0)
	m.On("GetTrickNumber").Return(3)
	m.On("GetGameEndFlag").Return(o.gameEnd)
	m.On("GetWinnerTeam").Return(o.winner)
	m.On("GetConfig").Return(domain.DefaultBidEuchreConfig())
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(true)
	m.On("BidEuchreValidPlays", 0).Return([]int{0})
	m.On("GetBids").Return([]*domain.BidEuchreBid{
		{Player: 1, Value: 4},
		{Player: 2, Value: 0},
		nil, // nil 混入でも落ちないこと
	})
	m.On("GetHighBid").Return(o.highBid)
	m.On("GetLastResult").Return(o.result)
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
		m.On("GetTricksWon", i).Return(i)
	}
	for team := range domain.BidEuchreTeamCnt {
		m.On("BidEuchreTeamTricks", team).Return(3 + team)
		m.On("GetScore", team).Return(10 * (team + 1))
	}
	return m
}

func defaultBidEuchreOpts() bidEuchreMockOpts {
	return bidEuchreMockOpts{
		phase:       domain.BidEuchrePhasePlay,
		declarer:    1,
		trump:       domain.BidEuchreTrumpSpade,
		trumpChosen: true,
		highBid:     &domain.BidEuchreBid{Player: 1, Value: 4},
		winner:      -1,
	}
}

func parseBidEuchreOutput(t *testing.T, s string) *controller.BidEuchreWebOutput {
	t.Helper()
	var out controller.BidEuchreWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return &out
}

// **キティが無く、誰の手札も公開されない。**
func TestBidEuchreWebPresenter_HidesEveryOtherHand(t *testing.T) {
	m := setupBidEuchreWebMock(defaultBidEuchreOpts())
	out := parseBidEuchreOutput(t, new(presenter.BidEuchreWebPresenter).Output(m, nil))

	assert.Len(t, out.Players, 4)
	assert.Len(t, out.Players[0].Cards, 1, "the human sees its own hand")
	for i := 1; i < 4; i++ {
		assert.Empty(t, out.Players[i].Cards, "seat %d must stay hidden", i)
		assert.Positive(t, out.Players[i].CardCount)
	}
	assert.Empty(t, out.Players[2].Cards, "even the partner's hand stays concealed")
}

func TestBidEuchreWebPresenter_RevealsAtTheSettlement(t *testing.T) {
	o := defaultBidEuchreOpts()
	o.phase = domain.BidEuchrePhaseHandEnd
	out := parseBidEuchreOutput(t, new(presenter.BidEuchreWebPresenter).Output(setupBidEuchreWebMock(o), nil))
	for i := range out.Players {
		assert.NotEmpty(t, out.Players[i].Cards, "seat %d must be revealed", i)
	}
}

// **24 枚を 4 人にちょうど配り切る。**キティが残る余地は無い。
func TestBidEuchreWebPresenter_SendsTheDealShape(t *testing.T) {
	out := parseBidEuchreOutput(t, new(presenter.BidEuchreWebPresenter).Output(setupBidEuchreWebMock(defaultBidEuchreOpts()), nil))
	assert.Equal(t, domain.BidEuchreHandSize, out.HandSize)
	assert.Equal(t, 6, out.HandSize)
	assert.Equal(t, domain.BidEuchreMinBid, out.MinBid)
	assert.Equal(t, 3, out.MinBid, "three tricks is the floor")
	assert.Equal(t, domain.BidEuchreMaxBid, out.MaxBid)
	assert.Equal(t, domain.BidEuchreGameTarget, out.GameTarget)
	assert.Equal(t, 32, out.GameTarget)
}

// **ノートランプは 2 種類あり、宣言そのものを送る必要がある。**
// TrumpSuit だけでは NT ハイと NT ローが区別できない。
func TestBidEuchreWebPresenter_DistinguishesTheTwoNoTrumpForms(t *testing.T) {
	for _, tc := range []struct {
		name  string
		trump domain.BidEuchreTrump
	}{
		{"no trump high", domain.BidEuchreTrumpNoHigh},
		{"no trump low", domain.BidEuchreTrumpNoLow},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := defaultBidEuchreOpts()
			o.trump = tc.trump
			out := parseBidEuchreOutput(t, new(presenter.BidEuchreWebPresenter).Output(setupBidEuchreWebMock(o), nil))
			assert.Equal(t, int(tc.trump), out.Trump)
			assert.Equal(t, 0, out.TrumpSuit, "a no-trump declaration has no trump suit")
			assert.True(t, out.TrumpChosen)
		})
	}

	// スート宣言では両方が埋まる。
	out := parseBidEuchreOutput(t, new(presenter.BidEuchreWebPresenter).Output(setupBidEuchreWebMock(defaultBidEuchreOpts()), nil))
	assert.Equal(t, int(domain.BidEuchreTrumpSpade), out.Trump)
	assert.Equal(t, domain.CardDesignSpade, out.TrumpSuit)
}

// **未達側は宣言額を失い、守備側は取ったトリックを得点する。**
func TestBidEuchreWebPresenter_SendsBothSidesPoints(t *testing.T) {
	o := defaultBidEuchreOpts()
	o.phase = domain.BidEuchrePhaseHandEnd
	o.result = &domain.BidEuchreHandResult{
		Points: [domain.BidEuchreTeamCnt]int{-5, 4},
		Tricks: [domain.BidEuchreTeamCnt]int{2, 4},
		Made:   false,
		Bid:    5,
	}
	out := parseBidEuchreOutput(t, new(presenter.BidEuchreWebPresenter).Output(setupBidEuchreWebMock(o), nil))

	assert.NotNil(t, out.LastResult)
	// **-5 であって -2 ではない。**引かれるのは宣言額。
	assert.Equal(t, -5, out.LastResult.Points[0])
	// **守備側は未達でも自分のトリックを得点する。**
	assert.Equal(t, 4, out.LastResult.Points[1])
	assert.Equal(t, [domain.BidEuchreTeamCnt]int{2, 4}, out.LastResult.Tricks)
	assert.False(t, out.LastResult.Made)
	assert.Equal(t, 5, out.LastResult.Bid)
	assert.Equal(t, "bideuchre.handSet", out.MessageCode)
}

// **出せる札はサーバーが決める。**左ボワーが切札扱いになるため。
func TestBidEuchreWebPresenter_SendsValidPlaysOnlyOnTheHumansPlayTurn(t *testing.T) {
	out := parseBidEuchreOutput(t, new(presenter.BidEuchreWebPresenter).Output(setupBidEuchreWebMock(defaultBidEuchreOpts()), nil))
	assert.Equal(t, []int{0}, out.ValidPlays)

	o := defaultBidEuchreOpts()
	o.phase = domain.BidEuchrePhaseBid
	bidding := setupBidEuchreWebMock(o)
	bidOut := parseBidEuchreOutput(t, new(presenter.BidEuchreWebPresenter).Output(bidding, nil))
	assert.Empty(t, bidOut.ValidPlays)
	bidding.AssertNotCalled(t, "BidEuchreValidPlays", 0)
}

func TestBidEuchreWebPresenter_TopLevelFields(t *testing.T) {
	out := parseBidEuchreOutput(t, new(presenter.BidEuchreWebPresenter).Output(setupBidEuchreWebMock(defaultBidEuchreOpts()), nil))

	assert.Equal(t, int(domain.BidEuchrePhasePlay), out.Phase)
	assert.Equal(t, 2, out.HandNumber)
	assert.Equal(t, 1, out.DeclarerIdx)
	// nil をまぜたトリックでも落ちない。
	assert.Len(t, out.Trick, 1)
	assert.Equal(t, 3, out.TrickNumber)
	// nil をまぜた宣言履歴でも落ちない。
	assert.Len(t, out.Bids, 2)
	assert.Equal(t, 4, out.Bids[0].Value)
	// **パスも 0 として送る。**
	assert.Equal(t, 0, out.Bids[1].Value)
	assert.Equal(t, [domain.BidEuchreTeamCnt]int{3, 4}, out.TeamTricks)
	assert.Equal(t, [domain.BidEuchreTeamCnt]int{10, 20}, out.Scores)
	// パートナーは向かい合わせ。
	assert.Equal(t, out.Players[0].Team, out.Players[2].Team)
	assert.NotEqual(t, out.Players[0].Team, out.Players[1].Team)
	assert.True(t, out.Players[3].IsDealer)
	assert.True(t, out.Players[1].IsDeclarer)
	assert.True(t, out.Players[0].IsCurrentTurn)
	assert.True(t, out.Config.AllowNoTrump)
}

func TestBidEuchreWebPresenter_Messages(t *testing.T) {
	for _, tc := range []struct {
		phase   domain.BidEuchrePhase
		result  *domain.BidEuchreHandResult
		wantKey string
	}{
		{domain.BidEuchrePhaseBid, nil, "bideuchre.bidPhase"},
		{domain.BidEuchrePhaseChooseTrump, nil, "bideuchre.trumpPhase"},
		{domain.BidEuchrePhasePlay, nil, "bideuchre.playPhase"},
		{domain.BidEuchrePhaseHandEnd, &domain.BidEuchreHandResult{Made: true}, "bideuchre.handMade"},
		{domain.BidEuchrePhaseHandEnd, &domain.BidEuchreHandResult{Made: false}, "bideuchre.handSet"},
		// 結果が無い局面でも落ちない。
		{domain.BidEuchrePhaseHandEnd, nil, "bideuchre.handMade"},
	} {
		o := defaultBidEuchreOpts()
		o.phase = tc.phase
		o.result = tc.result
		out := parseBidEuchreOutput(t, new(presenter.BidEuchreWebPresenter).Output(setupBidEuchreWebMock(o), nil))
		assert.Equal(t, tc.wantKey, out.MessageCode)
	}

	t.Run("an error wins over any phase message", func(t *testing.T) {
		out := parseBidEuchreOutput(t, new(presenter.BidEuchreWebPresenter).Output(setupBidEuchreWebMock(defaultBidEuchreOpts()), errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
		assert.Empty(t, out.MessageCode)
	})
}

// **勝敗はチームで決まる。**人間は席 0 = チーム 0。
func TestBidEuchreWebPresenter_GameEndIsByTeam(t *testing.T) {
	for _, tc := range []struct {
		name    string
		team    int
		wantKey string
	}{
		{"the human's team wins", 0, "bideuchre.result.humanWin"},
		{"the other team wins", 1, "bideuchre.result.cpuWin"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := defaultBidEuchreOpts()
			o.phase = domain.BidEuchrePhaseGameEnd
			o.gameEnd = true
			o.winner = tc.team
			out := parseBidEuchreOutput(t, new(presenter.BidEuchreWebPresenter).Output(setupBidEuchreWebMock(o), nil))
			assert.Equal(t, tc.wantKey, out.MessageCode)
			assert.True(t, out.GameEndFlag)
			assert.Equal(t, tc.team, out.WinnerTeam)
		})
	}
}

// **落札前に呼ばれても落ちない。**declarerIdx は -1 のまま。
func TestBidEuchreWebPresenter_NoContractYet(t *testing.T) {
	o := defaultBidEuchreOpts()
	o.phase = domain.BidEuchrePhaseBid
	o.declarer = -1
	o.highBid = nil
	o.trumpChosen = false
	out := parseBidEuchreOutput(t, new(presenter.BidEuchreWebPresenter).Output(setupBidEuchreWebMock(o), nil))
	assert.Nil(t, out.HighBid)
	assert.Equal(t, -1, out.DeclarerIdx)
	assert.Nil(t, out.LastResult)
	assert.False(t, out.TrumpChosen)
	// 宣言の範囲は落札前から送る。選択肢を出すのに要る。
	assert.Equal(t, 3, out.MinBid)
}

func TestBidEuchreWebPresenter_ActionLogOutput(t *testing.T) {
	m := setupBidEuchreWebMock(defaultBidEuchreOpts())
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotEmpty(t, new(presenter.BidEuchreWebPresenter).ActionLogOutput(m))
}
