//go:build !js || !wasm || casino

package domain

// dealerPlay ディーラーのカードドロー＆精算
func (b *BlackJack) dealerPlay() {
	if !b.allPlayerHandsDone() {
		b.DealerHit()
	} else {
		// 全ハンドバースト/サレンダーでもホールカードをカウント
		if b.config.CountingEnabled && !b.holeCardCounted && b.dealer.GetCardsSize() >= 2 {
			b.updateRunningCount(b.dealer.GetCard(1))
			b.holeCardCounted = true
		}
		b.endGame()
	}
}

// dealerShouldHit ディーラーがヒットすべきか判定（ソフト17ルール対応）
func (b *BlackJack) dealerShouldHit() bool {
	score := b.dealer.GetScore()
	if score < 17 {
		return true
	}
	if score == 17 && b.config.DealerHitsSoft17 && b.dealer.IsSoft() {
		return true
	}
	return false
}

// DealerHit ディーラーヒット
func (b *BlackJack) DealerHit() {
	// ホールカード（裏向きの2枚目）をカウント
	if b.config.CountingEnabled && !b.holeCardCounted && b.dealer.GetCardsSize() >= 2 {
		b.updateRunningCount(b.dealer.GetCard(1))
		b.holeCardCounted = true
	}
	for {
		if b.dealerShouldHit() {
			card := b.drawCard()
			if card == nil {
				b.appendLog(-1, "dealerstand", "dealer stand", nil)
				b.DealerStand()
				break
			}
			b.dealer.AddCard(card)
			b.updateRunningCount(card)
			b.appendLog(-1, "dealerhit", "dealer hit", []*Card{card})
		} else {
			b.appendLog(-1, "dealerstand", "dealer stand", nil)
			b.DealerStand()
			break
		}
	}
}

// DealerStand ディーラースタンド
func (b *BlackJack) DealerStand() {
	b.endGame()
}

// allPlayerHandsDone 全プレイヤーハンド（人間+CPU）がバーストまたはサレンダーしているか
func (b *BlackJack) allPlayerHandsDone() bool {
	for _, hand := range b.playerHands {
		if !hand.IsBusted() && !hand.IsSurrendered() {
			return false
		}
	}
	for _, cpu := range b.cpuPlayers {
		for _, hand := range cpu.GetHands() {
			if !hand.IsBusted() && !hand.IsSurrendered() {
				return false
			}
		}
	}
	return true
}

// endGame ゲーム終了処理
func (b *BlackJack) endGame() {
	b.resolvePayouts()
	b.resolvePayoutsCpu()
	b.gameEndFlag = true
	b.phase = BJPhaseEnd
	// 棋譜に結果を記録
	result := b.GameJudgment()
	var detail string
	switch result {
	case GameResultWin:
		detail = "player wins"
	case GameResultDraw:
		detail = "draw"
	case GameResultLose:
		detail = "player loses"
	}
	b.appendLog(-1, "result", detail, nil)
}

// judgeHandCore 共通ハンド勝敗判定ロジック
// fromSplit=true の場合はスプリット由来のためBJとして判定しない
func (b *BlackJack) judgeHandCore(hand *BlackJackHand, fromSplit bool) GameResult {
	playerScore := hand.GetScore()
	dealerScore := b.dealer.GetScore()

	if playerScore > 21 {
		return GameResultLose
	}
	if dealerScore > 21 {
		return GameResultWin
	}
	// スパニッシュ21等: プレイヤー21は常に勝利 (ディーラー21でも勝ち)
	if playerScore == 21 && b.variant != nil && b.variant.Player21AlwaysWins {
		return GameResultWin
	}
	if playerScore > dealerScore {
		return GameResultWin
	}
	if dealerScore > playerScore {
		return GameResultLose
	}

	// スコアが同じ場合、ナチュラルブラックジャックを確認
	dealerBJ := b.dealer.GetCardsSize() == 2 && dealerScore == 21
	playerBJ := hand.GetCardsSize() == 2 && playerScore == 21 && !fromSplit

	if playerBJ && !dealerBJ {
		return GameResultWin
	}
	if dealerBJ && !playerBJ {
		return GameResultLose
	}
	// 両者BJ時、バリアントによってはプレイヤー勝利
	if playerBJ && dealerBJ && b.variant != nil && b.variant.PlayerBJBeatsDealerBJ {
		return GameResultWin
	}

	return GameResultDraw
}

