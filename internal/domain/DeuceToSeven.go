//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// DeuceToSeven game phase constants.
const (
	DeuceToSevenPhaseInit     = 0 // freshly constructed, before Reset()
	DeuceToSevenPhaseDeal     = 1 // after deal, first betting round
	DeuceToSevenPhaseBet      = 2 // betting round after a draw (drawIndex identifies which)
	DeuceToSevenPhaseDraw     = 3 // draw round in progress (drawIndex 1..3)
	DeuceToSevenPhaseShowdown = 4 // showdown resolved
	DeuceToSevenPhaseEnd      = 5 // hand finished (alias of Showdown for external consumers)
)

// DeuceToSeven structural parameters.
const (
	// DeuceToSevenHandSize is the cards-per-player count (always 5).
	DeuceToSevenHandSize = 5
	// DeuceToSevenMaxDraws is the fixed number of draw rounds per hand.
	DeuceToSevenMaxDraws = 3
)

// Betting action constants (aliases of the shared betting constants so the
// adapter layer can keep an isolated vocabulary).
const (
	DeuceToSevenActionFold  = bettingActionFold
	DeuceToSevenActionCheck = bettingActionCheck
	DeuceToSevenActionCall  = bettingActionCall
	DeuceToSevenActionBet   = bettingActionBet
	DeuceToSevenActionRaise = bettingActionRaise
	DeuceToSevenActionAllIn = bettingActionAllIn
)

// DeuceToSevenResult captures the per-player showdown outcome.
type DeuceToSevenResult struct {
	PlayerIdx int    // player slice index
	HandRank  int    // poker category (PokerHandHighCard … PokerHandRoyalFlush)
	HandName  string // display name (High Card / One Pair / …)
	WonAmount int    // chips won from the pot(s)
}

// DeuceToSevenCpuAction records a single CPU betting decision for replay / UI.
type DeuceToSevenCpuAction struct {
	PlayerIdx  int
	Action     int
	Amount     int
	DrawIndex  int // 0 = initial bet, 1..3 = bet after the n-th draw
	RoundLabel string
}

// DeuceToSevenCpuExchange records a single CPU draw decision.
type DeuceToSevenCpuExchange struct {
	PlayerIdx     int
	DrawIndex     int // 1..3
	ExchangeCount int
}

// deuceToSevenRoundState holds the mutable per-hand state. Kept separate so
// Reset can recreate it cleanly without touching config / players.
type deuceToSevenRoundState struct {
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
	roundResults  []DeuceToSevenResult
	cpuActions    []DeuceToSevenCpuAction
	cpuExchanges  []DeuceToSevenCpuExchange
	actionLogBase
	gameEndFlag     bool
	lastCpuError    error
	lastHumanPlayMs int
}

// DeuceToSeven is the "Deuce to Seven" (2-7) Triple Draw lowball poker variant.
// Players start with 5 cards, take up to 3 draws across 4 betting rounds, and
// show down the lowest 5-card poker hand — with the Ace ALWAYS high and
// straights/flushes counting against the player, so the nut low is 7-5-4-3-2.
type DeuceToSeven struct {
	trumpCards   *TrumpCards
	players      []*DeuceToSevenPlayer
	config       DeuceToSevenConfig
	dealerIdx    int
	humanProfile *BettingHumanProfile
	round        deuceToSevenRoundState
	// muck はこのハンドで捨てられた札。**山が尽きたらここを切り直して引く。**
	// 52 枚 4 席だと配った時点の山は 32 枚しかなく、3 回のドローで全員が
	// 5 枚引くと最大 60 枚要るので、山切れは規則の範囲内で起きる。
	muck []*Card
}

// NewDeuceToSeven constructs a DeuceToSeven game. Callers typically use
// NewDefaultDeuceToSeven which picks the standard 4-seat, mixed-style lineup.
func NewDeuceToSeven(trumpCards *TrumpCards, players []*DeuceToSevenPlayer, config DeuceToSevenConfig) *DeuceToSeven {
	return &DeuceToSeven{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round: deuceToSevenRoundState{
			phase:         DeuceToSevenPhaseInit,
			sidePots:      make([]SidePot, 0),
			actedFlags:    make([]bool, len(players)),
			roundResults:  make([]DeuceToSevenResult, 0),
			cpuActions:    make([]DeuceToSevenCpuAction, 0),
			cpuExchanges:  make([]DeuceToSevenCpuExchange, 0),
			startingChips: make([]int, len(players)),
		},
	}
}

// NewDefaultDeuceToSeven returns DeuceToSeven with the canonical 4-seat (human +
// 3 CPU) lineup and DefaultDeuceToSevenConfig. Single source of truth for CLI /
// Web / Worker construction sites.
func NewDefaultDeuceToSeven() *DeuceToSeven {
	cfg := DefaultDeuceToSevenConfig()
	players := []*DeuceToSevenPlayer{
		NewDeuceToSevenPlayer(true, DeuceToSevenStyleBalanced),
		NewDeuceToSevenPlayer(false, DeuceToSevenStyleConservative),
		NewDeuceToSevenPlayer(false, DeuceToSevenStyleAggressive),
		NewDeuceToSevenPlayer(false, DeuceToSevenStyleBluffer),
	}
	return NewDeuceToSeven(NewTrumpCards(0), players, cfg)
}

