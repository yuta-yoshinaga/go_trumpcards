// API client for germanwhist. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { GermanWhistResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the German Whist /germanwhist/exec endpoint. */
export const germanwhistApi = {
  exec: (command: 'reset' | 'play' | 'giveup' | 'hint' | 'log', cardIndex?: number) =>
    gameExec<GermanWhistResponse>('germanwhist', { command, cardIndex }),
};
