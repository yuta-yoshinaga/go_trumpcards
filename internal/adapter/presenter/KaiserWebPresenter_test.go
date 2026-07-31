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

func kzTestCard(suit, value int) *domain.Card { return domain.NewCard(suit, value, true) }

// makeKaiserPlayers builds four seats, the first human, with the given hands.
func makeKaiserPlayers(hands ...[]*domain.Card) []*domain.KaiserPlayer {
	out := make([]*domain.KaiserPlayer, 0, len(hands))
	for i, hand := range hands {
		p := domain.NewKaiserPlayer(i == 0)
		for _, c := range hand {
			p.AddCard(c)
		}
		out = append(out, p)
	}
	return out
}

func setupKaiserWebMock(phase domain.KaiserPhase) *interfaces.MockKaiserGame {
	m := new(interfaces.MockKaiserGame)
	players := makeKaiserPlayers(
		[]*domain.Card{kzTestCard(domain.CardDesignHeart, 5), kzTestCard(domain.CardDesignSpade, 1)},
		[]*domain.Card{kzTestCard(domain.CardDesignSpade, 3)},
		[]*domain.Card{kzTestCard(domain.CardDesignClover, 7)},
		[]*domain.Card{kzTestCard(domain.CardDesignDiamond, 8)},
	)
	m.On("GetPhase").Return(phase)
	m.On("GetHandNumber").Return(2)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(3)
	m.On("GetDeclarerIdx").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetContract").Return(domain.KaiserContractTrump)
	m.On("GetKittySize").Return(0)
	m.On("GetTrick").Return([]*domain.Card{kzTestCard(domain.CardDesignHeart, 1), nil})
	m.On("GetTrickLeaderIdx").Return(0)
	m.On("GetTrickNumber").Return(3)
	m.On("GetHeartFiveBy").Return(0)
	m.On("GetSpadeThreeBy").Return(1)
	m.On("IsBidMade").Return(true)
	m.On("GetTargetScore").Return(domain.KaiserTargetScore)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetConfig").Return(domain.DefaultKaiserConfig())
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(true)
	m.On("KaiserValidPlays", 0).Return([]int{1})
	m.On("GetBids").Return([]*domain.KaiserBid{
		{Player: 1, Value: 8, Contract: domain.KaiserContractTrump},
		{Player: 2, Value: 0},
		nil, // nil 混入でも落ちないこと
	})
	m.On("GetHighBid").Return(&domain.KaiserBid{Player: 1, Value: 8, Contract: domain.KaiserContractTrump})
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
	}
	for team := range domain.KaiserTeamCnt {
		m.On("GetHandPoints", team).Return(5 * (team + 1))
		m.On("GetScore", team).Return(10 * (team + 1))
	}
	return m
}

func parseKaiserOutput(t *testing.T, s string) *controller.KaiserWebOutput {
	t.Helper()
	var out controller.KaiserWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return &out
}

// **相手の手札は伏せる。**枚数だけ送る。
func TestKaiserWebPresenter_HidesOpponentHandsDuringPlay(t *testing.T) {
	m := setupKaiserWebMock(domain.KaiserPhasePlay)
	out := parseKaiserOutput(t, new(presenter.KaiserWebPresenter).Output(m, nil))

	assert.Len(t, out.Players, 4)
	assert.Len(t, out.Players[0].Cards, 2, "the human sees its own hand")
	for i := 1; i < 4; i++ {
		assert.Empty(t, out.Players[i].Cards, "seat %d must stay hidden", i)
		assert.Positive(t, out.Players[i].CardCount, "the count is still public")
	}
}

func TestKaiserWebPresenter_RevealsAtTheSettlement(t *testing.T) {
	m := setupKaiserWebMock(domain.KaiserPhaseHandEnd)
	out := parseKaiserOutput(t, new(presenter.KaiserWebPresenter).Output(m, nil))
	for i := range out.Players {
		assert.NotEmpty(t, out.Players[i].Cards, "seat %d must be revealed", i)
	}
}

// **パートナーは向かい合わせ。**席 0/2 と 1/3。
func TestKaiserWebPresenter_SendsTeams(t *testing.T) {
	m := setupKaiserWebMock(domain.KaiserPhasePlay)
	out := parseKaiserOutput(t, new(presenter.KaiserWebPresenter).Output(m, nil))
	assert.Equal(t, out.Players[0].Team, out.Players[2].Team)
	assert.Equal(t, out.Players[1].Team, out.Players[3].Team)
	assert.NotEqual(t, out.Players[0].Team, out.Players[1].Team)
}

