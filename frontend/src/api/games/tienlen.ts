// API client for tienlen. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TienLenConfigInput, TienLenResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Tien Len /tienlen/exec endpoint. */
export const tienlenApi = {
  exec: (command: 'reset' | 'play', indices?: number[], config?: TienLenConfigInput) =>
    gameExec<TienLenResponse>('tienlen', { command, indices, config }),
};
