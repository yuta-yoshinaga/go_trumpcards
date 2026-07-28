// API client for zheng. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ZhengConfigInput, ZhengResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Zheng Shangyou /zheng/exec endpoint (empty indices = pass). */
export const zhengApi = {
  exec: (command: 'reset' | 'play', indices?: number[], config?: ZhengConfigInput) =>
    gameExec<ZhengResponse>('zheng', { command, indices, config }),
};
