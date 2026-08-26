//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// Badugi game phase constants.
const (
	BadugiPhaseInit     = 0 // freshly constructed, before Reset()
	BadugiPhaseDeal     = 1 // after deal, first betting round
	BadugiPhaseBet      = 2 // betting round after a draw (drawIndex identifies which)
	BadugiPhaseDraw     = 3 // draw round in progress (drawIndex 1..3)
	BadugiPhaseShowdown = 4 // showdown resolved
	BadugiPhaseEnd      = 5 // hand finished (alias of Showdown for external consumers)
)

// Badugi structural parameters.
const (
	// BadugiHandSize is the cards-per-player count (always 4 in Badugi).
	BadugiHandSize = 4
	// BadugiMaxDraws is the fixed number of draw rounds per hand.
	BadugiMaxDraws = 3
)

// Betting action constants (aliases of the shared betting constants so the
// adapter layer can keep an isolated vocabulary).
const (
	BadugiActionFold  = bettingActionFold
	BadugiActionCheck = bettingActionCheck
	BadugiActionCall  = bettingActionCall
	BadugiActionBet   = bettingActionBet
	BadugiActionRaise = bettingActionRaise
	BadugiActionAllIn = bettingActionAllIn
)

// BadugiResult captures the per-player showdown outcome.
type BadugiResult struct {
	PlayerIdx int    // player slice index
	HandSize  int    // BadugiHand.Size (1..4)
	HandName  string // display name (Badugi / 3-card / 2-card / 1-card)
	WonAmount int    // chips won from the pot(s)
}

// BadugiCpuAction records a single CPU betting decision for replay / UI.
type BadugiCpuAction struct {
	PlayerIdx  int
	Action     int
	Amount     int
	DrawIndex  int // 0 = initial bet, 1..3 = bet after the n-th draw
	RoundLabel string
}

// BadugiCpuExchange records a single CPU draw decision.
type BadugiCpuExchange struct {
	PlayerIdx     int
	DrawIndex     int // 1..3
	ExchangeCount int
}

// badugiRoundState holds the mutable per-hand state. Kept separate so Reset
// can recreate it cleanly without touching config / players.
type badugiRoundState struct {
	phase         int
	drawIndex     int // 0 during initial deal bet, 1..3 during/after the n-th draw
	pot           int
	currentTurn   int
	lastBet       int
	minRaise      int
	raiseCount    int
	actedFlags    []bool
	sidePots      []SidePot
	startingChips []int
	roundResults  []BadugiResult
	cpuActions    []BadugiCpuAction
	cpuExchanges  []BadugiCpuExchange
	actionLogBase
	gameEndFlag     bool
	lastCpuError    error
	lastHumanPlayMs int
}

// Badugi is the Asian-origin 4-card draw lowball poker variant. Players
// start with 4 cards, take up to 3 draws, and show down the best
// subset of cards with all-distinct ranks and all-distinct suits, with
// Ace low.
type Badugi struct {
	trumpCards   *TrumpCards
	players      []*BadugiPlayer
	config       BadugiConfig
	dealerIdx    int
	humanProfile *BettingHumanProfile
	round        badugiRoundState
	// muck はこのハンドで捨てられた札。**山が尽きたらここを切り直して引く。**
	// 4 席・手札 4 枚だと配った時点の山は 36 枚しかなく、3 回のドローで全員が
	// 4 枚引くと最大 48 枚要るので、山切れは規則の範囲内で起きる。
	muck []*Card
}

// NewBadugi constructs a Badugi game. Callers typically use NewDefaultBadugi
// which picks the standard 4-seat, mixed-style CPU lineup.
func NewBadugi(trumpCards *TrumpCards, players []*BadugiPlayer, config BadugiConfig) *Badugi {
	return &Badugi{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round: badugiRoundState{
			phase:         BadugiPhaseInit,
			sidePots:      make([]SidePot, 0),
			actedFlags:    make([]bool, len(players)),
			roundResults:  make([]BadugiResult, 0),
			cpuActions:    make([]BadugiCpuAction, 0),
			cpuExchanges:  make([]BadugiCpuExchange, 0),
			startingChips: make([]int, len(players)),
		},
	}
}

// NewDefaultBadugi returns Badugi with the canonical 4-seat (human + 3 CPU)
// lineup and DefaultBadugiConfig. Single source of truth for CLI / Web /
// Worker construction sites.
func NewDefaultBadugi() *Badugi {
	cfg := DefaultBadugiConfig()
	players := []*BadugiPlayer{
		NewBadugiPlayer(true, BadugiStyleBalanced),
		NewBadugiPlayer(false, BadugiStyleConservative),
		NewBadugiPlayer(false, BadugiStyleAggressive),
		NewBadugiPlayer(false, BadugiStyleBluffer),
	}
	return NewBadugi(NewTrumpCards(0), players, cfg)
}

