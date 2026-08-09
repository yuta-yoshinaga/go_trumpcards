//go:build !js || !wasm || casino

package domain

import (
	"fmt"
	"sort"
)

// BettingLimitType ベッティングリミットタイプ
type BettingLimitType int

// BettingLimitType定数
const (
	// BettingLimitFixed 固定リミット (最大4回レイズ)
	BettingLimitFixed BettingLimitType = iota
	// BettingLimitPotLimit ポットリミット (ベット/レイズ ≤ ポット)
	BettingLimitPotLimit
	// BettingLimitNoLimit ノーリミット
	BettingLimitNoLimit
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

// CalculateBettingLimits ベッティングリミット設定からmaxRaisesとmaxBetAmountを計算
func CalculateBettingLimits(limit BettingLimitType, pot, lastBet int) (maxRaises, maxBetAmount int) {
	switch limit {
	case BettingLimitPotLimit:
		maxRaises = bettingMaxRaisesPerRound
		maxBetAmount = pot + lastBet
	case BettingLimitNoLimit:
		maxRaises = 0
		maxBetAmount = 0
	default: // Fixed
		maxRaises = bettingMaxRaisesPerRound
		maxBetAmount = 0
	}
	return
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

// FindPotWinnersLowball 対象プレイヤーから最弱ハンドのプレイヤーを返す (2-7 Lowball: 低い手が勝ち)
// ランクが低いほど強い。同ランク時はカード値が低い方が勝つ (Ace=14 always)
func FindPotWinnersLowball(players []BettingPlayer, eligible []int) []int {
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

		if bestRank == -1 {
			bestRank = rank
			bestCards = cards
			winners = []int{idx}
			continue
		}

		if rank < bestRank {
			bestRank = rank
			bestCards = cards
			winners = []int{idx}
		} else if rank == bestRank {
			cmp := compareLowballCards(cards, bestCards)
			if cmp < 0 {
				bestCards = cards
				winners = []int{idx}
			} else if cmp == 0 {
				winners = append(winners, idx)
			}
		}
	}

	return winners
}

// compareLowballCards Lowball用カード比較 (Ace=14常時, 低い方が勝ち)
// a < b: -1 (aが強い), a > b: 1, a == b: 0
func compareLowballCards(a, b []*Card) int {
	aVals := lowballCardValues(a)
	bVals := lowballCardValues(b)
	sort.Sort(sort.Reverse(sort.IntSlice(aVals)))
	sort.Sort(sort.Reverse(sort.IntSlice(bVals)))
	for i := 0; i < len(aVals); i++ {
		if aVals[i] < bVals[i] {
			return -1
		}
		if aVals[i] > bVals[i] {
			return 1
		}
	}
	return 0
}

// lowballCardValues Lowball用カード値取得 (Ace=14, Joker=0)
func lowballCardValues(cards []*Card) []int {
	vals := make([]int, len(cards))
	for i, c := range cards {
		if c.GetDesign() == CardDesignJoker {
			vals[i] = 0
		} else if c.GetValue() == 1 {
			vals[i] = 14
		} else {
			vals[i] = c.GetValue()
		}
	}
	return vals
}

// FindPotWinnersRazz 対象プレイヤーから最弱ハンドのプレイヤーを返す (A-5 Lowball: Ace=1, 低い手が勝ち)
// ストレート・フラッシュは無視。ランクが低いほど強い。同ランク時はカード値が低い方が勝つ (Ace=1)
func FindPotWinnersRazz(players []BettingPlayer, eligible []int) []int {
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

		if bestRank == -1 {
			bestRank = rank
			bestCards = cards
			winners = []int{idx}
			continue
		}

		if rank < bestRank {
			bestRank = rank
			bestCards = cards
			winners = []int{idx}
		} else if rank == bestRank {
			cmp := compareRazzCards(cards, bestCards)
			if cmp < 0 {
				bestCards = cards
				winners = []int{idx}
			} else if cmp == 0 {
				winners = append(winners, idx)
			}
		}
	}

	return winners
}

// WinnerFunc ポッ���勝者判定��数型
type WinnerFunc func(players []BettingPlayer, eligible []int) []int

