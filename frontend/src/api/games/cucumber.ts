// API client for cucumber. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CucumberConfig, CucumberResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Cucumber /cucumber/exec endpoint.
 *
 * **`next` deals the following round.** The game stops at the round boundary so
 * the penalty can be read, since nothing on the board records it.
 */
export const cucumberApi = {
  exec: (
    command: 'reset' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<CucumberConfig>,
  ) => gameExec<CucumberResponse>('cucumber', { command, cardIndex, config }),
};
