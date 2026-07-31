// API client for bideuchre. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BidEuchreResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Bid Euchre game settings. */
export interface BidEuchreConfigInput {
  cpuDifficulty?: number;
  /** Whether the declarer may name a no-trump form. */
  allowNoTrump?: boolean;
}

/** Commands the /bideuchre/exec endpoint accepts. */
export type BidEuchreCommand = 'reset' | 'bid' | 'pass' | 'trump' | 'play' | 'next' | 'log';

/** Options carried alongside a Bid Euchre command. */
export interface BidEuchreExecOptions {
  /** Bid in tricks, 3-6. Three is the floor. Required for `bid`. */
  value?: number;
  /**
   * 0 = Spade, 1 = Club, 2 = Diamond, 3 = Heart, 4 = NoTrump high,
   * 5 = NoTrump LOW. Required for `trump`.
   */
  trump?: number;
  /** Hand index. Required for `play`. */
  cardIndex?: number;
  config?: BidEuchreConfigInput;
}

/**
 * API client for the Bid Euchre /bideuchre/exec endpoint.
 *
 * Bid Euchre deals the whole 24-card pack to four players, six each, so
 * **there is no kitty**. Bidding starts at three tricks and must beat the
 * standing bid — except for the dealer, who may **equal** it. The declarer then
 * names a trump suit or one of two no-trump forms; at **no trump low** the
 * ranking reverses and the nine is highest. Following suit is compulsory and
 * the left bower counts as a trump, so the server sends `validPlays`. Only
 * `reset` takes a `config`.
 */
export const bidEuchreApi = {
  exec: (command: BidEuchreCommand, opts?: BidEuchreExecOptions) =>
    gameExec<BidEuchreResponse>('bideuchre', {
      command,
      value: opts?.value,
      trump: opts?.trump,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