// payoutHand 共通ハンド精算ロジック
// fromSplit=true の場合はスプリット由来のためBJ 3:2配当なし
func payoutHand(player *BlackJackPlayer, hand *BlackJackHand, fromSplit bool, result GameResult) {
	bet := hand.GetBet()
	switch result {
	case GameResultWin:
		if hand.IsBlackJack() && !fromSplit {
			player.AddChips(bet + bet*3/2)
		} else {
			player.AddChips(bet * 2)
		}
	case GameResultDraw:
		player.AddChips(bet)
	case GameResultLose:
		// 没収（何もしない、既に減算済み）
	}
}

// payoutHandWithVariant はバリアント対応の精算ロジック
// 勝利かつバリアントがボーナス対象の場合、通常配当に代えてボーナス倍率で支払う
// (ナチュラルBJ・ダブルダウン後はボーナス対象外)
// 戻り値: 適用されたボーナス (nil = ボーナス未適用)
func (b *BlackJack) payoutHandWithVariant(player *BlackJackPlayer, hand *BlackJackHand, fromSplit bool, result GameResult) *BJBonusPayout {
	if result == GameResultWin && b.variant != nil && b.variant.BonusEval != nil &&
		!hand.IsBlackJack() && !hand.IsDoubled() {
		if bonus := b.variant.BonusEval(hand, b.dealer.GetCard(0)); bonus != nil {
			bet := hand.GetBet()
			// Num:Den の利益 (例 3:2 → bet*3/2) + ベット返却
			player.AddChips(bet + bet*bonus.MultiplierNum/bonus.MultiplierDen)
			return bonus
		}
	}
	payoutHand(player, hand, fromSplit, result)
	return nil
}

// judgeHand 個別ハンドの勝敗判定（人間プレイヤー用: スプリット由来ならBJ抑制）
func (b *BlackJack) judgeHand(hand *BlackJackHand) GameResult {
	return b.judgeHandCore(hand, hand.IsFromSplit())
}

// GameJudgment ゲーム勝敗判定（後方互換: ハンド0の結果を返す）
func (b *BlackJack) GameJudgment() GameResult {
	return b.GameJudgmentForHand(0)
}

// GameJudgmentForHand 指定ハンドの勝敗判定
func (b *BlackJack) GameJudgmentForHand(handIdx int) GameResult {
	if handIdx < 0 || handIdx >= len(b.playerHands) {
		return GameResultLose
	}
	return b.judgeHand(b.playerHands[handIdx])
}

// resolvePayouts 全ハンドの精算
func (b *BlackJack) resolvePayouts() {
	b.bonusKeys = nil // 当ラウンド分のボーナスを再集計
	dealerScore := b.dealer.GetScore()
	dealerBJ := b.dealer.GetCardsSize() == 2 && dealerScore == 21

	// インシュランスの精算
	if b.insuranceBet > 0 {
		if dealerBJ {
			// ディーラーがBJなのでインシュランス勝ち（2:1配当 = 元本+2倍）
			b.player.AddChips(b.insuranceBet * 3)
		}
		// ディーラーがBJでなければインシュランスは没収（何もしない）
	}

	for i, hand := range b.playerHands {
		if hand.IsSurrendered() {
			// サレンダー: 半額返却済み（PlayerSurrender内で処理）
			continue
		}
		result := b.judgeHand(hand)
		bonus := b.payoutHandWithVariant(b.player, hand, hand.IsFromSplit(), result)
		if bonus != nil {
			b.appendLog(i, "bonus", bonus.NameKey, nil)
			b.bonusKeys = append(b.bonusKeys, bonus.NameKey)
		}
	}
}

