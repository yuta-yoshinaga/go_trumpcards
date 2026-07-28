// API client for dragontiger. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { DragonTigerResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Dragon Tiger /dragontiger/exec endpoint. */
export const dragontigerApi = {
  exec: (command: 'reset' | 'bet' | 'clear' | 'log', amount?: number, betType?: number) =>
    gameExec<DragonTigerResponse>('dragontiger', { command, amount, betType }),
};
