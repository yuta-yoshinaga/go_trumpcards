// API client for schnapsen. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SchnapsenConfig, SchnapsenResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Schnapsen /schnapsen/exec endpoint. */
export const schnapsenApi = {
  exec: (
    command: 'reset' | 'play' | 'marriage' | 'next' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<SchnapsenConfig>,
  ) => gameExec<SchnapsenResponse>('schnapsen', { command, cardIndex, config }),
};
