package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// isRedSuit returns true for heart and diamond suits.
func isRedSuit(design int) bool {
	return design == domain.CardDesignHeart || design == domain.CardDesignDiamond
}

// cuiCardList is the minimal type constraint required by formatCardList.
type cuiCardList interface {
	GetCardsSize() int
	GetCard(idx int) *domain.Card
}

// cardFormatter formats a single card into a string.
type cardFormatter func(card *domain.Card) string

// formatCardList formats all cards in a cuiCardList using the given formatter and separator.
// When indexed is true, each card is prefixed with "[N]".
func formatCardList(hand cuiCardList, fmtCard cardFormatter, sep string, indexed bool) string {
	parts := make([]string, hand.GetCardsSize())
	for i := range parts {
		s := fmtCard(hand.GetCard(i))
		if indexed {
			s = fmt.Sprintf("[%d]%s", i, s)
		}
		parts[i] = s
	}
	return strings.Join(parts, sep)
}

// formatCardSlice formats a card slice using the given formatter and separator.
func formatCardSlice(cards []*domain.Card, fmtCard cardFormatter, sep string) string {
	parts := make([]string, len(cards))
	for i, c := range cards {
		parts[i] = fmtCard(c)
	}
	return strings.Join(parts, sep)
}

// cuiCardListStr returns a comma-separated card string for all cards in hand.
func cuiCardListStr(hand cuiCardList) string {
	return formatCardList(hand, cuiCardStr, ",", false)
}

// cuiPlayer is the minimal type constraint required by cuiPlayerName.
type cuiPlayer interface {
	comparable
	GetIsHuman() bool
}

// suitNames maps design constants to suit name strings.
// Index 0 is unused (joker); indices 1–4 correspond to CardDesignSpade–CardDesignDiamond.
var suitNames = []string{"", "SPADE", "CLOVER", "HEART", "DIAMOND"}

// cuiCardStr returns a text-based card string (e.g. "SPADE 5", "JOKER", "??").
// Used by BlackJack, OldMaid, Daifugo, Sevens, and Doubt CUI presenters.
func cuiCardStr(card *domain.Card) string {
	if card == nil {
		return "??"
	}
	if card.GetDesign() == domain.CardDesignJoker {
		return "JOKER"
	}
	name := cuiSuitName(card.GetDesign())
	if name == "UNKNOWN" {
		return "UNKNOWN"
	}
	s := name + " " + strconv.Itoa(card.GetValue())
	if isRedSuit(card.GetDesign()) {
		return color.Red(s)
	}
	return s
}

// cuiCardStrEmoji returns an emoji-based card string (e.g. "♠5", "🃏0").
// Used by Poker and Holdem CUI presenters.
func cuiCardStrEmoji(card *domain.Card) string {
	if card == nil {
		return "??"
	}
	designs := []string{"🃏", "♠", "♣", "♥", "♦"}
	d := card.GetDesign()
	if d < 0 || d >= len(designs) {
		d = 0
	}
	s := fmt.Sprintf("%s%d", designs[d], card.GetValue())
	if isRedSuit(card.GetDesign()) {
		return color.Red(s)
	}
	return s
}

// cuiSuitName returns the suit name string for a given design constant.
// Used by Daifugo and Sevens CUI presenters.
func cuiSuitName(suit int) string {
	if suit > 0 && suit < len(suitNames) {
		return suitNames[suit]
	}
	return "UNKNOWN"
}

// cuiRankLabel returns the card-face label for a rank value: A for 1, J/Q/K for
// 11/12/13, and the plain number otherwise (7–10). Locale-independent — matches
// the card-face notation used elsewhere in the UI. Used by the Watten CUI
// presenter for Schlag-rank display.
func cuiRankLabel(rank int) string {
	switch rank {
	case 1:
		return "A"
	case 11:
		return "J"
	case 12:
		return "Q"
	case 13:
		return "K"
	default:
		return strconv.Itoa(rank)
	}
}

// cuiPokerHandName returns the localized display name for a poker hand rank
// (0=High Card .. 10=Five of a Kind), resolved via the shared pokerHandRank*
// keys in cui_common. Out-of-range ranks fall back to the raw English
// domain.PokerHandNames entry (or "" when the index is invalid). Used by the
// UltimateTexasHoldem and MississippiStud CUI presenters.
func cuiPokerHandName(rank int) string {
	if rank < 0 || rank >= len(domain.PokerHandNames) {
		return ""
	}
	return i18n.T("pokerHandRank" + strconv.Itoa(rank))
}

// cuiPlayerName returns the human-friendly display name for a player:
// "You" / "あなた" for the human, "CPU N" for CPU opponents, or
// "UNKNOWN" if the player is nil/zero. Locale-aware via i18n.T (issue
// #1699 Phase 1). Used by OldMaid, Daifugo, Sevens, Doubt, Poker,
// Holdem, Omaha, Hearts, Spades, CrazyEights, GinRummy, and Memory.
func cuiPlayerName[P cuiPlayer](player P, idx int) string {
	var zero P
	if player == zero {
		return i18n.T("cuiPlayerUnknown")
	}
	if player.GetIsHuman() {
		return color.Bold(i18n.T("cuiPlayerYou"))
	}
	return color.Bold(i18n.Tf("cuiPlayerCpu", "idx", strconv.Itoa(idx)))
}

