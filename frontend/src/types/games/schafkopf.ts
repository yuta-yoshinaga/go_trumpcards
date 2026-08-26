// Type declarations for schafkopf. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Schafkopf game phase (0=Pick 1=Call 2=Play 3=TrickEnd 4=RoundEnd 5=GameEnd). */
export type SchafkopfPhase = 0 | 1 | 2 | 3 | 4 | 5;

/** The contract in play (0=Rufspiel 1=Wenz 2=Solo). Decides what counts as trump. */
export type SchafkopfContract = 0 | 1 | 2;

/** A Schafkopf player's public/own state. Cards are non-empty only for the human. */
export interface SchafkopfPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  chips: number;
}

/** A card played into the current Schafkopf trick. */
export interface SchafkopfTrickCard {
  playerIdx: number;
  card: Card;
}

/** Schafkopf game configuration. */
export interface SchafkopfConfig {
  cpuDifficulty: number;
  baseChips: number;
  startChips: number;
  targetChips: number;
}

/** A suggested hint for Schafkopf, computed by the backend. */
export interface SchafkopfHint {
  cardIndices: number[];
  /** Suggested called suit (0=none, 1=♠, 2=♣, 4=♦). Relevant in the Call phase. */
  suit: number;
  /** Whether the hint recommends declaring a contract (Pick phase). */
  pick: boolean;
  reason: string;
}

/** Full Schafkopf game state returned from the API. */
export interface SchafkopfResponse extends BaseGameResponse {
  players: SchafkopfPlayer[];
  phase: SchafkopfPhase;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  currentTrick: SchafkopfTrickCard[];
  /** Index of the picker, or -1 until decided. */
  pickerIdx: number;
  /** Contract in play. Rufspiel until someone declares otherwise. */
  contract: SchafkopfContract;
  /** Trump suit chosen for a Solo (1=♠ 2=♣ 3=♥ 4=♦); 0 for other contracts. */
  soloSuit: number;
  /**
   * Contracts this seat may still declare — anything that outranks the
   * standing bid. Empty once the auction has closed.
   */
  beatableContracts: SchafkopfContract[];
  /** Index of the picker's partner, or -1 until revealed/round end. */
  partnerIdx: number;
  /** Called partner suit (0=none, 1=♠, 2=♣, 4=♦); ♥ is trump, so never callable. */
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
  hint?: SchafkopfHint | null;
  config: SchafkopfConfig;
}

// --- Mus ---
