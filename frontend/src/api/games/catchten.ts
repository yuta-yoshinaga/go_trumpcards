// API client for catchten. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CatchTenConfig, CatchTenResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Catch the Ten /catchten/exec endpoint. */
export const catchtenApi = {
  exec: (
    command: 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<CatchTenConfig>,
  ) => gameExec<CatchTenResponse>('catchten', { command, cardIndex, config }),
};
