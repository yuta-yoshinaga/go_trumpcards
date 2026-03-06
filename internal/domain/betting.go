package domain

import "sort"

// BettingLimitType ベッティングリミットタイプ
type BettingLimitType int

const (
	BettingLimitFixed    BettingLimitType = iota // Fixed limit (max 4 raises)
	BettingLimitPotLimit                         // Pot limit (bet/raise ≤ pot)
	BettingLimitNoLimit                          // No limit
)

// BettingLimitNames リミットタイプ名
var BettingLimitNames = []string{
	"Fixed",
	"Pot Limit",
	"No Limit",
}

// ベッティングアクション定数 (Poker/Holdem共通)
const (
	bettingActionFold  = 0
	bettingActionCheck = 1
	bettingActionCall  = 2
	bettingActionBet   = 3
	bettingActionRaise = 4
	bettingActionAllIn = 5

	bettingMaxRaisesPerRound = 4
)

// BettingPlayer ベッティングに必要なプレイヤー操作インターフェース
type BettingPlayer interface {
	GetChips() int
	SubtractChips(int) bool
	AddChips(int)
	GetCurrentBet() int
	SetCurrentBet(int)
	GetFolded() bool
	SetFolded(bool)
	GetAllIn() bool
	SetAllIn(bool)
	GetHandRank() int
	GetComparisonCards() []*Card
}

// SidePot サイドポット (Poker/Holdem共通)
type SidePot struct {
	Amount          int   // ポット額
	EligiblePlayers []int // 受取対象プレイヤーインデックス
}

// BettingState ベッティング状態 (ポインタで渡して共有)
type BettingState struct {
	Pot        int
	LastBet    int
	MinRaise   int
	RaiseCount int
	ActedFlags []bool
}

