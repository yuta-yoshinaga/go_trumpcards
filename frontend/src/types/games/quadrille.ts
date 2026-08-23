// Type declarations for quadrille. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/**
 * Quadrille phase values (sync: internal/domain/Quadrille.go).
 *
 * 0=Bid, 1=KingCall, 2=Play, 3=TrickEnd, 4=RoundEnd, 5=GameEnd. The KingCall
 * phase does not exist in Ombre, so every later phase is one higher.
 */
export type QuadrillePhaseValue = 0 | 1 | 2 | 3 | 4 | 5;

/** A Quadrille player's public/own state. Cards are non-empty only for the human during play. */
export interface QuadrillePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player is the round's Quadrille (won the bid, plays alone). */
  isQuadrille: boolean;
}

/** A card played into the current Quadrille trick. */
export interface QuadrilleTrickCard {
  playerIdx: number;
  card: Card;
}

/** Quadrille game configuration. */
export interface QuadrilleConfig {
  cpuDifficulty: number;
  /** Number of deals that make up the match; the highest cumulative score wins. */
  targetRounds: number;
}

/** A suggested hint for Quadrille, computed by the backend. */
export interface QuadrilleHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Quadrille game state returned from the API.
 *
 * Quadrille is a 3-player soloist-vs-coalition trick-taker on a 40-card Spanish
 * deck (no 8/9/10). A Bid phase (pass / entrar / solo) plus a chosen trump suit
 * decides the Quadrille, who then plays alone against the coalition of the other
 * two. The trump group ranks Spadille (♠A) > Manille (7 of trump) > Basto (♣A)
 * > Punto (Ace of trump, red only) > K > Q > J > 6..2 of trump.
 */
export interface QuadrilleResponse extends BaseGameResponse {
  players: QuadrillePlayer[];
  phase: QuadrillePhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  /** Seat index of the player whose turn it is to bid. */
  currentBidderIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the forehand (first to bid / lead). */
  forehandIdx: number;
  /** Seat index of the round's Quadrille (bid winner), or -1 until decided. */
  quadrilleIdx: number;
  /** The winning bid (0=pass/none, 1=entrar, 2=solo). */
  winningBid: number;
  /** The trump suit (1=♠ 2=♣ 3=♥ 4=♦), or -1 until chosen. */
  trumpSuit: number;
  currentTrick: QuadrilleTrickCard[];
  /** Cumulative match scores per player — [p0, p1, p2]. */
  playerScores: number[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Deal outcome (0=None, 1=Sacar, 2=Puesta, 3=Codille). */
  outcome: number;
  /** Match result from the human's perspective (-1 lose, 0 none, 1 win). */
  result: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to play a card. */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to bid. */
  isHumanBidTurn: boolean;
  /** 王呼びフェーズで人間 (落札者) の指名待ちか。 */
  isHumanKingCallTurn: boolean;
  /** 呼ばれた王のスート (-1=未指名)。**呼び声は公開情報**。 */
  calledKingSuit: number;
  /** 落札者が呼べる王のスート (画面の選択肢)。 */
  callableKingSuits: number[];
  /** 味方の席。**呼ばれた王が場に出るまで -1** (誰が味方かは伏せる)。 */
  partnerIdx: number;
  /** 落札者が王を 4 枚とも持っていた単独プレイ。 */
  roiSeul: boolean;
  hint?: QuadrilleHint | null;
  config: QuadrilleConfig;
}

// --- Ulti (Ultimo) ---
