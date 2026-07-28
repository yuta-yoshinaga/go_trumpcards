// API client for mississippistud. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MississippiStudResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Mississippi Stud /mississippistud/exec endpoint. */
export const mississippiStudApi = {
  exec: (command: 'reset' | 'bet' | 'play' | 'fold' | 'log', amount?: number, multiplier?: number) =>
    gameExec<MississippiStudResponse>('mississippistud', { command, amount, multiplier }),
};
