package domain

// GinRummyOutcomeKind はラウンドがどう決着したか。
type GinRummyOutcomeKind int

const (
	// GinRummyOutcomeGin ノッカーがデッドウッド 0 で上がった。
	GinRummyOutcomeGin GinRummyOutcomeKind = iota
	// GinRummyOutcomeKnock 通常のノック。ノッカーがデッドウッド差を得る。
	GinRummyOutcomeKnock
	// GinRummyOutcomeUndercut 相手のデッドウッドがノッカー以下で、**相手**が得点する。
	GinRummyOutcomeUndercut
)

// GinRummyRoundOutcome はラウンド結果の種類と内訳。
type GinRummyRoundOutcome struct {
	Kind             GinRummyOutcomeKind
	WinnerIdx        int
	KnockerDeadwood  int
	OpponentDeadwood int
	Base             int // 得点のうちデッドウッド差 (Gin では相手のデッドウッド)
	Bonus            int
	Total            int
}

// GinRummyRoundOutcomeOf はノック時点の情報から結果を導く。
//
// **判定の順序と式は scoreRound と同一。**別に書くと、表示された内訳と
// 実際に動いた点が食い違う。特にアンダーカットは **勝者がノッカーではなく
// 相手に変わる**ので、種類を言わずに数字だけ出すと誰が得したのか読めない。
func GinRummyRoundOutcomeOf(knockerIdx, knockerDeadwood, opponentDeadwood int, isGin bool) GinRummyRoundOutcome {
	opponentIdx := 1 - knockerIdx
	if isGin {
		return GinRummyRoundOutcome{
			Kind: GinRummyOutcomeGin, WinnerIdx: knockerIdx,
			KnockerDeadwood: knockerDeadwood, OpponentDeadwood: opponentDeadwood,
			Base: opponentDeadwood, Bonus: GinRummyGinBonus,
			Total: opponentDeadwood + GinRummyGinBonus,
		}
	}
	if opponentDeadwood <= knockerDeadwood {
		diff := knockerDeadwood - opponentDeadwood
		return GinRummyRoundOutcome{
			Kind: GinRummyOutcomeUndercut, WinnerIdx: opponentIdx,
			KnockerDeadwood: knockerDeadwood, OpponentDeadwood: opponentDeadwood,
			Base: diff, Bonus: GinRummyUndercutBonus, Total: diff + GinRummyUndercutBonus,
		}
	}
	diff := opponentDeadwood - knockerDeadwood
	return GinRummyRoundOutcome{
		Kind: GinRummyOutcomeKnock, WinnerIdx: knockerIdx,
		KnockerDeadwood: knockerDeadwood, OpponentDeadwood: opponentDeadwood,
		Base: diff, Bonus: 0, Total: diff,
	}
}
