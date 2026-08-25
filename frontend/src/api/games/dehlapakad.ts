// API client for dehlapakad. Follows the split-out convention of gameApi.ts
// (issue #4434); gameApi.ts re-exports this file.

import type { DehlaPakadResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Dehla Pakad. */
export interface DehlaPakadConfigInput {
  cpuDifficulty?: number;
  /** Kots needed to take the match, 1-5. */
  targetKots?: number;
}

/** Commands accepted by the /dehlapakad/exec endpoint. */
export type DehlaPakadCommand = 'reset' | 'trump' | 'play' | 'nexthand' | 'hint' | 'log';

/**
 * API client for the Dehla Pakad /dehlapakad/exec endpoint.
 *
 *   - `trump` → `{ trumpSuit }` (1=spade 2=club 3=heart 4=diamond). Only the
 *     seat at `trumpChooserIdx` may call it, and only from its first five cards.
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 */
export const dehlaPakadApi = {
  exec: (
    command: DehlaPakadCommand,
    opts?: {
      cardIndex?: number;
      trumpSuit?: number;
      config?: DehlaPakadConfigInput;
    },
  ) =>
    gameExec<DehlaPakadResponse>('dehlapakad', {
      command,
      cardIndex: opts?.cardIndex,
      trumpSuit: opts?.trumpSuit,
      config: opts?.config,
    }),
};
