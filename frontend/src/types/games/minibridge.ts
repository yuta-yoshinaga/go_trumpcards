// Type declarations for minibridge. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Minibridge trick. */
export interface MinibridgeTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a Minibridge table. */
export interface MinibridgePlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. Exposed for every seat; the cards themselves are not. */
  cardCount: number;
  /** Populated for you and — once the contract is set — for the dummy. */
  cards: Card[];
  /**
   * High-card points this seat announced: A=4, K=3, Q=2, J=1. **With no
   * auction this is the game's only public information**, and the four seats
   * always add up to exactly 40.
   */
  hcp: number;
  /** `0` or `1`. Seats 0 and 2 are one pair, 1 and 3 the other. */
  team: number;
  trickCount: number;
}

/**
 * A suggestion. While choosing the contract it names one in `level`/`suit`
 * and carries no `cardIndex`; while playing it names a card.
 */
export interface MinibridgeHint {
  cardIndex?: number;
  /**
   * `minibridgeContract` before play; `minibridgeWinTrick` while playing your
   * own hand, `minibridgeDummy` when the card named is in the dummy's.
   */
  reason: string;
  /** Level to choose. `0` outside the contract phase. */
  level: number;
  /** Denomination to choose — `0` is no-trump. */
  suit: number;
}

/** Deal-count setting. */
export interface MinibridgeConfig {
  /** Deals to play (4..20, default 4 — one turn each as dealer). */
  rounds: number;
}

/** Full Minibridge game state returned from the API. */
export interface MinibridgeResponse extends BaseGameResponse {
  players: MinibridgePlayer[];
  /** `0` = Contract, `1` = Play, `2` = RoundEnd, `3` = GameEnd. */
  phase: number;
  roundNumber: number;
  trickNumber: number;
  /** `0` until the declarer has chosen. */
  contractLevel: number;
  /** `0` means no-trump, which is a choice rather than "unset". */
  contractSuit: number;
  /** Tricks the contract needs: `6 + contractLevel`. `0` before it is chosen. */
  requiredTricks: number;
  /** Decided from the announced HCP as soon as the cards are dealt. */
  declarerIdx: number;
  /** The declarer's partner. Their hand goes face up once the contract is set. */
  dummyIdx: number;
  /** The dummy's hand. Empty until the contract is chosen. */
  dummyHand: Card[];
  /** Whether the last deal's contract was made. */
  lastMade: boolean;
  /** Tricks the declaring side took in the last deal. */
  lastTricks: number;
  /** Running totals, indexed by team. */
  teamScores: number[];
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  currentTrick: MinibridgeTrickCard[];
  /**
   * Hand indices you may legally play, **for the seat you are controlling** —
   * the dummy's when it is the dummy's turn and you are the declarer.
   */
  validPlays: number[];
  gameEndFlag: boolean;
  /** `-1` until decided, and `-1` on a tie. */
  winnerTeam: number;
  hint?: MinibridgeHint;
  config: MinibridgeConfig;
}