// Reset clears per-hand state, shuffles, and posts antes. Returns an error if
// config validation fails. Callers must re-call Reset between hands.
func (d *DeuceToSeven) Reset() error {
	if err := d.config.Validate(); err != nil {
		return err
	}
	d.round = deuceToSevenRoundState{
		phase:         DeuceToSevenPhaseInit,
		minRaise:      d.config.MinBet,
		sidePots:      make([]SidePot, 0),
		actedFlags:    make([]bool, len(d.players)),
		roundResults:  make([]DeuceToSevenResult, 0),
		cpuActions:    make([]DeuceToSevenCpuAction, 0),
		cpuExchanges:  make([]DeuceToSevenCpuExchange, 0),
		startingChips: make([]int, len(d.players)),
	}

	if d.config.CpuMetaAI {
		if d.humanProfile != nil {
			d.humanProfile.GamesPlayed++
		} else {
			d.humanProfile = &BettingHumanProfile{}
		}
	}

	// 2-7 Triple Draw always uses a 52-card deck (no jokers).
	d.trumpCards = NewTrumpCards(0)
	d.trumpCards.Shuffle()
	// **マックはハンドごと。** 持ち越すと前のハンドの札が新しい山に混ざる。
	d.muck = nil

	activeSeatCount := d.config.CpuCount + 1
	if activeSeatCount > len(d.players) {
		activeSeatCount = len(d.players)
	}
	if activeSeatCount < 1 {
		activeSeatCount = 1
	}

	for i, pl := range d.players {
		pl.Reset()
		pl.SetFolded(false)
		pl.SetAllIn(false)
		pl.SetCurrentBet(0)
		pl.ResetDrawCounters()
		pl.SetHandRank(0)
		if pl.GetChips() <= 0 {
			pl.SetChips(d.config.InitChips)
		}
		if i >= activeSeatCount {
			pl.SetFolded(true)
		}
	}

	for i, pl := range d.players {
		d.round.startingChips[i] = pl.GetChips()
	}

	d.collectAntes()

	// Deal 5 cards to each active player (dealer's left first).
	for c := 0; c < DeuceToSevenHandSize; c++ {
		for j := 0; j < len(d.players); j++ {
			idx := (d.dealerIdx + 1 + j) % len(d.players)
			if d.players[idx].GetFolded() {
				continue
			}
			card := d.trumpCards.DrawCard()
			if card != nil {
				d.players[idx].AddCard(card)
			}
		}
	}

	d.round.phase = DeuceToSevenPhaseDeal
	d.round.drawIndex = 0
	d.round.currentTurn = d.findNextActive(d.dealerIdx)

	d.runCpuActions()
	return nil
}

// collectAntes posts an ante from every active player into the pot.
func (d *DeuceToSeven) collectAntes() {
	for i, pl := range d.players {
		if pl.GetFolded() {
			continue
		}
		ante := d.config.Ante
		if pl.GetChips() < ante {
			ante = pl.GetChips()
		}
		if ante <= 0 {
			continue
		}
		pl.SubtractChips(ante)
		d.round.pot += ante
		d.appendLog(i, "ante", fmt.Sprintf("ante %d chips", ante), nil)
	}
}

// PlayerAction applies the human player's betting action. humanPlayMs is the
// deliberation time in milliseconds (0 = not measured).
func (d *DeuceToSeven) PlayerAction(action, amount, humanPlayMs int) error {
	if d.round.gameEndFlag {
		return NewDomainError(ErrGameEnded, "Game has already ended.")
	}
	if d.round.phase != DeuceToSevenPhaseDeal && d.round.phase != DeuceToSevenPhaseBet {
		return NewDomainError(ErrWrongPhase, "Action is not allowed now.")
	}
	if !d.players[d.round.currentTurn].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "It is not your turn.")
	}

	d.round.lastHumanPlayMs = humanPlayMs
	if d.config.CpuMetaAI && d.humanProfile != nil {
		pl := d.players[d.round.currentTurn]
		pl.EvalHand()
		d.humanProfile.RecordAction(pl.GetStrength(), action)
		d.humanProfile.RecordHesitation(humanPlayMs)
		if d.round.lastBet > pl.GetCurrentBet() {
			d.humanProfile.RecordFoldToBet(action == DeuceToSevenActionFold)
		}
	}

	if err := d.executeAction(d.round.currentTurn, action, amount); err != nil {
		return err
	}

	d.advanceTurn()
	d.runCpuActions()

	// If the betting round closed, run any pending CPU draws.
	if d.round.phase == DeuceToSevenPhaseDraw {
		d.advanceDrawPhase()
	}
	return nil
}

// PlayerExchange replaces the cards at the given hand indices with fresh cards
// from the deck. Empty indices = stand pat. Indices outside [0,4] are silently
// ignored.
func (d *DeuceToSeven) PlayerExchange(indices []int) error {
	if d.round.phase != DeuceToSevenPhaseDraw {
		return NewDomainError(ErrWrongPhase, "Exchange is not allowed now.")
	}
	if !d.players[d.round.currentTurn].GetIsHuman() {
		return NewDomainError(ErrNotHumanTurn, "It is not your turn.")
	}

	d.applyExchange(d.round.currentTurn, indices)
	d.round.actedFlags[d.round.currentTurn] = true

	d.advanceTurn()
	d.advanceDrawPhase()
	return nil
}

// PlayerStand is sugar for PlayerExchange with no indices (stand pat).
func (d *DeuceToSeven) PlayerStand() error {
	return d.PlayerExchange(nil)
}

