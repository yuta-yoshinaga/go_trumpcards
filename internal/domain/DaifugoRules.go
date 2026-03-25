package domain

import (
	"fmt"
	"math/rand"
	"sort"
)

// triggerRevolutionIfNeeded 4枚出しで革命が起きるか判定し、起きた場合は革命フラグを切り替えて全プレイヤーの手札を再ソートする
// isSeq が true の場合は階段革命ルール (SequenceRevolutionEnabled) が有効な時のみ発動する
func (d *Daifugo) triggerRevolutionIfNeeded(cards []*Card, isSeq bool) {
	if len(cards) < 4 {
		return
	}
	if isSeq && !d.config.SequenceRevolutionEnabled {
		return
	}
	d.round.revolutionActive = !d.round.revolutionActive
	d.appendLog(-1, "revolution", "revolution!", nil)
	d.sortAllActiveHands()
}

// triggerEightCut 8切りチェック: 8が出されたら場をクリア
func (d *Daifugo) triggerEightCut(cards []*Card) bool {
	if !d.config.EightCutEnabled {
		return false
	}
	for _, c := range cards {
		if !IsJoker(c) && c.GetValue() == 8 {
			d.clearTableState()
			return true
		}
	}
	return false
}

// triggerSandstorm 砂嵐チェック: 3枚の非ジョーカー3が出されたら場をクリア
func (d *Daifugo) triggerSandstorm(cards []*Card) bool {
	if !d.config.SandstormEnabled {
		return false
	}
	if len(cards) != 3 {
		return false
	}
	for _, c := range cards {
		if IsJoker(c) || c.GetValue() != 3 {
			return false
		}
	}
	d.clearTableState()
	return true
}

// isValidEmperor エンペラー判定: 4枚の連番カード(全スート異なる)を場がクリアの時に出す
func (d *Daifugo) isValidEmperor(cards []*Card) bool {
	if !d.config.EmperorEnabled || len(cards) != 4 || d.round.tableCards != nil {
		return false
	}
	return d.isEmperorCards(cards)
}

// isEmperorCards カードがエンペラー条件を満たすか判定 (場の状態は見ない)
func (d *Daifugo) isEmperorCards(cards []*Card) bool {
	if len(cards) != 4 {
		return false
	}
	suits := make(map[int]bool)
	nonJokerValues := make([]int, 0, 4)
	jokerCount := 0
	for _, c := range cards {
		if IsJoker(c) {
			jokerCount++
			continue
		}
		if suits[c.GetDesign()] {
			return false // 同じスートが重複
		}
		suits[c.GetDesign()] = true
		nonJokerValues = append(nonJokerValues, d.cardStrength(c.GetValue()))
	}
	if len(nonJokerValues) == 0 {
		return false // 全ジョーカーは不可
	}
	sort.Ints(nonJokerValues)
	// 非ジョーカーの値が連続しているか確認
	gaps := 0
	for i := 1; i < len(nonJokerValues); i++ {
		diff := nonJokerValues[i] - nonJokerValues[i-1]
		if diff == 0 {
			return false // 重複値
		}
		gaps += diff - 1
	}
	// 非ジョーカー間の穴 + 両端への拡張で合計4枚になるか
	// 全体のスパン = max - min + 1 + 端に追加するジョーカー
	span := nonJokerValues[len(nonJokerValues)-1] - nonJokerValues[0] + 1
	remaining := jokerCount - gaps
	if remaining < 0 {
		return false // 穴を埋められない
	}
	totalSpan := span + remaining
	return totalSpan == 4
}

// triggerEmperor エンペラー発動: 革命を起こし場をクリア
func (d *Daifugo) triggerEmperor(cards []*Card) bool {
	if !d.config.EmperorEnabled {
		return false
	}
	if !d.isEmperorCards(cards) {
		return false
	}
	d.round.revolutionActive = !d.round.revolutionActive
	d.appendLog(-1, "revolution", "revolution!", nil)
	d.sortAllActiveHands()
	d.clearTableState()
	return true
}

// triggerElevenBack 11バックチェック: J(11)が出されたら11バック発動
func (d *Daifugo) triggerElevenBack(cards []*Card) {
	if !d.config.ElevenBackEnabled {
		return
	}
	for _, c := range cards {
		if !IsJoker(c) && c.GetValue() == 11 {
			d.round.elevenBackActive = !d.round.elevenBackActive
			d.sortAllActiveHands()
			return
		}
	}
}

