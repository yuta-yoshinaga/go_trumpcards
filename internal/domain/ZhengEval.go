//go:build !js || !wasm || solo

package domain

import "sort"

// Zheng Shangyou (争上游) のランク強さ: 3 < 4 < ... < K < A < 2 < 小ジョーカー < 大ジョーカー
// スートは強さに一切関係しない (Tien Lenとの最大の違い)。

// zhengRankStrength はカードのランク強度を返す (3=0, ..., A=11, 2=12, 小ジョーカー=13, 大ジョーカー=14)。
func zhengRankStrength(card *Card) int {
	if card.GetDesign() == CardDesignJoker {
		if card.GetValue() == 2 {
			return 14 // 大ジョーカー
		}
		return 13 // 小ジョーカー
	}
	switch card.GetValue() {
	case 2:
		return 12
	case 1:
		return 11
	default:
		return card.GetValue() - 3
	}
}

// ZhengPlayType プレイの種類
type ZhengPlayType int

// ZhengPlayType定数
const (
	ZhengPlayInvalid   ZhengPlayType = 0
	ZhengPlaySingle    ZhengPlayType = 1
	ZhengPlayPair      ZhengPlayType = 2
	ZhengPlayTriple    ZhengPlayType = 3
	ZhengPlayStraight  ZhengPlayType = 4 // 3枚以上の連続ランク (3..Aのみ)
	ZhengPlayPairRun   ZhengPlayType = 5 // 3組以上の連続ペア (3..Aのみ)
	ZhengPlayBomb      ZhengPlayType = 6 // フォーカード (爆弾)
	ZhengPlayJokerBomb ZhengPlayType = 7 // ジョーカー2枚 (最強爆弾)
)

// zhengAllSameNonJokerValue は全カードがジョーカー以外の同ランクかを返す。
// ジョーカー2枚はランクが異なるため決してペアにならない (ジョーカーボムのみ)。
func zhengAllSameNonJokerValue(cards []*Card) bool {
	if len(cards) == 0 {
		return false
	}
	for _, c := range cards {
		if c.GetDesign() == CardDesignJoker {
			return false
		}
	}
	for i := 1; i < len(cards); i++ {
		if cards[i].GetValue() != cards[0].GetValue() {
			return false
		}
	}
	return true
}

// zhengIsJokerBomb はちょうどジョーカー2枚 (小+大) かを返す。
func zhengIsJokerBomb(cards []*Card) bool {
	if len(cards) != 2 {
		return false
	}
	if cards[0].GetDesign() != CardDesignJoker || cards[1].GetDesign() != CardDesignJoker {
		return false
	}
	v0, v1 := cards[0].GetValue(), cards[1].GetValue()
	return (v0 == 1 && v1 == 2) || (v0 == 2 && v1 == 1)
}

// zhengClassifyPlay はカードの組み合わせをZhengのプレイタイプに分類する。
func zhengClassifyPlay(cards []*Card) ZhengPlayType {
	switch n := len(cards); n {
	case 0:
		return ZhengPlayInvalid
	case 1:
		return ZhengPlaySingle
	case 2:
		if zhengIsJokerBomb(cards) {
			return ZhengPlayJokerBomb
		}
		if zhengAllSameNonJokerValue(cards) {
			return ZhengPlayPair
		}
		return ZhengPlayInvalid
	case 3:
		if zhengAllSameNonJokerValue(cards) {
			return ZhengPlayTriple
		}
		if zhengCheckStraight(cards) {
			return ZhengPlayStraight
		}
		return ZhengPlayInvalid
	case 4:
		if zhengAllSameNonJokerValue(cards) {
			return ZhengPlayBomb
		}
		if zhengCheckStraight(cards) {
			return ZhengPlayStraight
		}
		return ZhengPlayInvalid
	default:
		if zhengCheckPairRun(cards) {
			return ZhengPlayPairRun
		}
		if zhengCheckStraight(cards) {
			return ZhengPlayStraight
		}
		return ZhengPlayInvalid
	}
}

// zhengCheckStraight は3枚以上の連続ランク (3..Aのみ、2とジョーカーは不可) かを返す。
func zhengCheckStraight(cards []*Card) bool {
	n := len(cards)
	if n < 3 {
		return false
	}
	strs := make([]int, n)
	for i, c := range cards {
		if c.GetDesign() == CardDesignJoker || c.GetValue() == 2 {
			return false
		}
		strs[i] = zhengRankStrength(c)
	}
	sort.Ints(strs)
	for i := 1; i < n; i++ {
		if strs[i] != strs[i-1]+1 {
			return false
		}
	}
	return true
}

// zhengCheckPairRun は3組以上の連続ペア (3..Aのみ、2とジョーカーは不可) かを返す。
func zhengCheckPairRun(cards []*Card) bool {
	n := len(cards)
	if n < 6 || n%2 != 0 {
		return false
	}
	freq := make(map[int]int)
	for _, c := range cards {
		if c.GetDesign() == CardDesignJoker || c.GetValue() == 2 {
			return false
		}
		freq[zhengRankStrength(c)]++
	}
	if len(freq) != n/2 {
		return false
	}
	strs := make([]int, 0, len(freq))
	for s, cnt := range freq {
		if cnt != 2 {
			return false
		}
		strs = append(strs, s)
	}
	sort.Ints(strs)
	for i := 1; i < len(strs); i++ {
		if strs[i] != strs[i-1]+1 {
			return false
		}
	}
	return true
}

