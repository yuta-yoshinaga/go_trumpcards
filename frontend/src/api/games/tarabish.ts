// API client for tarabish. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TarabishConfig, TarabishResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Tarabish /tarabish/exec endpoint. */
export const tarabishApi = {
  exec: (
    command: 'reset' | 'take' | 'pass' | 'play' | 'next' | 'giveup' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<TarabishConfig>,
  ) => gameExec<TarabishResponse>('tarabish', { command, cardIndex, config }),
};
