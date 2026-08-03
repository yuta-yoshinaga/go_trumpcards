//go:build !js || !wasm || solo

// Package domain タロッキーニ (Tarocchini / Ottocento) のドメインモデル。
//
// Tarocchini はイタリア・ボローニャの 62 枚タロット (タロッコ・ボロニェーゼ) を用いる
// 4 人・2 対 2 の固定チーム制トリックテイキングゲーム。対面同士がパートナーになる。
//
// # デッキ (62 枚)
//
//   - スート札 40 枚: design = 1..4 (4 スート)、value ∈ {1, 6, 7, 8, 9, 10, 11, 12, 13, 14}。
//     **2..5 のピップは抜かれている**のがボロニェーゼ版の特徴で、各スート 10 枚になる。
//   - 切り札 (trumps) 21 枚: design = TarocchiniTrumpDesign (5)、value = 1..21。
//   - マット (Matto / Fool) 1 枚: design = TarocchiniMattoDesign (6)、value = 0。
//
// # このゲームを他と分けているもの
//
// **同格のパパ 4 枚は「後から出した方」が勝つ。**他のトリックテイカーの勝敗判定は
// 「より強い札が出たら勝者を更新する」(厳密な >) で書けるが、パパ同士は順位が等しい
// ため、同値でも更新する (>=) 必要がある。Scarto の `trickWinner` をそのまま流用すると
// 先に出したパパが勝ってしまうので、本ゲームは専用の判定を持つ。
package domain

// TarocchiniPlayerCnt プレイヤー数 (人間 1 + CPU 3)。
const TarocchiniPlayerCnt = 4

// TarocchiniSuitCnt スート数。
const TarocchiniSuitCnt = 4

// TarocchiniTrumpDesign 切り札の仮想デザイン値。1..4 はスート、5 が切り札。
//
// Scarto / FrenchTarot と同じ割り当てにしてある。**新しいデッキ型は要らない** ——
// ADR-0033 の手続き描画パスがこの仮想デザインをそのまま描ける。
const TarocchiniTrumpDesign = 5

// TarocchiniMattoDesign マット (Matto / Fool) の仮想デザイン値。
const TarocchiniMattoDesign = 6

// TarocchiniMattoValue マットのカード値。
const TarocchiniMattoValue = 0

// TarocchiniMaxTrump 切り札の最大値。
const TarocchiniMaxTrump = 21

// TarocchiniKingValue 王 (Re) の値。
const TarocchiniKingValue = 14

// tarocchiniSuitValues 各スートに残る 10 枚の値。
//
// **2..5 は入らない。**ボロニェーゼ版は低位ピップを抜いた 40 枚構成で、これを
// 52 枚デッキの感覚で 1..10 に読み替えると枚数が合わなくなる。
var tarocchiniSuitValues = [...]int{1, 6, 7, 8, 9, 10, 11, 12, 13, 14}

// TarocchiniDeckSize デッキ枚数 (40 + 21 + 1)。
const TarocchiniDeckSize = TarocchiniSuitCnt*len(tarocchiniSuitValues) + TarocchiniMaxTrump + 1

// TarocchiniHandSize 各プレイヤーがトリックプレイで持つ札の枚数。
const TarocchiniHandSize = 15

// TarocchiniSurplus ディーラーが拾って捨てる余剰札の枚数。
//
// **62 は 4 で割り切れない。**15 枚ずつ配ると 60 枚で 2 枚余るので、その 2 枚は
// ディーラーが拾い、同数を捨てて手札を 15 枚に戻す (Scarto のスカルトと同型)。
const TarocchiniSurplus = TarocchiniDeckSize - TarocchiniPlayerCnt*TarocchiniHandSize

// TarocchiniTrickCount 1 ラウンドのトリック数。
const TarocchiniTrickCount = TarocchiniHandSize

// tarocchiniPapiLow / tarocchiniPapiHigh は同格に扱う切り札 (パパ / モーリ) の範囲。
//
// **この 4 枚の同定は未確定。**タロッコ・ボロニェーゼの切札は「ベガート(1) → 同格の
// パパ 4 枚 → 番号付きの中位 → 最上位 4 枚 (アンジェロ/モンド/ソーレ/ルーナ)」という
// 構成で、"同格 4 枚" と "最上位 4 枚" は別のグループ。issue #4409 本文は両者を
// 混同している疑いがあるため、ここではベガートの直上 4 枚 (切札 2..5) を採る。
// 同定が変われば**この 2 つの定数だけ**を差し替えればよい。
const (
	tarocchiniPapiLow  = 2
	tarocchiniPapiHigh = 5
)

