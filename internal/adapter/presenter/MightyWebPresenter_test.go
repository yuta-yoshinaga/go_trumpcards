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

// setupMightyWebMock returns a MockMightyGame with safe defaults for Web
// presenter tests.
func setupMightyWebMock() *interfaces.MockMightyGame {
	m := new(interfaces.MockMightyGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(0)
	m.On("GetWinningBidNoTrump").Return(false)
	m.On("GetPartnerCard").Return((*domain.Card)(nil))
	m.On("GetDeclarerIdx").Return(-1)
	m.On("GetPartnerIdx").Return(-1)
	m.On("GetPartnerRevealed").Return(false)
	m.On("GetHighestBid").Return(0)
	m.On("GetHighestBidder").Return(-1)
	m.On("GetKitty").Return(([]*domain.Card)(nil))
	m.On("GetCurrentTrick").Return([]*domain.MightyTrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.MightyPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(domain.MightyWinnerUndecided)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultMightyConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	// **Output() も受動ヒントを埋める**ようになった (#4483)。既定は「ヒント無し」。
	// **base だけに置く。**removeMockCall は最初の 1 件しか外さない。
	m.On("GetHint").Return(nil).Maybe()

	return m
}

func setupMightyWebMockWithPlayers() (*interfaces.MockMightyGame, []*domain.MightyPlayer) {
	m := setupMightyWebMock()
	players := makeMightyPlayers()
	m.On("GetPlayerCnt").Return(5)
	for i, p := range players {
		m.On("GetPlayer", i).Return(p)
	}
	return m, players
}

func TestMightyWebPresenter_Output(t *testing.T) {
	p := new(presenter.MightyWebPresenter)

	t.Run("base output + per-player fields", func(t *testing.T) {
		m, players := setupMightyWebMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[1].SetCumulativeScore(50)
		players[1].SetRoundScore(10)
		players[1].SetPointCards(3)
		players[1].SetBid(14)
		players[1].SetBidNoTrump(true)
		players[1].AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, false)})

		result := p.Output(m, nil)
		var resObj controller.MightyWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))

		// Base fields.
		assert.Equal(t, 5, len(resObj.Players))
		assert.Equal(t, int(domain.MightyPhasePlay), resObj.Phase)
		assert.Equal(t, 1, resObj.RoundNumber)
		assert.Equal(t, -1, resObj.DeclarerIdx)
		assert.Equal(t, -1, resObj.PartnerIdx)
		assert.Equal(t, domain.MightyWinnerUndecided, resObj.WinnerTeam)
		assert.Empty(t, resObj.CurrentTrick)
		// Human cards shown, CPU hidden.
		assert.True(t, resObj.Players[0].IsHuman)
		assert.Equal(t, "SPADE", resObj.Players[0].Cards[0].Design)
		assert.Len(t, resObj.Players[1].Cards, 0)
		assert.Equal(t, 1, resObj.Players[1].CardCount)
		// Per-player scoring & no-trump bid flag.
		assert.Equal(t, 50, resObj.Players[1].CumulativeScore)
		assert.Equal(t, 10, resObj.Players[1].RoundScore)
		assert.Equal(t, 3, resObj.Players[1].PointCards)
		assert.Equal(t, 14, resObj.Players[1].Bid)
		assert.True(t, resObj.Players[1].BidNoTrump)
		assert.Equal(t, 1, resObj.Players[1].TrickCount)
	})

	t.Run("declarer and partner flags surface", func(t *testing.T) {
		m, players := setupMightyWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPartnerRevealed")
		m.On("GetPartnerRevealed").Return(true)
		players[0].SetIsDeclarer(true)
		players[1].SetIsPartner(true)

		result := p.Output(m, nil)
		var resObj controller.MightyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)

		assert.True(t, resObj.Players[0].IsDeclarer)
		assert.True(t, resObj.Players[1].IsPartner)
		assert.True(t, resObj.PartnerRevealed)
	})

	t.Run("isPartner hidden when partner not revealed", func(t *testing.T) {
		m, players := setupMightyWebMockWithPlayers()
		players[1].SetIsPartner(true)
		result := p.Output(m, nil)
		var resObj controller.MightyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.False(t, resObj.Players[1].IsPartner)
	})

	t.Run("partner card populated", func(t *testing.T) {
		m, _ := setupMightyWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPartnerCard")
		m.On("GetPartnerCard").Return(domain.NewCard(domain.CardDesignHeart, 13, false))
		result := p.Output(m, nil)
		var resObj controller.MightyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.NotNil(t, resObj.PartnerCard)
		assert.Equal(t, "HEART", resObj.PartnerCard.Design)
		assert.Equal(t, 13, resObj.PartnerCard.Value)
	})

	t.Run("trick serialization with and without joker lead", func(t *testing.T) {
		// without joker lead → omitempty hides the IsJokerLead field
		m1, _ := setupMightyWebMockWithPlayers()
		m1.ExpectedCalls = removeMockCall(m1.ExpectedCalls, "GetCurrentTrick")
		m1.On("GetCurrentTrick").Return([]*domain.MightyTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
		})
		r1 := p.Output(m1, nil)
		assert.NotContains(t, r1, `"isJokerLead":true`)
		assert.Contains(t, r1, `"playerIdx":0`)

		// with joker lead → flag and demand suit appear
		m2, _ := setupMightyWebMockWithPlayers()
		m2.ExpectedCalls = removeMockCall(m2.ExpectedCalls, "GetCurrentTrick")
		m2.On("GetCurrentTrick").Return([]*domain.MightyTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignJoker, 1, false), IsJokerLead: true, LeadDemandSuit: 2},
		})
		r2 := p.Output(m2, nil)
		assert.Contains(t, r2, `"isJokerLead":true`)
		assert.Contains(t, r2, `"leadDemandSuit":2`)
	})

	t.Run("kitty visibility tied to KittyExchange phase", func(t *testing.T) {
		// Outside KittyExchange → kitty omitted.
		m1, _ := setupMightyWebMockWithPlayers()
		var resObj1 controller.MightyWebOutput
		_ = json.Unmarshal([]byte(p.Output(m1, nil)), &resObj1)
		assert.Nil(t, resObj1.Kitty)

		// Inside KittyExchange → kitty present.
		m2, _ := setupMightyWebMockWithPlayers()
		m2.ExpectedCalls = removeMockCall(m2.ExpectedCalls, "GetPhase")
		m2.ExpectedCalls = removeMockCall(m2.ExpectedCalls, "GetKitty")
		m2.On("GetPhase").Return(domain.MightyPhaseKittyExchange)
		m2.On("GetKitty").Return([]*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 10, false),
			domain.NewCard(domain.CardDesignHeart, 7, false),
		})
		var resObj2 controller.MightyWebOutput
		_ = json.Unmarshal([]byte(p.Output(m2, nil)), &resObj2)
		assert.Len(t, resObj2.Kitty, 3)
	})

	t.Run("config values flow through", func(t *testing.T) {
		m, _ := setupMightyWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetConfig")
		m.On("GetConfig").Return(domain.MightyConfig{
			CpuDifficulty: domain.MightyCpuDifficultyHard,
			MinBid:        15,
			NoTrumpExtra:  4,
			PointLimit:    200,
		})
		result := p.Output(m, nil)
		var resObj controller.MightyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, int(domain.MightyCpuDifficultyHard), resObj.Config.CpuDifficulty)
		assert.Equal(t, 15, resObj.Config.MinBid)
		assert.Equal(t, 4, resObj.Config.NoTrumpExtra)
		assert.Equal(t, 200, resObj.Config.PointLimit)
	})

	t.Run("winning bid no-trump flows through", func(t *testing.T) {
		m, _ := setupMightyWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinningBidNoTrump")
		m.On("GetWinningBidNoTrump").Return(true)
		result := p.Output(m, nil)
		var resObj controller.MightyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.WinningBidNoTrump)
	})

	t.Run("error message takes priority over phase message", func(t *testing.T) {
		m, _ := setupMightyWebMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		var resObj controller.MightyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "boom", resObj.Message)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("game end messageCodes by winner team", func(t *testing.T) {
		for team, code := range map[int]string{
			domain.MightyWinnerDeclarer:   "mighty.gameEnd.declarerWins",
			domain.MightyWinnerOpposition: "mighty.gameEnd.oppositionWins",
		} {
			m, _ := setupMightyWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
			m.On("GetGameEndFlag").Return(true)
			m.On("GetWinnerTeam").Return(team)
			result := p.Output(m, nil)
			var resObj controller.MightyWebOutput
			_ = json.Unmarshal([]byte(result), &resObj)
			assert.True(t, resObj.GameEndFlag)
			assert.Equal(t, code, resObj.MessageCode, "team=%d", team)
		}
	})

	t.Run("phase messageCodes", func(t *testing.T) {
		cases := []struct {
			phase domain.MightyPhase
			code  string
		}{
			{domain.MightyPhaseBid, "mighty.bidPhase"},
			{domain.MightyPhaseTrumpAndFriend, "mighty.trumpAndFriendPhase"},
			{domain.MightyPhaseKittyExchange, "mighty.kittyExchange"},
			{domain.MightyPhaseTrickEnd, "mighty.trickEnd"},
			{domain.MightyPhaseRoundEnd, "mighty.roundEnd"},
		}
		for _, c := range cases {
			m, _ := setupMightyWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.On("GetPhase").Return(c.phase)
			result := p.Output(m, nil)
			var resObj controller.MightyWebOutput
			_ = json.Unmarshal([]byte(result), &resObj)
			assert.Equal(t, c.code, resObj.MessageCode, "phase %v", c.phase)
		}
	})

	t.Run("play phase lead vs follow messageCodes", func(t *testing.T) {
		m, _ := setupMightyWebMockWithPlayers()
		result := p.Output(m, nil)
		var resObj controller.MightyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "mighty.playPhase.lead", resObj.MessageCode)

		m2, _ := setupMightyWebMockWithPlayers()
		m2.ExpectedCalls = removeMockCall(m2.ExpectedCalls, "GetCurrentTrick")
		m2.On("GetCurrentTrick").Return([]*domain.MightyTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
		})
		result2 := p.Output(m2, nil)
		var resObj2 controller.MightyWebOutput
		_ = json.Unmarshal([]byte(result2), &resObj2)
		assert.Equal(t, "mighty.playPhase.follow", resObj2.MessageCode)
	})
}

