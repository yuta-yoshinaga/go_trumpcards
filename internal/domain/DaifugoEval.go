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

// daifugoMaxPlayableCombos は出せるカードを探すときに調べる組み合わせ数の上限。
//
// **超えたら印を付けない。**中途半端に一部だけ調べて印を付けると、印の無い
// カードが「出せない」と読めてしまう。今までどおり無印のほうがまだ正直。
const daifugoMaxPlayableCombos = 50000

// GetPlayableCardIndices は現在の手番プレイヤーの手札のうち、**いま出せる
// 組み合わせに1つでも含まれる**カードのインデックスを返す (#4733)。
//
// 判定は PlayerPlay が使う isPlayable そのものを通す。革命・11バック・
// スートロック・階段縛り・スペ3返しといった場の状態は全部そちらが見るので、
// ここで規則を書き直さない。**別実装にすると「出せる」と印を付けた札が
// 実際には弾かれる。**
//
// nil を返すのは次の場合:
//   - 人間の手番でない / ゲーム終了 / ペンディングアクション待ち
//   - 場が空で階段縛り中 (階段の全長を数え上げると組み合わせが爆発する)
//   - 調べるべき組み合わせが daifugoMaxPlayableCombos を超える
//
// **場が空 (階段縛り無し) なら全札。**単騎はいつでも出せる。
func (d *Daifugo) GetPlayableCardIndices() []int {
	if d.round.gameEndFlag || d.round.pendingActionType != DaifugoPendingNone {
		return nil
	}
	player := d.players[d.round.currentTurn]
	if !player.GetIsHuman() {
		return nil
	}
	n := player.GetCardsSize()
	if n == 0 {
		return nil
	}

	if d.round.tableCards == nil {
		if d.round.sequenceLocked {
			return nil
		}
		all := make([]int, n)
		for i := range all {
			all[i] = i
		}
		return all
	}

	k := len(d.round.tableCards)
	if k > n || combinationCountCapped(n, k, daifugoMaxPlayableCombos) > daifugoMaxPlayableCombos {
		return nil
	}

	marked := make([]bool, n)
	idx := make([]int, k)
	cards := make([]*Card, k)
	var walk func(start, depth int)
	walk = func(start, depth int) {
		if depth == k {
			for i, j := range idx {
				cards[i] = player.GetCard(j)
			}
			if d.isPlayable(cards) {
				for _, j := range idx {
					marked[j] = true
				}
			}
			return
		}
		for i := start; i <= n-(k-depth); i++ {
			idx[depth] = i
			walk(i+1, depth+1)
		}
	}
	walk(0, 0)

	out := make([]int, 0, n)
	for i, ok := range marked {
		if ok {
			out = append(out, i)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// combinationCountCapped は C(n, k) を返す。cap を超えた時点で打ち切って
// cap+1 を返すので、大きな n でも桁あふれしない。
func combinationCountCapped(n, k, cap int) int {
	if k < 0 || k > n {
		return 0
	}
	if k > n-k {
		k = n - k
	}
	result := 1
	for i := 1; i <= k; i++ {
		result = result * (n - k + i) / i
		if result > cap {
			return cap + 1
		}
	}
	return result
}
