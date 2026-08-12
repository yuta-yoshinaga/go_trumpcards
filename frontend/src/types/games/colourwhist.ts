// Type declarations for colourwhist. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Trump value meaning "no trump" — the Miserie contract. Suits are 1..4. */
export const COLOUR_WHIST_NO_TRUMP = -1;

/**
 * The contracts that can be **bid**, weakest first.
 *
 * Troel is deliberately absent: it is forced at deal time by holding three
 * aces and cannot be declared, so it never belongs on a bidding button.
 */
export const COLOUR_WHIST_BIDDABLE = [
  { contract: 1, key: 'contract.samen' },
  { contract: 2, key: 'contract.alleen' },
  { contract: 3, key: 'contract.miserie' },
] as const;

/** Troel's contract value. Shown, never bid. */
export const COLOUR_WHIST_TROEL = 4;

/** One seat at a Colour Whist table. */
export interface ColourWhistPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  /** Only the human seat carries its cards. */
  cards: Card[];
  trickCount: number;
  /** Running score. **Negative is normal** — scoring is zero-sum. */
  score: number;
  /** Sides are set by the contract, not by seat. */
  isDeclarerSide: boolean;
  hasPassed: boolean;
}

/** A card played into the current trick. */
export interface ColourWhistTrickCard {
  playerIdx: number;
  card: Card;
}

/** A suggestion: either a contract to bid or a card to play. */
export interface ColourWhistHint {
  contract?: number;
  cardIndex?: number;
  /** `colourWhistBidStrength` or `colourWhistFollowSuit`. */
  reason: string;
}

/** Colour Whist game settings. */
export interface ColourWhistConfig {
  rounds: number;
}

/** Response payload for `/colourwhist/exec`. */
export interface ColourWhistResponse extends BaseGameResponse {
  players: ColourWhistPlayer[];
  /** 0=Bid, 1=Call, 2=Play, 3=RoundEnd, 4=GameEnd. */
  phase: number;
  validPlays: number[];
  dealerIdx: number;
  /** 0=none, 1=Samen, 2=Alleen, 3=Miserie, 4=Troel. */
  contract: number;
  declarerIdx: number;
  /** -1 until known. Troel names it at deal time; Samen hides it until the called card is played. */
  partnerIdx: number;
  /** The card called to find a partner. Samen only — troel is dealt, not called. */
  calledCard?: Card;
  /** 1..4, or -1 for no trump. */
  trumpSuit: number;
  /** True when three aces in one hand forced the contract without an auction. */
  troelForced: boolean;
  currentTurn: number;
  isHumanTurn: boolean;
  currentTrick: ColourWhistTrickCard[];
  lastTrick: ColourWhistTrickCard[];
  lastTrickWinner: number;
  trickCount: number;
  declarerTricks: number;
  roundNumber: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  config?: ColourWhistConfig;
}
