// Type declarations for dehlapakad. Follows the split-out convention of
// card.ts (issue #4366); card.ts re-exports this file.

import type { BaseGameResponse, Card } from '../common';

/** One seat at the four-handed table. Cards are non-empty only for the human. */
export interface DehlaPakadPlayer {
  id: number;
  isHuman: boolean;
  /** 0 or 1. **Partners sit opposite**, so your neighbours are always opponents. */
  team: number;
  cardCount: number;
  cards: Card[];
  /** How many cards this seat has gathered in from the centre pile. */
  gatheredCount: number;
  isDealer: boolean;
  /** True for the seat that calls the trump (the dealer's right). */
  isTrumpChooser: boolean;
}

/** One hand's result. */
export interface DehlaPakadHand {
  winnerTeam: number;
  /** Tens taken per team. */
  teamTens: number[];
  kot: boolean;
  /** "allTens" | "streak" | "". */
  kotReason: string;
  dealerIdx: number;
  trumpSuit: number;
}

/** Dehla Pakad game configuration. */
export interface DehlaPakadConfig {
  cpuDifficulty: number;
  /** Kots needed to take the match, 1-5. */
  targetKots: number;
}

/** A suggested hint, computed by the backend. */
export interface DehlaPakadHint {
  cardIndices: number[];
  /** i18n reason identifier. */
  reason: string;
}

/**
 * Full Dehla Pakad game state returned from the API.
 *
 * An Indian partnership trick-taker decided by the four tens. **Winning a trick
 * does not give you its cards**: they pile up in the centre and are gathered in
 * only when the same player wins two consecutive tricks.
 */
export interface DehlaPakadResponse extends BaseGameResponse {
  players: DehlaPakadPlayer[];
  /** "selectTrump" | "play" | "handEnd" | "gameEnd". */
  phase: string;
  handNumber: number;
  dealerIdx: number;
  /** The seat that calls the trump — the dealer's right. */
  trumpChooserIdx: number;
  /** Trump suit (1-4), or -1 while it is still being called. */
  trumpSuit: number;
  /** Stable key for the trump suit ("spade", "heart", ...), for i18n. */
  trumpSuitName: string;
  trickNumber: number;
  trickCount: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  currentTrick: { playerIdx: number; card: Card }[];
  lastTrick: { playerIdx: number; card: Card }[];
  lastTrickWinner: number;
  /**
   * Who won the previous trick.
   *
   * **They take the whole centre pile by winning the next one too** — this is
   * what makes a trick worth contesting, so a client should surface it.
   */
  prevTrickWinner: number;
  /** Cards nobody has gathered in yet. */
  centrePileCount: number;
  /** How many of those are tens. */
  centrePileTens: number;
  playableIndices: number[];
  /** Tens taken per team this hand. */
  teamTens: number[];
  /** Kots per team — this is the match score. */
  teamKots: number[];
  humanTeam: number;
  /** Team on a winning streak, or -1. */
  streakTeam: number;
  streakCount: number;
  lastHand?: DehlaPakadHand | null;
  handHistory: DehlaPakadHand[];
  gameEndFlag: boolean;
  winnerTeam: number;
  isHumanTurn: boolean;
  isTrumpPhase: boolean;
  hint?: DehlaPakadHint | null;
  /** Trump suit the backend suggests while calling, or -1. */
  hintTrumpSuit: number;
  config: DehlaPakadConfig;
}
