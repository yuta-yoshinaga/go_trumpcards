// API client for tute. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TuteResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Tute game settings. */
export interface TuteConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Tute /tute/exec endpoint. */
export type TuteCommand = 'reset' | 'play' | 'marriage' | 'tute' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Tute /tute/exec endpoint.
 *
 * Tute is a Spanish 4-player (2 vs 2) trump trick-taker. The play actions are:
 *   - `play` → `{ cardIndex: number }`
 *   - `marriage` → `{ suit: number }` (declare a King+Queen marriage; 1=♠ 2=♣ 3=♥ 4=♦)
 *   - `tute` → no extra fields (declare four Kings or four Queens for an instant win)
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const tuteApi = {
  exec: (
    command: TuteCommand,
    opts?: {
      cardIndex?: number;
      suit?: number;
      config?: TuteConfigInput;
    },
  ) =>
    gameExec<TuteResponse>('tute', {
      command,
      cardIndex: opts?.cardIndex,
      suit: opts?.suit,
      config: opts?.config,
    }),
};
