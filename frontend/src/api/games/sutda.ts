// API client for sutda. Follows the split-out convention of gameApi.ts
// (issue #4434); gameApi.ts re-exports this file.

import type { SutdaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Sutda. */
export interface SutdaConfigInput {
  cpuDifficulty?: number;
  /** Seats at the table, 2-5. */
  seats?: number;
  startChips?: number;
}

/** Commands accepted by the /sutda/exec endpoint. */
export type SutdaCommand = 'reset' | 'call' | 'raise' | 'fold' | 'nexthand' | 'hint' | 'log';

/**
 * API client for the Sutda /sutda/exec endpoint.
 *
 * `call`, `raise` and `fold` are commands in their own right — the server also
 * accepts them wrapped in `action`, but sending them directly keeps the client
 * from adding a pointless branch.
 */
export const sutdaApi = {
  exec: (command: SutdaCommand, opts?: { config?: SutdaConfigInput }) =>
    gameExec<SutdaResponse>('sutda', { command, config: opts?.config }),
};