// cuiPlayerWithStyle is the type constraint for players that have a play style.
type cuiPlayerWithStyle interface {
	cuiPlayer
	GetPlayStyleName() string
}

// cuiPlayerNameWithStyle returns cuiPlayerName with play style suffix for CPU.
// Used by Poker, Holdem, and Omaha CUI presenters.
func cuiPlayerNameWithStyle[P cuiPlayerWithStyle](player P, idx int) string {
	name := cuiPlayerName(player, idx)
	if !player.GetIsHuman() {
		name = fmt.Sprintf("%s (%s)", name, player.GetPlayStyleName())
	}
	return name
}

// cuiIndexedCardListStr returns a double-space separated indexed card string.
// e.g. "[0]SPADE 5  [1]HEART 3"
func cuiIndexedCardListStr(hand cuiCardList) string {
	return formatCardList(hand, cuiCardStr, "  ", true)
}

// CuiHoleMark は「その札は自分の手札から出したもの」を示す印。ショーダウンで
// ベストハンド5枚を並べたとき、どれがボード由来でどれが手札由来かを分ける。
//
// **オマハ系は手札から使う枚数が固定** (通常4枚のうち2枚、Big O は5枚のうち2枚)
// なので、10通りの組み合わせのどれが役になったのかは印が無いと追えない (#5484)。
const CuiHoleMark = "*"

// CuiLegalMark は「この札は今出せる」ことを示す印。CrazyEights / Wizard /
// Mushi / GoFish が以前から使っている後置の "*" に合わせている。
const CuiLegalMark = "*"

// cuiPlayableMarkedCardListStr returns an indexed card list where the cards at
// the given indices are suffixed with CuiLegalMark.
//
// **CUI プレイヤーだけが「どれを出せるか」を番号入力とエラーで学ぶしかなかった。**
// Web は validIndices でリング表示しているので、同じ情報をテキストでも出す。
// playable が nil または空のときは目印を付けない -- ビッド中や CPU の手番など、
// そもそも制限を出していない状態と区別するため (#4725)。
func cuiPlayableMarkedCardListStr(hand cuiCardList, playable []int) string {
	return cuiIndexMarkedCardListStr(hand, playable, CuiLegalMark)
}

// CuiBombMark は「その札はボムを構成できる」ことを示す印 (Tichu)。
//
// 合法手の "*" とも交換由来の "+" とも別の記号にする ── 同じ画面で意味の違う
// 印が同じ形だと、どちらの話をしているのか読めない (#5635)。
const CuiBombMark = "!"

// CuiKittyMark は「その札は交換で入ってきた」ことを示す印。
//
// **合法手の "*" とは別の記号にする。**同じ画面で意味の違う 2 つの印が同じ形だと、
// どちらの話をしているのか読めない (#5632)。
const CuiKittyMark = "+"

// CuiTrumpMark は「その札は切り札」であることを示す印 (Doppelkopf)。
//
// 合法手の "*"、ボムの "!"、交換由来の "+" とはまた別の記号にする ── 同じ画面で
// 意味の違う印が同じ形だと、どちらの話をしているのか読めない (#5639)。
const CuiTrumpMark = "^"

// CuiTopTrumpMark は「その札は固定の最上位切り札」であることを示す印
// (Auction Forty-Fives)。
//
// 切り札の印 "^" とも別にする ── Forty-Fives の最上位切り札は「強い」だけでなく
// **マストフォローが免除される** という別の意味を持つので、同じ形にすると
// どちらの話か読めない (#5643)。
const CuiTopTrumpMark = "!!"

// CuiWildMark は「その札はこのラウンドのワイルド」であることを示す印
// (Three Thirteen)。
//
// 合法手の "*"、ボムの "!"、交換由来の "+"、切り札の "^"、最上位切り札の "!!" とは
// また別の記号にする ── 同じ画面で意味の違う印が同じ形だと、どちらの話をして
// いるのか読めない (#5667)。
const CuiWildMark = "~"

// CuiNewestMark は「その札は今のストリートで新しく公開された」ことを示す印
// (Five Card Stud / Soko)。
//
// これまでの印 ("*" 合法手 / "!" ボム / "+" 交換由来 / "^" 切り札 / "!!" 最上位
// 切り札 / "~" ワイルド / "@" 交換候補) とはまた別の記号にする ── 同じ画面で
// 意味の違う印が同じ形だと、どちらの話をしているのか読めない (#5674)。
const CuiNewestMark = "<"

