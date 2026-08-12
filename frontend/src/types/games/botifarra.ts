// Type declarations for botifarra. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Trump value meaning "no trump" (botifarra). Not a suit — suits are 1..4. */
export const BOTIFARRA_NO_TRUMP = -1;

/** Points that move in one round: 60 from cards plus one per trick. */
export const BOTIFARRA_TOTAL_POINTS = 72;

/** Half of the round's points. Only the excess over this scores. */
export const BOTIFARRA_HALF_POINTS = 36;

/** One seat at a Botifarra table. */
export interface BotifarraPlayer {
  id: number;
  isHuman: boolean;
  /** 0 = seats 0 and 2, 1 = seats 1 and 3. Partners sit opposite. */
  team: number;
  cardCount: number;
  /**
   * The seat's hand. **Only the human seat carries its cards** — the others
   * arrive empty so the hand cannot be read off the wire.
   */
  cards: Card[];
  trickCount: number;
}

/** A card played into the current trick. */
export interface BotifarraTrickCard {
  playerIdx: number;
  card: Card;
}

/** A suggestion. Either a trump to declare or a card to play, never both. */
export interface BotifarraHint {
  cardIndex?: number;
  suit?: number;
  /** `botifarraDeclareLongest` or `botifarraMustWin`. */
  reason: string;
}

/** Botifarra game settings. */
export interface BotifarraConfig {
  targetScore: number;
  allowDoubling: boolean;
}

/**
 * Response payload for `/botifarra/exec`.
 *
 * **`validPlays` is often a strict subset of the hand.** Beyond following suit,
 * a player who can beat the trick must do so unless their partner is winning.
 */
export interface BotifarraResponse extends BaseGameResponse {
  players: BotifarraPlayer[];
  /** 0=Declare, 1=Delegated, 2=Double, 3=Play, 4=RoundEnd, 5=GameEnd. */
  phase: number;
  /** Hand indices the human may play. Always an array. */
  validPlays: number[];
  dealerIdx: number;
  /** -1 until trump is named. */
  declarerIdx: number;
  /** 1..4, or -1 for no trump. */
  trumpSuit: number;
  /** 1, 2 (contrar) or 4 (recontrar). */
  multiplier: number;
  currentTurn: number;
  isHumanTurn: boolean;
  currentTrick: BotifarraTrickCard[];
  lastTrick: BotifarraTrickCard[];
  lastTrickWinner: number;
  trickCount: number;
  /** Points taken this round per team; they sum to 72 once the round is dealt out. */
  roundPoints: number[];
  scores: number[];
  gameEndFlag: boolean;
  /** -1 until decided. */
  winnerTeam: number;
  config?: BotifarraConfig;
}
