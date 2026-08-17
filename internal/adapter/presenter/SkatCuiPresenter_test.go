//go:build test

package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupSkatCuiMock() *interfaces.MockSkatGame {
	m := new(interfaces.MockSkatGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(0)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetForehandIdx").Return(1)
	m.On("GetMiddlehandIdx").Return(2)
	m.On("GetRearhandIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(-1)
	m.On("GetCurrentBid").Return(0)
	m.On("GetActiveBidActorIdx").Return(2)
	m.On("GetGameType").Return(domain.SkatGameNone)
	m.On("GetTrumpSuit").Return(0)
	m.On("PickedSkat").Return(false)
	m.On("GetDeclarerCardPoints").Return(0)
	m.On("GetDefendersCardPoints").Return(0)
	m.On("GetWinnerSide").Return(domain.SkatWinnerUndecided)
	m.On("GetGameValue").Return(0)
	m.On("GetScoreBreakdown").Return((*domain.SkatScoreBreakdown)(nil)).Maybe()
	m.On("GetGameEndFlag").Return(false)
	m.On("GetLeadPlayerIdx").Return(-1)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetSkat").Return(([]*domain.Card)(nil))
	m.On("GetOriginalSkat").Return(([]*domain.Card)(nil))
	m.On("GetConfig").Return(domain.DefaultSkatConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetPlayerCnt").Return(3)
	for i := 0; i < 3; i++ {
		m.On("GetPlayer", i).Return(domain.NewSkatPlayer(i == 0))
	}
	m.On("GetPhase").Return(domain.SkatPhaseBid)
	return m
}

func TestSkatCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	// Output strings are localized; pin English for these assertions.
	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	p := new(presenter.SkatCuiPresenter)

	t.Run("bid phase", func(t *testing.T) {
		m := setupSkatCuiMock()
		out := p.Output(m, nil)
		assert.Contains(t, out, "Skat")
		assert.Contains(t, out, "Round: 1")
		assert.Contains(t, out, "Bidding")
	})

	t.Run("japanese locale renders translated phase text", func(t *testing.T) {
		i18n.SetLang("ja")
		defer i18n.SetLang("en")
		m := setupSkatCuiMock()
		out := p.Output(m, nil)
		assert.Contains(t, out, "ラウンド: 1")
		assert.Contains(t, out, "ビッド")
		assert.NotContains(t, out, "Round: 1")
		assert.NotContains(t, out, "Bidding:")
	})

	t.Run("error message rendered", func(t *testing.T) {
		m := setupSkatCuiMock()
		out := p.Output(m, errors.New("boom"))
		assert.Contains(t, out, "boom")
	})

	t.Run("game over", func(t *testing.T) {
		m := setupSkatCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		out := p.Output(m, nil)
		assert.Contains(t, out, "Game over")
	})

	t.Run("declared phase shows game label", func(t *testing.T) {
		m := setupSkatCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameType")
		m.On("GetGameType").Return(domain.SkatGameSuit)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentBid")
		m.On("GetCurrentBid").Return(18)
		out := p.Output(m, nil)
		assert.Contains(t, out, "Suit")
		assert.Contains(t, out, "trump=♠")
		assert.Contains(t, out, "Current bid: 18")
	})
}

func TestSkatCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	p := new(presenter.SkatCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m := setupSkatCuiMock()
		m.On("GetHint").Return((*domain.SkatHint)(nil))
		assert.Contains(t, p.HintOutput(m), "No hint")
	})

	t.Run("bid hint", func(t *testing.T) {
		m := setupSkatCuiMock()
		val := 1
		m.On("GetHint").Return(&domain.SkatHint{Bid: &val, Reason: "strategic_bid"})
		assert.Contains(t, p.HintOutput(m), "accept")
	})

	t.Run("pick skat hint", func(t *testing.T) {
		m := setupSkatCuiMock()
		yes := true
		m.On("GetHint").Return(&domain.SkatHint{PickSkat: &yes, Reason: "skat_pickup"})
		assert.Contains(t, p.HintOutput(m), "pick up")
	})

	t.Run("game choice hint", func(t *testing.T) {
		m := setupSkatCuiMock()
		gt := int(domain.SkatGameSuit)
		ts := domain.CardDesignSpade
		m.On("GetHint").Return(&domain.SkatHint{GameType: &gt, TrumpSuit: &ts, Reason: "game_choice"})
		assert.Contains(t, p.HintOutput(m), "Suit")
	})
}

func TestSkatCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SkatCuiPresenter)
	m := setupSkatCuiMock()
	assert.NotPanics(t, func() {
		_ = p.ActionLogOutput(m)
	})
}

// setupSkatCuiMockPhase replaces the GetPhase stub with the requested phase.
func setupSkatCuiMockPhase(m *interfaces.MockSkatGame, phase domain.SkatPhase) {
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
	m.On("GetPhase").Return(phase)
}

func TestSkatCuiPresenter_OutputPhases(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	p := new(presenter.SkatCuiPresenter)

	cases := []struct {
		name  string
		phase domain.SkatPhase
		want  string
	}{
		{"skat-pickup", domain.SkatPhaseSkatPickup, "Skat pickup"},
		{"discard", domain.SkatPhaseDiscard, "Discard 2 cards"},
		{"declaration", domain.SkatPhaseGameDeclaration, "Game declaration"},
		{"play", domain.SkatPhasePlay, "Turn:"},
		{"trick-end", domain.SkatPhaseTrickEnd, "Trick complete"},
		{"round-end", domain.SkatPhaseRoundEnd, "Round end"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := setupSkatCuiMock()
			setupSkatCuiMockPhase(m, c.phase)
			out := p.Output(m, nil)
			assert.Contains(t, out, c.want)
		})
	}
}

// TestSkatCuiPresenter_OutputDeclaredVariants exercises the trump-suit symbol
// branches and the Grand/Null game label paths.
func TestSkatCuiPresenter_OutputDeclaredVariants(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SkatCuiPresenter)

	suits := []struct {
		suit  int
		glyph string
	}{
		{domain.CardDesignClover, "♣"},
		{domain.CardDesignHeart, "♥"},
		{domain.CardDesignDiamond, "♦"},
	}
	for _, s := range suits {
		m := setupSkatCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameType")
		m.On("GetGameType").Return(domain.SkatGameSuit)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(s.suit)
		assert.Contains(t, p.Output(m, nil), s.glyph)
	}

	for _, gt := range []domain.SkatGameType{domain.SkatGameGrand, domain.SkatGameNull} {
		m := setupSkatCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameType")
		m.On("GetGameType").Return(gt)
		out := p.Output(m, nil)
		// Output should contain a non-trivial game-type label.
		assert.NotEmpty(t, out)
	}
}

// removeAllMockCalls strips every expected call matching the method name.
func removeAllMockCalls(m *interfaces.MockSkatGame, method string) {
	for {
		before := len(m.ExpectedCalls)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, method)
		if len(m.ExpectedCalls) == before {
			return
		}
	}
}

// TestSkatCuiPresenter_PlayerSummaryRoles exercises the declarer-flag and
// bid-display branches in skatPlayerStr.
func TestSkatCuiPresenter_PlayerSummaryRoles(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SkatCuiPresenter)

	m := setupSkatCuiMock()
	removeAllMockCalls(m, "GetPlayer")

	// Declarer player whose bid is 18 (numeric branch).
	pDecl := domain.NewSkatPlayer(false)
	pDecl.SetIsDeclarer(true)
	pDecl.SetBid(18)
	// Pass-bid player (bid==0).
	pPass := domain.NewSkatPlayer(false)
	pPass.SetBid(0)
	// Human with at least one card so the indexed-cards block runs.
	pHuman := domain.NewSkatPlayer(true)
	pHuman.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))

	m.On("GetPlayer", 0).Return(pHuman)
	m.On("GetPlayer", 1).Return(pDecl)
	m.On("GetPlayer", 2).Return(pPass)

	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	out := p.Output(m, nil)
	assert.Contains(t, out, "[Declarer]")
	assert.Contains(t, out, "bid=18")
	assert.Contains(t, out, "bid=pass")

	// The declarer role and pass label are localized under ja.
	i18n.SetLang("ja")
	outJa := p.Output(m, nil)
	assert.Contains(t, outJa, "[宣言者]")
	assert.Contains(t, outJa, "bid=パス")
	assert.NotContains(t, outJa, "[Declarer]")
}