// checkNaturalBlackJack ナチュラルBJチェック（ディール直後）
func (b *BlackJack) checkNaturalBlackJack() {
	dealerBJ := b.dealer.GetCardsSize() == 2 && b.dealer.GetScore() == 21

	// ディーラーBJなら即終了
	if dealerBJ {
		b.endGame()
		return
	}

	// 全プレイヤーハンドがBJか確認
	allPlayerBJ := true
	for _, hand := range b.playerHands {
		if !hand.IsBlackJack() {
			allPlayerBJ = false
			break
		}
	}

	if allPlayerBJ {
		// 全ハンドBJなら即終了
		b.endGame()
		return
	}

	// 一部ハンドがBJの場合、そのハンドを自動スタンドして続行
	for _, hand := range b.playerHands {
		if hand.IsBlackJack() {
			hand.SetStood(true)
		}
	}

	// advanceHand で最初の未完了ハンドへ進む（全完了ならディーラーターン）
	b.advanceHand()
}

// evaluateSideBets サイドベットの判定と精算
func (b *BlackJack) evaluateSideBets() {
	b.sideBetResults = nil
	card1 := b.playerHands[0].GetCard(0)
	card2 := b.playerHands[0].GetCard(1)
	dealerUpcard := b.dealer.GetCard(0)

	if b.perfectPairsBet > 0 {
		resultType, resultName := EvaluatePerfectPairs(card1, card2)
		payout := 0
		if resultType != BJPPNone {
			multiplier := PerfectPairsPayout(resultType)
			payout = b.perfectPairsBet * multiplier
			b.player.AddChips(b.perfectPairsBet + payout)
		}
		b.sideBetResults = append(b.sideBetResults, &BJSideBetResult{
			BetType:    BJSideBetPerfectPairs,
			ResultType: resultType,
			ResultName: resultName,
			BetAmount:  b.perfectPairsBet,
			Payout:     payout,
		})
	}

	if b.twentyOnePlus3Bet > 0 {
		resultType, resultName := Evaluate21Plus3(card1, card2, dealerUpcard)
		payout := 0
		if resultType != BJT3None {
			multiplier := TwentyOnePlus3Payout(resultType)
			payout = b.twentyOnePlus3Bet * multiplier
			b.player.AddChips(b.twentyOnePlus3Bet + payout)
		}
		b.sideBetResults = append(b.sideBetResults, &BJSideBetResult{
			BetType:    BJSideBet21Plus3,
			ResultType: resultType,
			ResultName: resultName,
			BetAmount:  b.twentyOnePlus3Bet,
			Payout:     payout,
		})
	}
}

// GetSideBetResults サイドベット結果取得
func (b *BlackJack) GetSideBetResults() []*BJSideBetResult {
	return b.sideBetResults
}

// hiLoValue Hi-Loカウンティングのカード値 (2-6: +1, 7-9: 0, 10/J/Q/K/A: -1)
func hiLoValue(card *Card) int {
	if card == nil {
		return 0
	}
	v := card.GetValue()
	switch {
	case v >= 2 && v <= 6:
		return 1
	case v >= 7 && v <= 9:
		return 0
	default: // 1(A), 10, 11(J), 12(Q), 13(K)
		return -1
	}
}

// koValue KOカウンティングのカード値 (2-7: +1, 8-9: 0, 10/J/Q/K/A: -1)
func koValue(card *Card) int {
	if card == nil {
		return 0
	}
	v := card.GetValue()
	switch {
	case v >= 2 && v <= 7:
		return 1
	case v >= 8 && v <= 9:
		return 0
	default: // 1(A), 10, 11(J), 12(Q), 13(K)
		return -1
	}
}

