// API client for klaverjas. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { KlaverjasResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Klaverjas game settings. */
export interface KlaverjasConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Klaverjas /klaverjas/exec endpoint. */
export type KlaverjasCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Klaverjas /klaverjas/exec endpoint.
 *
 * Klaverjas is a Dutch 4-player (2 vs 2) trump trick-taker with Roem (run/marriage)
 * bonuses. The only play action is playing a card; there are no declarations.
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const klaverjasApi = {
  exec: (
    command: KlaverjasCommand,
    opts?: {
      cardIndex?: number;
      config?: KlaverjasConfigInput;
    },
  ) =>
    gameExec<KlaverjasResponse>('klaverjas', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
