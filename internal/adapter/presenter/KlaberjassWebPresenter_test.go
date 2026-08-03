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

// makeKlaberjassPlayers builds two seats, the first human, with the given hands.
func makeKlaberjassPlayers(hands ...[]*domain.Card) []*domain.KlaberjassPlayer {
	out := make([]*domain.KlaberjassPlayer, 0, len(hands))
	for i, hand := range hands {
		p := domain.NewKlaberjassPlayer(i == 0)
		for _, c := range hand {
			p.AddCard(c)
		}
		out = append(out, p)
	}
	return out
}

func kjTestCard(suit, value int) *domain.Card { return domain.NewCard(suit, value, true) }

func setupKlaberjassWebMock(phase domain.KlaberjassPhase) (*interfaces.MockKlaberjassGame, []*domain.KlaberjassPlayer) {
	m := new(interfaces.MockKlaberjassGame)
	players := makeKlaberjassPlayers(
		[]*domain.Card{kjTestCard(domain.CardDesignSpade, 11), kjTestCard(domain.CardDesignHeart, 1)},
		[]*domain.Card{kjTestCard(domain.CardDesignClover, 7), kjTestCard(domain.CardDesignClover, 8)},
	)
	m.On("GetPhase").Return(phase)
	m.On("GetDealNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTurnUpCard").Return(kjTestCard(domain.CardDesignSpade, 13))
	m.On("GetMakerIdx").Return(1)
	m.On("GetTrick").Return([]*domain.Card{kjTestCard(domain.CardDesignHeart, 10), nil})
	m.On("GetTrickLeaderIdx").Return(0)
	m.On("GetTrickNumber").Return(2)
	m.On("GetSequenceWinner").Return(0)
	m.On("GetBelaHolder").Return(1)
	m.On("IsBelaScored").Return(true)
	m.On("IsDixUsed").Return(true)
	m.On("IsBete").Return(false)
	m.On("GetSchmeissBy").Return(-1)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultKlaberjassConfig())
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(true)
	m.On("KlaberjassValidPlays", 0).Return([]int{1})
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
		m.On("GetHandPoints", i).Return(10 * (i + 1))
		m.On("GetScore", i).Return(100 * (i + 1))
		m.On("GetSequences", i).Return([]*domain.KlaberjassSequence{
			{Suit: domain.CardDesignHeart, TopValue: 13, Length: 3, Points: 20},
			nil,
		})
	}
	return m, players
}

func parseKlaberjassOutput(t *testing.T, s string) *controller.KlaberjassWebOutput {
	t.Helper()
	var out controller.KlaberjassWebOutput
	assert.NoError(t, json.Unmarshal([]byte(s), &out))
	return &out
}

// **相手の手札も役も伏せる。**役は申告し合う勝負なので、見えたら成立しない。
func TestKlaberjassWebPresenter_HidesTheOpponentDuringPlay(t *testing.T) {
	m, _ := setupKlaberjassWebMock(domain.KlaberjassPhasePlay)
	out := parseKlaberjassOutput(t, new(presenter.KlaberjassWebPresenter).Output(m, nil))

	assert.Len(t, out.Players, 2)
	assert.Len(t, out.Players[0].Cards, 2, "the human sees its own hand")
	assert.Empty(t, out.Players[1].Cards, "the opponent's hand stays hidden")
	// 枚数は送る。何枚残っているかは公開情報。
	assert.Equal(t, 2, out.Players[1].CardCount)
	for i := range out.Players {
		assert.Empty(t, out.Players[i].Sequences, "sequences stay hidden until the settlement")
	}
}

func TestKlaberjassWebPresenter_RevealsAtTheSettlement(t *testing.T) {
	m, _ := setupKlaberjassWebMock(domain.KlaberjassPhaseHandEnd)
	out := parseKlaberjassOutput(t, new(presenter.KlaberjassWebPresenter).Output(m, nil))

	assert.Len(t, out.Players[1].Cards, 2, "hands are revealed once the hand is settled")
	// nil 混じりの役スライスでも落とさず、実体だけ送る。
	assert.Len(t, out.Players[0].Sequences, 1)
	assert.Equal(t, 20, out.Players[0].Sequences[0].Points)
	assert.Equal(t, 13, out.Players[0].Sequences[0].TopValue)
}

