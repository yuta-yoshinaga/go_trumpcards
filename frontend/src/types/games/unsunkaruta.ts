// Type declarations for unsunkaruta. Follows the split-out convention of
// card.ts (issue #4366); card.ts re-exports this file.

import type { BaseGameResponse, Card } from '../common';

/** Unsun Karuta phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type UnsunKarutaPhaseValue = 0 | 1 | 2 | 3;

/** One seat at the eight-handed table. Cards are non-empty only for the human. */
export interface UnsunKarutaPlayer {
  id: number;
  isHuman: boolean;
  /**
   * 0 or 1.
   *
   * **The seat number alone does not say who is on your side** — the two teams
   * sit alternately, so your neighbours are always opponents.
   */
  team: number;
  cardCount: number;
  cards: Card[];
  /** Tricks ("ko") this seat has taken in the current deal. */
  trickCount: number;
  isDealer: boolean;
}

/** A card played into the current trick. */
export interface UnsunKarutaTrickCard {
  playerIdx: number;
  card: Card;
}

/** Unsun Karuta game configuration. */
export interface UnsunKarutaConfig {
  cpuDifficulty: number;
  /** Deals per match, 1-8 (eight seats, one deal each). */
  targetDeals: number;
}

/** A suggested hint, computed by the backend. */
export interface UnsunKarutaHint {
  cardIndices: number[];
  /** i18n reason identifier. */
  reason: string;
}

/**
 * Full Unsun Karuta (八人メリ) game state returned from the API.
 *
 * The oldest surviving Japanese trick-taking game: 75 cards in five suits of
 * fifteen, eight players in two teams of four. The pip order flips with the
 * suit — nine is highest in the long suits, one is highest in the round suits —
 * and following suit is only required once the leader declares (meri/monchi).
 */
export interface UnsunKarutaResponse extends BaseGameResponse {
  players: UnsunKarutaPlayer[];
  phase: UnsunKarutaPhaseValue;
  roundNumber: number;
  trickNumber: number;
  /** Tricks in a deal (nine). */
  trickCount: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Team the human sits on. */
  humanTeam: number;
  /** Trump suit for this deal (1-5). */
  trumpSuit: number;
  /** Stable key for the trump suit ("pao", "kuru", ...), for i18n. */
  trumpSuitName: string;
  /** The card turned up to set the trump. */
  trumpCard?: Card | null;
  /** True while this trick carries the follow obligation (a declaration was made). */
  mustFollow: boolean;
  /** True when the declaration happened on this trick. */
  declared: boolean;
  /** True while the human may declare (they are on lead). */
  canDeclare: boolean;
  currentTrick: UnsunKarutaTrickCard[];
  /** Tricks taken per team in the current deal. */
  teamTricks: number[];
  /** Cumulative tricks per team across the match. */
  teamScores: number[];
  /** Seat that took the last trick, or -1. */
  lastTrickWinner: number;
  /** Match result from the human's perspective (0=none/draw, 1=win, 2=lose). */
  result: number;
  /** Indices in the human's hand that are legal to play. */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning team, or -1 for a draw / undecided. */
  winnerTeam: number;
  isHumanTurn: boolean;
  hint?: UnsunKarutaHint | null;
  config: UnsunKarutaConfig;
}
