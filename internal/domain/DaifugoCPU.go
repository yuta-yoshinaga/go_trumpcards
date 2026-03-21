package domain

import (
	"math"
	"sort"
)

// cardSearchOpts カード検索のオプション
type cardSearchOpts struct {
	skipRevolution  bool // 革命防止: 4枚以上のグループをスキップ (Normal用)
	selectStrongest bool // 最強のグループを選択 (Hard urgent用)
}

// suitCardEntry 同スートカードのインデックスと強さ
type suitCardEntry struct {
	idx      int
	strength int
}

// findBestPlay 難易度に応じたCPUプレイ戦略のディスパッチャー
func (d *Daifugo) findBestPlay(player *DaifugoPlayer) []int {
	switch d.config.CpuDifficulty {
	case DaifugoDifficultyEasy:
		return d.findBestPlayEasy(player)
	case DaifugoDifficultyHard:
		return d.findBestPlayHard(player)
	default:
		return d.findBestPlayNormal(player)
	}
}

// opponentMinCards 対戦相手の中で最小の手札枚数を返す
func (d *Daifugo) opponentMinCards(player *DaifugoPlayer) int {
	minCards := math.MaxInt
	for _, p := range d.players {
		if p == player || p.GetIsFinished() {
			continue
		}
		if p.GetCardsSize() < minCards {
			minCards = p.GetCardsSize()
		}
	}
	return minCards
}

// isUrgent 対戦相手が3枚以下の手札を持っているか (緊急モード判定)
func (d *Daifugo) isUrgent(player *DaifugoPlayer) bool {
	return d.opponentMinCards(player) <= 3
}

// shouldStrategicPass Hard AIの戦略的パス判定
// 場の強さが10以下で、出そうとしているカードがA(14)以上の強さで、手札が6枚以上の場合にパス
func (d *Daifugo) shouldStrategicPass(player *DaifugoPlayer, indices []int) bool {
	if player.GetCardsSize() <= 5 {
		return false
	}
	tableBase := getBaseValue(d.tableCards)
	var tableStrength int
	if tableBase < 0 {
		tableStrength = DaifugoJokerStrength
	} else {
		tableStrength = d.cardStrength(tableBase)
	}
	if tableStrength > 10 {
		return false
	}
	// 出そうとしているカードの強さを確認
	for _, idx := range indices {
		card := player.GetCard(idx)
		if IsJoker(card) {
			return true // ジョーカーは温存
		}
		if d.cardStrength(card.GetValue()) >= DaifugoCardStrength(1) { // A以上 (革命に依存しない固定閾値)
			return true
		}
	}
	return false
}

// calcTableStrength 場のカードの強さを計算する (d.tableCards が非nil前提)
func (d *Daifugo) calcTableStrength() int {
	tableBase := getBaseValue(d.tableCards)
	if tableBase < 0 {
		return DaifugoJokerStrength
	}
	return d.cardStrength(tableBase)
}

// searchCardGroupSuitCheck searchCardGroup内のスート縛りチェック
// 両縛り: 先頭カードのスートがロックスートと一致する必要がある
// 片縛り: グループ内に少なくとも1枚ロックスートと一致するカードがあればOK
func (d *Daifugo) searchCardGroupSuitCheck(player *DaifugoPlayer, start, end int) bool {
	if d.config.SuitLockMode == DaifugoSuitLockFull {
		return player.GetCard(start).GetDesign() == d.lockedSuit
	}
	// 片縛り
	for k := start; k < end; k++ {
		if player.GetCard(k).GetDesign() == d.lockedSuit {
			return true
		}
	}
	return false
}