// Reset clears per-hand state, shuffles, and posts antes. Returns an error
// if config validation fails. Callers must re-call Reset between hands.
func (b *Badugi) Reset() error {
	if err := b.config.Validate(); err != nil {
		return err
	}
	b.round = badugiRoundState{
		phase:         BadugiPhaseInit,
		minRaise:      b.config.MinBet,
		sidePots:      make([]SidePot, 0),
		actedFlags:    make([]bool, len(b.players)),
		roundResults:  make([]BadugiResult, 0),
		cpuActions:    make([]BadugiCpuAction, 0),
		cpuExchanges:  make([]BadugiCpuExchange, 0),
		startingChips: make([]int, len(b.players)),
	}

	if b.config.CpuMetaAI {
		if b.humanProfile != nil {
			b.humanProfile.GamesPlayed++
		} else {
			b.humanProfile = &BettingHumanProfile{}
		}
	}

	// Badugi always uses a 52-card deck (no jokers).
	b.trumpCards = NewTrumpCards(0)
	b.trumpCards.Shuffle()
	// **マックはハンドごと。** 持ち越すと前のハンドの札が新しい山に混ざる。
	b.muck = nil

	activeSeatCount := b.config.CpuCount + 1
	if activeSeatCount > len(b.players) {
		activeSeatCount = len(b.players)
	}
	if activeSeatCount < 1 {
		activeSeatCount = 1
	}

	for i, pl := range b.players {
		pl.Reset()
		pl.SetFolded(false)
		pl.SetAllIn(false)
		pl.SetCurrentBet(0)
		pl.ResetDrawCounters()
		pl.SetHandRank(0)
		if pl.GetChips() <= 0 {
			pl.SetChips(b.config.InitChips)
		}
		if i >= activeSeatCount {
			pl.SetFolded(true)
		}
	}

	for i, pl := range b.players {
		b.round.startingChips[i] = pl.GetChips()
	}

	b.collectAntes()

	// Deal 4 cards to each active player (dealer's left first).
	for c := 0; c < BadugiHandSize; c++ {
		for j := 0; j < len(b.players); j++ {
			idx := (b.dealerIdx + 1 + j) % len(b.players)
			if b.players[idx].GetFolded() {
				continue
			}
			card := b.trumpCards.DrawCard()
			if card != nil {
				b.players[idx].AddCard(card)
			}
		}
	}

	b.round.phase = BadugiPhaseDeal
	b.round.drawIndex = 0
	b.round.currentTurn = b.findNextActive(b.dealerIdx)

	b.runCpuActions()
	return nil
}

// collectAntes posts a ante from every active player into the pot.
func (b *Badugi) collectAntes() {
	for i, pl := range b.players {
		if pl.GetFolded() {
			continue
		}
		ante := b.config.Ante
		if pl.GetChips() < ante {
			ante = pl.GetChips()
		}
		if ante <= 0 {
			continue
		}
		pl.SubtractChips(ante)
		b.round.pot += ante
		b.appendLog(i, "ante", fmt.Sprintf("ante %d chips", ante), nil)
	}
}

// PlayerAction applies the human player's betting action.
// humanPlayMs is the deliberation time in milliseconds (0 = not measured).
func (b *Badugi) PlayerAction(action, amount, humanPlayMs int) error {
	if b.round.gameEndFlag {
		return NewDomainError(ErrGameEnded, "Game has already ended.")
	}
	if b.round.phase != BadugiPhaseDeal && b.round.phase != BadugiPhaseBet {
		return NewDomainError(ErrWrongPhase, "Action is not allowed now.")
	}
	if !b.players[b.round.currentTurn].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "It is not your turn.")
	}

	b.round.lastHumanPlayMs = humanPlayMs
	if b.config.CpuMetaAI && b.humanProfile != nil {
		pl := b.players[b.round.currentTurn]
		pl.EvalHand()
		b.humanProfile.RecordAction(pl.GetHandRank(), action)
		b.humanProfile.RecordHesitation(humanPlayMs)
		if b.round.lastBet > pl.GetCurrentBet() {
			b.humanProfile.RecordFoldToBet(action == BadugiActionFold)
		}
	}

	if err := b.executeAction(b.round.currentTurn, action, amount); err != nil {
		return err
	}

	b.advanceTurn()
	b.runCpuActions()

	// If the betting round closed, run any pending CPU draws.
	if b.round.phase == BadugiPhaseDraw {
		b.advanceDrawPhase()
	}
	return nil
}

// PlayerExchange replaces the cards at the given hand indices with fresh
// cards from the deck. Empty indices = stand pat. Indices outside [0,3] are
// silently ignored.
func (b *Badugi) PlayerExchange(indices []int) error {
	if b.round.phase != BadugiPhaseDraw {
		return NewDomainError(ErrWrongPhase, "Exchange is not allowed now.")
	}
	if !b.players[b.round.currentTurn].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "It is not your turn.")
	}

	b.applyExchange(b.round.currentTurn, indices)
	b.round.actedFlags[b.round.currentTurn] = true

	b.advanceTurn()
	b.advanceDrawPhase()
	return nil
}

