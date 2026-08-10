package presenter

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupMississippiStudCuiMockDefaults(m *interfaces.MockMississippiStudGame) {
	m.On("GetChips").Return(1000).Maybe()
	m.On("GetPhase").Return(domain.MississippiStudPhaseAnte).Maybe()
	m.On("GetPlayerHand").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunityCards").Return(([]*domain.Card)(nil)).Maybe()
	m.On("GetCommunityRevealed").Return([domain.MississippiStudCommunityCnt]bool{}).Maybe()
	m.On("GetGameEndFlag").Return(false).Maybe()
	m.On("GetAnteAmount").Return(0).Maybe()
	m.On("GetStreetMultipliers").Return([domain.MississippiStudStreetCnt]int{}).Maybe()
	m.On("GetFolded").Return(false).Maybe()
	m.On("GetTotalBet").Return(0).Maybe()
	m.On("GetResult").Return(domain.GameResult(0)).Maybe()
	m.On("GetHandRank").Return(0).Maybe()
	m.On("GetPayoutMultiplier").Return(0).Maybe()
	m.On("GetAntePayout").Return(0).Maybe()
	m.On("GetStreetPayouts").Return([domain.MississippiStudStreetCnt]int{}).Maybe()
	m.On("GetTotalPayout").Return(0).Maybe()
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil)).Maybe()
}

func TestMississippiStudCuiPresenter_Output_AntePhase(t *testing.T) {
	p := new(MississippiStudCuiPresenter)
	m := new(interfaces.MockMississippiStudGame)
	setupMississippiStudCuiMockDefaults(m)

	result := p.Output(m, nil)
	assert.Contains(t, result, "1000")
	assert.Contains(t, result, "ANTE")
	// No bet placed yet → no fold-loss note.
	assert.NotContains(t, result, strings.Split(i18n.T("mississippistud.foldLossLine"), "{{")[0])
}

func TestMississippiStudCuiPresenter_Output_Error(t *testing.T) {
	p := new(MississippiStudCuiPresenter)
	m := new(interfaces.MockMississippiStudGame)
	setupMississippiStudCuiMockDefaults(m)

	result := p.Output(m, errors.New("nope"))
	assert.Contains(t, result, "nope")
}

func TestMississippiStudCuiPresenter_Output_ThirdSt(t *testing.T) {
	p := new(MississippiStudCuiPresenter)
	m := new(interfaces.MockMississippiStudGame)

	hole := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 11, true),
		domain.NewCard(domain.CardDesignHeart, 11, true),
	}
	community := []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 2, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignSpade, 4, true),
	}
	m.On("GetChips").Return(900)
	m.On("GetPhase").Return(domain.MississippiStudPhaseThirdSt)
	m.On("GetPlayerHand").Return(hole)
	m.On("GetCommunityCards").Return(community)
	m.On("GetCommunityRevealed").Return([domain.MississippiStudCommunityCnt]bool{})
	m.On("GetGameEndFlag").Return(false)
	m.On("GetAnteAmount").Return(100)
	m.On("GetStreetMultipliers").Return([domain.MississippiStudStreetCnt]int{})
	m.On("GetFolded").Return(false)
	m.On("GetTotalBet").Return(100)
	m.On("GetResult").Return(domain.GameResult(0))
	m.On("GetHandRank").Return(0)
	m.On("GetPayoutMultiplier").Return(0)
	m.On("GetAntePayout").Return(0)
	m.On("GetStreetPayouts").Return([domain.MississippiStudStreetCnt]int{})
	m.On("GetTotalPayout").Return(0)

	result := p.Output(m, nil)
	assert.Contains(t, result, "3RD")
	assert.Contains(t, result, "??") // community masked
	// During the street, the accumulated bet and fold-loss note are shown.
	assert.Contains(t, result, strings.Split(i18n.T("mississippistud.foldLossLine"), "{{")[0])
}

func TestMississippiStudCuiPresenter_Output_Win(t *testing.T) {
	p := new(MississippiStudCuiPresenter)
	m := new(interfaces.MockMississippiStudGame)

	hole := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 11, true),
		domain.NewCard(domain.CardDesignHeart, 11, true),
	}
	community := []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 2, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignSpade, 4, true),
	}
	m.On("GetChips").Return(1600)
	m.On("GetPhase").Return(domain.MississippiStudPhaseEnd)
	m.On("GetPlayerHand").Return(hole)
	m.On("GetCommunityCards").Return(community)
	m.On("GetCommunityRevealed").Return([domain.MississippiStudCommunityCnt]bool{true, true, true})
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnteAmount").Return(100)
	m.On("GetStreetMultipliers").Return([domain.MississippiStudStreetCnt]int{3, 1, 1})
	m.On("GetFolded").Return(false)
	m.On("GetTotalBet").Return(600)
	m.On("GetResult").Return(domain.GameResultWin)
	m.On("GetHandRank").Return(domain.PokerHandOnePair)
	m.On("GetPayoutMultiplier").Return(domain.MississippiStudPayHighPair)
	m.On("GetAntePayout").Return(200)
	m.On("GetStreetPayouts").Return([domain.MississippiStudStreetCnt]int{600, 200, 200})
	m.On("GetTotalPayout").Return(1200)

	result := p.Output(m, nil)
	// The One Pair rank is localized (ja by default), not raw English.
	assert.Contains(t, result, "ワンペア")
	assert.Contains(t, result, "1200")

	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	assert.Contains(t, p.Output(m, nil), "One Pair")
}