// applyExchange swaps the cards at the given indices, updates counters, and logs
// the action. Shared by human / CPU paths.
func (d *DeuceToSeven) applyExchange(playerIdx int, indices []int) {
	pl := d.players[playerIdx]
	drawn := 0
	// **今この席が捨てた札は、この交換のあいだマックに入れない。** 先に混ぜると
	// 自分が捨てたばかりの札を引き直せてしまう (カジノでも現に引いている席の
	// 捨て札は脇に置く)。引き終えてからまとめて積む。
	var pending []*Card
	for _, idx := range indices {
		if idx < 0 || idx >= DeuceToSevenHandSize {
			continue
		}
		newCard := d.drawOrRecycleMuck()
		if newCard == nil {
			break
		}
		if old := pl.GetCard(idx); old != nil {
			pending = append(pending, old)
		}
		pl.ExchangeCard(idx, newCard)
		drawn++
	}
	d.muck = append(d.muck, pending...)
	pl.SetDrawCount(drawn)
	pl.AddToTotalDrawCount(drawn)
	d.appendLog(playerIdx, "exchange",
		fmt.Sprintf("draw %d: exchange %d card(s)", d.round.drawIndex, drawn), nil)
}

// drawOrRecycleMuck は山から 1 枚引く。山が尽きていたら捨て札を切り直して
// そこから引く。**どちらも空のときだけ nil を返す** —— 呼び出し側が黙って
// 打ち切ると、捨てたはずの札がそのまま手元に残る。
func (d *DeuceToSeven) drawOrRecycleMuck() *Card {
	if c := d.trumpCards.DrawCard(); c != nil {
		return c
	}
	if len(d.muck) == 0 {
		return nil
	}
	// 切り直しは 1 枚ごとではなく、山が尽きた最初の 1 回だけ意味がある。
	// ここでは残りを毎回シャッフルせず、無作為な位置から 1 枚抜く —— 結果は
	// 同じで、混ぜ直しのコストがドロー 1 回ぶんに収まる。
	i := rand.Intn(len(d.muck))
	c := d.muck[i]
	d.muck = append(d.muck[:i], d.muck[i+1:]...)
	return c
}

// bettingPlayers adapts the concrete player slice to the BettingPlayer
// interface slice consumed by the shared betting helpers.
func (d *DeuceToSeven) bettingPlayers() []BettingPlayer {
	return toBettingPlayers(d.players)
}

// executeAction runs a single betting action for playerIdx.
func (d *DeuceToSeven) executeAction(playerIdx, action, amount int) error {
	bp := d.bettingPlayers()
	state := &BettingState{
		Pot: d.round.pot, LastBet: d.round.lastBet, MinRaise: d.round.minRaise,
		RaiseCount: d.round.raiseCount, ActedFlags: d.round.actedFlags,
	}
	maxRaises, maxBetAmount := d.bettingLimits()
	err := ExecuteBettingAction(bp, state, playerIdx, action, amount, d.currentMinBet(), maxRaises, maxBetAmount)
	d.round.pot = state.Pot
	d.round.lastBet = state.LastBet
	d.round.minRaise = state.MinRaise
	d.round.raiseCount = state.RaiseCount
	if err != nil {
		return err
	}

	d.logBettingAction(playerIdx, action, amount)

	if d.countActivePlayers() == 1 {
		d.resolveLastPlayer()
	}
	return nil
}

// currentMinBet returns the per-round minimum bet. Fixed Limit uses MinBet for
// the first two rounds and 2×MinBet for the last two; Pot Limit and No Limit
// stay on MinBet as the minimum.
func (d *DeuceToSeven) currentMinBet() int {
	if d.config.BettingLimit != BettingLimitFixed {
		return d.config.MinBet
	}
	if d.isLateRound() {
		return d.config.MinBet * 2
	}
	return d.config.MinBet
}

// isLateRound reports whether the current betting round is a "big bet" round
// (after the 2nd or 3rd draw).
func (d *DeuceToSeven) isLateRound() bool {
	return d.round.phase == DeuceToSevenPhaseBet && d.round.drawIndex >= 2
}

// advanceTurn moves to the next acting player or, when a round completes,
// transitions to the next phase.
func (d *DeuceToSeven) advanceTurn() {
	if d.round.gameEndFlag {
		return
	}

	if d.round.phase == DeuceToSevenPhaseDeal || d.round.phase == DeuceToSevenPhaseBet {
		if d.isBettingRoundComplete() {
			d.advancePhase()
			return
		}
	}

	if d.round.phase == DeuceToSevenPhaseDraw {
		if d.isDrawComplete() {
			return
		}
	}

	for i := 1; i <= len(d.players); i++ {
		next := (d.round.currentTurn + i) % len(d.players)
		if !d.players[next].GetFolded() && !d.players[next].GetAllIn() && !d.round.actedFlags[next] {
			d.round.currentTurn = next
			return
		}
	}
}

func (d *DeuceToSeven) isRoundComplete() bool {
	for i, pl := range d.players {
		if pl.GetFolded() || pl.GetAllIn() {
			continue
		}
		if !d.round.actedFlags[i] {
			return false
		}
	}
	return true
}

func (d *DeuceToSeven) isBettingRoundComplete() bool { return d.isRoundComplete() }
func (d *DeuceToSeven) isDrawComplete() bool         { return d.isRoundComplete() }

