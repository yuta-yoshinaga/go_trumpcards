//go:build !js || !wasm || casino

package domain

// PokerDrawOdds ドローオッズ (各ハンドランクの確率)
type PokerDrawOdds struct {
	HandRank    int     // ハンドランク
	HandName    string  // ハンド名
	Probability float64 // 確率 (0.0 - 1.0)
	Count       int     // 該当する組み合わせ数
	Total       int     // 全組み合わせ数
}

// CalcDrawOdds 交換候補のカードインデックスに基づくドローオッズを計算する
// indices: 交換するカードのインデックス (0-4)
func (p *Poker) CalcDrawOdds(indices []int) ([]PokerDrawOdds, error) {
	if p.round.phase != PokerPhaseExchange {
		return nil, NewDomainError(ErrWrongPhase, "Odds calculation is only available during exchange phase.")
	}
	if !p.players[p.round.currentTurn].GetIsHuman() {
		return nil, NewDomainError(ErrNotHumanTurn, "It is not your turn.")
	}

	// インデックスバリデーション
	seen := make(map[int]bool)
	for _, idx := range indices {
		if idx < 0 || idx >= p.players[p.round.currentTurn].GetCardsSize() {
			return nil, NewDomainError(ErrInvalidIndices, "Card index out of range.")
		}
		if seen[idx] {
			return nil, NewDomainError(ErrInvalidIndices, "Duplicate card index.")
		}
		seen[idx] = true
	}

	human := p.players[p.round.currentTurn]
	hand := make([]*Card, human.GetCardsSize())
	for i := 0; i < human.GetCardsSize(); i++ {
		hand[i] = human.GetCard(i)
	}

	k := len(indices)

	// 交換なし (stand) → 現在のハンドランクが100%
	if k == 0 {
		rank := evalFiveCardHandWithJokers(hand)
		return buildOddsResult(rank, 1), nil
	}

	// 未知カードプールを構築
	pool := buildUnknownCards(hand, p.config.JokerCount)

	// 全組み合わせを列挙して各ハンドランクをカウント
	counts := make([]int, len(PokerHandNames))
	total := 0
	testHand := make([]*Card, len(hand))

	pokerCombinations(pool, k, func(combo []*Card) {
		// 交換後のハンドを構築
		copy(testHand, hand)
		for i, idx := range indices {
			testHand[idx] = combo[i]
		}
		rank := evalFiveCardHandWithJokers(testHand)
		counts[rank]++
		total++
	})

	// 結果を構築
	result := make([]PokerDrawOdds, len(PokerHandNames))
	for i := 0; i < len(PokerHandNames); i++ {
		prob := 0.0
		if total > 0 {
			prob = float64(counts[i]) / float64(total)
		}
		result[i] = PokerDrawOdds{
			HandRank:    i,
			HandName:    PokerHandNames[i],
			Probability: prob,
			Count:       counts[i],
			Total:       total,
		}
	}
	return result, nil
}

// buildOddsResult 交換なしの場合のオッズ結果を構築する
func buildOddsResult(rank int, total int) []PokerDrawOdds {
	result := make([]PokerDrawOdds, len(PokerHandNames))
	for i := 0; i < len(PokerHandNames); i++ {
		prob := 0.0
		count := 0
		if i == rank {
			prob = 1.0
			count = total
		}
		result[i] = PokerDrawOdds{
			HandRank:    i,
			HandName:    PokerHandNames[i],
			Probability: prob,
			Count:       count,
			Total:       total,
		}
	}
	return result
}

// buildUnknownCards 人間の手札を除いた未知カードプールを構築する
// CPU の手札は含む (情報リーク防止のため)
func buildUnknownCards(humanCards []*Card, jokerCount int) []*Card {
	// 人間の手札をセットに変換
	humanSet := make(map[[2]int]bool)
	humanJokerCount := 0
	for _, c := range humanCards {
		if c.GetDesign() == CardDesignJoker {
			humanJokerCount++
		} else {
			humanSet[[2]int{c.GetDesign(), c.GetValue()}] = true
		}
	}

	pool := make([]*Card, 0)
	// 通常カード52枚
	for d := CardDesignSpade; d <= CardDesignDiamond; d++ {
		for v := 1; v <= CardValueMax; v++ {
			if !humanSet[[2]int{d, v}] {
				pool = append(pool, NewCard(d, v, false))
			}
		}
	}
	// ジョーカー (人間が持っていない分)
	remainingJokers := jokerCount - humanJokerCount
	for i := 0; i < remainingJokers; i++ {
		pool = append(pool, NewCard(CardDesignJoker, CardValueJoker, false))
	}

	return pool
}

// pokerCombinations プールからk枚の組み合わせを列挙する (iterative)
func pokerCombinations(pool []*Card, k int, cb func(combo []*Card)) {
	n := len(pool)
	if k <= 0 || k > n {
		return
	}

	indices := make([]int, k)
	for i := 0; i < k; i++ {
		indices[i] = i
	}

	combo := make([]*Card, k)
	for {
		for i := 0; i < k; i++ {
			combo[i] = pool[indices[i]]
		}
		cb(combo)

		// 次の組み合わせを生成
		i := k - 1
		for i >= 0 && indices[i] == n-k+i {
			i--
		}
		if i < 0 {
			break
		}
		indices[i]++
		for j := i + 1; j < k; j++ {
			indices[j] = indices[j-1] + 1
		}
	}
}
