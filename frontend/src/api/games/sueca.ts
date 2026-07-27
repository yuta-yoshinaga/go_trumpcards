// API client for sueca. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SuecaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Sueca game settings. */
export interface SuecaConfigInput {
  cpuDifficulty?: number;
  targetGamePoints?: number;
}

/** Commands accepted by the Sueca /sueca/exec endpoint. */
export type SuecaCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Sueca /sueca/exec endpoint.
 *
 * Sueca is a Portuguese/Brazilian 4-player (2 vs 2) trump trick-taker. The only
 * play action is playing a card; there are no declarations.
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const suecaApi = {
  exec: (
    command: SuecaCommand,
    opts?: {
      cardIndex?: number;
      config?: SuecaConfigInput;
    },
  ) =>
    gameExec<SuecaResponse>('sueca', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
