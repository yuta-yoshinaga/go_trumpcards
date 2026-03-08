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
  EARLY_SURRENDER: 6,
} as const;

// Poker phase constants (sync: internal/domain/Poker.go)
export const PokerPhase = {
  INIT: 0,
  DEAL: 1,
  EXCHANGE: 2,
  SECOND_BET: 3,
  END: 4,
} as const;

// Poker action constants (sync: internal/domain/Poker.go)
export const PokerAction = {
  FOLD: 0,
  CHECK: 1,
  CALL: 2,
  BET: 3,
  RAISE: 4,
  ALL_IN: 5,
} as const;

// Texas Hold'em phase constants (sync: internal/domain/Holdem.go)
export const HoldemPhase = {
  INIT: 0,
  PRE_FLOP: 1,
  FLOP: 2,
  TURN: 3,
  RIVER: 4,
  SHOWDOWN: 5,
  END: 6,
  REBUY: 7,
} as const;

// Texas Hold'em rebuy phase type constants (sync: internal/domain/Holdem.go)
export const HoldemRebuyPhaseType = {
  NONE: 0,
  REBUY: 1,
  ADDON: 2,
} as const;