// **出せる札はサーバーが決める。**追随・切札・上乗せが全部強制なので、
// フロントで再現するとずれる。
func TestKlaberjassWebPresenter_SendsValidPlaysOnlyOnTheHumansPlayTurn(t *testing.T) {
	m, _ := setupKlaberjassWebMock(domain.KlaberjassPhasePlay)
	out := parseKlaberjassOutput(t, new(presenter.KlaberjassWebPresenter).Output(m, nil))
	assert.Equal(t, []int{1}, out.ValidPlays)

	// ビッド中は出せる札という概念が無い。
	bidding, _ := setupKlaberjassWebMock(domain.KlaberjassPhaseBidTurnUp)
	bidOut := parseKlaberjassOutput(t, new(presenter.KlaberjassWebPresenter).Output(bidding, nil))
	assert.Empty(t, bidOut.ValidPlays)
	bidding.AssertNotCalled(t, "KlaberjassValidPlays", 0)
}

func TestKlaberjassWebPresenter_TopLevelFields(t *testing.T) {
	m, _ := setupKlaberjassWebMock(domain.KlaberjassPhasePlay)
	out := parseKlaberjassOutput(t, new(presenter.KlaberjassWebPresenter).Output(m, nil))

	assert.Equal(t, int(domain.KlaberjassPhasePlay), out.Phase)
	assert.Equal(t, 1, out.DealNumber)
	assert.Equal(t, domain.CardDesignSpade, out.TrumpSuit)
	assert.NotNil(t, out.TurnUpCard)
	assert.Equal(t, 1, out.MakerIdx)
	// nil をまぜたトリックでも落ちない。
	assert.Len(t, out.Trick, 1)
	assert.Equal(t, 2, out.TrickNumber)
	assert.Equal(t, 0, out.SequenceWinner)
	assert.Equal(t, 1, out.BelaHolder)
	assert.True(t, out.BelaScored)
	assert.True(t, out.DixUsed)
	assert.Equal(t, -1, out.SchmeissBy)
	assert.Equal(t, domain.KlaberjassTargetScoreDefault, out.TargetScore)
	assert.True(t, out.Players[1].IsMaker)
	assert.True(t, out.Players[0].IsDealer)
	assert.True(t, out.Players[0].IsCurrentTurn)
	assert.Equal(t, 200, out.Players[1].Score)
	assert.Equal(t, 10, out.Players[0].HandPoints)
}

func TestKlaberjassWebPresenter_Messages(t *testing.T) {
	for _, tc := range []struct {
		phase   domain.KlaberjassPhase
		wantKey string
	}{
		{domain.KlaberjassPhaseBidTurnUp, "klaberjass.bidTurnUp"},
		{domain.KlaberjassPhaseBidFree, "klaberjass.bidFree"},
		{domain.KlaberjassPhaseSchmeiss, "klaberjass.schmeissPending"},
		{domain.KlaberjassPhasePlay, "klaberjass.playPhase"},
		{domain.KlaberjassPhaseHandEnd, "klaberjass.handEnd"},
	} {
		m, _ := setupKlaberjassWebMock(tc.phase)
		out := parseKlaberjassOutput(t, new(presenter.KlaberjassWebPresenter).Output(m, nil))
		assert.Equal(t, tc.wantKey, out.MessageCode)
	}

	t.Run("an error wins over any phase message", func(t *testing.T) {
		m, _ := setupKlaberjassWebMock(domain.KlaberjassPhasePlay)
		out := parseKlaberjassOutput(t, new(presenter.KlaberjassWebPresenter).Output(m, errors.New("boom")))
		assert.Equal(t, "boom", out.Message)
		assert.Empty(t, out.MessageCode)
	})
}

