// API client for quodlibet. Follows the split-out convention of gameApi.ts
// (issue #4434); gameApi.ts re-exports this file.

import type { QuodlibetResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Quodlibet. */
export interface QuodlibetConfigInput {
  cpuDifficulty?: number;
  autoSelectContract?: boolean;
}

/** Commands accepted by the /quodlibet/exec endpoint. */
export type QuodlibetCommand = 'reset' | 'contract' | 'play' | 'pass' | 'nextdeal' | 'hint' | 'log';

/**
 * API client for the Quodlibet /quodlibet/exec endpoint.
 *
 *   - `contract` → `{ contract }`, and only a value from the response's
 *     `availableContracts` is accepted — the dealer may pick only from the
 *     wheel currently in play.
 *   - `play` → `{ cardIndex }`
 *   - `pass` → nothing; legal only when a shedding contract leaves the seat
 *     with no playable card.
 *   - `reset` → `{ config }`
 */
export const quodlibetApi = {
  exec: (
    command: QuodlibetCommand,
    opts?: {
      cardIndex?: number;
      contract?: number;
      config?: QuodlibetConfigInput;
    },
  ) =>
    gameExec<QuodlibetResponse>('quodlibet', {
      command,
      cardIndex: opts?.cardIndex,
      contract: opts?.contract,
      config: opts?.config,
    }),
};
