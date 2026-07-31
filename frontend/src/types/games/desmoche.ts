// Type declarations for desmoche. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One Desmoche seat. */
export interface DesmochePlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. Always sent, including while {@link DesmochePlayer.hidden} is true. */
  cardCount: number;
  /** Empty while {@link DesmochePlayer.hidden} is true. */
  cards: Card[];
  /** Chip balance. Antes subtract, winning the pot adds. */
  score: number;
  /**
   * How many cards this seat has face up. Public for every seat — melds are laid
   * down face up, so this is the only progress indicator toward the ten needed
   * to go out.
   */
  meldedCount: number;
  hidden: boolean;
}

/** One meld on the table. */
export interface DesmocheMeld {
  owner: number;
  /** 0 = set (one rank), 1 = run (consecutive, one suit). */
  kind: number;
  cards: Card[];
}

/** Suggested move for the human seat. */
export interface DesmocheHintPayload {
  cardIndices?: number[];
  cardIndex?: number;
  drawStock: boolean;
  /** Reason identifier, e.g. `desmoche.hint.meld`. */
  reason: string;
}

/** Full Desmoche game state returned from the API. */
export interface DesmocheResponse extends BaseGameResponse {
  players: DesmochePlayer[];
  /** 0 = Draw, 1 = Act, 2 = RoundEnd, 3 = GameEnd. */
  phase: number;
  currentPlayerIdx: number;
  stockCount: number;
  discardTop?: Card;
  melds: DesmocheMeld[];
  roundNo: number;
  /**
   * Chips in the middle. A round that ends with no winner leaves this standing,
   * so it grows across deals — never recompute it from the ante.
   */
  pot: number;
  /** Cards that must be melded to take the pot (10, not the nine dealt). */
  goOutSize: number;
  /** Seat that got down to ten last round; -1 when nobody did. */
  roundWinner: number;
  /**
   * The stock ran out with nobody down to ten, which is why the pot carried
   * over.
   */
  roundExhausted: boolean;
  gameEndFlag: boolean;
  /** Seat with the most chips after the last round; -1 while the game is live. */
  winnerIdx: number;
  hint?: DesmocheHintPayload;
}