// TarocchiniIsPapa 同格に扱うパパか。
func TarocchiniIsPapa(c *Card) bool {
	return tarocchiniIsTrump(c) &&
		c.GetValue() >= tarocchiniPapiLow && c.GetValue() <= tarocchiniPapiHigh
}

// tarocchiniIsTrump 切り札か。
func tarocchiniIsTrump(c *Card) bool {
	return c != nil && c.GetDesign() == TarocchiniTrumpDesign
}

// tarocchiniIsMatto マットか。
func tarocchiniIsMatto(c *Card) bool {
	return c != nil && c.GetDesign() == TarocchiniMattoDesign
}

// buildTarocchiniDeck 62 枚のデッキを組む。
func buildTarocchiniDeck() []*Card {
	deck := make([]*Card, 0, TarocchiniDeckSize)
	for suit := 1; suit <= TarocchiniSuitCnt; suit++ {
		for _, val := range tarocchiniSuitValues {
			deck = append(deck, NewCard(suit, val, false))
		}
	}
	for val := 1; val <= TarocchiniMaxTrump; val++ {
		deck = append(deck, NewCard(TarocchiniTrumpDesign, val, false))
	}
	deck = append(deck, NewCard(TarocchiniMattoDesign, TarocchiniMattoValue, false))
	return deck
}

// TarocchiniTeamOf 席のチーム番号を返す。対面同士 (0-2 / 1-3) が組む。
func TarocchiniTeamOf(seat int) int { return seat % 2 }

// TarocchiniPhase ゲームフェーズ。入札は無い。
type TarocchiniPhase int

const (
	// TarocchiniPhaseScarto ディーラーが余剰 2 枚を捨てるフェーズ。
	TarocchiniPhaseScarto TarocchiniPhase = iota
	// TarocchiniPhasePlay トリックプレイ。
	TarocchiniPhasePlay
	// TarocchiniPhaseTrickEnd トリック終了 (結果表示待ち)。
	TarocchiniPhaseTrickEnd
	// TarocchiniPhaseRoundEnd ラウンド終了 (精算済み)。
	TarocchiniPhaseRoundEnd
	// TarocchiniPhaseGameEnd マッチ終了。
	TarocchiniPhaseGameEnd
)

// tarocchiniWinRank トリック比較用のランクを返す。高いほど強い。
//
// マットはトリックを取らないので常に最弱。切り札はリードスートより常に強い。
// **パパ同士の同格判定はここではしない** —— ランクだけでは「後出し優先」を
// 表現できないため、trickWinner 側で扱う。
func tarocchiniWinRank(c *Card, led int) int {
	switch {
	case c == nil || tarocchiniIsMatto(c):
		return -1
	case TarocchiniIsPapa(c):
		// **4 枚のパパは 1 つのランクに畳む。**素の値のままだと 2..5 が別ランクに
		// なり、同格であることも後出し優先も表現できない。
		return 1000 + tarocchiniPapiLow
	case tarocchiniIsTrump(c):
		return 1000 + c.GetValue()
	case c.GetDesign() == led:
		return c.GetValue()
	default:
		return -1
	}
}

// tarocchiniTrickWinnerOf は与えられたトリックの勝者席を返す。
//
// **同格のパパは後から出した方が勝つ。**そのため比較は「厳密に強い」ではなく、
// 「強い、または (両方パパで) 同じランク」で勝者を更新する。Scarto から
// `r > winRank` をそのまま持ってくると先に出したパパが勝ってしまう。
func tarocchiniTrickWinnerOf(trick []*TrickCard, led int) int {
	if len(trick) == 0 {
		return 0
	}
	winIdx := trick[0].PlayerIdx
	winRank := -1
	var winCard *Card
	for _, tc := range trick {
		if tc == nil {
			continue
		}
		r := tarocchiniWinRank(tc.Card, led)
		tie := r == winRank && TarocchiniIsPapa(tc.Card) && TarocchiniIsPapa(winCard)
		if r > winRank || tie {
			winRank, winIdx, winCard = r, tc.PlayerIdx, tc.Card
		}
	}
	return winIdx
}
