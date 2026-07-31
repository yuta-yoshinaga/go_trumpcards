// API client for boston. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BostonResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Boston game settings. */
export interface BostonConfigInput {
  cpuDifficulty?: number;
  /** Hands played before the game is decided (1-30, default 8). */
  targetHands?: number;
}

/** Commands the /boston/exec endpoint accepts. */
export type BostonCommand = 'reset' | 'bid' | 'pass' | 'callpartner' | 'play' | 'next' | 'log';

/** Options carried alongside a Boston command. */
export interface BostonExecOptions {
  /** Ladder step, 1-15. Required for `bid`. **Not a number of tricks.** */
  level?: number;
  /** Trump suit for a trick bid: 1=Spade, 2=Clover, 3=Heart, 4=Diamond. */
  suit?: number;
  /** Seat to partner, or -1 to play alone. Required for `callpartner`. */
  partner?: number;
  /** Hand index. Required for `play`. */
  cardIndex?: number;
  config?: BostonConfigInput;
}

/**
 * API client for the Boston /boston/exec endpoint.
 *
 * Boston's auction is a fifteen-step ladder in which the misère bids
 * **interleave** with the trick bids, so `bid` takes a ladder `level` rather
 * than a trick count; the server sends the whole ladder as `bidOptions`. Trick
 * bids stop at `callpartner`, where -1 means playing alone against three.
 * Following suit is compulsory, so the server sends `validPlays`. Only `reset`
 * takes a `config`.
 */
export const bostonApi = {
  exec: (command: BostonCommand, opts?: BostonExecOptions) =>
    gameExec<BostonResponse>('boston', {
      command,
      level: opts?.level,
      suit: opts?.suit,
      partner: opts?.partner,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
