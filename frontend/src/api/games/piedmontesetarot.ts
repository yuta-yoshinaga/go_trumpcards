// API client for piedmontesetarot. Follows the split-out convention of
// gameApi.ts (issue #4434); gameApi.ts re-exports this file.

import type { PiedmonteseTarotResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Tarocco Piemontese. */
export interface PiedmonteseTarotConfigInput {
  /** 3 or 4. Anything else is refused rather than rounded. */
  seats?: number;
  cpuDifficulty?: number;
  targetDeals?: number;
}

/** Commands accepted by the /piedmontesetarot/exec endpoint. */
export type PiedmonteseTarotCommand = 'reset' | 'scarto' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Tarocco Piemontese /piedmontesetarot/exec endpoint.
 *
 * A four-player (or three-player) Piedmontese tarot trick-taker on the 78-card
 * deck. The human is seat 0.
 *   - `scarto` → `{ cardIndices }` (the talon the dealer buries: 2 at four
 *     seats, 3 at three)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const piedmonteseTarotApi = {
  exec: (
    command: PiedmonteseTarotCommand,
    opts?: {
      cardIndices?: number[];
      cardIndex?: number;
      config?: PiedmonteseTarotConfigInput;
    },
  ) =>
    gameExec<PiedmonteseTarotResponse>('piedmontesetarot', {
      command,
      cardIndices: opts?.cardIndices,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
