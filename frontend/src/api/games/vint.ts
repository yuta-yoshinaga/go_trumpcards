// API client for vint. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { VintResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Vint game settings. */
export interface VintConfigInput {
  cpuDifficulty?: number;
}

/** Commands the /vint/exec endpoint accepts. */
export type VintCommand = 'reset' | 'bid' | 'pass' | 'play' | 'next' | 'log';

/** Options carried alongside a Vint command. */
export interface VintExecOptions {
  /** Bid level 1-7, contracting for 6 + level tricks. Required for `bid`. */
  level?: number;
  /** 0 = Spade, 1 = Club, 2 = Diamond, 3 = Heart, 4 = NoTrump. Required for `bid`. */
  denom?: number;
  /** Hand index. Required for `play`. */
  cardIndex?: number;
  config?: VintConfigInput;
}

/**
 * API client for the Vint /vint/exec endpoint.
 *
 * Vint is the Russian ancestor of contract bridge, played **without a dummy**.
 * Bids name a level and a denomination ranked spades < clubs < diamonds <
 * hearts < no trump — spades are the LOWEST, the reverse of bridge. Both sides
 * then score below the line for every trick they take, whether or not the
 * contract was made. Following suit is compulsory, so the server sends
 * `validPlays`. Only `reset` takes a `config`.
 */
export const vintApi = {
  exec: (command: VintCommand, opts?: VintExecOptions) =>
    gameExec<VintResponse>('vint', {
      command,
      level: opts?.level,
      denom: opts?.denom,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