// ExecuteBettingAction 共通ベッティングアクション実行
// maxRaises: 最大レイズ回数 (0以下で無制限=NoLimit)
// maxBetAmount: 最大ベット額 (0以下で無制限)
func ExecuteBettingAction(players []BettingPlayer, state *BettingState, playerIdx, action, amount, minBetAmount, maxRaises, maxBetAmount int) error {
	pl := players[playerIdx]

	switch action {
	case bettingActionFold:
		pl.SetFolded(true)
		state.ActedFlags[playerIdx] = true

	case bettingActionCheck:
		if state.LastBet > pl.GetCurrentBet() {
			return NewDomainError(ErrInvalidPlay, "Cannot check with outstanding bet.")
		}
		state.ActedFlags[playerIdx] = true

	case bettingActionCall:
		diff := state.LastBet - pl.GetCurrentBet()
		if diff <= 0 {
			return NewDomainError(ErrInvalidPlay, "Nothing to call.")
		}
		if pl.GetChips() <= diff {
			// オールイン (チップ不足)
			allInAmount := pl.GetChips()
			pl.SubtractChips(allInAmount)
			pl.SetCurrentBet(pl.GetCurrentBet() + allInAmount)
			state.Pot += allInAmount
			pl.SetAllIn(true)
		} else {
			pl.SubtractChips(diff)
			pl.SetCurrentBet(pl.GetCurrentBet() + diff)
			state.Pot += diff
		}
		state.ActedFlags[playerIdx] = true

	case bettingActionBet:
		if maxRaises > 0 && state.RaiseCount >= maxRaises {
			return NewDomainError(ErrInvalidPlay, "Maximum number of raises for this round has been reached.")
		}
		if state.LastBet > 0 {
			return NewDomainError(ErrInvalidPlay, "Cannot bet when there is an outstanding bet. Use raise.")
		}
		if amount < minBetAmount {
			return NewDomainError(ErrInvalidAmount, "Bet must be at least the minimum bet.")
		}
		if maxBetAmount > 0 && amount > maxBetAmount {
			return NewDomainError(ErrInvalidAmount, "Bet exceeds maximum allowed amount.")
		}
		if amount > pl.GetChips() {
			return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
		}
		pl.SubtractChips(amount)
		pl.SetCurrentBet(pl.GetCurrentBet() + amount)
		state.Pot += amount
		state.LastBet = pl.GetCurrentBet()
		state.MinRaise = amount
		state.RaiseCount++
		ResetActedExcept(players, state.ActedFlags, playerIdx)
		if pl.GetChips() == 0 {
			pl.SetAllIn(true)
		}

	case bettingActionRaise:
		if maxRaises > 0 && state.RaiseCount >= maxRaises {
			return NewDomainError(ErrInvalidPlay, "Maximum number of raises for this round has been reached.")
		}
		diff := state.LastBet - pl.GetCurrentBet()
		if diff < 0 {
			diff = 0
		}
		if amount < state.MinRaise {
			return NewDomainError(ErrInvalidAmount, "Raise must be at least the minimum raise.")
		}
		if maxBetAmount > 0 && amount > maxBetAmount {
			return NewDomainError(ErrInvalidAmount, "Raise exceeds maximum allowed amount.")
		}
		totalNeeded := diff + amount
		if totalNeeded >= pl.GetChips() {
			return ExecuteBettingAction(players, state, playerIdx, bettingActionAllIn, 0, minBetAmount, maxRaises, maxBetAmount)
		}
		pl.SubtractChips(totalNeeded)
		pl.SetCurrentBet(pl.GetCurrentBet() + totalNeeded)
		state.Pot += totalNeeded
		state.LastBet = pl.GetCurrentBet()
		state.MinRaise = amount
		state.RaiseCount++
		ResetActedExcept(players, state.ActedFlags, playerIdx)

	case bettingActionAllIn:
		allInAmount := pl.GetChips()
		if allInAmount <= 0 {
			return NewDomainError(ErrInsufficientChips, "No chips to go all-in.")
		}
		pl.SubtractChips(allInAmount)
		newBet := pl.GetCurrentBet() + allInAmount
		pl.SetCurrentBet(newBet)
		state.Pot += allInAmount
		pl.SetAllIn(true)
		if newBet > state.LastBet {
			raiseAmount := newBet - state.LastBet
			state.LastBet = newBet
			state.RaiseCount++
			if raiseAmount >= state.MinRaise {
				state.MinRaise = raiseAmount
				ResetActedExcept(players, state.ActedFlags, playerIdx)
			} else {
				state.ActedFlags[playerIdx] = true
			}
		} else {
			state.ActedFlags[playerIdx] = true
		}

	default:
		return NewDomainError(ErrInvalidPlay, "Unknown action.")
	}

	return nil
}

// ResetActedExcept 指定プレイヤー以外のactedフラグをリセット (フォールド・オールイン除く)
func ResetActedExcept(players []BettingPlayer, actedFlags []bool, exceptIdx int) {
	for i := range actedFlags {
		if i == exceptIdx {
			actedFlags[i] = true
			continue
		}
		if players[i].GetFolded() || players[i].GetAllIn() {
			continue
		}
		actedFlags[i] = false
	}
}

