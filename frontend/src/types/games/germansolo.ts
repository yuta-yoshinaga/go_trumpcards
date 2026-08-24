// Type declarations for germansolo. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/**
 * German Solo phase values (sync: internal/domain/GermanSolo.go).
 *
 * 0=Bid, 1=AceCall, 2=Play, 3=TrickEnd, 4=RoundEnd, 5=GameEnd. The AceCall
 * phase is entered only by the two partner contracts (Frage / Mussfrage);
 * Solo and Tout go straight from the auction to play.
 */
export type GermanSoloPhaseValue = 0 | 1 | 2 | 3 | 4 | 5;

/**
 * German Solo bid values (sync: internal/domain/GermanSolo.go).
 *
 * 0=pass, 1=Mussfrage, 2=Frage, 3=Solo, 4=Tout. **Mussfrage is never bid**: it
 * is forced on the holder of Spadille (♣Q) when every seat passes, so it can
 * appear as `winningBid` but never in `biddableBids`.
 */
export type GermanSoloBidValue = 0 | 1 | 2 | 3 | 4;

/** A German Solo player's public/own state. Cards are non-empty only for the human during play. */
export interface GermanSoloPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player took the contract this deal. */
  isDeclarer: boolean;
}

/** A card played into the current German Solo trick. */
export interface GermanSoloTrickCard {
  playerIdx: number;
  card: Card;
}

/** German Solo game configuration. */
export interface GermanSoloConfig {
  cpuDifficulty: number;
  /** Number of deals that make up the match; the highest cumulative score wins. */
  targetRounds: number;
}

/** A suggested hint for German Solo, computed by the backend. */
export interface GermanSoloHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full German Solo game state returned from the API.
 *
 * German Solo is the German branch of the Ombre → Quadrille family: four seats
 * on a 32-card Skat pack (A, K, Q, J, 10, 9, 8, 7), eight cards each and eight
 * tricks. Three matadors outrank every other trump whatever the trump suit is —
 * Spadille (♣Q) > Manille (the 7 of trumps) > Basta (♠Q) — so the two black
 * queens are always trumps and the trump suit itself runs A > K > (Q) > J > 10
 * > 9 > 8, the queen surviving only when trumps are red.
 */
export interface GermanSoloResponse extends BaseGameResponse {
  players: GermanSoloPlayer[];
  phase: GermanSoloPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  /** Seat index of the player whose turn it is to bid. */
  currentBidderIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the forehand — bids first and leads the first trick. */
  forehandIdx: number;
  /** Seat index of the declarer (bid winner), or -1 until the auction settles. */
  declarerIdx: number;
  /** The settled contract (0=none, 1=Mussfrage, 2=Frage, 3=Solo, 4=Tout). */
  winningBid: GermanSoloBidValue;
  /** Highest bid standing in the auction — what a new bid must exceed. */
  highestBid: GermanSoloBidValue;
  /** Contracts this seat may still declare. Empty once the auction has closed. */
  biddableBids: number[];
  /** Tricks the declaring side needs (8 for Tout, otherwise 5). */
  requiredTricks: number;
  /** Tricks taken so far by the declaring side (declarer + revealed partner). */
  declarerTricks: number;
  /** Tricks taken so far by the defenders. */
  defenderTricks: number;
  /** The trump suit (1=♠ 2=♣ 3=♥ 4=♦), or -1 until chosen. */
  trumpSuit: number;
  currentTrick: GermanSoloTrickCard[];
  /** Cumulative match scores per player — [p0, p1, p2, p3]. */
  playerScores: number[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Deal outcome (0=None, 1=Made, 2=Failed). */
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
  /** エース呼びフェーズで人間 (落札者) の指名待ちか。 */
  isHumanAceCallTurn: boolean;
  /** 呼ばれたエースのスート (-1=未指名)。**呼び声は公開情報**。 */
  calledAceSuit: number;
  /** 落札者が呼べるエースのスート (画面の選択肢)。持っている札と切り札スートは除かれる。 */
  callableAceSuits: number[];
  /** 味方の席。**呼ばれたエースが場に出るまで -1** (誰が味方かは伏せる)。 */
  partnerIdx: number;
  /** 単独契約 (Solo / Tout)、または呼べるエースが無かった Frage。 */
  playsAlone: boolean;
  hint?: GermanSoloHint | null;
  /**
   * エース呼びフェーズでヒントが勧めるスート (0=なし)。
   *
   * 共有の hint オブジェクトは札の索引しか運べないので、スートはここに載る。
   */
  hintAceSuit: number;
  config: GermanSoloConfig;
}
