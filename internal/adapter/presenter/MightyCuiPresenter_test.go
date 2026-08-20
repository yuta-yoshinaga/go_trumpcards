//go:build test

package presenter_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// setupMightyCuiMock returns a MockMightyGame with safe defaults so the
// presenter can render without touching unexpected accessors.
func setupMightyCuiMock() *interfaces.MockMightyGame {
	m := new(interfaces.MockMightyGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(0)
	m.On("GetWinningBidNoTrump").Return(false)
	m.On("GetPartnerCard").Return((*domain.Card)(nil))
	m.On("GetPartnerRevealed").Return(false)
	m.On("GetHighestBid").Return(0)
	m.On("GetCurrentTrick").Return([]*domain.MightyTrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.MightyPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(domain.MightyWinnerUndecided)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultMightyConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetKitty").Return([]*domain.Card{})
	return m
}

func makeMightyPlayers() []*domain.MightyPlayer {
	return []*domain.MightyPlayer{
		domain.NewMightyPlayer(true),
		domain.NewMightyPlayer(false),
		domain.NewMightyPlayer(false),
		domain.NewMightyPlayer(false),
		domain.NewMightyPlayer(false),
	}
}

func setupMightyCuiMockWithPlayers() (*interfaces.MockMightyGame, []*domain.MightyPlayer) {
	m := setupMightyCuiMock()
	players := makeMightyPlayers()
	m.On("GetPlayerCnt").Return(5)
	for i, p := range players {
		m.On("GetPlayer", i).Return(p)
	}
	return m, players
}

func TestMightyCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.MightyCuiPresenter)

	t.Run("base output includes player and hand", func(t *testing.T) {
		m, players := setupMightyCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		result := p.Output(m, nil)
		// Header block always rendered.
		assert.Contains(t, result, "==========")
		// Human cards are listed with indexes.
		assert.Contains(t, result, "[0]SPADE 1")
		assert.Contains(t, result, "[1]HEART 5")
		// Human name "あなた" (cui_common.json) appears for player 0.
		assert.Contains(t, result, "あなた")
		// CPU player name appears.
		assert.Contains(t, result, "CPU 1")
	})

	t.Run("trump suit glyph rendered when set", func(t *testing.T) {
		m, _ := setupMightyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
		result := p.Output(m, nil)
		assert.Contains(t, result, "♠")
	})

	t.Run("no-trump message when winning bid is NT and no trump suit", func(t *testing.T) {
		m, _ := setupMightyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinningBidNoTrump")
		m.On("GetWinningBidNoTrump").Return(true)
		result := p.Output(m, nil)
		// Unregistered key → renders the key itself.
		assert.Contains(t, result, "ノートランプ")
	})

	t.Run("partner card with hidden status", func(t *testing.T) {
		m, _ := setupMightyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPartnerCard")
		m.On("GetPartnerCard").Return(domain.NewCard(domain.CardDesignHeart, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "HEART 13")
		assert.Contains(t, result, "(非公開)")
	})

	t.Run("partner card with revealed status", func(t *testing.T) {
		m, _ := setupMightyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPartnerCard")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPartnerRevealed")
		m.On("GetPartnerCard").Return(domain.NewCard(domain.CardDesignDiamond, 1, false))
		m.On("GetPartnerRevealed").Return(true)
		result := p.Output(m, nil)
		assert.Contains(t, result, "DIAMOND 1")
		assert.Contains(t, result, "(公開済み)")
	})

	t.Run("highest bid shown when >0", func(t *testing.T) {
		m, _ := setupMightyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHighestBid")
		m.On("GetHighestBid").Return(17)
		result := p.Output(m, nil)
		assert.Contains(t, result, "17")
	})

	t.Run("declarer badge rendered for declarer player", func(t *testing.T) {
		m, players := setupMightyCuiMockWithPlayers()
		players[0].SetIsDeclarer(true)
		result := p.Output(m, nil)
		assert.Contains(t, result, "[宣言者]")
	})

	t.Run("partner badge rendered when revealed", func(t *testing.T) {
		m, players := setupMightyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPartnerRevealed")
		m.On("GetPartnerRevealed").Return(true)
		players[1].SetIsPartner(true)
		result := p.Output(m, nil)
		assert.Contains(t, result, "[パートナー]")
	})

	t.Run("partner badge hidden when not revealed", func(t *testing.T) {
		m, players := setupMightyCuiMockWithPlayers()
		players[1].SetIsPartner(true)
		result := p.Output(m, nil)
		assert.NotContains(t, result, "[パートナー]")
	})

	t.Run("trick line rendered when trick has cards", func(t *testing.T) {
		m, _ := setupMightyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.MightyTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignClover, 3, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignClover, 7, false)},
		}
		m.On("GetCurrentTrick").Return(trick)
		result := p.Output(m, nil)
		assert.Contains(t, result, "CLOVER 3")
		assert.Contains(t, result, "CLOVER 7")
	})

	t.Run("joker trick card renders as 'Joker'", func(t *testing.T) {
		m, _ := setupMightyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		m.On("GetCurrentTrick").Return([]*domain.MightyTrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignJoker, 1, false)},
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "Joker")
	})

	t.Run("error message rendered through cuiErrorBlock", func(t *testing.T) {
		m, _ := setupMightyCuiMockWithPlayers()
		result := p.Output(m, errors.New("invalid card index"))
		assert.Contains(t, result, "invalid card index")
	})

	t.Run("game end declarer wins", func(t *testing.T) {
		m, _ := setupMightyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(domain.MightyWinnerDeclarer)
		result := p.Output(m, nil)
		assert.Contains(t, result, "与党")
	})

	t.Run("game end opposition wins", func(t *testing.T) {
		m, _ := setupMightyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(domain.MightyWinnerOpposition)
		result := p.Output(m, nil)
		assert.Contains(t, result, "野党")
	})

	t.Run("kitty phase lists kitty cards with hand indices", func(t *testing.T) {
		m, players := setupMightyCuiMockWithPlayers()
		// Declarer (player 0, human) hand — the kitty is merged in and sorted.
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))  // idx 0
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))  // idx 1
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 9, false)) // idx 2
		kitty := []*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 5, false),
			nil, // defensive: a nil kitty entry is skipped
			domain.NewCard(domain.CardDesignClover, 9, false),
		}
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MightyPhaseKittyExchange)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetKitty")
		m.On("GetKitty").Return(kitty)

		result := p.Output(m, nil)
		assert.Contains(t, result, "キティのカード")
		// The kitty cards are shown at their positions in the declarer's hand,
		// matching the indices the `e` (exchange) command consumes.
		assert.Contains(t, result, "[1]HEART 5")
		assert.Contains(t, result, "[2]CLOVER 9")
	})

	t.Run("kitty line omitted when declarer is missing", func(t *testing.T) {
		m, _ := setupMightyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MightyPhaseKittyExchange)
		// Declarer index points at an absent player → GetPlayer returns nil.
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDeclarerIdx")
		m.On("GetDeclarerIdx").Return(9)
		m.On("GetPlayer", 9).Return((*domain.MightyPlayer)(nil))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetKitty")
		m.On("GetKitty").Return([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})

		result := p.Output(m, nil)
		// No kitty line, but the phase prompt still renders.
		assert.NotContains(t, result, "キティのカード")
		assert.Contains(t, result, "場札")
	})

	t.Run("phase prompts", func(t *testing.T) {
		cases := []struct {
			phase  domain.MightyPhase
			prompt string
		}{
			{domain.MightyPhaseBid, "ビッドフェーズ"},
			{domain.MightyPhaseTrumpAndFriend, "切り札とパートナー"},
			{domain.MightyPhaseKittyExchange, "場札"},
			{domain.MightyPhasePlay, "プレイフェーズ"},
			{domain.MightyPhaseTrickEnd, "トリック終了"},
			{domain.MightyPhaseRoundEnd, "ラウンド終了"},
		}
		for _, c := range cases {
			m, _ := setupMightyCuiMockWithPlayers()
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
			m.On("GetPhase").Return(c.phase)
			result := p.Output(m, nil)
			assert.Contains(t, result, c.prompt, "phase %v", c.phase)
		}
	})
}

func TestMightyCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.MightyCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m := new(interfaces.MockMightyGame)
		m.On("GetHint").Return((*domain.MightyHint)(nil))
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントなし")
	})

	t.Run("bid hint (no-trump tag included)", func(t *testing.T) {
		bid := 16
		noTrump := true
		m := new(interfaces.MockMightyGame)
		m.On("GetHint").Return(&domain.MightyHint{
			Bid:        &bid,
			BidNoTrump: &noTrump,
			Reason:     "strategic_bid",
		})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒント: ビッド")
		assert.Contains(t, result, "16")
		assert.Contains(t, result, "ノートランプ")
	})

	t.Run("trump suit hint with known suit", func(t *testing.T) {
		suit := domain.CardDesignSpade
		m := new(interfaces.MockMightyGame)
		m.On("GetHint").Return(&domain.MightyHint{
			TrumpSuit: &suit,
			Reason:    "strategic_declare",
		})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒント: 切り札")
		assert.Contains(t, result, "♠")
	})

	t.Run("trump suit hint with no-trump (suit not in glyph map)", func(t *testing.T) {
		suit := domain.MightyTrumpNone
		m := new(interfaces.MockMightyGame)
		m.On("GetHint").Return(&domain.MightyHint{
			TrumpSuit: &suit,
			Reason:    "strategic_declare",
		})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒント: 切り札")
	})

	t.Run("discard hint", func(t *testing.T) {
		m := new(interfaces.MockMightyGame)
		m.On("GetHint").Return(&domain.MightyHint{
			DiscardIndices: []int{0, 2, 4},
			Reason:         "strategic_discard",
		})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒント:")
	})

	t.Run("card index hint", func(t *testing.T) {
		idx := 1
		m := new(interfaces.MockMightyGame)
		m.On("GetHint").Return(&domain.MightyHint{
			CardIndex: &idx,
			Reason:    "play_low",
		})
		player := domain.NewMightyPlayer(true)
		player.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		player.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
		m.On("GetPlayer", 0).Return(player)
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒント:")
	})

	t.Run("hint with only an unsupported reason and no field falls back to none", func(t *testing.T) {
		m := new(interfaces.MockMightyGame)
		m.On("GetHint").Return(&domain.MightyHint{Reason: "unknown"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントなし")
	})
}

func TestMightyCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MightyCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		m := new(interfaces.MockMightyGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played SPADE 5"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		m.On("GetPlayer", mock.Anything).Return(domain.NewMightyPlayer(true)).Maybe()
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "play")
		assert.Contains(t, result, "あなた", "棋譜の座席名が他の行と揃っていない")
	})

	t.Run("game not ended", func(t *testing.T) {
		m := new(interfaces.MockMightyGame)
		m.On("GetGameEndFlag").Return(false)
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
	})
}

