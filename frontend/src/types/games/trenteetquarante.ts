// Type declarations for trenteetquarante. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/**
 * Trente et Quarante (Rouge et Noir) game state response.
 *
 * A pure banking game with no player card decisions: the player picks one of
 * four even-money bets (Noir, Rouge, Couleur, Inverse) plus a stake, then the
 * dealer immediately deals two rows — Noir (black) first, then Rouge (red) —
 * each summed until the total reaches 31 or more. The row with the LOWER total
 * wins. Betting resolves the round in one step, so there are only two phases:
 * Bet (0) and Result (1).
 */
export interface TrenteEtQuaranteResponse extends BaseGameResponse {
  /** Game phase: 0=Bet, 1=Result. */
  phase: number;
  /** Number of rounds resolved so far. */
  roundNumber: number;
  /** Player's remaining chip stack. */
  chips: number;
  /** Selected bet: 0=Noir, 1=Rouge, 2=Couleur, 3=Inverse. */
  currentBet: number;
  /** Amount wagered on the current round. */
  stake: number;
  /** Cards dealt to the Noir (black) row, summed until the total reaches 31+. */
  noirRow: Card[];
  /** Cards dealt to the Rouge (red) row, summed until the total reaches 31+. */
  rougeRow: Card[];
  /** Pip total of the Noir row (31–40 once the row is complete). */
  noirTotal: number;
  /** Pip total of the Rouge row (31–40 once the row is complete). */
  rougeTotal: number;
  /** Winning row (the LOWER total): 0=Noir, 1=Rouge; -1 for none (push/refait). */
  winningRow: number;
  /** Whether the first card dealt (Noir row's first card) is red — drives Couleur/Inverse display. */
  firstCardRed: boolean;
  /** True when both rows tie at 31 (a "Refait" — half the stake goes to the house). */
  refait: boolean;
  /** Round result from the player's perspective: 1=win, 0=push, -1=lose. */
  result: number;
  /** Gross chips returned to the stack this round (0=lose, stake/2=refait, stake=push, stake*2=win). */
  payout: number;
  /** Number of cards remaining in the shoe. */
  remainingDeck: number;
  /** True once the round has resolved. */
  gameEndFlag: boolean;
  /** Educational hint offered during the bet phase. */
  hint?: {
    /** Suggested bet type: 0=Noir, 1=Rouge, 2=Couleur, 3=Inverse. */
    bet: number;
    /** i18n reason key for the suggestion. */
    reason: string;
  };
  /** Local-rule configuration. */
  config: {
    /** Bet type pre-selected at the start of each round. */
    defaultBet: number;
  };
}
