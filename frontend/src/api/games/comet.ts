// API client for comet. Follows the split-out convention of gameApi.ts
// (issue #4434); gameApi.ts re-exports this file.

import type { CometResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Comet. */
export interface CometConfigInput {
  cpuDifficulty?: number;
  /** Seats at the table, 2-5. */
  players?: number;
  /** Points needed to win, 20-200. */
  targetScore?: number;
}

/** Commands accepted by the /comet/exec endpoint. */
export type CometCommand = 'reset' | 'play' | 'pass' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Comet /comet/exec endpoint.
 *
 * **Which cards are playable comes from the server.** The Comet (9 of
 * diamonds) stands in for any rank, so re-deriving the legal plays on the
 * client would drift from the domain.
 */
export const cometApi = {
  exec: (command: CometCommand, opts?: { handIndex?: number; config?: CometConfigInput }) =>
    gameExec<CometResponse>('comet', {
      command,
      handIndex: opts?.handIndex,
      config: opts?.config,
    }),
};
