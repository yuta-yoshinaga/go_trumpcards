//go:build test
// +build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestCuiCardStr(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	tests := []struct {
		name     string
		card     *domain.Card
		expected string
	}{
		{"nil card", nil, "??"},
		{"joker", domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false), "JOKER"},
		{"spade", domain.NewCard(domain.CardDesignSpade, 1, false), "SPADE 1"},
		{"clover", domain.NewCard(domain.CardDesignClover, 5, false), "CLOVER 5"},
		{"heart", domain.NewCard(domain.CardDesignHeart, 10, false), "HEART 10"},
		{"diamond", domain.NewCard(domain.CardDesignDiamond, 13, false), "DIAMOND 13"},
		{"unknown design", domain.NewCard(99, 1, false), "UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cuiCardStr(tt.card))
		})
	}
}

func TestCuiCardStrEmoji(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	tests := []struct {
		name     string
		card     *domain.Card
		expected string
	}{
		{"nil card", nil, "??"},
		{"joker", domain.NewCard(domain.CardDesignJoker, 0, false), "🃏0"},
		{"spade", domain.NewCard(domain.CardDesignSpade, 1, false), "♠1"},
		{"clover", domain.NewCard(domain.CardDesignClover, 5, false), "♣5"},
		{"heart", domain.NewCard(domain.CardDesignHeart, 10, false), "♥10"},
		{"diamond", domain.NewCard(domain.CardDesignDiamond, 13, false), "♦13"},
		{"negative design falls back to joker", domain.NewCard(-1, 7, false), "🃏7"},
		{"out-of-range design falls back to joker", domain.NewCard(99, 3, false), "🃏3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cuiCardStrEmoji(tt.card))
		})
	}
}

func TestCuiSuitName(t *testing.T) {
	tests := []struct {
		name     string
		suit     int
		expected string
	}{
		{"spade", domain.CardDesignSpade, "SPADE"},
		{"clover", domain.CardDesignClover, "CLOVER"},
		{"heart", domain.CardDesignHeart, "HEART"},
		{"diamond", domain.CardDesignDiamond, "DIAMOND"},
		{"unknown", 999, "UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cuiSuitName(tt.suit))
		})
	}
}

func TestCuiRankLabel(t *testing.T) {
	tests := []struct {
		rank     int
		expected string
	}{
		{1, "A"},
		{7, "7"},
		{10, "10"},
		{11, "J"},
		{12, "Q"},
		{13, "K"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, cuiRankLabel(tt.rank))
	}
}

func TestCuiPokerHandName(t *testing.T) {
	orig := i18n.Lang()
	i18n.SetLang("ja")
	defer i18n.SetLang(orig)

	assert.Equal(t, "ハイカード", cuiPokerHandName(0))
	assert.Equal(t, "ロイヤルフラッシュ", cuiPokerHandName(9))
	assert.Equal(t, "ファイブカード", cuiPokerHandName(10))
	// Out-of-range ranks fall back to an empty string.
	assert.Equal(t, "", cuiPokerHandName(-1))
	assert.Equal(t, "", cuiPokerHandName(999))

	i18n.SetLang("en")
	assert.Equal(t, "Royal Flush", cuiPokerHandName(9))
}

// mockCuiPlayer implements the cuiPlayer interface for testing.
type mockCuiPlayer struct {
	isHuman bool
}

func (m *mockCuiPlayer) GetIsHuman() bool {
	return m.isHuman
}

func TestCuiPlayerName(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	tests := []struct {
		name     string
		player   *mockCuiPlayer
		idx      int
		expected string
	}{
		{"nil player", nil, 0, "UNKNOWN"},
		{"human player", &mockCuiPlayer{isHuman: true}, 0, "あなた"},
		{"cpu player idx 1", &mockCuiPlayer{isHuman: false}, 1, "CPU 1"},
		{"cpu player idx 3", &mockCuiPlayer{isHuman: false}, 3, "CPU 3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cuiPlayerName(tt.player, tt.idx))
		})
	}
}

// mockCardList implements cuiCardList for testing.
type mockCardList struct {
	cards []*domain.Card
}

func (m *mockCardList) GetCardsSize() int            { return len(m.cards) }
func (m *mockCardList) GetCard(idx int) *domain.Card { return m.cards[idx] }

func TestCuiCardListStr(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	tests := []struct {
		name     string
		cards    []*domain.Card
		expected string
	}{
		{"empty hand", []*domain.Card{}, ""},
		{"single card", []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}, "SPADE 1"},
		{
			"multiple cards",
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignHeart, 5, false),
			},
			"SPADE 1,HEART 5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cuiCardListStr(&mockCardList{cards: tt.cards}))
		})
	}
}

func TestCuiIndexedCardListStr(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	tests := []struct {
		name     string
		cards    []*domain.Card
		expected string
	}{
		{"empty hand", []*domain.Card{}, ""},
		{"single card", []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}, "[0]SPADE 1"},
		{
			"multiple cards",
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignHeart, 5, false),
			},
			"[0]SPADE 1  [1]HEART 5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cuiIndexedCardListStr(&mockCardList{cards: tt.cards}))
		})
	}
}

