//go:build !js || !wasm || casino

package domain

import (
	"fmt"
	"math"
	"math/rand"
)

// runCpuActions CPUプレイヤーのアクションを実行
func (s *FiveCardStud) runCpuActions() error {
	if s.gameEndFlag {
		return nil
	}
	const maxIterations = 500
	iterations := 0
	for !s.gameEndFlag && s.phase >= FiveCardStudPhaseSecondStreet && s.phase <= FiveCardStudPhaseFifthStreet {
		iterations++
		if iterations > maxIterations {
			return fmt.Errorf("maxIterations reached in runCpuActions, possible infinite loop")
		}
		if s.players[s.currentTurn].GetIsHuman() {
			return nil
		}
		if s.players[s.currentTurn].GetFolded() || s.players[s.currentTurn].GetAllIn() {
			s.advanceTurn()
			continue
		}
		action, amount := s.cpuDecide(s.currentTurn)
		s.cpuActions = append(s.cpuActions, FiveCardStudCpuAction{
			PlayerIdx: s.currentTurn,
			Action:    action,
			Amount:    amount,
		})
		err := s.executeAction(s.currentTurn, action, amount)
		if err != nil {
			s.handleCpuActionError(s.currentTurn, action, err)
		}
		if s.gameEndFlag {
			return nil
		}
		s.advanceTurn()
	}
	return nil
}

// handleCpuActionError CPUアクション失敗時のフォールバック処理
func (s *FiveCardStud) handleCpuActionError(playerIdx, action int, err error) {
	s.lastCpuError = fmt.Errorf("CPU player %d action %d failed: %w", playerIdx, action, err)
	callAmt := s.lastBet - s.players[playerIdx].GetCurrentBet()
	if callAmt > 0 {
		_ = s.executeAction(playerIdx, FiveCardStudActionFold, 0)
	} else {
		_ = s.executeAction(playerIdx, FiveCardStudActionCheck, 0)
	}
}

// cpuDecide CPUプレイヤーの意思決定
func (s *FiveCardStud) cpuDecide(idx int) (int, int) {
	p := s.players[idx]
	style := p.GetPlayStyle()
	callAmount := s.lastBet - p.GetCurrentBet()

	// GTO: 混合戦略
	if style == HoldemStyleGTO {
		var action, amount int
		if s.phase == FiveCardStudPhaseSecondStreet {
			action, amount = s.cpuDecideThirdStreetGTO(idx, callAmount)
		} else {
			action, amount = s.cpuDecidePostThirdGTO(idx, callAmount)
		}
		maxRaises, maxBetAmount := s.bettingLimits()
		if maxBetAmount > 0 && amount > maxBetAmount {
			amount = maxBetAmount
		}
		if maxRaises > 0 && s.raiseCount >= maxRaises {
			if action == FiveCardStudActionRaise || action == FiveCardStudActionBet {
				if callAmount > 0 {
					return FiveCardStudActionCall, 0
				}
				return FiveCardStudActionCheck, 0
			}
		}
		return action, amount
	}

	params, ok := holdemStyleParamsMap[style]
	if !ok {
		return CpuCallOrCheck(callAmount)
	}

	// メタAI
	if s.config.CpuMetaAI && s.humanProfile != nil {
		adjusted := s.humanProfile.AdjustedBluffChance(float64(params.bluffRate))
		params.bluffRate = int(math.Round(adjusted))
	}

	var action, amount int
	if s.phase == FiveCardStudPhaseSecondStreet {
		action, amount = s.cpuDecideThirdStreet(idx, params, callAmount)
	} else {
		action, amount = s.cpuDecidePostThird(idx, params, callAmount)
	}

	// メタAI: コール確率調整
	if s.config.CpuMetaAI && s.humanProfile != nil && s.lastHumanPlayMs > 0 {
		if action == FiveCardStudActionFold && callAmount > 0 {
			handRank := p.EvalBestHand()
			bracket := bettingHandBracket(handRank)
			adjustedCall := s.humanProfile.AdjustedCallChance(0.0, bracket, s.lastHumanPlayMs)
			if adjustedCall > 0 && rand.Float64() < adjustedCall { //nolint:gosec // non-crypto random for game AI
				action = FiveCardStudActionCall
				amount = 0
			}
		}
	}

	maxRaises, maxBetAmount := s.bettingLimits()
	if maxBetAmount > 0 && amount > maxBetAmount {
		amount = maxBetAmount
	}
	if maxRaises > 0 && s.raiseCount >= maxRaises {
		if action == FiveCardStudActionRaise || action == FiveCardStudActionBet {
			if callAmount > 0 {
				return FiveCardStudActionCall, 0
			}
			return FiveCardStudActionCheck, 0
		}
	}
	return action, amount
}

