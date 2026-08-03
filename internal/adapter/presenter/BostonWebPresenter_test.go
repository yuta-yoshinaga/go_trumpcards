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

func bsTestCard(suit, value int) *domain.Card { return domain.NewCard(suit, value, true) }

// makeBostonPlayers builds four seats, the first human, with the given hands.
func makeBostonPlayers(hands ...[]*domain.Card) []*domain.BostonPlayer {
	out := make([]*domain.BostonPlayer, 0, len(hands))
	for i, hand := range hands {
		p := domain.NewBostonPlayer(i == 0)
		for _, c := range hand {
			p.AddCard(c)
		}
		out = append(out, p)
	}
	return out
}

// bostonMockOpts tunes the parts of the stub that individual tests vary.
type bostonMockOpts struct {
	phase       domain.BostonPhase
	declarer    int
	partner     int
	exposed     bool
	trickNumber int
	highBid     *domain.BostonBidRecord
	gameEnd     bool
	winner      int
	bidMade     bool
}

func setupBostonWebMock(o bostonMockOpts) *interfaces.MockBostonGame {
	m := new(interfaces.MockBostonGame)
	players := makeBostonPlayers(
		[]*domain.Card{bsTestCard(domain.CardDesignSpade, 1)},
		[]*domain.Card{bsTestCard(domain.CardDesignHeart, 2)},
		[]*domain.Card{bsTestCard(domain.CardDesignClover, 3)},
		[]*domain.Card{bsTestCard(domain.CardDesignDiamond, 4)},
	)
	m.On("GetPhase").Return(o.phase)
	m.On("GetHandNumber").Return(2)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(3)
	m.On("GetDeclarerIdx").Return(o.declarer)
	m.On("GetPartnerIdx").Return(o.partner)
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("IsExposed").Return(o.exposed)
	m.On("GetTrick").Return([]*domain.Card{bsTestCard(domain.CardDesignSpade, 5), nil})
	m.On("GetTrickLeaderIdx").Return(0)
	m.On("GetTrickNumber").Return(o.trickNumber)
	m.On("BostonDeclarerTricks").Return(4)
	m.On("IsBidMade").Return(o.bidMade)
	m.On("GetTargetHands").Return(domain.BostonTargetHandsDefault)
	m.On("GetGameEndFlag").Return(o.gameEnd)
	m.On("GetWinnerIdx").Return(o.winner)
	m.On("GetConfig").Return(domain.DefaultBostonConfig())
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(true)
	m.On("BostonValidPlays", 0).Return([]int{0})
	m.On("GetBids").Return([]*domain.BostonBidRecord{
		{Player: 1, Level: domain.BostonBidSeven, Suit: domain.CardDesignHeart},
		{Player: 2, Level: domain.BostonBidPass},
		nil, // nil 混入でも落ちないこと
	})
	m.On("GetHighBid").Return(o.highBid)
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
		m.On("GetTricksWon", i).Return(i)
		m.On("GetChips", i).Return(i * 10)
		m.On("BostonIsDeclarerSide", i).Return(i == o.declarer || (o.partner >= 0 && i == o.partner))
	}
	return m
}

func defaultBostonOpts() bostonMockOpts {
	return bostonMockOpts{
		phase:       domain.BostonPhasePlay,
		declarer:    1,
		partner:     -1,
		trickNumber: 3,
		highBid:     &domain.BostonBidRecord{Player: 1, Level: domain.BostonBidSeven, Suit: domain.CardDesignHeart},
		winner:      -1,
	}
}

func parseBostonOutput(t *testing.T, s string) *controller.BostonWebOutput {
	t.Helper()
	var out controller.BostonWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return &out
}