// advancePhase transitions from a betting round into the next draw, or from the
// final betting round into showdown.
func (d *DeuceToSeven) advancePhase() {
	switch d.round.phase {
	case DeuceToSevenPhaseDeal, DeuceToSevenPhaseBet:
		if d.round.drawIndex >= DeuceToSevenMaxDraws {
			d.resolveShowdown()
			return
		}
		d.round.drawIndex++
		d.round.phase = DeuceToSevenPhaseDraw
		d.resetBettingRound()
		d.round.currentTurn = d.findNextActive(d.dealerIdx)
		// Kick off any CPU draws that precede the human.
		d.advanceDrawPhase()
	}
}

// advanceDrawPhase runs remaining CPU draws and, once all active seats have
// drawn, opens the next betting round.
func (d *DeuceToSeven) advanceDrawPhase() {
	if d.round.gameEndFlag {
		return
	}
	d.runCpuDraws()
	if d.isDrawComplete() {
		d.startNextBettingRound()
	}
}

// startNextBettingRound opens the betting round following a draw. If only one or
// zero active players remain, it jumps straight to showdown.
func (d *DeuceToSeven) startNextBettingRound() {
	d.round.phase = DeuceToSevenPhaseBet
	d.resetBettingRound()

	activeCnt := 0
	for _, pl := range d.players {
		if !pl.GetFolded() && !pl.GetAllIn() {
			activeCnt++
		}
	}
	if activeCnt <= 1 {
		d.resolveShowdown()
		return
	}

	d.round.currentTurn = d.findNextActive(d.dealerIdx)
	d.runCpuActions()
}

// resetBettingRound clears per-round bets and acted flags.
func (d *DeuceToSeven) resetBettingRound() {
	for _, pl := range d.players {
		pl.SetCurrentBet(0)
	}
	d.round.lastBet = 0
	d.round.minRaise = d.currentMinBet()
	d.round.raiseCount = 0
	d.round.actedFlags = make([]bool, len(d.players))
	for i, pl := range d.players {
		if pl.GetFolded() || pl.GetAllIn() {
			d.round.actedFlags[i] = true
		}
	}
}

// findNextActive returns the next seat after fromIdx that is not folded /
// all-in. Used to select the first actor in a round.
func (d *DeuceToSeven) findNextActive(fromIdx int) int {
	return findNextActive(d.players, fromIdx)
}

func (d *DeuceToSeven) countActivePlayers() int {
	return countPlayers(d.players, func(p *DeuceToSevenPlayer) bool { return !p.GetFolded() })
}

// resolveLastPlayer awards the pot to the sole surviving player (everyone else
// folded).
func (d *DeuceToSeven) resolveLastPlayer() {
	for i, pl := range d.players {
		if !pl.GetFolded() {
			pl.AddChips(d.round.pot)
			d.round.roundResults = []DeuceToSevenResult{{
				PlayerIdx: i,
				WonAmount: d.round.pot,
			}}
			d.round.pot = 0
			break
		}
	}
	d.round.phase = DeuceToSevenPhaseEnd
	d.round.gameEndFlag = true
	d.dealerIdx = (d.dealerIdx + 1) % len(d.players)
}

// resolveShowdown evaluates all non-folded hands and distributes the pot.
func (d *DeuceToSeven) resolveShowdown() {
	for i, pl := range d.players {
		if !pl.GetFolded() {
			pl.EvalHand()
			cards := make([]*Card, pl.GetCardsSize())
			for j := 0; j < pl.GetCardsSize(); j++ {
				cards[j] = pl.GetCard(j)
			}
			d.appendLog(i, "showdown", fmt.Sprintf("showdown: %s", pl.GetHandName()), cards)
		}
	}

	bp := d.bettingPlayers()
	d.round.sidePots = CalculateSidePots(bp, d.round.pot, d.round.startingChips)
	wonAmounts := DistributePotsWithWinnerFunc(bp, d.round.sidePots, FindPotWinnersDeuceToSeven)

	d.round.roundResults = make([]DeuceToSevenResult, 0)
	for i, pl := range d.players {
		if pl.GetFolded() {
			continue
		}
		d.round.roundResults = append(d.round.roundResults, DeuceToSevenResult{
			PlayerIdx: i,
			HandRank:  pl.GetHandRank(),
			HandName:  pl.GetHandName(),
			WonAmount: wonAmounts[i],
		})
	}

	d.round.phase = DeuceToSevenPhaseEnd
	d.round.gameEndFlag = true
	d.dealerIdx = (d.dealerIdx + 1) % len(d.players)
}

// runCpuActions drives CPU betting until control returns to a human or the
// round completes.
func (d *DeuceToSeven) runCpuActions() {
	if d.round.gameEndFlag {
		return
	}
	for !d.round.gameEndFlag && (d.round.phase == DeuceToSevenPhaseDeal || d.round.phase == DeuceToSevenPhaseBet) {
		cur := d.round.currentTurn
		if d.players[cur].GetIsHuman() {
			return
		}
		if d.players[cur].GetFolded() || d.players[cur].GetAllIn() {
			d.advanceTurn()
			continue
		}
		action, amount := d.cpuDecide(cur)
		d.round.cpuActions = append(d.round.cpuActions, DeuceToSevenCpuAction{
			PlayerIdx:  cur,
			Action:     action,
			Amount:     amount,
			DrawIndex:  d.round.drawIndex,
			RoundLabel: d.currentRoundLabel(),
		})
		if err := d.executeAction(cur, action, amount); err != nil {
			d.round.lastCpuError = fmt.Errorf("CPU player %d action %d failed: %w", cur, action, err)
			callAmt := d.round.lastBet - d.players[cur].GetCurrentBet()
			if callAmt > 0 {
				_ = d.executeAction(cur, DeuceToSevenActionFold, 0)
			} else {
				_ = d.executeAction(cur, DeuceToSevenActionCheck, 0)
			}
		}
		if d.round.gameEndFlag {
			return
		}
		d.advanceTurn()
	}
}