// evalThirdStreetStrength 開始ストリートのハンド強度評価 (0-100)
func (s *FiveCardStud) evalThirdStreetStrength(idx int) int {
	p := s.players[idx]
	all := p.GetAllCards()
	if len(all) < 2 {
		return 0
	}

	score := 0

	// ペア判定
	freq := make(map[int]int)
	for _, c := range all {
		v := c.GetValue()
		if v == 1 {
			v = 14
		}
		freq[v]++
	}
	hasPair := false
	hasTrips := false
	maxVal := 0
	for v, cnt := range freq {
		if cnt >= 3 {
			hasTrips = true
		}
		if cnt >= 2 {
			hasPair = true
		}
		if v > maxVal {
			maxVal = v
		}
	}

	if hasTrips {
		score = 90
	} else if hasPair {
		// ペアの値によってスコア調整
		for v, cnt := range freq {
			if cnt >= 2 {
				score = 40 + v*3
				break
			}
		}
	} else {
		// ハイカード値
		score = maxVal * 2
	}

	// スーテッド (同スート) ボーナス
	suitCnt := make(map[int]int)
	for _, c := range all {
		suitCnt[c.GetDesign()]++
	}
	for _, cnt := range suitCnt {
		if cnt >= 3 {
			score += 15
			break
		}
	}

	// コネクタボーナス
	vals := make([]int, 0, len(all))
	for _, c := range all {
		v := c.GetValue()
		if v == 1 {
			v = 14
		}
		vals = append(vals, v)
	}
	connected := false
	for i := 0; i < len(vals); i++ {
		for j := i + 1; j < len(vals); j++ {
			diff := vals[i] - vals[j]
			if diff < 0 {
				diff = -diff
			}
			if diff <= 2 {
				connected = true
				break
			}
		}
		if connected {
			break
		}
	}
	if connected && !hasPair {
		score += 10
	}

	return clamp(score, 0, 100)
}

// cpuDecideThirdStreet 開始ストリートのCPU意思決定
func (s *FiveCardStud) cpuDecideThirdStreet(idx int, params cpuStyleParams, callAmount int) (int, int) {
	p := s.players[idx]
	strength := s.evalThirdStreetStrength(idx)

	if strength < params.preFlopFoldThreshold {
		if params.preFlopFoldCompound {
			if callAmount > s.config.SmallBet*params.preFlopFoldCallMult {
				return FiveCardStudActionFold, 0
			}
		} else {
			return CpuFoldOrCheck(callAmount)
		}
	}

	if params.aggressive {
		if strength >= params.preFlopRaiseThreshold || rand.Intn(100) < params.bluffRate { //nolint:gosec
			return CpuRaiseOrBet(p.GetChips(), callAmount, s.cpuPotBet(params.preFlopRaisePotPct))
		}
		return CpuCallOrCheck(callAmount)
	}

	if callAmount > 0 {
		return FiveCardStudActionCall, 0
	}
	if rand.Intn(100) < params.bluffRate { //nolint:gosec
		betAmt := s.cpuPotBet(params.preFlopBluffPotPct)
		if betAmt > p.GetChips() {
			return FiveCardStudActionAllIn, 0
		}
		return FiveCardStudActionBet, betAmt
	}
	return FiveCardStudActionCheck, 0
}