// triggerNineReverseIfNeeded 9リバースチェック: 非ジョーカーの9が出されたらターン方向を反転（階段時は無効）
func (d *Daifugo) triggerNineReverseIfNeeded(cards []*Card, isSeq bool) {
	if !d.config.NineReverseEnabled || isSeq {
		return
	}
	for _, c := range cards {
		if !IsJoker(c) && c.GetValue() == 9 {
			d.round.reverseDirection = !d.round.reverseDirection
			return
		}
	}
}

// triggerCoupDetatIfNeeded クーデターチェック: 3枚の非ジョーカー9で革命を起こす
func (d *Daifugo) triggerCoupDetatIfNeeded(cards []*Card) {
	if !d.config.CoupDetatEnabled {
		return
	}
	if len(cards) != 3 {
		return
	}
	for _, c := range cards {
		if IsJoker(c) || c.GetValue() != 9 {
			return
		}
	}
	d.round.revolutionActive = !d.round.revolutionActive
	d.appendLog(-1, "revolution", "revolution!", nil)
	d.sortAllActiveHands()
}

// updateSuitLock スート縛りの更新
func (d *Daifugo) updateSuitLock(cards []*Card) {
	if d.config.SuitLockMode == DaifugoSuitLockNone {
		// スート縛りなしでも数縛りは独立して発動可能
		d.updateNumberLock(cards)
		return
	}
	if len(d.round.tableCards) == 0 {
		// 場がクリアだった → 縛りなし、今出したカードのスートを記録のみ
		d.round.suitLocked = false
		d.round.lockedSuit = 0
		return
	}
	// 前の場のカードと今出したカードのスートが一致するか確認 (ジョーカー除く)
	prevSuit := d.getNonJokerSuit(d.round.tableCards)
	newSuit := d.getNonJokerSuit(cards)
	if prevSuit > 0 && newSuit > 0 && prevSuit == newSuit {
		if !d.round.suitLocked {
			d.round.suitLocked = true
			d.round.lockedSuit = prevSuit
			d.updateNumberLock(cards)
		}
	}
}

// updateSequenceLock 階段縛りの更新: 場が階段 && 新しいプレイも階段 → 縛り発動
func (d *Daifugo) updateSequenceLock(isSeq bool) {
	if !d.config.SequenceLockEnabled {
		return
	}
	if d.round.tableIsSequence && isSeq {
		d.round.sequenceLocked = true
	}
}

// updateNumberLock 数縛り (連番縛り) の更新
func (d *Daifugo) updateNumberLock(cards []*Card) {
	if !d.config.NumberLockEnabled {
		return
	}
	if len(d.round.tableCards) == 0 {
		return
	}
	prevBase := getBaseValue(d.round.tableCards)
	newBase := getBaseValue(cards)
	if prevBase > 0 && newBase > 0 {
		prevStr := d.cardStrength(prevBase)
		newStr := d.cardStrength(newBase)
		if newStr-prevStr == 1 {
			d.round.numberLocked = true
		}
	}
}

// triggerFiveSkipIfNeeded 5飛びチェック: 非ジョーカーの5が出されたらプレイヤーをスキップ（階段時は無効）
// 戻り値はスキップ回数 (0 = スキップなし)
func (d *Daifugo) triggerFiveSkipIfNeeded(cards []*Card, isSeq bool) int {
	if !d.config.FiveSkipEnabled || isSeq {
		return 0
	}
	fiveCount := 0
	for _, c := range cards {
		if !IsJoker(c) && c.GetValue() == 5 {
			fiveCount++
		}
	}
	if fiveCount == 0 {
		return 0
	}
	skipCount := d.config.FiveSkipCount
	if skipCount < 1 {
		skipCount = 1
	}
	totalSkips := skipCount * fiveCount
	// アクティブプレイヤー数 - 1 でキャップ (無限ループ防止)
	maxSkips := d.getActivePlayerCnt() - 1
	if maxSkips < 0 {
		maxSkips = 0
	}
	if totalSkips > maxSkips {
		totalSkips = maxSkips
	}
	return totalSkips
}

