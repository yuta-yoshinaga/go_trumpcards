// API client for kaiser. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { KaiserResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Kaiser game settings. */
export interface KaiserConfigInput {
  cpuDifficulty?: number;
  /** Whether no-trump bids are offered (default true). */
  allowNoTrump?: boolean;
}

/** Commands the /kaiser/exec endpoint accepts. */
export type KaiserCommand = 'reset' | 'bid' | 'pass' | 'trump' | 'discard' | 'play' | 'next' | 'log' | 'hint';

/** Options carried alongside a Kaiser command. */
export interface KaiserExecOptions {
  /** Points bid (7-12). Required for `bid`. */
  bid?: number;
  /** 0 = with trump, 1 = no trump, 2 = low no trump. Defaults to 0. */
  contract?: number;
  /** Trump suit for `trump`: 1=Spade, 2=Clover, 3=Heart, 4=Diamond. */
  suit?: number;
  /** The two hand indices to discard. Required for `discard`. */
  indices?: number[];
  /** Hand index. Required for `play`. */
  cardIndex?: number;
  config?: KaiserConfigInput;
}

/**
 * API client for the Kaiser /kaiser/exec endpoint.
 *
 * Kaiser is the Saskatchewan partnership bidding game on a 34-card pack — the
 * usual 32 plus the ♥5 (+5) and the ♠3 (−3), which is why four hands of eight
 * leave a two-card kitty. Bids are in **points**, from 7 to 12. The winning
 * bidder takes the kitty and discards two, but neither scoring card may go.
 * During play the server sends `validPlays`, because following suit is
 * compulsory. Only `reset` takes a `config`.
 */
export const kaiserApi = {
  exec: (command: KaiserCommand, opts?: KaiserExecOptions) =>
    gameExec<KaiserResponse>('kaiser', {
      command,
      bid: opts?.bid,
      contract: opts?.contract,
      suit: opts?.suit,
      indices: opts?.indices,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
