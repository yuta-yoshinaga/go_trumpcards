// Type declarations for allfours. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** All Fours player data (2-player: 0 = human elder hand, 1 = CPU dealer). */
export interface AllFoursPlayerData {
  /** Player index (0 = non-dealer/human, 1 = dealer/CPU). */
  id: number;
  /** Whether this player is the human. */
  isHuman: boolean;
  /** Number of cards in hand. */
  cardCount: number;
  /** Cards in hand (only populated for the human). */
  cards: Card[];
  /** Points scored this deal so far. */
  roundScore: number;
  /** Cumulative game score. */
  cumulativeScore: number;
  /** Number of tricks captured this deal. */
  trickCount: number;
}

/** A single card played to the current All Fours trick. */
export interface AllFoursTrickCard {
  playerIdx: number;
  card: Card;
}

/** All Fours hint payload (one of card/beg/run is set). */
export interface AllFoursHint {
  cardIndex?: number;
  beg?: boolean;
  run?: boolean;
  reason: string;
}

/** All Fours game configuration. */
export interface AllFoursConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A High/Low point award: the capturing player and the trump card, or -1 if unawarded. */
export interface AllFoursBreakdownAward {
  winnerIdx: number;
  card: Card | null;
}

/** The round-end point breakdown for All Fours (High/Low/Jack/Game). */
export interface AllFoursRoundBreakdown {
  high: AllFoursBreakdownAward;
  low: AllFoursBreakdownAward;
  /** Captor of the trump Jack, or -1 if no trump Jack was in play. */
  jack: { winnerIdx: number };
  /** Game point (most card pips): winner (-1 on tie/zero) and per-player pip totals. */
  game: { winnerIdx: number; points: number[] };
  /**
   * Whether these are mid-round provisional values rather than the settled result.
   *
   * **High and Low are decided by the trumps that appear over the whole round,**
   * so a trump still in someone's hand can take either away. The page labels a
   * provisional table so it is not read as final (#4771).
   */
  provisional: boolean;
}

/** Full All Fours game state returned from the API. */
export interface AllFoursResponse extends BaseGameResponse {
  players: AllFoursPlayerData[];
  /** Current phase (0=Beg, 1=Gift, 2=Play, 3=TrickEnd, 4=RoundEnd, 5=GameEnd). */
  phase: number;
  roundNumber: number;
  trickNumber: number;
  dealerIdx: number;
  nonDealerIdx: number;
  currentPlayerIdx: number;
  trumpSuit: number;
  /** The turn-up card that set the provisional trump, or null. */
  turnUp: Card | null;
  /** Number of "run the cards" attempts this deal. */
  runCount: number;
  currentTrick: AllFoursTrickCard[];
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  validPlayIndices: number[];
  /** Present only at ROUND_END / GAME_END: the High/Low/Jack/Game point breakdown. */
  /** Present during PLAY (provisional) and at ROUND_END / GAME_END (settled). */
  roundBreakdown?: AllFoursRoundBreakdown;
  config: AllFoursConfig;
  hint?: AllFoursHint;
}

// --- Guts ---
