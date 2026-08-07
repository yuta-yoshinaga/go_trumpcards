//go:build !js || !wasm || casino

package domain

import (
	"fmt"
	"math"
	"math/rand"
)

// runCpuActions CPUプレイヤーのアクションを実行
func (p *Pineapple) runCpuActions() error {
	if p.gameEndFlag {
		return nil
	}
	const maxIterations = 500
	iterations := 0
	for !p.gameEndFlag && p.phase >= PineapplePhasePreFlop && p.phase <= PineapplePhaseRiver {
		iterations++
		if iterations > maxIterations {
			return fmt.Errorf("maxIterations reached in runCpuActions, possible infinite loop")
		}
		if p.players[p.currentTurn].GetIsHuman() {
			return nil
		}
		if p.players[p.currentTurn].GetFolded() || p.players[p.currentTurn].GetAllIn() {
			p.advanceTurn()
			continue
		}
		action, amount := p.cpuDecide(p.currentTurn)
		p.cpuActions = append(p.cpuActions, HoldemCpuAction{
			PlayerIdx: p.currentTurn,
			Action:    action,
			Amount:    amount,
		})
		err := p.executeAction(p.currentTurn, action, amount)
		if err != nil {
			p.handleCpuActionError(p.currentTurn, action, err)
		}
		if p.gameEndFlag {
			return nil
		}
		p.advanceTurn()
	}
	return nil
}

// handleCpuActionError CPUアクション失敗時のフォールバック処理
func (p *Pineapple) handleCpuActionError(playerIdx, action int, err error) {
	p.lastCpuError = fmt.Errorf("CPU player %d action %d failed: %w", playerIdx, action, err)
	callAmt := p.lastBet - p.players[playerIdx].GetCurrentBet()
	if callAmt > 0 {
		_ = p.executeAction(playerIdx, PineappleActionFold, 0)
	} else {
		_ = p.executeAction(playerIdx, PineappleActionCheck, 0)
	}
}

// cpuDecide CPUプレイヤーの意思決定
func (p *Pineapple) cpuDecide(idx int) (int, int) {
	pl := p.players[idx]
	style := pl.GetPlayStyle()
	callAmount := p.lastBet - pl.GetCurrentBet()

	// GTO
	if style == HoldemStyleGTO {
		var action, amount int
		if p.phase == PineapplePhasePreFlop {
			action, amount = p.cpuDecidePreFlopGTO(idx, callAmount)
		} else {
			action, amount = p.cpuDecidePostFlopGTO(idx, callAmount)
		}
		maxRaises, maxBetAmount := p.bettingLimits()
		if maxBetAmount > 0 && amount > maxBetAmount {
			amount = maxBetAmount
		}
		if maxRaises > 0 && p.raiseCount >= maxRaises {
			if action == PineappleActionRaise || action == PineappleActionBet {
				if callAmount > 0 {
					return PineappleActionCall, 0
				}
				return PineappleActionCheck, 0
			}
		}
		return action, amount
	}

	params, ok := holdemStyleParamsMap[style]
	if !ok {
		return p.cpuCallOrCheck(callAmount)
	}

	// メタAI: ブラフ率を調整
	if p.config.CpuMetaAI && p.humanProfile != nil {
		adjusted := p.humanProfile.AdjustedBluffChance(float64(params.bluffRate))
		params.bluffRate = int(math.Round(adjusted))
	}

	var action, amount int
	if p.phase == PineapplePhasePreFlop {
		action, amount = p.cpuDecidePreFlop(idx, params, callAmount)
	} else {
		action, amount = p.cpuDecidePostFlop(idx, params, callAmount)
	}

	// メタAI: 人間のベット/レイズに対してコール確率を調整
	if p.config.CpuMetaAI && p.humanProfile != nil && p.lastHumanPlayMs > 0 {
		if action == PineappleActionFold && callAmount > 0 {
			handRank := pl.EvalBestHand(p.communityCards)
			bracket := bettingHandBracket(handRank)
			adjustedCall := p.humanProfile.AdjustedCallChance(0.0, bracket, p.lastHumanPlayMs)
			if adjustedCall > 0 && rand.Float64() < adjustedCall { //nolint:gosec // non-crypto random for game AI
				action = PineappleActionCall
				amount = 0
			}
		}
	}

	maxRaises, maxBetAmount := p.bettingLimits()

	if maxBetAmount > 0 && amount > maxBetAmount {
		amount = maxBetAmount
	}

	if maxRaises > 0 && p.raiseCount >= maxRaises {
		if action == PineappleActionRaise || action == PineappleActionBet {
			if callAmount > 0 {
				return PineappleActionCall, 0
			}
			return PineappleActionCheck, 0
		}
	}
	return action, amount
}

// cpuFoldOrCheck コール額がある場合はフォールド、なければチェック
func (p *Pineapple) cpuFoldOrCheck(callAmount int) (int, int) {
	return CpuFoldOrCheck(callAmount)
}