// triggerSevenPassIfNeeded 7渡しチェック: 非ジョーカーの7が出されたらペンディングアクションをセット
func (d *Daifugo) triggerSevenPassIfNeeded(cards []*Card, isSeq bool) {
	if !d.config.SevenPassEnabled || isSeq {
		return
	}
	player := d.players[d.round.currentTurn]
	// 出した後に手札が残っている場合のみ
	if player.GetCardsSize() == 0 {
		return
	}
	for _, c := range cards {
		if !IsJoker(c) && c.GetValue() == 7 {
			// 渡す対象: 次のアクティブなプレイヤー (自分以外)
			target := d.getNextActivePlayer(d.round.currentTurn)
			if target >= 0 && target != d.round.currentTurn {
				d.round.pendingActionType = DaifugoPendingSevenPass
				d.round.pendingActionTarget = target
			}
			return
		}
	}
}

// triggerTenDiscardIfNeeded 10捨てチェック: 非ジョーカーの10が出されたらペンディングアクションをセット
func (d *Daifugo) triggerTenDiscardIfNeeded(cards []*Card, isSeq bool) {
	if !d.config.TenDiscardEnabled || isSeq {
		return
	}
	player := d.players[d.round.currentTurn]
	// 出した後に手札が残っている場合のみ
	if player.GetCardsSize() == 0 {
		return
	}
	for _, c := range cards {
		if !IsJoker(c) && c.GetValue() == 10 {
			d.round.pendingActionType = DaifugoPendingTenDiscard
			d.round.pendingActionTarget = -1
			return
		}
	}
}

// triggerQueenBomberIfNeeded 12ボンバーチェック: 非ジョーカーのQ(12)がグループプレイで出されたらペンディングアクションをセット
func (d *Daifugo) triggerQueenBomberIfNeeded(cards []*Card, isSeq bool) {
	if !d.config.QueenBomberEnabled || isSeq {
		return
	}
	player := d.players[d.round.currentTurn]
	// 出した後に手札が残っている場合のみ
	if player.GetCardsSize() == 0 {
		return
	}
	for _, c := range cards {
		if !IsJoker(c) && c.GetValue() == 12 {
			d.round.pendingActionType = DaifugoPendingQueenBomber
			d.round.pendingActionTarget = -1
			return
		}
	}
}

// isSpadeThreeCounter スペ3返し判定: 場がジョーカー1枚でスペードの3を1枚出す場合
func (d *Daifugo) isSpadeThreeCounter(cards []*Card) bool {
	if !d.config.SpadeThreeEnabled {
		return false
	}
	// 場がジョーカー1枚のみ
	if len(d.round.tableCards) != 1 || !IsJoker(d.round.tableCards[0]) {
		return false
	}
	// 出すカードがスペードの3を1枚のみ
	if len(cards) != 1 {
		return false
	}
	c := cards[0]
	return !IsJoker(c) && c.GetDesign() == CardDesignSpade && c.GetValue() == 3
}

// resolvePendingAction ペンディングアクションを解決する
func (d *Daifugo) resolvePendingAction(indices []int) error {
	// 12ボンバーは indices[0] をカード値 (1-13) として解釈する
	if d.round.pendingActionType == DaifugoPendingQueenBomber {
		if len(indices) != 1 {
			return NewDomainError(ErrInvalidPlay, "queen bomber requires exactly 1 card value")
		}
		v := indices[0]
		if v < 1 || v > 13 {
			return NewDomainError(ErrInvalidCard, fmt.Sprintf("card value %d out of range (1-13)", v))
		}
		d.resolveQueenBomber(v)
		d.finishEmptyPlayers()
		d.round.humanAction = &DaifugoCpuAction{PlayerIdx: d.round.currentTurn, PlayedCards: nil}

		d.round.pendingActionType = DaifugoPendingNone
		d.round.pendingActionTarget = -1

		if !d.checkGameEnd() {
			d.advanceTurn()
			d.checkPassClear()
		}
		return nil
	}

	if len(indices) != 1 {
		return NewDomainError(ErrInvalidPlay, "pending action requires exactly 1 card index")
	}
	player := d.players[d.round.currentTurn]
	card := player.GetCard(indices[0])
	if card == nil {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("card index %d out of range", indices[0]))
	}

	switch d.round.pendingActionType {
	case DaifugoPendingSevenPass:
		// 7渡し: カードを対象プレイヤーに渡す
		removed := player.RemoveCards([]int{indices[0]})
		target := d.players[d.round.pendingActionTarget]
		target.AddCard(removed[0])
		target.SortCardsByStrength(d.cardStrengthForCard)
		d.round.humanAction = &DaifugoCpuAction{PlayerIdx: d.round.currentTurn, PlayedCards: removed}
	case DaifugoPendingTenDiscard:
		// 10捨て: カードを捨てる
		removed := player.RemoveCards([]int{indices[0]})
		d.round.humanAction = &DaifugoCpuAction{PlayerIdx: d.round.currentTurn, PlayedCards: removed}
	}

	d.round.pendingActionType = DaifugoPendingNone
	d.round.pendingActionTarget = -1

	d.advanceTurn()
	d.checkPassClear()
	return nil
}

