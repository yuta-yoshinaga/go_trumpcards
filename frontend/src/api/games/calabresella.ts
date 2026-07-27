// API client for calabresella. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CalabresellaResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Calabresella (Terziglio) game settings. */
export interface CalabresellaConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Calabresella /calabresella/exec endpoint. */
export type CalabresellaCommand = 'reset' | 'bid' | 'discard' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Calabresella (Terziglio) /calabresella/exec endpoint.
 *
 * Calabresella is a Calabrian/Italian 3-player 40-card (Tressette-family)
 * trick-taker with a Bid phase, a monte exchange (discard four) phase, and no
 * trump. One Soloist plays alone against the coalition of the other two.
 *   - `bid` → `{ bid: number }` (0=pass, 1=chiamo, 2=solo)
 *   - `discard` → `{ cardIndex }` (monte exchange: the Soloist discards one card
 *     per call, four times, from 16 down to 12)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const calabresellaApi = {
  exec: (
    command: CalabresellaCommand,
    opts?: {
      cardIndex?: number;
      bid?: number;
      config?: CalabresellaConfigInput;
    },
  ) =>
    gameExec<CalabresellaResponse>('calabresella', {
      command,
      cardIndex: opts?.cardIndex,
      bid: opts?.bid,
      config: opts?.config,
    }),
};
