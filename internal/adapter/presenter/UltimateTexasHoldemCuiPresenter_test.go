package presenter

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupUltimateTexasHoldemCuiMockDefaults(m *interfaces.MockUltimateTexasHoldemGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.UltimateTexasHoldemPhaseBet).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(0).Maybe()
	m.On("GetBlindBet").Return(0).Maybe()
	m.On("GetTripsBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetFolded").Return(false).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetBlindPayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetTripsPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetPlayerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerBest").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestUltimateTexasHoldemCuiPresenter_Output_BetPhase(t *testing.T) {
	p := new(UltimateTexasHoldemCuiPresenter)
	m := new(interfaces.MockUltimateTexasHoldemGame)
	setupUltimateTexasHoldemCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "1000")
	assert.NotEmpty(t, result)
	// Before the ante is placed (bet phase), no live bet summary is shown.
	assert.NotContains(t, result, strings.Split(i18n.T("ultimatetexasholdem.anteLine"), "{{")[0])
}

func TestUltimateTexasHoldemCuiPresenter_Output_PreFlopPhase_DealerHidden(t *testing.T) {
	p := new(UltimateTexasHoldemCuiPresenter)
	m := new(interfaces.MockUltimateTexasHoldemGame)
	m.On("GetChips").Return(800).Maybe()
	m.On("GetPhase").Return(domain.UltimateTexasHoldemPhasePreFlop).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
	}).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBlindBet").Return(100).Maybe()
	m.On("GetTripsBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetFolded").Return(false).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetBlindPayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetTripsPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "??", "dealer hand should be hidden in pre-flop")
	// During play the live bet summary is shown (ante placed, game not ended).
	assert.Contains(t, result, strings.Split(i18n.T("ultimatetexasholdem.anteLine"), "{{")[0])
}

func TestUltimateTexasHoldemCuiPresenter_Output_RiverPhase_TripsAndPlayBets(t *testing.T) {
	p := new(UltimateTexasHoldemCuiPresenter)
	m := new(interfaces.MockUltimateTexasHoldemGame)
	setupUltimateTexasHoldemCuiMockDefaults(m)
	m.ExpectedCalls = nil
	m.On("GetChips").Return(500).Maybe()
	m.On("GetPhase").Return(domain.UltimateTexasHoldemPhaseRiver).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBlindBet").Return(100).Maybe()
	m.On("GetTripsBet").Return(20).Maybe() // exercises the trips branch
	m.On("GetPlayBet").Return(100).Maybe() // exercises the play-bet branch
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, strings.Split(i18n.T("ultimatetexasholdem.tripsLine"), "{{")[0])
	assert.Contains(t, result, strings.Split(i18n.T("ultimatetexasholdem.playBetLine"), "{{")[0])
}

func TestUltimateTexasHoldemCuiPresenter_Output_EndPhase_PlayerWins(t *testing.T) {
	p := new(UltimateTexasHoldemCuiPresenter)
	m := new(interfaces.MockUltimateTexasHoldemGame)
	m.On("GetChips").Return(1500).Maybe()
	m.On("GetPhase").Return(domain.UltimateTexasHoldemPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 13, false),
		domain.NewCard(domain.CardDesignClover, 13, false),
	}).Maybe()
	m.On("GetDealerHand").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 7, false),
	}).Maybe()
	m.On("GetCommunity").Return([]*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 11, false),
		domain.NewCard(domain.CardDesignSpade, 12, false),
		domain.NewCard(domain.CardDesignClover, 3, false),
		domain.NewCard(domain.CardDesignHeart, 8, false),
		domain.NewCard(domain.CardDesignDiamond, 2, false),
	}).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBlindBet").Return(100).Maybe()
	m.On("GetTripsBet").Return(20).Maybe()
	m.On("GetPlayBet").Return(400).Maybe()
	m.On("GetFolded").Return(false).Maybe()
	m.On("GetResult").Return(domain.GameResultWin).Maybe()
	m.On("GetDealerQualified").Return(true).Maybe()
	m.On("GetAntePayout").Return(200).Maybe()
	m.On("GetBlindPayout").Return(100).Maybe()
	m.On("GetPlayPayout").Return(800).Maybe()
	m.On("GetTripsPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(1100).Maybe()
	m.On("GetPlayerHandRank").Return(domain.PokerHandOnePair).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()

	result := p.Output(m, nil)
	assert.Contains(t, result, "1100")
	// The player's One Pair rank is localized (ja by default), not raw English.
	assert.Contains(t, result, "ワンペア")
	assert.NotContains(t, result, "One Pair")

	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	assert.Contains(t, p.Output(m, nil), "One Pair")
}

