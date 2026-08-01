// API client for shengji. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ShengJiResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Sheng Ji game settings. */
export interface ShengJiConfigInput {
  cpuDifficulty?: number;
}

/** Commands the /shengji/exec endpoint accepts. */
export type ShengJiCommand = 'reset' | 'declare' | 'bury' | 'play' | 'next' | 'log';

/** Options carried alongside a Sheng Ji command. */
export interface ShengJiExecOptions {
  /**
   * Suit to declare: 1-4, or **0 to pass**. Required for `declare` — zero is a
   * real choice here, so it cannot be omitted to mean the same thing.
   */
  suit?: number;
  /** Hand indexes: the eight to bury, or the cards to play. */
  cardIndexes?: number[];
  config?: ShengJiConfigInput;
}

/**
 * API client for the Sheng Ji (升级 / 拖拉机) /shengji/exec endpoint.
 *
 * Sheng Ji deals **25 cards each from 108** (52x2 + 4 jokers) to four players
 * in two partnerships sitting opposite, leaving an **eight-card kitty**. Every
 * hand is played at a **level**, and the trump group is the trump suit **plus
 * every card of the level rank in all four suits plus all four jokers**. The
 * **defenders** collect the 5s, 10s and kings; the declarers win the hand by
 * holding them under **80** of the pack's 200. If the defenders take the last
 * trick the kitty's points reach them multiplied. Only `reset` takes a `config`.
 */
export const shengjiApi = {
  exec: (command: ShengJiCommand, opts?: ShengJiExecOptions) =>
    gameExec<ShengJiResponse>('shengji', {
      command,
      suit: opts?.suit,
      cardIndexes: opts?.cardIndexes,
      config: opts?.config,
    }),
};
