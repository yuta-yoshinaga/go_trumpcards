// API client for spoons. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SpoonsResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Spoons game settings. */
export interface SpoonsConfigInput {
  cpuDifficulty?: number;
}

/** Commands accepted by the Spoons /spoons/exec endpoint. */
export type SpoonsCommand = 'reset' | 'pass' | 'grab' | 'next' | 'log';

/**
 * API client for the Spoons /spoons/exec endpoint.
 *
 * Spoons is a 4-player pass-and-grab speed game. On the Pass phase the human
 * picks one of their four cards to pass to the next player (`pass` →
 * `{ cardIndex }`). When someone collects four of a kind the Grab window opens;
 * everyone races to `grab` a spoon — the one who misses out gains a letter
 * (S-P-O-O-N-S). `next` advances to the following round; `reset` applies the
 * config (CPU difficulty); `log` fetches the action log.
 *   - `pass` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `grab` / `next` / `log` carry no extra fields.
 */
export const spoonsApi = {
  exec: (
    command: SpoonsCommand,
    opts?: {
      cardIndex?: number;
      config?: SpoonsConfigInput;
    },
  ) =>
    gameExec<SpoonsResponse>('spoons', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
