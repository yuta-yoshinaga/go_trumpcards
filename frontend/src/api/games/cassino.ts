// API client for cassino. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CassinoResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Cassino game settings. */
export interface CassinoConfigInput {
  targetScore?: number;
  multiBuildEnabled?: boolean;
  sweepBonusEnabled?: boolean;
  cpuDifficulty?: number;
}

/** Command verbs accepted by the Cassino /cassino/exec endpoint. */
export type CassinoCommand = 'reset' | 'take' | 'build' | 'trail' | 'next' | 'log' | 'hint';

/** Extra payload fields for the Cassino /cassino/exec endpoint. */
export interface CassinoExecParams {
  handIndex?: number;
  tableIndices?: number[];
  buildIndices?: number[];
  declaredValue?: number;
  config?: CassinoConfigInput;
}

/** API client for the Cassino /cassino/exec endpoint. */
export const cassinoApi = {
  exec: (command: CassinoCommand, params?: CassinoExecParams) =>
    gameExec<CassinoResponse>('cassino', { command, ...(params ?? {}) }),
};
