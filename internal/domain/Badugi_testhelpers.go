//go:build test

package domain

// Test-only accessors. Kept in a separate file with the "test" build tag so
// that production binaries cannot mutate private game state.

// SetPhase overrides the current phase.
func (b *Badugi) SetPhase(phase int) { b.round.phase = phase }

// SetDrawIndex overrides the current draw round counter.
func (b *Badugi) SetDrawIndex(n int) { b.round.drawIndex = n }

// SetCurrentTurn overrides the current actor index.
func (b *Badugi) SetCurrentTurn(turn int) { b.round.currentTurn = turn }

// SetPot overrides the pot value.
func (b *Badugi) SetPot(pot int) { b.round.pot = pot }

// SetDealerIdx overrides the dealer seat.
func (b *Badugi) SetDealerIdx(idx int) { b.dealerIdx = idx }

// SetGameEndFlag toggles the game-end flag for targeted test scenarios.
func (b *Badugi) SetGameEndFlag(flag bool) { b.round.gameEndFlag = flag }

// SetLastBet overrides the last-bet value.
func (b *Badugi) SetLastBet(bet int) { b.round.lastBet = bet }

// SetMinRaise overrides the minimum raise increment.
func (b *Badugi) SetMinRaise(raise int) { b.round.minRaise = raise }

// SetRoundResults injects pre-baked round results.
func (b *Badugi) SetRoundResults(results []BadugiResult) { b.round.roundResults = results }

// SetCpuActions injects a CPU action log.
func (b *Badugi) SetCpuActions(actions []BadugiCpuAction) { b.round.cpuActions = actions }

// SetCpuExchanges injects a CPU exchange log.
func (b *Badugi) SetCpuExchanges(exchanges []BadugiCpuExchange) { b.round.cpuExchanges = exchanges }

// SetSidePots injects a side-pot layout.
func (b *Badugi) SetSidePots(pots []SidePot) { b.round.sidePots = pots }

// SetHumanProfile injects a meta-AI profile.
func (b *Badugi) SetHumanProfile(profile *BettingHumanProfile) { b.humanProfile = profile }

// GetLastHumanPlayMs returns the most recent human-action deliberation time.
func (b *Badugi) GetLastHumanPlayMs() int { return b.round.lastHumanPlayMs }

// GetTrumpCards returns the underlying deck for deterministic setup in tests.
func (b *Badugi) GetTrumpCards() *TrumpCards { return b.trumpCards }

// ResetPlayerHand clears a player's hand so tests can stage exact cards via
// AddCard without relying on the shuffled deck order.
func (b *Badugi) ResetPlayerHand(idx int) {
	if 0 <= idx && idx < len(b.players) {
		b.players[idx].Reset()
	}
}