// cuiIndexMarkedCardListStr returns an indexed card list where the cards at the
// given indices carry mark.
//
// indices が nil または空のときは目印を付けない -- 「制限が無い」状態と
// 「全部が対象」を取り違えないため。
func cuiIndexMarkedCardListStr(hand cuiCardList, indices []int, mark string) string {
	if len(indices) == 0 {
		return cuiIndexedCardListStr(hand)
	}
	marked := make(map[int]bool, len(indices))
	for _, i := range indices {
		marked[i] = true
	}
	parts := make([]string, hand.GetCardsSize())
	for i := range parts {
		parts[i] = fmt.Sprintf("[%d]%s", i, cuiCardStr(hand.GetCard(i)))
		if marked[i] {
			parts[i] += mark
		}
	}
	return strings.Join(parts, "  ")
}

// cuiCardListStrEmoji returns a double-space separated emoji card string (no index).
// e.g. "♠5  ♥3"
func cuiCardListStrEmoji(hand cuiCardList) string {
	return formatCardList(hand, cuiCardStrEmoji, "  ", false)
}

// cuiIndexedCardListStrEmoji returns a double-space separated indexed emoji card string.
// e.g. "[0]♠5  [1]♥3"
func cuiIndexedCardListStrEmoji(hand cuiCardList) string {
	return formatCardList(hand, cuiCardStrEmoji, "  ", true)
}

// cuiCardSliceStr returns a comma-space separated card string from a card slice.
// e.g. "SPADE 5, HEART 3"
func cuiCardSliceStr(cards []*domain.Card) string {
	return formatCardSlice(cards, cuiCardStr, ", ")
}

// cuiCardSliceStrEmoji returns a double-space separated emoji card string from a card slice.
// e.g. "♠5  ♥3"
func cuiCardSliceStrEmoji(cards []*domain.Card) string {
	return formatCardSlice(cards, cuiCardStrEmoji, "  ")
}

// cuiCardSliceStrEmojiNewest は cuiCardSliceStrEmoji と同じ並びに、末尾の 1 枚
// だけ CuiNewestMark を付けて返す。
//
// ストリートごとに 1 枚ずつ増える公開札で「今回増えたのはどれか」を示すための
// もの (#5674)。空スライスは空文字。
func cuiCardSliceStrEmojiNewest(cards []*domain.Card) string {
	if len(cards) == 0 {
		return ""
	}
	parts := make([]string, len(cards))
	for i, c := range cards {
		parts[i] = cuiCardStrEmoji(c)
	}
	parts[len(parts)-1] += CuiNewestMark
	return strings.Join(parts, "  ")
}

// cuiCaptureHintLine annotates each hand card with the table cards it can
// capture, e.g. `[0]♠5 → 場[1][3]`.
//
// Shared by the fishing games (Basra, Tablanet), which all get the pairing from
// the domain's own `GetCaptureOptions`. **Recomputing it per presenter is how
// the note starts disagreeing with what the server will accept** (#4922).
// Cards that capture nothing get no note; an empty result means no line at all.
func cuiCaptureHintLine(hand cuiCardList, opts map[int][]int, key string) string {
	if len(opts) == 0 {
		return ""
	}
	notes := make([]string, 0, hand.GetCardsSize())
	for h := range hand.GetCardsSize() {
		tableIdxs := opts[h]
		if len(tableIdxs) == 0 {
			continue
		}
		marks := make([]string, len(tableIdxs))
		for j, ti := range tableIdxs {
			marks[j] = "[" + strconv.Itoa(ti) + "]"
		}
		notes = append(notes, i18n.Tf(key,
			"hand", "["+strconv.Itoa(h)+"]"+cuiCardStr(hand.GetCard(h)),
			"table", strings.Join(marks, "")))
	}
	if len(notes) == 0 {
		return ""
	}
	return strings.Join(notes, "  ")
}

// cuiDiscardPileLines lists a whole discard pile, oldest first, in indexed form
// ("[0]SPADE 5 [1]HEART 9"), wrapped every cuiDiscardPerLine cards. Returns ""
// for an empty pile.
//
// **山ごと取るゲームでは捨て札の中身は公開情報。**Canasta/Burraco は一番上しか
// 出しておらず、「山全体を取る」判断を 1 枚で迫っていた (#4833, #5043)。
func cuiDiscardPileLines(pile []*domain.Card, key string) string {
	// 空の山は 1 行も出さない。長さ 0 なら下のループが回らないので、専用のガードは
	// 置かない (置いても分岐が到達不能になるだけ)。
	parts := make([]string, 0, len(pile))
	for i, c := range pile {
		parts = append(parts, "["+strconv.Itoa(i)+"]"+cuiCardStr(c))
	}
	var b strings.Builder
	for i := 0; i < len(parts); i += cuiDiscardPerLine {
		end := min(i+cuiDiscardPerLine, len(parts))
		b.WriteString(i18n.Tf(key, "cards", strings.Join(parts[i:end], " ")) + "\n")
	}
	return b.String()
}

// cuiDiscardPerLine は捨て札一覧の 1 行あたりの枚数。20 枚超の山を 1 行に並べると
// 折り返しで読めなくなる。
const cuiDiscardPerLine = 8