// TestSkatCuiPresenter_HintOutputAllBranches covers every hint-render branch:
// discard, card-play, and the various reason translations.
func TestSkatCuiPresenter_HintOutputAllBranches(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	p := new(presenter.SkatCuiPresenter)

	t.Run("discard hint", func(t *testing.T) {
		m := setupSkatCuiMock()
		idx := 0
		human := domain.NewSkatPlayer(true)
		human.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		removeAllMockCalls(m, "GetPlayer")
		m.On("GetPlayer", 0).Return(human)
		for i := 1; i < 3; i++ {
			m.On("GetPlayer", i).Return(domain.NewSkatPlayer(false))
		}
		m.On("GetHint").Return(&domain.SkatHint{DiscardIndex: &idx, Reason: "discard_low"})
		assert.Contains(t, p.HintOutput(m), "discard")
	})

	t.Run("play card hint", func(t *testing.T) {
		m := setupSkatCuiMock()
		idx := 0
		human := domain.NewSkatPlayer(true)
		human.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		removeAllMockCalls(m, "GetPlayer")
		m.On("GetPlayer", 0).Return(human)
		for i := 1; i < 3; i++ {
			m.On("GetPlayer", i).Return(domain.NewSkatPlayer(false))
		}
		m.On("GetHint").Return(&domain.SkatHint{CardIndex: &idx, Reason: "best_play"})
		assert.Contains(t, p.HintOutput(m), "play")
	})

	t.Run("pick skat decline hint", func(t *testing.T) {
		m := setupSkatCuiMock()
		no := false
		m.On("GetHint").Return(&domain.SkatHint{PickSkat: &no, Reason: "skat_pickup"})
		assert.Contains(t, p.HintOutput(m), "decline")
	})

	t.Run("bid pass hint", func(t *testing.T) {
		m := setupSkatCuiMock()
		val := 0
		m.On("GetHint").Return(&domain.SkatHint{Bid: &val, Reason: "strategic_bid"})
		assert.Contains(t, p.HintOutput(m), "pass")
	})

	t.Run("grand game choice hint", func(t *testing.T) {
		m := setupSkatCuiMock()
		gt := int(domain.SkatGameGrand)
		m.On("GetHint").Return(&domain.SkatHint{GameType: &gt, Reason: "game_choice"})
		assert.Contains(t, p.HintOutput(m), "Grand")
	})

	t.Run("empty hint struct returns no-hint", func(t *testing.T) {
		m := setupSkatCuiMock()
		m.On("GetHint").Return(&domain.SkatHint{Reason: "best_play"})
		assert.Contains(t, p.HintOutput(m), "No hint")
	})
}

func TestSkatCuiPresenter_HintOutput_Localized(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.SkatCuiPresenter)

	t.Run("ja renders Japanese hint + reason", func(t *testing.T) {
		i18n.SetLang("ja")
		m := setupSkatCuiMock()
		val := 1
		m.On("GetHint").Return(&domain.SkatHint{Bid: &val, Reason: "strategic_bid"})
		out := p.HintOutput(m)
		assert.Contains(t, out, "アクセプト")   // choiceBidAccept
		assert.Contains(t, out, "戦略的なビッド") // hintReasonStrategicBid
		assert.NotContains(t, out, "strategic bid")
	})

	t.Run("ja renders Japanese no-hint", func(t *testing.T) {
		i18n.SetLang("ja")
		m := setupSkatCuiMock()
		m.On("GetHint").Return((*domain.SkatHint)(nil))
		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})

	t.Run("en renders English hint + reason", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")
		m := setupSkatCuiMock()
		val := 1
		m.On("GetHint").Return(&domain.SkatHint{Bid: &val, Reason: "strategic_bid"})
		out := p.HintOutput(m)
		assert.Contains(t, out, "accept")
		assert.Contains(t, out, "strategic bid")
	})

	i18n.SetLang("ja")
}

// **どこまで受けて安全かの目安を出す。**Web は常時表示しているのに CUI には
// 無く、オーバービッドの危険を測れなかった (#4905)。
func TestSkatCuiPresenter_ShowsTheHandBidEstimate(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	i18n.SetLang("en")
	defer i18n.SetLang("ja")

	build := func(currentBid int, hand []*domain.Card) *interfaces.MockSkatGame {
		m := setupSkatCuiMock()
		human := domain.NewSkatPlayer(true)
		for _, c := range hand {
			human.AddCard(c)
		}
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayer")
		m.On("GetPlayer", 0).Return(human)
		for i := 1; i < 3; i++ {
			m.On("GetPlayer", i).Return(domain.NewSkatPlayer(false))
		}
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentBid")
		m.On("GetCurrentBid").Return(currentBid)
		return m
	}
	p := new(presenter.SkatCuiPresenter)

	// ♣J のみ = with 1 → グランドで (1+1)×24 = 48。
	hand := []*domain.Card{domain.NewCard(domain.CardDesignClover, 11, false)}
	out := p.Output(build(0, hand), nil)
	assert.Contains(t, out, "Hand estimate: up to 48")
	assert.NotContains(t, out, "above your hand estimate")

	// 目安ちょうどでは警告しない。超えたときだけ。
	assert.NotContains(t, p.Output(build(48, hand), nil), "above your hand estimate")
	assert.Contains(t, p.Output(build(59, hand), nil), "above your hand estimate")

	// 手札が無い局面では出さない。
	assert.NotContains(t, p.Output(build(0, nil), nil), "Hand estimate")
}

