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
func zhengIsPlayable(cards []*Card, tableCards []*Card, tablePlayType ZhengPlayType) bool {
	playType := zhengClassifyPlay(cards)
	if playType == ZhengPlayInvalid {
		return false
	}

	// リード時は任意の有効な組み合わせを出せる。
	if len(tableCards) == 0 {
		return true
	}

	// ジョーカーボムは全てに勝つ (ジョーカーボム同士は物理的に発生しない)。
	if playType == ZhengPlayJokerBomb {
		return tablePlayType != ZhengPlayJokerBomb
	}
	if tablePlayType == ZhengPlayJokerBomb {
		return false
	}

	// 爆弾は非爆弾役をタイプ・枚数無視で切れる。爆弾同士はランク比較。
	if playType == ZhengPlayBomb {
		if tablePlayType != ZhengPlayBomb {
			return true
		}
		return zhengPlayStrength(cards, playType) > zhengPlayStrength(tableCards, tablePlayType)
	}
	if tablePlayType == ZhengPlayBomb {
		return false
	}

	// 通常役: タイプと枚数が一致し、キーが厳密に高い場合のみ。
	if playType != tablePlayType || len(cards) != len(tableCards) {
		return false
	}
	return zhengPlayStrength(cards, playType) > zhengPlayStrength(tableCards, tablePlayType)
}