// runCpuDraws drives CPU draw decisions until control returns to a human or
// every active seat has drawn.
func (d *DeuceToSeven) runCpuDraws() {
	if d.round.gameEndFlag {
		return
	}
	for d.round.phase == DeuceToSevenPhaseDraw {
		if d.isDrawComplete() {
			return
		}
		cur := d.round.currentTurn
		if d.players[cur].GetIsHuman() {
			return
		}
		if d.players[cur].GetFolded() || d.players[cur].GetAllIn() {
			d.round.actedFlags[cur] = true
			d.advanceTurn()
			continue
		}
		indices := d.cpuDecideExchange(cur)
		d.applyExchange(cur, indices)
		d.round.cpuExchanges = append(d.round.cpuExchanges, DeuceToSevenCpuExchange{
			PlayerIdx:     cur,
			DrawIndex:     d.round.drawIndex,
			ExchangeCount: d.players[cur].GetDrawCount(),
		})
		d.round.actedFlags[cur] = true
		d.advanceTurn()
	}
}

// bettingLimits returns the shared limit settings for the current state.
func (d *DeuceToSeven) bettingLimits() (maxRaises, maxBetAmount int) {
	return CalculateBettingLimits(d.config.BettingLimit, d.round.pot, d.round.lastBet)
}

// currentRoundLabel produces the "pre-draw" / "after-draw-N" label used in
// CpuAction records. Helpful for UI debugging and replay.
func (d *DeuceToSeven) currentRoundLabel() string {
	if d.round.drawIndex == 0 {
		return "pre-draw"
	}
	return fmt.Sprintf("after-draw-%d", d.round.drawIndex)
}

// cpuDecide picks a betting action for the CPU at idx.
func (d *DeuceToSeven) cpuDecide(idx int) (int, int) {
	pl := d.players[idx]
	style := pl.GetPlayStyle()
	callAmount := d.round.lastBet - pl.GetCurrentBet()

	params, ok := deuceToSevenStyleParamsMap[style]
	if !ok {
		return d.cpuCallOrCheck(callAmount)
	}

	pl.EvalHand()
	strength := pl.GetStrength() // deuceLowStrength (1..4)

	var action, amount int
	if d.isLateRound() {
		action, amount = d.cpuDecideLateBet(idx, params, callAmount, strength)
	} else {
		action, amount = d.cpuDecideEarlyBet(idx, params, callAmount, strength)
	}

	maxRaises, maxBetAmount := d.bettingLimits()
	if maxBetAmount > 0 && amount > maxBetAmount {
		amount = maxBetAmount
	}
	if maxRaises > 0 && d.round.raiseCount >= maxRaises {
		if action == DeuceToSevenActionRaise || action == DeuceToSevenActionBet {
			if callAmount > 0 {
				return DeuceToSevenActionCall, 0
			}
			return DeuceToSevenActionCheck, 0
		}
	}
	return action, amount
}

// cpuDecideEarlyBet handles rounds 1-2 (deal + after 1st draw).
func (d *DeuceToSeven) cpuDecideEarlyBet(idx int, params deuceToSevenCpuStyleParams, callAmount, strength int) (int, int) {
	pl := d.players[idx]

	if strength <= params.earlyFoldSize {
		if !params.aggressive {
			if callAmount > d.currentMinBet()*params.earlyCallMult {
				return DeuceToSevenActionFold, 0
			}
		} else if rand.Intn(100) < params.bluffRate {
			betAmt := d.currentMinBet() * 2
			return d.cpuRaiseOrBet(pl, callAmount, betAmt)
		} else {
			return d.cpuFoldOrCheck(callAmount)
		}
	}
	if strength >= params.earlyBetSize || rand.Intn(100) < params.bluffRate {
		betAmt := d.currentMinBet()
		if params.aggressive {
			betAmt = d.currentMinBet() * 2
		}
		return d.cpuRaiseOrBet(pl, callAmount, betAmt)
	}
	return d.cpuCallOrCheck(callAmount)
}

// cpuDecideLateBet handles rounds 3-4 (after 2nd / 3rd draws).
func (d *DeuceToSeven) cpuDecideLateBet(idx int, params deuceToSevenCpuStyleParams, callAmount, strength int) (int, int) {
	pl := d.players[idx]

	// Opponent read: if opponents stood pat, tighten up.
	tighten := d.opponentsStoodPat(idx) > 0

	foldSize := params.lateFoldSize
	if tighten {
		foldSize++
	}

	if strength <= foldSize {
		if callAmount > d.currentMinBet()*params.lateCallMult {
			return DeuceToSevenActionFold, 0
		}
		if callAmount > 0 {
			return d.cpuCallOrCheck(callAmount)
		}
	}
	if strength >= params.lateBetSize || rand.Intn(100) < params.bluffRate {
		betAmt := d.currentMinBet() * 2
		return d.cpuRaiseOrBet(pl, callAmount, betAmt)
	}
	return d.cpuCallOrCheck(callAmount)
}

// opponentsStoodPat counts how many non-folded opponents exchanged 0 cards in
// the most recent draw round. 0 during the initial deal round.
func (d *DeuceToSeven) opponentsStoodPat(exceptIdx int) int {
	if d.round.drawIndex == 0 {
		return 0
	}
	cnt := 0
	for i, pl := range d.players {
		if i == exceptIdx || pl.GetFolded() {
			continue
		}
		if pl.GetDrawCount() == 0 {
			cnt++
		}
	}
	return cnt
}

