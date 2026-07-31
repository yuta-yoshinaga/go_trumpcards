//go:build !js || !wasm || extra3

package domain

// KilleDesign は Kille 札の仮想デザイン値。
//
// Kille は**単一スート**なので、標準の 1〜4 (スート) とは別の値をひとつだけ
// 使う。Rook が鳥札に 5 を割り当てているのと同じ手口。
const KilleDesign = 6

// KilleDeckID は手続き描画パス (ADR-0033) に渡すデッキ系統 ID。
const KilleDeckID = "kille"

// KilleRank は Kille の 21 種のうちのひとつ。**値が大きいほど強い。**
//
// 強さの順は Harlequin > Cuckoo > Hussar > Pig > Cavalier > Inn > 12…1 >
// Wreath > Flowerpot > Mask。数札の 1〜12 が絵札より弱く、さらに下に 3 枚ある。
type KilleRank int

// Kille の 21 種。**弱い順**に並べてある (Card の value にそのまま入る)。
const (
	// KilleMask 最弱
	KilleMask KilleRank = iota + 1
	// KilleFlowerpot
	KilleFlowerpot
	// KilleWreath
	KilleWreath
	// KilleNum1 〜 KilleNum12 は数札 1〜12。
	KilleNum1
	KilleNum2
	KilleNum3
	KilleNum4
	KilleNum5
	KilleNum6
	KilleNum7
	KilleNum8
	KilleNum9
	KilleNum10
	KilleNum11
	KilleNum12
	// KilleInn 以降が絵札。
	KilleInn
	// KilleCavalier
	KilleCavalier
	// KillePig
	KillePig
	// KilleHussar
	KilleHussar
	// KilleCuckoo
	KilleCuckoo
	// KilleHarlequin 最強 (ただし交換されると最弱扱いになる)
	KilleHarlequin
	// KilleRankCount 種類数
	KilleRankCount = KilleHarlequin
)

// KilleCopies は 1 種あたりの枚数。**21 種 × 2 枚 = 42 枚。**
const KilleCopies = 2

// KilleDeckSize は札の総数。
const KilleDeckSize = int(KilleRankCount) * KilleCopies

// killeNames は各種の表示名。
var killeNames = map[KilleRank]string{
	KilleMask:      "Mask",
	KilleFlowerpot: "Flowerpot",
	KilleWreath:    "Wreath",
	KilleInn:       "Inn",
	KilleCavalier:  "Cavalier",
	KillePig:       "Pig",
	KilleHussar:    "Hussar",
	KilleCuckoo:    "Cuckoo",
	KilleHarlequin: "Harlequin",
}

// killeGlyphs は絵札と下位札の記号。数札は数字そのものを使う。
var killeGlyphs = map[KilleRank]string{
	KilleMask:      "🎭",
	KilleFlowerpot: "🪴",
	KilleWreath:    "🎗",
	KilleInn:       "🏠",
	KilleCavalier:  "🐎",
	KillePig:       "🐖",
	KilleHussar:    "⚔",
	KilleCuckoo:    "🐦",
	KilleHarlequin: "🃏",
}

// KilleRankName は種の表示名を返す。数札は "1".."12"。
func KilleRankName(r KilleRank) string {
	if n, ok := killeNumberOf(r); ok {
		return killeItoa(n)
	}
	if name, ok := killeNames[r]; ok {
		return name
	}
	return "?"
}

// KilleRankGlyph は種の記号を返す。数札は数字。
func KilleRankGlyph(r KilleRank) string {
	if n, ok := killeNumberOf(r); ok {
		return killeItoa(n)
	}
	if g, ok := killeGlyphs[r]; ok {
		return g
	}
	return "?"
}

// KilleRankColor は種の色調トークンを返す。
//
// 効果を持つ 5 種を目立たせる。Harlequin は向きで強弱が反転する特別扱いなので
// 単独の色にしてある。
func KilleRankColor(r KilleRank) string {
	switch r {
	case KilleHarlequin:
		return "purple"
	case KilleCuckoo:
		return "blue"
	case KilleHussar:
		return "red"
	case KillePig:
		return "gold"
	case KilleCavalier, KilleInn:
		return "green"
	default:
		return "black"
	}
}

// killeNumberOf は数札なら 1..12 を返す。
func killeNumberOf(r KilleRank) (int, bool) {
	if r < KilleNum1 || r > KilleNum12 {
		return 0, false
	}
	return int(r-KilleNum1) + 1, true
}

// killeItoa は小さな非負整数を文字列にする。strconv を持ち込まないための
// 最小実装。
func killeItoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}

// KilleIsPicture は絵札 (Inn 以上) かを返す。
func KilleIsPicture(r KilleRank) bool { return r >= KilleInn }

// NewKilleCard は種から札を作る。
func NewKilleCard(r KilleRank) *Card { return NewCard(KilleDesign, int(r), true) }

// KilleRankOf は札の種を返す。Kille の札でなければ 0。
func KilleRankOf(c *Card) KilleRank {
	if c == nil || c.GetDesign() != KilleDesign {
		return 0
	}
	r := KilleRank(c.GetValue())
	if r < KilleMask || r > KilleHarlequin {
		return 0
	}
	return r
}

// newKilleDeck は 42 枚 (21 種 × 2) を生成する。
func newKilleDeck() []*Card {
	deck := make([]*Card, 0, KilleDeckSize)
	for r := KilleMask; r <= KilleHarlequin; r++ {
		for range KilleCopies {
			deck = append(deck, NewKilleCard(r))
		}
	}
	return deck
}
