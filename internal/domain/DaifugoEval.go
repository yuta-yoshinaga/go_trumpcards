package domain

import "sort"

// DaifugoCardStrength カードの強さを返す (3が最弱、2が最強)
// 3 < 4 < 5 < 6 < 7 < 8 < 9 < 10 < J(11) < Q(12) < K(13) < A(1) < 2(2)
func DaifugoCardStrength(v int) int {
	if v == 1 {
		return 14 // Ace
	}
	if v == 2 {
		return 15 // 2 は最強
	}
	return v
}

// DaifugoCardStrengthRevolution 革命中のカードの強さを返す (2が最弱、3が最強)
// 2 < A(1) < K(13) < Q(12) < J(11) < 10 < 9 < 8 < 7 < 6 < 5 < 4 < 3
func DaifugoCardStrengthRevolution(v int) int {
	return 18 - DaifugoCardStrength(v)
}

// DaifugoJokerStrength ジョーカーの強さ (常に最強)
const DaifugoJokerStrength = 16

// IsJoker カードがジョーカーかどうか判定
func IsJoker(card *Card) bool {
	return card.GetDesign() == CardDesignJoker
}

// cardStrength 現在の革命・11バック状態に応じたカード値の強さを返す
func (d *Daifugo) cardStrength(v int) int {
	// 革命と11バックのXOR: 両方有効なら打ち消し合う
	reversed := d.round.revolutionActive != d.round.elevenBackActive
	if reversed {
		return DaifugoCardStrengthRevolution(v)
	}
	return DaifugoCardStrength(v)
}

// cardStrengthForCard カードオブジェクトの強さを返す (ジョーカー対応)
func (d *Daifugo) cardStrengthForCard(card *Card) int {
	if card.GetDesign() == CardDesignJoker {
		return DaifugoJokerStrength
	}
	return d.cardStrength(card.GetValue())
}

// getBaseValue ジョーカーを除いたカード配列の共通値を取得 (全ジョーカーなら-1)
func getBaseValue(cards []*Card) int {
	for _, c := range cards {
		if !IsJoker(c) {
			return c.GetValue()
		}
	}
	return -1 // 全てジョーカー
}

// isValidGroup カード配列がグループ (同じ値 + ジョーカーワイルド) として有効かチェック
func isValidGroup(cards []*Card) bool {
	base := getBaseValue(cards)
	if base < 0 {
		// 全てジョーカー → 有効 (ジョーカーだけのグループ)
		return true
	}
	for _, c := range cards {
		if !IsJoker(c) && c.GetValue() != base {
			return false
		}
	}
	return true
}

// isValidSequence カード配列が階段 (同スートの連続した値 + ジョーカーワイルド) として有効かチェック
// 3枚以上が必要。革命・11バック状態に応じた強さを使用する。
func (d *Daifugo) isValidSequence(cards []*Card) bool {
	if len(cards) < 3 {
		return false
	}

	// ジョーカー以外のカードを収集
	suit := 0
	nonJokerValues := make([]int, 0)
	jokerCount := 0
	for _, c := range cards {
		if IsJoker(c) {
			jokerCount++
			continue
		}
		if suit == 0 {
			suit = c.GetDesign()
		} else if c.GetDesign() != suit {
			return false // スートが混在
		}
		nonJokerValues = append(nonJokerValues, d.cardStrength(c.GetValue()))
	}

	if len(nonJokerValues) == 0 {
		// 全ジョーカー → 階段としては不成立
		return false
	}

	sort.Ints(nonJokerValues)

	// ジョーカーで穴を埋められるか確認
	gaps := 0
	for i := 1; i < len(nonJokerValues); i++ {
		diff := nonJokerValues[i] - nonJokerValues[i-1]
		if diff == 0 {
			return false // 重複値
		}
		gaps += diff - 1
	}

	// 足りない分をジョーカーで埋める
	return gaps <= jokerCount
}