// cpuCallOrCheck コール額がある場合はコール、なければチェック
func (p *Pineapple) cpuCallOrCheck(callAmount int) (int, int) {
	return CpuCallOrCheck(callAmount)
}

// cpuRaiseOrBet コール額がある場合はレイズ、なければベット
func (p *Pineapple) cpuRaiseOrBet(pl *PineapplePlayer, callAmount, raiseAmt int) (int, int) {
	return CpuRaiseOrBet(pl.GetChips(), callAmount, raiseAmt)
}

// cpuBetOrAllIn ベットする (チップ不足時はオールイン)
func (p *Pineapple) cpuBetOrAllIn(pl *PineapplePlayer, betAmt int) (int, int) {
	if betAmt > pl.GetChips() {
		return PineappleActionAllIn, 0
	}
	return PineappleActionBet, betAmt
}

// cpuPotBet ポット比率ベースのベット額を計算
func (p *Pineapple) cpuPotBet(potPct int) int {
	bet := p.pot * potPct / 100
	if bet < p.config.BigBlind {
		bet = p.config.BigBlind
	}
	if bet < p.minRaise {
		bet = p.minRaise
	}
	return bet
}

// cpuDecidePreFlop プリフロップのCPU意思決定
func (p *Pineapple) cpuDecidePreFlop(idx int, params cpuStyleParams, callAmount int) (int, int) {
	pl := p.players[idx]
	strength := p.evalPreFlopStrength(idx)

	if strength < params.preFlopFoldThreshold {
		if params.preFlopFoldCompound {
			if callAmount > p.config.BigBlind*params.preFlopFoldCallMult {
				return PineappleActionFold, 0
			}
		} else {
			return p.cpuFoldOrCheck(callAmount)
		}
	}

	if params.aggressive {
		if strength >= params.preFlopRaiseThreshold || rand.Intn(100) < params.bluffRate {
			return p.cpuRaiseOrBet(pl, callAmount, p.cpuPotBet(params.preFlopRaisePotPct))
		}
		return p.cpuCallOrCheck(callAmount)
	}

	if callAmount > 0 {
		return PineappleActionCall, 0
	}
	if rand.Intn(100) < params.bluffRate {
		return p.cpuBetOrAllIn(pl, p.cpuPotBet(params.preFlopBluffPotPct))
	}
	return PineappleActionCheck, 0
}

// cpuDecidePostFlop フロップ以降のCPU意思決定
func (p *Pineapple) cpuDecidePostFlop(idx int, params cpuStyleParams, callAmount int) (int, int) {
	pl := p.players[idx]
	handRank := pl.EvalBestHand(p.communityCards)

	if params.aggressive {
		if handRank >= params.postFlopRaiseRank || rand.Intn(100) < params.bluffRate {
			return p.cpuRaiseOrBet(pl, callAmount, p.cpuPotBet(params.postFlopRaisePotPct))
		}
		if params.postFlopFallbackFold {
			if params.postFlopCondCallRank >= 0 && handRank >= params.postFlopCondCallRank && callAmount > 0 {
				return PineappleActionCall, 0
			}
			return p.cpuFoldOrCheck(callAmount)
		}
		if callAmount > 0 {
			if handRank <= params.postFlopAggrFoldRank && callAmount > p.config.BigBlind*params.postFlopAggrFoldMult {
				return PineappleActionFold, 0
			}
			return PineappleActionCall, 0
		}
		return PineappleActionCheck, 0
	}

	if callAmount > 0 {
		if handRank <= params.postFlopPassFoldRank {
			if params.postFlopPassFoldMult < 0 || callAmount > p.config.BigBlind*params.postFlopPassFoldMult {
				return PineappleActionFold, 0
			}
		}
		return PineappleActionCall, 0
	}
	if rand.Intn(100) < params.bluffRate {
		return p.cpuBetOrAllIn(pl, p.cpuPotBet(params.postFlopBluffPotPct))
	}
	return PineappleActionCheck, 0
}

// cpuDecidePreFlopGTO GTOプリフロップ意思決定
func (p *Pineapple) cpuDecidePreFlopGTO(idx, callAmount int) (int, int) {
	pl := p.players[idx]
	strength := p.evalPreFlopStrength(idx)
	dist := gtoPreFlopTable[gtoPreFlopIndex(strength)]
	decision := gtoRollAction(dist)

	switch decision {
	case 0:
		return p.cpuFoldOrCheck(callAmount)
	case 2:
		betAmt := p.cpuPotBet(gtoPreFlopBetPct)
		return p.cpuRaiseOrBet(pl, callAmount, betAmt)
	default:
		return p.cpuCallOrCheck(callAmount)
	}
}