// cpuDecidePostThird 3rd Street以降のCPU意思決定
func (s *FiveCardStud) cpuDecidePostThird(idx int, params cpuStyleParams, callAmount int) (int, int) {
	p := s.players[idx]
	handRank := p.EvalBestHand()

	if params.aggressive {
		if handRank >= params.postFlopRaiseRank || rand.Intn(100) < params.bluffRate { //nolint:gosec
			return CpuRaiseOrBet(p.GetChips(), callAmount, s.cpuPotBet(params.postFlopRaisePotPct))
		}
		if params.postFlopFallbackFold {
			if params.postFlopCondCallRank >= 0 && handRank >= params.postFlopCondCallRank && callAmount > 0 {
				return FiveCardStudActionCall, 0
			}
			return CpuFoldOrCheck(callAmount)
		}
		if callAmount > 0 {
			betSize := s.currentBetSize()
			if handRank <= params.postFlopAggrFoldRank && callAmount > betSize*params.postFlopAggrFoldMult {
				return FiveCardStudActionFold, 0
			}
			return FiveCardStudActionCall, 0
		}
		return FiveCardStudActionCheck, 0
	}

	// Passive
	if callAmount > 0 {
		betSize := s.currentBetSize()
		if handRank <= params.postFlopPassFoldRank {
			if params.postFlopPassFoldMult < 0 || callAmount > betSize*params.postFlopPassFoldMult {
				return FiveCardStudActionFold, 0
			}
		}
		return FiveCardStudActionCall, 0
	}
	if rand.Intn(100) < params.bluffRate { //nolint:gosec
		betAmt := s.cpuPotBet(params.postFlopBluffPotPct)
		if betAmt > p.GetChips() {
			return FiveCardStudActionAllIn, 0
		}
		return FiveCardStudActionBet, betAmt
	}
	return FiveCardStudActionCheck, 0
}

// cpuDecideThirdStreetGTO GTO開始ストリート意思決定
func (s *FiveCardStud) cpuDecideThirdStreetGTO(idx, callAmount int) (int, int) {
	p := s.players[idx]
	strength := s.evalThirdStreetStrength(idx)
	dist := gtoPreFlopTable[gtoPreFlopIndex(strength)]
	decision := gtoRollAction(dist)

	switch decision {
	case 0:
		return CpuFoldOrCheck(callAmount)
	case 2:
		betAmt := s.cpuPotBet(gtoPreFlopBetPct)
		return CpuRaiseOrBet(p.GetChips(), callAmount, betAmt)
	default:
		return CpuCallOrCheck(callAmount)
	}
}

// cpuDecidePostThirdGTO GTO 3rd Street以降意思決定
func (s *FiveCardStud) cpuDecidePostThirdGTO(idx, callAmount int) (int, int) {
	p := s.players[idx]
	handRank := p.EvalBestHand()
	category := classifyGTOHand(handRank)

	// ファイブカードスタッドにはコミュニティカードがないのでドライボード扱い
	dist := gtoPostFlopTable[category][0]
	decision := gtoRollAction(dist)

	potPct := gtoDryBoardBetPct

	switch decision {
	case 0:
		return CpuFoldOrCheck(callAmount)
	case 2:
		betAmt := s.cpuPotBet(potPct)
		return CpuRaiseOrBet(p.GetChips(), callAmount, betAmt)
	default:
		return CpuCallOrCheck(callAmount)
	}
}

// cpuPotBet ポット比率ベースのベット額を計算
func (s *FiveCardStud) cpuPotBet(potPct int) int {
	bet := s.pot * potPct / 100
	betSize := s.currentBetSize()
	if bet < betSize {
		bet = betSize
	}
	if bet < s.minRaise {
		bet = s.minRaise
	}
	return bet
}
