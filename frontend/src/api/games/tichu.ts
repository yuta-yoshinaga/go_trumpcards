// API client for tichu. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TichuResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Tichu /tichu/exec endpoint. */
export const tichuApi = {
  exec: (params: { command: string; indices?: number[]; declType?: number; config?: { cpuDifficulty?: number } }) =>
    gameExec<TichuResponse>('tichu', params),
};
