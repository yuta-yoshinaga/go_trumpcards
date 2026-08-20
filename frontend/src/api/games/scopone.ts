// API client for scopone. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ScoponeResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Scopone game settings. */
export interface ScoponeConfigInput {
  targetScore?: number;
  cpuDifficulty?: number;
}

/** Command verbs accepted by the Scopone /scopone/exec endpoint (short forms). */
export type ScoponeCommand = 'r' | 'n' | 'p' | 'log' | 'hint';

/** Extra payload fields for the Scopone /scopone/exec endpoint. */
export interface ScoponeExecParams {
  handIndex?: number;
  tableIndices?: number[];
  config?: ScoponeConfigInput;
}

/** API client for the Scopone /scopone/exec endpoint. */
export const scoponeApi = {
  exec: (command: ScoponeCommand, params?: ScoponeExecParams) =>
    gameExec<ScoponeResponse>('scopone', { command, ...(params ?? {}) }),
};
