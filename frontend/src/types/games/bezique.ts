// Type declarations for bezique. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Bezique game phase (0=Play 1=Meld 2=RoundEnd 3=GameEnd). */
export type BeziquePhaseValue = 0 | 1 | 2 | 3;

/** A Bezique player's public/own state. Cards are non-empty only for the human. */
export interface BeziquePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /** Points scored from melds and brisques in the current deal. */
  roundScore: number;
  /** Cumulative match score accumulated across deals. */
  cumulativeScore: number;
  trickCount: number;
}

/** A card played into the current Bezique trick. */
export interface BeziqueTrickCard {
  playerIdx: number;
  card: Card;
}

/**
 * A meld the trick winner may declare during the Meld phase. `type` is the
 * meld kind (0=marriage 1=Bezique 2=four aces 3=four kings 4=four queens
 * 5=four jacks); `suit` is the marriage suit (1=♠ 2=♣ 3=♥ 4=♦, or -1 for
 * Bezique and four-of-a-kind); `points` is the score it would award.
 */
export interface BeziqueMeld {
  type: number;
  suit: number;
  points: number;
}

/** Bezique game configuration. */
export interface BeziqueConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for Bezique, computed by the backend (may carry a card index OR a meld index, where -1 = skip). */
export interface BeziqueHint {
  /** Card index to play (Play phase). */
  cardIndex?: number;
  /** Meld index to declare (Meld phase); -1 means skip the meld. */
  meldIndex?: number;
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Server response for the Bezique game (2 players, melds, two-phase trick play). */
export interface BeziqueResponse extends BaseGameResponse {
  players: BeziquePlayer[];
  /** Points scored in the current deal, indexed by seat. */
  dealPoints: number[];
  /** Of the deal points, the portion from melds (trick portion = dealPoints - dealMeldPoints). */
  dealMeldPoints: number[];
  /** Cumulative match score, indexed by seat. */
  matchScore: number[];
  phase: BeziquePhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Trump suit (0=♠ 1=♣ 2=♥ 3=♦ — the deck's suit ordinal). */
  trumpSuit: number;
  /** The face-up card that fixed the trump (present until the stock empties). */
  trumpCard?: Card;
  currentTrick: BeziqueTrickCard[];
  /** Cards remaining in the stock (phase 2 begins when this reaches 0). */
  stockRemaining: number;
  /** Whether the deal has entered the strict must-follow endgame (phase 2). */
  isEndgame: boolean;
  /** Melds the human may declare this Meld phase (empty otherwise). */
  availableMelds: BeziqueMeld[];
  gameEndFlag: boolean;
  /** Winning seat index (0 or 1), or -1 until the game ends. */
  winnerIdx: number;
  hint?: BeziqueHint | null;
  config: BeziqueConfig;
}

// --- Écarté ---
