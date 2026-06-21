//go:build !js || !wasm || classic

package domain

import "math/rand"

// scopaPlan は CPU が選んだ 1 手の詳細。
type scopaPlan struct {
	handIdx   int
	tableIdxs []int // 空なら場に置く
}

// chooseCpuPlay は難易度に応じて CPU の 1 手を決定する。
// Easy: ランダムな合法手。
// Normal: 捕獲を優先し、最も得点価値の高い捕獲を選ぶ。なければ最も無難なカードを置く。
// Hard: Normal + 場に置く際に相手へ高得点カードを残さないよう減点。
func (s *Scopa) chooseCpuPlay(playerIdx int) scopaPlan {
	player := s.players[playerIdx]
	hand := make([]*Card, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		hand[i] = player.GetCard(i)
	}
	if len(hand) == 0 {
		return scopaPlan{handIdx: 0}
	}
	plans := s.enumerateCpuPlays(hand)
	if len(plans) == 0 {
		return scopaPlan{handIdx: 0}
	}
	if s.config.CpuDifficulty == ScopaDifficultyEasy {
		return plans[rand.Intn(len(plans))]
	}
	best := plans[0]
	bestScore := s.scoreCpuPlay(hand, plans[0])
	for _, p := range plans[1:] {
		sc := s.scoreCpuPlay(hand, p)
		if sc > bestScore {
			bestScore = sc
			best = p
		}
	}
	return best
}

// enumerateCpuPlays は手札全体からあらゆる合法手を列挙する。
// 捕獲が可能なカードについては各捕獲候補を、捕獲できないカードについては
// 「場に置く」プランを生成する。捕獲が必須のルールに合わせ、捕獲できる
// カードに対しては「場に置く」プランは生成しない。
func (s *Scopa) enumerateCpuPlays(hand []*Card) []scopaPlan {
	plans := make([]scopaPlan, 0)
	for handIdx, card := range hand {
		if card == nil {
			continue
		}
		captures := EnumerateScopaCaptures(card, s.round.tableCards)
		if len(captures) > 0 {
			for _, cap := range captures {
				plans = append(plans, scopaPlan{
					handIdx:   handIdx,
					tableIdxs: append([]int(nil), cap...),
				})
			}
			continue
		}
		plans = append(plans, scopaPlan{handIdx: handIdx})
	}
	return plans
}

// scoreCpuPlay はプランの期待得点。捕獲は得点価値の高いカードを優先。
// 場に置く手は残す手札価値のマイナス (Hard はさらに相手への漏れを減点)。
func (s *Scopa) scoreCpuPlay(hand []*Card, plan scopaPlan) int {
	card := hand[plan.handIdx]
	if len(plan.tableIdxs) == 0 {
		// 場に置く: 価値の低いカードを置きたいのでマイナス評価。
		base := -scopaCardValueScore(card)
		if s.config.CpuDifficulty == ScopaDifficultyHard {
			// 相手が取りやすい高得点カードを場に残すのを避ける。
			base -= scopaTrailPenalty(card)
		}
		return base
	}
	sc := 2 // 捕獲自体の基本点
	sc += scopaCardValueScore(card)
	used := make(map[int]bool, len(plan.tableIdxs))
	for _, idx := range plan.tableIdxs {
		used[idx] = true
		if idx >= 0 && idx < len(s.round.tableCards) {
			sc += scopaCardValueScore(s.round.tableCards[idx])
		}
	}
	// スコパ推定 (場札が全て捌けるか)。
	remain := 0
	for i := range s.round.tableCards {
		if !used[i] {
			remain++
		}
	}
	if remain == 0 {
		sc += 5
	}
	return sc
}

// scopaCardValueScore はカードの "得点価値" を返す (CPU 評価用)。
func scopaCardValueScore(c *Card) int {
	if c == nil {
		return 0
	}
	score := 1 // 1 枚で最多カードボーナスへ寄与
	if ScopaIsSetteBello(c) {
		score += 6
	} else if ScopaIsSeven(c) {
		score += 3
	}
	if ScopaIsDiamond(c) {
		score += 2
	}
	return score
}

// scopaTrailPenalty は場に置いた場合に相手へ与えうる価値の減点。
func scopaTrailPenalty(c *Card) int {
	if c == nil {
		return 0
	}
	penalty := 0
	if ScopaIsSetteBello(c) {
		penalty += 5
	} else if ScopaIsSeven(c) {
		penalty += 2
	}
	if ScopaIsDiamond(c) {
		penalty++
	}
	return penalty
}
