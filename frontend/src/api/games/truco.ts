// API client for truco. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TrucoConfig, TrucoResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Truco /truco/exec endpoint. */
export const trucoApi = {
  exec: (
    command: 'reset' | 'play' | 'truco' | 'accept' | 'decline' | 'next' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<TrucoConfig>,
  ) => gameExec<TrucoResponse>('truco', { command, cardIndex, config }),
};