// PlayerStand is sugar for PlayerExchange with no indices (stand pat).
func (b *Badugi) PlayerStand() error {
	return b.PlayerExchange(nil)
}

// applyExchange swaps the cards at the given indices, updates counters, and
// logs the action. Shared by human / CPU paths.
func (b *Badugi) applyExchange(playerIdx int, indices []int) {
	pl := b.players[playerIdx]
	drawn := 0
	// **今この席が捨てた札は、この交換のあいだマックに入れない。** 先に混ぜると
	// 自分が捨てたばかりの札を引き直せてしまう (カジノでも現に引いている席の
	// 捨て札は脇に置く)。引き終えてからまとめて積む。
	var pending []*Card
	for _, idx := range indices {
		if idx < 0 || idx >= BadugiHandSize {
			continue
		}
		newCard := b.drawOrRecycleMuck()
		if newCard == nil {
			break
		}
		if old := pl.GetCard(idx); old != nil {
			pending = append(pending, old)
		}
		pl.ExchangeCard(idx, newCard)
		drawn++
	}
	b.muck = append(b.muck, pending...)
	pl.SetDrawCount(drawn)
	pl.AddToTotalDrawCount(drawn)
	b.appendLog(playerIdx, "exchange",
		fmt.Sprintf("draw %d: exchange %d card(s)", b.round.drawIndex, drawn), nil)
}

// drawOrRecycleMuck は山から 1 枚引く。山が尽きていたら捨て札を切り直して
// そこから引く。**どちらも空のときだけ nil を返す** —— 呼び出し側が黙って
// 打ち切ると、捨てたはずの札がそのまま手元に残る。
func (b *Badugi) drawOrRecycleMuck() *Card {
	if c := b.trumpCards.DrawCard(); c != nil {
		return c
	}
	if len(b.muck) == 0 {
		return nil
	}
	// 切り直しは 1 枚ごとではなく、山が尽きた最初の 1 回だけ意味がある。
	// ここでは残りを毎回シャッフルせず、無作為な位置から 1 枚抜く —— 結果は
	// 同じで、混ぜ直しのコストがドロー 1 回ぶんに収まる。
	i := rand.Intn(len(b.muck))
	c := b.muck[i]
	b.muck = append(b.muck[:i], b.muck[i+1:]...)
	return c
}

// bettingPlayers adapts the concrete player slice to the BettingPlayer
// interface slice consumed by the shared betting helpers.
func (b *Badugi) bettingPlayers() []BettingPlayer {
	return toBettingPlayers(b.players)
}

// executeAction runs a single betting action for playerIdx.
func (b *Badugi) executeAction(playerIdx, action, amount int) error {
	bp := b.bettingPlayers()
	state := &BettingState{
		Pot: b.round.pot, LastBet: b.round.lastBet, MinRaise: b.round.minRaise,
		RaiseCount: b.round.raiseCount, ActedFlags: b.round.actedFlags,
	}
	maxRaises, maxBetAmount := b.bettingLimits()
	err := ExecuteBettingAction(bp, state, playerIdx, action, amount, b.currentMinBet(), maxRaises, maxBetAmount)
	b.round.pot = state.Pot
	b.round.lastBet = state.LastBet
	b.round.minRaise = state.MinRaise
	b.round.raiseCount = state.RaiseCount
	if err != nil {
		return err
	}

	b.logBettingAction(playerIdx, action, amount)

	if b.countActivePlayers() == 1 {
		b.resolveLastPlayer()
	}
	return nil
}

// currentMinBet returns the per-round minimum bet. Fixed Limit Badugi uses
// MinBet for the first two rounds and 2×MinBet for the last two; Pot Limit
// and No Limit stay on MinBet as the minimum.
func (b *Badugi) currentMinBet() int {
	if b.config.BettingLimit != BettingLimitFixed {
		return b.config.MinBet
	}
	if b.isLateRound() {
		return b.config.MinBet * 2
	}
	return b.config.MinBet
}

// isLateRound reports whether the current betting round is a "big bet"
// round (after the 2nd or 3rd draw).
func (b *Badugi) isLateRound() bool {
	return b.round.phase == BadugiPhaseBet && b.round.drawIndex >= 2
}

// advanceTurn moves to the next acting player or, when a round completes,
// transitions to the next phase.
func (b *Badugi) advanceTurn() {
	if b.round.gameEndFlag {
		return
	}

	if b.round.phase == BadugiPhaseDeal || b.round.phase == BadugiPhaseBet {
		if b.isBettingRoundComplete() {
			b.advancePhase()
			return
		}
	}

	if b.round.phase == BadugiPhaseDraw {
		if b.isDrawComplete() {
			return
		}
	}

	for i := 1; i <= len(b.players); i++ {
		next := (b.round.currentTurn + i) % len(b.players)
		if !b.players[next].GetFolded() && !b.players[next].GetAllIn() && !b.round.actedFlags[next] {
			b.round.currentTurn = next
			return
		}
	}
}

