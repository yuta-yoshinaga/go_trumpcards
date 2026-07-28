// API client for barbu. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BarbuResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Barbu game settings. */
export interface BarbuConfigInput {
  cpuDifficulty?: number;
}

/** Command verbs accepted by the Barbu /barbu/exec endpoint (short forms). */
export type BarbuCommand = 'r' | 'n' | 'c' | 'p' | 'log';

/** Extra payload fields for the Barbu /barbu/exec endpoint. */
export interface BarbuExecParams {
  contract?: number;
  trumpSuit?: number;
  handIndex?: number;
  tableIndices?: number[];
  config?: BarbuConfigInput;
}

/** API client for the Barbu /barbu/exec endpoint. */
export const barbuApi = {
  exec: (command: BarbuCommand, params?: BarbuExecParams) =>
    gameExec<BarbuResponse>('barbu', { command, ...(params ?? {}) }),
};
