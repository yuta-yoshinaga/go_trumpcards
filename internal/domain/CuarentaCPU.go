//go:build !js || !wasm || extra2

package domain

import "math/rand"

// cuarentaRankMatchIndexes は playedCard と同じランク (GetValue) を持つ
// 場札のインデックスを全て返す。クアレンタの捕獲は純粋なランク一致であり、
// 一致するカードは全て同時に捕獲される。
// CuarentaCaptureCount は playedCard を出したときに捕獲できる場札の枚数を返す。
//
// **同ランクの場札はまとめて捕獲する** (applyPlay)。画面はどの札で何枚取れるかを
// 出すが、判定を書き写すと実際の捕獲とずれるので、同じ関数から数える (#5673)。
func CuarentaCaptureCount(playedCard *Card, tableCards []*Card) int {
	return len(cuarentaRankMatchIndexes(playedCard, tableCards))
}

func cuarentaRankMatchIndexes(playedCard *Card, tableCards []*Card) []int {
	if playedCard == nil {
		return nil
	}
	target := playedCard.GetValue()
	out := make([]int, 0)
	for i, c := range tableCards {
		if c != nil && c.GetValue() == target {
			out = append(out, i)
		}
	}
	return out
}

// chooseCpuPlay は難易度に応じて CPU が出す手札インデックスを決定する。
//
//	Easy:   ランダムな手札。
//	Normal: 捕獲できる手を優先し、最も得点価値の高い手を選ぶ。
//	        捕獲できなければ最も価値の低いカードを場に置く。
//	Hard:   Normal に加え、caída を誘発しにくいよう場に残す価値を減点。
func (g *Cuarenta) chooseCpuPlay(playerIdx int) int {
	player := g.players[playerIdx]
	size := player.GetCardsSize()
	if size == 0 {
		return 0
	}
	if g.config.CpuDifficulty == CuarentaDifficultyEasy {
		return rand.Intn(size)
	}

	bestIdx := 0
	bestScore := g.scoreCpuCard(player.GetCard(0))
	for i := 1; i < size; i++ {
		sc := g.scoreCpuCard(player.GetCard(i))
		if sc > bestScore {
			bestScore = sc
			bestIdx = i
		}
	}
	return bestIdx
}

// scoreCpuCard は手札 1 枚を出したときの期待評価値。
func (g *Cuarenta) scoreCpuCard(card *Card) int {
	if card == nil {
		return -1000
	}
	matches := cuarentaRankMatchIndexes(card, g.round.tableCards)
	if len(matches) == 0 {
		// 捕獲なし: 価値の低いカードを置きたいのでマイナス評価。
		base := -cuarentaCardValueScore(card)
		if g.config.CpuDifficulty == CuarentaDifficultyHard {
			// 直後の相手に caída を取られにくいよう、高ランクを場に残さない。
			base -= card.GetValue() / 4
		}
		return base
	}
	// 捕獲: 基本点 + 取れる枚数 + caída / ronda / limpia ボーナス見込み。
	score := 5 + len(matches)*2
	if g.round.lastLaidCard != nil && g.round.lastLaidCard.GetValue() == card.GetValue() {
		score += CuarentaScoreCaida * 3 // caída を強く優先
	}
	total := len(matches) + 1
	if total >= 3 {
		score += (total - 2) * CuarentaScoreRondaPerExtra * 2
	}
	if len(matches) == len(g.round.tableCards) {
		score += CuarentaScoreLimpia * 3 // limpia を強く優先
	}
	return score
}

// cuarentaCardValueScore はカードの "得点価値" を返す (CPU 評価用)。
// 高ランクほど (相手に取られると痛いので) 価値が高い。
func cuarentaCardValueScore(c *Card) int {
	if c == nil {
		return 0
	}
	return c.GetValue()
}