func (b *Badugi) isRoundComplete() bool {
	for i, pl := range b.players {
		if pl.GetFolded() || pl.GetAllIn() {
			continue
		}
		if !b.round.actedFlags[i] {
			return false
		}
	}
	return true
}

func (b *Badugi) isBettingRoundComplete() bool { return b.isRoundComplete() }
func (b *Badugi) isDrawComplete() bool         { return b.isRoundComplete() }

// advancePhase transitions from a betting round into the next draw, or from
// the final betting round into showdown.
func (b *Badugi) advancePhase() {
	switch b.round.phase {
	case BadugiPhaseDeal, BadugiPhaseBet:
		if b.round.drawIndex >= BadugiMaxDraws {
			b.resolveShowdown()
			return
		}
		b.round.drawIndex++
		b.round.phase = BadugiPhaseDraw
		b.resetBettingRound()
		b.round.currentTurn = b.findNextActive(b.dealerIdx)
		// Kick off any CPU draws that precede the human.
		b.advanceDrawPhase()
	}
}

// advanceDrawPhase runs remaining CPU draws and, once all active seats have
// drawn, opens the next betting round.
func (b *Badugi) advanceDrawPhase() {
	if b.round.gameEndFlag {
		return
	}
	b.runCpuDraws()
	if b.isDrawComplete() {
		b.startNextBettingRound()
	}
}

// startNextBettingRound opens the betting round following a draw. If only
// one or zero active players remain, it jumps straight to showdown.
func (b *Badugi) startNextBettingRound() {
	b.round.phase = BadugiPhaseBet
	b.resetBettingRound()

	activeCnt := 0
	for _, pl := range b.players {
		if !pl.GetFolded() && !pl.GetAllIn() {
			activeCnt++
		}
	}
	if activeCnt <= 1 {
		b.resolveShowdown()
		return
	}

	b.round.currentTurn = b.findNextActive(b.dealerIdx)
	b.runCpuActions()
}

// resetBettingRound clears per-round bets and acted flags.
func (b *Badugi) resetBettingRound() {
	for _, pl := range b.players {
		pl.SetCurrentBet(0)
	}
	b.round.lastBet = 0
	b.round.minRaise = b.currentMinBet()
	b.round.raiseCount = 0
	b.round.actedFlags = make([]bool, len(b.players))
	for i, pl := range b.players {
		if pl.GetFolded() || pl.GetAllIn() {
			b.round.actedFlags[i] = true
		}
	}
}

// findNextActive returns the next seat after fromIdx that is not folded /
// all-in. Used to select the first actor in a round.
func (b *Badugi) findNextActive(fromIdx int) int {
	return findNextActive(b.players, fromIdx)
}

func (b *Badugi) countActivePlayers() int {
	return countPlayers(b.players, func(p *BadugiPlayer) bool { return !p.GetFolded() })
}

// resolveLastPlayer awards the pot to the sole surviving player (everyone
// else folded).
func (b *Badugi) resolveLastPlayer() {
	for i, pl := range b.players {
		if !pl.GetFolded() {
			pl.AddChips(b.round.pot)
			b.round.roundResults = []BadugiResult{{
				PlayerIdx: i,
				WonAmount: b.round.pot,
			}}
			b.round.pot = 0
			break
		}
	}
	b.round.phase = BadugiPhaseEnd
	b.round.gameEndFlag = true
	b.dealerIdx = (b.dealerIdx + 1) % len(b.players)
}

// resolveShowdown evaluates all non-folded hands and distributes the pot.
func (b *Badugi) resolveShowdown() {
	for i, pl := range b.players {
		if !pl.GetFolded() {
			pl.EvalHand()
			cards := make([]*Card, pl.GetCardsSize())
			for j := 0; j < pl.GetCardsSize(); j++ {
				cards[j] = pl.GetCard(j)
			}
			b.appendLog(i, "showdown", fmt.Sprintf("showdown: %s", pl.GetHandName()), cards)
		}
	}

	bp := b.bettingPlayers()
	b.round.sidePots = CalculateSidePots(bp, b.round.pot, b.round.startingChips)
	wonAmounts := DistributePotsWithWinnerFunc(bp, b.round.sidePots, FindPotWinnersBadugi)

	b.round.roundResults = make([]BadugiResult, 0)
	for i, pl := range b.players {
		if pl.GetFolded() {
			continue
		}
		b.round.roundResults = append(b.round.roundResults, BadugiResult{
			PlayerIdx: i,
			HandSize:  pl.GetHandRank(),
			HandName:  pl.GetHandName(),
			WonAmount: wonAmounts[i],
		})
	}

	b.round.phase = BadugiPhaseEnd
	b.round.gameEndFlag = true
	b.dealerIdx = (b.dealerIdx + 1) % len(b.players)
}

