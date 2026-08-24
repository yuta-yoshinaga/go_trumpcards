// API client for unsunkaruta. Follows the split-out convention of gameApi.ts
// (issue #4434); gameApi.ts re-exports this file.

import type { UnsunKarutaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Unsun Karuta. */
export interface UnsunKarutaConfigInput {
  cpuDifficulty?: number;
  /** Deals per match, 1-8. */
  targetDeals?: number;
}

/** Commands accepted by the /unsunkaruta/exec endpoint. */
export type UnsunKarutaCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Unsun Karuta (うんすんカルタ) /unsunkaruta/exec endpoint.
 *
 *   - `play` → `{ cardIndex, declare? }`. **The declaration rides with the
 *     card**: sending it separately would leave a board where someone has
 *     declared without playing. `declare` is only meaningful on a lead.
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const unsunKarutaApi = {
  exec: (
    command: UnsunKarutaCommand,
    opts?: {
      cardIndex?: number;
      declare?: boolean;
      config?: UnsunKarutaConfigInput;
    },
  ) =>
    gameExec<UnsunKarutaResponse>('unsunkaruta', {
      command,
      cardIndex: opts?.cardIndex,
      declare: opts?.declare,
      config: opts?.config,
    }),
};