// **出せる札はサーバーが決める。**追随が強制なのでフロントで再現するとずれる。
func TestKaiserWebPresenter_SendsValidPlaysOnlyOnTheHumansPlayTurn(t *testing.T) {
	m := setupKaiserWebMock(domain.KaiserPhasePlay)
	out := parseKaiserOutput(t, new(presenter.KaiserWebPresenter).Output(m, nil))
	assert.Equal(t, []int{1}, out.ValidPlays)

	bidding := setupKaiserWebMock(domain.KaiserPhaseBid)
	bidOut := parseKaiserOutput(t, new(presenter.KaiserWebPresenter).Output(bidding, nil))
	assert.Empty(t, bidOut.ValidPlays)
	bidding.AssertNotCalled(t, "KaiserValidPlays", 0)
}

func TestKaiserWebPresenter_TopLevelFields(t *testing.T) {
	m := setupKaiserWebMock(domain.KaiserPhasePlay)
	out := parseKaiserOutput(t, new(presenter.KaiserWebPresenter).Output(m, nil))

	assert.Equal(t, int(domain.KaiserPhasePlay), out.Phase)
	assert.Equal(t, 2, out.HandNumber)
	assert.Equal(t, 1, out.DeclarerIdx)
	assert.Equal(t, domain.CardDesignHeart, out.TrumpSuit)
	// nil をまぜたトリックでも落ちない。
	assert.Len(t, out.Trick, 1)
	assert.Equal(t, 3, out.TrickNumber)
	// nil をまぜたビッド履歴でも落ちない。
	assert.Len(t, out.Bids, 2)
	assert.Equal(t, 8, out.Bids[0].Value)
	// **パスは 0 として送る。**誰が降りたかは以後の読みに要る。
	assert.Equal(t, 0, out.Bids[1].Value)
	assert.NotNil(t, out.HighBid)
	assert.Equal(t, 8, out.HighBid.Value)
	assert.Equal(t, 0, out.HeartFiveBy)
	assert.Equal(t, 1, out.SpadeThreeBy)
	assert.Equal(t, domain.KaiserTargetScore, out.TargetScore)
	// **最低ビッドは 7。**フロントがボタンの範囲を出すのに要る。
	assert.Equal(t, domain.KaiserMinBid, out.MinBid)
	assert.Equal(t, 7, out.MinBid)
	assert.Equal(t, domain.KaiserMaxBid, out.MaxBid)
	assert.Equal(t, [domain.KaiserTeamCnt]int{5, 10}, out.TeamHandPoints)
	assert.Equal(t, [domain.KaiserTeamCnt]int{10, 20}, out.TeamScores)
	assert.True(t, out.Players[3].IsDealer)
	assert.True(t, out.Players[1].IsDeclarer)
	assert.True(t, out.Players[0].IsCurrentTurn)
}