func TestCuiCardListStrEmoji(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	tests := []struct {
		name     string
		cards    []*domain.Card
		expected string
	}{
		{"empty hand", []*domain.Card{}, ""},
		{"single card", []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}, "♠1"},
		{
			"multiple cards",
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignHeart, 5, false),
			},
			"♠1  ♥5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cuiCardListStrEmoji(&mockCardList{cards: tt.cards}))
		})
	}
}

func TestCuiIndexedCardListStrEmoji(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	tests := []struct {
		name     string
		cards    []*domain.Card
		expected string
	}{
		{"empty hand", []*domain.Card{}, ""},
		{"single card", []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}, "[0]♠1"},
		{
			"multiple cards",
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignHeart, 5, false),
			},
			"[0]♠1  [1]♥5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cuiIndexedCardListStrEmoji(&mockCardList{cards: tt.cards}))
		})
	}
}

func TestCuiCardSliceStr(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	tests := []struct {
		name     string
		cards    []*domain.Card
		expected string
	}{
		{"empty slice", []*domain.Card{}, ""},
		{"single card", []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}, "SPADE 1"},
		{
			"multiple cards",
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignHeart, 5, false),
			},
			"SPADE 1, HEART 5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cuiCardSliceStr(tt.cards))
		})
	}
}

func TestCuiCardSliceStrEmoji(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	tests := []struct {
		name     string
		cards    []*domain.Card
		expected string
	}{
		{"empty slice", []*domain.Card{}, ""},
		{"single card", []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}, "♠1"},
		{
			"multiple cards",
			[]*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignHeart, 5, false),
			},
			"♠1  ♥5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cuiCardSliceStrEmoji(tt.cards))
		})
	}
}

func TestFormatCardList(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	hand := &mockCardList{cards: []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
	}}

	tests := []struct {
		name     string
		fmtCard  cardFormatter
		sep      string
		indexed  bool
		expected string
	}{
		{"text no-index comma", cuiCardStr, ",", false, "SPADE 1,HEART 5"},
		{"text indexed double-space", cuiCardStr, "  ", true, "[0]SPADE 1  [1]HEART 5"},
		{"emoji no-index", cuiCardStrEmoji, "  ", false, "♠1  ♥5"},
		{"emoji indexed", cuiCardStrEmoji, "  ", true, "[0]♠1  [1]♥5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatCardList(hand, tt.fmtCard, tt.sep, tt.indexed))
		})
	}
}

func TestFormatCardList_Empty(t *testing.T) {
	hand := &mockCardList{cards: []*domain.Card{}}
	assert.Equal(t, "", formatCardList(hand, cuiCardStr, ",", false))
}

func TestFormatCardSlice(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	cards := []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 1, false),
		domain.NewCard(domain.CardDesignHeart, 5, false),
	}

	tests := []struct {
		name     string
		fmtCard  cardFormatter
		sep      string
		expected string
	}{
		{"text comma-space", cuiCardStr, ", ", "SPADE 1, HEART 5"},
		{"emoji double-space", cuiCardStrEmoji, "  ", "♠1  ♥5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatCardSlice(cards, tt.fmtCard, tt.sep))
		})
	}
}

func TestFormatCardSlice_Empty(t *testing.T) {
	assert.Equal(t, "", formatCardSlice([]*domain.Card{}, cuiCardStr, ", "))
}

func TestCuiBettingActionName(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	tests := []struct {
		name     string
		action   int
		expected string
	}{
		{"fold", domain.PokerActionFold, "フォールド"},
		{"check", domain.PokerActionCheck, "チェック"},
		{"call", domain.PokerActionCall, "コール"},
		{"bet", domain.PokerActionBet, "ベット"},
		{"raise", domain.PokerActionRaise, "レイズ"},
		{"all-in", domain.PokerActionAllIn, "オールイン"},
		{"unknown", 999, "不明"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cuiBettingActionName(tt.action))
		})
	}
}

// TestCuiBettingActionName_English verifies issue #1699 Phase 1: betting
// action names follow the active locale rather than always rendering as
// hardcoded Japanese.
func TestCuiBettingActionName_English(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	i18n.SetLang("en")
	defer i18n.SetLang("ja") // restore default for other tests

	tests := []struct {
		action   int
		expected string
	}{
		{domain.PokerActionFold, "Fold"},
		{domain.PokerActionCheck, "Check"},
		{domain.PokerActionCall, "Call"},
		{domain.PokerActionBet, "Bet"},
		{domain.PokerActionRaise, "Raise"},
		{domain.PokerActionAllIn, "All-in"},
		{999, "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, cuiBettingActionName(tt.action))
		})
	}
}

