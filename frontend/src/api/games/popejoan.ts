// API client for popejoan. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PopeJoanResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands the /popejoan/exec endpoint accepts. */
export type PopeJoanCommand = 'reset' | 'play' | 'next' | 'hint' | 'log';

/** API client for the Pope Joan /popejoan/exec endpoint. */
export const popejoanApi = {
  exec: (command: PopeJoanCommand, cardIndex?: number) =>
    gameExec<PopeJoanResponse>('popejoan', { command, cardIndex }),
};