// runCpuActions drives CPU betting until control returns to a human or the
// round completes.
func (b *Badugi) runCpuActions() {
	if b.round.gameEndFlag {
		return
	}
	for !b.round.gameEndFlag && (b.round.phase == BadugiPhaseDeal || b.round.phase == BadugiPhaseBet) {
		cur := b.round.currentTurn
		if b.players[cur].GetIsHuman() {
			return
		}
		if b.players[cur].GetFolded() || b.players[cur].GetAllIn() {
			b.advanceTurn()
			continue
		}
		action, amount := b.cpuDecide(cur)
		b.round.cpuActions = append(b.round.cpuActions, BadugiCpuAction{
			PlayerIdx:  cur,
			Action:     action,
			Amount:     amount,
			DrawIndex:  b.round.drawIndex,
			RoundLabel: b.currentRoundLabel(),
		})
		if err := b.executeAction(cur, action, amount); err != nil {
			b.round.lastCpuError = fmt.Errorf("CPU player %d action %d failed: %w", cur, action, err)
			callAmt := b.round.lastBet - b.players[cur].GetCurrentBet()
			if callAmt > 0 {
				_ = b.executeAction(cur, BadugiActionFold, 0)
			} else {
				_ = b.executeAction(cur, BadugiActionCheck, 0)
			}
		}
		if b.round.gameEndFlag {
			return
		}
		b.advanceTurn()
	}
}

// runCpuDraws drives CPU draw decisions until control returns to a human
// or every active seat has drawn.
func (b *Badugi) runCpuDraws() {
	if b.round.gameEndFlag {
		return
	}
	for b.round.phase == BadugiPhaseDraw {
		if b.isDrawComplete() {
			return
		}
		cur := b.round.currentTurn
		if b.players[cur].GetIsHuman() {
			return
		}
		if b.players[cur].GetFolded() || b.players[cur].GetAllIn() {
			b.round.actedFlags[cur] = true
			b.advanceTurn()
			continue
		}
		indices := b.cpuDecideExchange(cur)
		b.applyExchange(cur, indices)
		b.round.cpuExchanges = append(b.round.cpuExchanges, BadugiCpuExchange{
			PlayerIdx:     cur,
			DrawIndex:     b.round.drawIndex,
			ExchangeCount: b.players[cur].GetDrawCount(),
		})
		b.round.actedFlags[cur] = true
		b.advanceTurn()
	}
}

// bettingLimits returns the shared limit settings for the current state.
func (b *Badugi) bettingLimits() (maxRaises, maxBetAmount int) {
	return CalculateBettingLimits(b.config.BettingLimit, b.round.pot, b.round.lastBet)
}

// currentRoundLabel produces the "pre-draw" / "after-draw-N" label used in
// CpuAction records. Helpful for UI debugging and replay.
func (b *Badugi) currentRoundLabel() string {
	if b.round.drawIndex == 0 {
		return "pre-draw"
	}
	return fmt.Sprintf("after-draw-%d", b.round.drawIndex)
}

// cpuDecide picks a betting action for the CPU at idx.
func (b *Badugi) cpuDecide(idx int) (int, int) {
	pl := b.players[idx]
	style := pl.GetPlayStyle()
	callAmount := b.round.lastBet - pl.GetCurrentBet()

	params, ok := badugiStyleParamsMap[style]
	if !ok {
		return b.cpuCallOrCheck(callAmount)
	}

	pl.EvalHand()
	size := pl.GetHandRank() // BadugiHand.Size (1..4)

	var action, amount int
	if b.isLateRound() {
		action, amount = b.cpuDecideLateBet(idx, params, callAmount, size)
	} else {
		action, amount = b.cpuDecideEarlyBet(idx, params, callAmount, size)
	}

	maxRaises, maxBetAmount := b.bettingLimits()
	if maxBetAmount > 0 && amount > maxBetAmount {
		amount = maxBetAmount
	}
	if maxRaises > 0 && b.round.raiseCount >= maxRaises {
		if action == BadugiActionRaise || action == BadugiActionBet {
			if callAmount > 0 {
				return BadugiActionCall, 0
			}
			return BadugiActionCheck, 0
		}
	}
	return action, amount
}

// cpuDecideEarlyBet handles rounds 1-2 (deal + after 1st draw).
func (b *Badugi) cpuDecideEarlyBet(idx int, params badugiCpuStyleParams, callAmount, size int) (int, int) {
	pl := b.players[idx]

	if size <= params.earlyFoldSize {
		if !params.aggressive {
			if callAmount > b.currentMinBet()*params.earlyCallMult {
				return BadugiActionFold, 0
			}
		} else if rand.Intn(100) < params.bluffRate {
			betAmt := b.currentMinBet() * 2
			return b.cpuRaiseOrBet(pl, callAmount, betAmt)
		} else {
			return b.cpuFoldOrCheck(callAmount)
		}
	}
	if size >= params.earlyBetSize || rand.Intn(100) < params.bluffRate {
		betAmt := b.currentMinBet()
		if params.aggressive {
			betAmt = b.currentMinBet() * 2
		}
		return b.cpuRaiseOrBet(pl, callAmount, betAmt)
	}
	return b.cpuCallOrCheck(callAmount)
}

