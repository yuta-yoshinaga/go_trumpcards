package domain

// PresidentCardStrength カードの強さを返す (3が最弱、2が最強)
// 3 < 4 < 5 < 6 < 7 < 8 < 9 < 10 < J(11) < Q(12) < K(13) < A(1) < 2(2)
func PresidentCardStrength(v int) int {
	if v == 1 {
		return 14 // Ace
	}
	if v == 2 {
		return 15 // 2 は最強
	}
	return v
}

// PresidentCardStrengthRevolution 革命中のカードの強さを返す (2が最弱、3が最強)
func PresidentCardStrengthRevolution(v int) int {
	return 18 - PresidentCardStrength(v)
}

// cardStrength 現在の革命状態に応じたカード値の強さを返す
func (p *President) cardStrength(v int) int {
	if p.round.revolutionActive {
		return PresidentCardStrengthRevolution(v)
	}
	return PresidentCardStrength(v)
}

// cardStrengthForCard カードオブジェクトの強さを返す
func (p *President) cardStrengthForCard(card *Card) int {
	return p.cardStrength(card.GetValue())
}

// isValidGroup カード配列がグループ (同じ値) として有効かチェック
// プレジデントでは Singles/Pairs/Triples/Quads のみ
func isValidPresidentGroup(cards []*Card) bool {
	if len(cards) == 0 {
		return false
	}
	base := cards[0].GetValue()
	for _, c := range cards[1:] {
		if c.GetValue() != base {
			return false
		}
	}
	return true
}

// isPlayable 指定したカードが場のカードに対して出せるか判定
func (p *President) isPlayable(cards []*Card) bool {
	if len(cards) == 0 {
		return false
	}
	if len(cards) > 4 {
		return false
	}
	if !isValidPresidentGroup(cards) {
		return false
	}
	// 場がクリアなら何でも出せる
	if p.round.tableCards == nil {
		return true
	}
	// 枚数が一致しているか
	if len(cards) != len(p.round.tableCards) {
		return false
	}
	// 強さ比較
	tableStrength := p.cardStrength(p.round.tableCards[0].GetValue())
	playStrength := p.cardStrength(cards[0].GetValue())
	return playStrength > tableStrength
}