func (d *DeuceToSeven) cpuFoldOrCheck(callAmount int) (int, int) { return CpuFoldOrCheck(callAmount) }
func (d *DeuceToSeven) cpuCallOrCheck(callAmount int) (int, int) { return CpuCallOrCheck(callAmount) }

func (d *DeuceToSeven) cpuRaiseOrBet(pl *DeuceToSevenPlayer, callAmount, raiseAmt int) (int, int) {
	return CpuRaiseOrBet(pl.GetChips(), callAmount, raiseAmt)
}

// cpuDecideExchange picks which card indices (if any) the CPU at idx discards.
// Strategy:
//  1. Re-evaluate; if the made low is strong enough for this style, stand pat.
//  2. Occasionally bluff stand-pat on a weak hand.
//  3. Otherwise discard pair duplicates and high cards (9+ / Ace), keeping the
//     lowest distinct cards (2..8). If the hand is a made straight/flush with no
//     discard candidate, break it by dropping the single highest card.
func (d *DeuceToSeven) cpuDecideExchange(idx int) []int {
	pl := d.players[idx]
	pl.EvalHand()

	params := deuceToSevenStyleParamsMap[pl.GetPlayStyle()]

	if pl.GetStrength() >= params.drawStandPatSize {
		return []int{}
	}
	if params.bluffStandPatPct > 0 && rand.Intn(100) < params.bluffStandPatPct {
		return []int{}
	}
	return deuceToSevenDiscardIndices(pl)
}

// deuceToSevenDiscardIndices returns the hand indices a drawing player should
// discard: pair duplicates and high cards (value 9+ or Ace), keeping the lowest
// distinct cards (2..8). If keeping every card would leave a made straight or
// flush (no natural discard), the single highest card is dropped to break it.
// SuggestExchange は playerIdx の推奨交換カードインデックスを返す。空 (nil) の場合は
// 完成ロー (スタンドパット) 推奨。既に 8 以下で完成したローならスタンドパットを推奨し、
// それ以外は deuceToSevenDiscardIndices の交換候補を返す。
func (d *DeuceToSeven) SuggestExchange(playerIdx int) []int {
	players := d.GetPlayers()
	if playerIdx < 0 || playerIdx >= len(players) {
		return nil
	}
	pl := players[playerIdx]
	if deuceToSevenIsPatLow(pl) {
		return nil
	}
	return deuceToSevenDiscardIndices(pl)
}

// deuceToSevenIsPatLow は手札が完成ロー (8 以下で立てられるパット) かを判定する。
// 条件: 5 枚がすべて異なるランクで 8 以下 (エースは高札扱いで不可)、かつストレートでも
// フラッシュでもないこと。
func deuceToSevenIsPatLow(pl *DeuceToSevenPlayer) bool {
	if pl.GetCardsSize() != 5 {
		return false
	}
	ranks := make([]int, 0, 5)
	firstSuit := pl.GetCard(0).GetDesign()
	flush := true
	seen := make(map[int]bool, 5)
	for i := 0; i < 5; i++ {
		c := pl.GetCard(i)
		v := c.GetValue()
		if v == 1 {
			v = 14 // Ace is always high (bad for low)
		}
		if v > 8 || seen[v] {
			return false // too high or paired
		}
		seen[v] = true
		ranks = append(ranks, v)
		if c.GetDesign() != firstSuit {
			flush = false
		}
	}
	if flush {
		return false
	}
	sort.Ints(ranks)
	if ranks[len(ranks)-1]-ranks[0] == 4 {
		return false // 5 distinct consecutive ranks = straight
	}
	return true
}

func deuceToSevenDiscardIndices(pl *DeuceToSevenPlayer) []int {
	type cardInfo struct{ idx, low int } // low = 2-7 value with Ace = 14
	n := pl.GetCardsSize()
	infos := make([]cardInfo, n)
	for i := 0; i < n; i++ {
		v := pl.GetCard(i).GetValue()
		if v == 1 {
			v = 14 // Ace is always high
		}
		infos[i] = cardInfo{idx: i, low: v}
	}
	// Prefer keeping the lowest cards: sort ascending by value.
	sort.Slice(infos, func(i, j int) bool { return infos[i].low < infos[j].low })

	keptRanks := make(map[int]bool)
	keepIdx := make(map[int]bool)
	for _, info := range infos {
		if info.low <= 8 && !keptRanks[info.low] {
			keptRanks[info.low] = true
			keepIdx[info.idx] = true
		}
	}

	discards := make([]int, 0, n)
	for i := 0; i < n; i++ {
		if !keepIdx[i] {
			discards = append(discards, i)
		}
	}
	// Highest card values first so logs read "drop worst card" — purely cosmetic
	// (each discard pulls a fresh card regardless of order).
	sort.Slice(discards, func(i, j int) bool {
		return deuceHighValue(pl.GetCard(discards[i])) > deuceHighValue(pl.GetCard(discards[j]))
	})

	if len(discards) == 0 {
		// Every card was "kept" yet the hand is not pat — it must be a made
		// straight or flush. Break it by dropping the single highest card.
		worst, worstVal := 0, -1
		for i := 0; i < n; i++ {
			if v := deuceHighValue(pl.GetCard(i)); v > worstVal {
				worstVal, worst = v, i
			}
		}
		discards = []int{worst}
	}
	return discards
}