func TestKaiserWebPresenter_Messages(t *testing.T) {
	for _, tc := range []struct {
		phase   domain.KaiserPhase
		wantKey string
	}{
		{domain.KaiserPhaseBid, "kaiser.bidPhase"},
		{domain.KaiserPhaseDiscard, "kaiser.discardPhase"},
		{domain.KaiserPhasePlay, "kaiser.playPhase"},
		{domain.KaiserPhaseHandEnd, "kaiser.handMade"},
	} {
		m := setupKaiserWebMock(tc.phase)
		out := parseKaiserOutput(t, new(presenter.KaiserWebPresenter).Output(m, nil))
		assert.Equal(t, tc.wantKey, out.MessageCode)
	}

	t.Run("an error wins over any phase message", func(t *testing.T) {
		m := setupKaiserWebMock(domain.KaiserPhasePlay)
		out := parseKaiserOutput(t, new(presenter.KaiserWebPresenter).Output(m, errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
		assert.Empty(t, out.MessageCode)
	})
}

// 切札の指定待ちは捨て札の案内と別物。
func TestKaiserWebPresenter_AsksForTrumpBeforeTheDiscard(t *testing.T) {
	m := setupKaiserWebMock(domain.KaiserPhaseDiscard)
	m.ExpectedCalls = nil
	players := makeKaiserPlayers(nil, nil, nil, nil)
	m.On("GetPhase").Return(domain.KaiserPhaseDiscard)
	m.On("GetHandNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(1)
	m.On("GetBidPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(1)
	// **まだ切札が決まっていない。**
	m.On("GetTrumpSuit").Return(0)
	m.On("GetContract").Return(domain.KaiserContractTrump)
	m.On("GetKittySize").Return(0)
	m.On("GetTrick").Return([]*domain.Card{})
	m.On("GetTrickLeaderIdx").Return(-1)
	m.On("GetTrickNumber").Return(0)
	m.On("GetHeartFiveBy").Return(-1)
	m.On("GetSpadeThreeBy").Return(-1)
	m.On("IsBidMade").Return(false)
	m.On("GetTargetScore").Return(domain.KaiserTargetScore)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetConfig").Return(domain.DefaultKaiserConfig())
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(false)
	m.On("GetBids").Return([]*domain.KaiserBid{})
	m.On("GetHighBid").Return((*domain.KaiserBid)(nil))
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
	}
	for team := range domain.KaiserTeamCnt {
		m.On("GetHandPoints", team).Return(0)
		m.On("GetScore", team).Return(0)
	}

	out := parseKaiserOutput(t, new(presenter.KaiserWebPresenter).Output(m, nil))
	assert.Equal(t, "kaiser.nameTrump", out.MessageCode)
	assert.Nil(t, out.HighBid)
}

// ベートは達成と字面で区別する。全得点の行き先が変わる。
func TestKaiserWebPresenter_TellsASetHandApart(t *testing.T) {
	m := setupKaiserWebMock(domain.KaiserPhaseHandEnd)
	m.ExpectedCalls = nil
	players := makeKaiserPlayers(nil, nil, nil, nil)
	m.On("GetPhase").Return(domain.KaiserPhaseHandEnd)
	m.On("GetHandNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	m.On("GetContract").Return(domain.KaiserContractTrump)
	m.On("GetKittySize").Return(0)
	m.On("GetTrick").Return([]*domain.Card{})
	m.On("GetTrickLeaderIdx").Return(0)
	m.On("GetTrickNumber").Return(8)
	m.On("GetHeartFiveBy").Return(0)
	m.On("GetSpadeThreeBy").Return(0)
	m.On("IsBidMade").Return(false)
	m.On("GetTargetScore").Return(domain.KaiserTargetScore)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetConfig").Return(domain.DefaultKaiserConfig())
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(false)
	m.On("GetBids").Return([]*domain.KaiserBid{})
	m.On("GetHighBid").Return((*domain.KaiserBid)(nil))
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
	}
	for team := range domain.KaiserTeamCnt {
		m.On("GetHandPoints", team).Return(0)
		m.On("GetScore", team).Return(0)
	}

	out := parseKaiserOutput(t, new(presenter.KaiserWebPresenter).Output(m, nil))
	assert.Equal(t, "kaiser.handSet", out.MessageCode)
	assert.False(t, out.BidMade)
}

// **勝敗はチームで見る。**人間は席 0 = チーム 0。
func TestKaiserWebPresenter_GameEndIsByTeam(t *testing.T) {
	for _, tc := range []struct {
		name    string
		team    int
		wantKey string
	}{
		{"the human's team wins", 0, "kaiser.result.humanWin"},
		{"the other team wins", 1, "kaiser.result.cpuWin"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := setupKaiserWebMock(domain.KaiserPhaseGameEnd)
			m.ExpectedCalls = nil
			players := makeKaiserPlayers(nil, nil, nil, nil)
			m.On("GetPhase").Return(domain.KaiserPhaseGameEnd)
			m.On("GetHandNumber").Return(9)
			m.On("GetCurrentPlayerIdx").Return(0)
			m.On("GetBidPlayerIdx").Return(1)
			m.On("GetDealerIdx").Return(0)
			m.On("GetDeclarerIdx").Return(1)
			m.On("GetTrumpSuit").Return(domain.CardDesignHeart)
			m.On("GetContract").Return(domain.KaiserContractTrump)
			m.On("GetKittySize").Return(0)
			m.On("GetTrick").Return([]*domain.Card{})
			m.On("GetTrickLeaderIdx").Return(0)
			m.On("GetTrickNumber").Return(8)
			m.On("GetHeartFiveBy").Return(-1)
			m.On("GetSpadeThreeBy").Return(-1)
			m.On("IsBidMade").Return(true)
			m.On("GetTargetScore").Return(domain.KaiserTargetScore)
			m.On("GetGameEndFlag").Return(true)
			m.On("GetWinnerTeam").Return(tc.team)
			m.On("GetConfig").Return(domain.DefaultKaiserConfig())
			m.On("GetPlayers").Return(players)
			m.On("IsHumanTurn").Return(false)
			m.On("GetBids").Return([]*domain.KaiserBid{})
			m.On("GetHighBid").Return((*domain.KaiserBid)(nil))
			for i := range players {
				m.On("GetPlayer", i).Return(players[i])
			}
			for team := range domain.KaiserTeamCnt {
				m.On("GetHandPoints", team).Return(0)
				m.On("GetScore", team).Return(0)
			}

			out := parseKaiserOutput(t, new(presenter.KaiserWebPresenter).Output(m, nil))
			assert.Equal(t, tc.wantKey, out.MessageCode)
			assert.True(t, out.GameEndFlag)
			assert.Equal(t, tc.team, out.WinnerTeam)
		})
	}
}

func TestKaiserWebPresenter_ActionLogOutput(t *testing.T) {
	m := setupKaiserWebMock(domain.KaiserPhasePlay)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotEmpty(t, new(presenter.KaiserWebPresenter).ActionLogOutput(m))
}
