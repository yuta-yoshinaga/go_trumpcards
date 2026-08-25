//go:build !js || !wasm || extra2

package domain

// SutdaCPU.go は CPU のベット方策。
//
// **見えているのは自分の 2 枚だけ。** 相手の手札は伏せられているので、判断材料は
// 自分の役の強さと、コールに要る額がポットに対して見合うかしかない。

// 役の強さの目安 (これ以上なら攻め、これ未満なら降りる)。
const (
	// sutdaRaiseThreshold はレイズに踏み切る役の下限 (땡 以上)。
	sutdaRaiseThreshold = SutdaRankTtaeng
	// sutdaFoldThreshold はこれ未満なら降りる役の上限 (5끗 未満)。
	sutdaFoldThreshold = SutdaRankKkeut + 5
)

// cpuChooseAction は CPU の行動を返す。
func (s *Sutda) cpuChooseAction(playerIdx int) string {
	if s.config.CpuDifficulty == SutdaCpuDifficultyEasy {
		// **Easy でも降りることはある。** 常にコールだと卓が動かない。
		if s.GetCallAmount(playerIdx) > 0 && sutdaRandIntn(4) == 0 {
			return SutdaActionFold
		}
		return SutdaActionCall
	}
	return s.smartActionFor(playerIdx)
}

// smartActionFor は難易度に関わらず「良い」行動を返す。
func (s *Sutda) smartActionFor(playerIdx int) string {
	hand := s.GetHandOf(playerIdx)
	need := s.GetCallAmount(playerIdx)

	// **強い役は上げる。** 上げられるかぎり上げないと、ポットが育たない。
	if hand.Rank >= sutdaRaiseThreshold && s.CanRaise(playerIdx) {
		return SutdaActionRaise
	}
	// 追加で払う必要が無いなら降りる理由が無い (チェック)。
	if need <= 0 {
		return SutdaActionCall
	}
	// **弱い役で高い追随はしない。** ポットに対して要求が重いほど降りやすい。
	if hand.Rank < sutdaFoldThreshold && need*2 > s.pot {
		return SutdaActionFold
	}
	if need > s.players[playerIdx].GetChips() && s.players[playerIdx].GetChips() == 0 {
		return SutdaActionFold
	}
	return SutdaActionCall
}
