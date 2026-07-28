// API client for allfours. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { AllFoursResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for All Fours (Seven Up) game settings. */
export interface AllFoursConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the All Fours /allfours/exec endpoint. */
export const allfoursApi = {
  exec: (
    command: 'reset' | 'beg' | 'respond' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    beg?: boolean,
    run?: boolean,
    cardIndex?: number,
    config?: AllFoursConfigInput,
  ) => gameExec<AllFoursResponse>('allfours', { command, beg, run, cardIndex, config }),
};