// deuceHighValue returns a card's value with the Ace counted as 14 (high).
func deuceHighValue(c *Card) int {
	if v := c.GetValue(); v == 1 {
		return 14
	}
	return c.GetValue()
}

// --- Getters ---------------------------------------------------------------

// GetPhase returns the current game phase constant.
func (d *DeuceToSeven) GetPhase() int { return d.round.phase }

// GetDrawIndex returns the current draw round counter (0 = pre-draw betting,
// 1..3 = draw/bet rounds after each respective draw).
func (d *DeuceToSeven) GetDrawIndex() int { return d.round.drawIndex }

// GetPlayers returns the player slice.
func (d *DeuceToSeven) GetPlayers() []*DeuceToSevenPlayer { return d.players }

// GetPlayerCnt returns the number of seats at the table.
func (d *DeuceToSeven) GetPlayerCnt() int { return len(d.players) }

// GetPot returns the current pot value.
func (d *DeuceToSeven) GetPot() int { return d.round.pot }

// GetSidePots returns the calculated side pots (populated at showdown).
func (d *DeuceToSeven) GetSidePots() []SidePot { return d.round.sidePots }

// GetDealerIdx returns the button seat index.
func (d *DeuceToSeven) GetDealerIdx() int { return d.dealerIdx }

// GetCurrentTurn returns the seat index expected to act next.
func (d *DeuceToSeven) GetCurrentTurn() int { return d.round.currentTurn }

// GetGameEndFlag reports whether the current hand has been resolved.
func (d *DeuceToSeven) GetGameEndFlag() bool { return d.round.gameEndFlag }

// GetLastBet returns the last bet size in the current round.
func (d *DeuceToSeven) GetLastBet() int { return d.round.lastBet }

// GetMinRaise returns the minimum legal raise increment.
func (d *DeuceToSeven) GetMinRaise() int { return d.round.minRaise }

// GetRaiseCount returns the number of raises so far this round.
func (d *DeuceToSeven) GetRaiseCount() int { return d.round.raiseCount }

// GetAnte returns the configured ante value.
func (d *DeuceToSeven) GetAnte() int { return d.config.Ante }

// GetRoundResults returns the showdown results for the most recent hand.
func (d *DeuceToSeven) GetRoundResults() []DeuceToSevenResult { return d.round.roundResults }

// GetCpuActions returns the log of CPU betting decisions for this hand.
func (d *DeuceToSeven) GetCpuActions() []DeuceToSevenCpuAction { return d.round.cpuActions }

// GetCpuExchanges returns the log of CPU draw decisions for this hand.
func (d *DeuceToSeven) GetCpuExchanges() []DeuceToSevenCpuExchange { return d.round.cpuExchanges }

// GetConfig returns a copy of the active config.
func (d *DeuceToSeven) GetConfig() DeuceToSevenConfig { return d.config }

// SetConfig replaces the active config. Callers should Reset before the next
// hand for changes to take effect.
func (d *DeuceToSeven) SetConfig(cfg DeuceToSevenConfig) { d.config = cfg }

// GetLastCpuError returns the most recent CPU fallback error (test/debug).
func (d *DeuceToSeven) GetLastCpuError() error { return d.round.lastCpuError }

// GetHumanProfile returns the meta-AI profile (may be nil).
func (d *DeuceToSeven) GetHumanProfile() *BettingHumanProfile { return d.humanProfile }

// ResetProfile clears the meta-AI profile.
func (d *DeuceToSeven) ResetProfile() { d.humanProfile = nil }

// ExportProfile returns a marshalable copy of the profile, or nil.
func (d *DeuceToSeven) ExportProfile() any {
	if d.humanProfile == nil {
		return nil
	}
	data := d.humanProfile.Export()
	return &data
}

// ImportProfile loads a profile from JSON bytes (no-op on empty input).
func (d *DeuceToSeven) ImportProfile(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	pd, err := ImportBettingHumanProfileJSON(data)
	if err != nil {
		return err
	}
	d.humanProfile = &BettingHumanProfile{}
	d.humanProfile.Import(pd)
	return nil
}

// GetActionLog returns the chronological action log for this hand.
func (d *DeuceToSeven) GetActionLog() []*ActionLogEntry { return d.round.actionLog }

func (d *DeuceToSeven) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	d.round.appendLog(playerIdx, actionType, detail, cards)
}

func (d *DeuceToSeven) logBettingAction(playerIdx, action, _ int) {
	switch action {
	case DeuceToSevenActionFold:
		d.appendLog(playerIdx, "fold", "fold", nil)
	case DeuceToSevenActionCheck:
		d.appendLog(playerIdx, "check", "check", nil)
	case DeuceToSevenActionCall:
		d.appendLog(playerIdx, "call", fmt.Sprintf("call %d", d.players[playerIdx].GetCurrentBet()), nil)
	case DeuceToSevenActionBet:
		d.appendLog(playerIdx, "bet", fmt.Sprintf("bet %d", d.players[playerIdx].GetCurrentBet()), nil)
	case DeuceToSevenActionRaise:
		d.appendLog(playerIdx, "raise", fmt.Sprintf("raise to %d", d.players[playerIdx].GetCurrentBet()), nil)
	case DeuceToSevenActionAllIn:
		d.appendLog(playerIdx, "allin", fmt.Sprintf("all in %d", d.players[playerIdx].GetCurrentBet()), nil)
	}
}

// --- JSON round-trip -------------------------------------------------------

