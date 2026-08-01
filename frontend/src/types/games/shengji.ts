// Type declarations for shengji. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One Sheng Ji seat. */
export interface ShengJiPlayer {
  id: number;
  isHuman: boolean;
  /** 0 for seats 0/2, 1 for seats 1/3 — **partners sit opposite**. */
  team: number;
  cardCount: number;
  /** Your own hand only, until the hand is settled. */
  cards: Card[];
  /**
   * True for the side whose level is being played. **The declarers do not
   * collect points** — they win by holding the defenders under the target.
   */
  isDeclarer: boolean;
  isCurrentTurn: boolean;
}

/** The shape that was led into the current trick. */
export interface ShengJiCombo {
  /** 0 = none, 1 = single, 2 = pair, 3 = tractor (consecutive pairs). */
  kind: number;
  rank: number;
  size: number;
  /** True when the play belongs to the trump group. */
  trump: boolean;
  suit: number;
}

/** One seat's cards in the current trick. */
export interface ShengJiPlay {
  seat: number;
  cards: Card[];
}

/** The trump declaration that stands for this hand. */
export interface ShengJiDeclaration {
  seat: number;
  suit: number;
  /** 1 for a single level card, 2 for a pair. **Only a stronger showing overrides.** */
  strength: number;
}

/** How the previous hand settled. */
export interface ShengJiHandResult {
  declarerTeam: number;
  /** Points the **defenders** collected — they are the side that scores. */
  defenderPoints: number;
  /** Kitty points that reached the defenders, already multiplied. */
  kittyPoints: number;
  /** The multiplier applied to the kitty; 0 when the declarers took the last trick. */
  kittyMultiplier: number;
  /** True when the declarers kept the defenders under the target. */
  declarerHeld: boolean;
  advance: number;
  /** The team that climbed; -1 when nobody did. */
  advancingTeam: number;
}

/** Full Sheng Ji game state returned from the API. */
export interface ShengJiResponse extends BaseGameResponse {
  players: ShengJiPlayer[];
  /** 0 = Declare, 1 = Kitty, 2 = Play, 3 = HandEnd, 4 = GameEnd. */
  phase: number;
  handNumber: number;
  currentPlayerIdx: number;
  /**
   * This hand's level rank. **Every card of this rank, in all four suits, is a
   * trump** — not just the ones in the trump suit.
   */
  level: number;
  teamLevels: [number, number];
  declarerTeam: number;
  /** 1-4, or 0 for no trump suit (level cards and jokers are still trumps). */
  trumpSuit: number;
  declaration: ShengJiDeclaration | null;
  /** Suits the human can declare right now, keyed by suit number, valued by strength. */
  declarableSuits: Record<string, number>;
  kittySize: number;
  /** The kitty's cards — empty until the hand is settled. */
  kitty: Card[];
  trick: ShengJiPlay[];
  trickLeader: number;
  leadCombo: ShengJiCombo | null;
  /** Points each team has collected this hand. **Only the defenders' total counts.** */
  teamPoints: [number, number];
  trickCount: number;
  lastTrickWinner: number;
  lastResult: ShengJiHandResult | null;
  minLevel: number;
  maxLevel: number;
  kittySizeMax: number;
  /** Points in the pack: 200. */
  totalPoints: number;
  /** What the defenders need to take the deal: 80, two fifths of the pack. */
  defenderTarget: number;
  /** Every this many points above the target is one more level: 40. */
  advanceStep: number;
  gameEndFlag: boolean;
  /** Winning team; -1 while the game is live. */
  winnerTeam: number;
  config: ShengJiConfigOutput;
}

/** Settings echoed back with the game state. */
export interface ShengJiConfigOutput {
  cpuDifficulty: number;
}
