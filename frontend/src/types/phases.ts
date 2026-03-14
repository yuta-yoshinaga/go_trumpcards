/**
 * Game phase enums for all games.
 *
 * These values must stay in sync with the backend domain constants:
 *   - BlackJack:  internal/domain/BlackJack.go (BJPhaseBet, BJPhaseDeal, BJPhaseInsurance, BJPhaseAction, BJPhaseEnd)
 *   - Poker:      internal/domain/Poker.go     (PokerPhaseInit, PokerPhaseDeal, PokerPhaseExchange, PokerPhaseSecondBet, PokerPhaseEnd)
 *   - Hearts:     internal/domain/Hearts.go    (HeartsPhasePass, HeartsPhasePlay, HeartsPhaseTrickEnd, HeartsPhaseRoundEnd, HeartsPhaseGameEnd)
 *   - Memory:     internal/domain/Memory.go    (MemoryPhaseFlip1, MemoryPhaseFlip2, MemoryPhaseResult, MemoryPhaseGameEnd)
 *   - Klondike:   internal/domain/Klondike.go  (KlondikePhasePlaying, KlondikePhaseGameClear, KlondikePhaseGameOver)
 *   - Baccarat:   internal/domain/Baccarat.go  (BaccaratPhaseBet, BaccaratPhaseEnd)
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

// Hearts phase constants (sync: internal/domain/Hearts.go)
export const HeartsPhase = {
  PASS: 0,
  PLAY: 1,
  TRICK_END: 2,
  ROUND_END: 3,
  GAME_END: 4,
} as const;

// Memory phase constants (sync: internal/domain/Memory.go)
export const MemoryPhase = {
  FLIP1: 0,
  FLIP2: 1,
  RESULT: 2,
  GAME_END: 3,
} as const;

// Klondike phase constants (sync: internal/domain/Klondike.go)
export const KlondikePhase = {
  PLAYING: 0,
  GAME_CLEAR: 1,
  GAME_OVER: 2,
} as const;

// Baccarat bet type constants (sync: internal/domain/Baccarat.go)
export const BaccaratBetType = {
  PLAYER: 0,
  BANKER: 1,
  TIE: 2,
} as const;

// Baccarat phase constants (sync: internal/domain/Baccarat.go)
export const BaccaratPhase = {
  BET: 1,
  END: 2,
} as const;
