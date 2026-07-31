// Type declarations for loba. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One Loba seat. */
export interface LobaPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. Always sent, including while {@link LobaPlayer.hidden} is true. */
  cardCount: number;
  /** Empty while {@link LobaPlayer.hidden} is true. */
  cards: Card[];
  /**
   * Penalty total. Public for every seat: 101 knocks a player out, so how close
   * each one is drives every decision.
   */
  score: number;
  eliminated: boolean;
  /** Whether this seat has melded this round, which is what allows laying off. */
  hasMelded: boolean;
  hidden: boolean;
}

/** One meld on the table. */
export interface LobaMeld {
  owner: number;
  /** 0 = pierna (one rank, three different suits), 1 = escalera (a suit run). */
  kind: number;
  cards: Card[];
}

/** Suggested move for the human seat. */
export interface LobaHintPayload {
  cardIndices?: number[];
  cardIndex?: number;
  drawStock: boolean;
  /** Reason identifier, e.g. `loba.hint.meld`. */
  reason: string;
}

/** Full Loba game state returned from the API. */
export interface LobaResponse extends BaseGameResponse {
  players: LobaPlayer[];
  /** 0 = Draw, 1 = Act, 2 = RoundEnd, 3 = GameEnd. */
  phase: number;
  currentPlayerIdx: number;
  stockCount: number;
  discardTop?: Card;
  melds: LobaMeld[];
  roundNo: number;
  /** Penalty total that eliminates a player (101). */
  knockOut: number;
  /** Seat that went out last round; -1 when none. */
  roundWinner: number;
  /**
   * That seat went out without having melded earlier in the round, which is
   * what takes ten off.
   */
  roundClean: boolean;
  gameEndFlag: boolean;
  /** Last player standing; -1 while the game is live. */
  winnerIdx: number;
  hint?: LobaHintPayload;
}