// searchCardGroup 手札からカードグループを検索する共通ヘルパー
// 手札は強さ順 (弱→強) でソート済みなので、selectStrongest=true の場合は
// 後に見つかるグループほど強いため、上書きし続けて最後の結果を返す。
func (d *Daifugo) searchCardGroup(player *DaifugoPlayer, needed int, tableStrength int, opts cardSearchOpts) []int {
	jokerIndices := d.findJokerIndices(player)
	var bestIndices []int
	i := 0
	for i < player.GetCardsSize() {
		card := player.GetCard(i)
		if IsJoker(card) {
			i++
			continue
		}
		v := card.GetValue()
		j := i
		for j < player.GetCardsSize() && !IsJoker(player.GetCard(j)) && player.GetCard(j).GetValue() == v {
			j++
		}
		count := j - i
		if count >= needed && d.cardStrength(v) > tableStrength {
			// 革命防止: 4枚以上かつ革命未発動かつ出した後も手札が残る場合はスキップ
			if opts.skipRevolution && count >= 4 && !d.revolutionActive && player.GetCardsSize() > count {
				i = j
				continue
			}
			// スート縛りチェック
			if d.suitLocked && d.config.SuitLockMode != DaifugoSuitLockNone {
				if !d.searchCardGroupSuitCheck(player, i, j) {
					i = j
					continue
				}
			}
			indices := make([]int, needed)
			for k := 0; k < needed; k++ {
				indices[k] = i + k
			}
			if !opts.selectStrongest {
				return indices
			}
			bestIndices = indices
		}
		// ジョーカーで補完してグループを作れるか
		if count < needed && count > 0 && d.cardStrength(v) > tableStrength {
			if count+len(jokerIndices) >= needed {
				// スート縛りチェック
				if d.suitLocked && d.config.SuitLockMode != DaifugoSuitLockNone {
					if !d.searchCardGroupSuitCheck(player, i, j) {
						i = j
						continue
					}
				}
				indices := make([]int, 0, needed)
				for k := 0; k < count && len(indices) < needed; k++ {
					indices = append(indices, i+k)
				}
				for _, ji := range jokerIndices {
					if len(indices) >= needed {
						break
					}
					indices = append(indices, ji)
				}
				sort.Ints(indices)
				if !opts.selectStrongest {
					return indices
				}
				bestIndices = indices
			}
		}
		i = j
	}
	return bestIndices
}

// collectSuitCards startIdx のカードと同スートで、より強いカードを強さ昇順で収集する
func (d *Daifugo) collectSuitCards(player *DaifugoPlayer, startIdx int) []suitCardEntry {
	card := player.GetCard(startIdx)
	suit := card.GetDesign()
	startStrength := d.cardStrengthForCard(card)
	suitCards := []suitCardEntry{{startIdx, startStrength}}
	for nextIdx := startIdx + 1; nextIdx < player.GetCardsSize(); nextIdx++ {
		nextCard := player.GetCard(nextIdx)
		if IsJoker(nextCard) || nextCard.GetDesign() != suit {
			continue
		}
		nextStr := d.cardStrengthForCard(nextCard)
		if nextStr <= startStrength {
			continue
		}
		suitCards = append(suitCards, suitCardEntry{nextIdx, nextStr})
	}
	return suitCards
}

// tryBuildSequence suitCards の si 番目から始めて jokers で穴を埋めながら needed 枚の階段を構築する
func tryBuildSequence(suitCards []suitCardEntry, si int, jokerIndices []int, needed int) []int {
	indices := []int{suitCards[si].idx}
	lastStr := suitCards[si].strength
	jokersUsed := 0
	sci := si + 1
	for len(indices) < needed {
		targetStr := lastStr + 1
		found := false
		// suitCards は強さ昇順なので、sci 位置が targetStr と一致するかだけ確認すればよい
		if sci < len(suitCards) && suitCards[sci].strength == targetStr {
			indices = append(indices, suitCards[sci].idx)
			lastStr = targetStr
			sci++
			found = true
		}
		if !found {
			if jokersUsed < len(jokerIndices) {
				indices = append(indices, jokerIndices[jokersUsed])
				jokersUsed++
				lastStr = targetStr
			} else {
				break
			}
		}
	}
	if len(indices) != needed {
		return nil
	}
	return indices
}

// findOpeningSingleCard 場がクリアの時に、フィルタ条件に合うカードを探す (反則上がりチェック付き)
// filters は優先度順に適用され、reverse=true で手札末尾から探索する
func (d *Daifugo) findOpeningSingleCard(player *DaifugoPlayer, filters []func(*Card) bool, reverse bool) []int {
	var fallbackIdx *int
	n := player.GetCardsSize()
	for _, filter := range filters {
		for step := 0; step < n; step++ {
			i := step
			if reverse {
				i = n - 1 - step
			}
			if filter(player.GetCard(i)) {
				if !d.wouldCauseIllegalFinish(player, []int{i}) {
					return []int{i}
				}
				if fallbackIdx == nil {
					v := i
					fallbackIdx = &v
				}
			}
		}
	}
	if player.GetCardsSize() > 0 {
		if !d.wouldCauseIllegalFinish(player, []int{0}) {
			return []int{0}
		}
		if fallbackIdx == nil {
			v := 0
			fallbackIdx = &v
		}
	}
	if fallbackIdx != nil {
		return []int{*fallbackIdx}
	}
	return nil
}