// #5561: ラウンド終了行は最終値しか出さず、「なぜこの点数なのか」— とくに
// マタドール — を説明していなかった。
func TestSkatCuiPresenter_Output_ScoreBreakdown(t *testing.T) {
	build := func(bd *domain.SkatScoreBreakdown) string {
		m := setupSkatCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetScoreBreakdown")
		m.On("GetScoreBreakdown").Return(bd)
		setupSkatCuiMockPhase(m, domain.SkatPhaseRoundEnd)
		return new(presenter.SkatCuiPresenter).Output(m, nil)
	}

	out := build(&domain.SkatScoreBreakdown{Base: 11, Matadors: 2, Multiplier: 4, Hand: true, Value: 44})
	assert.Contains(t, out, i18n.Tf("skat.scoreBreakdownLine",
		"base", "11", "matadors", "2", "multiplier", "4",
		"bonuses", " ("+i18n.T("skat.bonusHand")+")"))

	// **ヌル契約には乗数の概念が無い。**内訳を出すと嘘になる。
	assert.NotContains(t, build(&domain.SkatScoreBreakdown{Base: 23, Multiplier: 1, Value: 23, Null: true}),
		strings.SplitN(i18n.T("skat.scoreBreakdownLine"), "{{", 2)[0])

	// 内訳が無いラウンド (未計算) でも落ちない。
	assert.NotContains(t, build(nil), strings.SplitN(i18n.T("skat.scoreBreakdownLine"), "{{", 2)[0])
}

// 各ボーナスが独立して並ぶこと。1 本にまとめると、フラグを 1 つ落とす変異が
// 別のフラグの出力に隠れる。出力面から見るので、内部関数は名指ししない。
func TestSkatCuiPresenter_ScoreBreakdownListsEachBonusIndependently(t *testing.T) {
	build := func(bd *domain.SkatScoreBreakdown) string {
		m := setupSkatCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetScoreBreakdown")
		m.On("GetScoreBreakdown").Return(bd)
		setupSkatCuiMockPhase(m, domain.SkatPhaseRoundEnd)
		return new(presenter.SkatCuiPresenter).Output(m, nil)
	}
	line := func(bonuses string) string {
		return i18n.Tf("skat.scoreBreakdownLine",
			"base", "11", "matadors", "2", "multiplier", "4", "bonuses", bonuses)
	}

	for _, tc := range []struct {
		name string
		bd   domain.SkatScoreBreakdown
		key  string
	}{
		{"hand", domain.SkatScoreBreakdown{Hand: true}, "skat.bonusHand"},
		{"schneider", domain.SkatScoreBreakdown{Schneider: true}, "skat.bonusSchneider"},
		{"schwarz", domain.SkatScoreBreakdown{Schwarz: true}, "skat.bonusSchwarz"},
		{"doubled", domain.SkatScoreBreakdown{Doubled: true}, "skat.bonusDoubled"},
		{"overbid", domain.SkatScoreBreakdown{Overbid: true}, "skat.bonusOverbid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bd := tc.bd
			bd.Base, bd.Matadors, bd.Multiplier, bd.Value = 11, 2, 4, 44
			// **完全一致で見る。**「含む」だけだと、全ボーナスを常に並べる実装でも通る。
			assert.Contains(t, build(&bd), line(" ("+i18n.T(tc.key)+")"))
		})
	}

	// 何も付いていなければ丸括弧ごと消えること。空文字を返さないと "… ()" と出る。
	assert.Contains(t, build(&domain.SkatScoreBreakdown{Base: 11, Matadors: 2, Multiplier: 4, Value: 44}), line(""))

	// 複数付いたらカンマ区切りで、順序は固定 (同じ局面なら毎回同じ行)。
	multi := domain.SkatScoreBreakdown{Base: 11, Matadors: 2, Multiplier: 4, Value: 44, Hand: true, Schwarz: true, Overbid: true}
	assert.Contains(t, build(&multi), line(" ("+i18n.T("skat.bonusHand")+", "+i18n.T("skat.bonusSchwarz")+", "+i18n.T("skat.bonusOverbid")+")"))
}
