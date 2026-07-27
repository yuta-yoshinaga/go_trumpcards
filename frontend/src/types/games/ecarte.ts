// Type declarations for ecarte. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Écarté game phase (0=Exchange 1=Play 2=RoundEnd 3=GameEnd). */
export type EcartePhaseValue = 0 | 1 | 2 | 3;

/**
 * Écarté negotiation sub-step within the Exchange phase
 * (0=ElderDecide 1=DealerRespond 2=ElderDiscard 3=DealerDiscard).
 */
export type EcarteNegStepValue = 0 | 1 | 2 | 3;

/** An Écarté player's public/own state. Cards are non-empty only for the human. */
export interface EcartePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /** Points scored in the current deal. */
  roundScore: number;
  /** Cumulative match score accumulated across deals. */
  cumulativeScore: number;
  trickCount: number;
}

/** A card played into the current Écarté trick. */
export interface EcarteTrickCard {
  playerIdx: number;
  card: Card;
}

/** Écarté game configuration. */
export interface EcarteConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/**
 * A suggested hint for Écarté, computed by the backend. During the Play phase
 * it carries a `cardIndex`; during the Exchange phase it carries an `action`
 * string (e.g. `propose`, `stand`, `accept`, `refuse`, `discard`).
 */
export interface EcarteHint {
  /** Card index to play (Play phase). */
  cardIndex?: number;
  /** Exchange-phase action identifier (Exchange phase). */
  action?: string;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Écarté game state returned from the API.
 *
 * Écarté is a 2-player French 32-card trick game. Before play, an Exchange
 * phase lets the elder (non-dealer) Propose or Stand; if proposed, the dealer
 * Accepts or Refuses; on accept, each player discards any number of cards and
 * draws replacements, then the elder decides again (repeating until the stock
 * empties). Play is 5 strict must-follow tricks (rank K>Q>J>A>10>9>8>7).
 * Winning 3+ tricks scores 1 point, all 5 (Vole) scores 2; holding the King of
 * trump scores +1, a turned King gives the dealer +1, and a dealer who refuses
 * then loses gives the elder +1. Scores accumulate to a target (default 5).
 */
export interface EcarteResponse extends BaseGameResponse {
  players: EcartePlayer[];
  /** Points scored in the current deal, indexed by seat. */
  dealPoints: number[];
  /** Cumulative match score, indexed by seat. */
  matchScore: number[];
  phase: EcartePhaseValue;
  /** Exchange-phase negotiation sub-step (only meaningful in phase 0). */
  negStep: EcarteNegStepValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the elder (non-dealer) player. */
  elderIdx: number;
  leadPlayerIdx: number;
  /** Trump suit (1=♠ 2=♣ 3=♥ 4=♦; 0=undeclared). */
  trumpSuit: number;
  /** The face-up card that fixed the trump (present until the stock empties). */
  trumpCard?: Card;
  currentTrick: EcarteTrickCard[];
  /** Cards remaining in the stock. */
  stockRemaining: number;
  /** Whether the dealer refused the most recent exchange proposal. */
  refusalByDealer: boolean;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  validPlays: number[];
  gameEndFlag: boolean;
  /** Winning seat index (0 or 1), or -1 until the game ends. */
  winnerIdx: number;
  hint?: EcarteHint | null;
  config: EcarteConfig;
}

// --- Three Card Brag ---
