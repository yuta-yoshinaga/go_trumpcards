// API client for gongzhu. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { GongZhuResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Gong Zhu game settings. */
export interface GongZhuConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Gong Zhu /gongzhu/exec endpoint. */
export const gongzhuApi = {
  exec: (
    command: 'reset' | 'expose' | 'play' | 'next' | 'nextround' | 'hint',
    cardIndices?: number[],
    cardIndex?: number,
    config?: GongZhuConfigInput,
  ) =>
    gameExec<GongZhuResponse>('gongzhu', {
      command,
      cardIndices,
      cardIndex,
      config,
    }),
};
