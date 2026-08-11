// Type declarations for polignac. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Polignac trick. */
export interface PolignacTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a Polignac table. */
export interface PolignacPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. The only hand information exposed for the CPU. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** Running penalty total across rounds. **Lower is better.** */
  score: number;
  /** Penalty taken in the current round. */
  roundPenalty: number;
  trickCount: number;
  /** Declared capot this round. */
  declaredCapot: boolean;
}

/** A suggested card to play. */
export interface PolignacHint {
  cardIndex?: number;
  /**
   * Why that card: `polignacAvoidJack` (a jack is on the table),
   * `polignacDumpJack` (safe trick — shed a jack now), `polignacLeadSafe`
   * (leading; play a low non-jack), `polignacBlockCapot` (take this trick to
   * break someone else's capot), or `polignacWinCapot` (you declared capot —
   * take every trick).
   */
  reason: string;
}

/** Round-count setting. */
export interface PolignacConfig {
  /** Rounds to play (1..12, default 4). */
  rounds: number;
}

/** Full Polignac game state returned from the API. */
export interface PolignacResponse extends BaseGameResponse {
  players: PolignacPlayer[];
  /** `0` = Declare, `1` = Play, `2` = RoundEnd, `3` = GameEnd. */
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat that declared capot, or `-1` when nobody did. */
  capotIdx: number;
  /** Tricks the capot declarer has taken so far. */
  capotTricks: number;
  currentTrick: PolignacTrickCard[];
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  gameEndFlag: boolean;
  /** `-1` until decided, and `-1` on a tie. */
  winnerIdx: number;
  hint?: PolignacHint;
  config: PolignacConfig;
}