// findJokerSinglePlay ジョーカーを単体で出せる場合にそのインデックスを返す
func (d *Daifugo) findJokerSinglePlay(player *DaifugoPlayer, tableStrength int) []int {
	for i := 0; i < player.GetCardsSize(); i++ {
		if IsJoker(player.GetCard(i)) && DaifugoJokerStrength > tableStrength {
			return []int{i}
		}
	}
	return nil
}

// findBestPlayNormal 通常難易度: 既存ロジック (最弱のカードを出す、8/ジョーカー温存)
func (d *Daifugo) findBestPlayNormal(player *DaifugoPlayer) []int {
	if d.tableCards == nil {
		return d.findNormalOpeningPlay(player)
	}
	return d.findNormalResponsePlay(player)
}

// findNormalOpeningPlay Normal AIの先手プレイ: エンペラー → 階段縛り時は階段 → 最弱非8非ジョーカー → 8 → ジョーカー
func (d *Daifugo) findNormalOpeningPlay(player *DaifugoPlayer) []int {
	if emperorIndices := d.findEmperorPlay(player); emperorIndices != nil {
		if !d.wouldCauseIllegalFinish(player, emperorIndices) {
			return emperorIndices
		}
	}
	if d.sequenceLocked {
		return d.findOpeningSequencePlay(player)
	}
	return d.findOpeningSingleCard(player, []func(*Card) bool{
		func(c *Card) bool { return !IsJoker(c) && c.GetValue() != 8 },
		func(c *Card) bool { return !IsJoker(c) && c.GetValue() == 8 },
	}, false)
}

// findNormalResponsePlay Normal AIの応手プレイ: 階段 → 最弱グループ → ジョーカー単体
func (d *Daifugo) findNormalResponsePlay(player *DaifugoPlayer) []int {
	if d.tableIsSequence && d.config.SequenceEnabled {
		return d.findBestSequencePlay(player)
	}

	needed := len(d.tableCards)
	tableStrength := d.calcTableStrength()

	if indices := d.searchCardGroup(player, needed, tableStrength, cardSearchOpts{skipRevolution: true}); indices != nil {
		return indices
	}

	if needed == 1 {
		if tableStrength >= DaifugoCardStrength(2) && player.GetCardsSize() > 3 {
			return nil
		}
		return d.findJokerSinglePlay(player, tableStrength)
	}

	return nil
}

// findBestPlayEasy 簡単難易度: 単純に出せる最弱のカードを出す (8/ジョーカー温存なし、エンペラー探索なし、革命防止なし)
func (d *Daifugo) findBestPlayEasy(player *DaifugoPlayer) []int {
	if d.tableCards == nil {
		// 階段縛り中は階段を探す
		if d.sequenceLocked {
			return d.findOpeningSequencePlay(player)
		}
		// 場がクリアなら最弱の1枚を出す (温存戦略なし、反則上がりチェックなし: Easy AIは戦略なしで失敗もする)
		if player.GetCardsSize() > 0 {
			return []int{0}
		}
		return nil
	}

	// 場が階段の場合
	if d.tableIsSequence && d.config.SequenceEnabled {
		return d.findBestSequencePlay(player)
	}

	needed := len(d.tableCards)
	tableStrength := d.calcTableStrength()

	// 最弱のグループを探す (革命防止なし、ジョーカー温存なし)
	if indices := d.searchCardGroup(player, needed, tableStrength, cardSearchOpts{}); indices != nil {
		return indices
	}

	// ジョーカー単体 (温存判定なし)
	if needed == 1 {
		return d.findJokerSinglePlay(player, tableStrength)
	}

	return nil
}

// findBestPlayHard 難しい難易度: 対戦相手の手札状況を考慮したヒューリスティックAI
// 終盤では「完全読み」ソルバーを使用して確実に上がれる手順を計算する
func (d *Daifugo) findBestPlayHard(player *DaifugoPlayer) []int {
	// 完全読み: 終盤で確実に上がれる手順があればそれを使う
	if indices := d.trySolveEndgame(player); indices != nil {
		return indices
	}
	if d.tableCards == nil {
		return d.findHardOpeningPlay(player)
	}
	return d.findHardResponsePlay(player)
}

