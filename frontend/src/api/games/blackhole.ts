// API client for blackhole. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BlackHoleResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Commands accepted by the Black Hole /blackhole/exec endpoint. */
export type BlackHoleCommand = 'reset' | 'mb' | 'g' | 'u' | 'undo_n' | 'hint' | 'log';

/** API client for the Black Hole /blackhole/exec endpoint. */
export const blackholeApi = {
  exec: (command: BlackHoleCommand, opts?: { fan?: number; n?: number }) =>
    gameExec<BlackHoleResponse>('blackhole', {
      command,
      fan: opts?.fan,
      n: opts?.n,
    }),
};
