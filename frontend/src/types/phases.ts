/**
 * Game phase enums for all games.
 *
 * These values must stay in sync with the backend domain constants:
 *   - BlackJack: internal/domain/BlackJack.go (BJPhaseBet, BJPhaseDeal, BJPhaseInsurance, BJPhaseAction, BJPhaseEnd)
 *   - Poker:     internal/domain/Poker.go     (PokerPhaseInit, PokerPhaseDeal, PokerPhaseExchange, PokerPhaseSecondBet, PokerPhaseEnd)
 */

// BlackJack phase constants (sync: internal/domain/BlackJack.go)
export const BjPhase = {
  BET: 1,
  DEAL: 2,
  INSURANCE: 3,
  ACTION: 4,
  END: 5,
} as const;

// Poker phase constants (sync: internal/domain/Poker.go)
export const PokerPhase = {
  INIT: 0,
  DEAL: 1,
  EXCHANGE: 2,
  SECOND_BET: 3,
  END: 4,
} as const;
