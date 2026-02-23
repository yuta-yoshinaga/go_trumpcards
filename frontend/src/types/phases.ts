/**
 * Game phase constants for all games.
 *
 * These values must stay in sync with the backend domain constants:
 *   - BlackJack: internal/domain/BlackJack.go (BJPhaseBet, BJPhaseDeal, BJPhaseInsurance, BJPhaseAction, BJPhaseEnd)
 *   - Poker:     internal/domain/Poker.go     (PokerPhaseInit, PokerPhaseDeal, PokerPhaseExchange, PokerPhaseSecondBet, PokerPhaseEnd)
 */

// BlackJack phase constants (sync: internal/domain/BlackJack.go)
export const BJ_PHASE_BET = 1;
export const BJ_PHASE_DEAL = 2;
export const BJ_PHASE_INSURANCE = 3;
export const BJ_PHASE_ACTION = 4;
export const BJ_PHASE_END = 5;

// Poker phase constants (sync: internal/domain/Poker.go)
export const POKER_PHASE_INIT = 0;
export const POKER_PHASE_DEAL = 1;
export const POKER_PHASE_EXCHANGE = 2;
export const POKER_PHASE_SECOND_BET = 3;
export const POKER_PHASE_END = 4;
