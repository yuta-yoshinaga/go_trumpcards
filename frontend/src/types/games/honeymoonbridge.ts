// Type declarations for honeymoonbridge. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Honeymoon Bridge trick. */
export interface HoneymoonBridgeTrickCard {
  playerIdx: number;
  card: Card;
}

/** One of the two seats at a Honeymoon Bridge table. */
export interface HoneymoonBridgePlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. The only hand information exposed for the CPU. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** This seat's latest bid level, or `0` for a pass. */
  bidLevel: number;
  /** This seat's latest bid suit. `0` is no-trump, not "unset". */
  bidSuit: number;
  trickCount: number;
  score: number;
}

/**
 * A suggestion. During the auction it names a contract in `level`/`suit` and
 * carries no `cardIndex`; while playing it names one card.
 */
export interface HoneymoonBridgeHint {
  cardIndex?: number;
  /**
   * `honeymoonbridgeDraw` in the draw phase (tricks there do not score),
   * `honeymoonbridgeBid` or `honeymoonbridgePass` during the auction,
   * `honeymoonbridgeWinTrick` while playing the contract.
   */
  reason: string;
  /** Level to bid, `0` to pass. `0` outside the auction. */
  level: number;
  /** Suit to bid — `0` is no-trump. */
  suit: number;
}

/** Target-score setting. */
export interface HoneymoonBridgeConfig {
  /** Points that end the game (50..500, default 100). */
  target: number;
}

/** Full Honeymoon Bridge game state returned from the API. */
export interface HoneymoonBridgeResponse extends BaseGameResponse {
  players: HoneymoonBridgePlayer[];
  /** `0` = Draw, `1` = Bid, `2` = Play, `3` = RoundEnd, `4` = GameEnd. */
  phase: number;
  roundNumber: number;
  trickNumber: number;
  /**
   * Cards left in the stock. 26 at the deal, two fewer after each draw-phase
   * trick, and `0` from the auction onwards.
   */
  stockSize: number;
  /** `0` means no-trump — during the draw phase and for a no-trump contract. */
  trumpSuit: number;
  /** `-1` until a contract is bought. */
  declarerIdx: number;
  /** `0` until a contract is bought. */
  contractLevel: number;
  /** Tricks the contract needs: `6 + contractLevel`. `0` before the auction. */
  requiredTricks: number;
  /**
   * The lowest bid that still outbids the table, so a client never has to
   * rederive the auction's suit order (which puts no-trump above diamonds).
   * `0` when no bid can outbid the table — pass is the only move — and `0`
   * outside the auction.
   */
  minBidLevel: number;
  /** Suit that goes with `minBidLevel`; `0` is no-trump. */
  minBidSuit: number;
  /** Whether the last deal's contract was made. */
  lastMade: boolean;
  /** Tricks the declarer took in the last deal. */
  lastTricks: number;
  /**
   * Points the deal that just ended actually moved.
   *
   * Scoring is level x 10 plus 5 an overtrick (or 10 a trick short to the
   * opponent), so the trick counts alone do not say what it was worth (#5760).
   */
  lastPoints?: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  currentTrick: HoneymoonBridgeTrickCard[];
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  gameEndFlag: boolean;
  /** `-1` until decided. */
  winnerIdx: number;
  hint?: HoneymoonBridgeHint;
  config: HoneymoonBridgeConfig;
}