// TestCuiPlayerName_English verifies issue #1699 Phase 1: player display
// names follow the active locale ("You" / "CPU N" instead of always
// "あなた" / "CPU N").
func TestCuiPlayerName_English(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	i18n.SetLang("en")
	defer i18n.SetLang("ja")

	assert.Equal(t, "You", cuiPlayerName(&mockCuiPlayer{isHuman: true}, 0))
	assert.Equal(t, "CPU 2", cuiPlayerName(&mockCuiPlayer{isHuman: false}, 2))
	assert.Equal(t, "UNKNOWN", cuiPlayerName[*mockCuiPlayer](nil, 0))
}

func TestIsRedSuit(t *testing.T) {
	assert.True(t, isRedSuit(domain.CardDesignHeart))
	assert.True(t, isRedSuit(domain.CardDesignDiamond))
	assert.False(t, isRedSuit(domain.CardDesignSpade))
	assert.False(t, isRedSuit(domain.CardDesignClover))
	assert.False(t, isRedSuit(domain.CardDesignJoker))
	assert.False(t, isRedSuit(99))
}

func TestCuiCardStrColor(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(false)
	defer color.SetNoColor(origNoColor)

	// Red suits should have ANSI red codes
	heartCard := domain.NewCard(domain.CardDesignHeart, 10, false)
	got := cuiCardStr(heartCard)
	assert.Contains(t, got, "\033[31m")
	assert.Contains(t, got, "\033[0m")
	assert.Contains(t, got, "HEART 10")

	diamondCard := domain.NewCard(domain.CardDesignDiamond, 5, false)
	got = cuiCardStr(diamondCard)
	assert.Contains(t, got, "\033[31m")

	// Black suits should NOT have ANSI red codes
	spadeCard := domain.NewCard(domain.CardDesignSpade, 1, false)
	got = cuiCardStr(spadeCard)
	assert.NotContains(t, got, "\033[31m")
	assert.Equal(t, "SPADE 1", got)

	// Joker should not be colored
	jokerCard := domain.NewCard(domain.CardDesignJoker, 0, false)
	got = cuiCardStr(jokerCard)
	assert.NotContains(t, got, "\033[31m")
	assert.Equal(t, "JOKER", got)

	// nil card should not be colored
	got = cuiCardStr(nil)
	assert.Equal(t, "??", got)

	// Unknown design should not be colored
	unknownCard := domain.NewCard(99, 1, false)
	got = cuiCardStr(unknownCard)
	assert.Equal(t, "UNKNOWN", got)
}

func TestCuiCardStrEmojiColor(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(false)
	defer color.SetNoColor(origNoColor)

	// Red suits should have ANSI red codes
	heartCard := domain.NewCard(domain.CardDesignHeart, 10, false)
	got := cuiCardStrEmoji(heartCard)
	assert.Contains(t, got, "\033[31m")
	assert.Contains(t, got, "♥10")

	diamondCard := domain.NewCard(domain.CardDesignDiamond, 5, false)
	got = cuiCardStrEmoji(diamondCard)
	assert.Contains(t, got, "\033[31m")
	assert.Contains(t, got, "♦5")

	// Black suits should NOT have ANSI red codes
	spadeCard := domain.NewCard(domain.CardDesignSpade, 1, false)
	got = cuiCardStrEmoji(spadeCard)
	assert.NotContains(t, got, "\033[31m")
	assert.Equal(t, "♠1", got)
}

func TestCuiPlayerNameColor(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(false)
	defer color.SetNoColor(origNoColor)

	// Human player should be bold
	got := cuiPlayerName(&mockCuiPlayer{isHuman: true}, 0)
	assert.Contains(t, got, "\033[1m")
	assert.Contains(t, got, "あなた")

	// CPU player should be bold
	got = cuiPlayerName(&mockCuiPlayer{isHuman: false}, 1)
	assert.Contains(t, got, "\033[1m")
	assert.Contains(t, got, "CPU 1")

	// nil player should not be colored
	got = cuiPlayerName[*mockCuiPlayer](nil, 0)
	assert.Equal(t, "UNKNOWN", got)
}

// 共有ヘルパの境界。**空を返す条件を取り違えると、空行だけが出る。**
func TestCuiCaptureHintLine(t *testing.T) {
	hand := domain.NewBasraPlayer(true)
	hand.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	hand.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))

	// 捕獲候補がある札だけ注記が付く。
	line := cuiCaptureHintLine(hand, map[int][]int{0: {1, 3}}, "tablanet.captureHint")
	assert.Contains(t, line, "[0]SPADE 5")
	assert.Contains(t, line, "[1][3]")
	assert.NotContains(t, line, "HEART 9")

	// 候補が無ければ 1 行も出さない (空文字。空白 1 文字ではない)。
	assert.Empty(t, cuiCaptureHintLine(hand, map[int][]int{}, "tablanet.captureHint"))
	assert.Empty(t, cuiCaptureHintLine(hand, nil, "tablanet.captureHint"))
	// 値が空スライスのキーしか無い場合も同じ。
	assert.Empty(t, cuiCaptureHintLine(hand, map[int][]int{0: {}}, "tablanet.captureHint"))
}
