// Type declarations for ramsch. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A Ramsch player's per-round state. */
export interface RamschPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /** Card points collected this round. **The most loses that many.** */
  cardPoints: number;
  roundsWon: number;
  roundsLost: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in a Ramsch trick. */
export interface RamschTrickCard {
  playerIdx: number;
  card: Card;
}

/** Ramsch game configuration. */
export interface RamschConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/**
 * A suggested hint for Ramsch.
 *
 * Only a card index: there is no auction and no contract to advise on.
 */
export interface RamschHint {
  cardIndex?: number;
  reason: string;
}

/** Full Ramsch game state returned from the API. */
export interface RamschResponse extends BaseGameResponse {
  players: RamschPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: RamschTrickCard[];
  forehandIdx: number;
  middlehandIdx: number;
  rearhandIdx: number;
  dealerIdx: number;
  /**
   * The two face-down cards. **Only sent once the round is over** — they go to
   * whoever wins the last trick, so revealing them earlier would turn the
   * endgame into perfect information.
   */
  skat?: Card[];
  /**
   * Who took the most card points and therefore loses them. `-1` while the
   * round is running, on a tie (everyone tied loses), and on a Durchmarsch.
   */
  loserIdx: number;
  /** Whether one player swept all ten tricks, which reverses the result. */
  durchmarsch: boolean;
  /** Who swept, or `-1`. */
  durchmarschIdx: number;
  gameEndFlag: boolean;
  leadPlayerIdx: number;
  config: RamschConfig;
  hint?: RamschHint;
}

// --- Shithead / Karma (シットヘッド / カーマ) ---