// getSequenceMinStrength 階段の最小強さ (最弱カードの強さ) を返す
func (d *Daifugo) getSequenceMinStrength(cards []*Card) int {
	minStr := DaifugoJokerStrength + 1
	for _, c := range cards {
		str := d.cardStrengthForCard(c)
		if str < minStr {
			minStr = str
		}
	}
	return minStr
}

// isPlayable 指定したカードが場のカードに対して出せるか判定
func (d *Daifugo) isPlayable(cards []*Card) bool {
	if len(cards) == 0 {
		return false
	}

	// スペ3返し特例: 場がジョーカー1枚のみ && スペードの3を1枚出す
	if d.isSpadeThreeCounter(cards) {
		return true
	}

	// プレイタイプ判定
	validGroup := isValidGroup(cards)
	validSeq := d.config.SequenceEnabled && d.isValidSequence(cards)
	validEmperor := d.isValidEmperor(cards)

	if !validGroup && !validSeq && !validEmperor {
		return false
	}

	if d.round.tableCards == nil {
		// 階段縛り中は階段またはエンペラーのみ
		if d.round.sequenceLocked {
			return validSeq || validEmperor
		}
		// 場がクリアなら何でも出せる
		return true
	}

	// 枚数が一致しているか
	if len(cards) != len(d.round.tableCards) {
		return false
	}

	// 場が階段の場合
	if d.round.tableIsSequence {
		if !validSeq {
			return false
		}
		tableMin := d.getSequenceMinStrength(d.round.tableCards)
		playMin := d.getSequenceMinStrength(cards)
		return playMin > tableMin
	}

	// グループプレイ
	if !validGroup {
		return false
	}

	// スート縛りチェック
	if d.round.suitLocked && d.config.SuitLockMode != DaifugoSuitLockNone {
		if d.config.SuitLockMode == DaifugoSuitLockFull {
			// 両縛り: 全てのカードがロックされたスートに一致する必要がある
			newSuit := d.getNonJokerSuit(cards)
			if newSuit > 0 && newSuit != d.round.lockedSuit {
				return false
			}
		} else {
			// 片縛り: 少なくとも1枚がロックされたスートに一致する必要がある
			if !d.hasMatchingSuit(cards, d.round.lockedSuit) {
				return false
			}
		}
	}

	// 強さ比較
	tableBase := getBaseValue(d.round.tableCards)
	playBase := getBaseValue(cards)

	var tableStrength, playStrength int
	if tableBase < 0 {
		tableStrength = DaifugoJokerStrength
	} else {
		tableStrength = d.cardStrength(tableBase)
	}
	if playBase < 0 {
		playStrength = DaifugoJokerStrength
	} else {
		playStrength = d.cardStrength(playBase)
	}

	// 数縛り: 連番縛り発動中は強さの差が1でなければ出せない
	// ジョーカーは連番縛りをバイパスし、通常の強さ比較のみ適用
	if d.round.numberLocked && d.config.NumberLockEnabled {
		if playBase > 0 && tableBase > 0 {
			return playStrength-tableStrength == 1
		}
	}

	return playStrength > tableStrength
}

// getNonJokerSuit カード配列からジョーカー以外のスートを取得 (全て同じスートなら返す、混在なら0)
func (d *Daifugo) getNonJokerSuit(cards []*Card) int {
	suit := 0
	for _, c := range cards {
		if IsJoker(c) {
			continue
		}
		if suit == 0 {
			suit = c.GetDesign()
		} else if suit != c.GetDesign() {
			return 0 // 混在
		}
	}
	return suit
}

// hasMatchingSuit カード配列に指定スートのカードが少なくとも1枚含まれるか判定 (片縛り用)
func (d *Daifugo) hasMatchingSuit(cards []*Card, suit int) bool {
	for _, c := range cards {
		if !IsJoker(c) && c.GetDesign() == suit {
			return true
		}
	}
	return false
}
