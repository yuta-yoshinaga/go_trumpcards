// Type declarations for fortyfives. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Auction Forty-Fives game phase (0=Bid 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type FortyFivesPhaseValue = 0 | 1 | 2 | 3 | 4;

/** An Auction Forty-Fives player's public/own state. Cards are non-empty only for the human. */
export interface FortyFivesPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of this player's TEAM (seats 0&2 = team 0, 1&3 = team 1). */
  teamScore: number;
  /** Whether this player is the round's declarer (the highest bidder). */
  isDeclarer: boolean;
}

/** A card played into the current Auction Forty-Fives trick. */
export interface FortyFivesTrickCard {
  playerIdx: number;
  card: Card;
}

/** Auction Forty-Fives game configuration. */
export interface FortyFivesConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Auction Forty-Fives, computed by the backend. */
export interface FortyFivesHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Server response for the Auction Forty-Fives game (4 players, 2 teams). */
export interface FortyFivesResponse extends BaseGameResponse {
  players: FortyFivesPlayer[];
  phase: FortyFivesPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the round's declarer (highest bidder), or -1 before bidding resolves. */
  declarerIdx: number;
  /** Winning contract value (0=Pass 15 20 25). */
  contract: number;
  /** Trump suit (0=none during bid, else 1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** Each player's bid this round — [p0, p1, p2, p3]. */
  bids: number[];
  currentTrick: FortyFivesTrickCard[];
  /** Cumulative match scores per team — [teamA, teamB]. */
  teamScores: number[];
  /** Points scored by each team this round — [teamA, teamB]. */
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
  hint?: FortyFivesHint | null;
  config: FortyFivesConfig;
}

// --- Twenty-Nine (29) ---
