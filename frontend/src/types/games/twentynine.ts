// Type declarations for twentynine. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Twenty-Nine (29) game phase (0=Bid 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type TwentyNinePhaseValue = 0 | 1 | 2 | 3 | 4;

/** A Twenty-Nine (29) player's public/own state. Cards are non-empty only for the human. */
export interface TwentyNinePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative game-point score of this player's TEAM (seats 0&2 = team 0, 1&3 = team 1). */
  teamScore: number;
  /** Whether this player is the round's declarer (the winning bidder). */
  isDeclarer: boolean;
}

/** A card played into the current Twenty-Nine (29) trick. */
export interface TwentyNineTrickCard {
  playerIdx: number;
  card: Card;
}

/** Twenty-Nine (29) game configuration. */
export interface TwentyNineConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Twenty-Nine (29), computed by the backend. */
export interface TwentyNineHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Server response for the Twenty-Nine (29) game (4 players, 2 teams, hidden trump). */
export interface TwentyNineResponse extends BaseGameResponse {
  players: TwentyNinePlayer[];
  phase: TwentyNinePhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the round's declarer (winning bidder), or -1 before bidding resolves. */
  declarerIdx: number;
  /** Winning contract value (0=Pass 16 20 24 28). */
  contract: number;
  /** Trump suit (0=none/hidden during bid, else 1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** Whether the hidden trump suit has been revealed yet. */
  trumpRevealed: boolean;
  /** Each player's bid this round — [p0, p1, p2, p3]. */
  bids: number[];
  currentTrick: TwentyNineTrickCard[];
  /** Cumulative game-point scores per team — [teamA, teamB]. */
  teamScores: number[];
  /** Card points captured by each team this round — [teamA, teamB]. */
  roundTeamPoints: number[];
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning team index (0 or 1), or -1 until the game ends. */
  winnerTeam: number;
  /** Whether it is currently the human's turn to play a card. */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to bid. */
  isHumanBidTurn: boolean;
  hint?: TwentyNineHint | null;
  config: TwentyNineConfig;
}

// --- Court Piece / Rang ---
