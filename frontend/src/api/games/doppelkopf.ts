// API client for doppelkopf. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { DoppelkopfResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Doppelkopf game settings. */
export interface DoppelkopfConfigInput {
  cpuDifficulty?: number;
  baseChips?: number;
  startChips?: number;
  targetChips?: number;
}

/** Commands accepted by the Doppelkopf /doppelkopf/exec endpoint. */
export type DoppelkopfCommand = 'reset' | 'play' | 'announce' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Doppelkopf /doppelkopf/exec endpoint.
 *
 * Doppelkopf is a plain trick-taking flow (no pick/bury/call). The only extra
 * action beyond playing a card is `announce` (Re/Kontra, first trick only):
 *   - `play` → `{ cardIndex: number }`
 *   - `announce` → no extra fields (declares Re or Kontra based on the human's team)
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const doppelkopfApi = {
  exec: (
    command: DoppelkopfCommand,
    opts?: {
      cardIndex?: number;
      config?: DoppelkopfConfigInput;
    },
  ) =>
    gameExec<DoppelkopfResponse>('doppelkopf', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
