package domain

import (
	"math/rand"
)

// cpuPlan は CPU が選んだアクションの詳細。
type cpuPlan struct {
	Type       CassinoActionType
	handIdx    int
	tableIdxs  []int
	buildIdxs  []int
	buildValue int
}

// chooseCpuAction は難易度に応じて CPU の 1 手を決定する。
// Easy: ランダムな合法手。
// Normal: take を優先、最も得点価値の高い組合せを選ぶ。なければ build か trail。
// Hard: Normal + 相手への得点漏れを回避 (trail で point card を残さない)。
func (c *Cassino) chooseCpuAction(playerIdx int) cpuPlan {
	player := c.players[playerIdx]
	hand := make([]*Card, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		hand[i] = player.GetCard(i)
	}
	if len(hand) == 0 {
		return cpuPlan{Type: CassinoActionTrail, handIdx: 0}
	}
	switch c.config.CpuDifficulty {
	case CassinoDifficultyEasy:
		return c.cpuEasyPlan(playerIdx, hand)
	case CassinoDifficultyHard:
		return c.cpuHardPlan(playerIdx, hand)
	default:
		return c.cpuNormalPlan(playerIdx, hand)
	}
}

// cpuEasyPlan は全ての合法手を列挙してランダムに選ぶ。
func (c *Cassino) cpuEasyPlan(playerIdx int, hand []*Card) cpuPlan {
	plans := c.enumerateCpuPlans(playerIdx, hand)
	if len(plans) == 0 {
		// 通常 enumerateCpuPlans は手札があれば必ず trail を含むので
		// ここは到達しないはず。安全策として最初の有効インデックスを trail。
		return cpuPlan{Type: CassinoActionTrail, handIdx: 0}
	}
	return plans[rand.Intn(len(plans))]
}

// cpuNormalPlan は take を優先し、スコアが最も高い take を選ぶ。
// take がなければ build、なければ trail。
func (c *Cassino) cpuNormalPlan(playerIdx int, hand []*Card) cpuPlan {
	plans := c.enumerateCpuPlans(playerIdx, hand)
	if len(plans) == 0 {
		return cpuPlan{Type: CassinoActionTrail, handIdx: 0}
	}
	best := plans[0]
	bestScore := c.scorePlan(playerIdx, hand, plans[0])
	for _, p := range plans[1:] {
		sc := c.scorePlan(playerIdx, hand, p)
		if sc > bestScore {
			bestScore = sc
			best = p
		}
	}
	return best
}

// cpuHardPlan は Normal のスコアに加え、trail 候補について相手への漏れを減点する。
func (c *Cassino) cpuHardPlan(playerIdx int, hand []*Card) cpuPlan {
	plans := c.enumerateCpuPlans(playerIdx, hand)
	if len(plans) == 0 {
		return cpuPlan{Type: CassinoActionTrail, handIdx: 0}
	}
	best := plans[0]
	bestScore := c.scorePlanHard(playerIdx, hand, plans[0])
	for _, p := range plans[1:] {
		sc := c.scorePlanHard(playerIdx, hand, p)
		if sc > bestScore {
			bestScore = sc
			best = p
		}
	}
	return best
}

// enumerateCpuPlans は手札全体からあらゆる合法手を列挙する。
func (c *Cassino) enumerateCpuPlans(playerIdx int, hand []*Card) []cpuPlan {
	plans := make([]cpuPlan, 0)
	for handIdx, card := range hand {
		if card == nil {
			continue
		}
		// take
		takes := EnumerateTakes(card, c.round.tableCards, c.round.builds)
		for _, t := range takes {
			plans = append(plans, cpuPlan{
				Type:      CassinoActionTake,
				handIdx:   handIdx,
				tableIdxs: append([]int(nil), t[0]...),
				buildIdxs: append([]int(nil), t[1]...),
			})
		}
		// build (自分がビルドを保有していない時のみ)
		// MultiBuildEnabled は複合ビルド(同値ビルドの合流)を許可するフラグであり、
		// 単独ビルド作成自体は常に許可される (人間プレイヤーと対称)。
		if !c.playerOwnsBuild(playerIdx) {
			cands := EnumerateBuilds(card, handIdx, hand, c.round.tableCards)
			for _, b := range cands {
				plans = append(plans, cpuPlan{
					Type:       CassinoActionBuild,
					handIdx:    handIdx,
					tableIdxs:  append([]int(nil), b.TableIdxs...),
					buildValue: b.DeclaredValue,
				})
			}
		}
		// trail
		if !c.playerOwnsBuild(playerIdx) {
			plans = append(plans, cpuPlan{
				Type:    CassinoActionTrail,
				handIdx: handIdx,
			})
		}
	}
	return plans
}

