// API client for scopa. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ScopaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Scopa game settings. */
export interface ScopaConfigInput {
  targetScore?: number;
  cpuDifficulty?: number;
}

/** Command verbs accepted by the Scopa /scopa/exec endpoint (short forms). */
export type ScopaCommand = 'r' | 'n' | 'p' | 'log' | 'hint';

/** Extra payload fields for the Scopa /scopa/exec endpoint. */
export interface ScopaExecParams {
  handIndex?: number;
  tableIndices?: number[];
  config?: ScopaConfigInput;
}

/** API client for the Scopa /scopa/exec endpoint. */
export const scopaApi = {
  exec: (command: ScopaCommand, params?: ScopaExecParams) =>
    gameExec<ScopaResponse>('scopa', { command, ...(params ?? {}) }),
};