// resolveQueenBomber 12ボンバー解決: 全プレイヤーから指定値のカードを除去する
func (d *Daifugo) resolveQueenBomber(value int) {
	for _, p := range d.players {
		if p.GetIsFinished() {
			continue
		}
		p.RemoveCardsByValue(value)
	}
}

// isIllegalFinish 反則上がり判定: 8切り/ジョーカー/革命で上がりとなる手かどうか
func (d *Daifugo) isIllegalFinish(cards []*Card, isSeq bool) bool {
	if !d.config.IllegalFinishEnabled {
		return false
	}
	// 8切り上がり: 非ジョーカーの8が含まれている && 8切りが有効
	if d.config.EightCutEnabled {
		for _, c := range cards {
			if !IsJoker(c) && c.GetValue() == 8 {
				return true
			}
		}
	}
	// ジョーカー上がり
	for _, c := range cards {
		if IsJoker(c) {
			return true
		}
	}
	// 革命上がり: 4枚以上 (グループまたは階段)
	if len(cards) >= 4 {
		// 階段の場合は階段革命が有効な時のみ
		if isSeq && !d.config.SequenceRevolutionEnabled {
			return false
		}
		return true
	}
	return false
}

// wouldCauseIllegalFinish 指定カードで上がった場合に反則上がりになるか判定 (CPU AI用)
func (d *Daifugo) wouldCauseIllegalFinish(player *DaifugoPlayer, indices []int) bool {
	if !d.config.IllegalFinishEnabled || len(indices) == 0 {
		return false
	}
	if player.GetCardsSize() != len(indices) {
		return false // 上がりにならない
	}
	cards := make([]*Card, len(indices))
	for i, idx := range indices {
		cards[i] = player.GetCard(idx)
	}
	isSeq := d.config.SequenceEnabled && d.isValidSequence(cards)
	return d.isIllegalFinish(cards, isSeq)
}

// applyIllegalFinishPenalty 反則上がりペナルティを適用する
// ペナルティを受けたプレイヤーを最下位に降格し、他のプレイヤーのランクを調整する
func (d *Daifugo) applyIllegalFinishPenalty() {
	if !d.config.IllegalFinishEnabled {
		return
	}
	penalized := make([]*DaifugoPlayer, 0)
	nonPenalized := make([]*DaifugoPlayer, 0)
	for _, p := range d.players {
		if p.GetIllegalFinishPenalty() {
			penalized = append(penalized, p)
		} else {
			nonPenalized = append(nonPenalized, p)
		}
	}
	if len(penalized) == 0 {
		return
	}
	sort.Slice(nonPenalized, func(i, j int) bool {
		return nonPenalized[i].GetRank() < nonPenalized[j].GetRank()
	})
	sort.Slice(penalized, func(i, j int) bool {
		return penalized[i].GetRank() < penalized[j].GetRank()
	})
	rank := 1
	for _, p := range nonPenalized {
		p.SetRank(rank)
		rank++
	}
	for _, p := range penalized {
		p.SetRank(rank)
		rank++
	}
}

