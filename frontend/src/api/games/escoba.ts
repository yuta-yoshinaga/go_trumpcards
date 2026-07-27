// API client for escoba. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { EscobaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Escoba game settings. */
export interface EscobaConfigInput {
  targetScore?: number;
  cpuDifficulty?: number;
}

/** Command verbs accepted by the Escoba /escoba/exec endpoint (short forms). */
export type EscobaCommand = 'r' | 'n' | 'p' | 'log';

/** Extra payload fields for the Escoba /escoba/exec endpoint. */
export interface EscobaExecParams {
  handIndex?: number;
  tableIndices?: number[];
  config?: EscobaConfigInput;
}

/** API client for the Escoba /escoba/exec endpoint. */
export const escobaApi = {
  exec: (command: EscobaCommand, params?: EscobaExecParams) =>
    gameExec<EscobaResponse>('escoba', { command, ...(params ?? {}) }),
};