// zenValue Zen Countカウンティングのカード値 (2-3: +1, 4-6: +2, 7: +1, 8: 0, 9: -1, 10/J/Q/K: -2, A: -1)
func zenValue(card *Card) int {
	if card == nil {
		return 0
	}
	v := card.GetValue()
	switch {
	case v >= 2 && v <= 3:
		return 1
	case v >= 4 && v <= 6:
		return 2
	case v == 7:
		return 1
	case v == 8:
		return 0
	case v == 9:
		return -1
	case v == 1: // A
		return -1
	default: // 10, 11(J), 12(Q), 13(K)
		return -2
	}
}

// omegaIIValue Omega IIカウンティングのカード値 (2-3: +1, 4-6: +2, 7: +1, 8: 0, 9: -1, 10/J/Q/K: -2, A: 0)
func omegaIIValue(card *Card) int {
	if card == nil {
		return 0
	}
	v := card.GetValue()
	switch {
	case v >= 2 && v <= 3:
		return 1
	case v >= 4 && v <= 6:
		return 2
	case v == 7:
		return 1
	case v == 8:
		return 0
	case v == 9:
		return -1
	case v == 1: // A
		return 0
	default: // 10, 11(J), 12(Q), 13(K)
		return -2
	}
}

// countingValue 指定カウンティングシステムでのカード値を返す
func countingValue(card *Card, system int) int {
	switch system {
	case BJCountingKO:
		return koValue(card)
	case BJCountingZen:
		return zenValue(card)
	case BJCountingOmegaII:
		return omegaIIValue(card)
	default: // BJCountingHiLo
		return hiLoValue(card)
	}
}

// updateRunningCount ランニングカウントを更新（countingEnabled時のみ）
func (b *BlackJack) updateRunningCount(card *Card) {
	if !b.config.CountingEnabled {
		return
	}
	b.runningCount += countingValue(card, b.config.CountingSystem)
}

// GetRunningCount ランニングカウント取得
func (b *BlackJack) GetRunningCount() int {
	return b.runningCount
}

// GetTrueCount トゥルーカウント取得 (アンバランスドシステム(KO)では0を返す)
func (b *BlackJack) GetTrueCount() float64 {
	if !IsBalancedCountingSystem(b.config.CountingSystem) {
		return 0
	}
	if b.trumpCards == nil {
		return 0
	}
	remaining := b.trumpCards.GetRemainingCount()
	decksRemaining := float64(remaining) / 52.0
	if decksRemaining < 1.0 {
		decksRemaining = 1.0
	}
	return float64(b.runningCount) / decksRemaining
}

// IsCountingEnabled カウンティング有効か
func (b *BlackJack) IsCountingEnabled() bool {
	return b.config.CountingEnabled
}

// GetBasicStrategySuggestion ベーシックストラテジーによる推奨アクションを返す
func (b *BlackJack) GetBasicStrategySuggestion() BJSuggestedAction {
	if !b.hintEnabled {
		return BJSuggestNone
	}
	if b.phase != BJPhaseAction && b.phase != BJPhaseInsurance && b.phase != BJPhaseEarlySurrender {
		return BJSuggestNone
	}
	if b.phase == BJPhaseInsurance {
		return BJSuggestDeclineInsurance
	}
	hand := b.playerHands[b.currentHandIdx]
	if hand.IsFinished() {
		return BJSuggestNone
	}
	dealerUpcard := b.dealer.GetCard(0)
	if dealerUpcard == nil {
		return BJSuggestNone
	}
	// **バリアントを渡す。**48枚デッキのスパニッシュ21に標準デッキの基本戦略を
	// 当てると、10 が抜けている分だけ助言がずれる (#4705)。
	if b.phase == BJPhaseEarlySurrender {
		action := GetVariantStrategyAction(hand, dealerUpcard, b.config.DealerHitsSoft17, b.config.Variant)
		if action == BJSuggestSurrender {
			return BJSuggestSurrender
		}
		return BJSuggestStand // "continue" = decline early surrender
	}
	return GetVariantStrategyAction(hand, dealerUpcard, b.config.DealerHitsSoft17, b.config.Variant)
}
