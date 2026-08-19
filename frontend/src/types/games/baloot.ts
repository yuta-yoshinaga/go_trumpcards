// Type declarations for baloot. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Baloot trick. */
export interface BalootTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a Baloot table. */
export interface BalootPlayer {
  id: number;
  isHuman: boolean;
  /** `0` or `1`. Seats 0+2 are one partnership, 1+3 the other. */
  team: number;
  /** Hand size. The only hand information exposed for the CPU. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** Holds the trump K and Q — worth 20, and possible in Hokom only. */
  hasBaloot: boolean;
  /**
   * Whether this seat's Baloot standing is public yet.
   *
   * The server sends `hasBaloot: false` while it is hidden, so read this first —
   * "no Baloot" and "not shown yet" are different things (#5750). A missing
   * field reads as hidden: showing a concealed Baloot would be the worse failure.
   */
  balootRevealed?: boolean;
  /** Has already declared or passed this round. */
  declared: boolean;
  trickCount: number;
}

/**
 * A suggestion. While declaring it advises a mode and carries no `cardIndex`
 * (naming the suit in `suit` when it recommends Hokom); during play it names a
 * card.
 */
export interface BalootHint {
  cardIndex?: number;
  /**
   * `balootDeclareSun` / `balootDeclareHokom` / `balootPassDeclare` while
   * declaring; `balootWinTrick` or `balootFeedPartner` (your partner is
   * winning — throw points on it) during play.
   */
  reason: string;
  /** Trump suit to declare when `reason` is `balootDeclareHokom`; `0` otherwise. */
  suit: number;
}

/** Target-score setting. */
export interface BalootConfig {
  /** Points needed to win (50..500, default 152). */
  target: number;
}

/** Full Baloot game state returned from the API. */
export interface BalootResponse extends BaseGameResponse {
  players: BalootPlayer[];
  /** `0` = Declare, `1` = Play, `2` = RoundEnd, `3` = GameEnd. */
  phase: number;
  /**
   * `0` = undeclared, `1` = Sun, `2` = Hokom. **The mode decides the rank
   * order itself**, not merely whether a trump exists.
   */
  mode: number;
  roundNumber: number;
  trickNumber: number;
  /** Meaningful in Hokom only; `0` under Sun, which has no trump. */
  trumpSuit: number;
  /** Seat that declared the mode, or `-1` before it is settled. */
  declarerIdx: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Running total per team, index 0 and 1. */
  scores: number[];
  /** Points taken this round per team, index 0 and 1. */
  roundPoints: number[];
  currentTrick: BalootTrickCard[];
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  gameEndFlag: boolean;
  /** `-1` until decided, and `-1` on a tie. */
  winnerTeam: number;
  hint?: BalootHint;
  config: BalootConfig;
}
