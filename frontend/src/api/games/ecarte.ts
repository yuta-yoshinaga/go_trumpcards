// API client for ecarte. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { EcarteResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Écarté game settings. */
export interface EcarteConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** Commands accepted by the Écarté /ecarte/exec endpoint. */
export type EcarteCommand =
  | 'reset'
  | 'propose'
  | 'stand'
  | 'respond'
  | 'discard'
  | 'play'
  | 'next'
  | 'hint'
  | 'log'
  | 'config';

/**
 * API client for the Écarté /ecarte/exec endpoint.
 *
 * Écarté is a 2-player French 32-card trick game with an Exchange phase. The
 * elder (non-dealer) chooses Propose or Stand; if proposed, the dealer Accepts
 * or Refuses; on accept, each player discards any number of cards and draws
 * replacements, then the elder decides again (until the stock empties). Play is
 * 5 strict must-follow tricks (rank K>Q>J>A>10>9>8>7). Scores accumulate to a
 * target (default 5).
 *   - `respond` → `{ accept: boolean }`
 *   - `discard` → `{ discardIndices: number[] }`
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` / `config` → `{ config }`
 *   - `propose` / `stand` / `next` / `hint` / `log` carry no extra fields.
 */
export const ecarteApi = {
  exec: (
    command: EcarteCommand,
    opts?: {
      accept?: boolean;
      cardIndex?: number;
      discardIndices?: number[];
      config?: EcarteConfigInput;
    },
  ) =>
    gameExec<EcarteResponse>('ecarte', {
      command,
      accept: opts?.accept,
      cardIndex: opts?.cardIndex,
      discardIndices: opts?.discardIndices,
      config: opts?.config,
    }),
};
