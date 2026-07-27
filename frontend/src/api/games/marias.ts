// API client for marias. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MariasResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Mariáš game settings. */
export interface MariasConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Mariáš /marias/exec endpoint. */
export type MariasCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Mariáš /marias/exec endpoint.
 *
 * Mariáš is a Czech/Slovak 3-player 32-card trump trick-taker. A rotating
 * Soloist plays alone against the 2 Defenders; trump is the Soloist's longest
 * suit (auto). The only play action is playing a card (must follow, trump when
 * void); there are no declarations.
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const mariasApi = {
  exec: (
    command: MariasCommand,
    opts?: {
      cardIndex?: number;
      config?: MariasConfigInput;
    },
  ) =>
    gameExec<MariasResponse>('marias', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
