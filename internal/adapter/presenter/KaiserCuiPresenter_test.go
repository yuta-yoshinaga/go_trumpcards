//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func setupKaiserCuiMock(phase domain.KaiserPhase, trump int, contract domain.KaiserContract, highBid *domain.KaiserBid) *interfaces.MockKaiserGame {
	m := new(interfaces.MockKaiserGame)
	players := makeKaiserPlayers(
		[]*domain.Card{kzTestCard(domain.CardDesignHeart, 5), kzTestCard(domain.CardDesignSpade, 1)},
		[]*domain.Card{kzTestCard(domain.CardDesignSpade, 3)},
		[]*domain.Card{kzTestCard(domain.CardDesignClover, 7)},
		[]*domain.Card{kzTestCard(domain.CardDesignDiamond, 8)},
	)
	m.On("GetPhase").Return(phase)
	m.On("GetHandNumber").Return(1)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(1)
	m.On("GetDealerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(1)
	m.On("GetTrumpSuit").Return(trump)
	m.On("GetContract").Return(contract)
	m.On("GetTrick").Return([]*domain.Card{kzTestCard(domain.CardDesignHeart, 1)})
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerTeam").Return(-1)
	m.On("IsBidMade").Return(true)
	m.On("GetTargetScore").Return(domain.KaiserTargetScore)
	m.On("GetHighBid").Return(highBid)
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(true)
	m.On("KaiserValidPlays", 0).Return([]int{1})
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
	}
	for team := range domain.KaiserTeamCnt {
		m.On("GetHandPoints", team).Return(team + 1)
		m.On("GetScore", team).Return(10 * (team + 1))
	}
	return m
}

func TestKaiserCuiPresenter_HidesOpponentHands(t *testing.T) {
	m := setupKaiserCuiMock(domain.KaiserPhasePlay, domain.CardDesignHeart, domain.KaiserContractTrump,
		&domain.KaiserBid{Player: 1, Value: 8})
	out := new(presenter.KaiserCuiPresenter).Output(m, nil)

	assert.Contains(t, out, "[0]")
	assert.Contains(t, out, "非公開")
	assert.Contains(t, out, "[親]")
	assert.Contains(t, out, "[宣言]")
	// **チームが読めないとパートナーが判らない。**
	assert.Contains(t, out, "(T0)")
	assert.Contains(t, out, "(T1)")
	assert.Contains(t, out, "この局:")
}

// **出せる札を出さないと操作できない。**追随が強制。
func TestKaiserCuiPresenter_ListsThePlayableIndexes(t *testing.T) {
	m := setupKaiserCuiMock(domain.KaiserPhasePlay, domain.CardDesignHeart, domain.KaiserContractTrump,
		&domain.KaiserBid{Player: 1, Value: 8})
	assert.Contains(t, new(presenter.KaiserCuiPresenter).Output(m, nil), "出せる札: 1")
}

func TestKaiserCuiPresenter_ShowsTheContractOnlyOnceBid(t *testing.T) {
	withBid := setupKaiserCuiMock(domain.KaiserPhasePlay, domain.CardDesignHeart, domain.KaiserContractTrump,
		&domain.KaiserBid{Player: 1, Value: 8})
	assert.Contains(t, new(presenter.KaiserCuiPresenter).Output(withBid, nil), "契約: 8")

	noBid := setupKaiserCuiMock(domain.KaiserPhaseBid, 0, domain.KaiserContractTrump, nil)
	assert.NotContains(t, new(presenter.KaiserCuiPresenter).Output(noBid, nil), "契約:")
}

// ロー・ノートランプは表示で区別できる必要がある。ランクが逆転するため。
func TestKaiserCuiPresenter_NamesTheContractKind(t *testing.T) {
	for _, tc := range []struct {
		contract domain.KaiserContract
		want     string
	}{
		{domain.KaiserContractTrump, "切札あり"},
		{domain.KaiserContractNoTrump, "ノートランプ"},
		{domain.KaiserContractLowNoTrump, "ロー・ノートランプ"},
	} {
		m := setupKaiserCuiMock(domain.KaiserPhasePlay, domain.CardDesignHeart, tc.contract,
			&domain.KaiserBid{Player: 1, Value: 8})
		assert.Contains(t, new(presenter.KaiserCuiPresenter).Output(m, nil), tc.want)
	}
}

