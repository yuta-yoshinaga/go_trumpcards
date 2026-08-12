//go:build test

package domain

// chemindeFerTestHand は合計が total になる 2 枚を返す。
//
// 10/J/Q/K が 0 点なので、**1 枚を K に固定すれば残り 1 枚で合計を作れる**。
func chemindeFerTestHand(total int) []*Card {
	second := total
	if total == 0 {
		second = 10 // 0 点の札。値 0 の札は存在しない。
	}
	return []*Card{NewCard(CardDesignSpade, 13, false), NewCard(CardDesignSpade, second, false)}
}

// SetupCoupForTest は「席 1 が 200 を賭け、指定の合計が配られた」局面を組み立てる。
//
// **配りは乱数のままでは固定できない** (TrumpCards.Shuffle はグローバルな rand を
// 使う) ので、盤面を直接組む。PlaceBet 経由で配らせると決着まで走り切ってしまい、
// 後から手札を差し替えると同じクーが 2 度精算される。
func (g *ChemindeFer) SetupCoupForTest(punterTotal, bankerTotal int, phase ChemindeFerPhase) {
	// **チップを初期値へ戻してから積む。** Reset は人間の番まで CPU を進めるので、
	// 呼び出し元がすでに 1 ラウンド遊んでいることがある。
	for _, p := range g.players {
		p.SetChips(g.config.InitialChips)
		p.SetBet(0)
	}
	const bet = 200
	g.stake = bet
	g.betOrder = nil
	g.betPos = -1
	const seat = 1
	g.players[seat].SubtractChips(bet)
	g.players[seat].SetBet(bet)
	g.represIdx = seat
	g.punterHand = chemindeFerTestHand(punterTotal)
	g.bankerHand = chemindeFerTestHand(bankerTotal)
	g.punterDrew = false
	g.result = ChemindeFerResultNone
	g.phase = phase
}

// StakeForTest は張りだけを行い、**CPU を進めない**。
//
// 公開の SetStake は次に人間の番が来るまで走り切るので、子が全員 CPU の卓では
// 1 クー遊び切ってしまう。賭けの途中を見たいテスト (別パッケージのプレゼンタ側)
// のための入口。
func (g *ChemindeFer) StakeForTest(amount int) error { return g.setStake(amount) }