func TestUltimateTexasHoldemCuiPresenter_Output_EndPhase_Folded(t *testing.T) {
	p := new(UltimateTexasHoldemCuiPresenter)
	m := new(interfaces.MockUltimateTexasHoldemGame)
	m.On("GetChips").Return(800).Maybe()
	m.On("GetPhase").Return(domain.UltimateTexasHoldemPhaseEnd).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetDealerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunity").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetAnteBet").Return(100).Maybe()
	m.On("GetBlindBet").Return(100).Maybe()
	m.On("GetTripsBet").Return(0).Maybe()
	m.On("GetPlayBet").Return(0).Maybe()
	m.On("GetFolded").Return(true).Maybe()
	m.On("GetResult").Return(domain.GameResultLose).Maybe()
	m.On("GetDealerQualified").Return(false).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetBlindPayout").Return(0).Maybe()
	m.On("GetPlayPayout").Return(0).Maybe()
	m.On("GetTripsPayout").Return(0).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetPlayerHandRank").Return(0).Maybe()
	m.On("GetDealerHandRank").Return(0).Maybe()

	result := p.Output(m, nil)
	assert.NotEmpty(t, result)
}

func TestUltimateTexasHoldemCuiPresenter_Output_ErrorRendered(t *testing.T) {
	p := new(UltimateTexasHoldemCuiPresenter)
	m := new(interfaces.MockUltimateTexasHoldemGame)
	setupUltimateTexasHoldemCuiMockDefaults(m)

	result := p.Output(m, errors.New("boom"))
	assert.Contains(t, result, "boom")
}

func TestUltimateTexasHoldemCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(UltimateTexasHoldemCuiPresenter)
	m := new(interfaces.MockUltimateTexasHoldemGame)
	m.On("GetGameEndFlag").Return(true).Maybe()
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "bet", Detail: "ante=100"},
	}).Maybe()

	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "bet")
}

// **CUI に 4x/3x/2x/1x/check/fold を選ぶ材料が何も無かった (#4709)。**
func TestUltimateTexasHoldemCuiPresenter_HintOutput(t *testing.T) {
	p := new(UltimateTexasHoldemCuiPresenter)
	game := func(rec string) *interfaces.MockUltimateTexasHoldemGame {
		m := new(interfaces.MockUltimateTexasHoldemGame)
		m.On("RecommendPlay").Return(rec)
		return m
	}

	// **倍率ごとに違う文言。**同じ文なら「レイズしろ」までしか伝わらない。
	t.Run("each multiplier gets its own line", func(t *testing.T) {
		seen := map[string]bool{}
		for _, rec := range []string{
			domain.UTHRecommendPlay4x, domain.UTHRecommendPlay3x,
			domain.UTHRecommendPlay2x, domain.UTHRecommendPlay1x,
			domain.UTHRecommendCheck, domain.UTHRecommendFold,
		} {
			out := p.HintOutput(game(rec))
			assert.False(t, seen[out], "%s の文言が他と重複している", rec)
			seen[out] = true
			assert.NotContains(t, out, i18n.T("ultimatetexasholdem.hintNone"))
		}
		assert.Len(t, seen, 6)
	})

	t.Run("names the 4x multiplier", func(t *testing.T) {
		assert.Contains(t, p.HintOutput(game(domain.UTHRecommendPlay4x)), "4倍")
	})

	t.Run("names the 2x multiplier", func(t *testing.T) {
		assert.Contains(t, p.HintOutput(game(domain.UTHRecommendPlay2x)), "2倍")
	})

	t.Run("says so when there is no decision to make", func(t *testing.T) {
		assert.Contains(t, p.HintOutput(game("")), i18n.T("ultimatetexasholdem.hintNone"))
	})
}

