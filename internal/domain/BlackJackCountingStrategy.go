//go:build !js || !wasm || casino

package domain

// countForDecision バランスドシステムならTC、KOならRCを返す
func countForDecision(trueCount float64, runningCount int, system int) float64 {
	if IsBalancedCountingSystem(system) {
		return trueCount
	}
	return float64(runningCount)
}

// GetCountingBetAmount カウンティングに基づくCPUベット額を計算する
// ベーススプレッド: count<=1→1x, 2→2x, 3→4x, 4→8x, >=5→16x (base=BJCpuBetAmount)
func GetCountingBetAmount(trueCount float64, runningCount int, system int, availableChips int) int {
	count := countForDecision(trueCount, runningCount, system)
	var multiplier int
	switch {
	case count >= 5:
		multiplier = 16
	case count >= 4:
		multiplier = 8
	case count >= 3:
		multiplier = 4
	case count >= 2:
		multiplier = 2
	default:
		multiplier = 1
	}
	bet := BJCpuBetAmount * multiplier
	if bet > availableChips {
		bet = availableChips
	}
	// BJMinBetの倍数に丸める
	bet = (bet / BJMinBet) * BJMinBet
	return bet
}

// ShouldTakeInsurance カウンティングに基づきインシュランスを取るべきか判定する
// count >= 3 でインシュランスを取る
func ShouldTakeInsurance(trueCount float64, runningCount int, system int) bool {
	return countForDecision(trueCount, runningCount, system) >= 3
}
