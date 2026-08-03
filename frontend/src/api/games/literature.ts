// API client for literature. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { LiteratureResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Literature game settings. */
export interface LiteratureConfigInput {
  cpuDifficulty?: number;
}

/** Commands the /literature/exec endpoint accepts. */
export type LiteratureCommand = 'reset' | 'ask' | 'claim' | 'log';

/** Options carried alongside a Literature command. */
export interface LiteratureExecOptions {
  /** Seat to ask. **Must be an opponent.** Required for `ask`. */
  target?: number;
  /** 1 = Spade, 2 = Clover, 3 = Heart, 4 = Diamond. Required for `ask`. */
  suit?: number;
  /** Rank. Required for `ask`. */
  value?: number;
  /** Half-suit index. Required for `claim`. */
  halfSuit?: number;
  /** Where each of the six cards is, in `halfSuitCards` order. Required for `claim`. */
  holders?: number[];
  config?: LiteratureConfigInput;
}

/**
 * API client for the Literature /literature/exec endpoint.
 *
 * Literature deals 48 cards (a standard pack with the eights removed) to six
 * players in two teams of three, seated alternately. You may only ask an
 * **opponent**, only for a half-suit you already hold, and only for a card you
 * do **not** hold. Claiming has **three** outcomes, not two: naming the wrong
 * teammate **cancels** the half-suit rather than handing it to the opponents,
 * so the totals need not add to eight. Winning takes **five** half-suits — a
 * majority of eight. Only `reset` takes a `config`.
 */
export const literatureApi = {
  exec: (command: LiteratureCommand, opts?: LiteratureExecOptions) =>
    gameExec<LiteratureResponse>('literature', {
      command,
      target: opts?.target,
      suit: opts?.suit,
      value: opts?.value,
      halfSuit: opts?.halfSuit,
      holders: opts?.holders,
      config: opts?.config,
    }),
};