// ベートは通常の精算とは別に伝える。全得点が移る特別な結果なので。
func TestKlaberjassWebPresenter_BeteHasItsOwnMessage(t *testing.T) {
	m := new(interfaces.MockKlaberjassGame)
	players := makeKlaberjassPlayers([]*domain.Card{}, []*domain.Card{})
	m.On("GetPhase").Return(domain.KlaberjassPhaseHandEnd)
	m.On("GetDealNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetTurnUpCard").Return((*domain.Card)(nil))
	m.On("GetMakerIdx").Return(0)
	m.On("GetTrick").Return([]*domain.Card{})
	m.On("GetTrickLeaderIdx").Return(0)
	m.On("GetTrickNumber").Return(9)
	m.On("GetSequenceWinner").Return(-1)
	m.On("GetBelaHolder").Return(-1)
	m.On("IsBelaScored").Return(false)
	m.On("IsDixUsed").Return(false)
	m.On("IsBete").Return(true)
	m.On("GetSchmeissBy").Return(-1)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetConfig").Return(domain.DefaultKlaberjassConfig())
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(false)
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
		m.On("GetHandPoints", i).Return(0)
		m.On("GetScore", i).Return(0)
		m.On("GetSequences", i).Return([]*domain.KlaberjassSequence{})
	}

	out := parseKlaberjassOutput(t, new(presenter.KlaberjassWebPresenter).Output(m, nil))
	assert.Equal(t, "klaberjass.bete", out.MessageCode)
	assert.True(t, out.Bete)
	assert.Nil(t, out.TurnUpCard)
}

func TestKlaberjassWebPresenter_GameEnd(t *testing.T) {
	for _, tc := range []struct {
		name    string
		winner  int
		wantKey string
	}{
		{"the human wins", 0, "klaberjass.result.humanWin"},
		{"the CPU wins", 1, "klaberjass.result.cpuWin"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, _ := setupKlaberjassWebMock(domain.KlaberjassPhaseGameEnd)
			m.ExpectedCalls = nil
			players := makeKlaberjassPlayers([]*domain.Card{}, []*domain.Card{})
			m.On("GetPhase").Return(domain.KlaberjassPhaseGameEnd)
			m.On("GetDealNumber").Return(9)
			m.On("GetCurrentPlayerIdx").Return(0)
			m.On("GetBidPlayerIdx").Return(1)
			m.On("GetDealerIdx").Return(0)
			m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
			m.On("GetTurnUpCard").Return((*domain.Card)(nil))
			m.On("GetMakerIdx").Return(0)
			m.On("GetTrick").Return([]*domain.Card{})
			m.On("GetTrickLeaderIdx").Return(0)
			m.On("GetTrickNumber").Return(9)
			m.On("GetSequenceWinner").Return(-1)
			m.On("GetBelaHolder").Return(-1)
			m.On("IsBelaScored").Return(false)
			m.On("IsDixUsed").Return(false)
			m.On("IsBete").Return(false)
			m.On("GetSchmeissBy").Return(-1)
			m.On("GetGameEndFlag").Return(true)
			m.On("GetWinnerIdx").Return(tc.winner)
			m.On("GetConfig").Return(domain.DefaultKlaberjassConfig())
			m.On("GetPlayers").Return(players)
			m.On("IsHumanTurn").Return(false)
			for i := range players {
				m.On("GetPlayer", i).Return(players[i])
				m.On("GetHandPoints", i).Return(0)
				m.On("GetScore", i).Return(0)
				m.On("GetSequences", i).Return([]*domain.KlaberjassSequence{})
			}

			out := parseKlaberjassOutput(t, new(presenter.KlaberjassWebPresenter).Output(m, nil))
			assert.Equal(t, tc.wantKey, out.MessageCode)
			assert.True(t, out.GameEndFlag)
		})
	}
}

func TestKlaberjassWebPresenter_ActionLogOutput(t *testing.T) {
	m, _ := setupKlaberjassWebMock(domain.KlaberjassPhasePlay)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotEmpty(t, new(presenter.KlaberjassWebPresenter).ActionLogOutput(m))
}