// cpuDecideLateBet handles rounds 3-4 (after 2nd / 3rd draws).
func (b *Badugi) cpuDecideLateBet(idx int, params badugiCpuStyleParams, callAmount, size int) (int, int) {
	pl := b.players[idx]

	// Opponent read: if opponents stood pat, tighten up.
	tighten := b.opponentsStoodPat(idx) > 0

	foldSize := params.lateFoldSize
	if tighten {
		foldSize++
	}

	if size <= foldSize {
		if callAmount > b.currentMinBet()*params.lateCallMult {
			return BadugiActionFold, 0
		}
		if callAmount > 0 {
			return b.cpuCallOrCheck(callAmount)
		}
	}
	if size >= params.lateBetSize || rand.Intn(100) < params.bluffRate {
		betAmt := b.currentMinBet() * 2
		return b.cpuRaiseOrBet(pl, callAmount, betAmt)
	}
	return b.cpuCallOrCheck(callAmount)
}

// opponentsStoodPat counts how many non-folded opponents exchanged 0 cards
// in the most recent draw round. 0 during the initial deal round.
func (b *Badugi) opponentsStoodPat(exceptIdx int) int {
	if b.round.drawIndex == 0 {
		return 0
	}
	cnt := 0
	for i, pl := range b.players {
		if i == exceptIdx || pl.GetFolded() {
			continue
		}
		if pl.GetDrawCount() == 0 {
			cnt++
		}
	}
	return cnt
}

func (b *Badugi) cpuFoldOrCheck(callAmount int) (int, int) { return CpuFoldOrCheck(callAmount) }
func (b *Badugi) cpuCallOrCheck(callAmount int) (int, int) { return CpuCallOrCheck(callAmount) }

func (b *Badugi) cpuRaiseOrBet(pl *BadugiPlayer, callAmount, raiseAmt int) (int, int) {
	return CpuRaiseOrBet(pl.GetChips(), callAmount, raiseAmt)
}

// cpuDecideExchange picks which card indices (if any) the CPU at idx discards.
// Strategy:
//  1. Re-evaluate; if already a Badugi (Size 4) and style allows it, stand pat.
//  2. Otherwise, discard the cards the current best subset does NOT include.
//  3. Among those discards, drop higher-ranked first (Badugi lowball).
func (b *Badugi) cpuDecideExchange(idx int) []int {
	pl := b.players[idx]
	pl.EvalHand()
	best := pl.GetBestHand()

	params := badugiStyleParamsMap[pl.GetPlayStyle()]

	if best.Size >= params.drawStandPatSize {
		return []int{}
	}

	// Occasional stand-pat bluff.
	if params.bluffStandPatPct > 0 && rand.Intn(100) < params.bluffStandPatPct {
		return []int{}
	}

	// Determine which hand positions are NOT in best.Cards. Compare by pointer
	// equality — the evaluator slices through pl.cards, so pointers line up.
	kept := make(map[*Card]bool, best.Size)
	for _, c := range best.Cards {
		kept[c] = true
	}

	type discard struct {
		idx int
		val int
	}
	discards := make([]discard, 0, BadugiHandSize-best.Size)
	for i := 0; i < pl.GetCardsSize(); i++ {
		c := pl.GetCard(i)
		if !kept[c] {
			discards = append(discards, discard{idx: i, val: c.GetValue()})
		}
	}
	// Highest values first so logs read "drop worst card". Order within the
	// slice does not affect the deck draw (each discard pulls a fresh card),
	// but we keep the ordering stable for reproducible CPU replays.
	sort.Slice(discards, func(i, j int) bool {
		return discards[i].val > discards[j].val
	})
	out := make([]int, len(discards))
	for i, d := range discards {
		out[i] = d.idx
	}
	return out
}

// --- Getters ---------------------------------------------------------------

// GetPhase returns the current game phase constant.
func (b *Badugi) GetPhase() int { return b.round.phase }

// GetDrawIndex returns the current draw round counter (0 = pre-draw betting,
// 1..3 = draw/bet rounds after each respective draw).
func (b *Badugi) GetDrawIndex() int { return b.round.drawIndex }

// GetPlayers returns the player slice.
func (b *Badugi) GetPlayers() []*BadugiPlayer { return b.players }

// GetPot returns the current pot value.
func (b *Badugi) GetPot() int { return b.round.pot }

// GetSidePots returns the calculated side pots (populated at showdown).
func (b *Badugi) GetSidePots() []SidePot { return b.round.sidePots }

// GetDealerIdx returns the button seat index.
func (b *Badugi) GetDealerIdx() int { return b.dealerIdx }

// GetCurrentTurn returns the seat index expected to act next.
func (b *Badugi) GetCurrentTurn() int { return b.round.currentTurn }