// findHardOpeningPlay Hard AIの先手プレイ: エンペラー → 階段縛り時は階段 → 緊急時は最強 / 非緊急時はNormal委譲
func (d *Daifugo) findHardOpeningPlay(player *DaifugoPlayer) []int {
	if emperorIndices := d.findEmperorPlay(player); emperorIndices != nil {
		if !d.wouldCauseIllegalFinish(player, emperorIndices) {
			return emperorIndices
		}
	}
	if d.sequenceLocked {
		return d.findOpeningSequencePlay(player)
	}
	if d.isUrgent(player) {
		return d.findOpeningSingleCard(player, []func(*Card) bool{
			func(c *Card) bool { return !IsJoker(c) },
		}, true)
	}
	return d.findNormalOpeningPlay(player)
}

// findHardResponsePlay Hard AIの応手プレイ: 緊急時は最強グループ / 非緊急時はNormal+戦略的パス
func (d *Daifugo) findHardResponsePlay(player *DaifugoPlayer) []int {
	if d.tableIsSequence && d.config.SequenceEnabled {
		return d.findBestSequencePlayHard(player)
	}

	needed := len(d.tableCards)
	tableStrength := d.calcTableStrength()

	if d.isUrgent(player) {
		if indices := d.searchCardGroup(player, needed, tableStrength, cardSearchOpts{selectStrongest: true}); indices != nil {
			return indices
		}
		if needed == 1 {
			return d.findJokerSinglePlay(player, tableStrength)
		}
		return nil
	}

	normalIndices := d.findNormalResponsePlay(player)
	if normalIndices != nil && d.shouldStrategicPass(player, normalIndices) {
		return nil
	}
	return normalIndices
}

// findBestSequencePlayHard Hard AIの階段モード
func (d *Daifugo) findBestSequencePlayHard(player *DaifugoPlayer) []int {
	if !d.isUrgent(player) {
		return d.findBestSequencePlay(player)
	}
	return d.findSequencePlay(player, true)
}

// findJokerIndices プレイヤーの手札中のジョーカーのインデックスを返す
func (d *Daifugo) findJokerIndices(player *DaifugoPlayer) []int {
	indices := make([]int, 0)
	for i := 0; i < player.GetCardsSize(); i++ {
		if IsJoker(player.GetCard(i)) {
			indices = append(indices, i)
		}
	}
	return indices
}

// findBestSequencePlay 階段モードで出せる最弱の階段を探す (ジョーカーで穴を埋められる)
func (d *Daifugo) findBestSequencePlay(player *DaifugoPlayer) []int {
	return d.findSequencePlay(player, false)
}

// findSequencePlay 階段モードで出せる階段を探す (selectStrongest=true で最強、false で最弱)
func (d *Daifugo) findSequencePlay(player *DaifugoPlayer, selectStrongest bool) []int {
	needed := len(d.tableCards)
	tableMinStr := d.getSequenceMinStrength(d.tableCards)
	jokerIndices := d.findJokerIndices(player)

	var bestIndices []int
	for startIdx := 0; startIdx < player.GetCardsSize(); startIdx++ {
		if IsJoker(player.GetCard(startIdx)) {
			continue
		}
		suitCards := d.collectSuitCards(player, startIdx)
		for si := 0; si < len(suitCards); si++ {
			indices := tryBuildSequence(suitCards, si, jokerIndices, needed)
			if indices == nil {
				continue
			}
			testCards := make([]*Card, len(indices))
			for i, idx := range indices {
				testCards[i] = player.GetCard(idx)
			}
			minStr := d.getSequenceMinStrength(testCards)
			if minStr > tableMinStr {
				sort.Ints(indices)
				if !selectStrongest {
					return indices
				}
				bestIndices = indices
			}
		}
	}
	return bestIndices
}

// findOpeningSequencePlay 階段縛り中に場がクリアの先手で出せる最短の階段を探す
func (d *Daifugo) findOpeningSequencePlay(player *DaifugoPlayer) []int {
	jokerIndices := d.findJokerIndices(player)
	const minLen = 3

	var bestIndices []int
	for startIdx := 0; startIdx < player.GetCardsSize(); startIdx++ {
		if IsJoker(player.GetCard(startIdx)) {
			continue
		}
		suitCards := d.collectSuitCards(player, startIdx)
		for si := 0; si < len(suitCards); si++ {
			indices := tryBuildSequence(suitCards, si, jokerIndices, minLen)
			if indices == nil {
				continue
			}
			testCards := make([]*Card, len(indices))
			for i, idx := range indices {
				testCards[i] = player.GetCard(idx)
			}
			if !d.isValidSequence(testCards) {
				continue
			}
			sort.Ints(indices)
			if !d.wouldCauseIllegalFinish(player, indices) {
				return indices
			}
			if bestIndices == nil {
				bestIndices = indices
			}
		}
	}
	return bestIndices
}