// applyCapitalFall 都落ちを適用する
// 前回大富豪だったプレイヤーが今回1位でない場合、そのプレイヤーと最下位のプレイヤーのランクを入れ替える
func (d *Daifugo) applyCapitalFall() {
	if !d.config.CapitalFallEnabled {
		return
	}
	// 前回大富豪のプレイヤーを探す
	prevDaifugoIdx := -1
	for i, p := range d.players {
		if p.GetPrevRank() == DaifugoRankDaifugo {
			prevDaifugoIdx = i
			break
		}
	}
	// 前回大富豪がいない (初回ゲーム) か今回も1位の場合はスキップ
	if prevDaifugoIdx < 0 || d.players[prevDaifugoIdx].GetRank() == DaifugoRankDaifugo {
		return
	}
	// 最下位のプレイヤーを探す
	lowestRank := 0
	lowestIdx := -1
	for i, p := range d.players {
		if p.GetRank() > lowestRank {
			lowestRank = p.GetRank()
			lowestIdx = i
		}
	}
	if lowestIdx < 0 || lowestIdx == prevDaifugoIdx {
		return
	}
	// ランクを入れ替える
	prevRank := d.players[prevDaifugoIdx].GetRank()
	d.players[prevDaifugoIdx].SetRank(d.players[lowestIdx].GetRank())
	d.players[lowestIdx].SetRank(prevRank)
}

// performCardExchange 前回のランクに基づいてカード交換を行う
func (d *Daifugo) performCardExchange() {
	d.round.exchangeActions = make([]*DaifugoExchangeAction, 0)

	// 前回のランクからプレイヤーインデックスのマッピングを作成
	rankToPlayer := make(map[int]int) // rank → current player index
	for i, p := range d.players {
		if p.GetPrevRank() > 0 {
			rankToPlayer[p.GetPrevRank()] = i
		}
	}

	// 大富豪 ↔ 大貧民: 2枚交換
	if idx1, ok1 := rankToPlayer[DaifugoRankDaifugo]; ok1 {
		if idx4, ok4 := rankToPlayer[DaifugoRankDaihinmin]; ok4 {
			d.exchangeCardsBetween(idx1, idx4, DaifugoExchangeCountDaifugo)
		}
	}

	// 富豪 ↔ 平民: 1枚交換
	if idx2, ok2 := rankToPlayer[DaifugoRankFugo]; ok2 {
		if idx3, ok3 := rankToPlayer[DaifugoRankHeimin]; ok3 {
			d.exchangeCardsBetween(idx2, idx3, DaifugoExchangeCountFugo)
		}
	}

	// 交換記録を棋譜に追加
	for _, ex := range d.round.exchangeActions {
		d.appendLog(ex.FromPlayerIdx, "exchange", fmt.Sprintf("exchanged %d card(s) with player %d", len(ex.Cards), ex.ToPlayerIdx), ex.Cards)
	}

	// 交換後に再ソート
	d.sortAllActiveHands()
}

// exchangeCardsBetween 上位プレイヤーと下位プレイヤー間でカード交換
// 下位→上位: 最強カードをcount枚渡す
// 上位→下位: 最弱カードをcount枚渡す
func (d *Daifugo) exchangeCardsBetween(upperIdx, lowerIdx, count int) {
	upper := d.players[upperIdx]
	lower := d.players[lowerIdx]

	if upper.GetCardsSize() < count || lower.GetCardsSize() < count {
		return
	}

	// 下位の最強カード(末尾)をcount枚取得
	lowerBestIndices := make([]int, count)
	for i := 0; i < count; i++ {
		lowerBestIndices[i] = lower.GetCardsSize() - count + i
	}
	lowerBestCards := lower.RemoveCards(lowerBestIndices)

	// 上位のカードをcount枚取得 (ブラインド交換: ランダム、通常: 最弱=先頭)
	var upperGiveIndices []int
	if d.config.BlindExchangeEnabled {
		perm := rand.Perm(upper.GetCardsSize())
		upperGiveIndices = perm[:count]
		sort.Ints(upperGiveIndices)
	} else {
		upperGiveIndices = make([]int, count)
		for i := 0; i < count; i++ {
			upperGiveIndices[i] = i
		}
	}
	upperWorstCards := upper.RemoveCards(upperGiveIndices)

	// カードを交換
	for _, c := range lowerBestCards {
		upper.AddCard(c)
	}
	for _, c := range upperWorstCards {
		lower.AddCard(c)
	}

	// 交換記録
	d.round.exchangeActions = append(d.round.exchangeActions, &DaifugoExchangeAction{
		FromPlayerIdx: lowerIdx,
		ToPlayerIdx:   upperIdx,
		Cards:         lowerBestCards,
	})
	d.round.exchangeActions = append(d.round.exchangeActions, &DaifugoExchangeAction{
		FromPlayerIdx: upperIdx,
		ToPlayerIdx:   lowerIdx,
		Cards:         upperWorstCards,
	})
}

