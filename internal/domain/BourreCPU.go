//go:build !js || !wasm || casino

package domain

import "math/rand"

// cpuShouldPlay CPUがこのハンドに参加すべきかを判定する。
// 推定獲得トリック数が難易度別のしきい値以上なら参加する。
func (b *Bourre) cpuShouldPlay(idx int) bool {
	est := b.estimateTricks(idx)
	var threshold float64
	switch b.config.CpuDifficulty {
	case BourreDifficultyHard:
		threshold = 1.3 // 慎重: ブーレ回避を優先
	case BourreDifficultyEasy:
		threshold = 0.5 // ルーズ: 弱くても参加しがち
	default:
		threshold = 1.0
	}
	return est >= threshold
}

// estimateTricks 手札と切り札から推定獲得トリック数を概算する
func (b *Bourre) estimateTricks(idx int) float64 {
	p := b.players[idx]
	score := 0.0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		r := bourreRank(c)
		if c.GetDesign() == b.trumpSuit {
			score += 0.5
			if r >= 12 { // 切り札の Q/K/A
				score += 0.4
			}
		} else if r == 14 { // 他スートのエース
			score += 0.6
		} else if r == 13 { // 他スートのキング
			score += 0.25
		}
	}
	return score
}

// cpuSelectDiscards CPUが交換する不要カードのインデックスを選ぶ。
// 切り札と高位札 (Q以上) は残し、それ以外の低位札を捨てる。
func (b *Bourre) cpuSelectDiscards(idx int) []int {
	p := b.players[idx]
	discards := make([]int, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == b.trumpSuit {
			continue
		}
		if bourreRank(c) >= 12 {
			continue
		}
		discards = append(discards, i)
	}
	return discards
}

// cpuSelectPlay CPUがプレイするカードのインデックスを選ぶ
func (b *Bourre) cpuSelectPlay(idx int) int {
	valid := b.legalPlays(idx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if b.config.CpuDifficulty == BourreDifficultyEasy {
		return valid[rand.Intn(len(valid))] //nolint:gosec // non-crypto random for game AI
	}
	p := b.players[idx]
	if len(b.currentTrick) == 0 {
		// リード: トリックを取りに行くため最も高いカードを出す
		return b.pickByRank(p, valid, true)
	}
	// フォロー: legalPlays が勝ち札集合を返している場合は最小の勝ち札、
	// そうでなければ温存のため最小のカードを出す
	return b.pickByRank(p, valid, false)
}

// pickByRank valid の中から最高 (highest=true) または最低ランクのインデックスを返す
func (b *Bourre) pickByRank(p *BourrePlayer, valid []int, highest bool) int {
	best := valid[0]
	bestRank := bourreRank(p.GetCard(best))
	for _, i := range valid[1:] {
		r := bourreRank(p.GetCard(i))
		if (highest && r > bestRank) || (!highest && r < bestRank) {
			best = i
			bestRank = r
		}
	}
	return best
}
