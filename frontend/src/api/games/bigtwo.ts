// API client for bigtwo. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BigTwoConfigInput, BigTwoResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Big Two /bigtwo/exec endpoint. */
export const bigtwoApi = {
  exec: (command: 'reset' | 'play', indices?: number[], config?: BigTwoConfigInput) =>
    gameExec<BigTwoResponse>('bigtwo', { command, indices, config }),
};
