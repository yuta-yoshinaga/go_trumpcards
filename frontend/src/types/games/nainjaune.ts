// Type declarations for nainjaune. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One of the five boxes on the board. */
export interface NainJauneBox {
  /** `ten` | `jack` | `queen` | `king` | `dwarf`. */
  name: string;
  /**
   * Chips sitting on this box. Whatever nobody claims stays put and everyone
   * antes again next deal, so the dwarf builds up fastest — it takes five a deal.
   */
  chips: number;
  /**
   * The exact card that claims this box. Sent rather than derived, because the
   * suit is part of it: an off-suit seven does not claim the dwarf.
   */
  card?: Card;
}

/** What a box paid out this deal. */
export interface NainJauneAward {
  box: string;
  player: number;
  chips: number;
}

/** One Le Nain Jaune seat. */
export interface NainJaunePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  /** Empty while {@link NainJaunePlayer.hidden} is true. */
  cards: Card[];
  chips: number;
  /**
   * What the hand is worth if the deal ends now. Public for every seat, because
   * settlement is in points rather than cards — the card count alone does not
   * say what a hand owes.
   */
  points: number;
  hidden: boolean;
}

/** Suggested move for the human seat. */
export interface NainJauneHintPayload {
  cardIndex?: number;
  /** Reason identifier, e.g. `nainjaune.hint.box`. */
  reason: string;
}

/** Full Le Nain Jaune game state returned from the API. */
export interface NainJauneResponse extends BaseGameResponse {
  players: NainJaunePlayer[];
  /** 0 = Play, 1 = DealEnd, 2 = GameEnd. */
  phase: number;
  /**
   * Hand indices the human may legally play, filled only on their turn.
   * runRank の次のランクを続ける義務があるので、出す前に示さないと押して初めて弾かれる (#4935)。
   */
  validPlays: number[];
  currentPlayerIdx: number;
  /** All five, always. */
  boxes: NainJauneBox[];
  /** Cards set aside undealt. Nobody uses them. */
  talonCount: number;
  awards: NainJauneAward[];
  playedPile: Card[];
  /**
   * Highest rank played in the current run. **0 means the run is stopped** and
   * any card may be led. There is no suit here — the run ignores suit entirely,
   * which is the decisive difference from Pope Joan.
   */
  runRank: number;
  dealNo: number;
  targetDeals: number;
  /** Seat that went out; -1 while the deal is live. */
  dealWinner: number;
  gameEndFlag: boolean;
  /** Seat with the most chips; -1 while the game is live. */
  winnerIdx: number;
  hint?: NainJauneHintPayload;
}