// zhengPlayStrength はプレイの比較キーとなるランク強度 (最高ランク) を返す。
// 無効なプレイには-1を返す。
func zhengPlayStrength(cards []*Card, playType ZhengPlayType) int {
	switch playType {
	case ZhengPlaySingle, ZhengPlayPair, ZhengPlayTriple,
		ZhengPlayStraight, ZhengPlayPairRun, ZhengPlayBomb:
		maxStr := -1
		for _, c := range cards {
			if s := zhengRankStrength(c); s > maxStr {
				maxStr = s
			}
		}
		return maxStr
	case ZhengPlayJokerBomb:
		return 14
	default:
		return -1
	}
}

// zhengIsBombType は爆弾系のプレイタイプかを返す。
func zhengIsBombType(playType ZhengPlayType) bool {
	return playType == ZhengPlayBomb || playType == ZhengPlayJokerBomb
}

// zhengCandidateStrength はタイプ横断で比較可能なプレイ強度を返す (CPU候補ソート用)。
// 爆弾系は通常役より常に強い扱いにして温存判断を単純化する。
func zhengCandidateStrength(cards []*Card) int {
	playType := zhengClassifyPlay(cards)
	switch playType {
	case ZhengPlayJokerBomb:
		return 1000
	case ZhengPlayBomb:
		return 100 + zhengPlayStrength(cards, playType)
	default:
		return zhengPlayStrength(cards, playType)
	}
}

// zhengIsPlayable は場のカードに対して出せるかを判定する。
//
// 通常役は場と同タイプ・同枚数でより高いキーのみ。爆弾 (フォーカード) は
// 任意の非爆弾役を切れる。爆弾同士はランク比較。ジョーカーボムは爆弾を含む
// 全てに勝つ。爆弾・ジョーカーボムはリードでも出せる。
// エカルトではなく「出せない理由」の識別子。Web (frontend/src/utils/zhengComboValidator.ts
// の ZhengInvalidReason) と同じ語を使い、両者は golden vector で縛ってある。
const (
	// ZhengInvalidType どの役にもならない組み合わせ。
	ZhengInvalidType = "invalidType"
	// ZhengInvalidWrongType 役にはなるが、場と役の種類が違う。
	ZhengInvalidWrongType = "wrongType"
	// ZhengInvalidWrongCount 役の種類は同じだが枚数が違う。
	ZhengInvalidWrongCount = "wrongCount"
	// ZhengInvalidTooWeak 種類も枚数も合っているが場より弱い。
	ZhengInvalidTooWeak = "tooWeak"
	// ZhengInvalidNeedBomb 場が爆弾なので通常役では切れない。
	ZhengInvalidNeedBomb = "needBomb"
	// ZhengInvalidUnbeatable 場がジョーカーボムなので何も勝てない。
	ZhengInvalidUnbeatable = "unbeatable"
)

// ZhengInvalidReason は cards を場に出せない理由を返す (出せるなら "")。
// 判定順は zhengIsPlayable と同じで、そちらはこの関数に委譲している。
func ZhengInvalidReason(cards []*Card, tableCards []*Card, tablePlayType ZhengPlayType) string {
	playType := zhengClassifyPlay(cards)
	if playType == ZhengPlayInvalid {
		return ZhengInvalidType
	}
	// リード時は任意の有効な組み合わせを出せる。
	if len(tableCards) == 0 {
		return ""
	}
	// ジョーカーボムは全てに勝つ (ジョーカーボム同士は物理的に発生しない)。
	if playType == ZhengPlayJokerBomb {
		if tablePlayType == ZhengPlayJokerBomb {
			return ZhengInvalidUnbeatable
		}
		return ""
	}
	if tablePlayType == ZhengPlayJokerBomb {
		return ZhengInvalidUnbeatable
	}
	// 爆弾は非爆弾役をタイプ・枚数無視で切れる。爆弾同士はランク比較。
	if playType == ZhengPlayBomb {
		if tablePlayType != ZhengPlayBomb {
			return ""
		}
		if zhengPlayStrength(cards, playType) > zhengPlayStrength(tableCards, tablePlayType) {
			return ""
		}
		return ZhengInvalidTooWeak
	}
	if tablePlayType == ZhengPlayBomb {
		return ZhengInvalidNeedBomb
	}
	// 通常役: タイプと枚数が一致し、キーが厳密に高い場合のみ。
	if playType != tablePlayType {
		return ZhengInvalidWrongType
	}
	if len(cards) != len(tableCards) {
		return ZhengInvalidWrongCount
	}
	if zhengPlayStrength(cards, playType) > zhengPlayStrength(tableCards, tablePlayType) {
		return ""
	}
	return ZhengInvalidTooWeak
}

func zhengIsPlayable(cards []*Card, tableCards []*Card, tablePlayType ZhengPlayType) bool {
	return ZhengInvalidReason(cards, tableCards, tablePlayType) == ""
}
