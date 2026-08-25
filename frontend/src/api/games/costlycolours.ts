// API client for costlycolours. Follows the split-out convention of gameApi.ts
// (issue #4434); gameApi.ts re-exports this file.

import type { CostlyColoursResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Costly Colours. */
export interface CostlyColoursConfigInput {
  cpuDifficulty?: number;
  /** Points needed to win, 31-121 (61 = Cotton, 121 = Parlett). */
  targetScore?: number;
}

/** Commands accepted by the /costlycolours/exec endpoint. */
export type CostlyColoursCommand = 'reset' | 'mog' | 'play' | 'nextdeal' | 'hint' | 'log';

/**
 * API client for the Costly Colours /costlycolours/exec endpoint.
 *
 * **The exchange decision is always explicit.** Refusing the mog pegs a point
 * for the opponent, so `accept` is never defaulted.
 */
export const costlycoloursApi = {
  exec: (
    command: CostlyColoursCommand,
    opts?: { handIndex?: number; accept?: boolean; config?: CostlyColoursConfigInput },
  ) =>
    gameExec<CostlyColoursResponse>('costlycolours', {
      command,
      handIndex: opts?.handIndex,
      accept: opts?.accept,
      config: opts?.config,
    }),
};