// scorePlan はプランの期待得点。take は捕獲カードで実質得点換算、
// build は相手に渡さない価値、trail は残す手札の危険度を基準。
func (c *Cassino) scorePlan(playerIdx int, hand []*Card, plan cpuPlan) int {
	switch plan.Type {
	case CassinoActionTake:
		sc := 3 // take 自体に基本点
		// 場札
		for _, idx := range plan.tableIdxs {
			if idx >= 0 && idx < len(c.round.tableCards) {
				sc += cassinoCardValueScore(c.round.tableCards[idx])
			}
		}
		// ビルド
		for _, bi := range plan.buildIdxs {
			if bi >= 0 && bi < len(c.round.builds) {
				for _, card := range c.round.builds[bi].AllCards() {
					sc += cassinoCardValueScore(card)
				}
			}
		}
		// 出す手札の価値はゲットできるので +1
		sc += cassinoCardValueScore(hand[plan.handIdx])
		// スイープ推定
		if c.estimateSweep(plan) {
			sc += 3
		}
		return sc
	case CassinoActionBuild:
		// ビルドは相手に奪われるリスクがあるので小さめ
		return 2
	default:
		// trail は残す手札価値のマイナス
		card := hand[plan.handIdx]
		return -cassinoCardValueScore(card)
	}
}

// scorePlanHard は Hard 難易度の評価。take は Normal と同じ、trail は相手の捕獲を深く減点。
func (c *Cassino) scorePlanHard(playerIdx int, hand []*Card, plan cpuPlan) int {
	base := c.scorePlan(playerIdx, hand, plan)
	if plan.Type != CassinoActionTrail {
		return base
	}
	card := hand[plan.handIdx]
	// trail 後の場に、他プレイヤーが簡単に取れる高得点カードが生まれるかを減点
	penalty := 0
	if CassinoIsSpade(card) {
		penalty += 2
	}
	if CassinoIsAce(card) {
		penalty += 2
	}
	if CassinoIsBigCasino(card) {
		penalty += 3
	}
	if CassinoIsLittleCasino(card) {
		penalty += 2
	}
	return base - penalty
}

// estimateSweep は plan の take が適用された直後に場・ビルドが空になるかどうか。
func (c *Cassino) estimateSweep(plan cpuPlan) bool {
	if plan.Type != CassinoActionTake {
		return false
	}
	// 場札の残り
	usedTable := make(map[int]bool)
	for _, idx := range plan.tableIdxs {
		usedTable[idx] = true
	}
	tableRemain := 0
	for i := range c.round.tableCards {
		if !usedTable[i] {
			tableRemain++
		}
	}
	usedBuilds := make(map[int]bool)
	for _, bi := range plan.buildIdxs {
		usedBuilds[bi] = true
	}
	buildRemain := 0
	for i := range c.round.builds {
		if !usedBuilds[i] {
			buildRemain++
		}
	}
	return tableRemain == 0 && buildRemain == 0
}

// cassinoCardValueScore はカードの"点数価値"を返す (CPU 評価用)。
func cassinoCardValueScore(c *Card) int {
	if c == nil {
		return 0
	}
	score := 1 // 1 枚獲得で 1 点相当 (最多カードボーナスへ寄与)
	if CassinoIsAce(c) {
		score += 2
	}
	if CassinoIsSpade(c) {
		score += 1
	}
	if CassinoIsBigCasino(c) {
		score += 4
	}
	if CassinoIsLittleCasino(c) {
		score += 2
	}
	return score
}
