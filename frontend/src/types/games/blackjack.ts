// Type declarations for blackjack. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single BlackJack hand with score, cards, and status flags. */
export interface BlackJackHand {
  score: number;
  cards: Card[];
  bet: number;
  stood: boolean;
  doubled: boolean;
  busted: boolean;
  isBlackJack: boolean;
  canSplit: boolean;
  surrendered: boolean;
  canSurrender: boolean;
}

/** BlackJack player (dealer or human) with chips and cards. */
export interface BlackJackPlayer {
  score?: number;
  cards?: Card[];
  chips: number;
}

/** BlackJack game phase (1=Bet, 2=Deal, 3=Insurance, 4=Action, 5=End, 6=EarlySurrender). */
export type BlackJackPhase = 1 | 2 | 3 | 4 | 5 | 6;

/** CPU player seat in BlackJack with chips and hands. */
export interface BlackJackCpuSeat {
  chips: number;
  hands: BlackJackHand[];
  insuranceBet: number;
}

/** Result of a BlackJack side bet (Perfect Pairs, 21+3). */
export interface BlackJackSideBetResult {
  betType: number;
  resultType: number;
  resultName: string;
  betAmount: number;
  payout: number;
}

/** Full BlackJack game state returned from the API. */
export interface BlackJackResponse extends BaseGameResponse {
  dealer: BlackJackPlayer;
  player: BlackJackPlayer;
  hands?: BlackJackHand[];
  currentHandIdx: number;
  phase: BlackJackPhase;
  insuranceBet: number;
  insuranceAvailable: boolean;
  hintEnabled: boolean;
  suggestedAction: number;
  deckCount: number;
  dealerHitsSoft17: boolean;
  countingEnabled: boolean;
  cpuPlayerCount: number;
  runningCount: number;
  trueCount: number;
  cpuPlayers?: BlackJackCpuSeat[];
  perfectPairsBet: number;
  twentyOnePlus3Bet: number;
  sideBetResults?: BlackJackSideBetResult[];
  /** i18n keys of variant bonuses achieved this round (Spanish 21, e.g. `spanish21.bonus.777.spade`). */
  bonuses?: string[];
  doubleAfterSplit: boolean;
  countingSystem: number;
  deckPenetration: number;
  multiHandCount: number;
  surrenderRule: number;
}
