// API client for pageone. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PageOneResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Page One game settings. */
export interface PageOneConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Page One /pageone/exec endpoint. */
export const pageoneApi = {
  exec: (
    command: 'reset' | 'play' | 'draw' | 'declare' | 'skip' | 'nextround' | 'hint',
    cardIndex?: number,
    config?: PageOneConfigInput,
  ) =>
    gameExec<PageOneResponse>('pageone', {
      command,
      cardIndex,
      config,
    }),
};