// #5594: `nt` を付けると最低ビッドが上がるのに、CUI はコマンド構文しか出して
// おらず、付けてエラーになって初めて気づく形だった。
func TestMightyCuiPresenter_ExplainsTheNoTrumpExtraWhileBidding(t *testing.T) {
	i18n.SetLang("ja")
	build := func(extra int) string {
		m, _ := setupMightyCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetConfig")
		cfg := domain.DefaultMightyConfig()
		cfg.NoTrumpExtra = extra
		m.On("GetPhase").Return(domain.MightyPhaseBid)
		m.On("GetConfig").Return(cfg)
		return new(presenter.MightyCuiPresenter).Output(m, nil)
	}

	line := func(points int) string {
		return i18n.Tf("mighty.promptBidNoTrumpExtra", "points", strconv.Itoa(points))
	}

	assert.Contains(t, build(1), line(1))
	// **設定を変えれば表示も変わる** (受け入れ条件2)。訳文に数字を焼き込んでいない証拠。
	assert.Contains(t, build(3), line(3))
	assert.NotContains(t, build(3), line(1))
}

// ビッド以外のフェーズでは出さない。`nt` を打てない局面の説明は雑音になる。
func TestMightyCuiPresenter_HidesTheNoTrumpExtraOutsideTheBidPhase(t *testing.T) {
	i18n.SetLang("ja")
	m, _ := setupMightyCuiMockWithPlayers()
	m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
	m.On("GetPhase").Return(domain.MightyPhasePlay)

	out := new(presenter.MightyCuiPresenter).Output(m, nil)
	assert.NotContains(t, out, strings.SplitN(i18n.T("mighty.promptBidNoTrumpExtra"), "{{", 2)[0])
}