// #5589: トリップスは**フォールドしても評価される**特殊なサイドベットなのに、
// CUI には倍率を知る手段が無く、賭け額を決める材料が欠けていた。
func TestUltimateTexasHoldemCuiPresenter_ShowsThePayoutTable(t *testing.T) {
	i18n.SetLang("ja")
	build := func(phase int) string {
		m := new(interfaces.MockUltimateTexasHoldemGame)
		setupUltimateTexasHoldemCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(phase)
		return new(UltimateTexasHoldemCuiPresenter).Output(m, nil)
	}

	out := build(domain.UltimateTexasHoldemPhaseBet)
	assert.Contains(t, out, i18n.T("ultimatetexasholdem.payoutRefTitle"))
	// **倍率はドメインの定数から。**文言に書き写すと、配当を変えたとき嘘の表が残る。
	assert.Contains(t, out, i18n.Tf("ultimatetexasholdem.payoutRefBlindRoyalFlush",
		"rate", strconv.Itoa(domain.UltimateTexasHoldemBlindPayRoyalFlush)))
	assert.Contains(t, out, i18n.Tf("ultimatetexasholdem.payoutRefTripsThreeOfAKind",
		"rate", strconv.Itoa(domain.UltimateTexasHoldemTripsPayThreeOfAKind)))
	// フラッシュのブラインドだけ 3:2。整数倍として出すと配当が変わる。
	assert.Contains(t, out, i18n.Tf("ultimatetexasholdem.payoutRefBlindFlush",
		"rate", strconv.Itoa(domain.UltimateTexasHoldemBlindPayFlushNum)+":"+
			strconv.Itoa(domain.UltimateTexasHoldemBlindPayFlushDen)))

	// **ブラインドとトリップスは倍率が違う。**同じ表を 2 度出す実装を弾く。
	assert.NotEqual(t, domain.UltimateTexasHoldemBlindPayRoyalFlush,
		domain.UltimateTexasHoldemTripsPayRoyalFlush)
	assert.Contains(t, out, i18n.Tf("ultimatetexasholdem.payoutRefTripsRoyalFlush",
		"rate", strconv.Itoa(domain.UltimateTexasHoldemTripsPayRoyalFlush)))
}

// ベットフェーズ以外では出さない (受け入れ条件3)。賭けた後の卓に配当表を並べても、
// いま起きたことが読み取りにくくなるだけ。
func TestUltimateTexasHoldemCuiPresenter_HidesThePayoutTableAfterTheBet(t *testing.T) {
	i18n.SetLang("ja")
	for _, phase := range []int{
		domain.UltimateTexasHoldemPhasePreFlop,
		domain.UltimateTexasHoldemPhaseFlop,
		domain.UltimateTexasHoldemPhaseRiver,
		domain.UltimateTexasHoldemPhaseEnd,
	} {
		m := new(interfaces.MockUltimateTexasHoldemGame)
		setupUltimateTexasHoldemCuiMockDefaults(m)
		m.ExpectedCalls = filterCalls(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(phase)

		assert.NotContains(t, new(UltimateTexasHoldemCuiPresenter).Output(m, nil),
			i18n.T("ultimatetexasholdem.payoutRefTitle"), "phase %d", phase)
	}
}