// GetGameEndFlag reports whether the current hand has been resolved.
func (b *Badugi) GetGameEndFlag() bool { return b.round.gameEndFlag }

// GetLastBet returns the last bet size in the current round.
func (b *Badugi) GetLastBet() int { return b.round.lastBet }

// GetMinRaise returns the minimum legal raise increment.
func (b *Badugi) GetMinRaise() int { return b.round.minRaise }

// GetRaiseCount returns the number of raises so far this round.
func (b *Badugi) GetRaiseCount() int { return b.round.raiseCount }

// GetAnte returns the configured ante value.
func (b *Badugi) GetAnte() int { return b.config.Ante }

// GetRoundResults returns the showdown results for the most recent hand.
func (b *Badugi) GetRoundResults() []BadugiResult { return b.round.roundResults }

// GetCpuActions returns the log of CPU betting decisions for this hand.
func (b *Badugi) GetCpuActions() []BadugiCpuAction { return b.round.cpuActions }

// GetCpuExchanges returns the log of CPU draw decisions for this hand.
func (b *Badugi) GetCpuExchanges() []BadugiCpuExchange { return b.round.cpuExchanges }

// GetConfig returns a copy of the active config.
func (b *Badugi) GetConfig() BadugiConfig { return b.config }

// SetConfig replaces the active config. Callers should Reset before the
// next hand for changes to take effect.
func (b *Badugi) SetConfig(cfg BadugiConfig) { b.config = cfg }

// GetLastCpuError returns the most recent CPU fallback error (test/debug).
func (b *Badugi) GetLastCpuError() error { return b.round.lastCpuError }

// GetHumanProfile returns the meta-AI profile (may be nil).
func (b *Badugi) GetHumanProfile() *BettingHumanProfile { return b.humanProfile }

// ResetProfile clears the meta-AI profile.
func (b *Badugi) ResetProfile() { b.humanProfile = nil }

// ExportProfile returns a marshalable copy of the profile, or nil.
func (b *Badugi) ExportProfile() any {
	if b.humanProfile == nil {
		return nil
	}
	d := b.humanProfile.Export()
	return &d
}

// ImportProfile loads a profile from JSON bytes (no-op on empty input).
func (b *Badugi) ImportProfile(data []byte) error {
	p, err := importBettingProfile(data)
	if err != nil || p == nil {
		return err
	}
	b.humanProfile = p
	return nil
}

// GetActionLog returns the chronological action log for this hand.
func (b *Badugi) GetActionLog() []*ActionLogEntry { return b.round.actionLog }

func (b *Badugi) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	b.round.appendLog(playerIdx, actionType, detail, cards)
}

func (b *Badugi) logBettingAction(playerIdx, action, _ int) {
	switch action {
	case BadugiActionFold:
		b.appendLog(playerIdx, "fold", "fold", nil)
	case BadugiActionCheck:
		b.appendLog(playerIdx, "check", "check", nil)
	case BadugiActionCall:
		b.appendLog(playerIdx, "call", fmt.Sprintf("call %d", b.players[playerIdx].GetCurrentBet()), nil)
	case BadugiActionBet:
		b.appendLog(playerIdx, "bet", fmt.Sprintf("bet %d", b.players[playerIdx].GetCurrentBet()), nil)
	case BadugiActionRaise:
		b.appendLog(playerIdx, "raise", fmt.Sprintf("raise to %d", b.players[playerIdx].GetCurrentBet()), nil)
	case BadugiActionAllIn:
		b.appendLog(playerIdx, "allin", fmt.Sprintf("all in %d", b.players[playerIdx].GetCurrentBet()), nil)
	}
}

// --- JSON round-trip -------------------------------------------------------

// badugiMaxSliceLen caps slice sizes during deserialisation to avoid
// adversarial payloads blowing up memory.
const badugiMaxSliceLen = 1000

type badugiRoundStateJSON struct {
	Phase           int                 `json:"ph"`
	DrawIndex       int                 `json:"dx"`
	Pot             int                 `json:"pt"`
	CurrentTurn     int                 `json:"ct"`
	LastBet         int                 `json:"lb"`
	MinRaise        int                 `json:"mr"`
	RaiseCount      int                 `json:"rc"`
	ActedFlags      []bool              `json:"af"`
	SidePots        []SidePot           `json:"sp"`
	StartingChips   []int               `json:"sc"`
	RoundResults    []BadugiResult      `json:"rr"`
	CpuActions      []BadugiCpuAction   `json:"ca"`
	CpuExchanges    []BadugiCpuExchange `json:"ce"`
	ActionLog       []*ActionLogEntry   `json:"al"`
	GameEndFlag     bool                `json:"ge"`
	LastHumanPlayMs int                 `json:"hm"`
}