// CalculateSidePots サイドポット計算
func CalculateSidePots(players []BettingPlayer, pot int, startingChips []int) []SidePot {
	type playerContrib struct {
		idx    int
		amount int
	}

	contribs := make([]playerContrib, 0, len(players))
	for i, pl := range players {
		invested := startingChips[i] - pl.GetChips()
		if invested < 0 {
			invested = 0
		}
		contribs = append(contribs, playerContrib{idx: i, amount: invested})
	}

	// オールインプレイヤーがいない場合はシンプルなメインポット
	hasAllIn := false
	for _, pl := range players {
		if pl.GetAllIn() && !pl.GetFolded() {
			hasAllIn = true
			break
		}
	}

	if !hasAllIn {
		eligible := make([]int, 0)
		for i, pl := range players {
			if !pl.GetFolded() {
				eligible = append(eligible, i)
			}
		}
		return []SidePot{{Amount: pot, EligiblePlayers: eligible}}
	}

	// オールインがある場合: 各オールイン額でポットを分割
	type allInLevel struct {
		amount int
		idx    int
	}
	levels := make([]allInLevel, 0)
	for _, c := range contribs {
		if players[c.idx].GetAllIn() && !players[c.idx].GetFolded() {
			levels = append(levels, allInLevel{amount: c.amount, idx: c.idx})
		}
	}
	sort.Slice(levels, func(i, j int) bool {
		return levels[i].amount < levels[j].amount
	})

	sidePots := make([]SidePot, 0)
	prevLevel := 0
	remaining := pot

	for _, lv := range levels {
		if lv.amount <= prevLevel {
			continue
		}
		layerAmount := lv.amount - prevLevel
		potAmount := 0
		eligible := make([]int, 0)
		for _, c := range contribs {
			contribution := c.amount - prevLevel
			if contribution <= 0 {
				continue
			}
			if contribution > layerAmount {
				contribution = layerAmount
			}
			potAmount += contribution
			if !players[c.idx].GetFolded() {
				eligible = append(eligible, c.idx)
			}
		}
		if potAmount > 0 {
			sidePots = append(sidePots, SidePot{Amount: potAmount, EligiblePlayers: eligible})
			remaining -= potAmount
		}
		prevLevel = lv.amount
	}

	// 残りのポット (非オールインプレイヤー分)
	if remaining > 0 {
		eligible := make([]int, 0)
		for i, pl := range players {
			if !pl.GetFolded() && !pl.GetAllIn() {
				eligible = append(eligible, i)
			}
		}
		if len(eligible) == 0 {
			// 全員オールインの場合は全未フォールドプレイヤーが対象
			for i, pl := range players {
				if !pl.GetFolded() {
					eligible = append(eligible, i)
				}
			}
		}
		sidePots = append(sidePots, SidePot{Amount: remaining, EligiblePlayers: eligible})
	}

	return sidePots
}

// FindPotWinners 対象プレイヤーから最強ハンドのプレイヤーを返す (複数ならスプリット)
func FindPotWinners(players []BettingPlayer, eligible []int) []int {
	bestRank := -1
	var bestCards []*Card
	var winners []int

	for _, idx := range eligible {
		pl := players[idx]
		if pl.GetFolded() {
			continue
		}
		rank := pl.GetHandRank()
		cards := pl.GetComparisonCards()
		if rank > bestRank {
			bestRank = rank
			bestCards = cards
			winners = []int{idx}
		} else if rank == bestRank {
			cmp := compareHighCardsSlice(cards, bestCards)
			if cmp > 0 {
				bestCards = cards
				winners = []int{idx}
			} else if cmp == 0 {
				winners = append(winners, idx)
			}
		}
	}

	return winners
}

// DistributePots サイドポットの勝者配分を計算しチップを付与
func DistributePots(players []BettingPlayer, sidePots []SidePot) map[int]int {
	wonAmounts := make(map[int]int)
	for _, sp := range sidePots {
		winners := FindPotWinners(players, sp.EligiblePlayers)
		if len(winners) == 0 {
			continue
		}
		share := sp.Amount / len(winners)
		remainder := sp.Amount % len(winners)
		for i, wIdx := range winners {
			won := share
			if i == 0 {
				won += remainder
			}
			players[wIdx].AddChips(won)
			wonAmounts[wIdx] += won
		}
	}
	return wonAmounts
}

// CpuFoldOrCheck コール額がある場合はフォールド、なければチェック
func CpuFoldOrCheck(callAmount int) (int, int) {
	if callAmount > 0 {
		return bettingActionFold, 0
	}
	return bettingActionCheck, 0
}

// CpuCallOrCheck コール額がある場合はコール、なければチェック
func CpuCallOrCheck(callAmount int) (int, int) {
	if callAmount > 0 {
		return bettingActionCall, 0
	}
	return bettingActionCheck, 0
}

// CpuRaiseOrBet レイズまたはベット (チップ不足時はオールイン)
func CpuRaiseOrBet(chips, callAmount, raiseAmt int) (int, int) {
	if raiseAmt > chips {
		return bettingActionAllIn, 0
	}
	if callAmount > 0 {
		if raiseAmt+callAmount > chips {
			return bettingActionAllIn, 0
		}
		return bettingActionRaise, raiseAmt
	}
	return bettingActionBet, raiseAmt
}
