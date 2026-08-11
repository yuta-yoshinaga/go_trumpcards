// Type declarations for hasenpfeffer. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Hasenpfeffer trick. */
export interface HasenpfefferTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a Hasenpfeffer table. */
export interface HasenpfefferPlayer {
  id: number;
  isHuman: boolean;
  /** `0` or `1`. Seats 0+2 are one partnership, 1+3 the other. */
  team: number;
  /** Hand size. The only hand information exposed for the CPU. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** Tricks bid, `0` if this seat passed, or `-1` before it answered. */
  bid: number;
  trickCount: number;
}

/**
 * A suggestion. While bidding it carries no `cardIndex` and puts the
 * recommended number in `value`; while discarding it names both a card and the
 * suit to make trump; during play it names a card.
 */
export interface HasenpfefferHint {
  cardIndex?: number;
  /**
   * `hasenpfefferBid` / `hasenpfefferPass` / `hasenpfefferMustBid` while
   * bidding; `hasenpfefferDiscard` while sorting the blind;
   * `hasenpfefferWinTrick` or `hasenpfefferFeedPartner` during play.
   */
  reason: string;
  /** Tricks to bid. `0` otherwise. */
  value: number;
  /** Suit to make trump when discarding. `0` otherwise. */
  suit: number;
}

/** Target-score setting. */
export interface HasenpfefferConfig {
  /** Points needed to win (5..50, default 10). */
  target: number;
}

/** Full Hasenpfeffer game state returned from the API. */
export interface HasenpfefferResponse extends BaseGameResponse {
  players: HasenpfefferPlayer[];
  /** `0` = Bid, `1` = Discard, `2` = Play, `3` = HandEnd, `4` = GameEnd. */
  phase: number;
  handNumber: number;
  trickNumber: number;
  /** `0` until the declarer names it along with the discard. */
  trumpSuit: number;
  /** Seat that won the bidding, or `-1` while the auction runs. */
  declarerIdx: number;
  /** Tricks contracted for. */
  contract: number;
  /**
   * Lowest bid that would beat the standing one. **`0` means nobody can bid
   * any more** — the maximum is already standing, so passing is the only move.
   */
  minBid: number;
  /**
   * You are the dealer and everybody else has passed, so **you cannot pass**.
   * Bidding in this game is compulsory; a hand never gets thrown in.
   */
  mustBid: boolean;
  /** Cards still face down; `0` once the declarer has taken the blind. */
  blindSize: number;
  /** Running total per team, index 0 and 1. */
  scores: number[];
  /** Tricks taken this hand per team, index 0 and 1. */
  teamTricks: number[];
  /** The declaring side failed its contract in the hand just scored. */
  lastHandEuchred: boolean;
  /** Tricks the declaring side took in the hand just scored. */
  lastHandTricks: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  currentTrick: HasenpfefferTrickCard[];
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  gameEndFlag: boolean;
  /** `-1` until decided, and `-1` on a tie. */
  winnerTeam: number;
  hint?: HasenpfefferHint;
  config: HasenpfefferConfig;
}