func TestMississippiStudCuiPresenter_Output_Push(t *testing.T) {
	p := new(MississippiStudCuiPresenter)
	m := new(interfaces.MockMississippiStudGame)

	hole := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, true),
		domain.NewCard(domain.CardDesignHeart, 6, true),
	}
	community := []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 2, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignSpade, 4, true),
	}
	m.On("GetChips").Return(1000)
	m.On("GetPhase").Return(domain.MississippiStudPhaseEnd)
	m.On("GetPlayerHand").Return(hole)
	m.On("GetCommunityCards").Return(community)
	m.On("GetCommunityRevealed").Return([domain.MississippiStudCommunityCnt]bool{true, true, true})
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnteAmount").Return(100)
	m.On("GetStreetMultipliers").Return([domain.MississippiStudStreetCnt]int{1, 1, 1})
	m.On("GetFolded").Return(false)
	m.On("GetTotalBet").Return(400)
	m.On("GetResult").Return(domain.GameResultDraw)
	m.On("GetHandRank").Return(domain.PokerHandOnePair)
	m.On("GetPayoutMultiplier").Return(domain.MississippiStudPayPush)
	m.On("GetAntePayout").Return(100)
	m.On("GetStreetPayouts").Return([domain.MississippiStudStreetCnt]int{100, 100, 100})
	m.On("GetTotalPayout").Return(400)

	result := p.Output(m, nil)
	assert.Contains(t, result, "400")
}

func TestMississippiStudCuiPresenter_Output_Fold(t *testing.T) {
	p := new(MississippiStudCuiPresenter)
	m := new(interfaces.MockMississippiStudGame)

	hole := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 6, true),
		domain.NewCard(domain.CardDesignHeart, 6, true),
	}
	community := []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 2, true),
		domain.NewCard(domain.CardDesignClover, 3, true),
		domain.NewCard(domain.CardDesignSpade, 4, true),
	}
	m.On("GetChips").Return(900)
	m.On("GetPhase").Return(domain.MississippiStudPhaseEnd)
	m.On("GetPlayerHand").Return(hole)
	m.On("GetCommunityCards").Return(community)
	m.On("GetCommunityRevealed").Return([domain.MississippiStudCommunityCnt]bool{})
	m.On("GetGameEndFlag").Return(true)
	m.On("GetAnteAmount").Return(100)
	m.On("GetStreetMultipliers").Return([domain.MississippiStudStreetCnt]int{})
	m.On("GetFolded").Return(true)
	m.On("GetTotalBet").Return(100)
	m.On("GetResult").Return(domain.GameResultLose)
	m.On("GetHandRank").Return(0)
	m.On("GetPayoutMultiplier").Return(0)
	m.On("GetAntePayout").Return(0)
	m.On("GetStreetPayouts").Return([domain.MississippiStudStreetCnt]int{})
	m.On("GetTotalPayout").Return(0)

	result := p.Output(m, nil)
	// Fold path should NOT include a hand name (player did not play)
	assert.NotContains(t, result, "One Pair")
}

func TestMississippiStudCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(MississippiStudCuiPresenter)
	m := new(interfaces.MockMississippiStudGame)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetGameEndFlag").Return(true).Maybe()

	result := p.ActionLogOutput(m)
	assert.NotEmpty(t, result)
}

// **CUI には役評価も配当対象判定も推奨倍率も無かった (#4710)。**
func TestMississippiStudCuiPresenter_HintOutput(t *testing.T) {
	p := new(MississippiStudCuiPresenter)
	game := func(rec string, made *domain.MississippiStudMadeHand) *interfaces.MockMississippiStudGame {
		m := new(interfaces.MockMississippiStudGame)
		m.On("RecommendBet").Return(rec)
		m.On("GetCurrentMadeHand").Return(made).Maybe()
		return m
	}

	t.Run("names the hand and that it pays", func(t *testing.T) {
		out := p.HintOutput(game(domain.MSRecommendPlay3x,
			&domain.MississippiStudMadeHand{Rank: domain.PokerHandOnePair, PaytableEligible: true}))
		assert.Contains(t, out, "ワンペア")
		assert.Contains(t, out, "配当あり")
	})

	// **役名だけでは足りない。**2 のペアは「ワンペア」でも配当が付かない。
	t.Run("says when a made hand does not pay", func(t *testing.T) {
		out := p.HintOutput(game(domain.MSRecommendFold,
			&domain.MississippiStudMadeHand{Rank: domain.PokerHandOnePair, PaytableEligible: false}))
		assert.Contains(t, out, "配当なし")
		assert.NotContains(t, out, "配当あり")
	})

	t.Run("each recommendation gets its own line", func(t *testing.T) {
		seen := map[string]bool{}
		for _, rec := range []string{domain.MSRecommendPlay3x, domain.MSRecommendPlay1x, domain.MSRecommendFold} {
			out := p.HintOutput(game(rec, nil))
			assert.False(t, seen[out], "%s の文言が他と重複している", rec)
			seen[out] = true
		}
		assert.Len(t, seen, 3)
	})

	t.Run("omits the made-hand line when nothing is made", func(t *testing.T) {
		out := p.HintOutput(game(domain.MSRecommendPlay1x, nil))
		assert.NotContains(t, out, "現在の役")
	})

	t.Run("says so outside the betting streets", func(t *testing.T) {
		assert.Contains(t, p.HintOutput(game("", nil)), i18n.T("mississippistud.hintNone"))
	})
}