// **序列表はサーバーが送る。**ミゼールがトリック宣言の間に挟まるので、
// クライアント側で組み直すと必ずずれる。
func TestBostonWebPresenter_SendsTheWholeLadderInRankOrder(t *testing.T) {
	m := setupBostonWebMock(defaultBostonOpts())
	out := parseBostonOutput(t, new(presenter.BostonWebPresenter).Output(m, nil))

	if len(out.BidOptions) != int(domain.BostonBidLevelCount)-1 {
		t.Fatalf("the ladder holds %d options, want %d", len(out.BidOptions), int(domain.BostonBidLevelCount)-1)
	}
	// 段が昇順に並んでいる。
	for i := 1; i < len(out.BidOptions); i++ {
		assert.Greater(t, out.BidOptions[i].Level, out.BidOptions[i-1].Level)
	}

	byName := map[string]*controller.BostonWebOutputBidOption{}
	for _, o := range out.BidOptions {
		byName[o.Name] = o
	}
	// **リトル・ミゼールは 7 トリックより下。**
	assert.Less(t, byName["littleMisere"].Level, byName["seven"].Level)
	assert.Greater(t, byName["littleMisere"].Level, byName["six"].Level)
	// **グランド・ミゼールは 9 トリックより下。**
	assert.Less(t, byName["grandMisere"].Level, byName["nine"].Level)
	// **ピッコリッシモはちょうど 1 トリック。**
	assert.Equal(t, 1, byName["piccolissimo"].Tricks)
	assert.Equal(t, int(domain.BostonKindPiccolissimo), byName["piccolissimo"].Kind)
	assert.False(t, byName["piccolissimo"].NeedsTrump)
	// **公開はそれ自体が別の宣言。**
	assert.False(t, byName["littleMisere"].Exposed)
	assert.True(t, byName["littleMisereTable"].Exposed)
	// **パートナーを呼べるのはトリック宣言だけ。**
	assert.True(t, byName["seven"].CanCallPartner)
	assert.False(t, byName["grandMisere"].CanCallPartner)
	assert.False(t, byName["chelem"].CanCallPartner)
}

// **相手の手札は伏せる。**枚数だけ送る。
func TestBostonWebPresenter_HidesOpponentHandsDuringPlay(t *testing.T) {
	m := setupBostonWebMock(defaultBostonOpts())
	out := parseBostonOutput(t, new(presenter.BostonWebPresenter).Output(m, nil))

	assert.Len(t, out.Players, 4)
	assert.Len(t, out.Players[0].Cards, 1, "the human sees its own hand")
	for i := 1; i < 4; i++ {
		assert.Empty(t, out.Players[i].Cards, "seat %d must stay hidden", i)
		assert.Positive(t, out.Players[i].CardCount)
	}
}

// **公開宣言では第1トリックのあとに落札者の手札が見える。**宣言と同時ではない。
func TestBostonWebPresenter_ExposesTheDeclarerOnlyAfterTheFirstTrick(t *testing.T) {
	before := defaultBostonOpts()
	before.exposed = true
	before.trickNumber = 0
	out := parseBostonOutput(t, new(presenter.BostonWebPresenter).Output(setupBostonWebMock(before), nil))
	assert.Empty(t, out.Players[1].Cards, "the hand stays hidden until the first trick is done")

	after := defaultBostonOpts()
	after.exposed = true
	after.trickNumber = 1
	out2 := parseBostonOutput(t, new(presenter.BostonWebPresenter).Output(setupBostonWebMock(after), nil))
	assert.NotEmpty(t, out2.Players[1].Cards, "the declarer's hand is exposed once the first trick is done")
	// **公開されるのは落札者だけ。**他家は伏せたまま。
	assert.Empty(t, out2.Players[2].Cards)
}

func TestBostonWebPresenter_RevealsAtTheSettlement(t *testing.T) {
	o := defaultBostonOpts()
	o.phase = domain.BostonPhaseHandEnd
	out := parseBostonOutput(t, new(presenter.BostonWebPresenter).Output(setupBostonWebMock(o), nil))
	for i := range out.Players {
		assert.NotEmpty(t, out.Players[i].Cards, "seat %d must be revealed", i)
	}
}

// **パートナーは落札側。**2 対 2 か 1 対 3 かがここで読める。
func TestBostonWebPresenter_MarksTheDeclaringSide(t *testing.T) {
	solo := parseBostonOutput(t, new(presenter.BostonWebPresenter).Output(setupBostonWebMock(defaultBostonOpts()), nil))
	assert.True(t, solo.Players[1].IsDeclarer)
	assert.True(t, solo.Players[1].IsDeclarerSide)
	assert.Equal(t, -1, solo.PartnerIdx)
	for _, i := range []int{0, 2, 3} {
		assert.False(t, solo.Players[i].IsDeclarerSide, "seat %d defends", i)
	}

	o := defaultBostonOpts()
	o.partner = 3
	pair := parseBostonOutput(t, new(presenter.BostonWebPresenter).Output(setupBostonWebMock(o), nil))
	assert.True(t, pair.Players[3].IsPartner)
	assert.True(t, pair.Players[3].IsDeclarerSide)
	assert.False(t, pair.Players[0].IsDeclarerSide)
}