// findEmperorPlay エンペラーの組み合わせを探す (場がクリアの時のみ)
func (d *Daifugo) findEmperorPlay(player *DaifugoPlayer) []int {
	if !d.config.EmperorEnabled || d.tableCards != nil {
		return nil
	}
	n := player.GetCardsSize()
	if n < 4 {
		return nil
	}
	// 全4枚の組み合わせを探索 (C(n,4))
	for a := 0; a < n-3; a++ {
		for b := a + 1; b < n-2; b++ {
			for c := b + 1; c < n-1; c++ {
				for dd := c + 1; dd < n; dd++ {
					testCards := []*Card{
						player.GetCard(a),
						player.GetCard(b),
						player.GetCard(c),
						player.GetCard(dd),
					}
					if d.isEmperorCards(testCards) {
						return []int{a, b, c, dd}
					}
				}
			}
		}
	}
	return nil
}

// findStrongestNonJokerIndex プレイヤーの手札中の最強の非ジョーカーカードのインデックスを返す (末尾から探索)
func (d *Daifugo) findStrongestNonJokerIndex(player *DaifugoPlayer) int {
	for i := player.GetCardsSize() - 1; i >= 0; i-- {
		if !IsJoker(player.GetCard(i)) {
			return i
		}
	}
	return 0
}

// findWeakestNonJokerIndex プレイヤーの手札中の最弱の非ジョーカーカードのインデックスを返す (先頭から探索)
func (d *Daifugo) findWeakestNonJokerIndex(player *DaifugoPlayer) int {
	for i := 0; i < player.GetCardsSize(); i++ {
		if !IsJoker(player.GetCard(i)) {
			return i
		}
	}
	return 0
}

// cpuResolvePendingAction CPUがペンディングアクションを自動解決する
func (d *Daifugo) cpuResolvePendingAction() {
	player := d.players[d.currentTurn]

	var idx int
	switch d.pendingActionType {
	case DaifugoPendingSevenPass:
		// 7渡し: 最強の非ジョーカーカードを渡す
		idx = d.findStrongestNonJokerIndex(player)
		removed := player.RemoveCards([]int{idx})
		target := d.players[d.pendingActionTarget]
		target.AddCard(removed[0])
		target.SortCardsByStrength(d.cardStrengthForCard)
		action := &DaifugoCpuAction{PlayerIdx: d.currentTurn, PlayedCards: removed}
		d.cpuActions = append(d.cpuActions, action)
	case DaifugoPendingTenDiscard:
		// 10捨て: 最弱の非ジョーカーカードを捨てる
		idx = d.findWeakestNonJokerIndex(player)
		removed := player.RemoveCards([]int{idx})
		action := &DaifugoCpuAction{PlayerIdx: d.currentTurn, PlayedCards: removed}
		d.cpuActions = append(d.cpuActions, action)
	case DaifugoPendingQueenBomber:
		// 12ボンバー: 対戦相手から最も多くカードを除去できる値を選ぶ
		bestValue := d.cpuChooseQueenBomberValue(player)
		d.resolveQueenBomber(bestValue)
		d.finishEmptyPlayers()
		action := &DaifugoCpuAction{PlayerIdx: d.currentTurn, PlayedCards: nil}
		d.cpuActions = append(d.cpuActions, action)
	}

	d.pendingActionType = DaifugoPendingNone
	d.pendingActionTarget = -1

	if d.checkGameEnd() {
		return
	}

	d.advanceTurn()
	d.checkPassClear()
}

// cpuChooseQueenBomberValue CPUが12ボンバーで選ぶ値を決定する (対戦相手から最も多くカードを除去できる値)
func (d *Daifugo) cpuChooseQueenBomberValue(self *DaifugoPlayer) int {
	bestValue := 1
	bestCount := 0
	for v := 1; v <= 13; v++ {
		count := 0
		for _, p := range d.players {
			if p == self || p.GetIsFinished() {
				continue
			}
			count += p.CountCardsByValue(v)
		}
		if count > bestCount {
			bestCount = count
			bestValue = v
		}
	}
	return bestValue
}