// countFinished 既に上がっているプレイヤー数を返す
func (d *Daifugo) countFinished() int {
	return countPlayers(d.players, func(p *DaifugoPlayer) bool { return p.GetIsFinished() })
}

// getActivePlayerCnt アクティブ (未上がり) プレイヤー数取得
func (d *Daifugo) getActivePlayerCnt() int {
	return len(d.players) - d.countFinished()
}

// getNextActivePlayer fromの次のアクティブなプレイヤーインデックスを取得
func (d *Daifugo) getNextActivePlayer(from int) int {
	direction := 1
	if d.round.reverseDirection {
		direction = -1
	}
	return nextActivePlayer(d.players, from, direction)
}

// checkGameEnd ゲーム終了チェック (残り1人以下なら終了)
func (d *Daifugo) checkGameEnd() bool {
	active := d.getActivePlayerCnt()
	if active <= 1 {
		for i, p := range d.players {
			if !p.GetIsFinished() {
				d.finishPlayer(i)
				break
			}
		}
		d.round.gameEndFlag = true
		d.applyCapitalFall()
		d.applyIllegalFinishPenalty()
		return true
	}
	return false
}

// finishPlayer プレイヤーを上がりにしてランクを付与
// ランクは現在の上がり済みプレイヤー数 + 1 として計算する
func (d *Daifugo) finishPlayer(idx int) {
	rank := d.countFinished() + 1
	d.players[idx].SetIsFinished(true)
	d.players[idx].SetRank(rank)
	d.appendLog(idx, "finish", fmt.Sprintf("player %d finished (rank %d)", idx, rank), nil)
	// 上がったプレイヤーが最後に出したプレイヤーなら場をクリア
	if d.round.lastPlayPlayerIdx == idx {
		d.clearTableState()
	}
}

// finishEmptyPlayers 手札が0枚になった非finished プレイヤーを上がりにする (12ボンバー後の処理)
func (d *Daifugo) finishEmptyPlayers() {
	for i, p := range d.players {
		if !p.GetIsFinished() && p.GetCardsSize() == 0 {
			d.finishPlayer(i)
		}
	}
}

// SortHumanHand 人間プレイヤーの手札を指定モードでソートする
func (d *Daifugo) SortHumanHand(mode DaifugoSortMode) error {
	if d.round.gameEndFlag {
		return ErrGameEnded
	}
	d.sortMode = mode
	for _, p := range d.players {
		if p.GetIsHuman() && !p.GetIsFinished() {
			d.sortPlayerCards(p)
			break
		}
	}
	return nil
}

// sortAllActiveHands 全アクティブプレイヤーの手札をソートする
// 人間プレイヤーは sortMode に従い、CPUは常に強さ順
func (d *Daifugo) sortAllActiveHands() {
	for _, p := range d.players {
		if p.GetIsFinished() {
			continue
		}
		if p.GetIsHuman() {
			d.sortPlayerCards(p)
		} else {
			p.SortCardsByStrength(d.cardStrengthForCard)
		}
	}
}

// sortPlayerCards プレイヤーの手札を sortMode に従ってソートする
func (d *Daifugo) sortPlayerCards(p *DaifugoPlayer) {
	switch d.sortMode {
	case DaifugoSortBySuit:
		d.sortBySuit(p)
	case DaifugoSortByNumber:
		d.sortByNumber(p)
	default:
		p.SortCardsByStrength(d.cardStrengthForCard)
	}
}

// sortBySuit スート順でソート (Spade < Clover < Heart < Diamond, 同スート内は値の昇順、ジョーカーは末尾)
func (d *Daifugo) sortBySuit(p *DaifugoPlayer) {
	p.SortCardsByStrength(func(c *Card) int {
		if IsJoker(c) {
			return jokerSortWeight // ジョーカーは末尾
		}
		return c.GetDesign()*100 + c.GetValue()
	})
}

// sortByNumber 数字順でソート (値の昇順、同値ならスートの昇順、ジョーカーは末尾)
func (d *Daifugo) sortByNumber(p *DaifugoPlayer) {
	p.SortCardsByStrength(func(c *Card) int {
		if IsJoker(c) {
			return jokerSortWeight // ジョーカーは末尾
		}
		return c.GetValue()*100 + c.GetDesign()
	})
}