// cpuDecidePostFlopGTO GTOポストフロップ意思決定
func (p *Pineapple) cpuDecidePostFlopGTO(idx, callAmount int) (int, int) {
	pl := p.players[idx]
	handRank := pl.EvalBestHand(p.communityCards)
	category := classifyGTOHand(handRank)
	bt := evalBoardTexture(p.communityCards)

	wetIdx := 0
	if bt.wet {
		wetIdx = 1
	}
	dist := gtoPostFlopTable[category][wetIdx]
	decision := gtoRollAction(dist)

	potPct := gtoDryBoardBetPct
	if bt.wet {
		potPct = gtoWetBoardBetPct
	}

	if bt.paired && decision == 2 && category <= gtoHandMedium && rand.Intn(100) < 30 {
		return p.cpuCallOrCheck(callAmount)
	}

	if bt.highCards >= 3 && decision == 2 && category <= gtoHandWeak && rand.Intn(100) < 40 {
		return p.cpuCallOrCheck(callAmount)
	}

	switch decision {
	case 0:
		return p.cpuFoldOrCheck(callAmount)
	case 2:
		betAmt := p.cpuPotBet(potPct)
		return p.cpuRaiseOrBet(pl, callAmount, betAmt)
	default:
		return p.cpuCallOrCheck(callAmount)
	}
}

// evalPreFlopStrength プリフロップハンド強度評価 (0-100)
// N枚のホールカードから最も強い2枚ペアの強度を評価
func (p *Pineapple) evalPreFlopStrength(idx int) int {
	pl := p.players[idx]
	n := pl.GetCardsSize()
	if n < 2 {
		return 0
	}
	if n == 2 {
		return evalTwoCardStrength(pl.GetCard(0), pl.GetCard(1))
	}

	// N枚からC(N,2)通りの2枚ペアを評価し、最高スコアを返す
	bestScore := 0
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			score := evalTwoCardStrength(pl.GetCard(i), pl.GetCard(j))
			if score > bestScore {
				bestScore = score
			}
		}
	}

	bonus := n + 2
	return clamp(bestScore+bonus, 0, 100)
}

// evalTwoCardStrength 2枚のカードからプリフロップ強度を計算 (Holdem互換)
func evalTwoCardStrength(c1, c2 *Card) int {
	v1 := c1.GetValue()
	v2 := c2.GetValue()
	d1 := c1.GetDesign()
	d2 := c2.GetDesign()

	if v1 == 1 {
		v1 = 14
	}
	if v2 == 1 {
		v2 = 14
	}

	suited := d1 == d2
	score := 0

	// ペア
	if v1 == v2 {
		score = 50 + v1*3
		if v1 >= 10 {
			score += 15
		}
		return clamp(score, 0, 100)
	}

	// ハイカード値
	high := v1
	low := v2
	if v2 > v1 {
		high = v2
		low = v1
	}

	score = high*2 + low
	if suited {
		score += 10
	}

	// コネクタ
	gap := high - low
	switch gap {
	case 1:
		score += 10
	case 2:
		score += 5
	}

	// ハイカードボーナス
	if high >= 12 {
		score += 10
	}

	return clamp(score, 0, 100)
}

// bestRankWithBoard は渡したホールカードと現在のボードから作れる最強の役を返す。
//
// **5枚に届かないときはポーカーの役ではなくプリフロップ用スコア (0-100) を
// 返す。**尺度が2つ混ざるので、同じ局面の中での大小比較にしか使えない。
// 外向きに役を名乗る GetHumanDiscardPreviews はボードが揃うまで何も返さない。
func (p *Pineapple) bestRankWithBoard(hole []*Card) int {
	all := make([]*Card, 0, len(hole)+len(p.communityCards))
	all = append(all, hole...)
	all = append(all, p.communityCards...)

	if len(all) < 5 {
		if len(hole) == 2 {
			return evalTwoCardStrength(hole[0], hole[1])
		}
		return -1
	}

	rank := -1
	for _, combo := range combinations(all, 5) {
		if r := evalFiveCardHand(combo); r > rank {
			rank = r
		}
	}
	return rank
}

// cpuDiscard CPUプレイヤーのディスカード意思決定
// C(N,2)通りの2枚ペアを全評価し、最適な2枚を残すように1枚を捨てる。
func (p *Pineapple) cpuDiscard(idx int) int {
	pl := p.players[idx]
	n := pl.GetCardsSize()
	if n <= 2 {
		return 0
	}

	bestKeepI, bestKeepJ := 0, 1
	bestRank := -1

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			rank := p.bestRankWithBoard([]*Card{pl.GetCard(i), pl.GetCard(j)})
			if rank > bestRank {
				bestRank = rank
				bestKeepI, bestKeepJ = i, j
			}
		}
	}

	for k := 0; k < n; k++ {
		if k != bestKeepI && k != bestKeepJ {
			return k
		}
	}
	return 0
}
