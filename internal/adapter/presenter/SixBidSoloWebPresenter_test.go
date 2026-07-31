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

func sbsTestCard(suit, value int) *domain.Card { return domain.NewCard(suit, value, true) }

// makeSixBidSoloPlayers builds three seats, the first human, with the given hands.
func makeSixBidSoloPlayers(hands ...[]*domain.Card) []*domain.SixBidSoloPlayer {
	out := make([]*domain.SixBidSoloPlayer, 0, len(hands))
	for i, hand := range hands {
		p := domain.NewSixBidSoloPlayer(i == 0)
		for _, c := range hand {
			p.AddCard(c)
		}
		out = append(out, p)
	}
	return out
}

// sixBidSoloMockOpts tunes the parts of the stub that individual tests vary.
type sixBidSoloMockOpts struct {
	phase      domain.SixBidSoloPhase
	declarer   int
	trumpSuit  int
	declared   bool
	spreadOpen bool
	highBid    *domain.SixBidSoloBid
	called     *domain.Card
	result     *domain.SixBidSoloHandResult
	gameEnd    bool
	winner     int
}

func setupSixBidSoloWebMock(o sixBidSoloMockOpts) *interfaces.MockSixBidSoloGame {
	m := new(interfaces.MockSixBidSoloGame)
	players := makeSixBidSoloPlayers(
		[]*domain.Card{sbsTestCard(domain.CardDesignSpade, 1)},
		[]*domain.Card{sbsTestCard(domain.CardDesignHeart, 10)},
		[]*domain.Card{sbsTestCard(domain.CardDesignClover, 13)},
	)
	m.On("GetPhase").Return(o.phase)
	m.On("GetHandNumber").Return(2)
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
		nil, // nil 混入でも落ちないこと
	})
	m.On("SixBidSoloWidowPoints").Return(21)
	m.On("GetTrick").Return([]*domain.Card{sbsTestCard(domain.CardDesignSpade, 13), nil})
	m.On("GetTrickLeaderIdx").Return(0)
	m.On("GetTrickNumber").Return(3)
	m.On("GetGameEndFlag").Return(o.gameEnd)
	m.On("GetWinnerIdx").Return(o.winner)
	m.On("GetConfig").Return(domain.DefaultSixBidSoloConfig())
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(true)
	m.On("SixBidSoloValidPlays", 0).Return([]int{0})
	m.On("GetBids").Return([]*domain.SixBidSoloBid{
		{Player: 1, Kind: domain.SixBidSoloBidGuarantee},
		{Player: 2, Kind: domain.SixBidSoloBidPass},
		nil, // nil 混入でも落ちないこと
	})
	m.On("GetHighBid").Return(o.highBid)
	m.On("GetLastResult").Return(o.result)
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
		m.On("GetPoints", i).Return(10 * i)
		m.On("GetTricksWon", i).Return(i)
		m.On("GetScore", i).Return(100 * i)
	}
	return m
}

func defaultSixBidSoloOpts() sixBidSoloMockOpts {
	return sixBidSoloMockOpts{
		phase:     domain.SixBidSoloPhasePlay,
		declarer:  1,
		trumpSuit: domain.CardDesignSpade,
		declared:  true,
		highBid:   &domain.SixBidSoloBid{Player: 1, Kind: domain.SixBidSoloBidGuarantee},
		winner:    -1,
	}
}

func parseSixBidSoloOutput(t *testing.T, s string) *controller.SixBidSoloWebOutput {
	t.Helper()
	var out controller.SixBidSoloWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return &out
}

// **ウィドウは精算まで伏せたまま。**枚数だけ送る。
func TestSixBidSoloWebPresenter_KeepsTheWidowFaceDownUntilTheSettlement(t *testing.T) {
	out := parseSixBidSoloOutput(t, new(presenter.SixBidSoloWebPresenter).Output(setupSixBidSoloWebMock(defaultSixBidSoloOpts()), nil))
	assert.Empty(t, out.Widow, "the widow must stay concealed during play")
	assert.Equal(t, 3, out.WidowSize, "its size is still sent so the table can show three face-down cards")

	o := defaultSixBidSoloOpts()
	o.phase = domain.SixBidSoloPhaseHandEnd
	revealed := parseSixBidSoloOutput(t, new(presenter.SixBidSoloWebPresenter).Output(setupSixBidSoloWebMock(o), nil))
	// nil 混入なので 2 枚。
	assert.Len(t, revealed.Widow, 2, "the widow is revealed once the hand settles")
}

// 他家の手札はプレイ中は見えない。
func TestSixBidSoloWebPresenter_HidesTheOtherHands(t *testing.T) {
	out := parseSixBidSoloOutput(t, new(presenter.SixBidSoloWebPresenter).Output(setupSixBidSoloWebMock(defaultSixBidSoloOpts()), nil))
	assert.Len(t, out.Players, 3)
	assert.Len(t, out.Players[0].Cards, 1, "the human sees its own hand")
	for i := 1; i < 3; i++ {
		assert.Empty(t, out.Players[i].Cards, "seat %d must stay hidden", i)
		assert.Positive(t, out.Players[i].CardCount)
	}
}

