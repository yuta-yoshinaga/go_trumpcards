// Type declarations for doubleattack. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Phases, matching the Go domain. */
export const DOUBLE_ATTACK_PHASE = { bet: 0, attack: 1, play: 2, result: 3 } as const;

/** Per-hand outcomes, matching the Go domain. */
export const DOUBLE_ATTACK_RESULT = {
  none: 0,
  win: 1,
  lose: 2,
  push: 3,
  blackjack: 4,
} as const;

/** Maximum hands after splitting. */
export const DOUBLE_ATTACK_MAX_HANDS = 4;

/** A suggestion for the current decision. */
export interface DoubleAttackHint {
  /** `attack`, `hit`, `stand`, `double` or `split`. */
  action: string;
  reason: string;
}

/** One player hand. Splitting produces more of these. */
export interface DoubleAttackHand {
  cards: Card[];
  score: number;
  bet: number;
  isSoft: boolean;
  stood: boolean;
  doubled: boolean;
  busted: boolean;
  /** Twenty-one on the first two cards. **Pays 1:1, not 3:2.** */
  blackjack: boolean;
  /** 0=none, 1=win, 2=lose, 3=push, 4=blackjack. */
  result: number;
}

/** Extra Bet Blackjack game settings. */
export interface DoubleAttackConfig {
  initialChips: number;
  defaultAnte: number;
}

/** Response payload for `/doubleattack/exec`. */
export interface DoubleAttackResponse extends BaseGameResponse {
  /** 0=Bet, 1=Attack, 2=Play, 3=Result. */
  phase: number;
  hands: DoubleAttackHand[];
  activeHand: number;
  /**
   * The dealer's cards.
   *
   * **Length 1 until the extra bet is placed** — the server does not hold a
   * second dealer card before that, so this is absence rather than masking.
   */
  dealerCards: Card[];
  /** Zero until the second dealer card exists (a one-card score is a hole-card hint). */
  dealerScore: number;
  dealerHoleDealt: boolean;
  /**
   * The most that may be added after seeing the up-card — at most the ante.
   *
   * **Do not re-derive this.** The rule lives in the domain.
   */
  maxAttackBet: number;
  canDouble: boolean;
  canSplit: boolean;
  anteBet: number;
  attackBet: number;
  bustItBet: number;
  payout: number;
  bustItPayout: number;
  chips: number;
  roundNumber: number;
  remainingCards: number;
  gameEndFlag: boolean;
  config?: DoubleAttackConfig;
}
