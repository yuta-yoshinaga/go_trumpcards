// API client for pigtail. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PigsTailResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Pig's Tail game API client. */
export const pigtailApi = {
  exec: (command: 'reset' | 'draw', cpuHesitationEnabled?: boolean, playerCount?: number) =>
    gameExec<PigsTailResponse>('pigtail', { command, cpuHesitationEnabled, playerCount }),
};
