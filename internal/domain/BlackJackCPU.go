package domain

// initCpuPlayers CPUプレイヤー初期化
func (b *BlackJack) initCpuPlayers() {
	count := b.config.CpuPlayerCount
	if count <= 0 {
		b.cpuPlayers = nil
		return
	}
	// 既存CPUプレイヤーを再利用（チップ引き継ぎ）
	if len(b.cpuPlayers) == count {
		for _, cpu := range b.cpuPlayers {
			cpu.Reset()
			if cpu.GetPlayer().GetChips() < BJMinBet {
				cpu.GetPlayer().SetChips(BJDefaultChips)
			}
		}
		return
	}
	b.cpuPlayers = make([]*BlackJackCpuSeat, count)
	for i := 0; i < count; i++ {
		b.cpuPlayers[i] = NewBlackJackCpuSeat()
	}
}

// cpuInsurance CPUプレイヤーのインシュランス判定
func (b *BlackJack) cpuInsurance() {
	if !b.config.CountingEnabled {
		return
	}
	for _, cpu := range b.cpuPlayers {
		if cpu.GetHands()[0].GetCardsSize() == 0 {
			continue
		}
		if !ShouldTakeInsurance(b.GetTrueCount(), b.GetRunningCount(), b.config.CountingSystem) {
			continue
		}
		cost := cpu.GetHands()[0].GetBet() / 2
		if !cpu.GetPlayer().SubtractChips(cost) {
			continue
		}
		cpu.SetInsuranceBet(cost)
	}
}

// cpuBetAndDeal CPUプレイヤーの自動ベットとカード配布
func (b *BlackJack) cpuBetAndDeal() {
	for _, cpu := range b.cpuPlayers {
		var betAmount int
		if b.config.CountingEnabled {
			betAmount = GetCountingBetAmount(b.GetTrueCount(), b.GetRunningCount(), b.config.CountingSystem, cpu.GetPlayer().GetChips())
		} else {
			betAmount = BJCpuBetAmount
			if cpu.GetPlayer().GetChips() < betAmount {
				betAmount = cpu.GetPlayer().GetChips()
			}
			// ベット額をBJMinBetの倍数に丸める（GetCountingBetAmountは内部で丸め済み）
			betAmount = (betAmount / BJMinBet) * BJMinBet
		}
		if betAmount < BJMinBet {
			continue
		}
		cpu.GetPlayer().SubtractChips(betAmount)
		hand := cpu.GetHands()[0]
		hand.SetBet(betAmount)
		// カードを2枚配る
		dealFailed := false
		for i := 0; i < 2; i++ {
			card := b.drawCard()
			if card != nil {
				hand.AddCard(card)
			} else {
				dealFailed = true
				break
			}
		}
		// 山札枯渇で必要な枚数を配れなかった場合、ベットを返却してリセット
		if dealFailed {
			cpu.GetPlayer().AddChips(betAmount)
			hand.Reset()
		}
	}
}

// cpuPlay CPUプレイヤーのベーシックストラテジープレイ
func (b *BlackJack) cpuPlay() {
	dealerUpcard := b.dealer.GetCard(0)
	if dealerUpcard == nil {
		return
	}
	for _, cpu := range b.cpuPlayers {
		b.cpuPlaySeat(cpu, dealerUpcard)
	}
}

// cpuPlaySeat 個別CPUプレイヤーのプレイ
func (b *BlackJack) cpuPlaySeat(cpu *BlackJackCpuSeat, dealerUpcard *Card) {
	handIdx := 0
	for handIdx < len(cpu.GetHands()) {
		hand := cpu.GetHands()[handIdx]
		if hand.IsFinished() || hand.GetCardsSize() == 0 {
			handIdx++
			continue
		}
		for !hand.IsFinished() {
			action := GetBasicStrategyAction(hand, dealerUpcard, b.config.DealerHitsSoft17)
			switch action {
			case BJSuggestHit:
				b.cpuHit(hand)
			case BJSuggestStand:
				hand.SetStood(true)
			case BJSuggestDouble:
				canDD := hand.GetCardsSize() == 2 && cpu.GetPlayer().GetChips() >= hand.GetBet()
				if canDD && (!hand.IsFromSplit() || b.config.DoubleAfterSplit) {
					b.cpuDoubleDown(cpu, hand)
				} else {
					b.cpuHit(hand)
				}
			case BJSuggestDoubleStand:
				canDD := hand.GetCardsSize() == 2 && cpu.GetPlayer().GetChips() >= hand.GetBet()
				if canDD && (!hand.IsFromSplit() || b.config.DoubleAfterSplit) {
					b.cpuDoubleDown(cpu, hand)
				} else {
					hand.SetStood(true)
				}
			case BJSuggestSplit:
				if hand.CanSplit() && len(cpu.GetHands()) < BJMaxHands && cpu.GetPlayer().GetChips() >= hand.GetBet() {
					b.cpuSplit(cpu, hand, handIdx, dealerUpcard)
					continue // cpuSplit may add hands, re-check current index
				}
				b.cpuHit(hand)
			case BJSuggestSurrender:
				if hand.CanSurrender() && b.config.SurrenderRule != BJSurrenderNone {
					halfBet := hand.GetBet() / 2
					cpu.GetPlayer().AddChips(halfBet)
					hand.SetSurrendered(true)
				} else {
					b.cpuHit(hand)
				}
			}
		}
		handIdx++
	}
}

