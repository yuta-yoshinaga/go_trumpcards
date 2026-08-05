// Type declarations for popejoan. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One of the eight compartments on the board. */
export interface PopeJoanCompartment {
  /** `ace` | `king` | `queen` | `jack` | `game` | `pope` | `matrimony` | `intrigue`. */
  name: string;
  /**
   * Chips sitting on this compartment. Whatever nobody claims stays put and the
   * dealer dresses the board again next deal, so Pope and Matrimony build up.
   */
  chips: number;
}

/** What the board paid out this deal. */
export interface PopeJoanAward {
  compartment: string;
  player: number;
  chips: number;
  /** The dealer took it outright because the turn-up matched. */
  byTurnUp: boolean;
}

/** One Pope Joan seat. */
export interface PopeJoanPlayer {
  id: number;
  isHuman: boolean;
  /**
   * Hand size. Public for every seat: going out is paid one chip per card left
   * in each other hand.
   */
  cardCount: number;
  /** Empty while {@link PopeJoanPlayer.hidden} is true. */
  cards: Card[];
  chips: number;
  /**
   * Public even while the hand is hidden — the Pope's holder is excused the
   * per-card payment, so hiding it would make the settlement unreadable.
   */
  holdsPope: boolean;
  hidden: boolean;
}

/** Suggested move for the human seat. */
export interface PopeJoanHintPayload {
  cardIndex?: number;
  /** Reason identifier, e.g. `popejoan.hint.lead`. */
  reason: string;
}

/** Full Pope Joan game state returned from the API. */
export interface PopeJoanResponse extends BaseGameResponse {
  players: PopeJoanPlayer[];
  /** 0 = Play, 1 = DealEnd, 2 = GameEnd. */
  phase: number;
  /**
   * Hand indices the human may legally play, filled only on their turn.
   * runSuit に従う義務があり、**自由リードでも自分の最も低い札に限られる**ので、出す前に示さないと押して初めて弾かれる (#4934)。
   */
  validPlays: number[];
  currentPlayerIdx: number;
  /** All eight, always. */
  compartments: PopeJoanCompartment[];
  /** Set by the dead hand's last card. Compartments pay only on this suit. */
  trumpSuit: number;
  turnUp?: Card;
  awards: PopeJoanAward[];
  playedPile: Card[];
  /**
   * Suit of the run in progress. **-1 means the run is stopped** and the next
   * player leads their lowest card of any suit — never collapse it to 0, which
   * reads as spades.
   */
  runSuit: number;
  /** Highest rank played in the current run (ace counts 14). */
  runRank: number;
  dealNo: number;
  targetDeals: number;
  /** Seat that went out; -1 while the deal is live. */
  dealWinner: number;
  gameEndFlag: boolean;
  /** Seat with the most chips; -1 while the game is live. */
  winnerIdx: number;
  hint?: PopeJoanHintPayload;
}