func TestKaiserCuiPresenter_PhasePrompts(t *testing.T) {
	for _, tc := range []struct {
		phase domain.KaiserPhase
		trump int
		want  string
	}{
		{domain.KaiserPhaseBid, domain.CardDesignHeart, "最低は7"},
		{domain.KaiserPhaseDiscard, 0, "切札を指定してください"},
		{domain.KaiserPhaseDiscard, domain.CardDesignHeart, "♥5 と ♠3 は捨てられません"},
		{domain.KaiserPhasePlay, domain.CardDesignHeart, "追随は強制"},
		{domain.KaiserPhaseHandEnd, domain.CardDesignHeart, "次の局へ"},
	} {
		m := setupKaiserCuiMock(tc.phase, tc.trump, domain.KaiserContractTrump,
			&domain.KaiserBid{Player: 1, Value: 8})
		assert.Contains(t, new(presenter.KaiserCuiPresenter).Output(m, nil), tc.want)
	}
}

// ベートは達成と字面で区別する。
func TestKaiserCuiPresenter_TellsASetHandApart(t *testing.T) {
	m := setupKaiserCuiMock(domain.KaiserPhaseHandEnd, domain.CardDesignHeart, domain.KaiserContractTrump,
		&domain.KaiserBid{Player: 1, Value: 8})
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
	m.On("GetTrick").Return([]*domain.Card{})
	m.On("GetGameEndFlag").Return(false)
	m.On("GetWinnerTeam").Return(-1)
	m.On("IsBidMade").Return(false)
	m.On("GetTargetScore").Return(domain.KaiserTargetScore)
	m.On("GetHighBid").Return(&domain.KaiserBid{Player: 1, Value: 8})
	m.On("GetPlayers").Return(players)
	m.On("IsHumanTurn").Return(false)
	for i := range players {
		m.On("GetPlayer", i).Return(players[i])
	}
	for team := range domain.KaiserTeamCnt {
		m.On("GetHandPoints", team).Return(0)
		m.On("GetScore", team).Return(0)
	}
	assert.Contains(t, new(presenter.KaiserCuiPresenter).Output(m, nil), "宣言額がそのままマイナス")
}

func TestKaiserCuiPresenter_ErrorAndGameEnd(t *testing.T) {
	m := setupKaiserCuiMock(domain.KaiserPhasePlay, domain.CardDesignHeart, domain.KaiserContractTrump,
		&domain.KaiserBid{Player: 1, Value: 8})
	assert.Contains(t, new(presenter.KaiserCuiPresenter).Output(m, errors.New("boom")), "boom")

	end := new(interfaces.MockKaiserGame)
	players := makeKaiserPlayers(nil, nil, nil, nil)
	end.On("GetPhase").Return(domain.KaiserPhaseGameEnd)
	end.On("GetHandNumber").Return(9)
	end.On("GetCurrentPlayerIdx").Return(0)
	end.On("GetBidPlayerIdx").Return(1)
	end.On("GetDealerIdx").Return(0)
	end.On("GetDeclarerIdx").Return(1)
	end.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	end.On("GetContract").Return(domain.KaiserContractTrump)
	end.On("GetTrick").Return([]*domain.Card{})
	end.On("GetGameEndFlag").Return(true)
	end.On("GetWinnerTeam").Return(0)
	end.On("IsBidMade").Return(true)
	end.On("GetTargetScore").Return(domain.KaiserTargetScore)
	end.On("GetHighBid").Return((*domain.KaiserBid)(nil))
	end.On("GetPlayers").Return(players)
	end.On("IsHumanTurn").Return(false)
	for i := range players {
		end.On("GetPlayer", i).Return(players[i])
	}
	for team := range domain.KaiserTeamCnt {
		end.On("GetHandPoints", team).Return(0)
		end.On("GetScore", team).Return(0)
	}
	assert.Contains(t, new(presenter.KaiserCuiPresenter).Output(end, nil), "ゲーム終了")
}

func TestKaiserCuiPresenter_ActionLogOutput(t *testing.T) {
	m := setupKaiserCuiMock(domain.KaiserPhasePlay, domain.CardDesignHeart, domain.KaiserContractTrump,
		&domain.KaiserBid{Player: 1, Value: 8})
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	assert.NotNil(t, new(presenter.KaiserCuiPresenter).ActionLogOutput(m))
}