// cpuHit CPUヒット
func (b *BlackJack) cpuHit(hand *BlackJackHand) {
	card := b.drawCard()
	if card == nil {
		hand.SetStood(true)
		return
	}
	hand.AddCard(card)
	b.updateRunningCount(card)
	if hand.GetScore() >= 22 {
		hand.SetBusted(true)
	}
}

// cpuDoubleDown CPUダブルダウン
func (b *BlackJack) cpuDoubleDown(cpu *BlackJackCpuSeat, hand *BlackJackHand) {
	bet := hand.GetBet()
	cpu.GetPlayer().SubtractChips(bet)
	hand.SetBet(bet * 2)
	hand.SetDoubled(true)
	card := b.drawCard()
	if card == nil {
		// デッキ枯渇: 元に戻す
		cpu.GetPlayer().AddChips(bet)
		hand.SetBet(bet)
		hand.SetDoubled(false)
		hand.SetStood(true)
		return
	}
	hand.AddCard(card)
	b.updateRunningCount(card)
	if hand.GetScore() >= 22 {
		hand.SetBusted(true)
	} else {
		hand.SetStood(true)
	}
}

// cpuSplit CPUスプリット
func (b *BlackJack) cpuSplit(cpu *BlackJackCpuSeat, hand *BlackJackHand, handIdx int, dealerUpcard *Card) {
	bet := hand.GetBet()
	cpu.GetPlayer().SubtractChips(bet)

	firstCard := hand.GetCard(0)
	secondCard := hand.GetCard(1)
	hand.Reset()
	hand.SetBet(bet)
	hand.SetFromSplit(true)
	hand.AddCard(firstCard)

	newHand := NewBlackJackHand()
	newHand.SetBet(bet)
	newHand.SetFromSplit(true)
	newHand.AddCard(secondCard)

	// 各ハンドに1枚ずつ配る
	card1 := b.drawCard()
	var card2 *Card
	if card1 != nil {
		hand.AddCard(card1)
		b.updateRunningCount(card1)
		card2 = b.drawCard()
	}

	if card1 == nil || card2 == nil {
		// デッキ枯渇: 元のハンドを復元してベットを返却
		hand.Reset()
		hand.SetBet(bet)
		hand.AddCard(firstCard)
		hand.AddCard(secondCard)
		cpu.GetPlayer().AddChips(bet)
		if card1 != nil {
			// card1のランニングカウント更新を元に戻す
			b.runningCount -= countingValue(card1, b.config.CountingSystem)
		}
		hand.SetStood(true)
		return
	}
	newHand.AddCard(card2)
	b.updateRunningCount(card2)

	// 新しいハンドを挿入
	hands := cpu.GetHands()
	newHands := make([]*BlackJackHand, 0, len(hands)+1)
	newHands = append(newHands, hands[:handIdx+1]...)
	newHands = append(newHands, newHand)
	newHands = append(newHands, hands[handIdx+1:]...)
	cpu.SetHands(newHands)

	// エースのスプリットの場合、両ハンドを自動スタンド
	if firstCard.GetValue() == 1 {
		hand.SetStood(true)
		newHand.SetStood(true)
	}
}

// resolvePayoutsCpu CPUプレイヤーの精算
func (b *BlackJack) resolvePayoutsCpu() {
	dealerScore := b.dealer.GetScore()
	dealerBJ := b.dealer.GetCardsSize() == 2 && dealerScore == 21

	for cpuIdx, cpu := range b.cpuPlayers {
		// CPUインシュランスの精算
		if cpu.GetInsuranceBet() > 0 {
			if dealerBJ {
				cpu.GetPlayer().AddChips(cpu.GetInsuranceBet() * 3)
			}
		}
		// CPU action log uses positive indices starting at 1 (0 is reserved for the human)
		cpuLogIdx := cpuIdx + 1
		for _, hand := range cpu.GetHands() {
			if hand.GetCardsSize() == 0 {
				continue
			}
			if hand.IsSurrendered() {
				continue
			}
			result := b.judgeCpuHand(hand)
			bonus := b.payoutHandWithVariant(cpu.GetPlayer(), hand, hand.IsFromSplit(), result)
			if bonus != nil {
				b.appendLog(cpuLogIdx, "bonus", bonus.NameKey, nil)
			}
		}
	}
}

// judgeCpuHand CPU個別ハンドの勝敗判定（fromSplit追跡による正確なBJ判定）
func (b *BlackJack) judgeCpuHand(hand *BlackJackHand) GameResult {
	return b.judgeHandCore(hand, hand.IsFromSplit())
}

// cpuEarlySurrender CPUプレイヤーのアーリーサレンダー判定
func (b *BlackJack) cpuEarlySurrender() {
	dealerUpcard := b.dealer.GetCard(0)
	if dealerUpcard == nil {
		return
	}
	for _, cpu := range b.cpuPlayers {
		for _, hand := range cpu.GetHands() {
			if hand.GetCardsSize() == 0 || hand.IsFinished() {
				continue
			}
			action := GetBasicStrategyAction(hand, dealerUpcard, b.config.DealerHitsSoft17)
			if action == BJSuggestSurrender {
				halfBet := hand.GetBet() / 2
				cpu.GetPlayer().AddChips(halfBet)
				hand.SetSurrendered(true)
			}
		}
	}
}
