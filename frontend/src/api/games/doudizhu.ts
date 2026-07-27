// API client for doudizhu. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { DoudizhuResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Dou Dizhu /doudizhu/exec endpoint. */
export const doudizhuApi = {
  exec: (params: { command: string; indices?: number[]; bidValue?: number; config?: { cpuDifficulty?: number } }) =>
    gameExec<DoudizhuResponse>('doudizhu', params),
};