// **スプレッド・ミゼールでは宣言者の手札が公開される。**それが賭けの中身。
func TestSixBidSoloWebPresenter_ExposesTheSpreadMisereHand(t *testing.T) {
	o := defaultSixBidSoloOpts()
	o.declarer = 1
	o.spreadOpen = true
	out := parseSixBidSoloOutput(t, new(presenter.SixBidSoloWebPresenter).Output(setupSixBidSoloWebMock(o), nil))
	assert.NotEmpty(t, out.Players[1].Cards, "the declarer's hand is laid down")
	assert.Empty(t, out.Players[2].Cards, "the other opponent stays concealed")
	assert.True(t, out.SpreadOpen)
}

// **目標点はビッドとスートの両方で決まる。**表をそのまま送る。
func TestSixBidSoloWebPresenter_SendsTheTargetTable(t *testing.T) {
	o := defaultSixBidSoloOpts()
	o.trumpSuit = domain.CardDesignSpade
	spades := parseSixBidSoloOutput(t, new(presenter.SixBidSoloWebPresenter).Output(setupSixBidSoloWebMock(o), nil))

	// 通常ビッドは 61 点。**60 ちょうどでは足りない。**
	assert.Equal(t, 61, spades.BidTargets[domain.SixBidSoloBidSolo])
	assert.Equal(t, 61, spades.BidTargets[domain.SixBidSoloBidHeartSolo])
	assert.Equal(t, 0, spades.BidTargets[domain.SixBidSoloBidMisere])
	assert.Equal(t, 80, spades.BidTargets[domain.SixBidSoloBidGuarantee], "guarantee at a black suit needs 80")
	assert.Equal(t, domain.SixBidSoloTotalPoints, spades.BidTargets[domain.SixBidSoloBidCall])

	// **♥ ならギャランティーは 74。**
	o.trumpSuit = domain.CardDesignHeart
	hearts := parseSixBidSoloOutput(t, new(presenter.SixBidSoloWebPresenter).Output(setupSixBidSoloWebMock(o), nil))
	assert.Equal(t, 74, hearts.BidTargets[domain.SixBidSoloBidGuarantee], "guarantee at hearts needs 74")

	assert.Equal(t, 120, spades.TotalPoints)
	assert.Equal(t, 60, spades.BaseTarget)
	assert.Equal(t, domain.SixBidSoloHandSize, spades.HandSize)
	assert.Equal(t, 11, spades.HandSize, "eleven each, not twelve")
}

// **未達でも精算が読める。**ウィドウ加算の内訳も送る。
func TestSixBidSoloWebPresenter_SendsTheSettlement(t *testing.T) {
	o := defaultSixBidSoloOpts()
	o.phase = domain.SixBidSoloPhaseHandEnd
	o.result = &domain.SixBidSoloHandResult{
		Kind:           domain.SixBidSoloBidSolo,
		Declarer:       1,
		DeclarerPoints: 65,
		WidowPoints:    25,
		Target:         61,
		Made:           true,
		Value:          10,
		Deltas:         [domain.SixBidSoloPlayerCnt]int{-10, 20, -10},
	}
	out := parseSixBidSoloOutput(t, new(presenter.SixBidSoloWebPresenter).Output(setupSixBidSoloWebMock(o), nil))

	assert.NotNil(t, out.LastResult)
	assert.Equal(t, 65, out.LastResult.DeclarerPoints)
	assert.Equal(t, 25, out.LastResult.WidowPoints, "the widow is credited to the declarer")
	assert.Equal(t, 61, out.LastResult.Target)
	assert.True(t, out.LastResult.Made)
	assert.Equal(t, 10, out.LastResult.Value)
	assert.Equal(t, [domain.SixBidSoloPlayerCnt]int{-10, 20, -10}, out.LastResult.Deltas)
	assert.Equal(t, "sixbidsolo.handMade", out.MessageCode)
}

// **出せる札はサーバーが決める。**追随が強制。
func TestSixBidSoloWebPresenter_SendsValidPlaysOnlyOnTheHumansPlayTurn(t *testing.T) {
	out := parseSixBidSoloOutput(t, new(presenter.SixBidSoloWebPresenter).Output(setupSixBidSoloWebMock(defaultSixBidSoloOpts()), nil))
	assert.Equal(t, []int{0}, out.ValidPlays)

	o := defaultSixBidSoloOpts()
	o.phase = domain.SixBidSoloPhaseBid
	bidding := setupSixBidSoloWebMock(o)
	bidOut := parseSixBidSoloOutput(t, new(presenter.SixBidSoloWebPresenter).Output(bidding, nil))
	assert.Empty(t, bidOut.ValidPlays)
	bidding.AssertNotCalled(t, "SixBidSoloValidPlays", 0)
}