// **出せる札はサーバーが決める。**追随が強制なのでフロントで再現するとずれる。
func TestBostonWebPresenter_SendsValidPlaysOnlyOnTheHumansPlayTurn(t *testing.T) {
	out := parseBostonOutput(t, new(presenter.BostonWebPresenter).Output(setupBostonWebMock(defaultBostonOpts()), nil))
	assert.Equal(t, []int{0}, out.ValidPlays)

	o := defaultBostonOpts()
	o.phase = domain.BostonPhaseBid
	bidding := setupBostonWebMock(o)
	bidOut := parseBostonOutput(t, new(presenter.BostonWebPresenter).Output(bidding, nil))
	assert.Empty(t, bidOut.ValidPlays)
	bidding.AssertNotCalled(t, "BostonValidPlays", 0)
}

func TestBostonWebPresenter_TopLevelFields(t *testing.T) {
	out := parseBostonOutput(t, new(presenter.BostonWebPresenter).Output(setupBostonWebMock(defaultBostonOpts()), nil))

	assert.Equal(t, int(domain.BostonPhasePlay), out.Phase)
	assert.Equal(t, 2, out.HandNumber)
	assert.Equal(t, 1, out.DeclarerIdx)
	assert.Equal(t, domain.CardDesignHeart, out.TrumpSuit)
	// nil をまぜたトリックでも落ちない。
	assert.Len(t, out.Trick, 1)
	assert.Equal(t, 3, out.TrickNumber)
	assert.Equal(t, 4, out.DeclarerTricks)
	assert.Equal(t, domain.BostonHandSize, out.HandSize)
	assert.Equal(t, domain.BostonTargetHandsDefault, out.TargetHands)
	// nil をまぜた宣言履歴でも落ちない。
	assert.Len(t, out.Bids, 2)
	assert.Equal(t, "seven", out.Bids[0].Name)
	// **パスも 0 として送る。**誰が降りたかは以後の読みに要る。
	assert.Equal(t, int(domain.BostonBidPass), out.Bids[1].Level)
	assert.NotNil(t, out.HighBid)
	assert.Equal(t, "seven", out.HighBid.Name)
	assert.Equal(t, 3, out.Players[3].TricksWon)
	assert.Equal(t, 30, out.Players[3].Chips)
	assert.True(t, out.Players[3].IsDealer)
	assert.True(t, out.Players[0].IsCurrentTurn)
}

func TestBostonWebPresenter_Messages(t *testing.T) {
	for _, tc := range []struct {
		phase   domain.BostonPhase
		bidMade bool
		wantKey string
	}{
		{domain.BostonPhaseBid, false, "boston.bidPhase"},
		{domain.BostonPhaseCallPartner, false, "boston.callPartner"},
		{domain.BostonPhasePlay, false, "boston.playPhase"},
		{domain.BostonPhaseHandEnd, true, "boston.handMade"},
		{domain.BostonPhaseHandEnd, false, "boston.handFailed"},
	} {
		o := defaultBostonOpts()
		o.phase = tc.phase
		o.bidMade = tc.bidMade
		out := parseBostonOutput(t, new(presenter.BostonWebPresenter).Output(setupBostonWebMock(o), nil))
		assert.Equal(t, tc.wantKey, out.MessageCode)
	}

	t.Run("an error wins over any phase message", func(t *testing.T) {
		out := parseBostonOutput(t, new(presenter.BostonWebPresenter).Output(setupBostonWebMock(defaultBostonOpts()), errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
		assert.Empty(t, out.MessageCode)
	})
}

func TestBostonWebPresenter_GameEnd(t *testing.T) {
	for _, tc := range []struct {
		name    string
		winner  int
		wantKey string
	}{
		{"the human wins", 0, "boston.result.humanWin"},
		{"a CPU wins", 2, "boston.result.cpuWin"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := defaultBostonOpts()
			o.phase = domain.BostonPhaseGameEnd
			o.gameEnd = true
			o.winner = tc.winner
			out := parseBostonOutput(t, new(presenter.BostonWebPresenter).Output(setupBostonWebMock(o), nil))
			assert.Equal(t, tc.wantKey, out.MessageCode)
			assert.True(t, out.GameEndFlag)
		})
	}
}

func TestBostonWebPresenter_NoContractYet(t *testing.T) {
	o := defaultBostonOpts()
	o.phase = domain.BostonPhaseBid
	o.declarer = -1
	o.highBid = nil
	out := parseBostonOutput(t, new(presenter.BostonWebPresenter).Output(setupBostonWebMock(o), nil))
	assert.Nil(t, out.HighBid)
	assert.Equal(t, -1, out.DeclarerIdx)
	// 序列表は宣言前から送る。選択肢を出すのに要る。
	assert.NotEmpty(t, out.BidOptions)
}

func TestBostonWebPresenter_ActionLogOutput(t *testing.T) {
	m := setupBostonWebMock(defaultBostonOpts())
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotEmpty(t, new(presenter.BostonWebPresenter).ActionLogOutput(m))
}