// deuceToSevenMaxSliceLen caps slice sizes during deserialisation to avoid
// adversarial payloads blowing up memory.
const deuceToSevenMaxSliceLen = 1000

type deuceToSevenRoundStateJSON struct {
	Phase           int                       `json:"ph"`
	DrawIndex       int                       `json:"dx"`
	Pot             int                       `json:"pt"`
	CurrentTurn     int                       `json:"ct"`
	LastBet         int                       `json:"lb"`
	MinRaise        int                       `json:"mr"`
	RaiseCount      int                       `json:"rc"`
	ActedFlags      []bool                    `json:"af"`
	SidePots        []SidePot                 `json:"sp"`
	StartingChips   []int                     `json:"sc"`
	RoundResults    []DeuceToSevenResult      `json:"rr"`
	CpuActions      []DeuceToSevenCpuAction   `json:"ca"`
	CpuExchanges    []DeuceToSevenCpuExchange `json:"ce"`
	ActionLog       []*ActionLogEntry         `json:"al"`
	GameEndFlag     bool                      `json:"ge"`
	LastHumanPlayMs int                       `json:"hm"`
}

type deuceToSevenJSON struct {
	TrumpCards *TrumpCards                `json:"tc"`
	Players    []*DeuceToSevenPlayer      `json:"pl"`
	Config     DeuceToSevenConfig         `json:"cf"`
	DealerIdx  int                        `json:"di"`
	Profile    *BettingHumanProfileData   `json:"pf,omitempty"`
	Round      deuceToSevenRoundStateJSON `json:"rd"`
	Muck       []*Card                    `json:"mk,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (d *DeuceToSeven) MarshalJSON() ([]byte, error) {
	j := deuceToSevenJSON{
		TrumpCards: d.trumpCards,
		Players:    d.players,
		Config:     d.config,
		DealerIdx:  d.dealerIdx,
		Muck:       d.muck,
		Round: deuceToSevenRoundStateJSON{
			Phase:           d.round.phase,
			DrawIndex:       d.round.drawIndex,
			Pot:             d.round.pot,
			CurrentTurn:     d.round.currentTurn,
			LastBet:         d.round.lastBet,
			MinRaise:        d.round.minRaise,
			RaiseCount:      d.round.raiseCount,
			ActedFlags:      d.round.actedFlags,
			SidePots:        d.round.sidePots,
			StartingChips:   d.round.startingChips,
			RoundResults:    d.round.roundResults,
			CpuActions:      d.round.cpuActions,
			CpuExchanges:    d.round.cpuExchanges,
			ActionLog:       d.round.actionLog,
			GameEndFlag:     d.round.gameEndFlag,
			LastHumanPlayMs: d.round.lastHumanPlayMs,
		},
	}
	if d.humanProfile != nil {
		pd := d.humanProfile.Export()
		j.Profile = &pd
	}
	return json.Marshal(j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *DeuceToSeven) UnmarshalJSON(data []byte) error {
	var j deuceToSevenJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > deuceToSevenMaxSliceLen || len(j.Round.ActedFlags) > deuceToSevenMaxSliceLen ||
		len(j.Round.SidePots) > deuceToSevenMaxSliceLen || len(j.Round.StartingChips) > deuceToSevenMaxSliceLen ||
		len(j.Round.RoundResults) > deuceToSevenMaxSliceLen || len(j.Round.CpuActions) > deuceToSevenMaxSliceLen ||
		len(j.Round.CpuExchanges) > deuceToSevenMaxSliceLen || len(j.Round.ActionLog) > deuceToSevenMaxSliceLen ||
		len(j.Muck) > deuceToSevenMaxSliceLen {
		return fmt.Errorf("deucetoseven: input array exceeds maximum allowed size")
	}
	// Consistency check: per-player slices must match the Players length,
	// otherwise ExecuteBettingAction's direct ActedFlags[playerIdx] access
	// would panic on restored state.
	if n := len(j.Players); n > 0 {
		if got := len(j.Round.ActedFlags); got != n {
			return fmt.Errorf("deucetoseven: ActedFlags length %d != Players length %d", got, n)
		}
		if got := len(j.Round.StartingChips); got != n {
			return fmt.Errorf("deucetoseven: StartingChips length %d != Players length %d", got, n)
		}
	}
	d.trumpCards = j.TrumpCards
	if d.trumpCards == nil {
		d.trumpCards = NewTrumpCards(0)
	}
	d.players = j.Players
	if d.players == nil {
		d.players = make([]*DeuceToSevenPlayer, 0)
	}
	d.config = j.Config
	d.dealerIdx = j.DealerIdx
	d.muck = j.Muck
	if j.Profile != nil {
		d.humanProfile = &BettingHumanProfile{}
		d.humanProfile.Import(*j.Profile)
	}
	d.round = deuceToSevenRoundState{
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
	if d.round.actedFlags == nil {
		d.round.actedFlags = make([]bool, 0)
	}
	if d.round.sidePots == nil {
		d.round.sidePots = make([]SidePot, 0)
	}
	if d.round.startingChips == nil {
		d.round.startingChips = make([]int, 0)
	}
	if d.round.roundResults == nil {
		d.round.roundResults = make([]DeuceToSevenResult, 0)
	}
	if d.round.cpuActions == nil {
		d.round.cpuActions = make([]DeuceToSevenCpuAction, 0)
	}
	if d.round.cpuExchanges == nil {
		d.round.cpuExchanges = make([]DeuceToSevenCpuExchange, 0)
	}
	if d.round.actionLog == nil {
		d.round.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