func TestSixBidSoloWebPresenter_TopLevelFields(t *testing.T) {
	o := defaultSixBidSoloOpts()
	o.called = sbsTestCard(domain.CardDesignHeart, 1)
	out := parseSixBidSoloOutput(t, new(presenter.SixBidSoloWebPresenter).Output(setupSixBidSoloWebMock(o), nil))

	assert.Equal(t, int(domain.SixBidSoloPhasePlay), out.Phase)
	assert.Equal(t, 2, out.HandNumber)
	assert.Equal(t, 1, out.DeclarerIdx)
	assert.Equal(t, domain.CardDesignSpade, out.TrumpSuit)
	assert.True(t, out.Declared)
	// nil をまぜたトリックでも落ちない。
	assert.Len(t, out.Trick, 1)
	assert.Equal(t, 3, out.TrickNumber)
	// nil をまぜた宣言履歴でも落ちない。
	assert.Len(t, out.Bids, 2)
	assert.Equal(t, int(domain.SixBidSoloBidGuarantee), out.Bids[0].Kind)
	// **パスも 0 として送る。**
	assert.Equal(t, int(domain.SixBidSoloBidPass), out.Bids[1].Kind)
	assert.NotNil(t, out.CalledCard)
	assert.Equal(t, domain.SixBidSoloDefaultHands, out.TargetHands)
	assert.True(t, out.Players[2].IsDealer)
	assert.True(t, out.Players[1].IsDeclarer)
	assert.True(t, out.Players[0].IsCurrentTurn)
	assert.Equal(t, 10, out.Players[1].Points)
	assert.Equal(t, 200, out.Players[2].Score)
}

func TestSixBidSoloWebPresenter_Messages(t *testing.T) {
	for _, tc := range []struct {
		phase   domain.SixBidSoloPhase
		result  *domain.SixBidSoloHandResult
		wantKey string
	}{
		{domain.SixBidSoloPhaseBid, nil, "sixbidsolo.bidPhase"},
		{domain.SixBidSoloPhaseDeclare, nil, "sixbidsolo.declarePhase"},
		{domain.SixBidSoloPhasePlay, nil, "sixbidsolo.playPhase"},
		{domain.SixBidSoloPhaseHandEnd, &domain.SixBidSoloHandResult{Made: true}, "sixbidsolo.handMade"},
		{domain.SixBidSoloPhaseHandEnd, &domain.SixBidSoloHandResult{Made: false}, "sixbidsolo.handSet"},
		// 結果が無い局面でも落ちない。
		{domain.SixBidSoloPhaseHandEnd, nil, "sixbidsolo.handMade"},
	} {
		o := defaultSixBidSoloOpts()
		o.phase = tc.phase
		o.result = tc.result
		out := parseSixBidSoloOutput(t, new(presenter.SixBidSoloWebPresenter).Output(setupSixBidSoloWebMock(o), nil))
		assert.Equal(t, tc.wantKey, out.MessageCode)
	}

	t.Run("an error wins over any phase message", func(t *testing.T) {
		out := parseSixBidSoloOutput(t, new(presenter.SixBidSoloWebPresenter).Output(setupSixBidSoloWebMock(defaultSixBidSoloOpts()), errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
		assert.Empty(t, out.MessageCode)
	})
}

func TestSixBidSoloWebPresenter_GameEnd(t *testing.T) {
	for _, tc := range []struct {
		name    string
		seat    int
		wantKey string
	}{
		{"the human wins", 0, "sixbidsolo.result.humanWin"},
		{"a cpu wins", 1, "sixbidsolo.result.cpuWin"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := defaultSixBidSoloOpts()
			o.phase = domain.SixBidSoloPhaseGameEnd
			o.gameEnd = true
			o.winner = tc.seat
			out := parseSixBidSoloOutput(t, new(presenter.SixBidSoloWebPresenter).Output(setupSixBidSoloWebMock(o), nil))
			assert.Equal(t, tc.wantKey, out.MessageCode)
			assert.True(t, out.GameEndFlag)
			assert.Equal(t, tc.seat, out.WinnerIdx)
		})
	}
}

// **落札前に呼ばれても落ちない。**declarerIdx は -1 のまま。
func TestSixBidSoloWebPresenter_NoContractYet(t *testing.T) {
	o := defaultSixBidSoloOpts()
	o.phase = domain.SixBidSoloPhaseBid
	o.declarer = -1
	o.highBid = nil
	o.declared = false
	o.trumpSuit = 0
	out := parseSixBidSoloOutput(t, new(presenter.SixBidSoloWebPresenter).Output(setupSixBidSoloWebMock(o), nil))
	assert.Nil(t, out.HighBid)
	assert.Equal(t, -1, out.DeclarerIdx)
	assert.Nil(t, out.LastResult)
	assert.Nil(t, out.CalledCard)
	assert.False(t, out.Declared)
	// 目標表は落札前から送る。選択肢を出すのに要る。
	assert.Equal(t, 61, out.BidTargets[domain.SixBidSoloBidSolo])
}

func TestSixBidSoloWebPresenter_ActionLogOutput(t *testing.T) {
	m := setupSixBidSoloWebMock(defaultSixBidSoloOpts())
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotEmpty(t, new(presenter.SixBidSoloWebPresenter).ActionLogOutput(m))
}
