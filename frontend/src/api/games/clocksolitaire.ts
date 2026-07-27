// API client for clocksolitaire. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ClockSolitaireResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Clock Solitaire /clocksolitaire/exec endpoint. */
export const clocksolitaireApi = {
  exec: (command: 'reset' | 'step' | 'autoplay' | 'undo' | 'log') =>
    gameExec<ClockSolitaireResponse>('clocksolitaire', { command }),
};
