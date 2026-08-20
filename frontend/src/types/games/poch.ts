// Type declarations for poch. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One of the nine pools on the board. */
export interface PochPool {
  /** `ace` | `king` | `queen` | `jack` | `ten` | `marriage` | `sequence` | `pocher` | `centre`. */
  name: string;
  /**
   * Chips sitting on this pool. Whatever nobody claims stays put and everyone
   * antes again next deal, so this grows — it is the motivation for the deal to
   * come, not just a running total.
   */
  chips: number;
}

/** What stage one paid out. */
export interface PochStakingAward {
  pool: string;
  player: number;
  chips: number;
}

/** One Poch seat. */
export interface PochPlayer {
  id: number;
  isHuman: boolean;
  /**
   * Hand size. Public for every seat, including while hidden: whoever goes out
   * is paid one chip per card still held, so this is a running liability.
   */
  cardCount: number;
  /** Empty while {@link PochPlayer.hidden} is true. */
  cards: Card[];
  chips: number;
  /** Staked in the current pochen round. */
  bet: number;
  folded: boolean;
  hidden: boolean;
}

/** Suggested move for the human seat. */
export interface PochHintPayload {
  /** `bet` | `fold` | `play`. */
  action: string;
  cardIndex?: number;
  /** Reason identifier, e.g. `poch.hint.play`. */
  reason: string;
}

/** Full Poch game state returned from the API. */
export interface PochResponse extends BaseGameResponse {
  players: PochPlayer[];
  /** 0 = Staking, 1 = Pochen, 2 = Stops, 3 = DealEnd, 4 = GameEnd. */
  phase: number;
  /**
   * Hand indices the human may legally play, filled only on their turn.
   * stopsSuit に従う義務があるので、出す前に示さないと押して初めて弾かれる (#4933)。
   */
  validPlays: number[];
  currentPlayerIdx: number;
  /** All nine, always. */
  pools: PochPool[];
  /** Suit of the turn-up. Stage one pays only on this suit, never on rank alone. */
  paySuit: number;
  turnUp?: Card;
  /**
   * Stage one resolves the instant the cards are dealt, so this record is the
   * only way to see what happened.
   */
  stakingAwards: PochStakingAward[];
  betTarget: number;
  /** Size of the human's own strongest same-rank set (0 = no set). */
  yourBestComboSize: number;
  /** Rank of that set (meaningless when the size is 0). */
  yourBestComboRank: number;
  /** Seat whose set won the comparison; -1 until settled. There is no declaration. */
  pochenWinner: number;
  pochenPot: number;
  playedPile: Card[];
  /**
   * Suit of the run in progress. **-1 means the run is stopped** and any card
   * may be led — never collapse it to 0, which reads as spades.
   */
  stopsSuit: number;
  /** Highest rank played in the current run (ace counts 14). */
  stopsRank: number;
  dealNo: number;
  targetDeals: number;
  /** Seat that went out; -1 while the deal is live. */
  dealWinner: number;
  gameEndFlag: boolean;
  /** Seat with the most chips; -1 while the game is live. */
  winnerIdx: number;
  hint?: PochHintPayload;
}
