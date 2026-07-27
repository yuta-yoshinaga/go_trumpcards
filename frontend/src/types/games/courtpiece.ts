// Type declarations for courtpiece. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Court Piece (Rang) game phase (0=TrumpDeclaration 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type CourtPiecePhaseValue = 0 | 1 | 2 | 3 | 4;

/** A Court Piece (Rang) player's public/own state. Cards are non-empty only for the human. */
export interface CourtPiecePlayer {
  id: number;
  isHuman: boolean;
  /** Team index (seats 0&2 = team 0, 1&3 = team 1). */
  team: number;
  cardCount: number;
  cards: Card[];
  /** Round points (tricks won this round). */
  roundScore: number;
  /** Cumulative game-point (Sar) score of this player's team. */
  cumulativeScore: number;
  trickCount: number;
}

/** A card played into the current Court Piece (Rang) trick. */
export interface CourtPieceTrickCard {
  playerIdx: number;
  card: Card;
}

/** Court Piece (Rang) game configuration. */
export interface CourtPieceConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A suggested hint for Court Piece (Rang), computed by the backend. */
export interface CourtPieceHint {
  /** Card index to play (Play phase). */
  cardIndex?: number;
  /** Trump suit to declare (TrumpDeclaration phase, 1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit?: number;
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Server response for the Court Piece (Rang) game (4 players, 2 teams, called trump). */
export interface CourtPieceResponse extends BaseGameResponse {
  players: CourtPiecePlayer[];
  phase: CourtPiecePhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  /** Seat index of the caller (Hakim) who declares the trump suit. */
  callerIdx: number;
  /** Trump suit (0=undeclared during TrumpDeclaration, else 1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  currentTrick: CourtPieceTrickCard[];
  /** Cumulative game-point (Sar) scores per team — [teamA, teamB]. */
  teamScores: number[];
  /** Consecutive round wins by the {@link lastWinnerTeam} (drives the Court bonus). */
  consecutiveWins: number;
  /** Team that won the previous round (or -1 before any round resolves). */
  lastWinnerTeam: number;
  /** Whether the previous round was a Court (sweep / consecutive bonus). */
  lastRoundCourt: boolean;
  gameEndFlag: boolean;
  /** Winning team index (0 or 1), or -1 until the game ends. */
  winnerTeam: number;
  /** Seat index of the player who led the current trick. */
  leadPlayerIdx: number;
  hint?: CourtPieceHint | null;
  config: CourtPieceConfig;
}

// --- Bezique ---
