import type { HoldemEquity } from './holdem';
// Type declarations for poker. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, BettingHumanProfileData, BettingMetaAI, Card } from '../common';

/** Poker player data including hand, chips, and status. */
export interface PokerPlayerData {
  id: number;
  isHuman: boolean;
  cards: Card[];
  chips: number;
  currentBet: number;
  folded: boolean;
  allIn: boolean;
  handRank: number;
  handName: string;
  exchangeCount: number;
  playStyleName: string;
}

/** CPU betting action in Poker. */
export interface PokerCpuAction {
  playerIdx: number;
  action: number;
  amount: number;
}

/** CPU card exchange result in Poker. */
export interface PokerCpuExchange {
  playerIdx: number;
  exchangeCount: number;
}

/** Poker round result for a single player. */
export interface PokerResult {
  playerIdx: number;
  handRank: number;
  handName: string;
  kickers: string;
  wonAmount: number;
}

/** Side pot in Poker with eligible players. */
export interface PokerSidePot {
  amount: number;
  eligiblePlayers: number[];
}

/** Probability of achieving a specific poker hand rank. */
export interface PokerOdds {
  handRank: number;
  handName: string;
  probability: number;
  count: number;
  total: number;
}

/** Poker game phase (0=Init, 1=Deal, 2=Exchange, 3=SecondBet, 4=End). */
export type PokerPhase = 0 | 1 | 2 | 3 | 4;

/** Full Poker game state returned from the API. */
export interface PokerResponse extends BaseGameResponse {
  players: PokerPlayerData[];
  pot: number;
  /**
   * 2 巡目ベットでの勝率。ベッティングフェーズ以外では undefined。
   *
   * Holdem 系と同じ `HoldemEquity` 形。判定はドメインの `GetEquity` が唯一の
   * 出どころで、フロントは表示だけ担当する。
   */
  equity?: HoldemEquity;
  /** コールに必要な額に対するポットオッズ (0-100)。 */
  potOdds?: number;
  sidePots: PokerSidePot[];
  dealerIdx: number;
  currentTurn: number;
  phase: PokerPhase;
  /**
   * Whether the human's exchange count is currently raising the CPUs' guard.
   *
   * The threshold lives in the domain (`calcExchangeWarning` shares it), so the
   * page must not re-derive it from `exchangeCount` -- a second copy drifts from
   * what the CPUs actually do.
   */
  exchangeRead: boolean;
  gameEndFlag: boolean;
  lastBet: number;
  minRaise: number;
  ante: number;
  jokerCount: number;
  bettingLimit: number;
  raiseCount: number;
  maxBetAmount: number;
  roundResults: PokerResult[];
  cpuActions: PokerCpuAction[];
  cpuExchanges: PokerCpuExchange[];
  odds?: PokerOdds[];
  isLowball: boolean;
  metaAI?: BettingMetaAI;
  profile?: BettingHumanProfileData;
}
