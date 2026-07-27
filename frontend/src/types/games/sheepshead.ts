// Type declarations for sheepshead. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Sheepshead game phase (0=Pick 1=Bury 2=Call 3=Play 4=TrickEnd 5=RoundEnd 6=GameEnd). */
export type SheepsheadPhase = 0 | 1 | 2 | 3 | 4 | 5 | 6;

/** A Sheepshead player's public/own state. Cards are non-empty only for the human. */
export interface SheepsheadPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  chips: number;
}

/** A card played into the current Sheepshead trick. */
export interface SheepsheadTrickCard {
  playerIdx: number;
  card: Card;
}

/** Sheepshead game configuration. */
export interface SheepsheadConfig {
  cpuDifficulty: number;
  baseChips: number;
  startChips: number;
  targetChips: number;
}

/** A suggested hint for Sheepshead, computed by the backend. */
export interface SheepsheadHint {
  cardIndices: number[];
  /** Suggested called suit (0=none, 1=♠, 2=♣, 3=♥). Relevant in the Call phase. */
  suit: number;
  /** Whether the hint recommends picking the blind (Pick phase). */
  pick: boolean;
  reason: string;
}

/** Full Sheepshead game state returned from the API. */
export interface SheepsheadResponse extends BaseGameResponse {
  players: SheepsheadPlayer[];
  phase: SheepsheadPhase;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  currentTrick: SheepsheadTrickCard[];
  /** Number of cards in the blind (only the count is exposed during the Pick phase). */
  blindCount: number;
  /** The two buried cards; empty until RoundEnd/GameEnd. */
  buried: Card[];
  /** Index of the picker, or -1 until decided. */
  pickerIdx: number;
  /** Index of the picker's partner, or -1 until revealed/round end. */
  partnerIdx: number;
  /** Called partner suit (0=none, 1=♠, 2=♣, 3=♥). */
  calledSuit: number;
  /** Whether the partner has been revealed. */
  partnerRevealed: boolean;
  /** Number of players who have passed in the Pick phase. */
  passCount: number;
  /** Suits the picker may call this turn (non-empty only in the Call phase). */
  callableSuits: number[];
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  /** Card points captured by the picker's team this round. */
  roundPickerPoints: number;
  /** Score multiplier applied to this round's result. */
  roundMultiplier: number;
  /** Whether the picker's team won the round. */
  roundPickerWon: boolean;
  gameEndFlag: boolean;
  /** Winning player index, or -1 until the game ends. */
  winnerIdx: number;
  hint?: SheepsheadHint | null;
  config: SheepsheadConfig;
}

// --- Mus ---