// DistributePotsWithWinnerFunc サイドポットの勝者配分を計算しチップを付与 (勝者判定関数を指定)
func DistributePotsWithWinnerFunc(players []BettingPlayer, sidePots []SidePot, winnerFunc WinnerFunc) map[int]int {
	wonAmounts := make(map[int]int)
	for _, sp := range sidePots {
		winners := winnerFunc(players, sp.EligiblePlayers)
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

// DistributePots サイドポットの勝者配分を計算しチップを付与
func DistributePots(players []BettingPlayer, sidePots []SidePot) map[int]int {
	return DistributePotsWithWinnerFunc(players, sidePots, FindPotWinners)
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

// afDisplay renders the aggression factor: "-" when the player has neither bet
// nor called, "∞" when they have been aggressive but never called, and the
// ratio otherwise. 6 betting games had this written out.
func afDisplay(betRaise, call int) string {
	if betRaise == 0 && call == 0 {
		return "-"
	}
	if call == 0 {
		return "∞"
	}
	return fmt.Sprintf("%.1f", float64(betRaise)/float64(call))
}

// humanChipHolder is a seat that can be identified as the human and asked for
// its chip count.
type humanChipHolder interface {
	GetIsHuman() bool
	GetChips() int
}

// rebuyAvailable reports whether the human may still rebuy: the option must be
// enabled, the rebuy period not yet past, and the human broke with rebuys left.
// 5 games had this written out.
func rebuyAvailable[P humanChipHolder](enabled bool, handCount, periodHands int, players []P, counts []int, maxCount int) bool {
	if !enabled || handCount > periodHands {
		return false
	}
	for i, p := range players {
		if p.GetIsHuman() && p.GetChips() <= 0 && counts[i] < maxCount {
			return true
		}
	}
	return false
}

// addonAvailable reports whether the human may take the add-on, which is
// offered on exactly one hand and only once. 5 games had this written out.
func addonAvailable[P humanChipHolder](enabled bool, handCount, afterHand int, players []P, used []bool) bool {
	if !enabled || handCount != afterHand {
		return false
	}
	for i, p := range players {
		if p.GetIsHuman() && !used[i] {
			return true
		}
	}
	return false
}

// potBet sizes a CPU bet as a percentage of the pot, floored by the big blind
// and by any outstanding minimum raise. 4 games had this written out.
func potBet(pot, potPct, bigBlind, minRaise int) int {
	bet := pot * potPct / 100
	if bet < bigBlind {
		bet = bigBlind
	}
	if bet < minRaise {
		bet = minRaise
	}
	return bet
}

// blindLogger records the blinds as they are posted.
type blindLogger interface {
	appendLog(playerIdx int, actionType, detail string, cards []*Card)
}

// postBlindsFor takes the small and big blinds from the two seats after the
// dealer, capping each at what the seat actually has and marking it all-in when
// that empties its stack. lastBet ends up at the big blind. 3 games had this
// written out.
//
// pot, lastBet and actedFlags are passed directly because the helper writes all
// three; keeping the fmt.Sprintf here means one copy rather than one per game.
func postBlindsFor[P BettingPlayer](players []P, dealerIdx, smallBlind, bigBlind int,
	pot, lastBet *int, actedFlags []bool, g blindLogger,
) {
	// Returns what was actually posted, which is less than asked for when the
	// seat is too short to cover the blind.
	post := func(idx, want int, label string) int {
		amount := want
		if players[idx].GetChips() < amount {
			amount = players[idx].GetChips()
		}
		players[idx].SubtractChips(amount)
		players[idx].SetCurrentBet(amount)
		*pot += amount
		g.appendLog(idx, "blind", fmt.Sprintf("posts %s %d", label, amount), nil)
		if players[idx].GetChips() == 0 {
			players[idx].SetAllIn(true)
			actedFlags[idx] = true
		}
		return amount
	}
	sbIdx := (dealerIdx + 1) % len(players)
	bbIdx := (dealerIdx + 2) % len(players)
	post(sbIdx, smallBlind, "small blind")
	*lastBet = post(bbIdx, bigBlind, "big blind")
}