type badugiJSON struct {
	TrumpCards *TrumpCards              `json:"tc"`
	Players    []*BadugiPlayer          `json:"pl"`
	Config     BadugiConfig             `json:"cf"`
	DealerIdx  int                      `json:"di"`
	Profile    *BettingHumanProfileData `json:"pf,omitempty"`
	Round      badugiRoundStateJSON     `json:"rd"`
	Muck       []*Card                  `json:"mk,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (b *Badugi) MarshalJSON() ([]byte, error) {
	j := badugiJSON{
		TrumpCards: b.trumpCards,
		Players:    b.players,
		Config:     b.config,
		DealerIdx:  b.dealerIdx,
		Muck:       b.muck,
		Round: badugiRoundStateJSON{
			Phase:           b.round.phase,
			DrawIndex:       b.round.drawIndex,
			Pot:             b.round.pot,
			CurrentTurn:     b.round.currentTurn,
			LastBet:         b.round.lastBet,
			MinRaise:        b.round.minRaise,
			RaiseCount:      b.round.raiseCount,
			ActedFlags:      b.round.actedFlags,
			SidePots:        b.round.sidePots,
			StartingChips:   b.round.startingChips,
			RoundResults:    b.round.roundResults,
			CpuActions:      b.round.cpuActions,
			CpuExchanges:    b.round.cpuExchanges,
			ActionLog:       b.round.actionLog,
			GameEndFlag:     b.round.gameEndFlag,
			LastHumanPlayMs: b.round.lastHumanPlayMs,
		},
	}
	if b.humanProfile != nil {
		d := b.humanProfile.Export()
		j.Profile = &d
	}
	return json.Marshal(j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Badugi) UnmarshalJSON(data []byte) error {
	var j badugiJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > badugiMaxSliceLen || len(j.Round.ActedFlags) > badugiMaxSliceLen ||
		len(j.Round.SidePots) > badugiMaxSliceLen || len(j.Round.StartingChips) > badugiMaxSliceLen ||
		len(j.Round.RoundResults) > badugiMaxSliceLen || len(j.Round.CpuActions) > badugiMaxSliceLen ||
		len(j.Round.CpuExchanges) > badugiMaxSliceLen || len(j.Round.ActionLog) > badugiMaxSliceLen ||
		len(j.Muck) > badugiMaxSliceLen {
		return fmt.Errorf("badugi: input array exceeds maximum allowed size")
	}
	// Consistency check: per-player slices (ActedFlags, StartingChips) must
	// match the Players length, otherwise ExecuteBettingAction's direct
	// ActedFlags[playerIdx] access would panic on restored state.
	if n := len(j.Players); n > 0 {
		if got := len(j.Round.ActedFlags); got != 0 && got != n {
			return fmt.Errorf("badugi: ActedFlags length %d != Players length %d", got, n)
		}
		if got := len(j.Round.StartingChips); got != 0 && got != n {
			return fmt.Errorf("badugi: StartingChips length %d != Players length %d", got, n)
		}
	}
	b.trumpCards = j.TrumpCards
	if b.trumpCards == nil {
		b.trumpCards = NewTrumpCards(0)
	}
	b.players = j.Players
	if b.players == nil {
		b.players = make([]*BadugiPlayer, 0)
	}
	b.config = j.Config
	b.dealerIdx = j.DealerIdx
	b.muck = j.Muck
	if j.Profile != nil {
		b.humanProfile = &BettingHumanProfile{}
		b.humanProfile.Import(*j.Profile)
	}
	b.round = badugiRoundState{
		phase:           j.Round.Phase,
		drawIndex:       j.Round.DrawIndex,
		pot:             j.Round.Pot,
		currentTurn:     j.Round.CurrentTurn,
		lastBet:         j.Round.LastBet,
		minRaise:        j.Round.MinRaise,
		raiseCount:      j.Round.RaiseCount,
		actedFlags:      j.Round.ActedFlags,
		sidePots:        j.Round.SidePots,
		startingChips:   j.Round.StartingChips,
		roundResults:    j.Round.RoundResults,
		cpuActions:      j.Round.CpuActions,
		cpuExchanges:    j.Round.CpuExchanges,
		actionLogBase:   actionLogBase{actionLog: j.Round.ActionLog},
		gameEndFlag:     j.Round.GameEndFlag,
		lastHumanPlayMs: j.Round.LastHumanPlayMs,
	}
	if b.round.actedFlags == nil {
		b.round.actedFlags = make([]bool, 0)
	}
	if b.round.sidePots == nil {
		b.round.sidePots = make([]SidePot, 0)
	}
	if b.round.startingChips == nil {
		b.round.startingChips = make([]int, 0)
	}
	if b.round.roundResults == nil {
		b.round.roundResults = make([]BadugiResult, 0)
	}
	if b.round.cpuActions == nil {
		b.round.cpuActions = make([]BadugiCpuAction, 0)
	}
	if b.round.cpuExchanges == nil {
		b.round.cpuExchanges = make([]BadugiCpuExchange, 0)
	}
	if b.round.actionLog == nil {
		b.round.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
