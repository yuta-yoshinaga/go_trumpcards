//go:build test

package domain

// Test-only accessors. Kept in a separate file with the "test" build tag so
// that production binaries cannot mutate private game state.

// SetPhase overrides the current phase.
func (d *DeuceToSeven) SetPhase(phase int) { d.round.phase = phase }

// SetDrawIndex overrides the current draw round counter.
func (d *DeuceToSeven) SetDrawIndex(n int) { d.round.drawIndex = n }

// SetCurrentTurn overrides the current actor index.
func (d *DeuceToSeven) SetCurrentTurn(turn int) { d.round.currentTurn = turn }

// SetPot overrides the pot value.
func (d *DeuceToSeven) SetPot(pot int) { d.round.pot = pot }

// SetDealerIdx overrides the dealer seat.
func (d *DeuceToSeven) SetDealerIdx(idx int) { d.dealerIdx = idx }

// SetGameEndFlag toggles the game-end flag for targeted test scenarios.
func (d *DeuceToSeven) SetGameEndFlag(flag bool) { d.round.gameEndFlag = flag }

// SetLastBet overrides the last-bet value.
func (d *DeuceToSeven) SetLastBet(bet int) { d.round.lastBet = bet }

// SetMinRaise overrides the minimum raise increment.
func (d *DeuceToSeven) SetMinRaise(raise int) { d.round.minRaise = raise }

// SetRoundResults injects pre-baked round results.
func (d *DeuceToSeven) SetRoundResults(results []DeuceToSevenResult) { d.round.roundResults = results }

// SetCpuActions injects a CPU action log.
func (d *DeuceToSeven) SetCpuActions(actions []DeuceToSevenCpuAction) { d.round.cpuActions = actions }

// SetCpuExchanges injects a CPU exchange log.
func (d *DeuceToSeven) SetCpuExchanges(exchanges []DeuceToSevenCpuExchange) {
	d.round.cpuExchanges = exchanges
}

// SetSidePots injects a side-pot layout.
func (d *DeuceToSeven) SetSidePots(pots []SidePot) { d.round.sidePots = pots }

// SetHumanProfile injects a meta-AI profile.
func (d *DeuceToSeven) SetHumanProfile(profile *BettingHumanProfile) { d.humanProfile = profile }

// GetLastHumanPlayMs returns the most recent human-action deliberation time.
func (d *DeuceToSeven) GetLastHumanPlayMs() int { return d.round.lastHumanPlayMs }

// GetTrumpCards returns the underlying deck for deterministic setup in tests.
func (d *DeuceToSeven) GetTrumpCards() *TrumpCards { return d.trumpCards }

// ResetPlayerHand clears a player's hand so tests can stage exact cards via
// AddCard without relying on the shuffled deck order.
func (d *DeuceToSeven) ResetPlayerHand(idx int) {
	if 0 <= idx && idx < len(d.players) {
		d.players[idx].Reset()
	}
}
