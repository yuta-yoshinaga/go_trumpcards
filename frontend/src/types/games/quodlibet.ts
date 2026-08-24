// Type declarations for quodlibet. Follows the split-out convention of
// card.ts (issue #4366); card.ts re-exports this file.

import type { BaseGameResponse, Card } from '../common';

/** One seat at the four-handed table. */
export interface QuodlibetPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  /**
   * The cards this seat's hand shows.
   *
   * **Not necessarily your own.** The third wheel makes visibility the rule:
   * under Open Trousers your hand comes back empty while the others are
   * populated, and under Good Hunting every hand is populated.
   */
  cards: Card[];
  trickCount: number;
  /** Running penalty total across the twelve deals. **Lower is better.** */
  penalty: number;
  /** What the last completed deal cost this seat. */
  dealPoints: number;
  /** Finishing position in a shedding contract (1-4, 0 = still holding cards). */
  outRank: number;
  isDealer: boolean;
}

/** One deal's penalty breakdown. */
export interface QuodlibetDeal {
  contract: number;
  /** Stable identifier for the contract, for i18n. */
  contractName: string;
  /** The wheel this contract belongs to (0-2). */
  round: number;
  dealerIdx: number;
  /** Penalty points per seat, indexed by seat. */
  points: number[];
}

/** Quodlibet game configuration. */
export interface QuodlibetConfig {
  cpuDifficulty: number;
  /** Let the game pick each deal's contract instead of asking. */
  autoSelectContract: boolean;
}

/** A suggested hint, computed by the backend. */
export interface QuodlibetHint {
  cardIndices: number[];
  /** i18n reason identifier. */
  reason: string;
}

/**
 * Full Quodlibet game state returned from the API.
 *
 * An Austrian compendium game: 32 cards, four players, eight cards each and no
 * trump. A game is twelve deals arranged as three wheels of four contracts, and
 * **every score is a penalty — the lowest total wins.**
 */
export interface QuodlibetResponse extends BaseGameResponse {
  players: QuodlibetPlayer[];
  /** "selectContract" | "play" | "dealEnd" | "gameEnd". */
  phase: string;
  /** 0-indexed deal number. */
  dealNumber: number;
  totalDeals: number;
  /** The wheel currently in play (1-3). */
  roundNumber: number;
  dealerIdx: number;
  /** Contract being played, or -1 while the dealer is still choosing. */
  currentContract: number;
  /** Stable identifier for the current contract, for i18n. */
  currentContractName: string;
  /** Contracts still unplayed **in this wheel** — the only legal choices. */
  availableContracts: number[];
  /** Stable identifiers in the same order as availableContracts. */
  availableContractNames: string[];
  /** True for Quadrature and Snack, which shed rather than take tricks. */
  isShedding: boolean;
  trickNumber: number;
  trickCount: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  currentTrick: { playerIdx: number; card: Card }[];
  lastTrick: { playerIdx: number; card: Card }[];
  lastTrickWinner: number;
  /** Indices in the human's hand that are legal to play. */
  playableIndices: number[];
  /** True only when a shedding contract leaves nothing playable. */
  canPass: boolean;
  /** Snack's layout: per suit, the rank positions already placed (0 = seven). */
  tablePlaced: number[][];
  /** Quadrature's covered pile, oldest first. */
  stack: Card[];
  lastDeal?: QuodlibetDeal | null;
  dealHistory: QuodlibetDeal[];
  /** Seats currently on the fewest penalty points. */
  winners: number[];
  gameEndFlag: boolean;
  isHumanTurn: boolean;
  isContractPhase: boolean;
  hint?: QuodlibetHint | null;
  /** Contract the backend suggests while choosing, or -1. */
  hintContract: number;
  config: QuodlibetConfig;
}
