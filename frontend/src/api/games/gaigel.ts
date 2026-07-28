// API client for gaigel. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { GaigelResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Gaigel game configuration input. */
export interface GaigelConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/**
 * API client for the Gaigel /gaigel/exec endpoint.
 *
 * The second positional slot is unused (Gaigel has no suit/bid argument); it
 * exists so the exec signature matches the `(command, arg1?, cardIndex?, config?)`
 * shape that `useTrickGameBase` dispatches for reset/play.
 */
export const gaigelApi = {
  exec: (
    command: 'reset' | 'play' | 'marriage' | 'next' | 'nextround' | 'hint',
    _unused?: number,
    cardIndex?: number,
    config?: GaigelConfigInput,
  ) =>
    gameExec<GaigelResponse>('gaigel', {
      command,
      cardIndex,
      config,
    }),
};
