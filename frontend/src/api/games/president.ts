// API client for president. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PresidentResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for President game settings. */
export interface PresidentConfigInput {
  revolutionEnabled?: boolean;
  cardExchangeEnabled?: boolean;
  passFieldFlushEnabled?: boolean;
  cpuDifficulty?: number;
}

/** Command verbs accepted by the President /president/exec endpoint. */
export type PresidentCommand = 'reset' | 'play' | 'log' | 'hint';

/** API client for the President /president/exec endpoint. */
export const presidentApi = {
  exec: (command: PresidentCommand, indices?: number[], config?: PresidentConfigInput) =>
    gameExec<PresidentResponse>('president', { command, indices, config }),
};