func TestMightyWebPresenter_HintOutput(t *testing.T) {
	p := new(presenter.MightyWebPresenter)
	idx, bid := 2, 16
	nt := true
	suit, pSuit, pVal, js := domain.CardDesignSpade, 3, 13, 2

	t.Run("hint variants serialize the expected fields", func(t *testing.T) {
		cases := []struct {
			name  string
			hint  *domain.MightyHint
			check func(t *testing.T, h *controller.MightyWebOutputHint)
		}{
			{
				name: "card index",
				hint: &domain.MightyHint{CardIndex: &idx, Reason: "play_low"},
				check: func(t *testing.T, h *controller.MightyWebOutputHint) {
					assert.Equal(t, &idx, h.CardIndex)
					assert.Equal(t, "play_low", h.Reason)
				},
			},
			{
				name: "bid with noTrump",
				hint: &domain.MightyHint{Bid: &bid, BidNoTrump: &nt, Reason: "strategic_bid"},
				check: func(t *testing.T, h *controller.MightyWebOutputHint) {
					assert.Equal(t, &bid, h.Bid)
					assert.Equal(t, &nt, h.BidNoTrump)
				},
			},
			{
				name: "trump + partner",
				hint: &domain.MightyHint{TrumpSuit: &suit, PartnerSuit: &pSuit, PartnerValue: &pVal, Reason: "strategic_declare"},
				check: func(t *testing.T, h *controller.MightyWebOutputHint) {
					assert.Equal(t, &suit, h.TrumpSuit)
					assert.Equal(t, &pSuit, h.PartnerSuit)
					assert.Equal(t, &pVal, h.PartnerValue)
				},
			},
			{
				name: "discard indices",
				hint: &domain.MightyHint{DiscardIndices: []int{0, 1, 2}, Reason: "strategic_discard"},
				check: func(t *testing.T, h *controller.MightyWebOutputHint) {
					assert.Equal(t, []int{0, 1, 2}, h.DiscardIndices)
				},
			},
			{
				name: "joker lead suit",
				hint: &domain.MightyHint{JokerLeadSuit: &js, Reason: "joker_lead"},
				check: func(t *testing.T, h *controller.MightyWebOutputHint) {
					assert.Equal(t, &js, h.JokerLeadSuit)
				},
			},
		}
		for _, c := range cases {
			m, _ := setupMightyWebMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
			m.On("GetHint").Return(c.hint)
			result := p.HintOutput(m)
			var resObj controller.MightyWebOutput
			assert.NoError(t, json.Unmarshal([]byte(result), &resObj), c.name)
			assert.NotNil(t, resObj.Hint, c.name)
			c.check(t, resObj.Hint)
		}
	})

	t.Run("no hint returns base output with nil hint", func(t *testing.T) {
		m, _ := setupMightyWebMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHint")
		m.On("GetHint").Return((*domain.MightyHint)(nil))
		result := p.HintOutput(m)
		var resObj controller.MightyWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Nil(t, resObj.Hint)
	})
}

func TestMightyWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MightyWebPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockMightyGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played SPADE 5"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"actionType":"play"`)
		assert.Contains(t, result, `"turnNumber":1`)
	})

	t.Run("game not ended yields empty entries", func(t *testing.T) {
		m := new(interfaces.MockMightyGame)
		m.On("GetGameEndFlag").Return(false)
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, `"entries":[]`)
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestMightyWebPresenterOutputCarriesTheHint(t *testing.T) {
	idx := 0
	mgg, _ := setupMightyWebMockWithPlayers()
	mgg.ExpectedCalls = removeMockCall(mgg.ExpectedCalls, "GetHint")
	mgg.On("GetHint").Return(&domain.MightyHint{CardIndex: &idx})

	result := new(presenter.MightyWebPresenter).Output(mgg, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	// **Output は「頼んだヒント」の印を付けない。**付けると CLI が毎回 HINT 行を出す。
	assert.NotContains(t, result, "mighty.hintRequested")
}

// **HintOutput は「頼んだヒント」だと分かる印を付ける。**
func TestMightyWebPresenterHintOutputMarksTheRequest(t *testing.T) {
	idx := 0
	mgg, _ := setupMightyWebMockWithPlayers()
	mgg.ExpectedCalls = removeMockCall(mgg.ExpectedCalls, "GetHint")
	mgg.On("GetHint").Return(&domain.MightyHint{CardIndex: &idx})
	assert.Contains(t, new(presenter.MightyWebPresenter).HintOutput(mgg), "mighty.hintRequested")

	none, _ := setupMightyWebMockWithPlayers()
	none.ExpectedCalls = removeMockCall(none.ExpectedCalls, "GetHint")
	none.On("GetHint").Return((*domain.MightyHint)(nil))
	assert.Contains(t, new(presenter.MightyWebPresenter).HintOutput(none), "mighty.noHint")
}
